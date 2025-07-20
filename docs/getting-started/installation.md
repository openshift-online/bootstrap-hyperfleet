# Installation Guide

**Audience**: New users  
**Complexity**: Beginner to Intermediate  
**Estimated Time**: 1-2 hours for basic setup  
**Prerequisites**: Basic OpenShift knowledge, AWS credentials

## Overview

This guide helps you set up the OpenShift Bootstrap hub cluster and deploy your first regional cluster. For production deployments and advanced configuration, see the [Complete Installation Guide](../../INSTALL.md).

## Quick Setup Path

### 1. Prerequisites

Before starting, ensure you have:

- **OpenShift 4.12+ cluster** with cluster-admin permissions
- **AWS credentials** for cluster provisioning  
- **Basic tools**: `oc`, `kubectl`, `kustomize`
- **Git repository access** for pushing configurations

### 2. Repository Setup

```bash
git clone https://github.com/openshift-online/bootstrap.git
cd bootstrap

# Log in to your OpenShift hub cluster
oc login https://api.your-hub-cluster.example.com:6443 --token=your-token
```

### 3. Bootstrap Hub Cluster

Run the automated bootstrap script to set up your hub cluster:

```bash
./bin/bootstrap.sh
```

**What this does:**
- Installs OpenShift GitOps (ArgoCD)
- Deploys Advanced Cluster Management (ACM)
- Sets up cluster role bindings and service accounts
- Creates necessary secrets for cluster provisioning
- Configures GitOps applications for cluster management

**Expected time:** 15-30 minutes

### 4. Verify Hub Setup

Check that core components are running:

```bash
# Check operators
oc get csv -n openshift-operators | grep -E "(gitops|advanced-cluster-management|pipelines)"

# Verify ACM hub
oc get mch -n open-cluster-management

# Check ArgoCD applications
oc get applications -n openshift-gitops
```

### 5. Access Management Interfaces

```bash
# Get ArgoCD admin password
oc extract secret/openshift-gitops-cluster -n openshift-gitops --to=-

# Get ArgoCD URL
oc get route openshift-gitops-server -n openshift-gitops

# Get ACM console URL
oc get route multicloud-console -n open-cluster-management
```

## Deploy Your First Cluster

### 1. Create Cluster Specification

Use the interactive tool to configure your first regional cluster:

```bash
./bin/new-cluster
```

Follow the prompts to specify:
- Cluster name (e.g., `my-first-cluster`)
- Cluster type (`ocp` for OpenShift, `eks` for EKS)
- AWS region (e.g., `us-east-1`)
- Instance type and replica count

### 2. Deploy via GitOps

The cluster generation tool automatically updates GitOps configurations. Deploy them:

```bash
./bin/bootstrap.sh
```

### 3. Monitor Progress

Track your cluster deployment:

```bash
# Check ApplicationSet creation
oc get applicationset -n openshift-gitops

# Monitor applications for your cluster
oc get applications -n openshift-gitops | grep my-first-cluster

# For EKS clusters
oc get awsmanagedcontrolplane my-first-cluster -n my-first-cluster

# For OCP clusters  
oc get clusterdeployment my-first-cluster -n my-first-cluster
```

**Expected time:** 15-45 minutes depending on cluster type

### 4. Verify Deployment

Check that your cluster is successfully provisioned and managed:

```bash
# Check managed cluster status
oc get managedcluster my-first-cluster

# Verify ArgoCD applications are synced
oc get applications -n openshift-gitops | grep my-first-cluster

# Run health check
./bin/health-check
```

## Next Steps

Once your hub cluster and first regional cluster are running:

1. **Explore**: [Deploy Your First Cluster](./first-cluster.md) - Detailed walkthrough
2. **Understand**: [Core Concepts](./concepts.md) - Architecture and workflow
3. **Scale**: [Complete Installation Guide](./production-installation.md) - Production setup
4. **Monitor**: [Monitoring Guide](../../guides/monitoring.md) - Health checking

## Common Issues

### Bootstrap Script Fails

```bash
# Check cluster connectivity
oc cluster-info

# Verify prerequisites
oc get subscription openshift-gitops-operator -n openshift-operators
```

### Cluster Provisioning Stuck

```bash
# Check AWS credentials
aws sts get-caller-identity

# For EKS clusters - check CAPI logs
oc logs -n capi-aws-system deployment/capa-controller-manager

# For OCP clusters - check Hive logs  
oc logs -n hive deployment/hive-controllers
```

### ArgoCD Application Issues

```bash
# Check application status
oc get applications -n openshift-gitops

# View application details
oc describe application my-first-cluster-cluster -n openshift-gitops
```

## Support

- **Quick Help**: [5-Minute Quickstart](./quickstart.md)
- **Detailed Guide**: [Deploy Your First Cluster](./first-cluster.md)
- **Production Setup**: [Complete Installation Guide](./production-installation.md)
- **Troubleshooting**: [Monitoring Guide](../../guides/monitoring.md)