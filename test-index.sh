#!/bin/bash

# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' | jq -r '.token')

echo "Token: $TOKEN"

# Get indexes for a collection
echo "Getting indexes for sampledata collection:"
curl -s -X GET http://localhost:8080/api/v1/collections/sampledata/indexes \
  -H "Authorization: Bearer $TOKEN" | jq '.'