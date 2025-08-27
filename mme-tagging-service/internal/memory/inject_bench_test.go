package memory

import "testing"

var benchSinkBytes int

//go:noinline
func consumeBytes(n int) { benchSinkBytes += n }

func BenchmarkInject_Pack(b *testing.B) {
	blocks := make([]MemoryBlock, 0, 300)
	for i := 0; i < 300; i++ {
		blocks = append(blocks, MemoryBlock{
			Tags:       []string{"grant","irap","deadline"},
			Status:     "completed",
			Content:    "lorem ipsum dolor sit amet " + string(rune('a'+(i%26))),
			Confidence: 0.9,
			Priority:   i % 3,
		})
	}
	tags := []string{"grant","irap"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selected := PromoteMemory(blocks, tags, "continue")
		sum := 0
		for _, it := range selected { sum += len(it.Content) }
		consumeBytes(sum)
	}
}
