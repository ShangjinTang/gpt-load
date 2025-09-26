package services

import (
	"context"
	"fmt"
	"gpt-load/internal/interfaces"
	"gpt-load/internal/models"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// IncrementalValidationConfig 增量验证配置
type IncrementalValidationConfig struct {
	// TimeWindow 时间窗口，只验证在此时间内添加的密钥
	TimeWindow time.Duration `json:"time_window"`

	// IncludeStates 包含的状态，只验证这些状态的密钥
	IncludeStates []string `json:"include_states"`

	// ExcludeRecentlyValidated 排除最近验证过的密钥
	ExcludeRecentlyValidated bool `json:"exclude_recently_validated"`

	// RecentValidationWindow 最近验证的时间窗口
	RecentValidationWindow time.Duration `json:"recent_validation_window"`

	// Concurrency 并发验证数量
	Concurrency int `json:"concurrency"`

	// BatchSize 批处理大小
	BatchSize int `json:"batch_size"`
}

// DefaultIncrementalValidationConfig 返回默认配置
func DefaultIncrementalValidationConfig() *IncrementalValidationConfig {
	return &IncrementalValidationConfig{
		TimeWindow:                 24 * time.Hour, // 默认验证24小时内添加的密钥
		IncludeStates:              []string{models.KeyStatusPending, models.KeyStatusInvalid},
		ExcludeRecentlyValidated:   true,
		RecentValidationWindow:     1 * time.Hour, // 1小时内验证过的不再验证
		Concurrency:                5,
		BatchSize:                  100,
	}
}

// 实现接口方法
func (c *IncrementalValidationConfig) GetTimeWindow() time.Duration { return c.TimeWindow }
func (c *IncrementalValidationConfig) GetIncludeStates() []string { return c.IncludeStates }
func (c *IncrementalValidationConfig) GetExcludeRecentlyValidated() bool { return c.ExcludeRecentlyValidated }
func (c *IncrementalValidationConfig) GetRecentValidationWindow() time.Duration { return c.RecentValidationWindow }
func (c *IncrementalValidationConfig) GetConcurrency() int { return c.Concurrency }
func (c *IncrementalValidationConfig) GetBatchSize() int { return c.BatchSize }

// IncrementalValidationResult 增量验证结果
type IncrementalValidationResult struct {
	GroupID           uint                           `json:"group_id"`
	GroupName         string                         `json:"group_name"`
	TotalKeys         int                            `json:"total_keys"`
	ValidatedKeys     int                            `json:"validated_keys"`
	SkippedKeys       int                            `json:"skipped_keys"`
	SuccessfulKeys    int                            `json:"successful_keys"`
	FailedKeys        int                            `json:"failed_keys"`
	Duration          time.Duration                  `json:"duration"`
	KeyResults        []IncrementalKeyValidationResult `json:"key_results,omitempty"`
	StartTime         time.Time                      `json:"start_time"`
	EndTime           time.Time                      `json:"end_time"`
}

// 实现接口方法
func (r *IncrementalValidationResult) GetGroupID() uint { return r.GroupID }
func (r *IncrementalValidationResult) GetGroupName() string { return r.GroupName }
func (r *IncrementalValidationResult) GetTotalKeys() int { return r.TotalKeys }
func (r *IncrementalValidationResult) GetValidatedKeys() int { return r.ValidatedKeys }
func (r *IncrementalValidationResult) GetSkippedKeys() int { return r.SkippedKeys }
func (r *IncrementalValidationResult) GetSuccessfulKeys() int { return r.SuccessfulKeys }
func (r *IncrementalValidationResult) GetFailedKeys() int { return r.FailedKeys }
func (r *IncrementalValidationResult) GetDuration() time.Duration { return r.Duration }

// IncrementalKeyValidationResult 单个密钥的验证结果
type IncrementalKeyValidationResult struct {
	KeyID        uint      `json:"key_id"`
	KeyHash      string    `json:"key_hash"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	IsValid      bool      `json:"is_valid"`
	Error        string    `json:"error,omitempty"`
	Duration     time.Duration `json:"duration"`
	ValidatedAt  time.Time `json:"validated_at"`
}

// IncrementalValidationService 增量验证服务
type IncrementalValidationService struct {
	db        *gorm.DB
	validator interfaces.KeyValidatorInterface
}

// NewIncrementalValidationService 创建新的增量验证服务
func NewIncrementalValidationService(
	db *gorm.DB,
	validator interfaces.KeyValidatorInterface,
) *IncrementalValidationService {
	return &IncrementalValidationService{
		db:        db,
		validator: validator,
	}
}

// ValidateGroup 对指定分组进行增量验证
func (ivs *IncrementalValidationService) ValidateGroup(
	ctx context.Context,
	groupID uint,
	config interfaces.IncrementalValidationConfig,
) (interfaces.IncrementalValidationResult, error) {
	if config == nil {
		config = DefaultIncrementalValidationConfig()
	}

	startTime := time.Now()

	// 获取分组信息
	var group models.Group
	if err := ivs.db.First(&group, groupID).Error; err != nil {
		return nil, fmt.Errorf("failed to find group %d: %w", groupID, err)
	}

	// 构建查询条件
	query := ivs.db.Model(&models.APIKey{}).Where("group_id = ?", groupID)

	// 时间窗口过滤
	if config.GetTimeWindow() > 0 {
		cutoffTime := startTime.Add(-config.GetTimeWindow())
		query = query.Where("created_at >= ?", cutoffTime)
	}

	// 状态过滤
	if len(config.GetIncludeStates()) > 0 {
		query = query.Where("status IN ?", config.GetIncludeStates())
	}

	// 排除最近验证过的密钥
	if config.GetExcludeRecentlyValidated() && config.GetRecentValidationWindow() > 0 {
		recentCutoff := startTime.Add(-config.GetRecentValidationWindow())
		query = query.Where("last_validated_at IS NULL OR last_validated_at < ?", recentCutoff)
	}

	// 获取需要验证的密钥总数
	var totalKeys int64
	if err := query.Count(&totalKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to count keys: %w", err)
	}

	result := &IncrementalValidationResult{
		GroupID:       groupID,
		GroupName:     group.Name,
		TotalKeys:     int(totalKeys),
		StartTime:     startTime,
		KeyResults:    make([]IncrementalKeyValidationResult, 0),
	}

	if totalKeys == 0 {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		logrus.WithFields(logrus.Fields{
			"groupID": groupID,
			"groupName": group.Name,
		}).Info("No keys found for incremental validation")
		return result, nil
	}

	logrus.WithFields(logrus.Fields{
		"groupID": groupID,
		"groupName": group.Name,
		"totalKeys": totalKeys,
		"config": config,
	}).Info("Starting incremental validation")

	// 批量处理密钥
	var processedKeys int
	batchErr := query.FindInBatches(&[]models.APIKey{}, config.GetBatchSize(), func(tx *gorm.DB, batch int) error {
		var keys []models.APIKey
		if dbErr := tx.Find(&keys).Error; dbErr != nil {
			return fmt.Errorf("failed to get keys in batch %d: %w", batch, dbErr)
		}

		batchResult, validateErr := ivs.validateKeysBatch(ctx, &group, keys, config)
		if validateErr != nil {
			logrus.WithError(validateErr).WithFields(logrus.Fields{
				"batch": batch,
				"groupID": groupID,
			}).Error("Failed to validate keys batch")
			return validateErr
		}

		// 合并结果
		result.ValidatedKeys += batchResult.ValidatedKeys
		result.SkippedKeys += batchResult.SkippedKeys
		result.SuccessfulKeys += batchResult.SuccessfulKeys
		result.FailedKeys += batchResult.FailedKeys
		result.KeyResults = append(result.KeyResults, batchResult.KeyResults...)

		processedKeys += len(keys)
		logrus.WithFields(logrus.Fields{
			"batch": batch,
			"processed": processedKeys,
			"total": totalKeys,
			"groupID": groupID,
		}).Debug("Batch validation completed")

		return nil
	}).Error

	if batchErr != nil {
		return nil, fmt.Errorf("failed to process keys in batches: %w", batchErr)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	logrus.WithFields(logrus.Fields{
		"groupID": groupID,
		"groupName": group.Name,
		"totalKeys": result.TotalKeys,
		"validatedKeys": result.ValidatedKeys,
		"successfulKeys": result.SuccessfulKeys,
		"failedKeys": result.FailedKeys,
		"duration": result.Duration,
	}).Info("Incremental validation completed")

	return result, nil
}

// validateKeysBatch 验证一批密钥
func (ivs *IncrementalValidationService) validateKeysBatch(
	ctx context.Context,
	group *models.Group,
	keys []models.APIKey,
	config interfaces.IncrementalValidationConfig,
) (*IncrementalValidationResult, error) {
	result := &IncrementalValidationResult{
		KeyResults: make([]IncrementalKeyValidationResult, 0, len(keys)),
	}

	// 创建工作通道
	keysChan := make(chan models.APIKey, len(keys))
	resultsChan := make(chan IncrementalKeyValidationResult, len(keys))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < config.GetConcurrency(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for key := range keysChan {
				select {
				case <-ctx.Done():
					return
				default:
					keyResult := ivs.validateSingleKey(ctx, group, key)
					resultsChan <- keyResult
				}
			}
		}()
	}

	// 发送密钥到工作通道
	go func() {
		defer close(keysChan)
		for _, key := range keys {
			select {
			case <-ctx.Done():
				return
			case keysChan <- key:
			}
		}
	}()

	// 收集结果
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for keyResult := range resultsChan {
		result.KeyResults = append(result.KeyResults, keyResult)
		result.ValidatedKeys++

		if keyResult.IsValid {
			result.SuccessfulKeys++
		} else {
			result.FailedKeys++
		}
	}

	return result, nil
}

// validateSingleKey 验证单个密钥
func (ivs *IncrementalValidationService) validateSingleKey(
	ctx context.Context,
	group *models.Group,
	key models.APIKey,
) IncrementalKeyValidationResult {
	startTime := time.Now()

	result := IncrementalKeyValidationResult{
		KeyID:       key.ID,
		KeyHash:     key.KeyHash,
		OldStatus:   key.Status,
		ValidatedAt: startTime,
	}

	// 执行验证
	isValid, err := ivs.validator.ValidateSingleKey(&key, group)
	result.IsValid = isValid

	if err != nil {
		result.Error = err.Error()
	}

	// 更新数据库中的验证时间和状态
	updates := map[string]interface{}{
		"last_validated_at": startTime,
	}

	// 根据验证结果更新状态
	if isValid {
		if key.Status == models.KeyStatusPending || key.Status == models.KeyStatusInvalid {
			updates["status"] = models.KeyStatusActive
			result.NewStatus = models.KeyStatusActive
		} else {
			result.NewStatus = key.Status
		}
	} else {
		if key.Status != models.KeyStatusInvalid {
			updates["status"] = models.KeyStatusInvalid
			result.NewStatus = models.KeyStatusInvalid
		} else {
			result.NewStatus = key.Status
		}
	}

	// 更新数据库
	if err := ivs.db.Model(&models.APIKey{}).Where("id = ?", key.ID).Updates(updates).Error; err != nil {
		logrus.WithError(err).WithField("keyID", key.ID).Error("Failed to update key validation status")
		if result.Error == "" {
			result.Error = fmt.Sprintf("Failed to update database: %v", err)
		}
	}

	result.Duration = time.Since(startTime)
	return result
}

// ValidateAllGroups 对所有分组进行增量验证
func (ivs *IncrementalValidationService) ValidateAllGroups(
	ctx context.Context,
	config interfaces.IncrementalValidationConfig,
) ([]interfaces.IncrementalValidationResult, error) {
	var groups []models.Group
	if err := ivs.db.Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	results := make([]interfaces.IncrementalValidationResult, 0, len(groups))

	for _, group := range groups {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result, err := ivs.ValidateGroup(ctx, group.ID, config)
		if err != nil {
			logrus.WithError(err).WithField("groupID", group.ID).Error("Failed to validate group")
			// 创建一个错误结果
			result = &IncrementalValidationResult{
				GroupID:   group.ID,
				GroupName: group.Name,
				StartTime: time.Now(),
				EndTime:   time.Now(),
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// GetValidationHistory 获取验证历史统计
func (ivs *IncrementalValidationService) GetValidationHistory(
	groupID uint,
	timeRange time.Duration,
) (map[string]interface{}, error) {
	cutoffTime := time.Now().Add(-timeRange)

	var stats struct {
		TotalKeys     int64 `gorm:"column:total_keys"`
		ValidatedKeys int64 `gorm:"column:validated_keys"`
		PendingKeys   int64 `gorm:"column:pending_keys"`
		ActiveKeys    int64 `gorm:"column:active_keys"`
		InvalidKeys   int64 `gorm:"column:invalid_keys"`
	}

	query := `
		SELECT
			COUNT(*) as total_keys,
			COUNT(CASE WHEN last_validated_at >= ? THEN 1 END) as validated_keys,
			COUNT(CASE WHEN status = ? THEN 1 END) as pending_keys,
			COUNT(CASE WHEN status = ? THEN 1 END) as active_keys,
			COUNT(CASE WHEN status = ? THEN 1 END) as invalid_keys
		FROM api_keys
		WHERE group_id = ?
	`

	err := ivs.db.Raw(query, cutoffTime, models.KeyStatusPending, models.KeyStatusActive, models.KeyStatusInvalid, groupID).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get validation history: %w", err)
	}

	return map[string]interface{}{
		"total_keys":     stats.TotalKeys,
		"validated_keys": stats.ValidatedKeys,
		"pending_keys":   stats.PendingKeys,
		"active_keys":    stats.ActiveKeys,
		"invalid_keys":   stats.InvalidKeys,
		"validation_rate": func() float64 {
			if stats.TotalKeys > 0 {
				return float64(stats.ValidatedKeys) / float64(stats.TotalKeys)
			}
			return 0
		}(),
	}, nil
}
