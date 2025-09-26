package services

import (
	"testing"

	"gpt-load/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKeyStateService 模拟密钥状态服务
type MockKeyStateService struct {
	mock.Mock
}

func (m *MockKeyStateService) HandleSuccess(keyID uint) error {
	args := m.Called(keyID)
	return args.Error(0)
}

func (m *MockKeyStateService) HandleFailure(keyID uint, errorMessage string) error {
	args := m.Called(keyID, errorMessage)
	return args.Error(0)
}

func (m *MockKeyStateService) ManuallyInvalidateKey(keyID uint, reason string) error {
	args := m.Called(keyID, reason)
	return args.Error(0)
}

func (m *MockKeyStateService) ManuallyDisableKey(keyID uint, reason string) error {
	args := m.Called(keyID, reason)
	return args.Error(0)
}

func (m *MockKeyStateService) ManuallyEnableKey(keyID uint) error {
	args := m.Called(keyID)
	return args.Error(0)
}

// MockPolicyEngine 模拟策略引擎
type MockPolicyEngine struct {
	mock.Mock
}

func (m *MockPolicyEngine) EvaluateRetryPolicies(groupID uint, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	args := m.Called(groupID, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PolicyEvaluationResult), args.Error(1)
}

func (m *MockPolicyEngine) EvaluateDegradationPolicies(groupID uint, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	args := m.Called(groupID, context)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PolicyEvaluationResult), args.Error(1)
}

func setupKeyPolicyHandlerTest() (*KeyPolicyHandler, *MockKeyStateService, *MockPolicyEngine) {
	mockKeyStateService := &MockKeyStateService{}
	mockPolicyEngine := &MockPolicyEngine{}

	handler := NewKeyPolicyHandler(mockPolicyEngine, mockKeyStateService)

	return handler, mockKeyStateService, mockPolicyEngine
}

func TestKeyPolicyHandler_HandleKeySuccess(t *testing.T) {
	handler, mockKeyStateService, _ := setupKeyPolicyHandlerTest()

	keyID := uint(123)

	// 设置期望
	mockKeyStateService.On("HandleSuccess", keyID).Return(nil)

	// 执行
	handler.HandleKeySuccess(keyID)

	// 验证调用
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeySuccess_Error(t *testing.T) {
	handler, mockKeyStateService, _ := setupKeyPolicyHandlerTest()

	keyID := uint(123)

	// 设置期望（返回错误）
	mockKeyStateService.On("HandleSuccess", keyID).Return(assert.AnError)

	// 执行（不应该panic，错误会被记录）
	handler.HandleKeySuccess(keyID)

	// 验证调用
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_NoMatchingPolicy(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 5,
		RequestCount: 100,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "API rate limit exceeded"

	// 设置策略引擎期望（没有匹配的策略）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(nil, nil)

	// 设置密钥状态服务期望（默认失败处理）
	mockKeyStateService.On("HandleFailure", apiKey.ID, errorMessage).Return(nil)

	// 执行
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_PolicyMatchInvalidate(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 2,
		RequestCount: 50,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "invalid api key"

	// 设置策略引擎期望（匹配无效化策略）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	policyResult := &models.PolicyEvaluationResult{
		Matched:    true,
		PolicyName: "invalidate-policy",
		RuleName:   "invalid-key-rule",
		Action:     models.RetryActionInvalidate,
		Reason:     "Matched invalidate policy",
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(policyResult, nil)

	// 设置密钥状态服务期望（无效化操作）
	mockKeyStateService.On("ManuallyInvalidateKey", apiKey.ID, policyResult.Reason).Return(nil)

	// 执行
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_PolicyMatchDisable(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 10,
		RequestCount: 200,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "too many requests"

	// 设置策略引擎期望（匹配禁用策略）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	policyResult := &models.PolicyEvaluationResult{
		Matched:    true,
		PolicyName: "disable-policy",
		RuleName:   "rate-limit-rule",
		Action:     models.RetryActionDisable,
		Reason:     "Matched disable policy",
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(policyResult, nil)

	// 设置密钥状态服务期望（禁用操作）
	mockKeyStateService.On("ManuallyDisableKey", apiKey.ID, policyResult.Reason).Return(nil)

	// 执行
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_PolicyMatchRetry(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 1,
		RequestCount: 10,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "temporary network error"

	// 设置策略引擎期望（匹配重试策略）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	policyResult := &models.PolicyEvaluationResult{
		Matched:    true,
		PolicyName: "retry-policy",
		RuleName:   "network-error-rule",
		Action:     models.RetryActionRetry,
		MaxRetries: 3,
		BackoffMs:  1000,
		Reason:     "Matched retry policy",
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(policyResult, nil)

	// 设置密钥状态服务期望（默认失败处理，因为是重试动作）
	mockKeyStateService.On("HandleFailure", apiKey.ID, errorMessage).Return(nil)

	// 执行
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_PolicyEngineError(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 3,
		RequestCount: 75,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "database connection error"

	// 设置策略引擎期望（返回错误）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(nil, assert.AnError)

	// 设置密钥状态服务期望（默认失败处理）
	mockKeyStateService.On("HandleFailure", apiKey.ID, errorMessage).Return(nil)

	// 执行
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_StateServiceError(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 1,
		RequestCount: 25,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "service unavailable"

	// 设置策略引擎期望（匹配无效化策略）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	policyResult := &models.PolicyEvaluationResult{
		Matched:    true,
		PolicyName: "invalidate-policy",
		RuleName:   "service-error-rule",
		Action:     models.RetryActionInvalidate,
		Reason:     "Service unavailable error",
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(policyResult, nil)

	// 设置密钥状态服务期望（返回错误）
	mockKeyStateService.On("ManuallyInvalidateKey", apiKey.ID, policyResult.Reason).Return(assert.AnError)

	// 执行（不应该panic，错误会被记录）
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}

func TestKeyPolicyHandler_HandleKeyFailure_PolicyNotMatched(t *testing.T) {
	handler, mockKeyStateService, mockPolicyEngine := setupKeyPolicyHandlerTest()

	apiKey := &models.APIKey{
		ID:           123,
		FailureCount: 1,
		RequestCount: 10,
	}
	group := &models.Group{
		ID: 456,
	}
	errorMessage := "unknown error"

	// 设置策略引擎期望（返回结果但不匹配）
	expectedContext := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}
	policyResult := &models.PolicyEvaluationResult{
		Matched: false, // 不匹配
	}
	mockPolicyEngine.On("EvaluateRetryPolicies", group.ID, expectedContext).Return(policyResult, nil)

	// 设置密钥状态服务期望（默认失败处理）
	mockKeyStateService.On("HandleFailure", apiKey.ID, errorMessage).Return(nil)

	// 执行
	handler.HandleKeyFailure(apiKey, group, errorMessage)

	// 验证调用
	mockPolicyEngine.AssertExpectations(t)
	mockKeyStateService.AssertExpectations(t)
}
