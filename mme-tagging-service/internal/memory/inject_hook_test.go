//go:build integration

package memory

import (
	"context"
	"time"
)

// TestHook allows injection of artificial delays for SLO testing
type TestHook struct {
	Delay time.Duration
}

var testHook *TestHook

// SetTestHook sets the test hook for SLO testing
func SetTestHook(hook *TestHook) {
	testHook = hook
}

// ClearTestHook clears the test hook
func ClearTestHook() {
	testHook = nil
}

// applyTestDelay applies artificial delay if test hook is set
func applyTestDelay() {
	if testHook != nil && testHook.Delay > 0 {
		time.Sleep(testHook.Delay)
	}
}

// BuildInjectionPackWithDelay is a test-only version that applies delays
func BuildInjectionPackWithDelay(ctx context.Context, request InjectRequest, propagationEnabled bool) (*InjectResponse, error) {
	// Apply test delay if set
	applyTestDelay()

	// Call the real function
	return buildInjectionPack(ctx, request, propagationEnabled)
}
