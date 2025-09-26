package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gpt-load/internal/db"
	"gpt-load/internal/tests"
	"gpt-load/internal/types"
)

func TestNewManager(t *testing.T) {
	// Setup test database and set global DB
	testDB := tests.SetupTestDB(t)
	db.DB = testDB
	defer func() { db.DB = nil }()

	settingsManager := NewSystemSettingsManager()

	// Initialize with minimal store
	store := &tests.MockMemoryStore{}
	err := settingsManager.Initialize(store, &tests.MockGroupManager{}, true)
	assert.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		// Set required environment variables
		os.Setenv("AUTH_KEY", "test-auth-key-12345")
		defer os.Unsetenv("AUTH_KEY")

		manager, err := NewManager(settingsManager)

		assert.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("missing AUTH_KEY", func(t *testing.T) {
		os.Unsetenv("AUTH_KEY")

		manager, err := NewManager(settingsManager)

		assert.Error(t, err)
		assert.Nil(t, manager)
		assert.Contains(t, err.Error(), "AUTH_KEY is required")
	})
}

func TestManager_ReloadConfig(t *testing.T) {
	settingsManager := NewSystemSettingsManager()
	manager := &Manager{settingsManager: settingsManager}

	t.Run("default configuration", func(t *testing.T) {
		// Set required environment variables
		os.Setenv("AUTH_KEY", "test-auth-key-12345")
		defer os.Unsetenv("AUTH_KEY")

		err := manager.ReloadConfig()

		assert.NoError(t, err)
		assert.NotNil(t, manager.config)
		assert.Equal(t, 3001, manager.config.Server.Port)
		assert.Equal(t, "0.0.0.0", manager.config.Server.Host)
		assert.Equal(t, "test-auth-key-12345", manager.config.Auth.Key)
	})

	t.Run("custom configuration", func(t *testing.T) {
		// Set custom environment variables
		os.Setenv("AUTH_KEY", "custom-auth-key-12345")
		os.Setenv("PORT", "8080")
		os.Setenv("HOST", "127.0.0.1")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("ENABLE_CORS", "true")
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
		defer func() {
			os.Unsetenv("AUTH_KEY")
			os.Unsetenv("PORT")
			os.Unsetenv("HOST")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("ENABLE_CORS")
			os.Unsetenv("ALLOWED_ORIGINS")
		}()

		err := manager.ReloadConfig()

		assert.NoError(t, err)
		assert.Equal(t, 8080, manager.config.Server.Port)
		assert.Equal(t, "127.0.0.1", manager.config.Server.Host)
		assert.Equal(t, "custom-auth-key-12345", manager.config.Auth.Key)
		assert.Equal(t, "debug", manager.config.Log.Level)
		assert.True(t, manager.config.CORS.Enabled)
		assert.Equal(t, []string{"http://localhost:3000"}, manager.config.CORS.AllowedOrigins)
	})
}

func TestManager_IsMaster(t *testing.T) {
	settingsManager := NewSystemSettingsManager()
	manager := &Manager{settingsManager: settingsManager}

	t.Run("default is master", func(t *testing.T) {
		os.Setenv("AUTH_KEY", "test-auth-key-12345")
		defer os.Unsetenv("AUTH_KEY")

		err := manager.ReloadConfig()
		assert.NoError(t, err)

		assert.True(t, manager.IsMaster())
	})

	t.Run("slave mode", func(t *testing.T) {
		os.Setenv("AUTH_KEY", "test-auth-key-12345")
		os.Setenv("IS_SLAVE", "true")
		defer func() {
			os.Unsetenv("AUTH_KEY")
			os.Unsetenv("IS_SLAVE")
		}()

		err := manager.ReloadConfig()
		assert.NoError(t, err)

		assert.False(t, manager.IsMaster())
	})
}

func TestManager_GetConfigs(t *testing.T) {
	settingsManager := NewSystemSettingsManager()
	manager := &Manager{settingsManager: settingsManager}

	// Set up valid configuration
	os.Setenv("AUTH_KEY", "test-auth-key-12345")
	os.Setenv("REDIS_DSN", "redis://localhost:6379")
	os.Setenv("ENCRYPTION_KEY", "test-encryption-key")
	defer func() {
		os.Unsetenv("AUTH_KEY")
		os.Unsetenv("REDIS_DSN")
		os.Unsetenv("ENCRYPTION_KEY")
	}()

	err := manager.ReloadConfig()
	assert.NoError(t, err)

	t.Run("GetAuthConfig", func(t *testing.T) {
		authConfig := manager.GetAuthConfig()
		assert.Equal(t, "test-auth-key-12345", authConfig.Key)
	})

	t.Run("GetCORSConfig", func(t *testing.T) {
		corsConfig := manager.GetCORSConfig()
		assert.False(t, corsConfig.Enabled) // Default is false
	})

	t.Run("GetPerformanceConfig", func(t *testing.T) {
		perfConfig := manager.GetPerformanceConfig()
		assert.Equal(t, 100, perfConfig.MaxConcurrentRequests) // Default value
	})

	t.Run("GetLogConfig", func(t *testing.T) {
		logConfig := manager.GetLogConfig()
		assert.Equal(t, "info", logConfig.Level) // Default value
		assert.Equal(t, "text", logConfig.Format) // Default value
	})

	t.Run("GetRedisDSN", func(t *testing.T) {
		redisDSN := manager.GetRedisDSN()
		assert.Equal(t, "redis://localhost:6379", redisDSN)
	})

	t.Run("GetDatabaseConfig", func(t *testing.T) {
		dbConfig := manager.GetDatabaseConfig()
		assert.Equal(t, "./data/gpt-load.db", dbConfig.DSN) // Default value
	})

	t.Run("GetEncryptionKey", func(t *testing.T) {
		encryptionKey := manager.GetEncryptionKey()
		assert.Equal(t, "test-encryption-key", encryptionKey)
	})

	t.Run("GetEffectiveServerConfig", func(t *testing.T) {
		serverConfig := manager.GetEffectiveServerConfig()
		assert.Equal(t, 3001, serverConfig.Port) // Default value
		assert.Equal(t, "0.0.0.0", serverConfig.Host) // Default value
		assert.True(t, serverConfig.IsMaster) // Default value
	})
}

