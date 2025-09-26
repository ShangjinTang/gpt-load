package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestPolicyEvaluationContext(t *testing.T) {
	context := &PolicyEvaluationContext{
		KeyID:        123,
		ErrorMessage: "Rate limit exceeded",
		StatusCode:   429,
		FailureCount: 3,
		GroupID:      1,
		Model:        "gpt-4",
		ErrorType:    "rate_limit",
		RequestCount: 100,
	}

	assert.Equal(t, uint(123), context.KeyID)
	assert.Equal(t, "Rate limit exceeded", context.ErrorMessage)
	assert.Equal(t, 429, context.StatusCode)
	assert.Equal(t, int64(3), context.FailureCount)
	assert.Equal(t, uint(1), context.GroupID)
	assert.Equal(t, "gpt-4", context.Model)
	assert.Equal(t, "rate_limit", context.ErrorType)
	assert.Equal(t, int64(100), context.RequestCount)
}

func TestPolicyEvaluationResult(t *testing.T) {
	result := &PolicyEvaluationResult{
		Action:     "disable",
		Reason:     "Too many failures",
		BackoffMs:  5000,
		PolicyID:   456,
		RuleName:   "failure-threshold",
		PolicyName: "retry-policy",
		MaxRetries: 3,
		Duration:   "5m",
		Priority:   1,
		Matched:    true,
	}

	assert.Equal(t, "disable", result.Action)
	assert.Equal(t, "Too many failures", result.Reason)
	assert.Equal(t, 5000, result.BackoffMs)
	assert.Equal(t, uint(456), result.PolicyID)
	assert.Equal(t, "failure-threshold", result.RuleName)
	assert.Equal(t, "retry-policy", result.PolicyName)
	assert.Equal(t, 3, result.MaxRetries)
	assert.Equal(t, "5m", result.Duration)
	assert.Equal(t, 1, result.Priority)
	assert.True(t, result.Matched)
}

func TestPolicy_GetRetryConfig(t *testing.T) {
	retryConfig := RetryPolicyConfig{
		Rules: []RetryRule{
			{
				Name:     "rate-limit-retry",
				Priority: 1,
				Conditions: []RetryCondition{
					{
						Type:     "status_code",
						Operator: "equals",
						Value:    "429",
					},
				},
				Action:     "retry",
				MaxRetries: 3,
				BackoffMs:  1000,
			},
		},
	}

	configJSON, err := json.Marshal(retryConfig)
	assert.NoError(t, err)

	policy := &Policy{
		Type:   PolicyTypeRetry,
		Config: datatypes.JSON(configJSON),
	}

	config, err := policy.GetRetryConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Rules, 1)
	assert.Equal(t, "rate-limit-retry", config.Rules[0].Name)
	assert.Equal(t, "retry", config.Rules[0].Action)
	assert.Equal(t, 3, config.Rules[0].MaxRetries)
}

func TestPolicy_GetRetryConfig_WrongType(t *testing.T) {
	policy := &Policy{
		Type: PolicyTypeDegradation,
	}

	config, err := policy.GetRetryConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "policy is not of type 'retry'")
}

func TestPolicy_GetDegradationConfig(t *testing.T) {
	degradationConfig := DegradationPolicyConfig{
		Rules: []DegradationRule{
			{
				Name:     "disable-on-failure",
				Priority: 1,
				Conditions: []RetryCondition{
					{
						Type:     "failure_count",
						Operator: "gte",
						Value:    "5",
					},
				},
				Action:   "disable",
				Duration: "10m",
			},
		},
	}

	configJSON, err := json.Marshal(degradationConfig)
	assert.NoError(t, err)

	policy := &Policy{
		Type:   PolicyTypeDegradation,
		Config: datatypes.JSON(configJSON),
	}

	config, err := policy.GetDegradationConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Rules, 1)
	assert.Equal(t, "disable-on-failure", config.Rules[0].Name)
	assert.Equal(t, "disable", config.Rules[0].Action)
	assert.Equal(t, "10m", config.Rules[0].Duration)
}

func TestPolicy_GetModelFilterConfig(t *testing.T) {
	filterConfig := ModelFilterPolicyConfig{
		Type:     "include",
		Patterns: []string{"gpt-4*", "gpt-3.5*"},
	}

	configJSON, err := json.Marshal(filterConfig)
	assert.NoError(t, err)

	policy := &Policy{
		Type:   PolicyTypeModelFilter,
		Config: datatypes.JSON(configJSON),
	}

	config, err := policy.GetModelFilterConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "include", config.Type)
	assert.Len(t, config.Patterns, 2)
	assert.Contains(t, config.Patterns, "gpt-4*")
	assert.Contains(t, config.Patterns, "gpt-3.5*")
}

func TestPolicy_GetRateLimitConfig(t *testing.T) {
	rateLimitConfig := RateLimitPolicyConfig{
		Limit:    100,
		Interval: "1m",
	}

	configJSON, err := json.Marshal(rateLimitConfig)
	assert.NoError(t, err)

	policy := &Policy{
		Type:   PolicyTypeRateLimit,
		Config: datatypes.JSON(configJSON),
	}

	config, err := policy.GetRateLimitConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, int64(100), config.Limit)
	assert.Equal(t, "1m", config.Interval)
}

func TestPolicy_FullStructure(t *testing.T) {
	policy := Policy{
		ID:          1,
		Name:        "retry-policy",
		Type:        PolicyTypeRetry,
		Description: "Policy for retrying failed requests",
		IsActive:    true,
		Priority:    10,
		Config:      datatypes.JSON(`{"rules":[]}`),
	}

	// Test basic fields
	assert.Equal(t, uint(1), policy.ID)
	assert.Equal(t, "retry-policy", policy.Name)
	assert.Equal(t, PolicyTypeRetry, policy.Type)
	assert.Equal(t, "Policy for retrying failed requests", policy.Description)
	assert.True(t, policy.IsActive)
	assert.Equal(t, 10, policy.Priority)
}

func TestRetryCondition(t *testing.T) {
	condition := RetryCondition{
		Type:     "status_code",
		Operator: "equals",
		Value:    "429",
		Values:   []string{"429", "503"},
	}

	assert.Equal(t, "status_code", condition.Type)
	assert.Equal(t, "equals", condition.Operator)
	assert.Equal(t, "429", condition.Value)
	assert.Len(t, condition.Values, 2)
	assert.Contains(t, condition.Values, "429")
	assert.Contains(t, condition.Values, "503")
}

func TestRetryRule(t *testing.T) {
	rule := RetryRule{
		Name:       "rate-limit-rule",
		Priority:   1,
		Action:     RetryActionRetry,
		MaxRetries: 3,
		BackoffMs:  1000,
		Conditions: []RetryCondition{
			{
				Type:     "status_code",
				Operator: "equals",
				Value:    "429",
			},
		},
	}

	assert.Equal(t, "rate-limit-rule", rule.Name)
	assert.Equal(t, 1, rule.Priority)
	assert.Equal(t, RetryActionRetry, rule.Action)
	assert.Equal(t, 3, rule.MaxRetries)
	assert.Equal(t, 1000, rule.BackoffMs)
	assert.Len(t, rule.Conditions, 1)
}
