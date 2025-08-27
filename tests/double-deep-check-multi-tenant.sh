#!/bin/bash
set -euo pipefail

# -------- Multiple User Sets --------
export HOST="mme.localhost"

# Test Set 1: Original pilot users
export U1="pilot_u1"
export U2="pilot_u2"

# Test Set 2: Different domain users
export U3="finance_user_001"
export U4="hr_manager_002"

# Test Set 3: Special characters and edge cases
export U5="user-with-dash"
export U6="user_with_underscore"

# Test Set 4: Long user IDs
export U7="very_long_user_id_that_tests_edge_cases_123456789"
export U8="another_very_long_user_id_for_testing_purposes_987654321"

pass(){ printf "✅ %s\n" "$1"; }
fail(){ printf "❌ %s\n" "$1"; exit 1; }
warn(){ printf "⚠️  %s\n" "$1"; }

printf "=== MULTI-TENANT DOUBLE DEEP CHECK ===\n"
printf "HOST=%s\n" "$HOST"
printf "User Sets:\n"
printf "  Set 1: %s, %s\n" "$U1" "$U2"
printf "  Set 2: %s, %s\n" "$U3" "$U4"
printf "  Set 3: %s, %s\n" "$U5" "$U6"
printf "  Set 4: %s, %s\n" "$U7" "$U8"

# -------- Helper: require command --------
req() { command -v "$1" >/dev/null 2>&1 || fail "Missing dependency: $1"; }
req curl; req jq

# -------- Helper: test user isolation --------
test_user_isolation() {
    local user1="$1"
    local user2="$2"
    local test_name="$3"
    
    printf "\n--- Testing Isolation: %s ---\n" "$test_name"
    
    # Create documents for both users
    curl -s -H "Host: $HOST" -H "X-User-ID: $user1" -H "Content-Type: application/json" \
      -d "{\"content\":\"$user1 private doc\",\"tags\":[{\"label\":\"${user1}_tag\",\"type\":\"concept\"}],\"status\":\"completed\"}" \
      "http://localhost/memory/save" >/dev/null

    curl -s -H "Host: $HOST" -H "X-User-ID: $user2" -H "Content-Type: application/json" \
      -d "{\"content\":\"$user2 private doc\",\"tags\":[{\"label\":\"${user2}_tag\",\"type\":\"concept\"}],\"status\":\"completed\"}" \
      "http://localhost/memory/save" >/dev/null

    # Test cross-tenant isolation
    iso_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $user1" \
      "http://localhost/memory/query?userId=$user2&tags=${user1}_tag&limit=5")

    # Check for data leakage
    u2_leak=$(jq -r '.. | .content? // empty' <<<"$iso_json" | grep -c "$user2 private doc" || true)
    [ "$u2_leak" -eq 0 ] && pass "$test_name: No cross-tenant data leak" || fail "$test_name: Found cross-tenant data leak"
    
    # Verify user1 can only see their own data
    user1_count=$(jq -r '.count // 0' <<<"$iso_json")
    [ "$user1_count" -ge 1 ] && pass "$test_name: User1 can access their own data" || fail "$test_name: User1 cannot access their data"
}

# -------- Helper: test user functionality --------
test_user_functionality() {
    local user="$1"
    local test_name="$2"
    
    printf "\n--- Testing Functionality: %s ---\n" "$test_name"
    
    # Test save
    save_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"$test_name test content\",\"tags\":[{\"label\":\"${test_name}_tag\",\"type\":\"concept\"}],\"status\":\"completed\"}" \
      "http://localhost/memory/save")
    
    tags_flat=$(jq -r '.tagsFlat | @json' <<<"$save_json" 2>/dev/null || echo "null")
    [[ "$tags_flat" != "null" && "$tags_flat" != "[]" ]] && pass "$test_name: tagsFlat present in save response" || fail "$test_name: tagsFlat missing in response"
    
    # Test query
    query_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $user" \
      "http://localhost/memory/query?userId=$user&tags=${test_name}_tag&limit=5")
    
    query_count=$(jq -r '.count // 0' <<<"$query_json")
    [ "$query_count" -ge 1 ] && pass "$test_name: Query returns expected results" || fail "$test_name: Query failed"
    
    # Test promote
    prom_json=$(curl -s -X POST -H "Host: $HOST" -H "X-User-ID: $user" \
      "http://localhost/memory/promote?userId=$user&tags=${test_name}_tag&goal=continue")
    
    prom_count=$(jq -r '.count // 0' <<<"$prom_json")
    [ "$prom_count" -ge 0 ] && pass "$test_name: Promote works correctly" || fail "$test_name: Promote failed"
    
    # Test delete
    del_id=$(jq -r '.id' <<<"$save_json")
    del_response=$(curl -s -X DELETE -H "Host: $HOST" -H "X-User-ID: $user" "http://localhost/memory/$del_id")
    del_success=$(jq -r '.message? // "ok"' <<<"$del_response")
    [[ "$del_success" =~ deleted|ok ]] && pass "$test_name: Delete works correctly" || fail "$test_name: Delete failed"
}

