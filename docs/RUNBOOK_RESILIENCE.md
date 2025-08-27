# MME Resilience Runbook (resilience policy v1 - 2025-08-20)

## Overview
This runbook provides timestamped, copy/paste ready steps for handling MME system resilience issues.

## Quick Reference
- **Degradation Script**: `./scripts/degradation-manager.sh {enable|disable|check}`
- **Traefik Dashboard**: http://localhost:9000
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

---

## 1. If p95 > 100ms for 10 min

**Timestamp**: `$(date '+%Y-%m-%d %H:%M:%S')`

### Immediate Actions:
```bash
# 1. Check current status
./scripts/degradation-manager.sh check

# 2. Enable degradation mode
./scripts/degradation-manager.sh enable

# 3. Verify changes
docker-compose exec mme-tagging-service env | grep -E "(ALLOW_NEIGHBORS|PROMOTE_TOPK|MME_DEGRADATION_MODE)"

# 4. Annotate Grafana
# Go to http://localhost:3000 → Annotations → Add annotation
# Text: "Degradation enabled - p95 > 100ms"
```

### Verification:
```bash
# Check service restarted successfully
docker-compose ps mme-tagging-service

# Monitor p95 improvement
curl -s "http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,rate(traefik_service_request_duration_seconds_bucket{service=\"mme-tagging@docker\"}[5m]))" | jq -r '.data.result[0].value[1]'
```

**Expected**: p95 should drop below 100ms within 2-3 minutes.

---

## 2. If error rate > 0.5%

**Timestamp**: `$(date '+%Y-%m-%d %H:%M:%S')`

### Investigation Steps:
```bash
# 1. Check Traefik retry hits on read routes
curl -s "http://localhost:9090/api/v1/query?query=rate(traefik_service_requests_total{service=\"mme-tagging@docker\",code=\"502\"}[5m])" | jq -r '.data.result[0].value[1]'

# 2. Verify Mongo pool status
curl -s "http://localhost:9090/api/v1/query?query=mongodb_connections{state=\"active\"}" | jq -r '.data.result[0].value[1]'

# 3. Check rate limit spikes
curl -s "http://localhost:9090/api/v1/query?query=rate(traefik_service_requests_total{service=\"mme-tagging@docker\",code=\"429\"}[5m])" | jq -r '.data.result[0].value[1]'

# 4. Review recent config changes
docker-compose logs --tail=50 mme-tagging-service | grep -E "(ERROR|WARN|config|restart)"
```

### Rollback if needed:
```bash
# If recent config change suspected
docker-compose exec mme-tagging-service sh -c "echo 'Rolling back config at $(date)' >> /tmp/rollback.log"
# Restore previous environment variables or restart with defaults
```

---

## 3. If isolation sentinel FAILS

**Timestamp**: `$(date '+%Y-%m-%d %H:%M:%S')`

### Emergency Actions:
```bash
# 1. Restrict access to pilot tenant (if applicable)
# Update Traefik labels to block specific user IDs

# 2. Disable promote routes temporarily
docker-compose exec traefik sh -c "echo 'Promote routes disabled at $(date)' >> /tmp/emergency.log"

# 3. Investigate logs
docker-compose logs --tail=100 mme-tagging-service | grep -E "(user|auth|isolation)"
docker-compose logs --tail=100 traefik | grep -E "(auth|forward|401)"

# 4. Check Traefik routes
curl -s "http://localhost:9000/api/http/routers" | jq '.[] | select(.service=="mme-tagging@docker")'
```

### Post-mortem template:
```bash
# Create post-mortem document
cat > /tmp/isolation-failure-$(date +%Y%m%d-%H%M%S).md << EOF
# Isolation Sentinel Failure - $(date)

## Timeline
- Detected: $(date)
- Actions taken: [list above]
- Root cause: [TBD]
- Resolution: [TBD]

## Investigation
- Logs reviewed: [yes/no]
- Config changes: [list any recent changes]
- User impact: [describe]

## Prevention
- [ ] Review auth middleware
- [ ] Test isolation with new users
- [ ] Update monitoring thresholds
EOF
```

