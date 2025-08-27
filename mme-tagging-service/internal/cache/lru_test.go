package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(3, 5*time.Second)

	// Test basic set and get
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Test non-existent key
	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)

	// Test cache size limit
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4") // Should evict key1

	_, exists = cache.Get("key1")
	assert.False(t, exists) // Should be evicted

	// Test LRU behavior
	cache.Get("key2") // Access key2 to make it most recently used
	cache.Set("key5", "value5") // Should evict key3 (least recently used)

	_, exists = cache.Get("key3")
	assert.False(t, exists) // Should be evicted
	_, exists = cache.Get("key2")
	assert.True(t, exists) // Should still exist
}

func TestLRUCacheTTL(t *testing.T) {
	cache := NewLRUCache(3, 100*time.Millisecond)

	// Set a value
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Value should be expired
	_, exists = cache.Get("key1")
	assert.False(t, exists)
}

func TestLRUCacheRemove(t *testing.T) {
	cache := NewLRUCache(3, 5*time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Remove existing key
	cache.Remove("key1")
	_, exists := cache.Get("key1")
	assert.False(t, exists)

	// Remove non-existent key (should not panic)
	cache.Remove("nonexistent")

	// Other keys should still exist
	value, exists := cache.Get("key2")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)
}

func TestLRUCacheClear(t *testing.T) {
	cache := NewLRUCache(3, 5*time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	assert.Equal(t, 2, cache.Size())

	cache.Clear()

	assert.Equal(t, 0, cache.Size())
	_, exists := cache.Get("key1")
	assert.False(t, exists)
	_, exists = cache.Get("key2")
	assert.False(t, exists)
}

func TestLRUCacheSize(t *testing.T) {
	cache := NewLRUCache(5, 5*time.Second)

	assert.Equal(t, 0, cache.Size())

	cache.Set("key1", "value1")
	assert.Equal(t, 1, cache.Size())

	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Size())

	cache.Remove("key1")
	assert.Equal(t, 1, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

func TestLRUCacheConcurrency(t *testing.T) {
	cache := NewLRUCache(10, 5*time.Second)
	done := make(chan bool, 2)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("key1")
		}
		done <- true
	}()

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("key1", "value1")
		}
		done <- true
	}()

	<-done
	<-done

	// Should not panic and should have consistent state
	assert.GreaterOrEqual(t, cache.Size(), 0)
}

func TestLRUCacheCleanup(t *testing.T) {
	cache := NewLRUCache(3, 50*time.Millisecond)

	// Set values that will expire
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Cleanup should remove expired entries
	cache.Cleanup()

	assert.Equal(t, 0, cache.Size())
}

func TestGenerateRelatedCacheKey(t *testing.T) {
	key := GenerateRelatedCacheKey("test-org", "python")
	expected := "related:test-org:python"
	assert.Equal(t, expected, key)

	// Test with empty org
	key = GenerateRelatedCacheKey("", "python")
	expected = "related::python"
	assert.Equal(t, expected, key)
}

func TestGeneratePropCacheKey(t *testing.T) {
	key := GeneratePropCacheKey("test-org", "seed-hash", "filters-hash")
	expected := "prop:test-org:seed-hash:filters-hash"
	assert.Equal(t, expected, key)

	// Test with empty values
	key = GeneratePropCacheKey("", "", "")
	expected = "prop:::"
	assert.Equal(t, expected, key)
}

func TestLRUCacheEdgeCases(t *testing.T) {
	// Test with zero capacity
	cache := NewLRUCache(0, 5*time.Second)
	cache.Set("key1", "value1")
	assert.Equal(t, 0, cache.Size())

	// Test with zero TTL
	cache = NewLRUCache(3, 0)
	cache.Set("key1", "value1")
	_, exists := cache.Get("key1")
	assert.False(t, exists) // Should be immediately expired

	// Test with negative TTL
	cache = NewLRUCache(3, -1*time.Second)
	cache.Set("key1", "value1")
	_, exists = cache.Get("key1")
	assert.False(t, exists) // Should be immediately expired
}

func TestLRUCacheUpdateExisting(t *testing.T) {
	cache := NewLRUCache(3, 5*time.Second)

	// Set initial value
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Update existing key
	cache.Set("key1", "value2")
	value, exists = cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value2", value)

	// Size should not change
	assert.Equal(t, 1, cache.Size())
}
