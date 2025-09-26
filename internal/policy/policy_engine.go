package policy

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gpt-load/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PolicyEngine 策略引擎
type PolicyEngine struct {
	db *gorm.DB
}

// NewPolicyEngine 创建新的策略引擎
func NewPolicyEngine(db *gorm.DB) *PolicyEngine {
	return &PolicyEngine{
		db: db,
	}
}

// EvaluatePolicies 评估分组的所有策略
func (pe *PolicyEngine) EvaluatePolicies(groupID uint, context *models.PolicyEvaluationContext) ([]*models.PolicyEvaluationResult, error) {
	// 获取分组关联的所有活跃策略
	policies, err := pe.GetGroupPolicies(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group policies: %w", err)
	}

	var results []*models.PolicyEvaluationResult

	// 按优先级排序策略
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})

	// 评估每个策略
	for _, groupPolicy := range policies {
		if !groupPolicy.IsActive || !groupPolicy.Policy.IsActive {
			continue
		}

		result, err := pe.evaluatePolicy(&groupPolicy.Policy, context)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"policy_id":   groupPolicy.PolicyID,
				"policy_name": groupPolicy.Policy.Name,
				"group_id":    groupID,
				"error":       err,
			}).Error("Failed to evaluate policy")
			continue
		}

		if result != nil {
			result.Priority = groupPolicy.Priority
			results = append(results, result)
		}
	}

	return results, nil
}

// EvaluateRetryPolicies 专门评估重试策略
func (pe *PolicyEngine) EvaluateRetryPolicies(groupID uint, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	policies, err := pe.GetGroupPoliciesByType(groupID, models.PolicyTypeRetry)
	if err != nil {
		return nil, fmt.Errorf("failed to get retry policies: %w", err)
	}

	// 按优先级排序
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})

	// 返回第一个匹配的策略结果
	for _, groupPolicy := range policies {
		if !groupPolicy.IsActive || !groupPolicy.Policy.IsActive {
			continue
		}

		result, err := pe.evaluateRetryPolicy(&groupPolicy.Policy, context)
		if err != nil {
			logrus.WithError(err).Error("Failed to evaluate retry policy")
			continue
		}

		if result != nil && result.Matched {
			result.Priority = groupPolicy.Priority
			return result, nil
		}
	}

	return nil, nil
}

// EvaluateDegradationPolicies 评估降级策略
func (pe *PolicyEngine) EvaluateDegradationPolicies(groupID uint, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	policies, err := pe.GetGroupPoliciesByType(groupID, models.PolicyTypeDegradation)
	if err != nil {
		return nil, fmt.Errorf("failed to get degradation policies: %w", err)
	}

	// 按优先级排序
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})

	// 返回第一个匹配的策略结果
	for _, groupPolicy := range policies {
		if !groupPolicy.IsActive || !groupPolicy.Policy.IsActive {
			continue
		}

		result, err := pe.evaluateDegradationPolicy(&groupPolicy.Policy, context)
		if err != nil {
			logrus.WithError(err).Error("Failed to evaluate degradation policy")
			continue
		}

		if result != nil && result.Matched {
			result.Priority = groupPolicy.Priority
			return result, nil
		}
	}

	return nil, nil
}

// EvaluateModelFilterPolicies 评估模型过滤策略
func (pe *PolicyEngine) EvaluateModelFilterPolicies(groupID uint, model string) (bool, error) {
	policies, err := pe.GetGroupPoliciesByType(groupID, models.PolicyTypeModelFilter)
	if err != nil {
		return true, fmt.Errorf("failed to get model filter policies: %w", err)
	}

	// 按优先级排序
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})

	// 评估每个模型过滤策略
	for _, groupPolicy := range policies {
		if !groupPolicy.IsActive || !groupPolicy.Policy.IsActive {
			continue
		}

		allowed, err := pe.evaluateModelFilterPolicy(&groupPolicy.Policy, model)
		if err != nil {
			logrus.WithError(err).Error("Failed to evaluate model filter policy")
			continue
		}

		// 如果任何一个策略拒绝，则拒绝
		if !allowed {
			return false, nil
		}
	}

	return true, nil
}

