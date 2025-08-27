# MME Observability & Monitoring

## Overview

This document describes the monitoring and observability setup for the MME (Memory Management Engine) system. The monitoring stack provides real-time visibility into system performance, security, and operational health.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Traefik       │    │   Prometheus    │    │   Grafana       │
│   (Metrics)     │───▶│   (Scraping)    │───▶│   (Dashboards)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MongoDB       │    │   Alertmanager  │    │   Isolation     │
│   Exporter      │    │   (Alerts)      │    │   Sentinel      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Components

### 1. Traefik Metrics
- **Endpoint**: `http://traefik:8082/metrics`
- **Metrics**: Request duration histograms, status codes, upstream health
- **Key Metrics**:
  - `traefik_service_request_duration_seconds_bucket`
  - `traefik_service_requests_total`
  - `traefik_service_up`

### 2. Prometheus
- **URL**: `http://localhost:9090`
- **Scrape Interval**: 15s (global), 10s (Traefik)
- **Retention**: 200 hours
- **Recording Rules**: p95/p50 latency, error rates, connection usage

### 3. Grafana
- **URL**: `http://localhost:3000`
- **Credentials**: admin/admin (configurable via `GRAFANA_PASSWORD`)
- **Dashboards**: MME Monitoring Dashboard (auto-provisioned)

### 4. Alertmanager
- **URL**: `http://localhost:9093`
- **Integration**: Slack (configurable via `SLACK_WEBHOOK_URL`)
- **Alerts**: Latency, error rates, isolation failures, MongoDB issues

### 5. MongoDB Exporter
- **Port**: 9216 (internal)
- **Metrics**: Connections, operations, slow queries, queues
- **Key Metrics**:
  - `mongodb_connections_current`
  - `mongodb_connections_available`
  - `mongodb_opcounters_total`

### 6. Isolation Sentinel
- **Frequency**: Every 5 minutes
- **Purpose**: Verify cross-tenant data isolation
- **Output**: Logs to `/logs/isolation_sentinel.log`
- **Metrics**: `mme_isolation_sentinel` gauge (0=fail, 1=pass)

### 7. Spike-trace Sampler
- **Frequency**: Daily at 1 AM
- **Purpose**: Analyze MME promote patterns
- **Output**: `/logs/spike_traces/` directory
- **Files**: Daily logs, analysis, summaries

## Key Metrics & SLOs

### MME Promote Endpoint
- **p95 Latency**: < 100ms
- **Error Rate**: < 0.5%
- **Request Rate**: Monitored for trends

### MongoDB
- **Connection Usage**: < 80%
- **Operation Counters**: Monitored for anomalies
- **Slow Queries**: Alerted on high counts

### Security
- **Isolation Sentinel**: Must pass every check
- **Cross-tenant Leakage**: Zero tolerance

## Alerts

### Critical Alerts (Immediate)
1. **MMEErrorRateHigh**: Error rate > 0.5% for 5 minutes
2. **IsolationSentinelFailure**: Cross-tenant isolation failure
3. **TraefikUpstreamErrors**: Service unavailable

### Warning Alerts (5 minutes)
1. **MMEPromoteHighLatency**: p95 > 100ms
2. **MongoDBConnectionsHigh**: Connection usage > 80%

## Dashboards

### MME Monitoring Dashboard
- **URL**: `http://localhost:3000/d/mme/mme-monitoring-dashboard`
- **Panels**:
  - MME Promote Request Rate
  - MME Promote p95 Latency
  - MME Error Rate
  - MongoDB Connections
  - Isolation Sentinel Status
  - Traefik Request Duration (p95)

## Verification Checklist

### Traefik
- [ ] `curl http://localhost:8082/metrics` returns metrics
- [ ] Grafana panels show Traefik data
- [ ] p50/p95 latency visible per service

### MME
- [ ] `/promote` traffic visible in Grafana
- [ ] p95 latency < 100ms consistently
- [ ] Error rate < 0.5%

### MongoDB
- [ ] Connections and ops visible in Grafana
- [ ] Alert triggers if > 80% pool usage

### Isolation
- [ ] `/logs/isolation_sentinel.log` appends every 5 min
- [ ] FAIL triggers alert (or Slack ping)
- [ ] Grafana shows isolation status

### Spike-trace
- [ ] `/logs/spike_trace_YYYYMMDD.log` created daily
- [ ] Analysis files contain meaningful data

## Troubleshooting

### Prometheus Not Scraping
1. Check service health: `docker-compose ps`
2. Verify network connectivity
3. Check Prometheus targets: `http://localhost:9090/targets`

### Grafana No Data
1. Verify Prometheus datasource
2. Check recording rules: `http://localhost:9090/rules`
3. Validate metric names in queries

### Isolation Sentinel Failing
1. Check MME service logs
2. Verify user isolation logic
3. Review sentinel script logs

### High Latency
1. Check MongoDB connection pool
2. Review MME service resource limits
3. Analyze spike-trace patterns

## Configuration Files

- `traefik.yml`: Traefik metrics configuration
- `monitoring/prometheus.yml`: Prometheus scraping config
- `monitoring/rules.yml`: Recording rules
- `monitoring/alert_rules.yml`: Alert definitions
- `monitoring/alertmanager.yml`: Alert routing
- `scripts/isolation_sentinel.sh`: Isolation checker
- `scripts/spike_trace_sampler.sh`: Trace analyzer

## Environment Variables

```bash
# Required
MONGODB_ROOT_PASSWORD=your_password

# Optional
GRAFANA_PASSWORD=admin
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
```

## Security Notes

- Grafana dashboard should be protected in production
- Alertmanager webhook URLs should be secured
- Isolation sentinel logs contain sensitive test data
- All services run on internal network except Traefik/Grafana

## Performance Notes

- Prometheus retention: 200 hours (adjust based on storage)
- Scrape intervals optimized for real-time monitoring
- Recording rules reduce query load on Grafana
- Log rotation prevents disk space issues
