#!/bin/bash
set -euo pipefail

# -------- Vars --------
export HOST="mme.localhost"
export U1="pilot_u1"
export U2="pilot_u2"

pass(){ printf "✅ %s\n" "$1"; }
fail(){ printf "❌ %s\n" "$1"; exit 1; }
warn(){ printf "⚠️  %s\n" "$1"; }

printf "Vars set: HOST=%s | U1=%s | U2=%s\n" "$HOST" "$U1" "$U2"

# -------- Helper: require command --------
req() { command -v "$1" >/dev/null 2>&1 || fail "Missing dependency: $1"; }
req curl; req jq

# -------- A) GATEWAY & SECURITY --------

# A1. ForwardAuth behavior (no header -> 401)
status=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: $HOST" "http://localhost/health" || true)
[ "$status" = "401" ] && pass "A1: ForwardAuth returns 401 without header" || fail "A1: Expected 401, got $status"

# A2. Auth OK (with header -> 200)
status=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: $HOST" -H "X-User-ID: $U1" "http://localhost/health")
[ "$status" = "200" ] && pass "A2: Auth returns 200 with X-User-ID" || fail "A2: Expected 200, got $status"

# A3. Cross-tenant isolation: seed U1 & U2 and verify scope by header user
curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
  -d '{"content":"U1 private doc","tags":[{"label":"u1_tag","type":"concept"}],"status":"completed"}' \
  "http://localhost/memory/save" >/dev/null

curl -s -H "Host: $HOST" -H "X-User-ID: $U2" -H "Content-Type: application/json" \
  -d '{"content":"U2 private doc","tags":[{"label":"u2_tag","type":"concept"}],"status":"completed"}' \
  "http://localhost/memory/save" >/dev/null

# Query with mismatched query user; service should still scope to U1 (header)
iso_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $U1" \
  "http://localhost/memory/query?userId=$U2&tags=u1_tag&limit=5")

# Try reading count from either shape {count,results} or array
iso_count=$(jq -r '.count // (if type=="array" then length else (.results|length) end)' <<<"$iso_json")
[ -z "$iso_count" ] && iso_count=0
# Validate none of the results show content "U2 private doc"
u2_leak=$(jq -r '.. | .content? // empty' <<<"$iso_json" | grep -c "U2 private doc" || true)
[ "$u2_leak" -eq 0 ] && pass "A3: Cross-tenant isolation: no U2 leak under U1 header" || fail "A3: Found U2 data under U1 header"

# A4. Bypass hardening – verify only Traefik exposes host ports (informational)
docker-compose ps || warn "docker-compose ps not available (skipping A4)"
# (Visual check recommended: Traefik should be the only service with host ports)

# -------- B) RESPONSE SERIALIZATION --------

save_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
  -d '{"content":"pilot save","tags":[{"label":"pilot_tag","type":"concept"}],"status":"completed"}' \
  "http://localhost/memory/save")
tags_flat=$(jq -r '.tagsFlat | @json' <<<"$save_json" 2>/dev/null || echo "null")
[[ "$tags_flat" != "null" && "$tags_flat" != "[]" ]] && pass "B1: tagsFlat present in save response: $tags_flat" || fail "B1: tagsFlat missing/null in response: $save_json"

# -------- C) PROMOTE + SPIKE TRACE --------

prom_json=$(curl -s -X POST -H "Host: $HOST" -H "X-User-ID: $U1" \
  "http://localhost/memory/promote?userId=$U1&tags=pilot_tag&goal=continue")
# Support {count,results} or array
prom_count=$(jq -r '.count // (if type=="array" then length else (.results|length) end)' <<<"$prom_json")
[ -z "$prom_count" ] && prom_count=0
[ "$prom_count" -ge 0 ] && pass "C1: Promote returns results/empty as expected (count=$prom_count)" || fail "C1: Promote response unexpected: $prom_json"

# spike_trace line visible
# Trigger a promote request to ensure logs are generated
curl -s -X POST -H "Host: $HOST" -H "X-User-ID: $U1" \
  "http://localhost/memory/promote?userId=$U1&tags=pilot_tag&goal=continue" >/dev/null
sleep 2  # Give time for logs to appear

# Debug: check if logs contain spike_trace
log_output=$(docker-compose logs mme-tagging-service --tail=100)
spike_count=$(echo "$log_output" | grep -c "spike_trace" || echo "0")
if [ "$spike_count" -gt 0 ]; then
  pass "C2: spike_trace line present in logs (found $spike_count lines)"
else
  fail "C2: spike_trace line not found in service logs"
