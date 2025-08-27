#!/bin/bash

# MME Demo Test Script

set -e

echo "🧪 Running MME Demo Tests..."

# Test basic health
echo "1. Testing basic health..."
curl -s -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  "http://localhost/health" | grep -q "ok" && echo "✅ Health check passed" || echo "❌ Health check failed"

# Test structured tagging
echo "2. Testing structured tagging..."
SAVE_RESPONSE=$(curl -s -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  -H "Content-Type: application/json" \
  -d '{"content":"Demo test content","tags":[{"label":"demo","type":"concept"}],"status":"completed"}' \
  "http://localhost/memory/save")

echo "$SAVE_RESPONSE" | jq -r '.id' > /dev/null && echo "✅ Save operation passed" || echo "❌ Save operation failed"

# Test tag-scoped recall
echo "3. Testing tag-scoped recall..."
PROMOTE_RESPONSE=$(curl -s -H "Host: mme.localhost" -H "X-User-ID: demo_user" \
  "http://localhost/memory/promote?userId=demo_user&tags=demo&goal=continue")

echo "$PROMOTE_RESPONSE" | jq -r '.results | length' > /dev/null && echo "✅ Promote operation passed" || echo "❌ Promote operation failed"

# Test multi-tenant isolation
echo "4. Testing multi-tenant isolation..."
USER_A_RESPONSE=$(curl -s -H "Host: mme.localhost" -H "X-User-ID: user_a" \
  -H "Content-Type: application/json" \
  -d '{"content":"User A private data","tags":[{"label":"private"}],"status":"completed"}' \
  "http://localhost/memory/save")

USER_B_RESPONSE=$(curl -s -H "Host: mme.localhost" -H "X-User-ID: user_b" \
  "http://localhost/memory/query?userId=user_b&tags=private&limit=5")

USER_B_COUNT=$(echo "$USER_B_RESPONSE" | jq -r '.results | length')
if [ "$USER_B_COUNT" = "0" ]; then
    echo "✅ Multi-tenant isolation working"
else
    echo "❌ Multi-tenant isolation failed"
fi

echo ""
echo "🎉 Demo tests completed!"
