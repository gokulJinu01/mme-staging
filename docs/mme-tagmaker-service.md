# MME Tagmaker Service

## Purpose & Responsibilities

The MME Tagmaker Service provides AI-powered tag extraction, semantic analysis, and memory tiering for agent outputs. It processes agent content, extracts relevant tags, and manages memory organization through automated learning and rebalancing.

## Tech Stack & Entry Points

- **Framework**: FastAPI (Python)
- **Port**: 8000 (HTTP)
- **Main Files**: 
  - `app/main.py` - FastAPI application and endpoints
  - `app/router.py` - Main API routes
  - `app/routes/edge_admin.py` - Edge learning admin routes
  - `app/security/` - Security middleware and handlers

## Public API

| Method | Path | Handler | Auth/Role | Rate Limit |
|--------|------|---------|-----------|------------|
| GET | `/health` | health | Public | 100/min |
| POST | `/feedback` | feedback | JWT Required | 100/min |
| GET | `/version` | version | Public | 30/min |
| POST | `/manual-rebalance` | router.manual_rebalance | JWT Required | 10/min |
| POST | `/generate-and-save` | router.generate_and_save | JWT Required | 200/min |
| POST | `/security/test/rate-limit` | security_router.test_rate_limit | JWT Required | 10/min |
| POST | `/security/test/threat-detection` | security_router.test_threat_detection | JWT Required | 10/min |
| POST | `/edge-admin/edge-learn/replay` | edge_admin_router.edge_learn_replay | Admin | 50/min |

## Internal/Admin API

| Method | Path | Handler | Auth/Role | Purpose |
|--------|------|---------|-----------|---------|
| GET | `/docs` | FastAPI docs | Public | API documentation |
| GET | `/redoc` | FastAPI redoc | Public | API documentation |
| GET | `/metrics` | Prometheus metrics | Public | Metrics endpoint |

## Dependencies

### In-Cluster
- **MongoDB** (27017) - Tag storage and metadata
- **MME Tagging Service** (8081) - Memory management integration
- **OpenAI API** - LLM for tag extraction

### External
- **OpenAI API** - GPT models for content analysis
- **Python packages** - FastAPI, prometheus, apscheduler

## Config & Secrets Matrix

| Variable | Source | Purpose | Required |
|----------|--------|---------|----------|
| `ENABLE_TAGGING_SERVICE` | env | Service enablement | ✅ |
| `MME_TAGGING_SERVICE_URL` | env | MME tagging service endpoint | ✅ |
| `PYTHON_ENV` | env | Python environment | ✅ |
| `LOG_LEVEL` | env | Logging level | ✅ |
| `OPENAI_API_KEY` | env | OpenAI API access | ✅ |
| `MONGODB_URI` | env | MongoDB connection | ✅ |
| `MONGODB_DATABASE` | env | MongoDB database name | ✅ |
| `MONGODB_COLLECTION` | env | MongoDB collection name | ✅ |

## Data Contracts

### Tag Generation Request
```json
{
  "content": "string",
  "userId": "string",
  "source": "string",
  "metadata": "object"
}
```

### Tag Generation Response
```json
{
  "tags": ["string"],
  "confidence": "number",
  "extraction_method": "string",
  "processing_time": "number"
}
```

### Edge Learning Request
```json
{
  "orgId": "string",
  "learning_rate": "number",
  "batch_size": "number"
}
```

### Security Test Response
```json
{
  "rate_limit_status": "string",
  "threat_detection_status": "string",
  "input_validation": "boolean",
  "timestamp": "timestamp"
}
```

## Workflows

- **Tag Extraction**: Agent output → Tagmaker → OpenAI analysis → Tag extraction → MME Tagging
- **Memory Tiering**: Content analysis → Semantic clustering → Memory organization
- **Edge Learning**: Historical data → Learning algorithm → Edge weight updates
- **Rebalancing**: Scheduled job → Tag rebalancing → Memory optimization

## Security Posture

### ✅ Implemented
- **JWT Authentication**: All endpoints require valid JWT (except health/version)
- **Rate Limiting**: Configurable rate limits per endpoint
- **Input Validation**: Request validation and sanitization
- **Threat Detection**: Security middleware for malicious input detection
- **API Key Protection**: OpenAI API key secured via environment

### 🔄 Residual Risks
- **Content Analysis**: AI model access and data privacy
- **Memory Content**: Sensitive information in agent outputs
- **API Rate Limits**: OpenAI API usage and costs

## Observability

### Health Endpoints
- **Health Check**: `http://localhost:8000/health`
- **Version Info**: `http://localhost:8000/version`
- **Metrics**: `http://localhost:8000/metrics`
- **API Docs**: `http://localhost:8000/docs`

### Key Metrics
- **Tag Extraction**: Success rate and processing time
- **OpenAI Usage**: API calls and response times
- **Memory Operations**: Tag generation and storage metrics
- **Scheduler Jobs**: Rebalancing and learning job success rates

### Dashboards
- **Grafana**: MME Tagmaker Service dashboard with AI metrics
- **Prometheus**: Custom metrics for tag extraction operations

## Scheduled Jobs

### Background Tasks
- **Daily Rebalancing**: 2 AM UTC - Tag rebalancing and optimization
- **Failed Delta Retry**: Every minute - Retry failed memory operations
- **Edge Learning**: Every 10 minutes - Continuous learning updates

## SLOs/SLIs

### Service Level Objectives
- **Availability**: 99.9% uptime
- **Latency**: P95 < 2s for tag extraction, P95 < 500ms for health checks
- **Throughput**: 200 tag extractions/minute

### Service Level Indicators
- **Error Rate**: < 2% for tag extraction operations
- **OpenAI API Success**: > 95% successful API calls
- **Tag Quality**: > 85% relevant tag extraction accuracy

### TODO: Verification Commands
```bash
# Check service health
curl http://localhost:8000/health

# Get service version
curl http://localhost:8000/version

# Test tag generation (requires JWT)
curl -X POST http://localhost:8000/generate-and-save \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"content": "test agent output", "userId": "user123", "source": "agent_output"}'

# Test security features (requires JWT)
curl -X POST http://localhost:8000/security/test/threat-detection \
  -H "Authorization: Bearer <JWT>"

# Get API documentation
curl http://localhost:8000/docs
```
