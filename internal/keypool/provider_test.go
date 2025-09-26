package keypool

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gpt-load/internal/config"
	"gpt-load/internal/encryption"
	app_errors "gpt-load/internal/errors"
	"gpt-load/internal/models"
	"gpt-load/internal/store"
	"gpt-load/internal/tests"
)

// MockStore 模拟 Store 接口
type MockStore struct {
	mock.Mock
}


func (m *MockStore) HSet(key string, fields map[string]interface{}) error {
	args := m.Called(key, fields)
	return args.Error(0)
}

func (m *MockStore) HGet(key, field string) (string, error) {
	args := m.Called(key, field)
	return args.String(0), args.Error(1)
}

func (m *MockStore) HGetAll(key string) (map[string]string, error) {
	args := m.Called(key)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(map[string]string), args.Error(1)
}

func (m *MockStore) HIncrBy(key, field string, incr int64) (int64, error) {
	args := m.Called(key, field, incr)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) LPush(key string, values ...interface{}) error {
	args := m.Called(key, values)
	return args.Error(0)
}

func (m *MockStore) LRem(key string, count int64, value interface{}) error {
	args := m.Called(key, count, value)
	return args.Error(0)
}

func (m *MockStore) Rotate(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStore) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStore) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockStore) Set(key string, value []byte, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockStore) Get(key string) ([]byte, error) {
	args := m.Called(key)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]byte), args.Error(1)
}

func (m *MockStore) Del(keys ...string) error {
	args := m.Called(keys)
	return args.Error(0)
}

func (m *MockStore) Exists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *MockStore) SetNX(key string, value []byte, ttl time.Duration) (bool, error) {
	args := m.Called(key, value, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *MockStore) SAdd(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *MockStore) SPopN(key string, count int64) ([]string, error) {
	args := m.Called(key, count)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]string), args.Error(1)
}

func (m *MockStore) Publish(channel string, message []byte) error {
	args := m.Called(channel, message)
	return args.Error(0)
}

func (m *MockStore) Subscribe(channel string) (store.Subscription, error) {
	args := m.Called(channel)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(store.Subscription), args.Error(1)
}

func TestNewProvider(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	assert.NotNil(t, provider)
	assert.Equal(t, db, provider.db)
	assert.Equal(t, mockStore, provider.store)
	assert.Equal(t, settingsManager, provider.settingsManager)
	assert.Equal(t, encryptionSvc, provider.encryptionSvc)
}

func TestKeyProvider_SelectKey(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful key selection", func(t *testing.T) {
		groupID := uint(1)
		keyID := "123"

		// 设置mock期望
		mockStore.On("Rotate", "group:1:active_keys").Return(keyID, nil).Once()

		keyDetails := map[string]string{
			"key_string":    "test-key-value",
			"status":        "active",
			"failure_count": "0",
			"created_at":    "1640995200", // 2022-01-01 00:00:00 UTC
		}
		mockStore.On("HGetAll", "key:123").Return(keyDetails, nil).Once()

		key, err := provider.SelectKey(groupID)

		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, uint(123), key.ID)
		assert.Equal(t, "test-key-value", key.KeyValue)
		assert.Equal(t, "active", key.Status)
		assert.Equal(t, int64(0), key.FailureCount)
		assert.Equal(t, groupID, key.GroupID)

		mockStore.AssertExpectations(t)
	})

	t.Run("no active keys", func(t *testing.T) {
		groupID := uint(1)

		mockStore.On("Rotate", "group:1:active_keys").Return("", store.ErrNotFound).Once()

		key, err := provider.SelectKey(groupID)

		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Equal(t, app_errors.ErrNoActiveKeys, err)

		mockStore.AssertExpectations(t)
	})

	t.Run("invalid key ID", func(t *testing.T) {
		groupID := uint(1)

		mockStore.On("Rotate", "group:1:active_keys").Return("invalid", nil).Once()

		key, err := provider.SelectKey(groupID)

		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "failed to parse key ID")

		mockStore.AssertExpectations(t)
	})

	t.Run("key details not found", func(t *testing.T) {
		groupID := uint(1)
		keyID := "123"

		mockStore.On("Rotate", "group:1:active_keys").Return(keyID, nil).Once()
		mockStore.On("HGetAll", "key:123").Return(nil, errors.New("key not found")).Once()

		key, err := provider.SelectKey(groupID)

		assert.Error(t, err)
		assert.Nil(t, key)
		assert.Contains(t, err.Error(), "failed to get key details")

		mockStore.AssertExpectations(t)
	})
}