// GetGroupPolicies 获取分组的所有策略
func (pe *PolicyEngine) GetGroupPolicies(groupID uint) ([]models.GroupPolicy, error) {
	var groupPolicies []models.GroupPolicy

	if err := pe.db.Preload("Policy").Where("group_id = ? AND is_active = ?", groupID, true).Find(&groupPolicies).Error; err != nil {
		return nil, fmt.Errorf("failed to query group policies: %w", err)
	}

	return groupPolicies, nil
}

// GetGroupPoliciesByType 获取分组指定类型的策略
func (pe *PolicyEngine) GetGroupPoliciesByType(groupID uint, policyType string) ([]models.GroupPolicy, error) {
	var groupPolicies []models.GroupPolicy

	if err := pe.db.Preload("Policy").
		Joins("JOIN policies ON policies.id = group_policies.policy_id").
		Where("group_policies.group_id = ? AND group_policies.is_active = ? AND policies.type = ? AND policies.is_active = ?",
			groupID, true, policyType, true).
		Find(&groupPolicies).Error; err != nil {
		return nil, fmt.Errorf("failed to query group policies by type: %w", err)
	}

	return groupPolicies, nil
}

// evaluatePolicy 评估单个策略
func (pe *PolicyEngine) evaluatePolicy(policy *models.Policy, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	switch policy.Type {
	case models.PolicyTypeRetry:
		return pe.evaluateRetryPolicy(policy, context)
	case models.PolicyTypeDegradation:
		return pe.evaluateDegradationPolicy(policy, context)
	case models.PolicyTypeModelFilter:
		// 模型过滤策略需要单独调用
		return nil, nil
	case models.PolicyTypeRateLimit:
		return pe.evaluateRateLimitPolicy(policy, context)
	default:
		return nil, fmt.Errorf("unknown policy type: %s", policy.Type)
	}
}

// evaluateRetryPolicy 评估重试策略
func (pe *PolicyEngine) evaluateRetryPolicy(policy *models.Policy, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	config, err := policy.GetRetryConfig()
	if err != nil {
		return nil, err
	}

	// 按优先级排序规则
	sort.Slice(config.Rules, func(i, j int) bool {
		return config.Rules[i].Priority < config.Rules[j].Priority
	})

	// 评估每个规则
	for _, rule := range config.Rules {
		if pe.evaluateConditions(rule.Conditions, context) {
			return &models.PolicyEvaluationResult{
				PolicyID:   policy.ID,
				PolicyName: policy.Name,
				RuleName:   rule.Name,
				Action:     rule.Action,
				MaxRetries: rule.MaxRetries,
				BackoffMs:  rule.BackoffMs,
				Priority:   rule.Priority,
				Reason:     fmt.Sprintf("Matched rule: %s", rule.Name),
				Matched:    true,
			}, nil
		}
	}

	return &models.PolicyEvaluationResult{
		PolicyID:   policy.ID,
		PolicyName: policy.Name,
		Matched:    false,
	}, nil
}

// evaluateDegradationPolicy 评估降级策略
func (pe *PolicyEngine) evaluateDegradationPolicy(policy *models.Policy, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	config, err := policy.GetDegradationConfig()
	if err != nil {
		return nil, err
	}

	// 按优先级排序规则
	sort.Slice(config.Rules, func(i, j int) bool {
		return config.Rules[i].Priority < config.Rules[j].Priority
	})

	// 评估每个规则
	for _, rule := range config.Rules {
		if pe.evaluateConditions(rule.Conditions, context) {
			return &models.PolicyEvaluationResult{
				PolicyID:   policy.ID,
				PolicyName: policy.Name,
				RuleName:   rule.Name,
				Action:     rule.Action,
				Duration:   rule.Duration,
				Priority:   rule.Priority,
				Reason:     fmt.Sprintf("Matched degradation rule: %s", rule.Name),
				Matched:    true,
			}, nil
		}
	}

	return &models.PolicyEvaluationResult{
		PolicyID:   policy.ID,
		PolicyName: policy.Name,
		Matched:    false,
	}, nil
}

