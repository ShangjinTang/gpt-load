package keypool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"

	"gpt-load/internal/config"
	"gpt-load/internal/encryption"
	"gpt-load/internal/models"
	"gpt-load/internal/tests"
	"gpt-load/internal/types"
)

// MockKeyValidator 模拟 KeyValidatorInterface
type MockKeyValidator struct {
	mock.Mock
}

func (m *MockKeyValidator) ValidateGroup(group *models.Group) (map[uint]interface{}, error) {
	args := m.Called(group)
	return args.Get(0).(map[uint]interface{}), args.Error(1)
}

func (m *MockKeyValidator) ValidateSingleKey(key *models.APIKey, group *models.Group) (bool, error) {
	args := m.Called(key, group)
	return args.Bool(0), args.Error(1)
}

func (m *MockKeyValidator) TestMultipleKeys(group *models.Group, keyValues []string) ([]interface{}, error) {
	args := m.Called(group, keyValues)
	return args.Get(0).([]interface{}), args.Error(1)
}

func TestNewCronChecker(t *testing.T) {
	db := tests.SetupTestDB(t)
	settingsManager := &config.SystemSettingsManager{}
	validator := &MockKeyValidator{}
	encryptionSvc, _ := encryption.NewService("test-password")

	checker := NewCronChecker(db, settingsManager, validator, encryptionSvc)

	assert.NotNil(t, checker)
	assert.Equal(t, db, checker.DB)
	assert.Equal(t, settingsManager, checker.SettingsManager)
	assert.Equal(t, validator, checker.Validator)
	assert.Equal(t, encryptionSvc, checker.EncryptionSvc)
	assert.NotNil(t, checker.stopChan)
}

func TestCronChecker_StartStop(t *testing.T) {
	db := tests.SetupTestDB(t)
	settingsManager := &config.SystemSettingsManager{}
	validator := &MockKeyValidator{}
	encryptionSvc, _ := encryption.NewService("test-password")

	checker := NewCronChecker(db, settingsManager, validator, encryptionSvc)

	// Start the checker
	checker.Start()

	// Stop the checker with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checker.Stop(ctx)

	// Verify that the checker stopped gracefully
	select {
	case <-ctx.Done():
		t.Fatal("CronChecker did not stop within timeout")
	default:
		// Success - checker stopped before timeout
	}
}

func TestCronChecker_submitValidationJobs(t *testing.T) {
	db := tests.SetupTestDB(t)
	settingsManager := config.NewSystemSettingsManager()
	validator := &MockKeyValidator{}
	encryptionSvc, _ := encryption.NewService("test-password")

	checker := NewCronChecker(db, settingsManager, validator, encryptionSvc)

	t.Run("no groups", func(t *testing.T) {
		// No groups in database, should not panic
		checker.submitValidationJobs()
	})

	t.Run("group with recent validation", func(t *testing.T) {
		// Create a group with recent validation
		recentTime := time.Now().Add(-1 * time.Minute)
		group := models.Group{
			Name:            "test-group",
			LastValidatedAt: &recentTime,
			Upstreams:       datatypes.JSON(`["http://localhost:8080"]`),
			Config: datatypes.JSONMap{
				"key_validation_interval_minutes": 5, // 5 minutes interval
			},
		}
		db.Create(&group)

		// Should not trigger validation since it's too recent
		checker.submitValidationJobs()

		// Clean up
		db.Delete(&group)
	})

	t.Run("group needs validation", func(t *testing.T) {
		// Create a group that needs validation
		oldTime := time.Now().Add(-10 * time.Minute)
		group := models.Group{
			Name:            "test-group",
			LastValidatedAt: &oldTime,
			Upstreams:       datatypes.JSON(`["http://localhost:8080"]`),
			Config: datatypes.JSONMap{
				"key_validation_interval_minutes": 5, // 5 minutes interval
				"key_validation_concurrency":      1,
			},
		}
		db.Create(&group)

		// Encrypt the test key properly
		encryptedKey, _ := encryptionSvc.Encrypt("test-key-value")

		// Create some invalid keys
		invalidKey := models.APIKey{
			KeyValue: encryptedKey,
			Status:   models.KeyStatusInvalid,
			GroupID:  group.ID,
		}
		db.Create(&invalidKey)

		// Mock validator to return valid
		validator.On("ValidateSingleKey", mock.AnythingOfType("*models.APIKey"), mock.AnythingOfType("*models.Group")).Return(true, nil).Once()

		checker.submitValidationJobs()

		// Wait a bit for goroutines to complete
		time.Sleep(200 * time.Millisecond)

		validator.AssertExpectations(t)

		// Clean up
		db.Delete(&invalidKey)
		db.Delete(&group)
	})
}

