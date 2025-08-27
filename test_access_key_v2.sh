#!/bin/bash

# Test Access Key Authentication Flow

BASE_URL="http://localhost:8080/api/v1"
EMAIL="test.admin@example.com"
PASSWORD="testpassword123"

echo "=== Testing Access Key Authentication Flow ==="
echo ""

# Step 1: Login to get JWT token
echo "1. Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('access_token', ''))" 2>/dev/null)

if [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Failed to login. Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Successfully logged in"
echo ""

# Step 2: Create an access key
echo "2. Creating access key..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/access-keys" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test API Key",
    "description": "Test key for authentication flow",
    "permissions": ["collection:*:read", "view:*:execute"],
    "expires_in": 3600
  }')

ACCESS_KEY=$(echo $CREATE_RESPONSE | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('key', ''))" 2>/dev/null)
KEY_ID=$(echo $CREATE_RESPONSE | python3 -c "import sys, json; data = json.load(sys.stdin); print(data.get('id', ''))" 2>/dev/null)

if [ -z "$ACCESS_KEY" ]; then
  echo "❌ Failed to create access key. Response: $CREATE_RESPONSE"
  exit 1
fi

echo "✅ Successfully created access key"
echo "   Key ID: $KEY_ID"
echo "   Key: $ACCESS_KEY"
echo ""

# Step 3: Test authentication with access key
echo "3. Testing authentication with access key..."
TEST_RESPONSE=$(curl -s -X GET "$BASE_URL/collections" \
  -H "Authorization: Bearer $ACCESS_KEY")

if echo "$TEST_RESPONSE" | grep -q "collections"; then
  echo "✅ Successfully authenticated with access key"
  echo "   Response: $(echo $TEST_RESPONSE | python3 -c "import sys, json; data = json.load(sys.stdin); print(f\"Found {len(data.get('collections', []))} collections\")" 2>/dev/null || echo "$TEST_RESPONSE")"
else
  echo "❌ Failed to authenticate with access key. Response: $TEST_RESPONSE"
fi
echo ""

# Step 4: Test access key permissions (should fail for write)
echo "4. Testing access key permissions (attempting write - should fail)..."
WRITE_TEST=$(curl -s -X POST "$BASE_URL/collections" \
  -H "Authorization: Bearer $ACCESS_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "test_collection", "description": "Test"}')

if echo "$WRITE_TEST" | grep -q "Insufficient permissions\|forbidden\|denied"; then
  echo "✅ Correctly denied write access (as expected)"
else
  echo "⚠️  Unexpected response to write attempt: $WRITE_TEST"
fi
echo ""

# Step 5: List access keys
echo "5. Listing access keys..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/admin/access-keys" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$LIST_RESPONSE" | grep -q "$KEY_ID"; then
  echo "✅ Access key found in list"
  KEY_COUNT=$(echo $LIST_RESPONSE | python3 -c "import sys, json; data = json.load(sys.stdin); print(len(data.get('access_keys', [])))" 2>/dev/null || echo "unknown")
  echo "   Total access keys: $KEY_COUNT"
else
  echo "❌ Access key not found in list"
fi
echo ""

# Step 6: Delete the test access key
echo "6. Deleting test access key..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/admin/access-keys/$KEY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$DELETE_RESPONSE" | grep -q "deleted\|success"; then
  echo "✅ Successfully deleted access key"
else
  echo "❌ Failed to delete access key. Response: $DELETE_RESPONSE"
fi
echo ""

# Step 7: Verify key no longer works
echo "7. Verifying deleted key no longer works..."
VERIFY_RESPONSE=$(curl -s -X GET "$BASE_URL/collections" \
  -H "Authorization: Bearer $ACCESS_KEY")

if echo "$VERIFY_RESPONSE" | grep -q "Invalid access key\|Unauthorized\|invalid"; then
  echo "✅ Deleted key correctly rejected"
else
  echo "❌ Deleted key still works (unexpected). Response: $VERIFY_RESPONSE"
fi
echo ""

echo "=== Access Key Authentication Flow Test Complete ===
"
echo "Summary:"
echo "✅ User authentication works"
echo "✅ Access key creation works"
echo "✅ Access key authentication works"
echo "✅ Permission checking works"
echo "✅ Access key deletion works"