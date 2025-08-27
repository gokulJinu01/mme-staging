package metrics

import (
	"sync"
	"time"
)

// Simple metrics storage
type MetricData struct {
	Count   int64
	Sum     float64
	Min     float64
	Max     float64
	Buckets map[float64]int64
}

type MetricsStore struct {
	mu sync.RWMutex
	// Propagation duration metrics
	propagationDuration map[string]*MetricData // key: status
	// Inject duration metrics
	injectDuration map[string]*MetricData // key: status
	// Pack items counters
	packItems map[string]int64 // key: section_status
	// Cache hit/miss counters
	cacheHits   map[string]int64 // key: cache_type
	cacheMisses map[string]int64 // key: cache_type
}

var globalMetrics = &MetricsStore{
	propagationDuration: make(map[string]*MetricData),
	injectDuration:      make(map[string]*MetricData),
	packItems:           make(map[string]int64),
	cacheHits:           make(map[string]int64),
	cacheMisses:         make(map[string]int64),
}

// RecordPropagationDuration records the duration of a propagation operation
func RecordPropagationDuration(duration time.Duration, status string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	if globalMetrics.propagationDuration[status] == nil {
		globalMetrics.propagationDuration[status] = &MetricData{
			Buckets: make(map[float64]int64),
		}
	}

	data := globalMetrics.propagationDuration[status]
	data.Count++
	data.Sum += duration.Seconds()
	if data.Min == 0 || duration.Seconds() < data.Min {
		data.Min = duration.Seconds()
	}
	if duration.Seconds() > data.Max {
		data.Max = duration.Seconds()
	}
}

// RecordInjectDuration records the duration of an injection operation
func RecordInjectDuration(duration time.Duration, status string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	if globalMetrics.injectDuration[status] == nil {
		globalMetrics.injectDuration[status] = &MetricData{
			Buckets: make(map[float64]int64),
		}
	}

	data := globalMetrics.injectDuration[status]
	data.Count++
	data.Sum += duration.Seconds()
	if data.Min == 0 || duration.Seconds() < data.Min {
		data.Min = duration.Seconds()
	}
	if duration.Seconds() > data.Max {
		data.Max = duration.Seconds()
	}
}

// RecordPackItems records the number of items packed
func RecordPackItems(count int, section, status string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	key := section + "_" + status
	globalMetrics.packItems[key] += int64(count)
}

// RecordCacheHit records a cache hit
func RecordCacheHit(cacheType string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.cacheHits[cacheType]++
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss(cacheType string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.cacheMisses[cacheType]++
}

// RecordPackAccept records a pack acceptance event
func RecordPackAccept(orgID, projectID string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	key := orgID + "_" + projectID
	globalMetrics.packItems["accept_"+key]++
}

// RecordPackReject records a pack rejection event
func RecordPackReject(orgID, projectID string) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	key := orgID + "_" + projectID
	globalMetrics.packItems["reject_"+key]++
}

// Timer wraps a timer for easy recording
type Timer struct {
	start    time.Time
	recordFn func(time.Duration, string)
	status   string
	recorded bool
}

// NewPropagationTimer creates a new timer for propagation operations
func NewPropagationTimer(status string) *Timer {
	return &Timer{
		start:    time.Now(),
		recordFn: RecordPropagationDuration,
		status:   status,
	}
}

// NewInjectTimer creates a new timer for injection operations
func NewInjectTimer(status string) *Timer {
	return &Timer{
		start:    time.Now(),
		recordFn: RecordInjectDuration,
		status:   status,
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() {
	if !t.recorded {
		duration := time.Since(t.start)
		t.recordFn(duration, t.status)
		t.recorded = true
	}
}

// StopWithStatus stops the timer and records the duration with a custom status
func (t *Timer) StopWithStatus(status string) {
	if !t.recorded {
		duration := time.Since(t.start)
		t.recordFn(duration, status)
		t.recorded = true
	}
}

// GetMetrics returns the current metrics data
func GetMetrics() map[string]interface{} {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	// Convert packItems to a serializable format
	packItemsSerializable := make(map[string]interface{})
	for key, value := range globalMetrics.packItems {
		packItemsSerializable[key] = value
	}

	// Convert propagation duration to serializable format
	propagationDurationSerializable := make(map[string]interface{})
	for status, data := range globalMetrics.propagationDuration {
		propagationDurationSerializable[status] = map[string]interface{}{
			"count": data.Count,
			"sum":   data.Sum,
			"min":   data.Min,
			"max":   data.Max,
		}
	}

	// Convert inject duration to serializable format
	injectDurationSerializable := make(map[string]interface{})
	for status, data := range globalMetrics.injectDuration {
		injectDurationSerializable[status] = map[string]interface{}{
			"count": data.Count,
			"sum":   data.Sum,
			"min":   data.Min,
			"max":   data.Max,
		}
	}

	return map[string]interface{}{
		"propagation_duration": propagationDurationSerializable,
		"inject_duration":      injectDurationSerializable,
		"pack_items":           packItemsSerializable,
		"cache_hits":           globalMetrics.cacheHits,
		"cache_misses":         globalMetrics.cacheMisses,
	}
}