func TestCronChecker_validateGroupKeys(t *testing.T) {
	db := tests.SetupTestDB(t)
	settingsManager := config.NewSystemSettingsManager()
	validator := &MockKeyValidator{}
	encryptionSvc, _ := encryption.NewService("test-password")

	checker := NewCronChecker(db, settingsManager, validator, encryptionSvc)

	t.Run("no invalid keys", func(t *testing.T) {
		group := &models.Group{
			ID:        1,
			Name:      "test-group",
			Upstreams: datatypes.JSON(`["http://localhost:8080"]`),
			EffectiveConfig: types.SystemSettings{
				KeyValidationConcurrency: 1,
			},
		}
		db.Create(&group)

		checker.validateGroupKeys(group)

		// Verify last_validated_at was updated
		var updatedGroup models.Group
		db.First(&updatedGroup, group.ID)
		assert.NotNil(t, updatedGroup.LastValidatedAt)

		// Clean up
		db.Delete(&group)
	})

	t.Run("has invalid keys", func(t *testing.T) {
		group := &models.Group{
			ID:        1,
			Name:      "test-group",
			Upstreams: datatypes.JSON(`["http://localhost:8080"]`),
			EffectiveConfig: types.SystemSettings{
				KeyValidationConcurrency: 2,
			},
		}
		db.Create(&group)

		// Encrypt the test key
		encryptedKey, _ := encryptionSvc.Encrypt("test-key-value")

		// Create invalid keys
		invalidKeys := []models.APIKey{
			{
				KeyValue: encryptedKey,
				Status:   models.KeyStatusInvalid,
				GroupID:  group.ID,
			},
			{
				KeyValue: encryptedKey,
				Status:   models.KeyStatusInvalid,
				GroupID:  group.ID,
			},
		}
		db.Create(&invalidKeys)

		// Mock validator responses
		validator.On("ValidateSingleKey", mock.AnythingOfType("*models.APIKey"), mock.AnythingOfType("*models.Group")).Return(true, nil).Times(2)

		checker.validateGroupKeys(group)

		// Wait a bit for goroutines to complete
		time.Sleep(200 * time.Millisecond)

		validator.AssertExpectations(t)

		// Verify last_validated_at was updated
		var updatedGroup models.Group
		db.First(&updatedGroup, group.ID)
		assert.NotNil(t, updatedGroup.LastValidatedAt)

		// Clean up
		db.Delete(&invalidKeys)
		db.Delete(&group)
	})

	t.Run("decryption failure", func(t *testing.T) {
		group := &models.Group{
			ID:        1,
			Name:      "test-group",
			Upstreams: datatypes.JSON(`["http://localhost:8080"]`),
			EffectiveConfig: types.SystemSettings{
				KeyValidationConcurrency: 1,
			},
		}
		db.Create(&group)

		// Create invalid key with bad encrypted value
		invalidKey := models.APIKey{
			KeyValue: "bad-encrypted-value",
			Status:   models.KeyStatusInvalid,
			GroupID:  group.ID,
		}
		db.Create(&invalidKey)

		// Should not call validator due to decryption failure
		checker.validateGroupKeys(group)

		// Wait a bit for goroutines to complete
		time.Sleep(200 * time.Millisecond)

		// Verify last_validated_at was updated despite decryption failure
		var updatedGroup models.Group
		db.First(&updatedGroup, group.ID)
		assert.NotNil(t, updatedGroup.LastValidatedAt)

		// Clean up
		db.Delete(&invalidKey)
		db.Delete(&group)
	})
}

func TestCronChecker_StopWithTimeout(t *testing.T) {
	db := tests.SetupTestDB(t)
	settingsManager := &config.SystemSettingsManager{}
	validator := &MockKeyValidator{}
	encryptionSvc, _ := encryption.NewService("test-password")

	checker := NewCronChecker(db, settingsManager, validator, encryptionSvc)

	// Start the checker
	checker.Start()

	// Stop with very short timeout to test timeout scenario
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should timeout
	checker.Stop(ctx)

	// The test should complete without hanging
}

func TestCronChecker_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := tests.SetupTestDB(t)
	settingsManager := config.NewSystemSettingsManager()
	validator := &MockKeyValidator{}
	encryptionSvc, _ := encryption.NewService("test-password")

	checker := NewCronChecker(db, settingsManager, validator, encryptionSvc)

	// Create a group that needs validation
	group := models.Group{
		Name:      "integration-test-group",
		Upstreams: datatypes.JSON(`["http://localhost:8080"]`),
		Config: datatypes.JSONMap{
			"key_validation_interval_minutes": 1, // 1 minute interval
			"key_validation_concurrency":      2,
		},
	}
	db.Create(&group)

	// Encrypt the test key
	encryptedKey, _ := encryptionSvc.Encrypt("integration-test-key")

	// Create an invalid key
	invalidKey := models.APIKey{
		KeyValue: encryptedKey,
		Status:   models.KeyStatusInvalid,
		GroupID:  group.ID,
	}
	db.Create(&invalidKey)

	// Mock validator to return valid
	validator.On("ValidateSingleKey", mock.AnythingOfType("*models.APIKey"), mock.AnythingOfType("*models.Group")).Return(true, nil).Maybe()

	// Start the checker
	checker.Start()

	// Wait for at least one validation cycle
	time.Sleep(1 * time.Second)

	// Stop the checker
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	checker.Stop(ctx)

	// Verify that validation occurred
	var updatedGroup models.Group
	db.First(&updatedGroup, group.ID)
	assert.NotNil(t, updatedGroup.LastValidatedAt)

	// Clean up
	db.Delete(&invalidKey)
	db.Delete(&group)
}
