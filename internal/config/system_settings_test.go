package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"gpt-load/internal/db"
	"gpt-load/internal/models"
	"gpt-load/internal/tests"
)

func TestNewSystemSettingsManager(t *testing.T) {
	manager := NewSystemSettingsManager()

	assert.NotNil(t, manager)
}

func TestSystemSettingsManager_Initialize(t *testing.T) {
	// Setup test database and set global DB
	testDB := tests.SetupTestDB(t)
	db.DB = testDB
	defer func() { db.DB = nil }()

	manager := NewSystemSettingsManager()
	store := &tests.MockMemoryStore{}
	groupManager := &tests.MockGroupManager{}

	t.Run("successful initialization as master", func(t *testing.T) {
		err := manager.Initialize(store, groupManager, true)
		assert.NoError(t, err)
	})

	t.Run("successful initialization as slave", func(t *testing.T) {
		err := manager.Initialize(store, groupManager, false)
		assert.NoError(t, err)
	})
}

func TestSystemSettingsManager_GetSettings(t *testing.T) {
	manager := NewSystemSettingsManager()

	t.Run("uninitialized manager returns default settings", func(t *testing.T) {
		settings := manager.GetSettings()

		// Should return default settings
		assert.Equal(t, 600, settings.RequestTimeout)
		assert.Equal(t, 3, settings.MaxRetries)
		assert.Equal(t, 3, settings.BlacklistThreshold)
	})

	t.Run("initialized manager", func(t *testing.T) {
		// Setup test database and set global DB
		testDB := tests.SetupTestDB(t)
		db.DB = testDB
		defer func() { db.DB = nil }()

		store := &tests.MockMemoryStore{}
		groupManager := &tests.MockGroupManager{}

		err := manager.Initialize(store, groupManager, true)
		assert.NoError(t, err)

		settings := manager.GetSettings()
		assert.NotNil(t, settings)
	})
}

func TestSystemSettingsManager_GetEffectiveConfig(t *testing.T) {
	// Setup test database and set global DB
	testDB := tests.SetupTestDB(t)
	db.DB = testDB
	defer func() { db.DB = nil }()

	manager := NewSystemSettingsManager()
	store := &tests.MockMemoryStore{}
	groupManager := &tests.MockGroupManager{}

	err := manager.Initialize(store, groupManager, true)
	assert.NoError(t, err)

	t.Run("nil group config returns system settings", func(t *testing.T) {
		effectiveConfig := manager.GetEffectiveConfig(nil)

		// Should be the same as system settings
		systemSettings := manager.GetSettings()
		assert.Equal(t, systemSettings, effectiveConfig)
	})

	t.Run("empty group config returns system settings", func(t *testing.T) {
		groupConfig := datatypes.JSONMap{}
		effectiveConfig := manager.GetEffectiveConfig(groupConfig)

		// Should be the same as system settings
		systemSettings := manager.GetSettings()
		assert.Equal(t, systemSettings, effectiveConfig)
	})

	t.Run("group config overrides system settings", func(t *testing.T) {
		requestTimeout := 60
		maxRetries := 5

		groupConfig := datatypes.JSONMap{
			"request_timeout": requestTimeout,
			"max_retries":    maxRetries,
		}

		effectiveConfig := manager.GetEffectiveConfig(groupConfig)

		assert.Equal(t, requestTimeout, effectiveConfig.RequestTimeout)
		assert.Equal(t, maxRetries, effectiveConfig.MaxRetries)
		// Other values should remain from system settings
		assert.Equal(t, manager.GetSettings().BlacklistThreshold, effectiveConfig.BlacklistThreshold)
	})

	t.Run("invalid group config JSON", func(t *testing.T) {
		// Create an invalid JSON map that can't be marshaled properly
		groupConfig := datatypes.JSONMap{
			"invalid": make(chan int), // channels can't be marshaled to JSON
		}

		effectiveConfig := manager.GetEffectiveConfig(groupConfig)

		// Should fallback to system settings
		systemSettings := manager.GetSettings()
		assert.Equal(t, systemSettings, effectiveConfig)
	})
}

