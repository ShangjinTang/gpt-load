package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyStates(t *testing.T) {
	// Test that all key states are defined
	assert.Equal(t, "pending", KeyStatusPending)
	assert.Equal(t, "active", KeyStatusActive)
	assert.Equal(t, "degraded", KeyStatusDegraded)
	assert.Equal(t, "disabled", KeyStatusDisabled)
	assert.Equal(t, "invalid", KeyStatusInvalid)
}

func TestKeyStateMachine_IsValidState(t *testing.T) {
	machine := NewKeyStateMachine()

	testCases := []struct {
		status   string
		expected bool
	}{
		{KeyStatusPending, true},
		{KeyStatusActive, true},
		{KeyStatusDegraded, true},
		{KeyStatusDisabled, true},
		{KeyStatusInvalid, true},
		{"unknown", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run("status_"+tc.status, func(t *testing.T) {
			result := machine.IsValidState(tc.status)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestKeyStateMachine_TransitionState(t *testing.T) {
	machine := NewKeyStateMachine()

	testCases := []struct {
		name                string
		currentStatus       string
		isSuccess           bool
		consecutiveFailures int64
		expectedStatus      string
	}{
		// Pending state transitions
		{"pending_success", KeyStatusPending, true, 0, KeyStatusActive},
		{"pending_failure_low", KeyStatusPending, false, 1, KeyStatusPending},
		{"pending_failure_high", KeyStatusPending, false, 3, KeyStatusDisabled},

		// Active state transitions
		{"active_success", KeyStatusActive, true, 0, KeyStatusActive},
		{"active_failure", KeyStatusActive, false, 1, KeyStatusDegraded},

		// Degraded state transitions
		{"degraded_success", KeyStatusDegraded, true, 0, KeyStatusActive},
		{"degraded_failure_low", KeyStatusDegraded, false, 2, KeyStatusDegraded},
		{"degraded_failure_high", KeyStatusDegraded, false, 3, KeyStatusDisabled},

		// Disabled state transitions
		{"disabled_success", KeyStatusDisabled, true, 0, KeyStatusDegraded},
		{"disabled_failure", KeyStatusDisabled, false, 5, KeyStatusDisabled},

		// Invalid state (no transitions)
		{"invalid_success", KeyStatusInvalid, true, 0, KeyStatusInvalid},
		{"invalid_failure", KeyStatusInvalid, false, 1, KeyStatusInvalid},

		// Unknown state
		{"unknown_state", "unknown", false, 1, KeyStatusPending},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := machine.TransitionState(tc.currentStatus, tc.isSuccess, tc.consecutiveFailures)
			assert.Equal(t, tc.expectedStatus, result)
		})
	}
}

func TestKeyStateMachine_CalculateBackoffDuration(t *testing.T) {
	machine := NewKeyStateMachine()

	testCases := []struct {
		level    int
		expected time.Duration
	}{
		{0, 1 * time.Minute},      // Minimum 1 minute
		{1, 2 * time.Minute},      // 2^1 = 2 minutes
		{2, 4 * time.Minute},      // 2^2 = 4 minutes
		{3, 8 * time.Minute},      // 2^3 = 8 minutes
		{4, 16 * time.Minute},     // 2^4 = 16 minutes
		{5, 30 * time.Minute},     // Cap at 30 minutes (2^5 = 32, but capped)
		{10, 30 * time.Minute},    // Cap at 30 minutes
		{-1, 1 * time.Minute},     // Negative level, minimum 1 minute
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("level_%d", tc.level), func(t *testing.T) {
			duration := machine.CalculateBackoffDuration(tc.level)
			assert.Equal(t, tc.expected, duration)
		})
	}
}

func TestNewKeyStateMachine(t *testing.T) {
	machine := NewKeyStateMachine()
	assert.NotNil(t, machine)
}
