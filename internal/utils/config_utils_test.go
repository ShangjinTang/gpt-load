package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvOrDefault(t *testing.T) {
	testCases := []struct {
		name         string
		key          string
		defaultValue string
		setValue     string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			setValue:     "actual_value",
			expected:     "actual_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default_value",
			setValue:     "",
			expected:     "default_value",
		},
		{
			name:         "empty default value",
			key:          "ANOTHER_NON_EXISTENT_VAR",
			defaultValue: "",
			setValue:     "",
			expected:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable if needed
			if tc.setValue != "" {
				t.Setenv(tc.key, tc.setValue)
			}

			result := GetEnvOrDefault(tc.key, tc.defaultValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseInteger(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			value:        "123",
			defaultValue: 456,
			expected:     123,
		},
		{
			name:         "invalid integer",
			value:        "not_a_number",
			defaultValue: 789,
			expected:     789,
		},
		{
			name:         "empty string",
			value:        "",
			defaultValue: 100,
			expected:     100,
		},
		{
			name:         "negative number",
			value:        "-42",
			defaultValue: 0,
			expected:     -42,
		},
		{
			name:         "zero",
			value:        "0",
			defaultValue: 999,
			expected:     0,
		},
		{
			name:         "very large number",
			value:        "999999999",
			defaultValue: 1,
			expected:     999999999,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseInteger(tc.value, tc.defaultValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseBoolean(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true string",
			value:        "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false string",
			value:        "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "1 as true",
			value:        "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "0 as false",
			value:        "0",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "invalid boolean",
			value:        "not_a_bool",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "empty string",
			value:        "",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "mixed case true",
			value:        "True",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "mixed case false",
			value:        "False",
			defaultValue: true,
			expected:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseBoolean(tc.value, tc.defaultValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}
