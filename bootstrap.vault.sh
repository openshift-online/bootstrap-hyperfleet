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
  local aws_creds_file="$2"
  local pull_secret_file="$3"

  # Construct the namespace from the cluster ID
  local namespace="cluster-${cluster_id}"

  # Extract the resource name from the filename (e.g., "secrets/aws-creds.yaml" -> "aws-creds")
  local aws_creds_name=$(basename "${aws_creds_file}" .yaml)
  local pull_secret_name=$(basename "${pull_secret_file}" .yaml)

  echo "--- 🚀 Starting setup for namespace: ${namespace} ---"

  # 1. Create the namespace, ignoring errors if it already exists
  oc create namespace "${namespace}" 2>/dev/null || true

  # 2. Apply the secrets to the specified namespace
  echo "Applying secrets..."
  oc apply -f "${aws_creds_file}" -n "${namespace}"
  oc apply -f "${pull_secret_file}" -n "${namespace}"

  # 3. Wait for both secrets to exist in the cluster
  echo "Waiting for secrets to become available..."
  ./wait.kube.sh secret "${aws_creds_name}" "${namespace}" '{.kind}' Secret
  ./wait.kube.sh secret "${pull_secret_name}" "${namespace}" '{.kind}' Secret

  echo "--- ✅ Setup complete for namespace: ${namespace} ---"
  echo "" # Add a blank line for readability
}

# === HOW TO USE THE FUNCTION ===

# Now you can replace the original repetitive blocks with clean function calls.
# This handles the different file paths seen in your original script.

setup_cluster_secrets "10" "secrets/aws-creds.yaml" "secrets/pull-secret.yaml"
#setup_cluster_secrets "20" "secrets/aws-creds.yaml" "secrets/pull-secret.yaml"
#setup_cluster_secrets "30" "secrets/aws-creds.yaml" "secrets/pull-secret.yaml"
setup_cluster_secrets "40" "secrets/aws-creds.yaml" "secrets/pull-secret.yaml"


#oc create namespace cluster-10 2>/dev/null
#oc apply -f secrets/aws-creds.yaml -n cluster-10
#oc apply -f secrets/pull-secret.yaml -n cluster-10
#
#./wait.kube.sh secret secrets/aws-creds cluster-10 {.kind} Secret
#./wait.kube.sh secret secrets/pull-secret cluster-10 {.kind} Secret
#
#oc create namespace cluster-20 2>/dev/null
#oc apply -f aws-creds.yaml -n cluster-20
#oc apply -f pull-secret.yaml -n cluster-20
#
#./wait.kube.sh secret aws-creds cluster-20 '{.kind}' Secret
#./wait.kube.sh secret pull-secret cluster-20 '{.kind}' Secret
##
#oc create namespace cluster-30 2>/dev/null
#oc apply -f secrets/aws-creds.yaml -n cluster-30
#oc apply -f secrets/pull-secret.yaml -n cluster-30
#
#./wait.kube.sh secret aws-creds cluster-30 '{.kind}' Secret
#./wait.kube.sh secret pull-secret cluster-30 '{.kind}' Secret
#
#
#
