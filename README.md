# MME Staging Environment

**Multi-Modal Memory Extractor (MME) - Production Staging Environment**

## Overview

This repository contains the staging deployment configuration for MME, a governed packer system that provides deterministic, diverse, and budget-controlled memory management with vector similarity integration.

## Architecture

- **MME Tagging Service** (Go/Fiber) - Core packer service
- **MME Tagmaker Service** (Python/FastAPI) - AI tagging service
- **MongoDB** - Primary database
- **Redis** - Caching layer
- **Traefik** - API gateway with authentication
- **JWT Verifier** - Authentication service
- **Monitoring Stack** - Prometheus, Grafana, Alertmanager

## Quick Start

### Prerequisites

- Docker & Docker Compose
- GCP project with required APIs enabled
- Static IP: `34.58.167.157`
- Domain or nip.io hostnames configured

### Deployment

1. **Clone and setup**:
   ```bash
   git clone git@github.com:gokulJinu01/mme-staging.git
   cd mme-staging
   ```

2. **Configure environment**:
   ```bash
   cp deployment/config/env.example deployment/.env
   # Edit deployment/.env with your configuration
   ```

3. **Deploy**:
   ```bash
   cd deployment
   docker compose up -d
   ```

4. **Verify**:
   ```bash
   docker compose ps
   curl -H "Host: mme.34.58.167.157.nip.io" http://34.58.167.157/health
   ```

## Configuration

### Environment Variables

Key configuration in `deployment/.env`:

```bash
# Hosts
TRAEFIK_HOST_MME=mme.34.58.167.157.nip.io
TRAEFIK_HOST_TAGMAKER=tagmaker.34.58.167.157.nip.io

# vecSim & union (Mode C - Production Ready)
MME_VECSIM_ENABLED=true
MME_BETA_VECTOR_SIMILARITY=0.20
MME_UNION_ENABLED=true
MME_UNION_TOP_M=50

# Governance
MME_PACKER_JACCARD_HARD_CUT=0.95
MME_PACKER_DIVERSITY_MODE=soft
MME_PACKER_TOKEN_BUDGET=2048
```

### Modes

- **Mode A**: vecSim OFF, union OFF (Baseline)
- **Mode B**: vecSim ON, union OFF (vecSim only)
- **Mode C**: vecSim ON, union ON (Production - Recommended)

## API Endpoints

### MME Tagging Service
- `GET /health` - Health check
- `GET /memory/query?tags=<tags>&limit=<n>` - Query memories
- `POST /memory/save` - Save memory
- `POST /search/semantic` - Semantic search
- `GET /memory/tokenizer-health` - Tokenizer status

### Tagmaker Service
- `GET /health` - Health check
- `POST /tags/generate` - Generate tags
- `GET /tags/learn` - Edge learning

## Performance

### Benchmarked Results
- **Mode A**: 12ms p95 (baseline)
- **Mode B**: 20ms p95 (vecSim only)
- **Mode C**: 32ms p95 (production - recommended)

### Quality Metrics
- **P@10**: 0.89-0.91
- **nDCG@10**: 0.86-0.88
- **Recall@K**: 0.91-0.93

## Monitoring

### Dashboards
- **Grafana**: http://34.58.167.157:3000 (admin/admin)
- **Prometheus**: http://34.58.167.157:9090
- **Traefik**: http://34.58.167.157:9000

### Key Metrics
- Response time (p95 < 120ms)
- Error rate (< 1%)
- Token budget compliance
- Diversity breaches (0 expected)
- Determinism (perfect)

## Testing

### Load Testing
```bash
# Install k6
curl -L https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz | tar xz
sudo cp k6-v0.47.0-linux-amd64/k6 /usr/local/bin/

# Run tests
k6 run test_eff/gcp_staging/k6_mme_staging.js
```

### Health Checks
```bash
# Service health
curl -H "Host: mme.34.58.167.157.nip.io" http://34.58.167.157/health

# API test
curl -H "Host: mme.34.58.167.157.nip.io" \
     -H "X-User-ID: test_user" \
     http://34.58.167.157/memory/query?tags=demo&limit=5
```

## Troubleshooting

### Common Issues

1. **Services not starting**:
   ```bash
   docker compose logs -f
   ```

2. **Health checks failing**:
   ```bash
   docker compose ps
   docker compose restart <service>
   ```

3. **Performance issues**:
   ```bash
   docker stats
   docker compose logs mme-tagging-service
   ```

### Rollback Procedures

1. **Disable vecSim**:
   ```bash
   # Edit deployment/.env
   MME_VECSIM_ENABLED=false
   MME_UNION_ENABLED=false
   docker compose restart mme-tagging-service
   ```

2. **Switch tokenizer**:
   ```bash
   # Edit deployment/.env
   MME_PACKER_TOKENIZER_MODE=heuristic
   docker compose restart mme-tagging-service
   ```

## Security

- JWT authentication required for all endpoints
- Secrets managed via environment variables
- Network isolation via Docker networks
- Security headers via Traefik middleware

## Support

For issues or questions:
1. Check logs: `docker compose logs -f`
2. Verify configuration: `docker compose config`
3. Test endpoints: Use provided health checks
4. Review monitoring: Check Grafana dashboards

## License

Proprietary - MME Staging Environment
# mme-staging
