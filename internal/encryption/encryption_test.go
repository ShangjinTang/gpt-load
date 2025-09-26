package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Run("with encryption key", func(t *testing.T) {
		service, err := NewService("test-key-123")
		require.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("without encryption key", func(t *testing.T) {
		service, err := NewService("")
		require.NoError(t, err)
		assert.NotNil(t, service)
	})
}

func TestEncryptionService_Behavior(t *testing.T) {
	t.Run("with key should encrypt/decrypt", func(t *testing.T) {
		service, err := NewService("test-key")
		require.NoError(t, err)

		plaintext := "test data"
		encrypted, err := service.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, encrypted) // Should be encrypted

		decrypted, err := service.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("without key should passthrough", func(t *testing.T) {
		service, err := NewService("")
		require.NoError(t, err)

		plaintext := "test data"
		encrypted, err := service.Encrypt(plaintext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, encrypted) // Should passthrough

		decrypted, err := service.Decrypt(plaintext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestEncryptionService_EncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		plaintext string
	}{
		{
			name:      "basic encryption",
			key:       "test-key-123",
			plaintext: "hello world",
		},
		{
			name:      "empty plaintext",
			key:       "test-key-456",
			plaintext: "",
		},
		{
			name:      "long plaintext",
			key:       "test-key-789",
			plaintext: "this is a very long plaintext that should be encrypted and decrypted correctly without any issues",
		},
		{
			name:      "special characters",
			key:       "test-key-special",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode characters",
			key:       "test-key-unicode",
			plaintext: "你好世界 🌍 こんにちは世界",
		},
		{
			name:      "json data",
			key:       "test-key-json",
			plaintext: `{"name":"test","value":123,"nested":{"key":"value"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewService(tc.key)
		require.NoError(t, err)

			// 加密
			encrypted, err := service.Encrypt(tc.plaintext)
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tc.plaintext, encrypted)

			// 解密
			decrypted, err := service.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tc.plaintext, decrypted)
		})
	}
}

func TestEncryptionService_EncryptDecrypt_Disabled(t *testing.T) {
	service, err := NewService("")
	require.NoError(t, err)
	plaintext := "test data"

	// 当加密未启用时，Encrypt应该返回原始文本
	encrypted, err := service.Encrypt(plaintext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, encrypted)

	// 当加密未启用时，Decrypt应该返回原始文本
	decrypted, err := service.Decrypt(plaintext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptionService_DecryptInvalidData(t *testing.T) {
	service, err := NewService("test-key")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		encrypted string
	}{
		{
			name:      "invalid base64",
			encrypted: "invalid-base64-data!@#",
		},
		{
			name:      "valid base64 but invalid encrypted data",
			encrypted: "dGhpcyBpcyBub3QgZW5jcnlwdGVkIGRhdGE=", // "this is not encrypted data" in base64
		},
		{
			name:      "empty string",
			encrypted: "",
		},
		{
			name:      "too short data",
			encrypted: "YWJj", // "abc" in base64, too short for encrypted data
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.Decrypt(tc.encrypted)
			assert.Error(t, err)
		})
	}
}

func TestEncryptionService_EncryptionConsistency(t *testing.T) {
	service, err := NewService("consistent-key")
	require.NoError(t, err)
	plaintext := "test data for consistency"

	// 多次加密同一个文本应该产生不同的密文（因为使用了随机IV）
	encrypted1, err := service.Encrypt(plaintext)
	require.NoError(t, err)

	encrypted2, err := service.Encrypt(plaintext)
	require.NoError(t, err)

	// 密文应该不同（因为IV不同）
	assert.NotEqual(t, encrypted1, encrypted2)

	// 但解密后应该得到相同的明文
	decrypted1, err := service.Decrypt(encrypted1)
	require.NoError(t, err)

	decrypted2, err := service.Decrypt(encrypted2)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
}

func TestEncryptionService_DifferentKeys(t *testing.T) {
	service1, err := NewService("key-1")
	require.NoError(t, err)
	service2, err := NewService("key-2")
	require.NoError(t, err)
	plaintext := "secret data"

	// 用第一个密钥加密
	encrypted, err := service1.Encrypt(plaintext)
	require.NoError(t, err)

	// 用第二个密钥解密应该失败
	_, err = service2.Decrypt(encrypted)
	assert.Error(t, err)

	// 用正确的密钥解密应该成功
	decrypted, err := service1.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptionService_KeyDerivation(t *testing.T) {
	// 测试相同的密码应该产生相同的派生密钥
	service1, err := NewService("same-password")
	require.NoError(t, err)
	service2, err := NewService("same-password")
	require.NoError(t, err)
	plaintext := "test data"

	// 用第一个服务加密
	encrypted, err := service1.Encrypt(plaintext)
	require.NoError(t, err)

	// 用第二个服务解密应该成功（因为使用相同的密码）
	decrypted, err := service2.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptionService_EdgeCases(t *testing.T) {

	t.Run("very long key", func(t *testing.T) {
		longKeyService, err := NewService("this-is-a-very-long-encryption-key-that-should-be-handled-correctly-by-the-key-derivation-function")
		require.NoError(t, err)
		plaintext := "test with long key"

		encrypted, err := longKeyService.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := longKeyService.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("short key", func(t *testing.T) {
		shortKeyService, err := NewService("a")
		require.NoError(t, err)
		plaintext := "test with short key"

		encrypted, err := shortKeyService.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := shortKeyService.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("key with special characters", func(t *testing.T) {
		specialKeyService, err := NewService("key!@#$%^&*()_+-=[]{}|;':\",./<>?")
		require.NoError(t, err)
		plaintext := "test with special key"

		encrypted, err := specialKeyService.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := specialKeyService.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}

func TestEncryptionService_ConcurrentAccess(t *testing.T) {
	service, err := NewService("concurrent-test-key")
	require.NoError(t, err)
	plaintext := "concurrent test data"

	// 并发测试
	const numGoroutines = 10
	const numOperations = 100

	results := make(chan error, numGoroutines*numOperations*2) // *2 for encrypt and decrypt

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// 加密
				encrypted, err := service.Encrypt(plaintext)
				if err != nil {
					results <- err
					continue
				}

				// 解密
				decrypted, err := service.Decrypt(encrypted)
				if err != nil {
					results <- err
					continue
				}

				if decrypted != plaintext {
					results <- assert.AnError
					continue
				}

				results <- nil // success
			}
		}(i)
	}

	// 收集结果
	for i := 0; i < numGoroutines*numOperations; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

func BenchmarkEncryptionService_Encrypt(b *testing.B) {
	service, err := NewService("benchmark-key")
	if err != nil {
		b.Fatal(err)
	}
	plaintext := "benchmark test data that is somewhat longer to simulate real usage"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Encrypt(plaintext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptionService_Decrypt(b *testing.B) {
	service, err := NewService("benchmark-key")
	if err != nil {
		b.Fatal(err)
	}
	plaintext := "benchmark test data that is somewhat longer to simulate real usage"

	// 预先加密数据
	encrypted, err := service.Encrypt(plaintext)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Decrypt(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptionService_EncryptDecrypt(b *testing.B) {
	service, err := NewService("benchmark-key")
	if err != nil {
		b.Fatal(err)
	}
	plaintext := "benchmark test data that is somewhat longer to simulate real usage"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, err := service.Encrypt(plaintext)
		if err != nil {
			b.Fatal(err)
		}

		_, err = service.Decrypt(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}
