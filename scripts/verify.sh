#!/bin/bash

BASE_URL="http://localhost:8090/api/v1"

echo "1. Add permissions to Tenant 'tenant-1' (Add)..."
curl -X POST "$BASE_URL/tenant/permissions/add" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_id": "res-1", "action_id": "act-1"}
    ]
  }'
echo -e "\n"

echo "2. Add another permission to Tenant 'tenant-1' (Add)..."
curl -X POST "$BASE_URL/tenant/permissions/add" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_id": "res-1", "action_id": "act-2"}
    ]
  }'
echo -e "\n"

echo "3. Remove permission from Tenant 'tenant-1' (Remove)..."
curl -X POST "$BASE_URL/tenant/permissions/remove" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_id": "res-1", "action_id": "act-1"}
    ]
  }'
echo -e "\n"

echo "4. Sync permissions for Tenant 'tenant-1' (Sync - Replace with act-1, act-2)..."
curl -X PUT "$BASE_URL/tenant/permissions" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_id": "res-1", "action_id": "act-1"},
      {"resource_id": "res-1", "action_id": "act-2"}
    ]
  }'
echo -e "\n"

echo "5. Create Role 'admin' in 'tenant-1'..."
curl -X POST "$BASE_URL/roles" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "admin",
    "tenant_id": "tenant-1"
  }' > role_response.json
ROLE_ID=$(grep -o '"id":"[^"]*"' role_response.json | cut -d'"' -f4)
echo "Role ID: $ROLE_ID"
echo -e "\n"

echo "6. Sync permissions for Role 'admin' (Sync)..."
curl -X PUT "$BASE_URL/roles/$ROLE_ID/permissions" \
  -H "Content-Type: application/json" \
  -d '{
    "permissions": [
      {"resource_id": "res-1", "action_id": "act-1"},
      {"resource_id": "res-1", "action_id": "act-2"}
    ]
  }'
echo -e "\n"

echo "7. Assign User 'user-1' to Role 'admin'..."
curl -X POST "$BASE_URL/roles/$ROLE_ID/users/bulk" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": ["user-1"]
  }'
echo -e "\n"

echo "8. Check Permission: user-1, tenant-1, order:read (Should be TRUE)..."
curl -X POST "$BASE_URL/check-permission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-1",
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_code": "order", "action_code": "read"}
    ]
  }'
echo -e "\n"

echo "9. Check Permission: user-1, tenant-1, product:read (Should be FALSE)..."
curl -X POST "$BASE_URL/check-permission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-1",
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_code": "product", "action_code": "read"}
    ]
  }'
echo -e "\n"

echo "10. Check Permission: user-1, tenant-1, order:write (Should be TRUE)..."
curl -X POST "$BASE_URL/check-permission" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-1",
    "tenant_id": "tenant-1",
    "permissions": [
      {"resource_code": "order", "action_code": "write"}
    ]
  }'
echo -e "\n"
