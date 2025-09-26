package policy

import (
	"encoding/json"
	"gpt-load/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PolicyService 负责策略的 CRUD 操作和与分组的关联。
type PolicyService struct {
	db *gorm.DB
}

// NewPolicyService 创建一个新的 PolicyService 实例。
func NewPolicyService(db *gorm.DB) *PolicyService {
	return &PolicyService{db: db}
}

// CreatePolicy 创建新策略
func (ps *PolicyService) CreatePolicy(policy *models.Policy) error {
	return ps.db.Create(policy).Error
}

// UpdatePolicy 更新策略
func (ps *PolicyService) UpdatePolicy(policy *models.Policy) error {
	return ps.db.Save(policy).Error
}

// DeletePolicy 删除策略
func (ps *PolicyService) DeletePolicy(policyID uint) error {
	// 同时删除分组关联
	err := ps.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("policy_id = ?", policyID).Delete(&models.GroupPolicy{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Policy{}, policyID).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

// GetPolicyByID 根据 ID 获取策略
func (ps *PolicyService) GetPolicyByID(policyID uint) (*models.Policy, error) {
	var policy models.Policy
	err := ps.db.First(&policy, policyID).Error
	return &policy, err
}

// GetPoliciesByType 根据类型获取策略
func (ps *PolicyService) GetPoliciesByType(policyType string) ([]models.Policy, error) {
	var policies []models.Policy
	err := ps.db.Where("type = ?", policyType).Find(&policies).Error
	return policies, err
}

// AddPolicyToGroup 将策略添加到分组
func (ps *PolicyService) AddPolicyToGroup(groupID, policyID uint, priority int) error {
	assoc := &models.GroupPolicy{
		GroupID:  groupID,
		PolicyID: policyID,
		Priority: priority,
		IsActive: true,
	}
	return ps.db.Create(assoc).Error
}

// RemovePolicyFromGroup 从分组移除策略
func (ps *PolicyService) RemovePolicyFromGroup(groupID, policyID uint) error {
	return ps.db.Where("group_id = ? AND policy_id = ?", groupID, policyID).Delete(&models.GroupPolicy{}).Error
}

// CreateDefaultPolicies 创建默认策略
func (ps *PolicyService) CreateDefaultPolicies() error {
	defaultPolicies := []models.Policy{
		{
			Name:        "Default Retry Policy",
			Description: "A default policy for retrying common transient errors.",
			Type:        models.PolicyTypeRetry,
			IsActive:    true,
		},
		{
			Name:        "Default Degradation Policy",
			Description: "A default policy for handling critical errors that should cause a key to be temporarily disabled.",
			Type:        models.PolicyTypeDegradation,
			IsActive:    true,
		},
	}

	retryConfig := &models.RetryPolicyConfig{
		Rules: []models.RetryRule{
			{
				Name:     "Rate Limit Retry",
				Priority: 1,
				Conditions: []models.RetryCondition{{
					Type:     "error_message",
					Operator: "contains",
					Value:    "rate limit",
				}},
				Action:     models.RetryActionRetry,
				MaxRetries: 3,
				BackoffMs:  2000,
			},
			{
				Name:     "Handle Invalid Key Errors",
				Priority: 2,
				Conditions: []models.RetryCondition{{
					Type:     "error_message",
					Operator: "contains",
					Value:    "invalid api key",
				}},
				Action: models.RetryActionInvalidate,
			},
			{
				Name:     "Handle Insufficient Quota",
				Priority: 3,
				Conditions: []models.RetryCondition{{
					Type:     "error_message",
					Operator: "contains",
					Value:    "insufficient quota",
				}},
				Action: models.RetryActionDisable,
			},
		},
	}

	degradationConfig := &models.DegradationPolicyConfig{
		Rules: []models.DegradationRule{
			{
				Name:     "Disable on Critical Errors",
				Priority: 1,
				Conditions: []models.RetryCondition{{
					Type:     "status_code",
					Operator: "in",
					Values:   []string{"500", "502", "503", "504"},
				}},
				Action:   models.RetryActionDisable,
				Duration: "5m",
			},
			{
				Name:     "Invalidate on Auth Errors",
				Priority: 2,
				Conditions: []models.RetryCondition{{
					Type:     "status_code",
					Operator: "in",
					Values:   []string{"401", "403"},
				}},
				Action:   models.RetryActionInvalidate,
				Duration: "1h",
			},
		},
	}

	// Marshal configs to JSON
	retryConfigJSON, _ := json.Marshal(retryConfig)
	degradationConfigJSON, _ := json.Marshal(degradationConfig)

	defaultPolicies[0].Config = retryConfigJSON
	defaultPolicies[1].Config = degradationConfigJSON

	for _, policy := range defaultPolicies {
		err := ps.db.FirstOrCreate(&policy, models.Policy{Name: policy.Name}).Error
		if err != nil {
			logrus.WithError(err).Errorf("Failed to create or find default policy %s", policy.Name)
		}
	}

	return nil
}
