#!/bin/bash

# Test script for Traefik ForwardAuth setup
echo "=== TRAEFIK FORWARDAUTH TEST SUITE ==="

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Test 1: Health check (should work without auth)
echo "Test 1: Health check (no auth required)"
curl -s -i "http://localhost/health" | head -n1

# Test 2: JWT verifier health check
echo "Test 2: JWT verifier health check"
curl -s -i "http://auth.localhost/health" | head -n1

# Test 3: MME service without auth (should be 401)
echo "Test 3: MME service without auth (should be 401)"
curl -s -i "http://mme.localhost/memory/query?userId=test&limit=1" | head -n1

# Test 4: MME service with X-User-ID header (should be 200)
echo "Test 4: MME service with X-User-ID header (should be 200)"
curl -s -i -H "X-User-ID: test_user" "http://mme.localhost/memory/query?userId=test_user&limit=1" | head -n1

# Test 5: MME service with user mismatch (should be 401)
echo "Test 5: MME service with user mismatch (should be 401)"
curl -s -i -H "X-User-ID: user1" "http://mme.localhost/memory/query?userId=user2&limit=1" | head -n1

# Test 6: Save memory with auth
echo "Test 6: Save memory with auth"
curl -s -X POST -H "X-User-ID: test_user" -H "Content-Type: application/json" \
  -d '{"userId":"test_user","content":"test content","tags":[{"label":"test_tag","type":"concept"}],"status":"completed"}' \
  "http://mme.localhost/memory/save" | jq '{id, tagsFlat}'

# Test 7: Query memory with auth
echo "Test 7: Query memory with auth"
curl -s -H "X-User-ID: test_user" "http://mme.localhost/memory/query?userId=test_user&limit=5" | jq '.count'

# Test 8: Promote memory with auth
echo "Test 8: Promote memory with auth"
curl -s -X POST -H "X-User-ID: test_user" \
  "http://mme.localhost/memory/promote?userId=test_user&tags=test_tag&goal=continue" | jq '.count'

echo "=== TEST COMPLETE ==="