func TestKeyProvider_UpdateStatus(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("success update", func(t *testing.T) {
		apiKey := &models.APIKey{
			ID:      1,
			GroupID: 1,
		}
		group := &models.Group{
			ID: 1,
		}

		// Mock for success handling
		keyDetails := map[string]string{
			"failure_count": "1",
			"status":        "active",
		}
		mockStore.On("HGetAll", "key:1").Return(keyDetails, nil).Once()
		mockStore.On("HSet", "key:1", mock.AnythingOfType("map[string]interface {}")).Return(nil).Once()

		// Create a test key in the database for the transaction
		testKey := models.APIKey{
			ID:           1,
			KeyValue:     "test-key",
			Status:       "active",
			FailureCount: 1,
			GroupID:      1,
		}
		db.Create(&testKey)

		// Call UpdateStatus and wait a bit for goroutine to complete
		provider.UpdateStatus(apiKey, group, true, "")
		time.Sleep(100 * time.Millisecond)

		mockStore.AssertExpectations(t)
	})

	t.Run("uncounted error", func(t *testing.T) {
		apiKey := &models.APIKey{
			ID:      1,
			GroupID: 1,
		}
		group := &models.Group{
			ID: 1,
		}

		// Call UpdateStatus with uncounted error - should not call store methods
		provider.UpdateStatus(apiKey, group, false, "resource has been exhausted")
		time.Sleep(100 * time.Millisecond)

		// No expectations should be called for uncounted errors
	})
}

func TestKeyProvider_AddKeys(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful add keys", func(t *testing.T) {
		groupID := uint(1)
		keys := []models.APIKey{
			{
				KeyValue: "test-key-1",
				Status:   models.KeyStatusActive,
				GroupID:  groupID,
			},
			{
				KeyValue: "test-key-2",
				Status:   models.KeyStatusActive,
				GroupID:  groupID,
			},
		}

		// Mock store operations
		mockStore.On("HSet", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(2)
		mockStore.On("LRem", mock.AnythingOfType("string"), int64(0), mock.Anything).Return(nil).Times(2)
		mockStore.On("LPush", mock.AnythingOfType("string"), mock.Anything).Return(nil).Times(2)

		err := provider.AddKeys(groupID, keys)

		assert.NoError(t, err)

		// Verify keys were added to database
		var dbKeys []models.APIKey
		db.Where("group_id = ?", groupID).Find(&dbKeys)
		assert.Len(t, dbKeys, 2)

		mockStore.AssertExpectations(t)
	})

	t.Run("empty keys slice", func(t *testing.T) {
		groupID := uint(1)
		var keys []models.APIKey

		err := provider.AddKeys(groupID, keys)

		assert.NoError(t, err)
	})
}

func TestKeyProvider_RemoveKeys(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful remove keys", func(t *testing.T) {
		groupID := uint(1)

		// Create test keys in database first
		testKeys := []models.APIKey{
			{
				KeyValue: "test-key-1",
				KeyHash:  encryptionSvc.Hash("test-key-1"),
				Status:   models.KeyStatusActive,
				GroupID:  groupID,
			},
			{
				KeyValue: "test-key-2",
				KeyHash:  encryptionSvc.Hash("test-key-2"),
				Status:   models.KeyStatusActive,
				GroupID:  groupID,
			},
		}
		db.Create(&testKeys)

		keyValues := []string{"test-key-1", "test-key-2"}

		// Mock store operations for removal
		mockStore.On("LRem", mock.AnythingOfType("string"), int64(0), mock.Anything).Return(nil).Times(2)
		mockStore.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(2)

		deletedCount, err := provider.RemoveKeys(groupID, keyValues)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), deletedCount)

		// Verify keys were removed from database
		var dbKeys []models.APIKey
		db.Where("group_id = ?", groupID).Find(&dbKeys)
		assert.Len(t, dbKeys, 0)

		mockStore.AssertExpectations(t)
	})

	t.Run("empty key values", func(t *testing.T) {
		groupID := uint(1)
		var keyValues []string

		deletedCount, err := provider.RemoveKeys(groupID, keyValues)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), deletedCount)
	})
}

func TestKeyProvider_RestoreKeys(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful restore keys", func(t *testing.T) {
		groupID := uint(1)

		// Create test invalid keys in database
		testKeys := []models.APIKey{
			{
				KeyValue:     "test-key-1",
				Status:       models.KeyStatusInvalid,
				FailureCount: 5,
				GroupID:      groupID,
			},
			{
				KeyValue:     "test-key-2",
				Status:       models.KeyStatusInvalid,
				FailureCount: 3,
				GroupID:      groupID,
			},
		}
		db.Create(&testKeys)

		// Mock store operations for restoration
		mockStore.On("HSet", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(2)
		mockStore.On("LRem", mock.AnythingOfType("string"), int64(0), mock.Anything).Return(nil).Times(2)
		mockStore.On("LPush", mock.AnythingOfType("string"), mock.Anything).Return(nil).Times(2)

		restoredCount, err := provider.RestoreKeys(groupID)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), restoredCount)

		// Verify keys were restored in database
		var dbKeys []models.APIKey
		db.Where("group_id = ? AND status = ?", groupID, models.KeyStatusActive).Find(&dbKeys)
		assert.Len(t, dbKeys, 2)
		for _, key := range dbKeys {
			assert.Equal(t, int64(0), key.FailureCount)
		}

		mockStore.AssertExpectations(t)
	})

	t.Run("no invalid keys to restore", func(t *testing.T) {
		groupID := uint(2)

		restoredCount, err := provider.RestoreKeys(groupID)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), restoredCount)
	})
}