---

## 4. If rate-limit noise

**Timestamp**: `$(date '+%Y-%m-%d %H:%M:%S')`

### Adjustments:
```bash
# 1. Lower burst on /save if legitimate traffic
# Edit traefik.yml - increase burst from 40 to 80
# Restart Traefik: docker-compose restart traefik

# 2. Whitelist CI IP (if applicable)
# Add to Traefik labels: whitelist-source-range

# 3. Monitor rate limit metrics
curl -s "http://localhost:9090/api/v1/query?query=rate(traefik_service_requests_total{service=\"mme-tagging@docker\",code=\"429\"}[5m])" | jq -r '.data.result[0].value[1]'
```

---

## 5. If Redis/Mongo under pressure

**Timestamp**: `$(date '+%Y-%m-%d %H:%M:%S')`

### Investigation:
```bash
# 1. Check connection pool usage
curl -s "http://localhost:9090/api/v1/query?query=mongodb_connections{state=\"active\"}" | jq -r '.data.result[0].value[1]'

# 2. Check slow queries
docker-compose exec mongodb mongosh --eval "db.currentOp({'secs_running': {'$gt': 5}}).pretty()"

# 3. Validate indexes
docker-compose exec mongodb mongosh --eval "db.memories.getIndexes().forEach(printjson)"

# 4. Check Redis memory
docker-compose exec redis redis-cli info memory | grep used_memory_human
```

### Actions:
```bash
# 1. Raise minPoolSize if connections are thrashing
# Update MONGODB_URI in docker-compose.yml: minPoolSize=20
# Restart: docker-compose restart mme-tagging-service

# 2. Scale vertically if needed
# Update docker-compose.yml resources section
# Restart: docker-compose up -d --scale mme-tagging-service=1
```

---

## 6. Performance Pulse Test

**Timestamp**: `$(date '+%Y-%m-%d %H:%M:%S')`

### Quick Health Check:
```bash
# 1. Basic latency test
curl -w "TTOTAL=%{time_total}s\n" -s -o /dev/null \
  -H "Host: mme.localhost" -H "X-User-ID: test_user" \
  "http://localhost/memory/promote?userId=test_user&tags=test&goal=continue"

# 2. Concurrency test
seq 1 10 | xargs -I{} -P10 bash -c \
  'curl -s -H "Host: mme.localhost" -H "X-User-ID: test_user" \
  "http://localhost/memory/promote?userId=test_user&tags=test&goal=continue" > /dev/null'

# 3. Check all targets are up
curl -s "http://localhost:9090/api/v1/targets" | jq '.data.activeTargets[] | select(.health=="up") | .labels.job'
```

**Expected**: TTOTAL < 0.10s, all 10 concurrent requests succeed, all targets up.

---

## 7. Emergency Contacts

- **On-call**: [Add contact]
- **Escalation**: [Add contact]
- **Slack**: #mme-alerts

---

## 8. Recovery Procedures

### Restore Normal Operation:
```bash
# 1. Disable degradation if active
./scripts/degradation-manager.sh disable

# 2. Verify normal parameters
docker-compose exec mme-tagging-service env | grep -E "(ALLOW_NEIGHBORS|PROMOTE_TOPK|MME_DEGRADATION_MODE)"

# 3. Monitor recovery
./scripts/degradation-manager.sh check
```

### Full System Restart (if needed):
```bash
# 1. Stop all services
docker-compose down

# 2. Start core services first
docker-compose up -d mongodb redis

# 3. Wait for health checks
sleep 30

# 4. Start application services
docker-compose up -d mme-tagging-service mme-tagmaker-service

# 5. Start monitoring
docker-compose up -d prometheus alertmanager grafana

# 6. Verify all services
docker-compose ps
```

---

## Notes

- All timestamps should be logged for audit trail
- Always verify changes with monitoring before considering resolved
- Document any deviations from this runbook
- Update this runbook based on lessons learned
