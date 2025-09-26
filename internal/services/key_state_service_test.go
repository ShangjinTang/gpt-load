package services

import (
	"fmt"
	"testing"
	"time"

	"gpt-load/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupKeyStateServiceTest(t *testing.T) (*KeyStateService, *gorm.DB, func()) {
	// 设置内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 迁移数据库结构
	err = db.AutoMigrate(&models.APIKey{})
	require.NoError(t, err)

	// 创建服务
	service := NewKeyStateService(db)

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return service, db, cleanup
}

func TestKeyStateService_HandleSuccess(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建测试密钥
	key := &models.APIKey{
		KeyValue:            "test-key",
		Status:              models.KeyStatusDegraded,
		ConsecutiveFailures: 2,
		BackoffLevel:        1,
		LastErrorMessage:    "Previous error",
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 处理成功
	err = service.HandleSuccess(key.ID)
	require.NoError(t, err)

	// 验证状态更新
	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.KeyStatusActive, updatedKey.Status)
	assert.Equal(t, int64(0), updatedKey.ConsecutiveFailures)
	assert.Equal(t, 0, updatedKey.BackoffLevel)
	assert.Empty(t, updatedKey.LastErrorMessage)
	assert.Nil(t, updatedKey.DisabledUntil)
	assert.NotNil(t, updatedKey.LastSuccessAt)
}

func TestKeyStateService_HandleFailure_FirstFailure(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建活跃密钥
	key := &models.APIKey{
		KeyValue: "test-key",
		Status:   models.KeyStatusActive,
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 处理第一次失败
	errorMessage := "API rate limit exceeded"
	err = service.HandleFailure(key.ID, errorMessage)
	require.NoError(t, err)

	// 验证状态更新
	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.KeyStatusDegraded, updatedKey.Status)
	assert.Equal(t, int64(0), updatedKey.FailureCount) // HandleFailure不增加FailureCount
	assert.Equal(t, int64(1), updatedKey.ConsecutiveFailures)
	assert.Equal(t, errorMessage, updatedKey.LastErrorMessage)
	assert.NotNil(t, updatedKey.LastFailureAt)
}

func TestKeyStateService_HandleFailure_MultipleFailures(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建已经有2次连续失败的密钥
	key := &models.APIKey{
		KeyValue:            "test-key",
		Status:              models.KeyStatusDegraded,
		ConsecutiveFailures: 2,
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 处理第三次失败（应该触发禁用）
	errorMessage := "Connection timeout"
	err = service.HandleFailure(key.ID, errorMessage)
	require.NoError(t, err)

	// 验证状态更新
	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.KeyStatusDisabled, updatedKey.Status)
	assert.Equal(t, int64(0), updatedKey.FailureCount) // HandleFailure不增加FailureCount
	assert.Equal(t, int64(3), updatedKey.ConsecutiveFailures)
	assert.Equal(t, 1, updatedKey.BackoffLevel)
	assert.Equal(t, errorMessage, updatedKey.LastErrorMessage)
	assert.NotNil(t, updatedKey.DisabledUntil)
	assert.True(t, updatedKey.DisabledUntil.After(time.Now()))
}

func TestKeyStateService_ManuallyInvalidateKey(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建活跃密钥
	key := &models.APIKey{
		KeyValue: "test-key",
		Status:   models.KeyStatusActive,
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 手动标记为无效
	reason := "Manually invalidated by admin"
	err = service.ManuallyInvalidateKey(key.ID, reason)
	require.NoError(t, err)

	// 验证状态更新
	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.KeyStatusInvalid, updatedKey.Status)
	assert.Equal(t, fmt.Sprintf("Manually invalidated: %s", reason), updatedKey.LastErrorMessage)
	assert.Equal(t, int64(0), updatedKey.ConsecutiveFailures)
	assert.Equal(t, 0, updatedKey.BackoffLevel)
	assert.Nil(t, updatedKey.DisabledUntil)
}

