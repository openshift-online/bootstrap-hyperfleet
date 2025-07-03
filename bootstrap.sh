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
./wait.kube.sh cd cluster-10 cluster-10 '{.status.conditions[?(@.type=="Provisioned")].message}' "Cluster is provisioned"

./wait.kube.sh cd cluster-20 cluster-20 '{.status.conditions[?(@.type=="Provisioned")].message}' "Cluster is provisioned"