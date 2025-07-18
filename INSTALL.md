# OpenShift Bootstrap Installation Guide

This guide provides comprehensive instructions for setting up and managing multi-cluster OpenShift deployments using GitOps automation.

## Overview

This repository enables automated management of OpenShift clusters across multiple regions using:
- **Hub Cluster**: OpenShift cluster running ArgoCD, ACM, and cluster management operators
- **Managed Clusters**: Regional OpenShift (OCP) or EKS clusters provisioned and managed via GitOps
- **Automated Provisioning**: Using `bin/generate-cluster` tool for consistent cluster deployment

## Table of Contents

1. [Hub Cluster Setup](#hub-cluster-setup) (One-time setup)
2. [Adding New Regions](#adding-new-regions) (Repeatable process)
3. [Monitoring and Management](#monitoring-and-management)
4. [Troubleshooting](#troubleshooting)

---

## Hub Cluster Setup

**Prerequisites:**
- OpenShift 4.12+ cluster with cluster-admin permissions
- Minimum 16GB RAM, 4 vCPUs for control plane workloads
- Network connectivity to target regions for cluster provisioning
- AWS credentials for cluster provisioning

### 1. Repository Setup

```bash
git clone https://github.com/openshift-online/bootstrap.git
cd bootstrap
```

### 2. Authentication

```bash
# Log in to your OpenShift hub cluster
oc login https://api.your-hub-cluster.example.com:6443 --token=your-token
```

### 3. Bootstrap Control Plane

```bash
# Run the automated bootstrap script
./bootstrap.sh
```

**What this script does:**
1. **Deploys Prerequisites** (`oc apply -k ./prereqs`):
   - OpenShift GitOps Operator subscription
   - Cluster role bindings for ArgoCD
   - Service accounts for cluster import
   - ArgoCD Tekton resource exclusions

2. **Creates Secrets** (`./bootstrap.vault.sh`):
   - AWS credentials for cluster provisioning
   - Pull secrets for OpenShift installations
   - SSH keys for cluster access

3. **Deploys GitOps Applications** (`oc apply -k ./gitops-applications`):
   - ACM operator and MultiClusterHub instance
   - Regional cluster ApplicationSets
   - Tekton pipelines for CI/CD

4. **Waits for Completion**:
   - OpenShift GitOps route availability
   - ACM hub components readiness
   - Regional cluster provisioning status

### 4. Access Control Plane

```bash
# Get the ArgoCD admin password
oc extract secret/openshift-gitops-cluster -n openshift-gitops --to=-

# Access the ArgoCD UI
oc get route openshift-gitops-server -n openshift-gitops
```

### 5. Verify Hub Setup

```bash
# Check operators are running
oc get csv -n openshift-operators | grep -E "(gitops|advanced-cluster-management|pipelines)"

# Verify ACM hub status
oc get mch -n open-cluster-management

# Check ArgoCD applications
oc get applications -n openshift-gitops
```

---

## Adding New Regions

Use this process to add new OpenShift (OCP) or EKS clusters to your hub.

### Prerequisites

- Hub cluster successfully bootstrapped
- AWS credentials configured for target region
- Cluster naming convention decided (e.g., cluster-XX)

### Step 1: Create Regional Specification

```bash
# Create regional spec directory
mkdir -p regions/[AWS_REGION]/[CLUSTER_NAME]/

# Create region.yaml specification file
cat > regions/[AWS_REGION]/[CLUSTER_NAME]/region.yaml << EOF
type: [CLUSTER_TYPE]  # "ocp" or "eks"
name: [CLUSTER_NAME]  # e.g., cluster-50
region: [AWS_REGION]  # e.g., us-west-2
domain: rosa.mturansk-test.csu2.i3.devshift.org
instanceType: [COMPUTE_TYPE]  # e.g., m5.large
replicas: [NODE_COUNT]  # e.g., 3
EOF
```

**Example configurations:**

```bash
# OpenShift cluster in us-west-2
mkdir -p regions/us-west-2/cluster-50/
cat > regions/us-west-2/cluster-50/region.yaml << EOF
type: ocp
name: cluster-50
region: us-west-2
domain: rosa.mturansk-test.csu2.i3.devshift.org
instanceType: m5.large
replicas: 3
EOF

# EKS cluster in eu-west-1
mkdir -p regions/eu-west-1/cluster-51/
cat > regions/eu-west-1/cluster-51/region.yaml << EOF
type: eks
name: cluster-51
region: eu-west-1
domain: rosa.mturansk-test.csu2.i3.devshift.org
instanceType: m5.xlarge
replicas: 5
EOF
```

### Step 2: Generate Complete Cluster Overlay

```bash
# Run automated cluster generator
./bin/generate-cluster regions/[AWS_REGION]/[CLUSTER_NAME]/
```

**What this generates:**
- **Cluster overlay** (`clusters/[CLUSTER_NAME]/`) - OCP (Hive) or EKS (CAPI) resources
- **Pipeline overlays** - Hello World and Cloud Infrastructure pipelines  
- **Operators overlay** - OpenShift Pipelines operator configuration
- **Deployments overlay** - OCM service deployments
- **ApplicationSet** - GitOps applications with proper sync waves
- **Updated kustomization** - Adds ApplicationSet to gitops-applications

### Step 3: Validate Generated Manifests

```bash
# Test all kustomize builds
kustomize build clusters/[CLUSTER_NAME]/
kustomize build pipelines/hello-world/[CLUSTER_NAME]/
kustomize build pipelines/cloud-infrastructure-provisioning/[CLUSTER_NAME]/
kustomize build operators/openshift-pipelines/[CLUSTER_NAME]/
kustomize build deployments/ocm/[CLUSTER_NAME]/
kustomize build gitops-applications/

# Dry-run validation
oc --dry-run=client apply -k clusters/[CLUSTER_NAME]/
```

### Step 4: Deploy via GitOps

```bash
# Deploy new cluster through GitOps
./bootstrap.sh
```

### Step 5: Monitor Deployment

```bash
# Check ApplicationSet status
oc get applicationset [CLUSTER_NAME]-applications -n openshift-gitops

# Check generated applications (sync wave order: 1=cluster, 2=operators, 3=pipelines, 4=deployments)
oc get applications -n openshift-gitops | grep [CLUSTER_NAME]

# Monitor cluster provisioning
# For OCP clusters:
oc get clusterdeployment [CLUSTER_NAME] -n [CLUSTER_NAME]
oc get machinepool [CLUSTER_NAME]-worker -n [CLUSTER_NAME]

# For EKS clusters:
oc get awsmanagedcontrolplane [CLUSTER_NAME] -n [CLUSTER_NAME]
oc get awsmanagedmachinepool [CLUSTER_NAME] -n [CLUSTER_NAME]

# Check ACM import
oc get managedcluster [CLUSTER_NAME]
oc get klusterletaddonconfig [CLUSTER_NAME] -n [CLUSTER_NAME]
```

### Step 6: Verify Deployment

```bash
# Wait for cluster provisioning completion (typically 30-45 minutes)
# Check ArgoCD for application sync status
oc get applications -n openshift-gitops

# Verify services deployed (once cluster is ready)
# Note: Requires kubeconfig for target cluster
oc get deployment -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig
oc get pipeline -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig
```

---

## Monitoring and Management

### ArgoCD Applications

```bash
# List all applications
oc get applications -n openshift-gitops

# Check specific application status
oc describe application [CLUSTER_NAME]-cluster -n openshift-gitops

# View ArgoCD UI
oc get route openshift-gitops-server -n openshift-gitops
```

### Cluster Status

```bash
# List all managed clusters
oc get managedclusters

# Check cluster health
oc get clusterdeployments -A  # For OCP clusters
oc get awsmanagedcontrolplane -A  # For EKS clusters

# View cluster details
oc describe managedcluster [CLUSTER_NAME]
```

### ACM Console

```bash
# Access ACM console
oc get route multicloud-console -n open-cluster-management
```

---

## Troubleshooting

### Common Issues

#### 1. Bootstrap Script Fails

```bash
# Check cluster connectivity
oc cluster-info

# Verify prerequisites
oc get subscription openshift-gitops-operator -n openshift-operators

# Check operator status
oc get csv -n openshift-operators | grep gitops
```

#### 2. Cluster Provisioning Stuck

```bash
# Check cluster deployment events
oc describe clusterdeployment [CLUSTER_NAME] -n [CLUSTER_NAME]

# Verify AWS credentials
oc get secret aws-credentials -n [CLUSTER_NAME] -o yaml

# Check Hive controller logs (for OCP clusters)
oc logs -n hive deployment/hive-controllers

# Check CAPI controller logs (for EKS clusters)
oc logs -n capa-system deployment/capa-controller-manager
```

#### 3. ArgoCD Application Sync Issues

```bash
# Check application status
oc get applications -n openshift-gitops

# View application details
oc describe application [CLUSTER_NAME]-cluster -n openshift-gitops

# Check ArgoCD server logs
oc logs -n openshift-gitops deployment/openshift-gitops-server
```

#### 4. Kustomize Build Failures

```bash
# Test individual builds
kustomize build clusters/[CLUSTER_NAME]/
kustomize build gitops-applications/

# Validate YAML syntax
oc --dry-run=client apply -k clusters/[CLUSTER_NAME]/
```

### Validation Commands

```bash
# Test all major components
kustomize build ./clusters/[CLUSTER_NAME]
kustomize build ./gitops-applications
oc get managedclusters
oc get mch -n open-cluster-management
```

### Rollback Procedures

If deployment fails, clean up with:

```bash
# Delete ApplicationSet and all generated applications
oc delete applicationset [CLUSTER_NAME]-applications -n openshift-gitops

# Delete cluster namespace
oc delete namespace [CLUSTER_NAME]

# Remove overlay directories
rm -rf ./clusters/[CLUSTER_NAME]
rm -rf ./pipelines/hello-world/[CLUSTER_NAME]
rm -rf ./pipelines/cloud-infrastructure-provisioning/[CLUSTER_NAME]
rm -rf ./operators/openshift-pipelines/[CLUSTER_NAME]
rm -rf ./deployments/ocm/[CLUSTER_NAME]

# Remove ApplicationSet file
rm ./gitops-applications/[CLUSTER_NAME].yaml

# Remove regional specification
rm -rf ./regions/[AWS_REGION]/[CLUSTER_NAME]

# Update gitops-applications/kustomization.yaml to remove ApplicationSet reference
# Remove line: - ./[CLUSTER_NAME].yaml
```

---

## Architecture Notes

### Hub-Spoke Model
- **Hub Cluster**: Runs ArgoCD, ACM, and all cluster management operators
- **Managed Clusters**: Regional clusters (OCP or EKS) managed by the hub
- **GitOps Flow**: Hub → Cluster Provisioning → Operator Installation → Pipeline Deployment → Service Deployment

### Sync Wave Ordering
Applications deploy in this order via ApplicationSet sync waves:
1. **Wave 1**: Cluster provisioning (CAPI/Hive resources)
2. **Wave 2**: Operators installation (OpenShift Pipelines)  
3. **Wave 3**: Pipeline deployment (Tekton resources)
4. **Wave 4**: Service deployment (OCM services)

### Resource Types by Platform
- **OCP Clusters**: ClusterDeployment + MachinePool (via Hive)
- **EKS Clusters**: AWSManagedControlPlane + AWSManagedMachinePool (via CAPI v1beta2)
- **GitOps**: ApplicationSets with sync waves for proper dependency management

### Security
- All secrets managed through OpenShift native secret management
- Cluster access controlled via RBAC policies
- Network policies isolate cluster management workloads
- Automated ArgoCD cluster registration via ACM GitOpsCluster resources

This installation method provides a production-ready multi-cluster management platform with full GitOps automation for both OpenShift and EKS clusters.