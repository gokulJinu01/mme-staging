# MME (Memory Management Engine)

## Overview

The MME (Memory Management Engine) provides intelligent memory management for AI systems using bounded semantic graph search and LLM-powered tag extraction. Perfect for enhancing RAG systems and building memory-aware AI applications.

## Services Included

### 1. MME Tagging Service (Go/Fiber)
- **Port**: 8081 (mapped from container port 8080)
- **Technology**: Go with Fiber framework
- **Purpose**: Memory storage, bounded graph search, semantic retrieval

### 2. MME Tagmaker Service (Python/FastAPI)
- **Port**: 8000
- **Technology**: Python with FastAPI
- **Purpose**: LLM-powered tag extraction, edge learning, memory tiering

## Key Features

✅ **Bounded Semantic Graph Search** - Mathematical performance guarantees O(M×D×B)  
✅ **LLM-Powered Tagging** - GPT-4o-mini semantic analysis  
✅ **Continuous Learning** - Edge weight updates from feedback  
✅ **Multi-tenant Ready** - Organization-scoped data isolation  
✅ **Production Ready** - Health checks, metrics, JWT authentication  

## Quick Start

### Prerequisites
- Docker and Docker Compose
- MongoDB (included in docker-compose.yml)
- OpenAI API Key
- Redis (optional, for caching)

### Environment Setup

1. Copy environment template:
```bash
cp .env.example .env
```

2. Configure required environment variables in `.env`:
```bash
# OpenAI Configuration (Required)
OPENAI_API_KEY=your_openai_api_key_here

# JWT Configuration (Required)
JWT_SECRET=your_jwt_secret_here

# Optional: Service Configuration
CORS_ALLOWED_ORIGINS=*
ENABLE_TAGGING_SERVICE=true
```

### Launch Services

```bash
# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8081/health  # Tagging Service
curl http://localhost:8000/health  # Tagmaker Service

# View logs
docker-compose logs -f mme-tagging-service
docker-compose logs -f mme-tagmaker-service
```

## API Endpoints

### MME Tagmaker Service (Port 8000)

**Tag Generation:**
```bash
curl -X POST http://localhost:8000/generate-and-save \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Completed machine learning project using neural networks",
    "userId": "user123",
    "source": "agent_output"
  }'
```

**Health & Status:**
```bash
curl http://localhost:8000/health
curl http://localhost:8000/version
curl http://localhost:8000/queue-status
```

### MME Tagging Service (Port 8081)

**Memory Storage:**
```bash
curl -X POST http://localhost:8081/memory/save \
  -H "X-User-ID: user123" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Project completed successfully",
    "tags": ["project", "completion", "success"],
    "section": "work",
    "status": "completed"
  }'
```

**Semantic Search:**
```bash
curl -X POST http://localhost:8081/search/semantic \
  -H "X-User-ID: user123" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning projects",
    "limit": 10,
    "minScore": 0.1
  }'
```

**Memory Query:**
```bash
curl "http://localhost:8081/memory/query?tags=project,completion&limit=10" \
  -H "X-User-ID: user123"
```

## Configuration

### Bounded Graph Parameters

The MME uses configurable bounded parameters for performance guarantees:

| Parameter | Environment Variable | Default | Description |
|-----------|---------------------|---------|-------------|
| Max Edges Per Tag (M) | `MME_MAX_EDGES_PER_TAG` | 32 | Maximum edges maintained per tag |
| Max Depth (D) | `MME_MAX_DEPTH` | 2 | Maximum propagation depth |
| Beam Width (B) | `MME_BEAM_WIDTH` | 128 | Maximum activations maintained |
| Decay Alpha (α) | `MME_DECAY_ALPHA` | 0.85 | Activation decay factor |
| Min Activation (θ) | `MME_MIN_ACTIVATION` | 0.05 | Minimum activation threshold |

### Service Configuration

**MME Tagging Service** (`mme-tagging-service/config.prod.yaml`):
```yaml
env: prod
logLevel: WARN
```

