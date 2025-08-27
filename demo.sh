#!/bin/bash

# AnyBase Demo Script - Collections, Views, and Governance
echo "🚀 AnyBase Demo - Firebase-like Features with Governance"
echo "========================================================="

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"

# Step 1: Login to get token
echo -e "\n1️⃣ Logging in as admin user..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"AdminPass123!"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
echo "✅ Logged in successfully!"

# Step 2: Create a collection with schema and permissions
echo -e "\n2️⃣ Creating 'products' collection with schema and permissions..."
curl -s -X POST "$API_URL/collections" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "products",
    "description": "Product catalog",
    "schema": {
      "fields": [
        {
          "name": "name",
          "type": "string",
          "required": true,
          "indexed": true
        },
        {
          "name": "price",
          "type": "number",
          "required": true
        },
        {
          "name": "category",
          "type": "string",
          "indexed": true
        },
        {
          "name": "stock",
          "type": "number",
          "permissions": {
            "read": ["admin", "manager"],
            "write": ["admin"]
          }
        }
      ],
      "required": ["name", "price"]
    },
    "permissions": {
      "read": {
        "public": true,
        "roles": []
      },
      "write": {
        "public": false,
        "roles": ["admin", "manager"]
      },
      "update": {
        "public": false,
        "roles": ["admin", "manager"]
      },
      "delete": {
        "public": false,
        "roles": ["admin"]
      }
    },
    "settings": {
      "versioning": true,
      "soft_delete": true,
      "auditing": true
    }
  }' | jq '.'

echo "✅ Collection created!"

# Step 3: Insert documents into the collection
echo -e "\n3️⃣ Inserting products into collection..."
curl -s -X POST "$API_URL/data/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop Pro",
    "price": 1299.99,
    "category": "Electronics",
    "stock": 50
  }' | jq '.'

curl -s -X POST "$API_URL/data/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Mouse",
    "price": 29.99,
    "category": "Accessories",
    "stock": 200
  }' | jq '.'

echo "✅ Products inserted!"

# Step 4: Create a view for budget products
echo -e "\n4️⃣ Creating a view for budget products (price < 100)..."
curl -s -X POST "$API_URL/views" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "budget_products",
    "description": "Products under $100",
    "collection": "products",
    "filter": {
      "price": {"$lt": 100}
    },
    "fields": ["name", "price", "category"],
    "sort": {"price": 1},
    "permissions": {
      "public": true,
      "roles": []
    }
  }' | jq '.'

echo "✅ View created!"

# Step 5: Query the view
echo -e "\n5️⃣ Querying budget products view..."
curl -s -X GET "$API_URL/views/budget_products/query" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Step 6: List available roles
echo -e "\n6️⃣ Listing available roles..."
curl -s -X GET "$API_URL/roles" \
  -H "Authorization: Bearer $TOKEN" | jq '.roles[] | {name: .name, description: .description}'

# Step 7: Show collections
echo -e "\n7️⃣ Listing user's accessible collections..."
curl -s -X GET "$API_URL/collections" \
  -H "Authorization: Bearer $TOKEN" | jq '.collections[] | {name: .name, description: .description}'

echo -e "\n✅ Demo completed!"
echo -e "\n📚 Available Features:"
echo "  • Collections with schemas and validation"
echo "  • Field-level permissions (e.g., stock field restricted)"
echo "  • Views for filtered/transformed data access"
echo "  • Role-based access control (RBAC)"
echo "  • Audit logging for compliance"
echo "  • Soft deletes and versioning"
echo -e "\n🔗 Explore more at http://localhost:8080/docs"