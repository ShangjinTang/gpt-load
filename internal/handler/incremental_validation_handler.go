package handler

import (
	"context"
	"strconv"
	"time"

	app_errors "gpt-load/internal/errors"
	"gpt-load/internal/interfaces"
	"gpt-load/internal/response"
	"gpt-load/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// IncrementalValidationHandler 增量验证处理器
type IncrementalValidationHandler struct {
	validationService interfaces.IncrementalValidationServiceInterface
}

// NewIncrementalValidationHandler 创建新的增量验证处理器
func NewIncrementalValidationHandler(
	validationService interfaces.IncrementalValidationServiceInterface,
) *IncrementalValidationHandler {
	return &IncrementalValidationHandler{
		validationService: validationService,
	}
}

// ValidateGroupRequest 验证分组请求
type ValidateGroupRequest struct {
	TimeWindowHours          int      `json:"time_window_hours" binding:"min=0,max=168"`           // 最大7天
	IncludeStates            []string `json:"include_states"`                                      // 包含的状态
	ExcludeRecentlyValidated bool     `json:"exclude_recently_validated"`                         // 排除最近验证过的
	RecentValidationHours    int      `json:"recent_validation_hours" binding:"min=0,max=24"`     // 最近验证时间窗口（小时）
	Concurrency              int      `json:"concurrency" binding:"min=1,max=20"`                 // 并发数
	BatchSize                int      `json:"batch_size" binding:"min=10,max=1000"`               // 批大小
}

// toConfig 将请求转换为配置
func (r *ValidateGroupRequest) toConfig() *services.IncrementalValidationConfig {
	config := services.DefaultIncrementalValidationConfig()

	if r.TimeWindowHours > 0 {
		config.TimeWindow = time.Duration(r.TimeWindowHours) * time.Hour
	}

	if len(r.IncludeStates) > 0 {
		config.IncludeStates = r.IncludeStates
	}

	config.ExcludeRecentlyValidated = r.ExcludeRecentlyValidated

	if r.RecentValidationHours > 0 {
		config.RecentValidationWindow = time.Duration(r.RecentValidationHours) * time.Hour
	}

	if r.Concurrency > 0 {
		config.Concurrency = r.Concurrency
	}

	if r.BatchSize > 0 {
		config.BatchSize = r.BatchSize
	}

	return config
}

// ValidateGroup 验证指定分组的密钥
func (h *IncrementalValidationHandler) ValidateGroup(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.Error(c, app_errors.NewAPIError(app_errors.ErrBadRequest, "Invalid group ID"))
		return
	}

	var req ValidateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, app_errors.NewAPIError(app_errors.ErrBadRequest, "Invalid request: "+err.Error()))
		return
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	config := req.toConfig()

	logrus.WithFields(logrus.Fields{
		"groupID": groupID,
		"config":  config,
	}).Info("Starting incremental validation for group")

	result, err := h.validationService.ValidateGroup(ctx, uint(groupID), config)
	if err != nil {
		logrus.WithError(err).WithField("groupID", groupID).Error("Failed to validate group")
		response.Error(c, app_errors.NewAPIError(app_errors.ErrInternalServer, "Validation failed: "+err.Error()))
		return
	}

	response.Success(c, gin.H{
		"group_id":        result.GetGroupID(),
		"group_name":      result.GetGroupName(),
		"total_keys":      result.GetTotalKeys(),
		"validated_keys":  result.GetValidatedKeys(),
		"successful_keys": result.GetSuccessfulKeys(),
		"failed_keys":     result.GetFailedKeys(),
		"duration_ms":     result.GetDuration().Milliseconds(),
	})
}

// ValidateAllGroups 验证所有分组的密钥
func (h *IncrementalValidationHandler) ValidateAllGroups(c *gin.Context) {
	var req ValidateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, app_errors.NewAPIError(app_errors.ErrBadRequest, "Invalid request: "+err.Error()))
		return
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	config := req.toConfig()

	logrus.WithFields(logrus.Fields{
		"config": config,
	}).Info("Starting incremental validation for all groups")

	results, err := h.validationService.ValidateAllGroups(ctx, config)
	if err != nil {
		logrus.WithError(err).Error("Failed to validate all groups")
		response.Error(c, app_errors.NewAPIError(app_errors.ErrInternalServer, "Validation failed: "+err.Error()))
		return
	}

	// 转换结果
	responseResults := make([]gin.H, len(results))
	for i, result := range results {
		responseResults[i] = gin.H{
			"group_id":        result.GetGroupID(),
			"group_name":      result.GetGroupName(),
			"total_keys":      result.GetTotalKeys(),
			"validated_keys":  result.GetValidatedKeys(),
			"successful_keys": result.GetSuccessfulKeys(),
			"failed_keys":     result.GetFailedKeys(),
			"duration_ms":     result.GetDuration().Milliseconds(),
		}
	}

	response.Success(c, gin.H{
		"results": responseResults,
		"summary": gin.H{
			"total_groups": len(results),
			"total_keys": func() int {
				sum := 0
				for _, result := range results {
					sum += result.GetTotalKeys()
				}
				return sum
			}(),
			"total_validated": func() int {
				sum := 0
				for _, result := range results {
					sum += result.GetValidatedKeys()
				}
				return sum
			}(),
			"total_successful": func() int {
				sum := 0
				for _, result := range results {
					sum += result.GetSuccessfulKeys()
				}
				return sum
			}(),
		},
	})
}

// GetValidationHistory 获取验证历史统计
func (h *IncrementalValidationHandler) GetValidationHistory(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.Error(c, app_errors.NewAPIError(app_errors.ErrBadRequest, "Invalid group ID"))
		return
	}

	// 获取时间范围参数，默认24小时
	timeRangeHours := 24
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil && hours > 0 && hours <= 168 { // 最大7天
			timeRangeHours = hours
		}
	}

	timeRange := time.Duration(timeRangeHours) * time.Hour

	history, err := h.validationService.GetValidationHistory(uint(groupID), timeRange)
	if err != nil {
		logrus.WithError(err).WithField("groupID", groupID).Error("Failed to get validation history")
		response.Error(c, app_errors.NewAPIError(app_errors.ErrInternalServer, "Failed to get validation history: "+err.Error()))
		return
	}

	response.Success(c, gin.H{
		"group_id":     groupID,
		"time_range":   timeRangeHours,
		"history":      history,
	})
}
