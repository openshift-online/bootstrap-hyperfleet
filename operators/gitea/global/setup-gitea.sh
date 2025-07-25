#!/bin/bash
set -e

GITEA_URL="http://localhost:3000"
ADMIN_USER="admin"
ADMIN_PASS="acmeprototype321#"
ADMIN_EMAIL="admin@bootstrap.local"

echo "Setting up Gitea initial configuration..."

# Wait for Gitea to be ready
echo "Waiting for Gitea to be ready..."
until curl -s "${GITEA_URL}/api/healthz" | grep -q "pass"; do
  echo "Waiting for Gitea..."
  sleep 2
done

echo "Gitea is ready!"

# Create admin user using Gitea CLI inside the pod
POD_NAME=$(oc get pods -n gitea-system -l app=gitea -o jsonpath='{.items[0].metadata.name}')

echo "Creating admin user in pod: ${POD_NAME}"
oc exec -n gitea-system ${POD_NAME} -- gitea admin user create \
  --username "${ADMIN_USER}" \
  --password "${ADMIN_PASS}" \
  --email "${ADMIN_EMAIL}" \
  --admin || true

echo "Admin user created (or already exists)"

# Create access token for admin user
echo "Creating access token..."
TOKEN=$(curl -s -X POST "${GITEA_URL}/api/v1/users/${ADMIN_USER}/tokens" \
  -H "Content-Type: application/json" \
  -u "${ADMIN_USER}:${ADMIN_PASS}" \
  -d '{
    "name": "bootstrap-token",
    "scopes": ["write:repository", "write:user"]
  }' | jq -r '.sha1' 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
  echo "Warning: Could not create token via API, will use basic auth"
  TOKEN="${ADMIN_USER}:${ADMIN_PASS}"
fi

echo "Token created: ${TOKEN:0:8}..."

# Create bootstrap organization
echo "Creating bootstrap organization..."
curl -s -X POST "${GITEA_URL}/api/v1/orgs" \
  -H "Content-Type: application/json" \
  -H "Authorization: token ${TOKEN}" \
  -d '{
    "username": "bootstrap",
    "full_name": "Bootstrap Organization",
    "description": "Organization for bootstrap repositories"
  }' || true

# Create bootstrap repository
echo "Creating bootstrap repository..."
curl -s -X POST "${GITEA_URL}/api/v1/orgs/bootstrap/repos" \
  -H "Content-Type: application/json" \
  -H "Authorization: token ${TOKEN}" \
  -d '{
    "name": "bootstrap",
    "description": "Bootstrap cluster configurations",
    "private": false,
    "auto_init": true
  }' || true

echo "Gitea setup complete!"
echo "URL: ${GITEA_URL}"
echo "Admin user: ${ADMIN_USER}"
echo "Admin password: ${ADMIN_PASS}"
echo "Repository: ${GITEA_URL}/bootstrap/bootstrap"