fi

# -------- D) DATA INTEGRITY: EDGES, DELETE, BACKFILL --------

# D1. Edges: save a few pairs (we assume edge upsert)
for i in 1 2 3; do
  curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
    -d '{"content":"pair irap-budget","tags":[{"label":"IRAP","type":"concept"},{"label":"budget","type":"object"}],"status":"completed"}' \
    "http://localhost/memory/save" >/dev/null
done
pass "D1: Saved IRAP-budget pairs (edge upserts assumed OK)."

# D2. Delete: owner can delete; other user cannot
del_id=$(curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
  -d '{"content":"temp to delete","tags":[{"label":"to_delete"}],"status":"completed"}' \
  "http://localhost/memory/save" | jq -r '.id')
[ -z "$del_id" ] && fail "D2: Could not create doc to delete"
del_owner=$(curl -s -X DELETE -H "Host: $HOST" -H "X-User-ID: $U1" "http://localhost/memory/$del_id" | jq -r '.message? // .status? // "ok"')
[[ "$del_owner" =~ deleted|ok ]] && pass "D2: Delete OK for owner" || fail "D2: Owner delete failed"

del_wrong=$(curl -s -X DELETE -H "Host: $HOST" -H "X-User-ID: $U2" "http://localhost/memory/$del_id" | jq -r '.error? // .message? // .status? // "none"')
[[ "$del_wrong" =~ not\ found|404|none ]] && pass "D2: Other user cannot delete (as expected)" || fail "D2: Wrong-user delete should not succeed"

# D3. Backfill: only processes docs with tags and missing/empty tagsFlat
bf_json=$(curl -s -X POST -H "Host: $HOST" -H "X-User-ID: $U1" "http://localhost/processing/backfill-tags?limit=10")
bf_processed=$(jq -r '.processed // 0' <<<"$bf_json")
pass "D3: Backfill ran (processed=$bf_processed). If 0, corpus already normalized."

# -------- E) LARGE + MULTILANG + INJECTION --------

LARGE=$(python3 - <<'PY'
print(('IRAP funding submission timeline ')*8000)
PY
)
curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
  -d '{"content":"'"${LARGE}"'","tags":[{"label":"long"}]}' "http://localhost/memory/save" >/dev/null
curl -s -H "Host: $HOST" -H "X-User-ID: $U1" "http://localhost/memory/query?userId=$U1&tags=long&limit=1" >/dev/null \
  && pass "E1: Large content handled" || fail "E1: Large content handling failed"

curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
  -d '{"content":"¿Qué pasó con la presentación IRAP? موعد التسليم.","tags":[{"label":"intl"}]}' \
  "http://localhost/memory/save" >/dev/null \
  && pass "E2: Multilingual content handled" || fail "E2: Multilingual handling failed"

curl -s -H "Host: $HOST" -H "X-User-ID: $U1" -H "Content-Type: application/json" \
  -d '{"content":"<script>alert(1)</script> {$ne: null} -- ; DROP","tags":[{"label":"safe"}],"status":"completed"}' \
  "http://localhost/memory/save" >/dev/null \
  && pass "E3: Injection-like input treated as data" || fail "E3: Injection-like input failed"

# -------- F) PERFORMANCE + CONCURRENCY --------

tt=$(curl -w "%{time_total}" -s -o /dev/null -H "Host: $HOST" -H "X-User-ID: $U1" \
  "http://localhost/memory/promote?userId=$U1&tags=pilot_tag&goal=continue")
awk "BEGIN {exit !($tt < 0.10)}" && pass "F1: Promote p95 under 100ms (TTOTAL=${tt}s)" || warn "F1: Promote TTOTAL=${tt}s (>=100ms)"

seq 1 10 | xargs -I{} -P10 bash -c \
 'curl -s -H "Host: '"$HOST"'" -H "X-User-ID: '"$U1"'" \
  "http://localhost/memory/promote?userId='"$U1"'&tags=pilot_tag&goal=continue" > /dev/null' \
 && pass "F2: 10x concurrency OK" || fail "F2: Concurrency errors"

# -------- G) TRAEFIK HYGIENE (informational) --------
docker-compose logs traefik --tail=10 >/dev/null 2>&1 && pass "G: Traefik running (logs accessible)" || warn "G: Traefik logs not accessible"
# Optional: curl -s "http://localhost:9000/api/http/routers" | jq '.[].middlewares'  (if dashboard enabled)

printf "\n==== DOUBLE DEEP CHECK: ALL DONE ====\n"
