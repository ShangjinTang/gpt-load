// Package proxy - Request execution logic refactored for better testability
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gpt-load/internal/channel"
	app_errors "gpt-load/internal/errors"
	"gpt-load/internal/models"
	"gpt-load/internal/response"
	"gpt-load/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestContext encapsulates all the context needed for a single request attempt
type RequestContext struct {
	GinContext     *gin.Context
	ChannelHandler channel.ChannelProxy
	Group          *models.Group
	BodyBytes      []byte
	IsStream       bool
	StartTime      time.Time
	RetryCount     int
	APIKey         *models.APIKey
	UpstreamURL    string
}

// RequestResult represents the result of a request execution
type RequestResult struct {
	Response    *http.Response
	Error       error
	StatusCode  int
	ErrorBody   string
	ParsedError string
}

// RequestExecutor handles the execution of individual requests with proper separation of concerns
type RequestExecutor struct {
	proxyServer *ProxyServer
}

// NewRequestExecutor creates a new RequestExecutor
func NewRequestExecutor(ps *ProxyServer) *RequestExecutor {
	return &RequestExecutor{proxyServer: ps}
}

// ExecuteWithRetry is the main entry point for request execution with retry logic
func (re *RequestExecutor) ExecuteWithRetry(ctx *RequestContext) {
	// Step 1: Select API key
	if err := re.selectAPIKey(ctx); err != nil {
		re.handleKeySelectionError(ctx, err)
		return
	}

	// Step 2: Build upstream URL
	if err := re.buildUpstreamURL(ctx); err != nil {
		re.handleUpstreamURLError(ctx, err)
		return
	}

	// Step 3: Execute the request
	result := re.executeRequest(ctx)

	// Step 4: Handle the result
	if re.shouldRetry(ctx, result) {
		re.handleRetryableError(ctx, result)
		return
	}

	if result.Error != nil || result.StatusCode >= 400 {
		re.handleFinalError(ctx, result)
		return
	}

	// Success case
	re.handleSuccess(ctx, result)
}

// selectAPIKey selects an appropriate API key for the request
func (re *RequestExecutor) selectAPIKey(ctx *RequestContext) error {
	apiKey, err := re.proxyServer.keyProvider.SelectKey(ctx.Group.ID)
	if err != nil {
		return fmt.Errorf("failed to select key for group %s on attempt %d: %w",
			ctx.Group.Name, ctx.RetryCount+1, err)
	}
	ctx.APIKey = apiKey
	return nil
}

// buildUpstreamURL builds the upstream URL for the request
func (re *RequestExecutor) buildUpstreamURL(ctx *RequestContext) error {
	upstreamURL, err := ctx.ChannelHandler.BuildUpstreamURL(ctx.GinContext.Request.URL, ctx.Group)
	if err != nil {
		return fmt.Errorf("failed to build upstream URL: %w", err)
	}
	ctx.UpstreamURL = upstreamURL
	return nil
}

// executeRequest performs the actual HTTP request
func (re *RequestExecutor) executeRequest(ctx *RequestContext) *RequestResult {
	// Create HTTP request
	req, err := re.buildHTTPRequest(ctx)
	if err != nil {
		return &RequestResult{
			Error:      err,
			StatusCode: 500,
		}
	}

	// Select appropriate HTTP client
	client := re.selectHTTPClient(ctx)

	// Execute request
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return &RequestResult{
			Error:      err,
			StatusCode: 500,
		}
	}

	// Handle response
	if resp.StatusCode >= 400 {
		errorBody, parsedError := re.extractErrorFromResponse(resp)
		return &RequestResult{
			Response:    resp,
			StatusCode:  resp.StatusCode,
			ErrorBody:   errorBody,
			ParsedError: parsedError,
		}
	}

	return &RequestResult{
		Response:   resp,
		StatusCode: resp.StatusCode,
	}
}

