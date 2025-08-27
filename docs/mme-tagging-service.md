# MME Tagging Service

## Purpose & Responsibilities

The MME (Memory Management Engine) Tagging Service provides semantic memory management, tag extraction, and memory querying capabilities. It handles memory block storage, semantic search, and tag-based retrieval for AI agents.

## Tech Stack & Entry Points

- **Framework**: Go with Fiber
- **Port**: 8081 (HTTP, mapped from container port 8080)
- **Main Files**: 
  - `api/routes.go` - API route definitions
  - `internal/memory/` - Memory management handlers
  - `internal/db/` - Database operations
  - `internal/metrics/` - Metrics collection

## Public API

| Method | Path | Handler | Auth/Role | Rate Limit |
|--------|------|---------|-----------|------------|
| GET | `/health` | Health check | Public | 100/min |
| GET | `/metrics` | Metrics endpoint | Public | 30/min |
| POST | `/memory/save` | memory.SaveBlock | JWT Required | 200/min |
| GET | `/memory/query` | memory.QueryBlocks | JWT Required | 300/min |
| GET | `/memory/recent` | memory.GetRecent | JWT Required | 300/min |
| DELETE | `/memory/:id` | memory.DeleteBlock | JWT Required | 100/min |
| POST | `/memory/promote` | memory.PromoteBlocks | JWT Required | 100/min |
| POST | `/memory/inject` | memory.HandleMemoryInject | JWT Required | 200/min |
| POST | `/tags/extract` | memory.ExtractTagsFromPrompt | JWT Required | 200/min |
| POST | `/tags/query` | memory.QueryBlocksByPrompt | JWT Required | 300/min |
| POST | `/tags/delta` | memory.UpdateTagDelta | JWT Required | 200/min |
| POST | `/search/semantic` | memory.SemanticSearch | JWT Required | 300/min |
| POST | `/events/pack` | memory.HandlePackEvent | JWT Required | 100/min |

## Internal/Admin API

| Method | Path | Handler | Auth/Role | Purpose |
|--------|------|---------|-----------|---------|
| GET | `/admin/stats` | handleAdminStats | Admin Role | System statistics |
| POST | `/admin/cleanup` | handleAdminCleanup | Admin Role | Memory cleanup |
| POST | `/admin/edges/prune` | handleAdminEdgePrune | Admin Role | Edge pruning |
| GET | `/admin/features` | feature.HandleGetFeatures | Admin Role | Feature flags |
| POST | `/admin/features` | feature.HandleSetFeatures | Admin Role | Feature configuration |

## Dependencies

### In-Cluster
- **MongoDB** (27017) - Memory blocks and tags storage
- **Redis** (6379) - Caching and session management (optional)

### External
- **Go modules** - Dependencies via go.mod

## Config & Secrets Matrix

| Variable | Source | Purpose | Required |
|----------|--------|---------|----------|
| `MONGODB_URI` | env | MongoDB connection string | ✅ |
| `GO_ENV` | env | Environment (production/development) | ✅ |
| `MME_LOG_LEVEL` | env | Logging level | ✅ |
| `PORT` | env | Service port | ✅ |
| `JWT_SECRET` | env | JWT signing secret | ✅ |
| `CORS_ALLOWED_ORIGINS` | env | CORS configuration | ✅ |

## Data Contracts

### Memory Block
```json
{
  "id": "string",
  "content": "string",
  "tags": ["string"],
  "userId": "string",
  "orgId": "string",
  "createdAt": "timestamp",
  "updatedAt": "timestamp",
  "metadata": "object"
}
```

### Tag Query
```json
{
  "prompt": "string",
  "tags": ["string"],
  "limit": "number",
  "userId": "string",
  "orgId": "string"
}
```

### Semantic Search
```json
{
  "query": "string",
  "limit": "number",
  "threshold": "number",
  "userId": "string",
  "orgId": "string"
}
```

### Admin Stats
```json
{
  "total_memory_blocks": "number",
  "total_tags": "number",
  "last_cleanup_time": "timestamp",
  "uptime_seconds": "number",
  "memory_usage_mb": "number",
  "requests_per_minute": "number",
  "timestamp": "timestamp",
  "requested_by": "string"
}
```

## Workflows

- **Memory Storage**: Agent → MME Tagging → MongoDB (memory blocks)
- **Tag Extraction**: Prompt → MME Tagging → Tag extraction → MongoDB (tags)
- **Semantic Search**: Query → MME Tagging → Vector search → Relevant memories
- **Memory Cleanup**: Admin → MME Tagging → Expired memory deletion

## Security Posture

### ✅ Implemented
- **JWT Authentication**: All endpoints require valid JWT (except health/metrics)
- **Role-based Access**: Admin endpoints require admin role
- **Organization Isolation**: Multi-tenant data segregation
- **Input Validation**: Request validation and sanitization
- **Rate Limiting**: Configurable rate limits per endpoint

### 🔄 Residual Risks
- **Data Privacy**: Memory content may contain sensitive information
- **Access Control**: Organization-level data isolation
- **Audit Logging**: Memory access and modification tracking

## Observability

### Health Endpoints
- **Health Check**: `http://localhost:8081/health`
- **Metrics**: `http://localhost:8081/metrics`

### Key Metrics
- **Memory Blocks**: Total memory blocks per organization
- **Tags**: Total tags and tag usage statistics
- **Query Performance**: Semantic search response times
- **Storage Usage**: MongoDB collection sizes
- **Request Rates**: API endpoint usage patterns

### Dashboards
- **Grafana**: MME Tagging Service dashboard with memory metrics
- **Prometheus**: Custom metrics for memory operations

## SLOs/SLIs

### Service Level Objectives
- **Availability**: 99.9% uptime
- **Latency**: P95 < 500ms for memory queries, P95 < 1s for semantic search
- **Throughput**: 300 queries/minute, 200 saves/minute

### Service Level Indicators
- **Error Rate**: < 1% for all endpoints
- **Memory Query Success**: > 98% successful queries
- **Tag Extraction Accuracy**: > 90% relevant tag extraction

### TODO: Verification Commands
```bash
# Check service health
curl http://localhost:8081/health

# Get service metrics
curl http://localhost:8081/metrics

# Test memory save (requires JWT)
curl -X POST http://localhost:8081/memory/save \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"content": "test memory", "tags": ["test"], "userId": "user123", "orgId": "org123"}'

# Get admin stats (requires admin JWT)
curl -H "Authorization: Bearer <ADMIN_JWT>" http://localhost:8081/admin/stats
```