func TestKeyProvider_RemoveInvalidKeys(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful remove invalid keys", func(t *testing.T) {
		groupID := uint(1)

		// Create test keys with mixed statuses
		testKeys := []models.APIKey{
			{
				KeyValue: "test-key-1",
				Status:   models.KeyStatusInvalid,
				GroupID:  groupID,
			},
			{
				KeyValue: "test-key-2",
				Status:   models.KeyStatusActive,
				GroupID:  groupID,
			},
			{
				KeyValue: "test-key-3",
				Status:   models.KeyStatusInvalid,
				GroupID:  groupID,
			},
		}
		db.Create(&testKeys)

		// Mock store operations for removal
		mockStore.On("LRem", mock.AnythingOfType("string"), int64(0), mock.Anything).Return(nil).Times(2)
		mockStore.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(2)

		removedCount, err := provider.RemoveInvalidKeys(groupID)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), removedCount)

		// Verify only invalid keys were removed
		var dbKeys []models.APIKey
		db.Where("group_id = ?", groupID).Find(&dbKeys)
		assert.Len(t, dbKeys, 1)
		assert.Equal(t, models.KeyStatusActive, dbKeys[0].Status)

		mockStore.AssertExpectations(t)
	})
}

func TestKeyProvider_RemoveAllKeys(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful remove all keys", func(t *testing.T) {
		groupID := uint(1)

		// Create test keys
		testKeys := []models.APIKey{
			{
				KeyValue: "test-key-1",
				Status:   models.KeyStatusActive,
				GroupID:  groupID,
			},
			{
				KeyValue: "test-key-2",
				Status:   models.KeyStatusInvalid,
				GroupID:  groupID,
			},
		}
		db.Create(&testKeys)

		// Mock store operations for removal
		mockStore.On("LRem", mock.AnythingOfType("string"), int64(0), mock.Anything).Return(nil).Times(2)
		mockStore.On("Delete", mock.AnythingOfType("string")).Return(nil).Times(2)

		removedCount, err := provider.RemoveAllKeys(groupID)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), removedCount)

		// Verify all keys were removed
		var dbKeys []models.APIKey
		db.Where("group_id = ?", groupID).Find(&dbKeys)
		assert.Len(t, dbKeys, 0)

		mockStore.AssertExpectations(t)
	})
}

func TestKeyProvider_RemoveKeysFromStore(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful remove keys from store", func(t *testing.T) {
		groupID := uint(1)
		keyIDs := []uint{1, 2, 3}

		// Mock store operations
		mockStore.On("Delete", "group:1:active_keys").Return(nil).Once()
		mockStore.On("Delete", "key:1").Return(nil).Once()
		mockStore.On("Delete", "key:2").Return(nil).Once()
		mockStore.On("Delete", "key:3").Return(nil).Once()

		err := provider.RemoveKeysFromStore(groupID, keyIDs)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("empty key IDs", func(t *testing.T) {
		groupID := uint(1)
		var keyIDs []uint

		err := provider.RemoveKeysFromStore(groupID, keyIDs)

		assert.NoError(t, err)
	})
}

func TestKeyProvider_LoadKeysFromDB(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	t.Run("successful load keys from DB", func(t *testing.T) {
		// Create test keys in database
		testKeys := []models.APIKey{
			{
				KeyValue: "test-key-1",
				Status:   models.KeyStatusActive,
				GroupID:  1,
			},
			{
				KeyValue: "test-key-2",
				Status:   models.KeyStatusActive,
				GroupID:  1,
			},
		}
		db.Create(&testKeys)

		// Mock store operations - only 2 keys now
		mockStore.On("HSet", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil).Times(2)
		mockStore.On("Delete", "group:1:active_keys").Return(nil).Once()
		mockStore.On("LPush", "group:1:active_keys", mock.Anything).Return(nil).Once()

		err := provider.LoadKeysFromDB()

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})
}

func TestPluckIDs(t *testing.T) {
	keys := []models.APIKey{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	ids := pluckIDs(keys)

	expected := []uint{1, 2, 3}
	assert.Equal(t, expected, ids)
}

func TestKeyProvider_apiKeyToMap(t *testing.T) {
	db := tests.SetupTestDB(t)
	mockStore := &MockStore{}
	settingsManager := &config.SystemSettingsManager{}
	encryptionSvc, _ := encryption.NewService("test-password")

	provider := NewProvider(db, mockStore, settingsManager, encryptionSvc)

	createdAt := time.Now()
	key := &models.APIKey{
		ID:           123,
		KeyValue:     "test-key-value",
		Status:       models.KeyStatusActive,
		FailureCount: 5,
		GroupID:      1,
		CreatedAt:    createdAt,
	}

	result := provider.apiKeyToMap(key)

	expected := map[string]any{
		"id":            "123",
		"key_string":    "test-key-value",
		"status":        models.KeyStatusActive,
		"failure_count": int64(5),
		"group_id":      uint(1),
		"created_at":    createdAt.Unix(),
	}

	assert.Equal(t, expected, result)
}
