#!/bin/bash

# Test Access Key Authentication Flow

BASE_URL="http://localhost:8080/api/v1"
EMAIL="admin@example.com"
PASSWORD="admin123"

echo "=== Testing Access Key Authentication Flow ==="
echo ""

# Step 1: Login to get JWT token
echo "1. Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

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

ACCESS_KEY=$(echo $CREATE_RESPONSE | grep -o '"key":"[^"]*' | cut -d'"' -f4)
KEY_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*' | cut -d'"' -f4)

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
else
  echo "❌ Failed to authenticate with access key. Response: $TEST_RESPONSE"
fi
echo ""

# Step 4: List access keys
echo "4. Listing access keys..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/admin/access-keys" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$LIST_RESPONSE" | grep -q "$KEY_ID"; then
  echo "✅ Access key found in list"
else
  echo "❌ Access key not found in list"
fi
echo ""

# Step 5: Delete the test access key
echo "5. Deleting test access key..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/admin/access-keys/$KEY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
  echo "✅ Successfully deleted access key"
else
  echo "❌ Failed to delete access key. Response: $DELETE_RESPONSE"
fi
echo ""

echo "=== Access Key Authentication Flow Test Complete ==="