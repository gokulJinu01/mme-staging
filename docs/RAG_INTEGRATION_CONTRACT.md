# RAG Integration Contract for MME System

## Overview
This document defines the contract for integrating MME (Memory Management Engine) with RAG (Retrieval Augmented Generation) systems.

## Data Flow

### 1. Memory Indexing (MME → RAG)

When a memory is saved in MME, it should be indexed in the RAG system with the following metadata:

```json
{
  "userId": "u1",
  "docId": "mem_123",
  "text": "IRAP submission draft v3: budget assumptions revised; timeline shifted by 2 weeks.",
  "metadata": {
    "tags": ["IRAP", "submission", "budget", "timeline"],
    "section": "funding-proposal",
    "scope": "shared",
    "status": "completed",
    "createdAt": "2025-07-08T15:44:00Z",
    "confidence": 0.95,
    "source": "agent_output"
  }
}
```

### 2. Scoped Query (RAG with MME Context)

When querying the RAG system, include MME tag context for scoped retrieval:

```json
{
  "userId": "u1",
  "q": "Continue IRAP submission",
  "filter": { 
    "tags": { "$in": ["IRAP", "submission"] },
    "scope": "shared"
  },
  "limit": 5
}
```

### 3. Response Format

RAG should return results with MME metadata preserved:

```json
{
  "results": [
    {
      "docId": "mem_123",
      "text": "IRAP submission draft v3: budget assumptions revised...",
      "score": 0.92,
      "metadata": {
        "tags": ["IRAP", "submission", "budget", "timeline"],
        "section": "funding-proposal",
        "scope": "shared",
        "createdAt": "2025-07-08T15:44:00Z"
      }
    }
  ],
  "count": 1,
  "query": "Continue IRAP submission"
}
```

## Integration Points

### MME Endpoints for RAG Integration

1. **Memory Save Hook**: When `/memory/save` is called, trigger RAG indexing
2. **Tag Updates**: When tags are modified, update RAG metadata
3. **Memory Deletion**: When `/memory/delete` is called, remove from RAG index

### RAG Endpoints Expected

1. **Index/Upsert**: `POST /rag/index` - Index memory with metadata
2. **Query**: `POST /rag/query` - Query with tag filters
3. **Delete**: `DELETE /rag/index/{docId}` - Remove from index

## Acceptance Criteria

1. **Scoped Results**: Querying with tag filters returns only relevant content
2. **Metadata Preservation**: All MME metadata is preserved in RAG
3. **Performance**: RAG queries complete within 100ms
4. **Fallback**: If RAG is unavailable, MME continues to work independently

## Implementation Notes

- Use structured tags for precise filtering
- Maintain `tagsFlat` for backward compatibility
- Preserve all MME metadata fields in RAG
- Implement proper error handling for RAG failures
- Use async indexing to avoid blocking MME operations
