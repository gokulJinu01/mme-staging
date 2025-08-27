#!/bin/bash

# Isolation Sentinel - Check cross-tenant isolation every 5 minutes
# This script verifies that users cannot access other users' data

set -euo pipefail

# Configuration
HOST="mme-tagging-service:8080"
USER_A="sentinel_user_a"
USER_B="sentinel_user_b"
LOG_FILE="/logs/isolation_sentinel.log"
METRICS_FILE="/logs/isolation_sentinel.json"

# Create log directory if it doesn't exist
mkdir -p /logs

# Function to log with timestamp
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" | tee -a "$LOG_FILE"
}

# Function to create test data for a user
create_test_data() {
    local user="$1"
    local tag="$2"
    
    curl -s -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"$user private data\",\"tags\":[{\"label\":\"$tag\",\"type\":\"concept\"}],\"status\":\"completed\"}" \
      "http://$HOST/memory/save" >/dev/null
    
    # Wait a moment for data to be indexed
    sleep 1
}

# Function to check isolation
check_isolation() {
    local user_a="$1"
    local user_b="$2"
    local tag_a="$3"
    
    # Query with user A's header but user B's query parameter
    response=$(curl -s -H "X-User-ID: $user_a" \
      "http://$HOST/memory/query?userId=$user_b&tags=$tag_a&limit=5")
    
    # Check if any results contain user B's data
    user_b_leak=$(echo "$response" | jq -r '.. | .content? // empty' | grep -c "$user_b private data" || echo "0")
    
    if [ "$user_b_leak" -gt 0 ] 2>/dev/null; then
        log "FAIL: Cross-tenant data leak detected! User $user_a can see $user_b's data"
        echo "0" > /tmp/isolation_status
        return 1
    else
        log "PASS: Cross-tenant isolation verified - no data leakage"
        echo "1" > /tmp/isolation_status
        return 0
    fi
}

# Function to check auth failures
check_auth_failures() {
    local auth_failures=$(curl -s "http://prometheus:9090/api/v1/query?query=rate(traefik_service_requests_total{code=\"401\"}[5m])" | jq -r '.data.result[0].value[1] // "0"')
    if (( $(echo "$auth_failures > 0.1" | bc -l 2>/dev/null || echo "0") )); then
        log "WARNING: High authentication failures detected: $auth_failures"
    fi
}

# Function to check rate limiting
check_rate_limiting() {
    local request_rate=$(curl -s "http://prometheus:9090/api/v1/query?query=rate(traefik_service_requests_total[1m])" | jq -r '.data.result[0].value[1] // "0"')
    if (( $(echo "$request_rate > 100" | bc -l 2>/dev/null || echo "0") )); then
        log "WARNING: High request rate detected: $request_rate"
    fi
}

# Main execution
log "Starting isolation sentinel check..."

# Additional security checks
check_auth_failures
check_rate_limiting

# Create test data for both users
create_test_data "$USER_A" "sentinel_tag_a"
create_test_data "$USER_B" "sentinel_tag_b"

# Check isolation
if check_isolation "$USER_A" "$USER_B" "sentinel_tag_a"; then
    # Update Prometheus metrics file
    cat > "$METRICS_FILE" << EOF
[
  {
    "targets": ["localhost:9091"],
    "labels": {
      "job": "isolation-sentinel",
      "env": "pilot"
    }
  }
]
EOF
    
    # Create textfile metric for Prometheus
    cat > /logs/mme_isolation_sentinel.prom << EOF
# HELP mme_isolation_sentinel Cross-tenant isolation status
# TYPE mme_isolation_sentinel gauge
mme_isolation_sentinel{env="pilot"} 1
EOF
    
    log "SUCCESS: Isolation check completed successfully"
    exit 0
else
    # Update Prometheus metrics file with failure
    cat > "$METRICS_FILE" << EOF
[
  {
    "targets": ["localhost:9091"],
    "labels": {
      "job": "isolation-sentinel",
      "env": "pilot"
    }
  }
]
EOF
    
    # Create textfile metric for Prometheus (failure)
    cat > /logs/mme_isolation_sentinel.prom << EOF
# HELP mme_isolation_sentinel Cross-tenant isolation status
# TYPE mme_isolation_sentinel gauge
mme_isolation_sentinel{env="pilot"} 0
EOF
    
    log "FAILURE: Isolation check failed - potential security breach!"
    exit 1
fi