func TestKeyStateService_ManuallyDisableKey(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建活跃密钥
	key := &models.APIKey{
		KeyValue: "test-key",
		Status:   models.KeyStatusActive,
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 手动禁用
	reason := "Temporarily disabled for maintenance"
	err = service.ManuallyDisableKey(key.ID, reason)
	require.NoError(t, err)

	// 验证状态更新
	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.KeyStatusDisabled, updatedKey.Status)
	assert.Equal(t, fmt.Sprintf("Manually disabled: %s", reason), updatedKey.LastErrorMessage)
	assert.Equal(t, 0, updatedKey.BackoffLevel) // 手动禁用不会增加退避级别
	assert.NotNil(t, updatedKey.DisabledUntil)
}

func TestKeyStateService_ManuallyEnableKey(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建禁用的密钥
	disabledUntil := time.Now().Add(1 * time.Hour)
	key := &models.APIKey{
		KeyValue:            "test-key",
		Status:              models.KeyStatusDisabled,
		ConsecutiveFailures: 3,
		BackoffLevel:        2,
		DisabledUntil:       &disabledUntil,
		LastErrorMessage:    "Previous error",
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 手动启用
	err = service.ManuallyEnableKey(key.ID)
	require.NoError(t, err)

	// 验证状态更新
	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)

	assert.Equal(t, models.KeyStatusDegraded, updatedKey.Status) // 手动启用后设为降级状态
	assert.Equal(t, int64(0), updatedKey.ConsecutiveFailures)
	assert.Equal(t, 0, updatedKey.BackoffLevel)
	assert.Nil(t, updatedKey.DisabledUntil)
	assert.Empty(t, updatedKey.LastErrorMessage) // 错误消息被清空
	assert.Nil(t, updatedKey.LastSuccessAt)      // 手动启用不设置成功时间
}

func TestKeyStateService_UpdateKeyStatus_InvalidState(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建测试密钥
	key := &models.APIKey{
		KeyValue: "test-key",
		Status:   models.KeyStatusActive,
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 尝试设置无效状态
	err = service.UpdateKeyStatus(key.ID, "invalid-status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestKeyStateService_HandleSuccess_NonexistentKey(t *testing.T) {
	service, _, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 尝试处理不存在的密钥
	err := service.HandleSuccess(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find key")
}

func TestKeyStateService_HandleFailure_NonexistentKey(t *testing.T) {
	service, _, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 尝试处理不存在的密钥
	err := service.HandleFailure(999, "test error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find key")
}

func TestKeyStateService_StateMachine_Integration(t *testing.T) {
	service, db, cleanup := setupKeyStateServiceTest(t)
	defer cleanup()

	// 创建pending状态的密钥
	key := &models.APIKey{
		KeyValue: "test-key",
		Status:   models.KeyStatusPending,
	}
	err := db.Create(key).Error
	require.NoError(t, err)

	// 1. 首次成功 -> Active
	err = service.HandleSuccess(key.ID)
	require.NoError(t, err)

	var updatedKey models.APIKey
	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.KeyStatusActive, updatedKey.Status)

	// 2. 失败 -> Degraded
	err = service.HandleFailure(key.ID, "First failure")
	require.NoError(t, err)

	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.KeyStatusDegraded, updatedKey.Status)
	assert.Equal(t, int64(1), updatedKey.ConsecutiveFailures)

	// 3. 再次成功 -> Active (重置连续失败)
	err = service.HandleSuccess(key.ID)
	require.NoError(t, err)

	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.KeyStatusActive, updatedKey.Status)
	assert.Equal(t, int64(0), updatedKey.ConsecutiveFailures)

	// 4. 连续3次失败 -> Disabled
	for i := 0; i < 3; i++ {
		err = service.HandleFailure(key.ID, "Consecutive failure")
		require.NoError(t, err)
	}

	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.KeyStatusDisabled, updatedKey.Status)
	assert.Equal(t, int64(3), updatedKey.ConsecutiveFailures)
	assert.Equal(t, 1, updatedKey.BackoffLevel)
	assert.NotNil(t, updatedKey.DisabledUntil)

	// 5. 手动启用 -> Degraded (需要验证后才能变为Active)
	err = service.ManuallyEnableKey(key.ID)
	require.NoError(t, err)

	err = db.First(&updatedKey, key.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.KeyStatusDegraded, updatedKey.Status) // 手动启用后为降级状态
	assert.Equal(t, int64(0), updatedKey.ConsecutiveFailures)
	assert.Equal(t, 0, updatedKey.BackoffLevel)
}
