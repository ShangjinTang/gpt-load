package services

import (
	"gpt-load/internal/interfaces"
	"gpt-load/internal/models"

	"github.com/sirupsen/logrus"
)

// KeyPolicyHandler 处理密钥相关的策略评估和状态更新
type KeyPolicyHandler struct {
	policyEngine    interfaces.PolicyEngineInterface
	keyStateService interfaces.KeyStateServiceInterface
}

// NewKeyPolicyHandler 创建一个新的 KeyPolicyHandler 实例
func NewKeyPolicyHandler(
	policyEngine interfaces.PolicyEngineInterface,
	keyStateService interfaces.KeyStateServiceInterface,
) *KeyPolicyHandler {
	return &KeyPolicyHandler{
		policyEngine:    policyEngine,
		keyStateService: keyStateService,
	}
}

// HandleKeySuccess 处理密钥成功的情况
func (kph *KeyPolicyHandler) HandleKeySuccess(keyID uint) {
	if err := kph.keyStateService.HandleSuccess(keyID); err != nil {
		logrus.WithFields(logrus.Fields{"keyID": keyID, "error": err}).Error("Failed to handle key success")
	}
}

// HandleKeyFailure 处理密钥失败的情况，包括策略评估
func (kph *KeyPolicyHandler) HandleKeyFailure(apiKey *models.APIKey, group *models.Group, errorMessage string) {
	// 创建策略评估上下文
	context := &models.PolicyEvaluationContext{
		ErrorMessage: errorMessage,
		KeyID:        apiKey.ID,
		GroupID:      group.ID,
		FailureCount: apiKey.FailureCount,
		RequestCount: apiKey.RequestCount,
	}

	// 评估重试策略
	retryResult, err := kph.policyEngine.EvaluateRetryPolicies(group.ID, context)
	if err != nil {
		logrus.WithFields(logrus.Fields{"keyID": apiKey.ID, "error": err}).Error("Failed to evaluate retry policies")
	}

	// 根据策略结果处理
	if retryResult != nil && retryResult.Matched {
		logrus.WithFields(logrus.Fields{
			"keyID":  apiKey.ID,
			"policy": retryResult.PolicyName,
			"rule":   retryResult.RuleName,
			"action": retryResult.Action,
			"reason": retryResult.Reason,
		}).Info("Policy-based failure handling")

		// 根据策略动作处理
		switch retryResult.Action {
		case models.RetryActionInvalidate:
			if err := kph.keyStateService.ManuallyInvalidateKey(apiKey.ID, retryResult.Reason); err != nil {
				logrus.WithError(err).Error("Failed to invalidate key based on policy")
			}
		case models.RetryActionDisable:
			if err := kph.keyStateService.ManuallyDisableKey(apiKey.ID, retryResult.Reason); err != nil {
				logrus.WithError(err).Error("Failed to disable key based on policy")
			}
		default:
			// 对于 retry 和 degrade 动作，使用默认的失败处理
			if err := kph.keyStateService.HandleFailure(apiKey.ID, errorMessage); err != nil {
				logrus.WithFields(logrus.Fields{"keyID": apiKey.ID, "error": err}).Error("Failed to handle key failure")
			}
		}
	} else {
		// 没有匹配的策略，使用默认处理
		if err := kph.keyStateService.HandleFailure(apiKey.ID, errorMessage); err != nil {
			logrus.WithFields(logrus.Fields{"keyID": apiKey.ID, "error": err}).Error("Failed to handle key failure")
		}
	}
}