func TestManager_Validate(t *testing.T) {
	settingsManager := NewSystemSettingsManager()
	manager := &Manager{settingsManager: settingsManager}

	t.Run("valid configuration", func(t *testing.T) {
		manager.config = &Config{
			Server: types.ServerConfig{
				Port:                    8080,
				GracefulShutdownTimeout: 15,
			},
			Auth: types.AuthConfig{
				Key: "valid-auth-key-12345",
			},
			Performance: types.PerformanceConfig{
				MaxConcurrentRequests: 50,
			},
			CORS: types.CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{"http://localhost:3000"},
			},
		}

		err := manager.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid port", func(t *testing.T) {
		manager.config = &Config{
			Server: types.ServerConfig{
				Port: 70000, // Invalid port
			},
			Auth: types.AuthConfig{
				Key: "valid-auth-key-12345",
			},
			Performance: types.PerformanceConfig{
				MaxConcurrentRequests: 50,
			},
		}

		err := manager.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port must be between")
	})

	t.Run("empty auth key", func(t *testing.T) {
		manager.config = &Config{
			Server: types.ServerConfig{
				Port: 8080,
			},
			Auth: types.AuthConfig{
				Key: "", // Empty auth key
			},
			Performance: types.PerformanceConfig{
				MaxConcurrentRequests: 50,
			},
		}

		err := manager.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AUTH_KEY is required")
	})

	t.Run("invalid max concurrent requests", func(t *testing.T) {
		manager.config = &Config{
			Server: types.ServerConfig{
				Port: 8080,
			},
			Auth: types.AuthConfig{
				Key: "valid-auth-key-12345",
			},
			Performance: types.PerformanceConfig{
				MaxConcurrentRequests: 0, // Invalid value
			},
		}

		err := manager.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max concurrent requests cannot be less than 1")
	})

	t.Run("short graceful shutdown timeout gets reset", func(t *testing.T) {
		manager.config = &Config{
			Server: types.ServerConfig{
				Port:                    8080,
				GracefulShutdownTimeout: 5, // Too short
			},
			Auth: types.AuthConfig{
				Key: "valid-auth-key-12345",
			},
			Performance: types.PerformanceConfig{
				MaxConcurrentRequests: 50,
			},
		}

		err := manager.Validate()
		assert.NoError(t, err)
		assert.Equal(t, 10, manager.config.Server.GracefulShutdownTimeout) // Should be reset to 10
	})

	t.Run("CORS enabled but no allowed origins", func(t *testing.T) {
		manager.config = &Config{
			Server: types.ServerConfig{
				Port: 8080,
			},
			Auth: types.AuthConfig{
				Key: "valid-auth-key-12345",
			},
			Performance: types.PerformanceConfig{
				MaxConcurrentRequests: 50,
			},
			CORS: types.CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{}, // Empty origins
			},
		}

		err := manager.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CORS is enabled but ALLOWED_ORIGINS is not set")
	})
}

func TestManager_DisplayServerConfig(t *testing.T) {
	settingsManager := NewSystemSettingsManager()
	manager := &Manager{settingsManager: settingsManager}

	// Set up valid configuration
	os.Setenv("AUTH_KEY", "test-auth-key-12345")
	defer os.Unsetenv("AUTH_KEY")

	err := manager.ReloadConfig()
	assert.NoError(t, err)

	// This test just ensures DisplayServerConfig doesn't panic
	// The actual output is logged, so we can't easily test it
	assert.NotPanics(t, func() {
		manager.DisplayServerConfig()
	})
}

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, 1, DefaultConstants.MinPort)
	assert.Equal(t, 65535, DefaultConstants.MaxPort)
	assert.Equal(t, 1, DefaultConstants.MinTimeout)
	assert.Equal(t, 30, DefaultConstants.DefaultTimeout)
	assert.Equal(t, 50, DefaultConstants.DefaultMaxSockets)
	assert.Equal(t, 10, DefaultConstants.DefaultMaxFreeSockets)
}
