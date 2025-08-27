# BENCHMARK MODES CONFIGURATION

## Mode Definitions

### L0: Vector top-K only
- **Description**: Naive vector similarity top-K (no packer)
- **Configuration**: 
  - vecSim OFF
  - union OFF
  - Take top-K from vector candidates only
- **Expected**: Fast, non-deterministic, no budget governance

### L1: BM25 top-K only  
- **Description**: Naive BM25 top-K (no packer)
- **Configuration**:
  - vecSim OFF
  - union OFF
  - Take top-K from BM25 candidates only
- **Expected**: Fast, non-deterministic, no budget governance

### L2: Hybrid top-K (simple fusion)
- **Description**: Simple fusion of vector and BM25 (no packer)
- **Configuration**:
  - vecSim OFF
  - union OFF
  - Interleave or fuse vector + BM25 candidates
- **Expected**: Moderate speed, non-deterministic, no budget governance

### L3: MME (tags/graph only)
- **Description**: MME with vecSim OFF, union OFF
- **Configuration**:
  - MME_BETA_VECTOR_SIMILARITY=0.0
  - MME_VECSIM_ENABLED=false
  - MME_UNION_ENABLED=false
- **Expected**: Deterministic, governed, tag-based only

### L4: MME (vecSim ON βᵥ=0.2, union OFF)
- **Description**: MME with vector similarity enabled
- **Configuration**:
  - MME_BETA_VECTOR_SIMILARITY=0.2
  - MME_VECSIM_ENABLED=true
  - MME_UNION_ENABLED=false
- **Expected**: Deterministic, governed, vector + tag hybrid

### L5: MME (vecSim ON βᵥ=0.2, union ON M=50)
- **Description**: MME with vector similarity and union enabled
- **Configuration**:
  - MME_BETA_VECTOR_SIMILARITY=0.2
  - MME_VECSIM_ENABLED=true
  - MME_UNION_ENABLED=true
  - MME_UNION_TOP_M=50
- **Expected**: Deterministic, governed, vector + tag + union

## Rollback Commands

### L0/L1/L2 (Baseline Retrievers)
```bash
# No MME configuration needed - direct retriever calls
```

### L3 (MME Baseline)
```bash
MME_BETA_VECTOR_SIMILARITY=0.0
MME_VECSIM_ENABLED=false
MME_UNION_ENABLED=false
docker compose restart mme-tagging-service mme-tagmaker-service
```

### L4 (MME + vecSim)
```bash
MME_BETA_VECTOR_SIMILARITY=0.2
MME_VECSIM_ENABLED=true
MME_UNION_ENABLED=false
docker compose restart mme-tagging-service mme-tagmaker-service
```

### L5 (MME + vecSim + Union)
```bash
MME_BETA_VECTOR_SIMILARITY=0.2
MME_VECSIM_ENABLED=true
MME_UNION_ENABLED=true
MME_UNION_TOP_M=50
docker compose restart mme-tagging-service mme-tagmaker-service
```

## Test Sequence
1. L0 → L1 → L2 (baseline retrievers)
2. L3 → L4 → L5 (MME modes)
3. Cross-comparison and leaderboard generation
