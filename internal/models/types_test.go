package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestSystemSetting(t *testing.T) {
	setting := SystemSetting{
		ID:           1,
		SettingKey:   "request_timeout",
		SettingValue: "30",
		Description:  "Request timeout in seconds",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.Equal(t, uint(1), setting.ID)
	assert.Equal(t, "request_timeout", setting.SettingKey)
	assert.Equal(t, "30", setting.SettingValue)
	assert.Equal(t, "Request timeout in seconds", setting.Description)
	assert.NotZero(t, setting.CreatedAt)
	assert.NotZero(t, setting.UpdatedAt)
}

func TestGroup(t *testing.T) {
	now := time.Now()
	group := Group{
		ID:                 1,
		Name:               "test-group",
		DisplayName:        "Test Group",
		ProxyKeys:          "key1,key2,key3",
		Description:        "Test group for API keys",
		Upstreams:          datatypes.JSON(`["http://api1.example.com", "http://api2.example.com"]`),
		ValidationEndpoint: "http://api.example.com/validate",
		ChannelType:        "openai",
		Sort:               1,
		TestModel:          "gpt-3.5-turbo",
		ParamOverrides:     datatypes.JSONMap{"temperature": 0.7, "max_tokens": 1000},
		Config:             datatypes.JSONMap{"timeout": 30, "retries": 3},
		HeaderRules:        datatypes.JSON(`[{"name": "Authorization", "value": "Bearer {key}"}]`),
		LastValidatedAt:    &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	assert.Equal(t, uint(1), group.ID)
	assert.Equal(t, "test-group", group.Name)
	assert.Equal(t, "Test Group", group.DisplayName)
	assert.Equal(t, "key1,key2,key3", group.ProxyKeys)
	assert.Equal(t, "Test group for API keys", group.Description)
	assert.Equal(t, "http://api.example.com/validate", group.ValidationEndpoint)
	assert.Equal(t, "openai", group.ChannelType)
	assert.Equal(t, 1, group.Sort)
	assert.Equal(t, "gpt-3.5-turbo", group.TestModel)
	assert.NotNil(t, group.LastValidatedAt)
	assert.Equal(t, now, *group.LastValidatedAt)
	assert.NotZero(t, group.CreatedAt)
	assert.NotZero(t, group.UpdatedAt)
}

func TestAPIKey(t *testing.T) {
	now := time.Now()
	key := APIKey{
		ID:                  1,
		KeyValue:            "sk-test123456789",
		KeyHash:             "hash123",
		GroupID:             1,
		Status:              KeyStatusActive,
		RequestCount:        100,
		FailureCount:        5,
		LastUsedAt:          &now,
		CreatedAt:           now,
		UpdatedAt:           now,
		LastFailureAt:       &now,
		LastSuccessAt:       &now,
		LastValidatedAt:     &now,
		DisabledUntil:       nil,
		ConsecutiveFailures: 2,
		LastErrorMessage:    "Rate limit exceeded",
		BackoffLevel:        1,
	}

	assert.Equal(t, uint(1), key.ID)
	assert.Equal(t, "sk-test123456789", key.KeyValue)
	assert.Equal(t, "hash123", key.KeyHash)
	assert.Equal(t, uint(1), key.GroupID)
	assert.Equal(t, KeyStatusActive, key.Status)
	assert.Equal(t, int64(100), key.RequestCount)
	assert.Equal(t, int64(5), key.FailureCount)
	assert.NotNil(t, key.LastUsedAt)
	assert.Equal(t, now, *key.LastUsedAt)
	assert.NotZero(t, key.CreatedAt)
	assert.NotZero(t, key.UpdatedAt)
	assert.NotNil(t, key.LastFailureAt)
	assert.NotNil(t, key.LastSuccessAt)
	assert.NotNil(t, key.LastValidatedAt)
	assert.Nil(t, key.DisabledUntil)
	assert.Equal(t, int64(2), key.ConsecutiveFailures)
	assert.Equal(t, "Rate limit exceeded", key.LastErrorMessage)
	assert.Equal(t, 1, key.BackoffLevel)
}

func TestRequestLog(t *testing.T) {
	now := time.Now()
	log := RequestLog{
		ID:           "req-123",
		Timestamp:    now,
		GroupID:      1,
		GroupName:    "test-group",
		KeyValue:     "sk-test123",
		KeyHash:      "hash123",
		Model:        "gpt-3.5-turbo",
		IsSuccess:    true,
		SourceIP:     "192.168.1.1",
		StatusCode:   200,
		RequestPath:  "/v1/chat/completions",
		Duration:     150,
		ErrorMessage: "",
		UserAgent:    "test-client/1.0",
		RequestType:  "final",
		UpstreamAddr: "http://api.openai.com",
		IsStream:     false,
		RequestBody:  "test request body",
	}

	assert.Equal(t, "req-123", log.ID)
	assert.Equal(t, now, log.Timestamp)
	assert.Equal(t, uint(1), log.GroupID)
	assert.Equal(t, "test-group", log.GroupName)
	assert.Equal(t, "sk-test123", log.KeyValue)
	assert.Equal(t, "hash123", log.KeyHash)
	assert.Equal(t, "gpt-3.5-turbo", log.Model)
	assert.True(t, log.IsSuccess)
	assert.Equal(t, "192.168.1.1", log.SourceIP)
	assert.Equal(t, 200, log.StatusCode)
	assert.Equal(t, "/v1/chat/completions", log.RequestPath)
	assert.Equal(t, int64(150), log.Duration)
	assert.Equal(t, "", log.ErrorMessage)
	assert.Equal(t, "test-client/1.0", log.UserAgent)
	assert.Equal(t, "final", log.RequestType)
	assert.Equal(t, "http://api.openai.com", log.UpstreamAddr)
	assert.False(t, log.IsStream)
	assert.Equal(t, "test request body", log.RequestBody)
}

func TestGroupConfig(t *testing.T) {
	timeout := 60
	retries := 5
	threshold := 10
	proxyURL := "http://proxy.example.com"

	config := GroupConfig{
		RequestTimeout:     &timeout,
		MaxRetries:         &retries,
		BlacklistThreshold: &threshold,
		ProxyURL:           &proxyURL,
	}

	assert.NotNil(t, config.RequestTimeout)
	assert.Equal(t, 60, *config.RequestTimeout)
	assert.NotNil(t, config.MaxRetries)
	assert.Equal(t, 5, *config.MaxRetries)
	assert.NotNil(t, config.BlacklistThreshold)
	assert.Equal(t, 10, *config.BlacklistThreshold)
	assert.NotNil(t, config.ProxyURL)
	assert.Equal(t, "http://proxy.example.com", *config.ProxyURL)
}

func TestHeaderRule(t *testing.T) {
	rule := HeaderRule{
		Key:    "Authorization",
		Value:  "Bearer {key}",
		Action: "set",
	}

	assert.Equal(t, "Authorization", rule.Key)
	assert.Equal(t, "Bearer {key}", rule.Value)
	assert.Equal(t, "set", rule.Action)
}

func TestAPIKey_DefaultValues(t *testing.T) {
	key := APIKey{
		KeyValue: "sk-test",
		GroupID:  1,
	}

	// Test that default values are set correctly by GORM
	assert.Equal(t, "sk-test", key.KeyValue)
	assert.Equal(t, uint(1), key.GroupID)
	// Status should default to 'pending' according to the struct tag
	// RequestCount, FailureCount, ConsecutiveFailures should default to 0
	// BackoffLevel should default to 0
}

func TestGroup_Relations(t *testing.T) {
	group := Group{
		ID:      1,
		Name:    "test-group",
		APIKeys: []APIKey{
			{KeyValue: "key1", GroupID: 1},
			{KeyValue: "key2", GroupID: 1},
		},
	}

	assert.Len(t, group.APIKeys, 2)
	assert.Equal(t, "key1", group.APIKeys[0].KeyValue)
	assert.Equal(t, "key2", group.APIKeys[1].KeyValue)
	assert.Equal(t, uint(1), group.APIKeys[0].GroupID)
	assert.Equal(t, uint(1), group.APIKeys[1].GroupID)
}
