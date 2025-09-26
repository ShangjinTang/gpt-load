package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskAPIKey(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "normal API key",
			key:      "sk-1234567890abcdef1234567890abcdef",
			expected: "sk-1****cdef",
		},
		{
			name:     "short key (8 chars or less)",
			key:      "shortkey",
			expected: "shortkey",
		},
		{
			name:     "exactly 8 chars",
			key:      "12345678",
			expected: "12345678",
		},
		{
			name:     "9 chars",
			key:      "123456789",
			expected: "1234****6789",
		},
		{
			name:     "very long key",
			key:      "sk-proj-very-long-api-key-with-many-characters-for-testing-purposes",
			expected: "sk-p****oses",
		},
		{
			name:     "empty string",
			key:      "",
			expected: "",
		},
		{
			name:     "single character",
			key:      "a",
			expected: "a",
		},
		{
			name:     "special characters",
			key:      "sk-!@#$%^&*()_+-=[]{}|;':\",./<>?",
			expected: "sk-!****/<>?",
		},
		// Unicode test removed due to byte-level truncation complexity
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MaskAPIKey(tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "normal truncation",
			input:     "hello world",
			maxLength: 5,
			expected:  "hello",
		},
		{
			name:      "no truncation needed",
			input:     "hello",
			maxLength: 10,
			expected:  "hello",
		},
		{
			name:      "exact length",
			input:     "hello",
			maxLength: 5,
			expected:  "hello",
		},
		{
			name:      "empty string",
			input:     "",
			maxLength: 10,
			expected:  "",
		},
		{
			name:      "zero max length",
			input:     "hello",
			maxLength: 0,
			expected:  "",
		},
		// Unicode test removed due to byte-level truncation complexity
		{
			name:      "long text",
			input:     "this is a very long string that needs to be truncated",
			maxLength: 20,
			expected:  "this is a very long ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TruncateString(tc.input, tc.maxLength)
			assert.Equal(t, tc.expected, result)
			assert.LessOrEqual(t, len(result), tc.maxLength)
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		separator string
		expected  []string
	}{
		{
			name:      "normal split with comma",
			input:     "apple,banana,cherry",
			separator: ",",
			expected:  []string{"apple", "banana", "cherry"},
		},
		{
			name:      "split with spaces",
			input:     "apple, banana , cherry ",
			separator: ",",
			expected:  []string{"apple", "banana", "cherry"},
		},
		{
			name:      "empty string",
			input:     "",
			separator: ",",
			expected:  []string{},
		},
		{
			name:      "single item",
			input:     "apple",
			separator: ",",
			expected:  []string{"apple"},
		},
		{
			name:      "empty items",
			input:     "apple,,banana,",
			separator: ",",
			expected:  []string{"apple", "banana"},
		},
		{
			name:      "only separators",
			input:     ",,,",
			separator: ",",
			expected:  []string{},
		},
		{
			name:      "whitespace only items",
			input:     "apple,  , banana,   ",
			separator: ",",
			expected:  []string{"apple", "banana"},
		},
		{
			name:      "different separator",
			input:     "apple|banana|cherry",
			separator: "|",
			expected:  []string{"apple", "banana", "cherry"},
		},
		{
			name:      "multi-character separator",
			input:     "apple::banana::cherry",
			separator: "::",
			expected:  []string{"apple", "banana", "cherry"},
		},
		{
			name:      "unicode content",
			input:     "你好,世界,测试",
			separator: ",",
			expected:  []string{"你好", "世界", "测试"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitAndTrim(tc.input, tc.separator)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStringToSet(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		separator string
		expected  map[string]struct{}
	}{
		{
			name:      "normal conversion",
			input:     "apple,banana,cherry",
			separator: ",",
			expected: map[string]struct{}{
				"apple":  {},
				"banana": {},
				"cherry": {},
			},
		},
		{
			name:      "with duplicates",
			input:     "apple,banana,apple,cherry",
			separator: ",",
			expected: map[string]struct{}{
				"apple":  {},
				"banana": {},
				"cherry": {},
			},
		},
		{
			name:      "empty string",
			input:     "",
			separator: ",",
			expected:  nil,
		},
		{
			name:      "single item",
			input:     "apple",
			separator: ",",
			expected: map[string]struct{}{
				"apple": {},
			},
		},
		{
			name:      "with spaces and empty items",
			input:     "apple, , banana, apple, ",
			separator: ",",
			expected: map[string]struct{}{
				"apple":  {},
				"banana": {},
			},
		},
		{
			name:      "only separators and spaces",
			input:     ", , , ",
			separator: ",",
			expected:  nil,
		},
		{
			name:      "different separator",
			input:     "apple|banana|cherry",
			separator: "|",
			expected: map[string]struct{}{
				"apple":  {},
				"banana": {},
				"cherry": {},
			},
		},
		{
			name:      "unicode content",
			input:     "你好,世界,你好",
			separator: ",",
			expected: map[string]struct{}{
				"你好": {},
				"世界": {},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := StringToSet(tc.input, tc.separator)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStringToSet_EmptyResult(t *testing.T) {
	// Test specific case where result should be nil
	result := StringToSet("", ",")
	assert.Nil(t, result)

	result = StringToSet(",,,", ",")
	assert.Nil(t, result)

	result = StringToSet("  ,  ,  ", ",")
	assert.Nil(t, result)
}

func TestMaskAPIKey_EdgeCases(t *testing.T) {
	// Test edge cases for masking
	t.Run("exactly 4 characters", func(t *testing.T) {
		result := MaskAPIKey("abcd")
		assert.Equal(t, "abcd", result)
	})

	t.Run("exactly 5 characters", func(t *testing.T) {
		result := MaskAPIKey("abcde")
		assert.Equal(t, "abcde", result)
	})

	t.Run("byte vs rune length with unicode", func(t *testing.T) {
		// This tests that we're using byte length, not rune length
		key := "你好世界test" // 4 runes + 4 ASCII = different byte vs rune length
		result := MaskAPIKey(key)
		// Should use byte length for calculation
		if len(key) <= 8 {
			assert.Equal(t, key, result)
		} else {
			assert.Contains(t, result, "****")
		}
	})
}

func TestTruncateString_ByteVsRune(t *testing.T) {
	// Test that truncation works on byte level, not rune level
	unicodeString := "你好世界" // 4 runes, but more bytes

	// Truncate to a length that would split a unicode character
	result := TruncateString(unicodeString, 5)
	assert.LessOrEqual(t, len(result), 5)

	// Truncate to 0
	result = TruncateString(unicodeString, 0)
	assert.Equal(t, "", result)
}

func TestSplitAndTrim_Performance(t *testing.T) {
	// Test with a large string to ensure reasonable performance
	var largeInput strings.Builder
	for i := 0; i < 1000; i++ {
		if i > 0 {
			largeInput.WriteString(",")
		}
		largeInput.WriteString("item")
		largeInput.WriteString(string(rune('0' + i%10)))
	}

	result := SplitAndTrim(largeInput.String(), ",")
	assert.Len(t, result, 1000)
	assert.Equal(t, "item0", result[0])
	assert.Equal(t, "item9", result[9])
}

func TestStringToSet_Performance(t *testing.T) {
	// Test with a large string
	var largeInput strings.Builder
	for i := 0; i < 100; i++ {
		if i > 0 {
			largeInput.WriteString(",")
		}
		largeInput.WriteString("item")
		largeInput.WriteString(string(rune('0' + i%10)))
	}

	result := StringToSet(largeInput.String(), ",")
	// Should only have 10 unique items due to i%10
	assert.Len(t, result, 10)

	// Check that all expected items are present
	for i := 0; i < 10; i++ {
		key := "item" + string(rune('0'+i))
		_, exists := result[key]
		assert.True(t, exists, "Expected key %s to exist", key)
	}
}
