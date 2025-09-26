package services

import (
	"fmt"
	"gpt-load/internal/models"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// KeyStateService 负责管理 API Key 的状态转换和持久化。
type KeyStateService struct {
	db           *gorm.DB
	stateMachine *models.KeyStateMachine
}

// NewKeyStateService 创建一个新的 KeyStateService 实例。
func NewKeyStateService(db *gorm.DB) *KeyStateService {
	return &KeyStateService{
		db:           db,
		stateMachine: models.NewKeyStateMachine(),
	}
}

// HandleSuccess 处理密钥验证成功的情况
func (kss *KeyStateService) HandleSuccess(keyID uint) error {
	return kss.db.Transaction(func(tx *gorm.DB) error {
		var key models.APIKey
		if err := tx.First(&key, keyID).Error; err != nil {
			return fmt.Errorf("failed to find key %d: %w", keyID, err)
		}

		now := time.Now()
		newStatus := kss.stateMachine.TransitionState(key.Status, true, key.ConsecutiveFailures)

		updates := map[string]interface{}{
			"status":               newStatus,
			"last_success_at":      &now,
			"consecutive_failures": 0, // 成功后重置连续失败次数
			"backoff_level":        0, // 成功后重置退避级别
			"disabled_until":       nil, // 清除禁用时间
			"last_error_message":   "", // 清除错误信息
		}

		if err := tx.Model(&key).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update key %d status: %w", keyID, err)
		}

		logrus.WithFields(logrus.Fields{
			"keyID":     keyID,
			"oldStatus": key.Status,
			"newStatus": newStatus,
		}).Debug("Key status updated after success")

		return nil
	})
}

// HandleFailure 处理密钥验证失败的情况
func (kss *KeyStateService) HandleFailure(keyID uint, errorMessage string) error {
	return kss.db.Transaction(func(tx *gorm.DB) error {
		var key models.APIKey
		if err := tx.First(&key, keyID).Error; err != nil {
			return fmt.Errorf("failed to find key %d: %w", keyID, err)
		}

		now := time.Now()
		consecutiveFailures := key.ConsecutiveFailures + 1
		newStatus := kss.stateMachine.TransitionState(key.Status, false, consecutiveFailures)

		updates := map[string]interface{}{
			"status":               newStatus,
			"last_failure_at":      &now,
			"consecutive_failures": consecutiveFailures,
			"last_error_message":   errorMessage,
		}

		// 如果转换到禁用状态，设置禁用时间和退避级别
		if newStatus == models.KeyStatusDisabled {
			backoffLevel := key.BackoffLevel + 1
			disabledUntil := now.Add(kss.stateMachine.CalculateBackoffDuration(backoffLevel))
			updates["backoff_level"] = backoffLevel
			updates["disabled_until"] = &disabledUntil
		}

		if err := tx.Model(&key).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update key %d status: %w", keyID, err)
		}

		logrus.WithFields(logrus.Fields{
			"keyID":              keyID,
			"oldStatus":          key.Status,
			"newStatus":          newStatus,
			"consecutiveFailures": consecutiveFailures,
			"errorMessage":       errorMessage,
		}).Debug("Key status updated after failure")

		return nil
	})
}

// ManuallyInvalidateKey 手动将密钥标记为无效
func (kss *KeyStateService) ManuallyInvalidateKey(keyID uint, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":             models.KeyStatusInvalid,
		"last_error_message": fmt.Sprintf("Manually invalidated: %s", reason),
		"last_failure_at":    &now,
	}

	if err := kss.db.Model(&models.APIKey{}).Where("id = ?", keyID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to invalidate key %d: %w", keyID, err)
	}

	logrus.WithFields(logrus.Fields{
		"keyID":  keyID,
		"reason": reason,
	}).Info("Key manually invalidated")

	return nil
}

// ManuallyDisableKey 手动禁用密钥
func (kss *KeyStateService) ManuallyDisableKey(keyID uint, reason string) error {
	now := time.Now()
	disabledUntil := now.Add(30 * time.Minute) // 手动禁用30分钟

	updates := map[string]interface{}{
		"status":             models.KeyStatusDisabled,
		"disabled_until":     &disabledUntil,
		"last_error_message": fmt.Sprintf("Manually disabled: %s", reason),
		"last_failure_at":    &now,
	}

	if err := kss.db.Model(&models.APIKey{}).Where("id = ?", keyID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to disable key %d: %w", keyID, err)
	}

	logrus.WithFields(logrus.Fields{
		"keyID":        keyID,
		"reason":       reason,
		"disabledUntil": disabledUntil,
	}).Info("Key manually disabled")

	return nil
}

// ManuallyEnableKey 手动启用密钥（设置为降级状态）
func (kss *KeyStateService) ManuallyEnableKey(keyID uint) error {
	updates := map[string]interface{}{
		"status":             models.KeyStatusDegraded, // 手动启用后设为降级状态，需要验证后才能变为活跃
		"disabled_until":     nil,
		"consecutive_failures": 0,
		"backoff_level":      0,
		"last_error_message": "",
	}

	if err := kss.db.Model(&models.APIKey{}).Where("id = ?", keyID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to enable key %d: %w", keyID, err)
	}

	logrus.WithFields(logrus.Fields{
		"keyID": keyID,
	}).Info("Key manually enabled")

	return nil
}

// UpdateKeyStatus 直接更新密钥状态（用于迁移等场景）
func (kss *KeyStateService) UpdateKeyStatus(keyID uint, newStatus string) error {
	if !kss.stateMachine.IsValidState(newStatus) {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	if err := kss.db.Model(&models.APIKey{}).Where("id = ?", keyID).Update("status", newStatus).Error; err != nil {
		return fmt.Errorf("failed to update key %d status to %s: %w", keyID, newStatus, err)
	}

	return nil
}
