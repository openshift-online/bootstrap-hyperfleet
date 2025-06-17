#!/bin/bash

# Verify that the user is logged into a Kubernetes Cluster
if [[ ! $(oc cluster-info) ]]; then
  echo "Please log in to an OpenShift cluster using 'oc login'"
  exit 1
fi

echo "Content will be deployed to the cluster running at:";
oc cluster-info

# Deploy prereqs, including the OpenShift GitOps Operator
echo "Applying a subscription to the OpenShift GitOps Operator"
oc apply -k ./prereqs

# Wait for the Application CR to become available
echo "Waiting for the Application CR to become available"
while [[ ! $(oc get CustomResourceDefinition/applications.argoproj.io) && i < 10 ]]; do
  echo "Waiting for Application CR to become available..."
  sleep 5
  i=$((i + 1))
done

# Apply the GitOps Applications to complete bootstrap
echo "Applying the GitOps Applications to complete bootstrap"
oc apply -k ./gitops-applications
