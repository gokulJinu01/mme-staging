# MME (Memory Management Engine) - Demo

A **Memory Management Engine** that enhances RAG systems with precise, explainable, tag-scoped recall, secured behind Traefik ForwardAuth with enterprise-grade resilience and observability.

## 🎯 What It Does

MME enhances RAG by providing:
- **Precise, tag-scoped recall** (better than vector similarity)
- **Explainable results** (spike_trace logging)
- **Graph-based relationships** (tag edges)
- **Quality gates** (precision mode)
- **Multi-tenant isolation** (secure, scoped access)

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- 4GB RAM minimum
- OpenAI API Key

### 1. Setup Environment
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your OpenAI API key
nano .env
```

### 2. Start Services
```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps
```

### 3. Verify Installation
```bash
# Test basic functionality
curl -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  "http://localhost/health"
```

## 📊 Demo URLs

- **MME API**: `http://localhost/` (via Traefik)
- **Grafana**: `http://localhost:3000` (admin/admin)
- **Prometheus**: `http://localhost:9090`
- **Traefik Dashboard**: `http://localhost:9000`

## 🧪 Demo Test Cases

### Basic Functionality
```bash
# Test structured tagging
curl -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  -H "Content-Type: application/json" \
  -d '{"content":"IRAP funding submission","tags":[{"label":"IRAP","type":"concept"},{"label":"funding","type":"object"}],"status":"completed"}' \
  "http://localhost/memory/save"

# Test tag-scoped recall
curl -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  "http://localhost/memory/promote?userId=demo_user&tags=IRAP&goal=continue"
```

### Multi-Tenant Isolation
```bash
# User A saves data
curl -H "Host: mme.localhost" -H "X-User-ID: user_a" \
  -H "Content-Type: application/json" \
  -d '{"content":"User A private data","tags":[{"label":"private"}],"status":"completed"}' \
  "http://localhost/memory/save"

# User B cannot access User A's data
curl -H "Host: mme.localhost" -H "X-User-ID: user_b" \
  "http://localhost/memory/query?userId=user_b&tags=private&limit=5"
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Traefik       │    │   JWT Verifier  │    │   MME Tagging   │
│   (Gateway)     │───▶│   (Auth)        │───▶│   (Core Logic)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Grafana       │    │   Prometheus    │    │   MongoDB       │
│   (Dashboards)  │    │   (Metrics)     │    │   (Storage)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔧 Key Features

### Core MME
- **Save/Query/Promote/Delete** operations
- **Structured tags** with metadata
- **Tag edges** for graph relationships
- **Precision mode** for quality gates

### Security
- **Traefik ForwardAuth** gateway
- **Multi-tenant isolation**
- **No direct app port exposure**

### Resilience
- **Timeouts** (5s max)
- **Retries** (idempotent routes only)
- **Rate limiting** (20 rps / 40 burst)
- **Graceful degradation**

### Observability
- **Prometheus** metrics collection
- **Grafana** dashboards
- **Alertmanager** alerting
- **Isolation sentinel** monitoring

## 📁 Project Structure

```
mme/
├── docker-compose.yml          # Main orchestration
├── traefik.yml                 # Gateway configuration
├── .env                        # Environment variables
├── .env.example               # Environment template
├── docs/                       # Documentation
│   ├── api/                   # API documentation
│   ├── deployment/            # Deployment guides
│   ├── monitoring/            # Monitoring setup
│   └── resilience/            # Resilience procedures
├── scripts/                    # Operational scripts
│   ├── setup/                 # Setup scripts
│   ├── maintenance/           # Maintenance scripts
│   └── backup/                # Backup scripts
├── tests/                      # Test suites
│   ├── integration/           # Integration tests
│   ├── performance/           # Performance tests
│   └── security/              # Security tests
├── monitoring/                 # Monitoring configs
├── mme-tagging-service/        # Core MME service (Go)
├── mme-tagmaker-service/       # AI tagging service (Python)
├── jwt-verifier/              # Authentication service
└── logs/                      # Application logs
```

## 🛠️ Operations

### Health Checks
```bash
# Check all services
docker-compose ps

# Check MME health
curl -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  "http://localhost/health"
```

### Monitoring
```bash
# View Grafana dashboards
open http://localhost:3000

# Check Prometheus targets
open http://localhost:9090/targets
```

### Resilience Testing
```bash
# Run resilience test suite
./tests/test-resilience-features.sh

# Check degradation status
./scripts/maintenance/degradation-manager.sh check
```

### Logs
```bash
# View MME service logs
docker-compose logs mme-tagging-service

# View Traefik logs
docker-compose logs traefik
```

## 🔍 Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   # Check logs
   docker-compose logs <service-name>
   
   # Restart service
   docker-compose restart <service-name>
   ```

2. **Authentication fails**
   ```bash
   # Check JWT verifier
   docker-compose logs jwt-verifier
   
   # Verify environment variables
   docker-compose exec mme-tagging-service env | grep JWT
   ```

3. **Performance issues**
   ```bash
   # Check resource usage
   docker stats
   
   # Enable degradation mode
   ./scripts/maintenance/degradation-manager.sh enable
   ```

### Getting Help

- **Documentation**: See `docs/` directory
- **Runbook**: `docs/resilience/RUNBOOK_RESILIENCE.md`
- **API Docs**: `docs/api/`
- **Monitoring**: `docs/monitoring/`

## 📈 Performance

- **Typical Response**: < 30ms
- **P95 Target**: < 100ms
- **Concurrency**: 10x bursts supported
- **Rate Limiting**: 20 rps average, 40 rps burst

## 🔒 Security

- **Authentication**: JWT-based via Traefik ForwardAuth
- **Isolation**: Zero cross-tenant data leakage
- **Network**: No direct service port exposure
- **Monitoring**: Security events logged and alerted

## 📄 License

See `LICENSE` file for details.

---

**Ready for production deployment with enterprise-grade resilience and observability.**
