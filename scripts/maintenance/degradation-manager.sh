#!/bin/bash

# MME Degradation Manager (resilience policy v1 - 2025-08-20)
# Handles graceful degradation based on Prometheus alerts

set -e

# Configuration
COMPOSE_FILE="docker-compose.yml"
SERVICE_NAME="mme-tagging-service"
LOG_FILE="./logs/degradation.log"
METRICS_URL="http://prometheus:9090"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if we're in degradation mode
is_degraded() {
    docker-compose exec -T "$SERVICE_NAME" env | grep -q "MME_DEGRADATION_MODE=true"
}

# Enable degradation mode
enable_degradation() {
    log "ENABLING DEGRADATION MODE"
    
    # Update environment variables
    docker-compose exec -T "$SERVICE_NAME" sh -c "
        export ALLOW_NEIGHBORS=false
        export PROMOTE_TOPK=5
        export MME_DEGRADATION_MODE=true
        echo 'Degradation enabled at $(date)' >> /tmp/degradation.log
    "
    
    # Restart service to apply changes
    docker-compose restart "$SERVICE_NAME"
    
    log "Degradation mode enabled - neighbors disabled, topK reduced to 5"
}

# Disable degradation mode
disable_degradation() {
    log "DISABLING DEGRADATION MODE"
    
    # Update environment variables
    docker-compose exec -T "$SERVICE_NAME" sh -c "
        export ALLOW_NEIGHBORS=true
        export PROMOTE_TOPK=10
        export MME_DEGRADATION_MODE=false
        echo 'Degradation disabled at $(date)' >> /tmp/degradation.log
    "
    
    # Restart service to apply changes
    docker-compose restart "$SERVICE_NAME"
    
    log "Degradation mode disabled - normal operation restored"
}

# Check current performance metrics
check_performance() {
    local p95_query="rate(traefik_service_request_duration_seconds{service=\"mme-tagging@docker\"}[5m])"
    local p95_result=$(curl -s "$METRICS_URL/api/v1/query?query=$p95_query" | jq -r '.data.result[0].value[1] // "0"')
    
    local error_query="rate(traefik_service_requests_total{service=\"mme-tagging@docker\",code=~\"5..\"}[5m])"
    local error_result=$(curl -s "$METRICS_URL/api/v1/query?query=$error_query" | jq -r '.data.result[0].value[1] // "0"')
    
    echo "p95_latency=$p95_result,error_rate=$error_result"
}

# Main degradation logic
main() {
    case "${1:-check}" in
        "enable")
            enable_degradation
            ;;
        "disable")
            disable_degradation
            ;;
        "check")
            if is_degraded; then
                log "Current status: DEGRADED"
            else
                log "Current status: NORMAL"
            fi
            check_performance
            ;;
        *)
            echo "Usage: $0 {enable|disable|check}"
            exit 1
            ;;
    esac
}

main "$@"
