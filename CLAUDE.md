# CLAUDE.md - Project Overview

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Red Hat OpenShift bootstrap repository that contains GitOps infrastructure for deploying and managing OpenShift clusters across multiple regions. The project uses OpenShift GitOps (ArgoCD), Red Hat Advanced Cluster Management (ACM), and Hive for cluster lifecycle management.

## Architecture

The codebase is organized into several key components:

### Core Components
- **Bootstrap Control Plane**: Uses OpenShift GitOps to manage the initial cluster setup
- **Cluster Provisioning**: Uses CAPI (Cluster API) for automated cluster creation
- **Regional Management**: Uses ACM for multi-cluster management across regions
- **Configuration Management**: Pure Kustomize-based approach for generating cluster manifests

### Directory Structure
- `clusters/`: Cluster deployment configurations (base + overlays)
- `operators/`: Operator deployments organized by operator type
  - `operators/advanced-cluster-management/global/`: ACM hub cluster deployment
  - `operators/openshift-pipelines/global/`: Pipelines hub cluster deployment  
  - `operators/openshift-pipelines/cluster-*/`: Tekton Pipelines operator per managed cluster
- `prereqs/`: Prerequisites for bootstrap process
- `pipelines/`: Tekton pipeline configurations deployed per region
- `deployments/`: Service deployments (OCM services) per cluster
- `gitops-applications/`: ArgoCD ApplicationSets for GitOps automation

## Key Technologies

- **OpenShift GitOps (ArgoCD)**: Continuous deployment and cluster management
- **Red Hat Advanced Cluster Management (ACM)**: Multi-cluster management with CAPI integration
- **Cluster API (CAPI)**: Kubernetes-native cluster lifecycle management
- **Hive**: OpenShift cluster provisioning operator (for OCP clusters)
- **Infrastructure Providers**: AWS, Azure, GCP, vSphere, OpenStack, BareMetal (via ACM)
- **Kustomize**: YAML configuration management and templating
- **Tekton Pipelines**: CI/CD workflows

## Claude Memories

- Don't run `bootstrap.sh` from a Claude session
- When provisioning or managing OpenShift, always use `oc` client

## Documentation Navigation

For detailed information, see:
- **[Architecture](./ARCHITECTURE.md)** - Visual diagrams and technical architecture
- **[Installation](./INSTALL.md)** - Complete setup guide
- **[Cluster Creation](./guides/cluster-creation.md)** - End-to-end cluster deployment
- **[Monitoring](./guides/monitoring.md)** - Status checking and troubleshooting
- **[Documentation Index](./docs/INDEX.md)** - Complete documentation reference