# VECTOR SIMILARITY FULL-CYCLE VALIDATION RUN SHEET

## Configuration Modes

### Baseline A: vecSim OFF, union OFF
```bash
# Current state - no additional env vars needed
# MME_BETA_VECTOR_SIMILARITY defaults to 0.0
# MME_VECSIM_ENABLED defaults to false
# MME_UNION_ENABLED defaults to false
```

### Treatment B: vecSim ON βᵥ=0.20, union OFF
```bash
# Set in docker-compose.yml or environment
MME_BETA_VECTOR_SIMILARITY=0.20
MME_VECSIM_ENABLED=true
MME_UNION_ENABLED=false
```

### Treatment C: vecSim ON βᵥ=0.20, union ON M=50
```bash
# Set in docker-compose.yml or environment
MME_BETA_VECTOR_SIMILARITY=0.20
MME_VECSIM_ENABLED=true
MME_UNION_ENABLED=true
MME_UNION_TOP_M=50
```

## Rollback Instructions

### Emergency Rollback (vecSim OFF)
```bash
# Set these environment variables:
MME_BETA_VECTOR_SIMILARITY=0.0
MME_VECSIM_ENABLED=false
MME_UNION_ENABLED=false

# Restart services:
docker compose restart mme-tagging-service mme-tagmaker-service
```

### Partial Rollback (vecSim ON, union OFF)
```bash
# Keep vecSim but disable union:
MME_BETA_VECTOR_SIMILARITY=0.20
MME_VECSIM_ENABLED=true
MME_UNION_ENABLED=false

# Restart services:
docker compose restart mme-tagging-service mme-tagmaker-service
```

### Conservative Rollback (lower βᵥ)
```bash
# Reduce vecSim weight:
MME_BETA_VECTOR_SIMILARITY=0.15
MME_VECSIM_ENABLED=true
MME_UNION_ENABLED=true
MME_UNION_TOP_M=30

# Restart services:
docker compose restart mme-tagging-service mme-tagmaker-service
```

## Test Queries for Semantic Validation

1. **Synonyms**: "docker" vs "containerization"
2. **Paraphrases**: "machine learning" vs "ML algorithms"
3. **Related concepts**: "redis" vs "in-memory database"
4. **Technical variations**: "kubernetes" vs "k8s orchestration"
5. **Domain terms**: "vector database" vs "embedding storage"

## Golden Pairs for Quality Evaluation

1. `rag:vector-db`
2. `mme:token-budget`
3. `mse:llm-security`
4. `embedding:ranker`
5. `docker:kubernetes`
6. `redis:mongo`
7. `grant:proposal`
8. `recency:importance`

## Acceptance Criteria

- **Recall@200**: ≥ 0.80 target
- **P@10**: ≥ 0.60 target, expect lift from A→B
- **Determinism**: 20 identical queries → 1 unique result
- **Diversity**: 0 Jaccard breaches
- **Budget**: ΣtokenCost ≤ tokenBudget
- **SLO p95**: A→B unchanged, B→C bounded increase
- **Envelopes**: All business endpoints standardized