// evaluateModelFilterPolicy 评估模型过滤策略
func (pe *PolicyEngine) evaluateModelFilterPolicy(policy *models.Policy, model string) (bool, error) {
	config, err := policy.GetModelFilterConfig()
	if err != nil {
		return true, err
	}

	// 检查每个模式
	for _, pattern := range config.Patterns {
		matched, err := regexp.MatchString(pattern, model)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"policy": policy.Name,
				"pattern": pattern,
				"model": model,
				"error": err,
			}).Warn("Invalid regex pattern in model filter policy")
			continue
		}

		if matched {
			// 如果是包含模式且匹配，则允许
			// 如果是排除模式且匹配，则拒绝
			return config.Type == "include", nil
		}
	}

	// 如果没有匹配：
	// - 包含模式：默认拒绝
	// - 排除模式：默认允许
	return config.Type != "include", nil
}

// evaluateRateLimitPolicy 评估限流策略
func (pe *PolicyEngine) evaluateRateLimitPolicy(policy *models.Policy, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error) {
	_, err := policy.GetRateLimitConfig()
	if err != nil {
		return nil, err
	}

	// 这里简化处理，实际实现需要配合 Redis 等存储来跟踪请求计数
	// 目前返回不匹配
	return &models.PolicyEvaluationResult{
		PolicyID:   policy.ID,
		PolicyName: policy.Name,
		Matched:    false,
	}, nil
}

// evaluateConditions 评估条件列表
func (pe *PolicyEngine) evaluateConditions(conditions []models.RetryCondition, context *models.PolicyEvaluationContext) bool {
	if len(conditions) == 0 {
		return true
	}

	// 所有条件都必须满足（AND 逻辑）
	for _, condition := range conditions {
		if !pe.evaluateCondition(condition, context) {
			return false
		}
	}

	return true
}

// evaluateCondition 评估单个条件
func (pe *PolicyEngine) evaluateCondition(condition models.RetryCondition, context *models.PolicyEvaluationContext) bool {
	var actualValue string

	// 获取实际值
	switch condition.Type {
	case "status_code":
		actualValue = strconv.Itoa(context.StatusCode)
	case "error_message":
		actualValue = context.ErrorMessage
	case "error_type":
		actualValue = context.ErrorType
	case "model":
		actualValue = context.Model
	case "failure_count":
		actualValue = strconv.FormatInt(context.FailureCount, 10)
	case "request_count":
		actualValue = strconv.FormatInt(context.RequestCount, 10)
	default:
		return false
	}

	// 根据操作符进行比较
	switch condition.Operator {
	case "equals":
		return actualValue == condition.Value
	case "contains":
		return strings.Contains(actualValue, condition.Value)
	case "regex":
		matched, err := regexp.MatchString(condition.Value, actualValue)
		if err != nil {
			logrus.WithError(err).Warn("Invalid regex in policy condition")
			return false
		}
		return matched
	case "in":
		for _, value := range condition.Values {
			if actualValue == value {
				return true
			}
		}
		return false
	case "not_in":
		for _, value := range condition.Values {
			if actualValue == value {
				return false
			}
		}
		return true
	case "gt":
		actual, err1 := strconv.ParseFloat(actualValue, 64)
		expected, err2 := strconv.ParseFloat(condition.Value, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return actual > expected
	case "lt":
		actual, err1 := strconv.ParseFloat(actualValue, 64)
		expected, err2 := strconv.ParseFloat(condition.Value, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return actual < expected
	case "gte":
		actual, err1 := strconv.ParseFloat(actualValue, 64)
		expected, err2 := strconv.ParseFloat(condition.Value, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return actual >= expected
	case "lte":
		actual, err1 := strconv.ParseFloat(actualValue, 64)
		expected, err2 := strconv.ParseFloat(condition.Value, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return actual <= expected
	default:
		return false
	}
}
