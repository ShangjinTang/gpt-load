package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gpt-load/internal/types"
)

// MockConfigManager 模拟配置管理器
type MockConfigManager struct {
	mock.Mock
}

func (m *MockConfigManager) IsMaster() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfigManager) GetAuthConfig() types.AuthConfig {
	args := m.Called()
	return args.Get(0).(types.AuthConfig)
}

func (m *MockConfigManager) GetCORSConfig() types.CORSConfig {
	args := m.Called()
	return args.Get(0).(types.CORSConfig)
}

func (m *MockConfigManager) GetPerformanceConfig() types.PerformanceConfig {
	args := m.Called()
	return args.Get(0).(types.PerformanceConfig)
}

func (m *MockConfigManager) GetLogConfig() types.LogConfig {
	args := m.Called()
	return args.Get(0).(types.LogConfig)
}

func (m *MockConfigManager) GetRedisDSN() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigManager) GetDatabaseConfig() types.DatabaseConfig {
	args := m.Called()
	return args.Get(0).(types.DatabaseConfig)
}

func (m *MockConfigManager) GetEncryptionKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigManager) GetEffectiveServerConfig() types.ServerConfig {
	args := m.Called()
	return args.Get(0).(types.ServerConfig)
}

func (m *MockConfigManager) ReloadConfig() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigManager) Validate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigManager) DisplayServerConfig() {
	m.Called()
}

func TestNewStore(t *testing.T) {
	t.Run("returns memory store when no Redis DSN", func(t *testing.T) {
		mockConfig := &MockConfigManager{}
		mockConfig.On("GetRedisDSN").Return("")

		store, err := NewStore(mockConfig)

		assert.NoError(t, err)
		assert.NotNil(t, store)
		assert.IsType(t, &MemoryStore{}, store)
		mockConfig.AssertExpectations(t)
	})

	t.Run("returns error for invalid Redis DSN", func(t *testing.T) {
		mockConfig := &MockConfigManager{}
		mockConfig.On("GetRedisDSN").Return("invalid-dsn")

		store, err := NewStore(mockConfig)

		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "failed to parse redis DSN")
		mockConfig.AssertExpectations(t)
	})

	t.Run("returns error for unreachable Redis", func(t *testing.T) {
		mockConfig := &MockConfigManager{}
		// Use a valid DSN format but unreachable address
		mockConfig.On("GetRedisDSN").Return("redis://localhost:9999")

		store, err := NewStore(mockConfig)

		assert.Error(t, err)
		assert.Nil(t, store)
		assert.Contains(t, err.Error(), "failed to connect to redis")
		mockConfig.AssertExpectations(t)
	})

	// Note: Testing successful Redis connection would require a real Redis instance
	// This is typically done in integration tests rather than unit tests
}
