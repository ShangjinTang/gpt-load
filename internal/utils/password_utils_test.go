package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeriveAESKey(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{"normal password", "mypassword123"},
		{"empty password", ""},
		{"long password", "this-is-a-very-long-password-for-testing-purposes"},
		{"special characters", "!@#$%^&*()"},
		{"unicode", "你好世界"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := DeriveAESKey(tc.password)
			assert.Equal(t, 32, len(key)) // AES-256 key should be 32 bytes
		})
	}
}

func TestDeriveAESKey_Consistency(t *testing.T) {
	password := "test-password"
	key1 := DeriveAESKey(password)
	key2 := DeriveAESKey(password)

	assert.Equal(t, key1, key2, "Same password should produce same key")
}

func TestDeriveAESKey_Different(t *testing.T) {
	key1 := DeriveAESKey("password1")
	key2 := DeriveAESKey("password2")

	assert.NotEqual(t, key1, key2, "Different passwords should produce different keys")
}

func TestValidatePasswordStrength(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		context  string
	}{
		{"short password", "short", "TEST"},
		{"weak pattern", "password123", "TEST"},
		{"strong password", "MyStr0ngP@ssw0rd!", "TEST"},
		{"empty password", "", "TEST"},
		{"very long password", "this-is-a-very-long-and-complex-password-that-should-be-strong", "TEST"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This function logs warnings but doesn't return anything
			// We just test that it doesn't panic
			assert.NotPanics(t, func() {
				ValidatePasswordStrength(tc.password, tc.context)
			})
		})
	}
}
