#!/bin/bash

# Spike-trace Sampler - Collect daily samples of spike_trace logs
# This script extracts and analyzes spike_trace patterns from MME logs

set -euo pipefail

# Configuration
LOG_DIR="/logs"
SAMPLE_DIR="/logs/spike_traces"
ROTATION_DAYS=7

# Create directories
mkdir -p "$SAMPLE_DIR"

# Get today's date
TODAY=$(date '+%Y%m%d')
YESTERDAY=$(date -d 'yesterday' '+%Y%m%d' 2>/dev/null || date -v-1d '+%Y%m%d' 2>/dev/null || date '+%Y%m%d')

# Function to log with timestamp
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

# Extract spike_trace logs from MME service
extract_spike_traces() {
    local output_file="$1"
    
    log "Extracting spike_trace logs from MME service..."
    
    # Get logs from the last 24 hours and extract spike_trace lines
    docker logs --since "24h" mme-tagging-service 2>&1 | \
        grep "spike_trace" | \
        tail -1000 > "$output_file"
    
    local count=$(wc -l < "$output_file")
    log "Extracted $count spike_trace entries"
}

# Analyze spike_trace patterns
analyze_patterns() {
    local input_file="$1"
    local analysis_file="$2"
    
    log "Analyzing spike_trace patterns..."
    
    # Count occurrences by seed and tier
    echo "# Spike-trace Analysis - $(date '+%Y-%m-%d %H:%M:%S')" > "$analysis_file"
    echo "" >> "$analysis_file"
    
    # Extract and count seed patterns
    echo "## Top Seeds by Frequency:" >> "$analysis_file"
    grep -o 'seed=\[[^]]*\]' "$input_file" | \
        sort | uniq -c | sort -nr | head -10 | \
        while read count pattern; do
            echo "$count: $pattern" >> "$analysis_file"
        done
    
    echo "" >> "$analysis_file"
    
    # Extract and count tier patterns
    echo "## Tier Distribution:" >> "$analysis_file"
    grep -o 'tier=[^[:space:]]*' "$input_file" | \
        sort | uniq -c | sort -nr | \
        while read count tier; do
            echo "$count: $tier" >> "$analysis_file"
        done
    
    echo "" >> "$analysis_file"
    
    # Performance trends analysis
    echo "## Performance Trends Analysis:" >> "$analysis_file"
    local direct_tier=$(grep "tier=direct" "$input_file" | wc -l)
    local fallback_usage=$(grep "tier=neighbors\|tier=recent" "$input_file" | wc -l)
    
    echo "Direct tier usage: $direct_tier" >> "$analysis_file"
    echo "Fallback tier usage: $fallback_usage" >> "$analysis_file"
    
    if [ "$fallback_usage" -gt 0 ]; then
        local fallback_percent=$(echo "scale=2; $fallback_usage * 100 / ($direct_tier + $fallback_usage)" | bc -l 2>/dev/null || echo "0")
        echo "Fallback usage percentage: ${fallback_percent}%" >> "$analysis_file"
        
        if (( $(echo "$fallback_percent > 20" | bc -l 2>/dev/null || echo "0") )); then
            echo "⚠️  WARNING: High fallback usage detected (>20%)" >> "$analysis_file"
        fi
    fi
    
    echo "" >> "$analysis_file"
    
    # Show sample entries
    echo "## Sample Entries:" >> "$analysis_file"
    head -20 "$input_file" >> "$analysis_file"
}

# Rotate old files
rotate_old_files() {
    log "Rotating files older than $ROTATION_DAYS days..."
    
    find "$SAMPLE_DIR" -name "spike_trace_*.log" -mtime +$ROTATION_DAYS -delete
    find "$SAMPLE_DIR" -name "spike_analysis_*.txt" -mtime +$ROTATION_DAYS -delete
}

# Main execution
log "Starting spike-trace sampling..."

# Extract today's spike traces
today_log="$SAMPLE_DIR/spike_trace_$TODAY.log"
today_analysis="$SAMPLE_DIR/spike_analysis_$TODAY.txt"

extract_spike_traces "$today_log"
analyze_patterns "$today_log" "$today_analysis"

# Rotate old files
rotate_old_files

# Create a summary
summary_file="$SAMPLE_DIR/summary_$TODAY.txt"
{
    echo "Spike-trace Daily Summary - $TODAY"
    echo "================================"
    echo ""
    echo "Total entries: $(wc -l < "$today_log")"
    echo "Unique seeds: $(grep -o 'seed=\[[^]]*\]' "$today_log" | sort -u | wc -l)"
    echo "Tiers used: $(grep -o 'tier=[^[:space:]]*' "$today_log" | sort -u | wc -l)"
    echo ""
    echo "Top 5 seeds:"
    grep -o 'seed=\[[^]]*\]' "$today_log" | sort | uniq -c | sort -nr | head -5
    echo ""
    echo "Tier distribution:"
    grep -o 'tier=[^[:space:]]*' "$today_log" | sort | uniq -c | sort -nr
} > "$summary_file"

log "Spike-trace sampling completed successfully"
log "Files created:"
log "  - $today_log"
log "  - $today_analysis"
log "  - $summary_file"