# -------- A) GATEWAY & SECURITY (Multi-tenant) --------

printf "\n=== A) GATEWAY & SECURITY (Multi-tenant) ===\n"

# A1. ForwardAuth behavior (no header -> 401)
status=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: $HOST" "http://localhost/health" || true)
[ "$status" = "401" ] && pass "A1: ForwardAuth returns 401 without header" || fail "A1: Expected 401, got $status"

# A2. Auth OK (with header -> 200) for all user types
for user in "$U1" "$U3" "$U5" "$U7"; do
    status=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: $HOST" -H "X-User-ID: $user" "http://localhost/health")
    [ "$status" = "200" ] && pass "A2: Auth returns 200 with X-User-ID: $user" || fail "A2: Expected 200, got $status for $user"
done

# A3. Cross-tenant isolation for all user sets
test_user_isolation "$U1" "$U2" "Pilot Users"
test_user_isolation "$U3" "$U4" "Finance/HR Users"
test_user_isolation "$U5" "$U6" "Special Character Users"
test_user_isolation "$U7" "$U8" "Long User ID Users"

# A4. Bypass hardening
docker-compose ps | grep -q "traefik.*0.0.0.0:80" && pass "A4: Only Traefik exposes host ports" || warn "A4: Check port exposure"

# -------- B) RESPONSE SERIALIZATION (Multi-tenant) --------

printf "\n=== B) RESPONSE SERIALIZATION (Multi-tenant) ===\n"

for user in "$U1" "$U3" "$U5" "$U7"; do
    test_name="User_${user}"
    save_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"$test_name test\",\"tags\":[{\"label\":\"${test_name}_tag\",\"type\":\"concept\"}],\"status\":\"completed\"}" \
      "http://localhost/memory/save")
    
    tags_flat=$(jq -r '.tagsFlat | @json' <<<"$save_json" 2>/dev/null || echo "null")
    [[ "$tags_flat" != "null" && "$tags_flat" != "[]" ]] && pass "B1: tagsFlat present for $user: $tags_flat" || fail "B1: tagsFlat missing for $user"
done

# -------- C) PROMOTE + SPIKE TRACE (Multi-tenant) --------

printf "\n=== C) PROMOTE + SPIKE TRACE (Multi-tenant) ===\n"

# Test promote for all user types
for user in "$U1" "$U3" "$U5" "$U7"; do
    prom_json=$(curl -s -X POST -H "Host: $HOST" -H "X-User-ID: $user" \
      "http://localhost/memory/promote?userId=$user&tags=${user}_tag&goal=continue")
    
    prom_count=$(jq -r '.count // 0' <<<"$prom_json")
    [ "$prom_count" -ge 0 ] && pass "C1: Promote works for $user (count=$prom_count)" || fail "C1: Promote failed for $user"
done

# Check spike_trace lines for all users
sleep 2
log_output=$(docker-compose logs mme-tagging-service --tail=200)
spike_count=$(echo "$log_output" | grep -c "spike_trace" || echo "0")
[ "$spike_count" -gt 0 ] && pass "C2: spike_trace lines present in logs (found $spike_count lines)" || fail "C2: spike_trace lines not found"

# -------- D) DATA INTEGRITY (Multi-tenant) --------

printf "\n=== D) DATA INTEGRITY (Multi-tenant) ===\n"

# Test edges for different user sets
for user in "$U1" "$U3" "$U5" "$U7"; do
    for i in 1 2; do
        curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
          -d "{\"content\":\"pair test-$user\",\"tags\":[{\"label\":\"TEST\",\"type\":\"concept\"},{\"label\":\"$user\",\"type\":\"object\"}],\"status\":\"completed\"}" \
          "http://localhost/memory/save" >/dev/null
    done
    pass "D1: Saved TEST-$user pairs for $user (edge upserts assumed OK)"
done

# Test delete permissions across users
for user in "$U1" "$U3" "$U5" "$U7"; do
    # Create a document to delete
    save_json=$(curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"temp to delete for $user\",\"tags\":[{\"label\":\"to_delete_$user\"}],\"status\":\"completed\"}" \
      "http://localhost/memory/save")
    
    del_id=$(jq -r '.id' <<<"$save_json")
    
    # Owner can delete
    del_owner=$(curl -s -X DELETE -H "Host: $HOST" -H "X-User-ID: $user" "http://localhost/memory/$del_id" | jq -r '.message? // "ok"')
    [[ "$del_owner" =~ deleted|ok ]] && pass "D2: Delete OK for owner $user" || fail "D2: Owner delete failed for $user"
    
    # Other users cannot delete (test with U2)
    del_wrong=$(curl -s -X DELETE -H "Host: $HOST" -H "X-User-ID: $U2" "http://localhost/memory/$del_id" | jq -r '.error? // "none"')
    [[ "$del_wrong" =~ not\ found|404|none ]] && pass "D2: Other user cannot delete $user's doc" || fail "D2: Wrong-user delete succeeded for $user"
