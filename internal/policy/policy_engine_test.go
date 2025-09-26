package policy

import (
	"encoding/json"
	"testing"

	"gpt-load/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestPolicyEngine(t *testing.T) (*PolicyEngine, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 创建表
	err = db.AutoMigrate(&models.Policy{}, &models.GroupPolicy{}, &models.Group{})
	require.NoError(t, err)

	engine := NewPolicyEngine(db)
	return engine, db
}

func createTestRetryPolicy(t *testing.T, db *gorm.DB) *models.Policy {
	config := models.RetryPolicyConfig{
		Rules: []models.RetryRule{
			{
				Name: "auth_error_invalidate",
				Conditions: []models.RetryCondition{
					{
						Type:     "status_code",
						Operator: "in",
						Values:   []string{"401", "403"},
					},
				},
				Action:   models.RetryActionInvalidate,
				Priority: 0,
			},
			{
				Name: "rate_limit_retry",
				Conditions: []models.RetryCondition{
					{
						Type:     "status_code",
						Operator: "equals",
						Value:    "429",
					},
				},
				Action:     models.RetryActionRetry,
				MaxRetries: 3,
				BackoffMs:  5000,
				Priority:   1,
			},
		},
	}

	configBytes, err := json.Marshal(config)
	require.NoError(t, err)

	policy := &models.Policy{
		Name:        "test_retry_policy",
		Description: "Test Retry Policy",
		Type:        models.PolicyTypeRetry,
		Config:      configBytes,
		IsActive:    true,
	}

	err = db.Create(policy).Error
	require.NoError(t, err)

	return policy
}

func createTestModelFilterPolicy(t *testing.T, db *gorm.DB) *models.Policy {
	config := models.ModelFilterPolicyConfig{
		Type:     "include",
		Patterns: []string{"gpt-.*", "claude-.*"},
	}

	configBytes, err := json.Marshal(config)
	require.NoError(t, err)

	policy := &models.Policy{
		Name:        "test_model_filter_policy",
		Description: "Test Model Filter Policy",
		Type:        models.PolicyTypeModelFilter,
		Config:      configBytes,
		IsActive:    true,
	}

	err = db.Create(policy).Error
	require.NoError(t, err)

	return policy
}

func createTestGroup(t *testing.T, db *gorm.DB) *models.Group {
	group := &models.Group{
		Name:      "test_group",
		Upstreams: []byte(`["openai"]`),
	}

	err := db.Create(group).Error
	require.NoError(t, err)

	return group
}

func TestPolicyEngine_EvaluateRetryPolicies(t *testing.T) {
	engine, db := setupTestPolicyEngine(t)

	// 创建测试数据
	group := createTestGroup(t, db)
	policy := createTestRetryPolicy(t, db)

	// 创建分组策略关联
	groupPolicy := &models.GroupPolicy{
		GroupID:  group.ID,
		PolicyID: policy.ID,
		Priority: 1,
		IsActive: true,
	}
	err := db.Create(groupPolicy).Error
	require.NoError(t, err)

	testCases := []struct {
		name           string
		context        *models.PolicyEvaluationContext
		expectedAction string
		expectedMatch  bool
	}{
		{
			name: "401 error should invalidate",
			context: &models.PolicyEvaluationContext{
				StatusCode:   401,
				ErrorMessage: "Unauthorized",
				KeyID:        1,
				GroupID:      group.ID,
			},
			expectedAction: models.RetryActionInvalidate,
			expectedMatch:  true,
		},
		{
			name: "403 error should invalidate",
			context: &models.PolicyEvaluationContext{
				StatusCode:   403,
				ErrorMessage: "Forbidden",
				KeyID:        1,
				GroupID:      group.ID,
			},
			expectedAction: models.RetryActionInvalidate,
			expectedMatch:  true,
		},
		{
			name: "429 error should retry",
			context: &models.PolicyEvaluationContext{
				StatusCode:   429,
				ErrorMessage: "Too Many Requests",
				KeyID:        1,
				GroupID:      group.ID,
			},
			expectedAction: models.RetryActionRetry,
			expectedMatch:  true,
		},
		{
			name: "500 error should not match",
			context: &models.PolicyEvaluationContext{
				StatusCode:   500,
				ErrorMessage: "Internal Server Error",
				KeyID:        1,
				GroupID:      group.ID,
			},
			expectedMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateRetryPolicies(group.ID, tc.context)
			require.NoError(t, err)

			if tc.expectedMatch {
				require.NotNil(t, result)
				assert.True(t, result.Matched)
				assert.Equal(t, tc.expectedAction, result.Action)
				assert.Equal(t, policy.ID, result.PolicyID)
				assert.Equal(t, policy.Name, result.PolicyName)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestPolicyEngine_EvaluateModelFilterPolicies(t *testing.T) {
	engine, db := setupTestPolicyEngine(t)

	// 创建测试数据
	group := createTestGroup(t, db)
	policy := createTestModelFilterPolicy(t, db)

	// 创建分组策略关联
	groupPolicy := &models.GroupPolicy{
		GroupID:  group.ID,
		PolicyID: policy.ID,
		Priority: 1,
		IsActive: true,
	}
	err := db.Create(groupPolicy).Error
	require.NoError(t, err)

	testCases := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "gpt-3.5-turbo should be allowed",
			model:    "gpt-3.5-turbo",
			expected: true,
		},
		{
			name:     "gpt-4 should be allowed",
			model:    "gpt-4",
			expected: true,
		},
		{
			name:     "claude-3-opus should be allowed",
			model:    "claude-3-opus",
			expected: true,
		},
		{
			name:     "gemini-pro should be rejected",
			model:    "gemini-pro",
			expected: false,
		},
		{
			name:     "unknown-model should be rejected",
			model:    "unknown-model",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := engine.EvaluateModelFilterPolicies(group.ID, tc.model)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, allowed)
		})
	}
}

