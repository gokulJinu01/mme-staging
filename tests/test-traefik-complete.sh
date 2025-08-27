#!/bin/bash

echo "=== TRAEFIK FORWARDAUTH COMPREHENSIVE TEST ==="

# Test 1: Health check without auth (should work)
echo "Test 1: Health check (no auth required)"
curl -s -i -H "Host: mme.localhost" "http://localhost/health" | head -n1

# Test 2: Memory query without auth (should be 401)
echo "Test 2: Memory query without auth (should be 401)"
curl -s -i -H "Host: mme.localhost" "http://localhost/memory/query?userId=test&limit=1" | head -n1

# Test 3: Memory query with auth (should be 200)
echo "Test 3: Memory query with auth (should be 200)"
curl -s -i -H "Host: mme.localhost" -H "X-User-ID: test_user" "http://localhost/memory/query?userId=test_user&limit=1" | head -n1

# Test 4: Save memory with auth
echo "Test 4: Save memory with auth"
SAVE_RESPONSE=$(curl -s -X POST -H "Host: mme.localhost" -H "X-User-ID: test_user" -H "Content-Type: application/json" \
  -d '{"content":"test content for traefik","tags":[{"label":"traefik_test","type":"concept"}],"status":"completed"}' \
  "http://localhost/memory/save")
echo "Save response: $SAVE_RESPONSE" | jq '{id, tagsFlat}'

# Test 5: Query memory with auth
echo "Test 5: Query memory with auth"
QUERY_RESPONSE=$(curl -s -H "Host: mme.localhost" -H "X-User-ID: test_user" "http://localhost/memory/query?userId=test_user&limit=5")
echo "Query count: $(echo $QUERY_RESPONSE | jq '.count')"

# Test 6: Promote memory with auth (should show spike_trace)
echo "Test 6: Promote memory with auth (should show spike_trace)"
PROMOTE_RESPONSE=$(curl -s -X POST -H "Host: mme.localhost" -H "X-User-ID: test_user" \
  "http://localhost/memory/promote?userId=test_user&tags=traefik_test&goal=continue")
echo "Promote count: $(echo $PROMOTE_RESPONSE | jq '.count')"

# Test 7: Cross-tenant isolation (should be 401)
echo "Test 7: Cross-tenant isolation (should be 401)"
curl -s -i -H "Host: mme.localhost" -H "X-User-ID: user1" "http://localhost/memory/query?userId=user2&limit=1" | head -n1

# Test 8: JWT verifier health check
echo "Test 8: JWT verifier health check"
curl -s -i -H "Host: auth.localhost" "http://localhost/health" | head -n1

# Test 9: Traefik dashboard
echo "Test 9: Traefik dashboard"
curl -s -i "http://localhost:9000" | head -n1

echo "=== TEST COMPLETE ==="
echo "Check logs for spike_trace output:"
echo "docker-compose logs mme-tagging-service --tail=10 | grep spike_trace"
