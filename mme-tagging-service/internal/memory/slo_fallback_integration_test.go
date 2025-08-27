//go:build integration

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mme-tagging-service/internal/feature"
)

func TestSLOFallbackIntegration(t *testing.T) {
	// Set up test environment
	t.Setenv("MME_INJECT_SLO_MS", "5")      // Very low threshold
	t.Setenv("MME_SLO_COOLDOWN_SECS", "60") // Short cooldown

	orgID := "test-org"

	// Enable SLO guard and propagation
	feature.GlobalFeatureStore.SetFlags(orgID, &feature.FeatureFlags{
		PropagationOn: true,
		SLOGuardOn:    true,
	})

	// Create test request
	request := InjectRequest{
		OrgID:       orgID,
		ProjectID:   "test-project",
		Prompt:      "Tell me about Python and FastAPI",
		Limit:       10,
		TokenBudget: 2048,
	}

	ctx := context.Background()

	// First, verify propagation is enabled
	assert.True(t, feature.IsPropagationEnabled(orgID))
	assert.True(t, feature.IsSLOGuardEnabled(orgID))

	// Inject artificial delays to trigger SLO fallback
	// We need to make several requests to build up the latency buffer
	for i := 0; i < 10; i++ {
		SetTestHook(&TestHook{Delay: 10 * time.Millisecond}) // 10ms delay

		_, err := BuildInjectionPackWithDelay(ctx, request, true)
		require.NoError(t, err)

		ClearTestHook()
	}

	// Wait a moment for SLO calculation
	time.Sleep(100 * time.Millisecond)

	// Verify that propagation has been disabled due to SLO breach
	assert.False(t, feature.IsPropagationEnabled(orgID), "Propagation should be disabled due to SLO breach")
}

func TestSLOFallbackRecovery(t *testing.T) {
	// Set up test environment with longer cooldown
	t.Setenv("MME_INJECT_SLO_MS", "5")
	t.Setenv("MME_SLO_COOLDOWN_SECS", "1") // 1 second cooldown

	orgID := "test-org"

	// Enable SLO guard and propagation
	feature.GlobalFeatureStore.SetFlags(orgID, &feature.FeatureFlags{
		PropagationOn: true,
		SLOGuardOn:    true,
	})

	request := InjectRequest{
		OrgID:       orgID,
		ProjectID:   "test-project",
		Prompt:      "Test prompt",
		Limit:       10,
		TokenBudget: 2048,
	}

	ctx := context.Background()

	// Trigger SLO fallback
	for i := 0; i < 10; i++ {
		SetTestHook(&TestHook{Delay: 10 * time.Millisecond})
		_, err := BuildInjectionPackWithDelay(ctx, request, true)
		require.NoError(t, err)
		ClearTestHook()
	}

	// Wait for SLO calculation
	time.Sleep(100 * time.Millisecond)

	// Verify propagation is disabled
	assert.False(t, feature.IsPropagationEnabled(orgID))

	// Wait for cooldown to expire
	time.Sleep(2 * time.Second)

	// Verify propagation is re-enabled after cooldown
	assert.True(t, feature.IsPropagationEnabled(orgID), "Propagation should be re-enabled after cooldown")
}

func TestSLOFallbackBasic(t *testing.T) {
	// Set up test environment
	t.Setenv("MME_INJECT_SLO_MS", "5")
	t.Setenv("MME_SLO_COOLDOWN_SECS", "60")

	orgID := "test-org"

	// Enable SLO guard and propagation
	feature.GlobalFeatureStore.SetFlags(orgID, &feature.FeatureFlags{
		PropagationOn: true,
		SLOGuardOn:    true,
	})

	// Verify initial state
	assert.True(t, feature.IsPropagationEnabled(orgID))
	assert.True(t, feature.IsSLOGuardEnabled(orgID))

	// Test that SLO guard is working
	feature.GlobalSLOGuard.RecordLatency(orgID, 100*time.Millisecond)

	// Verify SLO guard recorded the latency
	// Note: We can't easily test the internal state, but we can verify the function doesn't panic
	assert.True(t, feature.IsPropagationEnabled(orgID))
}
