package models

import (
	"math"
	"time"
)

// Key状态
const (
	KeyStatusPending  = "pending"  // 新增，待验证
	KeyStatusActive   = "active"   // 活跃，正常使用
	KeyStatusDegraded = "degraded" // 降级，部分失败，但仍可使用
	KeyStatusDisabled = "disabled" // 禁用，连续失败，暂时禁用（指数退避）
	KeyStatusInvalid  = "invalid"  // 无效，永久无效，需要手动干预
)

// KeyStateMachine 定义了API Key的状态机
type KeyStateMachine struct{}

// NewKeyStateMachine 创建一个新的状态机实例
func NewKeyStateMachine() *KeyStateMachine {
	return &KeyStateMachine{}
}

// GetValidStates 返回所有有效的状态
func (ksm *KeyStateMachine) GetValidStates() []string {
	return []string{
		KeyStatusPending,
		KeyStatusActive,
		KeyStatusDegraded,
		KeyStatusDisabled,
		KeyStatusInvalid,
	}
}

// IsValidState 检查状态是否有效
func (ksm *KeyStateMachine) IsValidState(status string) bool {
	validStates := ksm.GetValidStates()
	for _, validStatus := range validStates {
		if status == validStatus {
			return true
		}
	}
	return false
}

// GetStateDescription 获取状态描述
func (ksm *KeyStateMachine) GetStateDescription(status string) string {
	descriptions := map[string]string{
		KeyStatusPending:  "新添加的密钥，等待首次验证",
		KeyStatusActive:   "验证成功，正常工作状态",
		KeyStatusDegraded: "部分失败，但仍可使用",
		KeyStatusDisabled: "连续失败，暂时禁用（指数退避中）",
		KeyStatusInvalid:  "永久无效，需要手动干预",
	}

	if desc, exists := descriptions[status]; exists {
		return desc
	}
	return "未知状态"
}

// TransitionState 执行状态转换
func (ksm *KeyStateMachine) TransitionState(currentStatus string, isSuccess bool, consecutiveFailures int64) string {
	switch currentStatus {
	case KeyStatusPending:
		if isSuccess {
			return KeyStatusActive
		}
		// 首次验证失败，根据失败次数决定
		if consecutiveFailures >= 3 {
			return KeyStatusDisabled
		}
		return KeyStatusPending // 保持pending，等待重试

	case KeyStatusActive:
		if isSuccess {
			return KeyStatusActive // 保持活跃
		}
		// 从活跃状态失败，先降级
		return KeyStatusDegraded

	case KeyStatusDegraded:
		if isSuccess {
			return KeyStatusActive // 恢复到活跃状态
		}
		// 降级状态下继续失败，检查是否需要禁用
		if ksm.shouldDisable(consecutiveFailures) {
			return KeyStatusDisabled
		}
		return KeyStatusDegraded // 保持降级状态

	case KeyStatusDisabled:
		if isSuccess {
			return KeyStatusDegraded // 从禁用恢复到降级状态，而不是直接到活跃
		}
		return KeyStatusDisabled // 禁用状态下失败，保持禁用

	case KeyStatusInvalid:
		// 无效状态需要手动干预，不会自动转换
		return KeyStatusInvalid

	default:
		// 未知状态，默认设为pending
		return KeyStatusPending
	}
}

// shouldDegrade 判断是否应该降级
func (ksm *KeyStateMachine) shouldDegrade(consecutiveFailures int64) bool {
	return consecutiveFailures >= 1
}

// shouldDisable 判断是否应该禁用
func (ksm *KeyStateMachine) shouldDisable(consecutiveFailures int64) bool {
	return consecutiveFailures >= 3
}

// CalculateBackoffDuration 计算指数退避时间
func (ksm *KeyStateMachine) CalculateBackoffDuration(backoffLevel int) time.Duration {
	if backoffLevel <= 0 {
		return time.Minute // 最小1分钟
	}

	// 指数退避：2^level 分钟，最大30分钟
	minutes := math.Pow(2, float64(backoffLevel))
	if minutes > 30 {
		minutes = 30
	}

	return time.Duration(minutes) * time.Minute
}

// GetStateInfo 获取状态的详细信息
func (ksm *KeyStateMachine) GetStateInfo(key *APIKey) map[string]interface{} {
	info := map[string]interface{}{
		"status":               key.Status,
		"description":          ksm.GetStateDescription(key.Status),
		"consecutive_failures": key.ConsecutiveFailures,
		"last_error":          key.LastErrorMessage,
	}

	// 如果是禁用状态，计算重试时间
	if key.Status == KeyStatusDisabled && key.DisabledUntil != nil {
		info["disabled_until"] = key.DisabledUntil
		info["retry_in_seconds"] = int64(time.Until(*key.DisabledUntil).Seconds())
	}

	return info
}
