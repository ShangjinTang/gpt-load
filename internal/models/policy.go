package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
)

// PolicyType 策略类型
const (
	PolicyTypeRetry       = "retry"        // 重试策略
	PolicyTypeDegradation = "degradation"  // 降级策略
	PolicyTypeModelFilter = "model_filter" // 模型过滤策略
	PolicyTypeRateLimit   = "rate_limit"   // 限流策略
)

// PolicyAction 策略动作
const (
	RetryActionRetry     = "retry"     // 重试
	RetryActionDegrade   = "degrade"   // 降级
	RetryActionDisable   = "disable"   // 禁用
	RetryActionInvalidate = "invalidate" // 标记无效
)

// Policy 策略模型
type Policy struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null;unique" json:"name"`
	Description string         `gorm:"type:varchar(512)" json:"description"`
	Type        string         `gorm:"type:varchar(50);not null" json:"type"`
	Config      datatypes.JSON `gorm:"type:json;not null" json:"config"`
	Priority    int            `gorm:"not null;default:0" json:"priority"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// GroupPolicy 分组和策略的关联模型
type GroupPolicy struct {
	GroupID   uint      `gorm:"primaryKey" json:"group_id"`
	PolicyID  uint      `gorm:"primaryKey" json:"policy_id"`
	Policy    Policy    `gorm:"foreignKey:PolicyID" json:"policy"`
	Priority  int       `gorm:"not null;default:0" json:"priority"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RetryCondition 重试策略条件
type RetryCondition struct {
	Type     string   `json:"type"`     // "status_code", "error_message", "error_type", "model", "failure_count", "request_count"
	Operator string   `json:"operator"` // "equals", "contains", "regex", "in", "not_in", "gt", "lt", "gte", "lte"
	Value    string   `json:"value,omitempty"`
	Values   []string `json:"values,omitempty"`
}

// RetryRule 重试策略规则
type RetryRule struct {
	Name       string           `json:"name"`
	Priority   int              `json:"priority"`
	Conditions []RetryCondition `json:"conditions"`
	Action     string           `json:"action"` // "retry", "degrade", "disable", "invalidate"
	MaxRetries int              `json:"max_retries,omitempty"`
	BackoffMs  int              `json:"backoff_ms,omitempty"`
}

// RetryPolicyConfig 重试策略配置
type RetryPolicyConfig struct {
	Rules []RetryRule `json:"rules"`
}

// DegradationRule 降级策略规则
type DegradationRule struct {
	Name       string           `json:"name"`
	Priority   int              `json:"priority"`
	Conditions []RetryCondition `json:"conditions"`
	Action     string           `json:"action"`   // "disable", "invalidate"
	Duration   string           `json:"duration"` // e.g., "5m", "1h"
}

// DegradationPolicyConfig 降级策略配置
type DegradationPolicyConfig struct {
	Rules []DegradationRule `json:"rules"`
}

// ModelFilterPolicyConfig 模型过滤策略配置
type ModelFilterPolicyConfig struct {
	Type     string   `json:"type"`
	Patterns []string `json:"patterns"`
}

// RateLimitPolicyConfig 限流策略配置
type RateLimitPolicyConfig struct {
	Limit    int64  `json:"limit"`
	Interval string `json:"interval"` // e.g., "1s", "1m", "1h"
}

// PolicyEvaluationContext 策略评估上下文
type PolicyEvaluationContext struct {
	GroupID       uint
	KeyID         uint
	Model         string
	StatusCode    int
	ErrorMessage  string
	ErrorType     string
	FailureCount  int64
	RequestCount  int64
}

// PolicyEvaluationResult 策略评估结果
type PolicyEvaluationResult struct {
	PolicyID   uint
	PolicyName string
	RuleName   string
	Action     string
	MaxRetries int
	BackoffMs  int
	Duration   string
	Priority   int
	Reason     string
	Matched    bool
}

// GetRetryConfig 从 Policy.Config 中解析重试策略配置
func (p *Policy) GetRetryConfig() (*RetryPolicyConfig, error) {
	if p.Type != PolicyTypeRetry {
		return nil, fmt.Errorf("policy is not of type 'retry'")
	}
	var config RetryPolicyConfig
	err := json.Unmarshal(p.Config, &config)
	return &config, err
}

// GetDegradationConfig ...
func (p *Policy) GetDegradationConfig() (*DegradationPolicyConfig, error) {
	if p.Type != PolicyTypeDegradation {
		return nil, fmt.Errorf("policy is not of type 'degradation'")
	}
	var config DegradationPolicyConfig
	err := json.Unmarshal(p.Config, &config)
	return &config, err
}

// GetModelFilterConfig ...
func (p *Policy) GetModelFilterConfig() (*ModelFilterPolicyConfig, error) {
	if p.Type != PolicyTypeModelFilter {
		return nil, fmt.Errorf("policy is not of type 'model_filter'")
	}
	var config ModelFilterPolicyConfig
	err := json.Unmarshal(p.Config, &config)
	return &config, err
}

// GetRateLimitConfig ...
func (p *Policy) GetRateLimitConfig() (*RateLimitPolicyConfig, error) {
	if p.Type != PolicyTypeRateLimit {
		return nil, fmt.Errorf("policy is not of type 'rate_limit'")
	}
	var config RateLimitPolicyConfig
	err := json.Unmarshal(p.Config, &config)
	return &config, err
}
