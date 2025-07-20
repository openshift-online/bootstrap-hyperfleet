#!/bin/bash

echo "🔐 Setting up Vault-based secret management..."

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to set up a namespace and apply Vault-managed secrets
# @param {string} cluster_name - The semantic name of the cluster (e.g., ocp-01, eks-02, hcp-01-test)
setup_cluster_vault_secrets() {
  local cluster_name="$1"
  local namespace="${cluster_name}"

  echo "--- 🚀 Starting Vault integration for namespace: ${namespace} ---"

  # 1. Create the namespace, ignoring errors if it already exists
  oc create namespace "${namespace}" 2>/dev/null || true
  sleep .2

  # 2. Create service account for Vault access
  oc create serviceaccount vault-secret-reader -n "${namespace}" 2>/dev/null || true
  sleep .2

  # 3. Create service account token secret (required for Vault auth)
  cat <<EOF | oc apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: vault-secret-reader-token
  namespace: ${namespace}
  annotations:
    kubernetes.io/service-account.name: vault-secret-reader
type: kubernetes.io/service-account-token
EOF
  sleep .2

  # 4. Apply ExternalSecrets for this namespace
  echo "Applying ExternalSecrets for ${namespace}..."
  sed "s/CLUSTER_NAMESPACE/${namespace}/g" bases/ocm/external-secrets-template.yaml | oc apply -f -
  sleep .2

  # 5. ExternalSecrets created - secrets will sync when ESO is ready and cluster namespace exists
  echo "✅ ExternalSecrets configured for ${namespace}"
  echo "    Secrets will be synced automatically by ESO when:"
  echo "    - External Secret Operator is running on hub cluster"
  echo "    - ClusterSecretStore vault-cluster-store is configured" 
  echo "    - Cluster namespace ${namespace} exists (after cluster provisioning)"
  echo ""
  echo "    Monitor secret sync status: oc get externalsecret -n ${namespace}"
  echo "    Debug if needed: oc describe externalsecret aws-credentials -n ${namespace}"

  echo "--- ✅ Vault setup complete for namespace: ${namespace} ---"
  echo ""
}

# === MAIN EXECUTION ===

echo ""
echo -e "${YELLOW}🔐 Setting up Vault-based cluster secrets...${NC}"

# Verify Vault is running
if ! oc get pod vault-0 -n vault >/dev/null 2>&1; then
  echo -e "${RED}❌ Vault pod not found. Please ensure Vault is deployed first.${NC}"
  exit 1
fi

# Verify ClusterSecretStore exists
if ! oc get clustersecretstore vault-cluster-store >/dev/null 2>&1; then
  echo -e "${RED}❌ ClusterSecretStore 'vault-cluster-store' not found.${NC}"
  echo "Please run the Vault setup first:"
  echo "  oc apply -f operators/vault/global/cluster-secret-store.yaml"
  exit 1
fi

# Check if secrets are stored in Vault
echo "Verifying secrets are stored in Vault..."
if ! oc exec vault-0 -n vault -- vault kv get secret/aws-credentials >/dev/null 2>&1; then
  echo -e "${RED}❌ AWS credentials not found in Vault at secret/aws-credentials${NC}"
  echo "Please store secrets in Vault first. See documentation for instructions."
  exit 1
fi

if ! oc exec vault-0 -n vault -- vault kv get secret/pull-secret >/dev/null 2>&1; then
  echo -e "${RED}❌ Pull secret not found in Vault at secret/pull-secret${NC}"
  echo "Please store secrets in Vault first. See documentation for instructions."
  exit 1
fi

echo -e "${GREEN}✅ Vault prerequisites verified${NC}"

# Dynamically discover all cluster configurations and set up Vault-managed secrets
echo "🔍 Discovering cluster configurations..."

# Find all cluster directories and extract cluster names
cluster_names=()
if [ -d "clusters" ]; then
  for cluster_dir in clusters/*/; do
    if [ -d "$cluster_dir" ]; then
      cluster_name=$(basename "$cluster_dir")
      cluster_names+=("$cluster_name")
      echo "  Found cluster: $cluster_name"
    fi
  done
fi

if [ ${#cluster_names[@]} -eq 0 ]; then
  echo -e "${YELLOW}⚠️  No cluster configurations found in clusters/ directory${NC}"
  echo "This is expected if no clusters have been configured yet."
  exit 0
fi

echo "📋 Setting up Vault secrets for ${#cluster_names[@]} cluster(s)..."
echo ""

# Set up Vault-managed secrets for all discovered clusters
for cluster_name in "${cluster_names[@]}"; do
  setup_cluster_vault_secrets "$cluster_name"
done

echo -e "${GREEN}🎉 All Vault-based cluster secrets setup complete!${NC}"
echo ""
echo "📋 Summary:"
echo "  • Vault ClusterSecretStore: vault-cluster-store"
echo "  • Secret refresh interval: 1 hour"
echo "  • Managed secrets: aws-credentials, pull-secret"
echo ""
echo "🔍 Monitoring:"
echo "  • Check ExternalSecret status: oc get externalsecret -A"
echo "  • Check secret sync status: oc describe externalsecret <name> -n <namespace>"
echo "  • View Vault secrets: oc exec vault-0 -n vault -- vault kv list secret/"
echo ""
echo "🔧 Troubleshooting:"
echo "  • If secrets fail to sync, check Vault auth: oc logs -n external-secrets deployment/eso-external-secrets"
echo "  • Vault UI access: oc port-forward -n vault vault-0 8200:8200 (token: root)"