func TestSystemSettingsManager_ValidateSettings(t *testing.T) {
	manager := NewSystemSettingsManager()

	t.Run("valid settings", func(t *testing.T) {
		settingsMap := map[string]any{
			"request_timeout":    float64(30),
			"max_retries":        float64(3),
			"blacklist_threshold": float64(5),
		}

		err := manager.ValidateSettings(settingsMap)
		assert.NoError(t, err)
	})

	t.Run("invalid setting key", func(t *testing.T) {
		settingsMap := map[string]any{
			"invalid_key": "value",
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid setting key: invalid_key")
	})

	t.Run("invalid type for integer field", func(t *testing.T) {
		settingsMap := map[string]any{
			"request_timeout": "not_a_number",
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type for request_timeout")
	})

	t.Run("non-integer float value", func(t *testing.T) {
		settingsMap := map[string]any{
			"request_timeout": 30.5, // Not an integer
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be an integer")
	})

	t.Run("value below minimum", func(t *testing.T) {
		settingsMap := map[string]any{
			"request_timeout": float64(0), // Below minimum of 1
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is below minimum value")
	})

	t.Run("invalid type for boolean field", func(t *testing.T) {
		settingsMap := map[string]any{
			"enable_request_body_logging": "not_a_boolean",
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type for enable_request_body_logging")
	})

	t.Run("invalid type for string field", func(t *testing.T) {
		settingsMap := map[string]any{
			"app_url": 123, // Should be string
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid type for app_url")
	})
}

func TestSystemSettingsManager_UpdateSettings(t *testing.T) {
	testDB := tests.SetupTestDB(t)
	db.DB = testDB
	defer func() { db.DB = nil }()

	manager := NewSystemSettingsManager()
	store := &tests.MockMemoryStore{}
	groupManager := &tests.MockGroupManager{}

	err := manager.Initialize(store, groupManager, true)
	assert.NoError(t, err)

	t.Run("update valid settings", func(t *testing.T) {
		settingsMap := map[string]any{
			"request_timeout": float64(45),
			"max_retries":     float64(5),
		}

		err := manager.UpdateSettings(settingsMap)
		assert.NoError(t, err)

		// Verify settings were updated in the database
		var settings []models.SystemSetting
		testDB.Find(&settings)

		found := make(map[string]string)
		for _, setting := range settings {
			found[setting.SettingKey] = setting.SettingValue
		}

		assert.Equal(t, "45", found["request_timeout"])
		assert.Equal(t, "5", found["max_retries"])
	})

	t.Run("update invalid settings", func(t *testing.T) {
		settingsMap := map[string]any{
			"invalid_key": "value",
		}

		err := manager.UpdateSettings(settingsMap)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid setting key")
	})

	t.Run("empty settings map", func(t *testing.T) {
		settingsMap := map[string]any{}

		err := manager.UpdateSettings(settingsMap)
		assert.NoError(t, err) // Should succeed with no updates
	})
}


func TestSystemSettingsManager_EdgeCases(t *testing.T) {
	manager := NewSystemSettingsManager()

	t.Run("GetSettings before initialization", func(t *testing.T) {
		// Should not panic and return default settings
		settings := manager.GetSettings()
		assert.NotNil(t, settings)
	})

	t.Run("GetEffectiveConfig with malformed JSON", func(t *testing.T) {
		// This test ensures graceful handling of malformed JSON
		groupConfig := datatypes.JSONMap{
			"request_timeout": "invalid_number_string",
		}

		effectiveConfig := manager.GetEffectiveConfig(groupConfig)
		assert.NotNil(t, effectiveConfig)
	})

	t.Run("ValidateSettings with mixed valid and invalid", func(t *testing.T) {
		settingsMap := map[string]any{
			"request_timeout": float64(30), // valid
			"invalid_key":     "value",     // invalid
		}

		err := manager.ValidateSettings(settingsMap)
		assert.Error(t, err) // Should fail due to invalid key
	})
}
