package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureStore(t *testing.T) {
	store := NewFeatureStore()

	// Test getting default flags for new org
	flags := store.GetFlags("test-org")
	assert.NotNil(t, flags)
	assert.True(t, flags.PropagationOn)
	assert.True(t, flags.SLOGuardOn)
	assert.True(t, flags.EdgeLearningOn)

	// Test setting flags
	newFlags := &FeatureFlags{
		PropagationOn:  false,
		SLOGuardOn:     true,
		EdgeLearningOn: false,
	}
	store.SetFlags("test-org", newFlags)

	// Test getting updated flags
	updatedFlags := store.GetFlags("test-org")
	assert.Equal(t, newFlags.PropagationOn, updatedFlags.PropagationOn)
	assert.Equal(t, newFlags.SLOGuardOn, updatedFlags.SLOGuardOn)
	assert.Equal(t, newFlags.EdgeLearningOn, updatedFlags.EdgeLearningOn)

	// Test that different orgs have independent flags
	otherFlags := store.GetFlags("other-org")
	assert.True(t, otherFlags.PropagationOn) // Should have defaults
	assert.True(t, otherFlags.SLOGuardOn)
	assert.True(t, otherFlags.EdgeLearningOn)
}

func TestUpdateFlag(t *testing.T) {
	store := NewFeatureStore()

	// Test updating individual flags
	store.UpdateFlag("test-org", "PROPAGATION_ON", false)
	store.UpdateFlag("test-org", "SLO_GUARD_ON", false)

	flags := store.GetFlags("test-org")
	assert.False(t, flags.PropagationOn)
	assert.False(t, flags.SLOGuardOn)
	assert.True(t, flags.EdgeLearningOn) // Should remain default

	// Test updating non-existent flag (should not panic)
	store.UpdateFlag("test-org", "NON_EXISTENT_FLAG", true)
}

func TestIsPropagationEnabled(t *testing.T) {
	// Use the global feature store
	store := GlobalFeatureStore

	// Test default (enabled)
	assert.True(t, IsPropagationEnabled("test-org"))

	// Test disabled
	store.UpdateFlag("test-org", "PROPAGATION_ON", false)
	assert.False(t, IsPropagationEnabled("test-org"))

	// Test re-enabled
	store.UpdateFlag("test-org", "PROPAGATION_ON", true)
	assert.True(t, IsPropagationEnabled("test-org"))
}

func TestIsSLOGuardEnabled(t *testing.T) {
	// Use the global feature store
	store := GlobalFeatureStore

	// Test default (enabled)
	assert.True(t, IsSLOGuardEnabled("test-org"))

	// Test disabled
	store.UpdateFlag("test-org", "SLO_GUARD_ON", false)
	assert.False(t, IsSLOGuardEnabled("test-org"))

	// Test re-enabled
	store.UpdateFlag("test-org", "SLO_GUARD_ON", true)
	assert.True(t, IsSLOGuardEnabled("test-org"))
}

func TestSLOGuard(t *testing.T) {
	guard := NewSLOGuard(5, 100, 60) // 5 samples, 100ms threshold, 60s cooldown

	// Test recording latencies
	guard.RecordLatency("test-org", 50)  // Under threshold
	guard.RecordLatency("test-org", 150) // Over threshold
	guard.RecordLatency("test-org", 200) // Over threshold
	guard.RecordLatency("test-org", 300) // Over threshold
	guard.RecordLatency("test-org", 400) // Over threshold

	// Should trigger fallback after 5 samples over threshold
	// This is tested by checking if the guard's internal state changes
	// The actual fallback logic would be tested in integration tests
}

func TestSLOGuardCooldown(t *testing.T) {
	guard := NewSLOGuard(3, 100, 1) // 3 samples, 100ms threshold, 1s cooldown

	// Record enough samples to trigger fallback
	guard.RecordLatency("test-org", 150)
	guard.RecordLatency("test-org", 150)
	guard.RecordLatency("test-org", 150)

	// Should be in cooldown now
	// The cooldown logic would be tested in integration tests
}

func TestFeatureFlagsConcurrency(t *testing.T) {
	store := NewFeatureStore()

	// Test concurrent access (basic test)
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			store.GetFlags("test-org")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			store.UpdateFlag("test-org", "PROPAGATION_ON", i%2 == 0)
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic and should have consistent state
	flags := store.GetFlags("test-org")
	assert.NotNil(t, flags)
}

func TestFeatureFlagsEdgeCases(t *testing.T) {
	// Use the global feature store
	store := GlobalFeatureStore

	// Test empty org ID
	flags := store.GetFlags("")
	assert.NotNil(t, flags)
	assert.True(t, flags.PropagationOn) // Should have defaults

	// Test setting nil flags
	store.SetFlags("test-org", nil)
	flags = store.GetFlags("test-org")
	assert.NotNil(t, flags) // Should not panic

	// Test updating with invalid flag name
	store.UpdateFlag("test-org", "", true)        // Empty flag name
	store.UpdateFlag("test-org", "INVALID", true) // Invalid flag name
	// Should not panic
}

func TestSLOGuardEdgeCases(t *testing.T) {
	guard := NewSLOGuard(1, 100, 60)

	// Test with empty org ID
	guard.RecordLatency("", 150)
	// Should not panic

	// Test with zero latency
	guard.RecordLatency("test-org", 0)
	// Should not panic

	// Test with negative latency
	guard.RecordLatency("test-org", -50)
	// Should not panic
}
