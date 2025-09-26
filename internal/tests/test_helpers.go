package tests

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gpt-load/internal/models"
	"gpt-load/internal/store"
)

// SetupTestDB 创建一个用于测试的内存数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.Group{},
		&models.APIKey{},
		&models.RequestLog{},
		&models.SystemSetting{},
		&models.Policy{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// MockMemoryStore 模拟内存存储
type MockMemoryStore struct {
	data map[string][]byte
}

func NewMockMemoryStore() *MockMemoryStore {
	return &MockMemoryStore{
		data: make(map[string][]byte),
	}
}

func (m *MockMemoryStore) Set(key string, value []byte, ttl time.Duration) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = value
	return nil
}

func (m *MockMemoryStore) Get(key string) ([]byte, error) {
	if m.data == nil {
		return nil, store.ErrNotFound
	}
	value, exists := m.data[key]
	if !exists {
		return nil, store.ErrNotFound
	}
	return value, nil
}

func (m *MockMemoryStore) Delete(key string) error {
	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *MockMemoryStore) Del(keys ...string) error {
	if m.data != nil {
		for _, key := range keys {
			delete(m.data, key)
		}
	}
	return nil
}

func (m *MockMemoryStore) Exists(key string) (bool, error) {
	if m.data == nil {
		return false, nil
	}
	_, exists := m.data[key]
	return exists, nil
}

func (m *MockMemoryStore) SetNX(key string, value []byte, ttl time.Duration) (bool, error) {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	if _, exists := m.data[key]; exists {
		return false, nil
	}
	m.data[key] = value
	return true, nil
}

func (m *MockMemoryStore) HSet(key string, values map[string]any) error {
	return nil
}

func (m *MockMemoryStore) HGetAll(key string) (map[string]string, error) {
	return make(map[string]string), nil
}

func (m *MockMemoryStore) HIncrBy(key, field string, incr int64) (int64, error) {
	return incr, nil
}

func (m *MockMemoryStore) LPush(key string, values ...any) error {
	return nil
}

func (m *MockMemoryStore) LRem(key string, count int64, value any) error {
	return nil
}

func (m *MockMemoryStore) Rotate(key string) (string, error) {
	return "", store.ErrNotFound
}

func (m *MockMemoryStore) SAdd(key string, members ...any) error {
	return nil
}

func (m *MockMemoryStore) SPopN(key string, count int64) ([]string, error) {
	return []string{}, nil
}

func (m *MockMemoryStore) Close() error {
	return nil
}

func (m *MockMemoryStore) Publish(channel string, message []byte) error {
	return nil
}

func (m *MockMemoryStore) Subscribe(channel string) (store.Subscription, error) {
	return &MockSubscription{}, nil
}

func (m *MockMemoryStore) Clear() error {
	m.data = make(map[string][]byte)
	return nil
}

// MockSubscription 模拟订阅
type MockSubscription struct{}

func (m *MockSubscription) Channel() <-chan *store.Message {
	ch := make(chan *store.Message)
	close(ch)
	return ch
}

func (m *MockSubscription) Close() error {
	return nil
}

// MockGroupManager 模拟组管理器
type MockGroupManager struct{}

func (m *MockGroupManager) Invalidate() error {
	return nil
}
