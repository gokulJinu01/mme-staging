package feature

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FeatureFlags represents the available feature flags
type FeatureFlags struct {
	PropagationOn  bool `json:"PROPAGATION_ON"`
	SLOGuardOn     bool `json:"SLO_GUARD_ON"`
	EdgeLearningOn bool `json:"EDGE_LEARNING_ON"`
}

// FeatureStore manages feature flags per organization
type FeatureStore struct {
	flags map[string]*FeatureFlags
	mutex sync.RWMutex
}

// Global feature store instance
var GlobalFeatureStore = NewFeatureStore()

// NewFeatureStore creates a new feature store
func NewFeatureStore() *FeatureStore {
	return &FeatureStore{
		flags: make(map[string]*FeatureFlags),
	}
}

// GetFlags returns the feature flags for an organization
func (fs *FeatureStore) GetFlags(orgID string) *FeatureFlags {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	flags, exists := fs.flags[orgID]
	if !exists {
		// Return default flags
		return &FeatureFlags{
			PropagationOn:  true,
			SLOGuardOn:     true,
			EdgeLearningOn: true,
		}
	}
	return flags
}

// SetFlags sets the feature flags for an organization
func (fs *FeatureStore) SetFlags(orgID string, flags *FeatureFlags) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// If flags is nil, remove the entry so GetFlags returns defaults
	if flags == nil {
		delete(fs.flags, orgID)
	} else {
		fs.flags[orgID] = flags
	}
}

// UpdateFlag updates a specific flag for an organization
func (fs *FeatureStore) UpdateFlag(orgID, flagName string, value bool) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	flags, exists := fs.flags[orgID]
	if !exists {
		flags = &FeatureFlags{
			PropagationOn:  true,
			SLOGuardOn:     true,
			EdgeLearningOn: true,
		}
		fs.flags[orgID] = flags
	}

	switch flagName {
	case "PROPAGATION_ON":
		flags.PropagationOn = value
	case "SLO_GUARD_ON":
		flags.SLOGuardOn = value
	case "EDGE_LEARNING_ON":
		flags.EdgeLearningOn = value
	}
}

// SLOGuard manages SLO monitoring and automatic fallbacks
type SLOGuard struct {
	latencyBuffer map[string][]time.Duration // orgID -> ring buffer of latencies
	mutex         sync.RWMutex
	bufferSize    int
	sloThreshold  time.Duration
	cooldownTime  time.Duration
	cooldowns     map[string]time.Time // orgID -> cooldown end time
}

// Global SLO guard instance
var GlobalSLOGuard = NewSLOGuard(500, 250*time.Millisecond, 10*time.Minute)

// NewSLOGuard creates a new SLO guard
func NewSLOGuard(bufferSize int, sloThreshold, cooldownTime time.Duration) *SLOGuard {
	return &SLOGuard{
		latencyBuffer: make(map[string][]time.Duration),
		bufferSize:    bufferSize,
		sloThreshold:  sloThreshold,
		cooldownTime:  cooldownTime,
		cooldowns:     make(map[string]time.Time),
	}
}

// RecordLatency records a latency measurement for an organization
func (sg *SLOGuard) RecordLatency(orgID string, latency time.Duration) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	// Initialize buffer if needed
	if sg.latencyBuffer[orgID] == nil {
		sg.latencyBuffer[orgID] = make([]time.Duration, 0, sg.bufferSize)
	}

	// Add latency to buffer
	buffer := sg.latencyBuffer[orgID]
	buffer = append(buffer, latency)

	// Maintain buffer size
	if len(buffer) > sg.bufferSize {
		buffer = buffer[1:]
	}

	sg.latencyBuffer[orgID] = buffer

	// Check if we need to trigger SLO fallback
	sg.checkSLOFallback(orgID)
}

// checkSLOFallback checks if SLO threshold is breached and triggers fallback
func (sg *SLOGuard) checkSLOFallback(orgID string) {
	// Check if in cooldown
	if cooldownEnd, exists := sg.cooldowns[orgID]; exists && time.Now().Before(cooldownEnd) {
		return
	}

	buffer := sg.latencyBuffer[orgID]
	if len(buffer) < 100 { // Need minimum samples
		return
	}

	// Calculate p95
	sorted := make([]time.Duration, len(buffer))
	copy(sorted, buffer)

	// Simple sort (in production, use sort.Slice)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	p95Index := int(float64(len(sorted)) * 0.95)
	p95Latency := sorted[p95Index]

	// If p95 exceeds threshold, trigger fallback
	if p95Latency > sg.sloThreshold {
		GlobalFeatureStore.UpdateFlag(orgID, "PROPAGATION_ON", false)
		sg.cooldowns[orgID] = time.Now().Add(sg.cooldownTime)

		// Record metric
		// metrics.RecordInjectFallback(orgID, "slo")
	}
}

// IsPropagationEnabled checks if propagation is enabled for an organization
func IsPropagationEnabled(orgID string) bool {
	flags := GlobalFeatureStore.GetFlags(orgID)
	return flags.PropagationOn
}

// IsSLOGuardEnabled checks if SLO guard is enabled for an organization
func IsSLOGuardEnabled(orgID string) bool {
	flags := GlobalFeatureStore.GetFlags(orgID)
	return flags.SLOGuardOn
}

// HandleGetFeatures handles GET /admin/features endpoint
func HandleGetFeatures(c *fiber.Ctx) error {
	orgID := c.Query("orgId")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "orgId query parameter is required",
		})
	}

	flags := GlobalFeatureStore.GetFlags(orgID)
	return c.JSON(fiber.Map{
		"orgId": orgID,
		"flags": flags,
	})
}

// HandleSetFeatures handles POST /admin/features endpoint
func HandleSetFeatures(c *fiber.Ctx) error {
	var request struct {
		OrgID string       `json:"orgId"`
		Flags FeatureFlags `json:"flags"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if request.OrgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "orgId is required",
		})
	}

	GlobalFeatureStore.SetFlags(request.OrgID, &request.Flags)

	return c.JSON(fiber.Map{
		"status": "success",
		"orgId":  request.OrgID,
		"flags":  request.Flags,
	})
}
