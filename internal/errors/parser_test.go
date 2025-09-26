package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUpstreamError(t *testing.T) {
	testCases := []struct {
		name     string
		body     []byte
		expected string
	}{
		{
			name:     "standard OpenAI format",
			body:     []byte(`{"error": {"message": "Invalid API key provided"}}`),
			expected: "Invalid API key provided",
		},
		{
			name:     "vendor specific format",
			body:     []byte(`{"error_msg": "Rate limit exceeded"}`),
			expected: "Rate limit exceeded",
		},
		{
			name:     "simple error format",
			body:     []byte(`{"error": "Bad request format"}`),
			expected: "Bad request format",
		},
		{
			name:     "root message format",
			body:     []byte(`{"message": "Service unavailable"}`),
			expected: "Service unavailable",
		},
		{
			name:     "invalid JSON",
			body:     []byte(`invalid json content`),
			expected: "invalid json content",
		},
		{
			name:     "empty JSON",
			body:     []byte(`{}`),
			expected: "{}",
		},
		{
			name:     "nested error structure",
			body:     []byte(`{"error": {"message": "Model not found", "type": "invalid_request_error"}}`),
			expected: "Model not found",
		},
		{
			name:     "empty message in standard format",
			body:     []byte(`{"error": {"message": ""}}`),
			expected: `{"error": {"message": ""}}`,
		},
		{
			name:     "whitespace only message",
			body:     []byte(`{"error": {"message": "   "}}`),
			expected: `{"error": {"message": "   "}}`,
		},
		{
			name:     "multiple formats present - standard wins",
			body:     []byte(`{"error": {"message": "Standard format"}, "error_msg": "Vendor format"}`),
			expected: "Standard format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseUpstreamError(tc.body)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseUpstreamError_LongMessages(t *testing.T) {
	// Test message truncation
	longMessage := make([]byte, 3000)
	for i := range longMessage {
		longMessage[i] = 'a'
	}

	body := `{"error": {"message": "` + string(longMessage) + `"}}`
	result := ParseUpstreamError([]byte(body))

	// Should be truncated to maxErrorBodyLength (2048)
	assert.LessOrEqual(t, len(result), 2048)
}

func TestParseUpstreamError_EdgeCases(t *testing.T) {
	t.Run("empty body", func(t *testing.T) {
		result := ParseUpstreamError([]byte{})
		assert.Equal(t, "", result)
	})

	t.Run("null values", func(t *testing.T) {
		body := []byte(`{"error": {"message": null}}`)
		result := ParseUpstreamError(body)
		assert.Equal(t, string(body), result) // Should return raw body when message is null
	})

	t.Run("numeric message", func(t *testing.T) {
		body := []byte(`{"error": {"message": 123}}`)
		result := ParseUpstreamError(body)
		assert.Equal(t, string(body), result) // Should return raw body when message is not string
	})

	t.Run("array message", func(t *testing.T) {
		body := []byte(`{"error": {"message": ["error1", "error2"]}}`)
		result := ParseUpstreamError(body)
		assert.Equal(t, string(body), result) // Should return raw body when message is array
	})
}

func TestTruncateString(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "short string",
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
			name:      "long string truncated",
			input:     "hello world",
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
		{
			name:      "unicode string",
			input:     "你好世界",
			maxLength: 6,
			expected:  "你好",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := truncateString(tc.input, tc.maxLength)
			assert.Equal(t, tc.expected, result)
			assert.LessOrEqual(t, len(result), tc.maxLength)
		})
	}
}

func TestParseUpstreamError_RealWorldExamples(t *testing.T) {
	testCases := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name: "OpenAI API error",
			body: `{
				"error": {
					"message": "You exceeded your current quota, please check your plan and billing details.",
					"type": "insufficient_quota",
					"param": null,
					"code": "insufficient_quota"
				}
			}`,
			expected: "You exceeded your current quota, please check your plan and billing details.",
		},
		{
			name: "Claude API error",
			body: `{
				"error": {
					"type": "authentication_error",
					"message": "x-api-key header is required"
				}
			}`,
			expected: "x-api-key header is required",
		},
		{
			name: "Gemini API error",
			body: `{
				"error": {
					"code": 400,
					"message": "API key not valid. Please pass a valid API key.",
					"status": "INVALID_ARGUMENT"
				}
			}`,
			expected: "API key not valid. Please pass a valid API key.",
		},
		{
			name:     "Simple error message",
			body:     `{"error": "Model not found"}`,
			expected: "Model not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseUpstreamError([]byte(tc.body))
			assert.Equal(t, tc.expected, result)
		})
	}
}
