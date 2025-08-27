package memory

import (
	"context"
	"os"
	"testing"
	"time"
)

// sinks to defeat DCE
var benchSinkInt int

//go:noinline
func consumeInt(n int) { benchSinkInt += n }

// OPTIONAL: seed a tiny neighborhood so propagation does real work.
// This uses MONGO_URI if present; otherwise it's a no-op.
func maybeSeedTinyGraph() {
	uri := os.Getenv("TEST_MONGO_URI")
	if uri == "" {
		return
	}
	// minimal in-process seeding path if your EdgeStore is available via a helper.
	// If not, rely on perf/seed_edges.sh (below) instead.
}

func BenchmarkBounded_Cold(b *testing.B) {
	maybeSeedTinyGraph()

	cfg := DefaultBoundedConfig()
	seeds := []string{"grant","irap","deadline","proposal","nda"}
	if len(seeds) > cfg.MaxSeedTags {
		seeds = seeds[:cfg.MaxSeedTags]
	}
	_ = performBoundedPropagation(context.Background(), seeds, cfg) // prime

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tags := performBoundedPropagation(context.Background(), seeds, cfg)
		consumeInt(len(tags))
	}
}

func BenchmarkBounded_Warm(b *testing.B) {
	cfg := DefaultBoundedConfig()
	seeds := []string{"grant","irap","deadline","proposal","nda"}
	if len(seeds) > cfg.MaxSeedTags {
		seeds = seeds[:cfg.MaxSeedTags]
	}
	_ = performBoundedPropagation(context.Background(), seeds, cfg)
	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tags := performBoundedPropagation(context.Background(), seeds, cfg)
		consumeInt(len(tags))
	}
}
