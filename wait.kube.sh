#!/bin/bash

# ====================================================================================
#
# Description: This script waits for any specified Kubernetes resource to meet
#              a condition defined by a JSONPath expression.
#
# Usage: ./wait.kube.sh <type> <name> <namespace> <jsonpath> <expected-value> [timeout]
#
# Arguments:
#   <type>:           The type of the resource (e.g., pod, deployment, service, crd).
#   <name>:           The name of the resource.
#   <namespace>:      The namespace of the resource. Use "" for cluster-scoped resources.
#   <jsonpath>:       The JSONPath expression to query the resource's status.
#   <expected-value>: The value the JSONPath query should return for the condition to be met.
#   [timeout]:        (Optional) The maximum time to wait in seconds. Defaults to 180.
#
# Examples:
#   # Wait for a Pod named 'my-pod' in the 'default' namespace to be 'Ready'.
#   ./wait-for-resource.sh pod my-pod default '{.status.conditions[?(@.type=="Ready")].status}' "True"
#
#   # Wait for a Deployment to complete its rollout (ready replicas match desired replicas).
#   DESIRED=$(oc get deployment my-app -n web -o jsonpath='{.spec.replicas}')
#   ./wait-for-resource.sh deployment my-app web '{.status.readyReplicas}' "$DESIRED" 300
#
#   # Wait for a CustomResourceDefinition to be 'Established'.
#   ./wait-for-resource.sh crd my-crd.example.com "" '{.status.conditions[?(@.type=="Established")].status}' "True" 120
#
# ====================================================================================

# --- Script Configuration ---
RESOURCE_TYPE="$1"
RESOURCE_NAME="$2"
NAMESPACE="$3"
JSONPATH_CONDITION="$4"
EXPECTED_STATUS="$5"
TIMEOUT=${6:-1800}
SLEEP_INTERVAL=60

# --- Argument Validation ---
if [ -z "$RESOURCE_TYPE" ] || [ -z "$RESOURCE_NAME" ] || [ -z "$JSONPATH_CONDITION" ] || [ -z "$EXPECTED_STATUS" ]; then
  echo "Error: Missing required arguments."
  echo "Usage: $0 <type> <name> <namespace> <jsonpath> <expected-value> [timeout]"
  echo "Note: For cluster-scoped resources, provide an empty string for the namespace: \"\""
  exit 1
fi

# --- Namespace Flag Handling ---
# Construct the namespace flag for oc. If the namespace is empty, no flag is added.
if [ -n "$NAMESPACE" ]; then
  NAMESPACE_FLAG="-n $NAMESPACE"
  namespace_log_name=$NAMESPACE
else
  NAMESPACE_FLAG=""
  namespace_log_name="cluster-scoped"
fi

echo "Waiting for resource '$RESOURCE_TYPE/$RESOURCE_NAME' in namespace '$namespace_log_name'..."
echo "Condition: JSONPath '$JSONPATH_CONDITION' must be '$EXPECTED_STATUS'."

# --- Main Logic ---
start_time=$(date +%s)

# Loop until the resource condition is met or the timeout is reached.
while true; do
  current_time=$(date +%s)
  elapsed_time=$((current_time - start_time))

  if [ "$elapsed_time" -ge "$TIMEOUT" ]; then
    echo "Error: Timeout of ${TIMEOUT}s reached. Resource condition not met."
    exit 1
  fi

  # Query for the resource and check the condition using the provided JSONPath.
  # The 2>/dev/null suppresses errors if the resource doesn't exist yet, which is expected.
  current_status=$(oc get "$RESOURCE_TYPE" "$RESOURCE_NAME" $NAMESPACE_FLAG -o jsonpath="$JSONPATH_CONDITION" 2>/dev/null)

  # Check if the current status matches the expected status.
  if [ "$current_status" == "$EXPECTED_STATUS" ]; then
    echo "Success: Resource '$RESOURCE_TYPE/$RESOURCE_NAME' has met the condition."
    break
  else

    oc get "$RESOURCE_TYPE" "$RESOURCE_NAME" $NAMESPACE_FLAG -o jsonpath="$JSONPATH_CONDITION"

    echo "Resource not ready yet. Current status: '${current_status:-<not-found>}'. Retrying in ${SLEEP_INTERVAL}s... (Elapsed: ${elapsed_time}s)"
    sleep "$SLEEP_INTERVAL"
  fi
done

exit 0