// buildHTTPRequest creates the HTTP request with proper headers and context
func (re *RequestExecutor) buildHTTPRequest(ctx *RequestContext) (*http.Request, error) {
	// Create context with appropriate timeout
	reqCtx, cancel := re.createRequestContext(ctx)
	defer cancel()

	// Create HTTP request
	req, err := http.NewRequestWithContext(
		reqCtx,
		ctx.GinContext.Request.Method,
		ctx.UpstreamURL,
		bytes.NewReader(ctx.BodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.ContentLength = int64(len(ctx.BodyBytes))

	// Clone and clean headers
	req.Header = ctx.GinContext.Request.Header.Clone()
	re.cleanAuthHeaders(req)

	// Apply channel-specific modifications
	ctx.ChannelHandler.ModifyRequest(req, ctx.APIKey, ctx.Group)

	// Apply custom header rules
	re.applyHeaderRules(req, ctx)

	// Set streaming headers if needed
	if ctx.IsStream {
		req.Header.Set("X-Accel-Buffering", "no")
	}

	return req, nil
}

// createRequestContext creates appropriate context with timeout for the request
func (re *RequestExecutor) createRequestContext(ctx *RequestContext) (context.Context, context.CancelFunc) {
	if ctx.IsStream {
		return context.WithCancel(ctx.GinContext.Request.Context())
	}

	timeout := time.Duration(ctx.Group.EffectiveConfig.RequestTimeout) * time.Second
	return context.WithTimeout(ctx.GinContext.Request.Context(), timeout)
}

// cleanAuthHeaders removes client authentication headers
func (re *RequestExecutor) cleanAuthHeaders(req *http.Request) {
	req.Header.Del("Authorization")
	req.Header.Del("X-Api-Key")
	req.Header.Del("X-Goog-Api-Key")
}

// applyHeaderRules applies custom header rules if configured
func (re *RequestExecutor) applyHeaderRules(req *http.Request, ctx *RequestContext) {
	if len(ctx.Group.HeaderRuleList) > 0 {
		headerCtx := utils.NewHeaderVariableContextFromGin(ctx.GinContext, ctx.Group, ctx.APIKey)
		utils.ApplyHeaderRules(req, ctx.Group.HeaderRuleList, headerCtx)
	}
}

// selectHTTPClient selects the appropriate HTTP client based on request type
func (re *RequestExecutor) selectHTTPClient(ctx *RequestContext) *http.Client {
	if ctx.IsStream {
		return ctx.ChannelHandler.GetStreamClient()
	}
	return ctx.ChannelHandler.GetHTTPClient()
}

// extractErrorFromResponse extracts and parses error information from HTTP response
func (re *RequestExecutor) extractErrorFromResponse(resp *http.Response) (string, string) {
	errorBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logrus.Errorf("Failed to read error body: %v", readErr)
		return "Failed to read error body", "Failed to read error body"
	}

	errorBody = handleGzipCompression(resp, errorBody)
	errorMessage := string(errorBody)
	parsedError := app_errors.ParseUpstreamError(errorBody)

	return errorMessage, parsedError
}

// shouldRetry determines if the request should be retried
func (re *RequestExecutor) shouldRetry(ctx *RequestContext, result *RequestResult) bool {
	// Don't retry if it's the last attempt
	if ctx.RetryCount >= ctx.Group.EffectiveConfig.MaxRetries {
		return false
	}

	// Don't retry if there's no error
	if result.Error == nil && result.StatusCode < 400 {
		return false
	}

	// Don't retry 404 errors
	if result.StatusCode == http.StatusNotFound {
		return false
	}

	// Don't retry ignorable client-side errors
	if result.Error != nil && app_errors.IsIgnorableError(result.Error) {
		logrus.Debugf("Client-side ignorable error for key %s, aborting retries: %v",
			utils.MaskAPIKey(ctx.APIKey.KeyValue), result.Error)
		return false
	}

	return true
}

// handleKeySelectionError handles errors during key selection
func (re *RequestExecutor) handleKeySelectionError(ctx *RequestContext, err error) {
	logrus.Errorf("Failed to select a key for group %s on attempt %d: %v",
		ctx.Group.Name, ctx.RetryCount+1, err)
	response.Error(ctx.GinContext, app_errors.NewAPIError(app_errors.ErrNoKeysAvailable, err.Error()))
	re.proxyServer.logRequest(ctx.GinContext, ctx.Group, nil, ctx.StartTime,
		http.StatusServiceUnavailable, err, ctx.IsStream, "", ctx.ChannelHandler,
		ctx.BodyBytes, models.RequestTypeFinal)
}

// handleUpstreamURLError handles errors during upstream URL building
func (re *RequestExecutor) handleUpstreamURLError(ctx *RequestContext, err error) {
	response.Error(ctx.GinContext, app_errors.NewAPIError(app_errors.ErrInternalServer, err.Error()))
}

// handleRetryableError handles errors that should trigger a retry
func (re *RequestExecutor) handleRetryableError(ctx *RequestContext, result *RequestResult) {
	// Update key status
	errorMsg := result.ParsedError
	if errorMsg == "" && result.Error != nil {
		errorMsg = result.Error.Error()
	}

	re.proxyServer.keyProvider.UpdateStatus(ctx.APIKey, ctx.Group, false, errorMsg)

	// Log the retry attempt
	var logError error
	if result.Error != nil {
		logError = result.Error
	} else {
		logError = errors.New(result.ParsedError)
	}

	re.proxyServer.logRequest(ctx.GinContext, ctx.Group, ctx.APIKey, ctx.StartTime,
		result.StatusCode, logError, ctx.IsStream, ctx.UpstreamURL, ctx.ChannelHandler,
		ctx.BodyBytes, models.RequestTypeRetry)

	// Log debug information
	if result.Error != nil {
		logrus.Debugf("Request failed (attempt %d/%d) for key %s: %v",
			ctx.RetryCount+1, ctx.Group.EffectiveConfig.MaxRetries,
			utils.MaskAPIKey(ctx.APIKey.KeyValue), result.Error)
	} else {
		logrus.Debugf("Request failed with status %d (attempt %d/%d) for key %s. Parsed Error: %s",
			result.StatusCode, ctx.RetryCount+1, ctx.Group.EffectiveConfig.MaxRetries,
			utils.MaskAPIKey(ctx.APIKey.KeyValue), result.ParsedError)
	}

	// Retry with incremented count
	retryCtx := *ctx
	retryCtx.RetryCount++
	re.ExecuteWithRetry(&retryCtx)
}

// handleFinalError handles errors on the final attempt (no more retries)
func (re *RequestExecutor) handleFinalError(ctx *RequestContext, result *RequestResult) {
	// Update key status
	errorMsg := result.ParsedError
	if errorMsg == "" && result.Error != nil {
		errorMsg = result.Error.Error()
	}

	re.proxyServer.keyProvider.UpdateStatus(ctx.APIKey, ctx.Group, false, errorMsg)

	// Log the final error
	var logError error
	if result.Error != nil {
		logError = result.Error
	} else {
		logError = errors.New(result.ParsedError)
	}

	re.proxyServer.logRequest(ctx.GinContext, ctx.Group, ctx.APIKey, ctx.StartTime,
		result.StatusCode, logError, ctx.IsStream, ctx.UpstreamURL, ctx.ChannelHandler,
		ctx.BodyBytes, models.RequestTypeFinal)

	// Return error to client
	if result.Error != nil && app_errors.IsIgnorableError(result.Error) {
		// For ignorable errors, just return without sending response (client disconnected)
		return
	}

	// Try to parse error as JSON first
	if result.ErrorBody != "" {
		var errorJSON map[string]any
		if err := json.Unmarshal([]byte(result.ErrorBody), &errorJSON); err == nil {
			ctx.GinContext.JSON(result.StatusCode, errorJSON)
			return
		}
	}

	// Fallback to structured error response
	errorMessage := result.ErrorBody
	if errorMessage == "" && result.Error != nil {
		errorMessage = result.Error.Error()
	}
	response.Error(ctx.GinContext, app_errors.NewAPIErrorWithUpstream(
		result.StatusCode, "UPSTREAM_ERROR", errorMessage))
}

// handleSuccess handles successful requests
func (re *RequestExecutor) handleSuccess(ctx *RequestContext, result *RequestResult) {
	logrus.Debugf("Request for group %s succeeded on attempt %d with key %s",
		ctx.Group.Name, ctx.RetryCount+1, utils.MaskAPIKey(ctx.APIKey.KeyValue))

	// Copy response headers
	for key, values := range result.Response.Header {
		for _, value := range values {
			ctx.GinContext.Header(key, value)
		}
	}
	ctx.GinContext.Status(result.Response.StatusCode)

	// Handle response body
	if ctx.IsStream {
		re.proxyServer.handleStreamingResponse(ctx.GinContext, result.Response)
	} else {
		re.proxyServer.handleNormalResponse(ctx.GinContext, result.Response)
	}

	// Log successful request
	re.proxyServer.logRequest(ctx.GinContext, ctx.Group, ctx.APIKey, ctx.StartTime,
		result.Response.StatusCode, nil, ctx.IsStream, ctx.UpstreamURL, ctx.ChannelHandler,
		ctx.BodyBytes, models.RequestTypeFinal)
}