**MME Tagmaker Service** (`mme-tagmaker-service/config.prod.yaml`):
```yaml
env: prod
logLevel: INFO
```

## Integration Examples

### RAG Enhancement Pattern

```python
# Enhance existing RAG with MME
import requests

# 1. Extract tags from query
response = requests.post('http://localhost:8000/generate-and-save', 
    headers={'Authorization': f'Bearer {jwt_token}'},
    json={'content': user_query, 'userId': user_id, 'source': 'rag_query'})

# 2. Use tags for semantic search
tags = response.json()['cues']
search_response = requests.post('http://localhost:8081/search/semantic',
    headers={'X-User-ID': user_id},
    json={'query': user_query, 'limit': 20, 'minScore': 0.1})

# 3. Combine with traditional RAG results
mme_results = search_response.json()['results']
# Merge with vector search results...
```

### Agent Memory Pattern

```javascript
// Store agent execution context
await fetch('http://localhost:8081/memory/save', {
  method: 'POST',
  headers: {
    'X-User-ID': userId,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    content: agentOutput,
    tags: extractedTags,
    section: 'agent_execution',
    status: 'completed'
  })
});

// Retrieve relevant context for next execution
const contextResponse = await fetch(`http://localhost:8081/memory/query?tags=${tags.join(',')}&limit=10`, {
  headers: { 'X-User-ID': userId }
});
const relevantMemories = await contextResponse.json();
```

## Monitoring & Operations

### Health Checks
- **Tagging Service**: `GET /health` (every 30s)
- **Tagmaker Service**: `GET /health` (every 30s)
- **Database Status**: `GET /database-status` (Tagmaker)

### Metrics (Prometheus)
- Memory block operations
- Tag extraction performance
- Semantic search latency
- Edge learning progress

### Background Jobs
- **Tag Rebalancing**: Daily at 2 AM UTC
- **Failed Delta Retry**: Every minute
- **Edge Learning**: Every 10 minutes

## Production Deployment

### Resource Requirements
- **Tagging Service**: 1GB RAM, 0.5 CPU
- **Tagmaker Service**: 1GB RAM, 0.5 CPU
- **MongoDB**: 4GB RAM, 1 CPU (shared with other services)

### Security Considerations
- JWT authentication required for all endpoints
- Multi-tenant data isolation by organization
- Rate limiting configured per endpoint
- Input validation and sanitization

### Scaling
- Services scale independently
- MongoDB can be sharded by organization
- Redis clustering for cache scaling
- Load balancer friendly (stateless services)

## Troubleshooting

### Common Issues

**Service Won't Start:**
```bash
# Check environment variables
docker-compose config

# Check logs
docker-compose logs mme-tagging-service
docker-compose logs mme-tagmaker-service
```

**OpenAI API Errors:**
```bash
# Verify API key
curl http://localhost:8000/database-status

# Check rate limits in logs
docker-compose logs mme-tagmaker-service | grep "rate limit"
```

**Memory Query Issues:**
```bash
# Check MongoDB connection
docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')"

# Verify user context
curl "http://localhost:8081/memory/query?tags=test" -H "X-User-ID: debug-user"
```

### Debug Endpoints

**Tagmaker Service:**
- `GET /database-status` - MongoDB connection health
- `GET /database-debug` - Detailed connection info
- `GET /queue-status` - Failed operation queue status

**Tagging Service:**
- `GET /admin/stats` - Memory and tag statistics (requires admin role)
- `GET /metrics` - Prometheus metrics

## Documentation

- **Comprehensive Analysis**: `MME_SERVICES_COMPREHENSIVE_ANALYSIS.md`
- **Service Architecture**: `docs/mme-tagging-service.md`, `docs/mme-tagmaker-service.md`
- **API Documentation**: Available at `http://localhost:8000/docs` (FastAPI auto-docs)

## Support

## License

This project is licensed under the MIT License - see the LICENSE file for details.
