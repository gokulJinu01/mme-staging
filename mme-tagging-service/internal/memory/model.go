package memory

import "time"

// Tag represents a structured tag with metadata and graph capabilities
type Tag struct {
	Label       string    `json:"label" bson:"label"`
	Section     string    `json:"section,omitempty" bson:"section,omitempty"`
	Origin      string    `json:"origin,omitempty" bson:"origin,omitempty"`     // agent|system|user|unknown
	Scope       string    `json:"scope,omitempty" bson:"scope,omitempty"`       // local|shared|global
	Type        string    `json:"type,omitempty" bson:"type,omitempty"`         // concept|action|object|error|status|misc
	Confidence  float64   `json:"confidence,omitempty" bson:"confidence,omitempty"`
	Links       []string  `json:"links,omitempty" bson:"links,omitempty"`
	UsageCount  int       `json:"usageCount,omitempty" bson:"usageCount,omitempty"`
	LastUsed    time.Time `json:"lastUsed,omitempty" bson:"lastUsed,omitempty"`
}

type MemoryBlock struct {
	ID         string     `json:"id" bson:"_id,omitempty"`
	UserID     string     `json:"userId" bson:"userId"`
	Tags       []Tag      `json:"tags,omitempty" bson:"tags,omitempty"`
	TagsFlat   []string   `json:"tagsFlat,omitempty" bson:"tagsFlat,omitempty"` // for compat & index
	Section    string     `json:"section,omitempty" bson:"section,omitempty"`
	Status     string     `json:"status,omitempty" bson:"status,omitempty"`
	Content    string     `json:"content" bson:"content"`
	Source     string     `json:"source,omitempty" bson:"source,omitempty"`
	CreatedAt  time.Time  `json:"createdAt" bson:"createdAt"`
	Hash       string     `json:"hash,omitempty" bson:"hash,omitempty"`
	Confidence float64    `json:"confidence,omitempty" bson:"confidence,omitempty"`
	Priority   int        `json:"priority,omitempty" bson:"priority,omitempty"`
	TTL        int64      `json:"ttl,omitempty" bson:"ttl,omitempty"`
}

// PromptInput represents a request to extract tags from a prompt
type PromptInput struct {
	UserID string `json:"userId" validate:"required"`
	Prompt string `json:"prompt" validate:"required"`
}

// PromptQueryInput represents a request to query memory using a prompt
type PromptQueryInput struct {
	UserID  string `json:"userId" validate:"required"`
	Prompt  string `json:"prompt" validate:"required"`
	Section string `json:"section,omitempty"`
	Status  string `json:"status,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// TagExtractionResponse represents the response from tag extraction
type TagExtractionResponse struct {
	UserID string   `json:"userId"`
	Tags   []string `json:"tags"`
	Count  int      `json:"count"`
}

// SemanticSearchRequest represents a semantic search request
type SemanticSearchRequest struct {
	Query    string  `json:"query" validate:"required"`
	UserID   string  `json:"userId,omitempty"`
	Limit    int     `json:"limit,omitempty"`
	MinScore float64 `json:"minScore,omitempty"`
}

// SemanticSearchResponse represents the response from semantic search
type SemanticSearchResponse struct {
	Query   string                   `json:"query"`
	Results []SemanticSearchResult   `json:"results"`
	Count   int                      `json:"count"`
	UserID  string                   `json:"userId"`
}

// SemanticSearchResult represents a single semantic search result
type SemanticSearchResult struct {
	MemoryBlock MemoryBlock `json:"memoryBlock"`
	Score       float64     `json:"score"`
	Relevance   string      `json:"relevance"`
}
