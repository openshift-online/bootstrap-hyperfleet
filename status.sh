#!/bin/bash
# ==============================================================================
#
# Description: This script waits for a specific Kubernetes
#              CustomResourceDefinition (CRD) to be established.
#
# Usage: ./wait-for-crd.sh <crd-name> [timeout-in-seconds]
#
# Example: ./wait-for-crd.sh certificates.cert-manager.io 120
#
# ==============================================================================

# --- Script Configuration ---

# The name of the CustomResourceDefinition to wait for.
# This is taken from the first command-line argument.
CRD_NAME="$1"

# The maximum time to wait in seconds.
# Defaults to 120 seconds (2 minutes) if not provided as the second argument.
TIMEOUT=${2:-120}

# The interval in seconds between checks.
SLEEP_INTERVAL=5

# --- Main Logic ---

# Check if a CRD name was provided.
if [ -z "$CRD_NAME" ]; then
  echo "Error: No CRD name provided."
  echo "Usage: $0 <crd-name> [timeout-in-seconds]"
  exit 1
fi

echo "Waiting for CRD '$CRD_NAME' to be established..."

# Record the start time.
start_time=$(date +%s)

# Loop until the CRD is found or the timeout is reached.
while true; do
  # Check the current time.
  current_time=$(date +%s)
  elapsed_time=$((current_time - start_time))

  # Check if the timeout has been exceeded.
  if [ "$elapsed_time" -ge "$TIMEOUT" ]; then
    echo "Timeout of ${TIMEOUT}s reached. CRD '$CRD_NAME' not found."
    exit 1
  fi

  # Query for the CRD and check its 'Established' condition.
  # We use `kubectl get` with a JSONPath expression to extract the status
  # of the 'Established' condition.
  # The command will succeed (exit code 0) and output "True" if the CRD is ready.
  status=$(kubectl get crd "$CRD_NAME" -o jsonpath='{.status.conditions[?(@.type=="Established")].status}' 2>/dev/null)

  # Check if the status is "True".
  if [ "$status" == "True" ]; then
    echo "Found '$CRD_NAME'"
    break
  else
    echo "Waiting for '$CRD_NAME' to be established. Retrying in ${SLEEP_INTERVAL}s... (Elapsed: ${elapsed_time}s)"
    sleep "$SLEEP_INTERVAL"
  fi
done

exit 0
