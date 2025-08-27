package cache

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached entry with expiration and access time
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
	LastAccess time.Time
}

// LRUCache represents a thread-safe LRU cache with TTL
type LRUCache struct {
	cache    map[string]CacheEntry
	mutex    sync.RWMutex
	capacity int
	ttl      time.Duration
}

// NewLRUCache creates a new LRU cache with the specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		cache:    make(map[string]CacheEntry),
		capacity: capacity,
		ttl:      ttl,
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	entry, exists := c.cache[key]
	if !exists {
		c.mutex.RUnlock()
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.Expiration) {
		c.mutex.RUnlock()
		// Remove expired entry (we'll do this in a write lock)
		go c.removeExpired(key)
		return nil, false
	}

	// Update last access time (need write lock)
	c.mutex.RUnlock()
	c.mutex.Lock()
	if entry, exists := c.cache[key]; exists {
		entry.LastAccess = time.Now()
		c.cache[key] = entry
	}
	c.mutex.Unlock()

	return entry.Value, true
}

// Set stores a value in the cache
func (c *LRUCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If capacity is 0, don't store anything
	if c.capacity == 0 {
		return
	}

	// Check if we need to evict entries due to capacity
	if len(c.cache) >= c.capacity {
		c.evictOldest()
	}

	// Store the entry with expiration and access time
	c.cache[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
		LastAccess: time.Now(),
	}
}

// Remove removes a key from the cache
func (c *LRUCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.cache, key)
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]CacheEntry)
}

// Size returns the current number of entries in the cache
func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// evictOldest removes the least recently used entry from the cache
func (c *LRUCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.cache {
		if first || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}

// removeExpired removes an expired entry from the cache
func (c *LRUCache) removeExpired(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.cache, key)
}

// Cleanup removes all expired entries from the cache
func (c *LRUCache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.After(entry.Expiration) {
			delete(c.cache, key)
		}
	}
}

// Global cache instances
var (
	RelatedCache *LRUCache
	PropCache    *LRUCache
)

// InitCaches initializes the global cache instances
func InitCaches(ttlSeconds int) {
	ttl := time.Duration(ttlSeconds) * time.Second
	RelatedCache = NewLRUCache(1000, ttl) // 1000 entries for related tags
	PropCache = NewLRUCache(500, ttl)     // 500 entries for propagation results
}

// GetRelatedCache returns the global related cache instance
func GetRelatedCache() *LRUCache {
	return RelatedCache
}

// GetPropCache returns the global propagation cache instance
func GetPropCache() *LRUCache {
	return PropCache
}

// GenerateRelatedCacheKey generates a cache key for related tags with org prefix
func GenerateRelatedCacheKey(orgID, tag string) string {
	return fmt.Sprintf("related:%s:%s", orgID, tag)
}

// GeneratePropCacheKey generates a cache key for propagation results with org prefix
func GeneratePropCacheKey(orgID, seedHash, filtersHash string) string {
	return fmt.Sprintf("prop:%s:%s:%s", orgID, seedHash, filtersHash)
}
