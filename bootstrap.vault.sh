#!/bin/bash

echo "Temporary workaround: Creating cluster namespaces so we can hack create necessary secrets"#

# A script to wait for a Kubernetes resource to be ready.
# Ensure wait.kube.sh is in your PATH or in the current directory.

# ---
# Function to set up a namespace and apply standard secrets.
#
# @param {string} cluster_id       - The ID of the cluster (e.g., 10, 20).
# @param {string} aws_creds_file   - The file path to the AWS credentials secret.
# @param {string} pull_secret_file - The file path to the pull secret.
# ---
setup_cluster_secrets() {
  # Assign arguments to named local variables for clarity
  local cluster_id="$1"

  # Construct the namespace from the cluster ID
  local namespace="cluster-${cluster_id}"

  echo "--- 🚀 Starting setup for namespace: ${namespace} ---"

  # 1. Create the namespace, ignoring errors if it already exists
  oc create namespace "${namespace}" 2>/dev/null || true
  sleep .2

  # 2. Apply the secrets to the specified namespace
  echo "Applying secrets..."
  oc apply -f "$2" -n "${namespace}"
  sleep .2
  oc apply -f "$3" -n "${namespace}"
  sleep .2

  # 3. Wait for both secrets to exist in the cluster
  echo "Waiting for secrets to become available..."
  ./wait.kube.sh secret "aws-credentials" "${namespace}" '{.kind}' Secret
  sleep .2
  ./wait.kube.sh secret "pull-secret" "${namespace}" '{.kind}' Secret
  sleep .2

  echo "--- ✅ Setup complete for namespace: ${namespace} ---"
  echo "" # Add a blank line for readability
}

# === MAIN EXECUTION ===

echo ""
echo -e "${YELLOW}🚀 Setting up cluster secrets...${NC}"

# Now you can replace the original repetitive blocks with clean function calls.
# This handles the different file paths seen in your original script.

setup_cluster_secrets "10" "secrets/aws-credentials.yaml" "secrets/pull-secret.yaml"
setup_cluster_secrets "20" "secrets/aws-credentials.yaml" "secrets/pull-secret.yaml"
setup_cluster_secrets "30" "secrets/aws-credentials.yaml" "secrets/pull-secret.yaml"
setup_cluster_secrets "40" "secrets/aws-credentials.yaml" "secrets/pull-secret.yaml"
setup_cluster_secrets "50" "secrets/aws-credentials.yaml" "secrets/pull-secret.yaml"

echo -e "${GREEN}🎉 All cluster secrets setup complete!${NC}"
