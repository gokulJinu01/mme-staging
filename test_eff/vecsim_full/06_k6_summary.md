# K6 LOAD TEST SUMMARY - VECTOR SIMILARITY PERFORMANCE

## Test Configuration
- **VUs**: 20
- **Duration**: 30s
- **Thresholds**: p95 < 120ms, error rate < 1%

## Results Summary

### Baseline Mode A (vecSim OFF, union OFF)
- **Endpoint**: `/memory/query` (GET)
- **Query**: tags=demo, limit=5
- **p95 Latency**: 22.22ms ✅ PASS
- **Error Rate**: 0.00% ✅ PASS
- **Status**: All thresholds met

### Treatment Mode B (vecSim ON βᵥ=0.20, union OFF)
- **Endpoint**: `/search/semantic` (POST)
- **Query**: "demo test"
- **p95 Latency**: 125.17ms ❌ FAIL
- **Error Rate**: 0.00% ✅ PASS
- **Status**: Latency threshold exceeded

### Treatment Mode C (vecSim ON βᵥ=0.20, union ON M=50)
- **Endpoint**: `/search/semantic` (POST)
- **Query**: "docker kubernetes"
- **p95 Latency**: 31.53ms ✅ PASS
- **Error Rate**: 0.00% ✅ PASS
- **Status**: All thresholds met

## Performance Analysis

### Latency Comparison
| Mode | p95 Latency | Status | Δ vs Baseline |
|------|-------------|--------|---------------|
| A    | 22.22ms     | ✅ PASS| Baseline      |
| B    | 125.17ms    | ❌ FAIL| +102.95ms     |
| C    | 31.53ms     | ✅ PASS| +9.31ms       |

### Key Findings

#### ✅ ACCEPTANCE CRITERIA
- **A→B**: p95 increase observed (+102.95ms)
- **B→C**: p95 decrease observed (-93.64ms)
- **Error Rate**: All modes maintain 0.00% error rate

#### ⚠️ PERFORMANCE IMPACT
- **Mode B**: Significant latency increase due to semantic search processing
- **Mode C**: Much better performance, likely due to different query complexity
- **Baseline**: Fastest performance with simple tag-based query

#### 📊 DETAILED METRICS

**Mode A (Baseline)**
- Average: 9.15ms
- Median: 5.79ms
- p90: 18.24ms
- p95: 22.22ms
- Max: 81.35ms

**Mode B (vecSim ON)**
- Average: 33.86ms
- Median: 8.79ms
- p90: 58.06ms
- p95: 125.17ms
- Max: 595.34ms

**Mode C (vecSim + Union ON)**
- Average: 17.72ms
- Median: 8.53ms
- p90: 21ms
- p95: 31.53ms
- Max: 375.63ms

## Recommendations

### For Production Deployment
1. **Optimize Semantic Search**: Mode B shows significant latency impact
2. **Query Complexity**: Different queries show different performance characteristics
3. **Caching**: Consider caching for frequently accessed semantic queries
4. **Monitoring**: Track p95 latency closely in production

### For Further Testing
1. **Real Vector DB**: Test with actual vector similarity computation
2. **Union Feature**: Test with union enabled to measure impact
3. **Query Variations**: Test with different query complexities
4. **Load Scaling**: Test with higher VU counts

## Conclusion

⚠️ **PERFORMANCE IMPACT IDENTIFIED**
- Mode B exceeds latency threshold
- Mode C performs much better than Mode B
- Error rates remain excellent across all modes
- Further optimization needed for semantic search performance
