package interfaces

import (
	"context"
	"time"

	"gpt-load/internal/models"
)

// KeyProviderInterface 定义密钥提供者接口
type KeyProviderInterface interface {
	AddKeys(groupID uint, keys []models.APIKey) error
	RestoreMultipleKeys(groupID uint, keyValues []string) (int64, error)
	RestoreKeys(groupID uint) (int64, error)
	RemoveInvalidKeys(groupID uint) (int64, error)
	RemoveAllKeys(groupID uint) (int64, error)
	RemoveKeys(groupID uint, keyValues []string) (int64, error)
	RemoveKeysFromStore(groupID uint, keyIDs []uint) error
}

// KeyValidationResult 定义了单个密钥验证的结果。
type KeyValidationResult struct {
	IsValid bool
	Error   error
}

// KeyStatusUpdater 定义密钥状态更新接口
type KeyStatusUpdater interface {
	UpdateStatus(apiKey *models.APIKey, group *models.Group, isSuccess bool, errorMessage string)
}

// KeyValidatorInterface 定义密钥验证器接口
type KeyValidatorInterface interface {
	ValidateGroup(group *models.Group) (map[uint]interface{}, error)
	ValidateSingleKey(key *models.APIKey, group *models.Group) (bool, error)
	TestMultipleKeys(group *models.Group, keyValues []string) ([]interface{}, error)
}

// PolicyEngineInterface 定义策略引擎接口
type PolicyEngineInterface interface {
	EvaluateRetryPolicies(groupID uint, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error)
	EvaluateDegradationPolicies(groupID uint, context *models.PolicyEvaluationContext) (*models.PolicyEvaluationResult, error)
}

// KeyStateServiceInterface 定义密钥状态服务接口
type KeyStateServiceInterface interface {
	HandleSuccess(keyID uint) error
	HandleFailure(keyID uint, errorMessage string) error
	ManuallyInvalidateKey(keyID uint, reason string) error
	ManuallyDisableKey(keyID uint, reason string) error
	ManuallyEnableKey(keyID uint) error
}

// KeyPolicyHandlerInterface 定义密钥策略处理器接口
type KeyPolicyHandlerInterface interface {
	HandleKeySuccess(keyID uint)
	HandleKeyFailure(apiKey *models.APIKey, group *models.Group, errorMessage string)
}

// IncrementalValidationConfig 增量验证配置接口
type IncrementalValidationConfig interface {
	GetTimeWindow() time.Duration
	GetIncludeStates() []string
	GetExcludeRecentlyValidated() bool
	GetRecentValidationWindow() time.Duration
	GetConcurrency() int
	GetBatchSize() int
}

// IncrementalValidationResult 增量验证结果接口
type IncrementalValidationResult interface {
	GetGroupID() uint
	GetGroupName() string
	GetTotalKeys() int
	GetValidatedKeys() int
	GetSkippedKeys() int
	GetSuccessfulKeys() int
	GetFailedKeys() int
	GetDuration() time.Duration
}

// IncrementalValidationServiceInterface 定义增量验证服务接口
type IncrementalValidationServiceInterface interface {
	ValidateGroup(ctx context.Context, groupID uint, config IncrementalValidationConfig) (IncrementalValidationResult, error)
	ValidateAllGroups(ctx context.Context, config IncrementalValidationConfig) ([]IncrementalValidationResult, error)
	GetValidationHistory(groupID uint, timeRange time.Duration) (map[string]interface{}, error)
}
