package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "TEST_ERROR",
		Message:    "This is a test error",
	}

	assert.Equal(t, "This is a test error", err.Error())
}

func TestPredefinedErrors(t *testing.T) {
	testCases := []struct {
		name       string
		err        *APIError
		httpStatus int
		code       string
		message    string
	}{
		{"ErrBadRequest", ErrBadRequest, http.StatusBadRequest, "BAD_REQUEST", "Invalid request parameters"},
		{"ErrInvalidJSON", ErrInvalidJSON, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format"},
		{"ErrValidation", ErrValidation, http.StatusBadRequest, "VALIDATION_FAILED", "Input validation failed"},
		{"ErrDuplicateResource", ErrDuplicateResource, http.StatusConflict, "DUPLICATE_RESOURCE", "Resource already exists"},
		{"ErrResourceNotFound", ErrResourceNotFound, http.StatusNotFound, "NOT_FOUND", "Resource not found"},
		{"ErrInternalServer", ErrInternalServer, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred"},
		{"ErrDatabase", ErrDatabase, http.StatusInternalServerError, "DATABASE_ERROR", "Database operation failed"},
		{"ErrUnauthorized", ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed"},
		{"ErrForbidden", ErrForbidden, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource"},
		{"ErrTaskInProgress", ErrTaskInProgress, http.StatusConflict, "TASK_IN_PROGRESS", "A task is already in progress"},
		{"ErrBadGateway", ErrBadGateway, http.StatusBadGateway, "BAD_GATEWAY", "Upstream service error"},
		{"ErrNoActiveKeys", ErrNoActiveKeys, http.StatusServiceUnavailable, "NO_ACTIVE_KEYS", "No active API keys available for this group"},
		{"ErrMaxRetriesExceeded", ErrMaxRetriesExceeded, http.StatusBadGateway, "MAX_RETRIES_EXCEEDED", "Request failed after maximum retries"},
		{"ErrNoKeysAvailable", ErrNoKeysAvailable, http.StatusServiceUnavailable, "NO_KEYS_AVAILABLE", "No API keys available to process the request"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotNil(t, tc.err)
			assert.Equal(t, tc.httpStatus, tc.err.HTTPStatus)
			assert.Equal(t, tc.code, tc.err.Code)
			assert.Equal(t, tc.message, tc.err.Message)
			assert.Equal(t, tc.message, tc.err.Error())
		})
	}
}

func TestNewAPIError(t *testing.T) {
	base := ErrBadRequest
	customMessage := "Custom error message"

	newErr := NewAPIError(base, customMessage)

	assert.Equal(t, base.HTTPStatus, newErr.HTTPStatus)
	assert.Equal(t, base.Code, newErr.Code)
	assert.Equal(t, customMessage, newErr.Message)
	assert.Equal(t, customMessage, newErr.Error())
}

func TestNewAPIErrorWithUpstream(t *testing.T) {
	statusCode := http.StatusBadGateway
	code := "UPSTREAM_ERROR"
	message := "Upstream service returned an error"

	err := NewAPIErrorWithUpstream(statusCode, code, message)

	assert.Equal(t, statusCode, err.HTTPStatus)
	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, message, err.Error())
}

func TestErrorComparison(t *testing.T) {
	t.Run("errors.Is works correctly", func(t *testing.T) {
		err := ErrNoActiveKeys
		assert.True(t, errors.Is(err, ErrNoActiveKeys))
		assert.False(t, errors.Is(err, ErrBadRequest))
	})

	t.Run("different instances of same type", func(t *testing.T) {
		err1 := NewAPIError(ErrBadRequest, "Custom message 1")
		err2 := NewAPIError(ErrBadRequest, "Custom message 2")

		// They should not be equal via errors.Is since they're different instances
		assert.False(t, errors.Is(err1, err2))

		// But they should have the same HTTP status and code
		assert.Equal(t, err1.HTTPStatus, err2.HTTPStatus)
		assert.Equal(t, err1.Code, err2.Code)
	})
}

func TestErrorUniqueness(t *testing.T) {
	// 确保所有预定义错误都是唯一的
	allErrors := []*APIError{
		ErrBadRequest,
		ErrInvalidJSON,
		ErrValidation,
		ErrDuplicateResource,
		ErrResourceNotFound,
		ErrInternalServer,
		ErrDatabase,
		ErrUnauthorized,
		ErrForbidden,
		ErrTaskInProgress,
		ErrBadGateway,
		ErrNoActiveKeys,
		ErrMaxRetriesExceeded,
		ErrNoKeysAvailable,
	}

	// 检查错误代码的唯一性
	codes := make(map[string]bool)
	for _, err := range allErrors {
		assert.False(t, codes[err.Code], "Duplicate error code: %s", err.Code)
		codes[err.Code] = true
	}

	// 检查错误消息的唯一性
	messages := make(map[string]bool)
	for _, err := range allErrors {
		msg := err.Message
		assert.False(t, messages[msg], "Duplicate error message: %s", msg)
		messages[msg] = true
	}
}

func TestParseDBError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		result := ParseDBError(nil)
		assert.Nil(t, result)
	})

	t.Run("record not found", func(t *testing.T) {
		err := errors.New("record not found")
		// Mock gorm.ErrRecordNotFound
		result := ParseDBError(err)
		// Since we can't easily mock gorm.ErrRecordNotFound without importing gorm,
		// this will return ErrDatabase for a generic error
		assert.Equal(t, ErrDatabase, result)
	})

	t.Run("SQLite unique constraint", func(t *testing.T) {
		err := errors.New("UNIQUE constraint failed: users.email")
		result := ParseDBError(err)
		assert.Equal(t, ErrDuplicateResource, result)
	})

	t.Run("generic database error", func(t *testing.T) {
		err := errors.New("some database error")
		result := ParseDBError(err)
		assert.Equal(t, ErrDatabase, result)
	})
}
