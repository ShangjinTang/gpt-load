package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUnCounted(t *testing.T) {
	testCases := []struct {
		name         string
		errorMessage string
		expected     bool
	}{
		// Uncounted errors (should not affect key statistics)
		{"resource exhausted", "resource has been exhausted", true},
		{"Resource Exhausted", "Resource has been exhausted", true},
		{"RESOURCE EXHAUSTED", "RESOURCE HAS BEEN EXHAUSTED", true},
		{"reduce message length", "please reduce the length of the messages", true},
		{"Reduce Message Length", "Please reduce the length of the messages", true},
		{"REDUCE MESSAGE LENGTH", "PLEASE REDUCE THE LENGTH OF THE MESSAGES", true},

		// Counted errors (should affect key statistics)
		{"invalid api key", "invalid api key", false},
		{"rate limit exceeded", "rate limit exceeded", false},
		{"insufficient quota", "insufficient quota", false},
		{"model not found", "model not found", false},
		{"server error", "internal server error", false},
		{"bad request", "bad request", false},
		{"timeout", "request timeout", false},
		{"network error", "network connection failed", false},
		{"context canceled", "context canceled", false}, // Not in the actual uncounted list
		{"client disconnected", "client disconnected", false}, // Not in the actual uncounted list

		// Edge cases
		{"empty string", "", false},
		{"whitespace", "   ", false},
		{"partial match", "resource", false}, // Should not match partial words
		{"contains uncounted", "error: resource has been exhausted occurred", true}, // Should match if contains the phrase
		{"multiple errors", "resource has been exhausted and invalid api key", true}, // Should match if any uncounted error is present

		// Case variations
		{"mixed case resource", "Resource Has Been Exhausted", true},
		{"mixed case reduce", "Please Reduce The Length Of The Messages", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsUnCounted(tc.errorMessage)
			assert.Equal(t, tc.expected, result, "Error message: %s", tc.errorMessage)
		})
	}
}

func TestIsUnCountedCaseSensitivity(t *testing.T) {
	// Test various case combinations for each uncounted error type
	uncountedPatterns := []string{
		"resource has been exhausted",
		"please reduce the length of the messages",
	}

	caseCombinations := []func(string) string{
		func(s string) string { return s },                    // original
		func(s string) string { return strings.ToUpper(s) },  // UPPERCASE
		func(s string) string { return strings.ToLower(s) },  // lowercase
		func(s string) string { return strings.Title(s) },    // Title Case
	}

	for _, pattern := range uncountedPatterns {
		for i, caseFunc := range caseCombinations {
			t.Run(fmt.Sprintf("%s_case_%d", strings.ReplaceAll(pattern, " ", "_"), i), func(t *testing.T) {
				transformed := caseFunc(pattern)
				assert.True(t, IsUnCounted(transformed), "Should be uncounted: %s", transformed)
			})
		}
	}
}

func TestIsUnCountedWithContext(t *testing.T) {
	// Test uncounted errors within larger error messages
	testCases := []struct {
		name    string
		message string
		expected bool
	}{
		{"prefix resource", "error: resource has been exhausted", true},
		{"suffix resource", "resource has been exhausted: additional info", true},
		{"middle resource", "request failed: resource has been exhausted during processing", true},
		{"prefix reduce", "timeout: please reduce the length of the messages", true},
		{"suffix reduce", "please reduce the length of the messages while waiting", true},
		{"middle reduce", "operation failed: please reduce the length of the messages in handler", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsUnCounted(tc.message)
			assert.Equal(t, tc.expected, result, "Message: %s", tc.message)
		})
	}
}

func TestIsUnCountedEdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		assert.False(t, IsUnCounted(""))
	})

	t.Run("whitespace only", func(t *testing.T) {
		assert.False(t, IsUnCounted("   "))
		assert.False(t, IsUnCounted("\t\n"))
	})

	t.Run("partial matches should not count", func(t *testing.T) {
		// These should NOT be considered uncounted
		assert.False(t, IsUnCounted("resource"))
		assert.False(t, IsUnCounted("exhausted"))
		assert.False(t, IsUnCounted("reduce"))
		assert.False(t, IsUnCounted("length"))
		assert.False(t, IsUnCounted("messages"))
	})

	t.Run("similar but different phrases", func(t *testing.T) {
		// These should NOT be considered uncounted
		assert.False(t, IsUnCounted("resource exhausted"))
		assert.False(t, IsUnCounted("reduce message length"))
		assert.False(t, IsUnCounted("resource has exhausted"))
		assert.False(t, IsUnCounted("please reduce length"))
	})

	t.Run("multiple uncounted errors", func(t *testing.T) {
		message := "resource has been exhausted and please reduce the length of the messages"
		assert.True(t, IsUnCounted(message))
	})

	t.Run("mixed counted and uncounted", func(t *testing.T) {
		message := "invalid api key but resource has been exhausted"
		assert.True(t, IsUnCounted(message)) // Should be true if ANY uncounted error is present
	})
}

func TestIsUnCountedPerformance(t *testing.T) {
	// Test performance with various message lengths
	testMessages := []string{
		"resource has been exhausted",
		"this is a medium length error message that contains resource has been exhausted in the middle",
		"this is a very long error message that goes on and on and on with lots of details and explanations and resource has been exhausted appears somewhere in this very verbose error description that might be returned by some systems",
	}

	for _, msg := range testMessages {
		t.Run(fmt.Sprintf("length_%d", len(msg)), func(t *testing.T) {
			// Just ensure it works correctly regardless of message length
			result := IsUnCounted(msg)
			assert.True(t, result, "Should detect uncounted error in message of length %d", len(msg))
		})
	}
}

func TestUnCountedErrorsDocumentation(t *testing.T) {
	// This test serves as documentation for what errors are considered "uncounted"
	// These errors typically indicate resource exhaustion or message length issues
	// that should not count against the API key's failure statistics

	uncountedErrors := map[string]string{
		"resource has been exhausted":                 "Resource exhaustion - typically temporary server-side issue",
		"please reduce the length of the messages":   "Message too long - client should adjust request size",
	}

	for errorMsg, description := range uncountedErrors {
		t.Run(strings.ReplaceAll(errorMsg, " ", "_"), func(t *testing.T) {
			assert.True(t, IsUnCounted(errorMsg), "Error should be uncounted: %s (%s)", errorMsg, description)
		})
	}

	// Document what errors ARE counted (examples)
	countedErrors := map[string]string{
		"invalid api key":         "Authentication failure - key issue",
		"rate limit exceeded":     "Server-side rate limiting - key usage issue",
		"insufficient quota":      "Billing/quota issue - key limitation",
		"model not found":         "Invalid request - model specification issue",
		"internal server error":   "Server-side error - not client's fault but counts toward key stats",
		"bad request":            "Client error in request format - counts as usage",
	}

	for errorMsg, description := range countedErrors {
		t.Run(errorMsg+"_counted", func(t *testing.T) {
			assert.False(t, IsUnCounted(errorMsg), "Error should be counted: %s (%s)", errorMsg, description)
		})
	}
}
