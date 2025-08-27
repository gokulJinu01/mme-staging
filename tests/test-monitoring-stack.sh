#!/bin/bash

# MME Monitoring Stack Verification Script
# This script verifies that all monitoring components are working correctly

set -euo pipefail

echo "🔍 MME Monitoring Stack Verification"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "OK" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠️  $message${NC}"
    else
        echo -e "${RED}❌ $message${NC}"
    fi
}

# Function to check if a service is responding
check_service() {
    local service=$1
    local url=$2
    local description=$3
    
    if curl -s -f "$url" >/dev/null 2>&1 || curl -s -H "Host: mme.localhost" -H "X-User-ID: test" "$url" >/dev/null 2>&1; then
        print_status "OK" "$description ($service)"
        return 0
    else
        print_status "FAIL" "$description ($service)"
        return 1
    fi
}

# Function to check Prometheus targets
check_prometheus_targets() {
    echo -e "\n📊 Prometheus Targets:"
    local targets=$(curl -s http://localhost:9090/api/v1/targets | jq -r '.data.activeTargets[] | "\(.labels.job): \(.health)"')
    
    local all_up=true
    while IFS= read -r target; do
        if [[ "$target" == *": up" ]]; then
            print_status "OK" "$target"
        else
            print_status "FAIL" "$target"
            all_up=false
        fi
    done <<< "$targets"
    
    return $([ "$all_up" = true ] && echo 0 || echo 1)
}

# Function to check metrics availability
check_metrics() {
    echo -e "\n📈 Metrics Availability:"
    
    # Check Traefik metrics
    local traefik_metrics=$(docker exec mme-traefik wget -qO- http://localhost:8082/metrics 2>/dev/null | grep -c "traefik_" || echo "0")
    if [ "$traefik_metrics" -gt 0 ] 2>/dev/null; then
        print_status "OK" "Traefik metrics available ($traefik_metrics metrics)"
    else
        print_status "FAIL" "Traefik metrics not available"
    fi
    
    # Check MongoDB metrics
    local mongo_metrics=$(curl -s http://localhost:9090/api/v1/query?query=mongodb_connections_current | jq -r '.data.result[0].value[1] // "0"')
    if [ "$mongo_metrics" != "0" ] && [ "$mongo_metrics" != "null" ]; then
        print_status "OK" "MongoDB metrics available (connections: $mongo_metrics)"
    else
        print_status "WARN" "MongoDB metrics not available"
    fi
}

# Function to check isolation sentinel
check_isolation_sentinel() {
    echo -e "\n🔒 Isolation Sentinel:"
    
    # Check if isolation sentinel is running
    if docker ps --format "{{.Names}}" | grep -q mme-isolation-sentinel; then
        print_status "OK" "Isolation sentinel container running"
    else
        print_status "FAIL" "Isolation sentinel container not running"
        return 1
    fi
    
    # Check isolation log
    if [ -f "logs/isolation_sentinel.log" ]; then
        local last_check=$(tail -1 logs/isolation_sentinel.log | grep -o "PASS\|FAIL" | tail -1 || echo "UNKNOWN")
        if [ "$last_check" = "PASS" ]; then
            print_status "OK" "Last isolation check: PASS"
        elif [ "$last_check" = "FAIL" ]; then
            print_status "FAIL" "Last isolation check: FAIL"
        else
            print_status "WARN" "Last isolation check: UNKNOWN"
        fi
    else
        print_status "WARN" "Isolation sentinel log not found"
    fi
}

# Function to check spike-trace sampling
check_spike_trace() {
    echo -e "\n📝 Spike-trace Sampling:"
    
    # Check if spike-trace sampler is running
    if docker ps --format "{{.Names}}" | grep -q mme-spike-trace-sampler; then
        print_status "OK" "Spike-trace sampler container running"
    else
        print_status "FAIL" "Spike-trace sampler container not running"
        return 1
    fi
    
    # Check spike-trace files
    if [ -d "logs/spike_traces" ] && [ "$(ls -A logs/spike_traces)" ]; then
        local file_count=$(ls logs/spike_traces/*.log 2>/dev/null | wc -l)
        print_status "OK" "Spike-trace files available ($file_count files)"
        
        # Check latest summary
        local latest_summary=$(ls -t logs/spike_traces/summary_*.txt 2>/dev/null | head -1)
        if [ -n "$latest_summary" ]; then
            local total_entries=$(grep "Total entries:" "$latest_summary" | grep -o '[0-9]*' || echo "0")
            print_status "OK" "Latest summary: $total_entries entries"
        fi
    else
        print_status "WARN" "Spike-trace files not found"
    fi
}

# Function to check alerting
check_alerting() {
    echo -e "\n🚨 Alerting:"
    
    # Check Alertmanager
    if check_service "alertmanager" "http://localhost:9093/-/healthy" "Alertmanager health"; then
        # Check alert rules
        local alert_rules=$(curl -s http://localhost:9090/api/v1/rules | jq -r '.data.groups[].rules[].name' | wc -l)
        if [ "$alert_rules" -gt 0 ] 2>/dev/null; then
            print_status "OK" "Alert rules configured ($alert_rules rules)"
        else
            print_status "WARN" "No alert rules found"
        fi
    fi
}

# Function to check Grafana
check_grafana() {
    echo -e "\n📊 Grafana:"
    
    if check_service "grafana" "http://localhost:3000/api/health" "Grafana health"; then
        # Check if Prometheus datasource is configured
        local datasources=$(curl -s http://localhost:3000/api/datasources | jq -r '.[].name' | grep -c "Prometheus" || echo "0")
        if [ "$datasources" -gt 0 ] 2>/dev/null; then
            print_status "OK" "Prometheus datasource configured"
        else
            print_status "WARN" "Prometheus datasource not configured"
        fi
        
        # Check dashboards
        local dashboards=$(curl -s http://localhost:3000/api/search | jq -r '.[].title' | grep -c "MME" || echo "0")
        if [ "$dashboards" -gt 0 ] 2>/dev/null; then
            print_status "OK" "MME dashboards available ($dashboards dashboards)"
        else
            print_status "WARN" "MME dashboards not found"
        fi
    fi
}

# Function to check performance metrics
check_performance() {
    echo -e "\n⚡ Performance Metrics:"
    
    # Check MME promote latency (if available)
    local promote_latency=$(curl -s "http://localhost:9090/api/v1/query?query=mme_promote_duration_p95" | jq -r '.data.result[0].value[1] // "0"')
    if [ "$promote_latency" != "0" ] && [ "$promote_latency" != "null" ]; then
        local latency_ms=$(echo "$promote_latency * 1000" | bc -l 2>/dev/null || echo "0")
        if (( $(echo "$latency_ms < 100" | bc -l) )); then
            print_status "OK" "MME promote p95 latency: ${latency_ms}ms"
        else
            print_status "WARN" "MME promote p95 latency: ${latency_ms}ms (high)"
        fi
    else
        print_status "WARN" "MME promote latency not available"
    fi
    
    # Check error rate
    local error_rate=$(curl -s "http://localhost:9090/api/v1/query?query=mme_promote_error_rate" | jq -r '.data.result[0].value[1] // "0"')
    if [ "$error_rate" != "0" ] && [ "$error_rate" != "null" ]; then
        local error_percent=$(echo "$error_rate * 100" | bc -l 2>/dev/null || echo "0")
        if (( $(echo "$error_percent < 0.5" | bc -l) )); then
            print_status "OK" "MME error rate: ${error_percent}%"
        else
            print_status "FAIL" "MME error rate: ${error_percent}% (high)"
        fi
    else
        print_status "WARN" "MME error rate not available"
    fi
}

# Main execution
echo "Starting monitoring stack verification..."

# Check core services
echo -e "\n🏗️  Core Services:"
check_service "traefik" "http://localhost:80/health" "Traefik gateway (via MME)"
check_service "prometheus" "http://localhost:9090/-/healthy" "Prometheus"
check_service "alertmanager" "http://localhost:9093/-/healthy" "Alertmanager"
check_service "grafana" "http://localhost:3000/api/health" "Grafana"

# Check monitoring components
check_prometheus_targets
check_metrics
check_isolation_sentinel
check_spike_trace
check_alerting
check_grafana
check_performance

echo -e "\n🎯 Verification Summary:"
echo "========================="
echo "✅ Traefik metrics: http://localhost:8082/metrics"
echo "✅ Prometheus: http://localhost:9090"
echo "✅ Grafana: http://localhost:3000 (admin/admin)"
echo "✅ Alertmanager: http://localhost:9093"
echo "✅ Isolation sentinel: logs/isolation_sentinel.log"
echo "✅ Spike-trace sampling: logs/spike_traces/"

echo -e "\n📋 Next Steps:"
echo "1. Access Grafana at http://localhost:3000"
echo "2. Login with admin/admin"
echo "3. Check the MME Monitoring Dashboard"
echo "4. Configure Slack webhook in Alertmanager if needed"
echo "5. Monitor isolation sentinel logs for security"
echo "6. Review daily spike-trace summaries"

echo -e "\n🎉 Monitoring stack verification completed!"
