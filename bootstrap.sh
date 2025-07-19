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

./status.sh applications.argoproj.io

# Apply the GitOps Applications to complete bootstrap
echo "Applying the GitOps Applications to complete bootstrap"
oc apply -k ./gitops-applications

echo "Waiting for openshift-gitops (aka Argo) to complete"
./wait.kube.sh route openshift-gitops-server openshift-gitops {.kind} Route

echo "Waiting for ACM to complete"
./wait.kube.sh mch multiclusterhub open-cluster-management '{.status.conditions[?(@.type=="Complete")].message}' "All hub components ready."

echo "Control plane is ready: $(oc whoami --show-console)"

echo "Waiting for regional clusters to provision"
# Note: Cluster provisioning is handled by GitOps - clusters will be created automatically
# Wait times depend on cluster type: EKS (~15min), OCP via Hive (~45min), HCP (~10min)

echo "GitOps has begun cluster provisioning. You can monitor progress with:"
echo "  oc get applications -n openshift-gitops"
echo "  oc get clusterdeployments -A  # For OCP clusters"
echo "  oc get clusters -A            # For EKS clusters" 
echo "  oc get hostedclusters -A      # For HCP clusters"
echo ""

echo "Setting up Vault-based secret management for provisioned cluster namespaces"
echo "Note: ExternalSecrets will be created but won't sync until clusters are fully provisioned"
./bootstrap.vault-integration.sh