func TestPolicyEngine_evaluateCondition(t *testing.T) {
	engine, _ := setupTestPolicyEngine(t)

	context := &models.PolicyEvaluationContext{
		StatusCode:   429,
		ErrorMessage: "Rate limit exceeded",
		ErrorType:    "rate_limit",
		Model:        "gpt-3.5-turbo",
		FailureCount: 5,
		RequestCount: 100,
	}

	testCases := []struct {
		name      string
		condition models.RetryCondition
		expected  bool
	}{
		{
			name: "status_code equals",
			condition: models.RetryCondition{
				Type:     "status_code",
				Operator: "equals",
				Value:    "429",
			},
			expected: true,
		},
		{
			name: "error_message contains",
			condition: models.RetryCondition{
				Type:     "error_message",
				Operator: "contains",
				Value:    "Rate limit",
			},
			expected: true,
		},
		{
			name: "status_code in list",
			condition: models.RetryCondition{
				Type:     "status_code",
				Operator: "in",
				Values:   []string{"401", "403", "429"},
			},
			expected: true,
		},
		{
			name: "status_code not in list",
			condition: models.RetryCondition{
				Type:     "status_code",
				Operator: "not_in",
				Values:   []string{"200", "201", "204"},
			},
			expected: true,
		},
		{
			name: "failure_count greater than",
			condition: models.RetryCondition{
				Type:     "failure_count",
				Operator: "gt",
				Value:    "3",
			},
			expected: true,
		},
		{
			name: "model regex match",
			condition: models.RetryCondition{
				Type:     "model",
				Operator: "regex",
				Value:    "^gpt-.*",
			},
			expected: true,
		},
		{
			name: "status_code not equals",
			condition: models.RetryCondition{
				Type:     "status_code",
				Operator: "equals",
				Value:    "500",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := engine.evaluateCondition(tc.condition, context)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPolicyEngine_GetGroupPolicies(t *testing.T) {
	engine, db := setupTestPolicyEngine(t)

	// 创建测试数据
	group := createTestGroup(t, db)
	policy1 := createTestRetryPolicy(t, db)
	policy2 := createTestModelFilterPolicy(t, db)

	// 创建分组策略关联
	groupPolicies := []*models.GroupPolicy{
		{
			GroupID:  group.ID,
			PolicyID: policy1.ID,
			Priority: 2,
			IsActive: true,
		},
		{
			GroupID:  group.ID,
			PolicyID: policy2.ID,
			Priority: 1,
			IsActive: true,
		},
	}

	for _, gp := range groupPolicies {
		err := db.Create(gp).Error
		require.NoError(t, err)
	}

	// 测试获取分组策略
	result, err := engine.GetGroupPolicies(group.ID)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// 验证策略已加载
	for _, gp := range result {
		assert.NotEmpty(t, gp.Policy.Name)
		assert.True(t, gp.IsActive)
	}
}

func TestPolicyEngine_GetGroupPoliciesByType(t *testing.T) {
	engine, db := setupTestPolicyEngine(t)

	// 创建测试数据
	group := createTestGroup(t, db)
	retryPolicy := createTestRetryPolicy(t, db)
	filterPolicy := createTestModelFilterPolicy(t, db)

	// 创建分组策略关联
	groupPolicies := []*models.GroupPolicy{
		{
			GroupID:  group.ID,
			PolicyID: retryPolicy.ID,
			Priority: 1,
			IsActive: true,
		},
		{
			GroupID:  group.ID,
			PolicyID: filterPolicy.ID,
			Priority: 2,
			IsActive: true,
		},
	}

	for _, gp := range groupPolicies {
		err := db.Create(gp).Error
		require.NoError(t, err)
	}

	// 测试获取重试策略
	retryPolicies, err := engine.GetGroupPoliciesByType(group.ID, models.PolicyTypeRetry)
	require.NoError(t, err)
	assert.Len(t, retryPolicies, 1)
	assert.Equal(t, models.PolicyTypeRetry, retryPolicies[0].Policy.Type)

	// 测试获取模型过滤策略
	filterPolicies, err := engine.GetGroupPoliciesByType(group.ID, models.PolicyTypeModelFilter)
	require.NoError(t, err)
	assert.Len(t, filterPolicies, 1)
	assert.Equal(t, models.PolicyTypeModelFilter, filterPolicies[0].Policy.Type)

	// 测试获取不存在的策略类型
	rateLimitPolicies, err := engine.GetGroupPoliciesByType(group.ID, models.PolicyTypeRateLimit)
	require.NoError(t, err)
	assert.Len(t, rateLimitPolicies, 0)
}