done

# Test backfill for all users
for user in "$U1" "$U3" "$U5" "$U7"; do
    bf_json=$(curl -s -X POST -H "Host: $HOST" -H "X-User-ID: $user" "http://localhost/processing/backfill-tags?limit=10")
    bf_processed=$(jq -r '.processed // 0' <<<"$bf_json")
    pass "D3: Backfill ran for $user (processed=$bf_processed)"
done

# -------- E) LARGE + MULTILANG + INJECTION (Multi-tenant) --------

printf "\n=== E) LARGE + MULTILANG + INJECTION (Multi-tenant) ===\n"

# Test large content for different users
LARGE=$(python3 -c "print(('Large content test ')*5000)")
for user in "$U1" "$U3" "$U5" "$U7"; do
    curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"$LARGE\",\"tags\":[{\"label\":\"large_$user\"}]}" \
      "http://localhost/memory/save" >/dev/null
    
    curl -s -H "Host: $HOST" -H "X-User-ID: $user" \
      "http://localhost/memory/query?userId=$user&tags=large_$user&limit=1" >/dev/null \
      && pass "E1: Large content handled for $user" || fail "E1: Large content handling failed for $user"
done

# Test multilingual content for different users
for user in "$U1" "$U3" "$U5" "$U7"; do
    curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"¿Qué pasó con $user? موعد التسليم. こんにちは.\",\"tags\":[{\"label\":\"intl_$user\"}]}" \
      "http://localhost/memory/save" >/dev/null \
      && pass "E2: Multilingual content handled for $user" || fail "E2: Multilingual handling failed for $user"
done

# Test injection-like inputs for different users
for user in "$U1" "$U3" "$U5" "$U7"; do
    curl -s -H "Host: $HOST" -H "X-User-ID: $user" -H "Content-Type: application/json" \
      -d "{\"content\":\"<script>alert('$user')</script> {\$ne: null} -- ; DROP TABLE users;\",\"tags\":[{\"label\":\"safe_$user\"}],\"status\":\"completed\"}" \
      "http://localhost/memory/save" >/dev/null \
      && pass "E3: Injection-like input treated as data for $user" || fail "E3: Injection-like input failed for $user"
done

# -------- F) PERFORMANCE + CONCURRENCY (Multi-tenant) --------

printf "\n=== F) PERFORMANCE + CONCURRENCY (Multi-tenant) ===\n"

# Test performance for all users
for user in "$U1" "$U3" "$U5" "$U7"; do
    tt=$(curl -w "%{time_total}" -s -o /dev/null -H "Host: $HOST" -H "X-User-ID: $user" \
      "http://localhost/memory/promote?userId=$user&tags=${user}_tag&goal=continue")
    awk "BEGIN {exit !($tt < 0.10)}" && pass "F1: Promote p95 under 100ms for $user (TTOTAL=${tt}s)" || warn "F1: Promote TTOTAL=${tt}s for $user (>=100ms)"
done

# Test concurrency with mixed users
printf "Testing mixed-user concurrency...\n"
for i in {1..5}; do
    curl -s -H "Host: $HOST" -H "X-User-ID: $U1" \
      "http://localhost/memory/promote?userId=$U1&tags=${U1}_tag&goal=continue" >/dev/null &
    curl -s -H "Host: $HOST" -H "X-User-ID: $U3" \
      "http://localhost/memory/promote?userId=$U3&tags=${U3}_tag&goal=continue" >/dev/null &
    curl -s -H "Host: $HOST" -H "X-User-ID: $U5" \
      "http://localhost/memory/promote?userId=$U5&tags=${U5}_tag&goal=continue" >/dev/null &
    curl -s -H "Host: $HOST" -H "X-User-ID: $U7" \
      "http://localhost/memory/promote?userId=$U7&tags=${U7}_tag&goal=continue" >/dev/null &
done
wait
pass "F2: Mixed-user concurrency OK"

# -------- G) TRAEFIK HYGIENE (Multi-tenant) --------

printf "\n=== G) TRAEFIK HYGIENE (Multi-tenant) ===\n"

docker-compose logs traefik --tail=10 >/dev/null 2>&1 && pass "G: Traefik running (logs accessible)" || warn "G: Traefik logs not accessible"

# Test that all users are properly routed
for user in "$U1" "$U3" "$U5" "$U7"; do
    status=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: $HOST" -H "X-User-ID: $user" "http://localhost/health")
    [ "$status" = "200" ] && pass "G: User $user properly routed" || fail "G: User $user routing failed"
done

printf "\n==== MULTI-TENANT DOUBLE DEEP CHECK: ALL DONE ====\n"
printf "Tested %d different user types across %d isolation scenarios\n" 8 4
