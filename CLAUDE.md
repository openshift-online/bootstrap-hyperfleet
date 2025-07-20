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

The project uses semantic directory organization with consistent patterns:

**Top-level "things":**
- `clusters/`: Cluster deployment configurations (auto-generated from regions/)
- `operators/`: Operator/application deployments following {operator-name}/{deployment-target} pattern
- `pipelines/`: Tekton pipeline configurations following {pipeline-name}/{cluster-name} pattern
- `deployments/`: Service deployments following {service-name}/{cluster-name} pattern
- `regions/`: Regional cluster specifications (input for generation)
- `bases/`: Reusable Kustomize base components
- `gitops-applications/`: ArgoCD ApplicationSets for GitOps automation
- `prereqs/`: Prerequisites for bootstrap process

**Operator deployments organized semantically:**
- `operators/advanced-cluster-management/global/`: ACM hub cluster deployment
- `operators/gitops-integration/global/`: GitOps integration policies and configurations
- `operators/openshift-pipelines/global/`: Pipelines hub cluster deployment  
- `operators/openshift-pipelines/{cluster-name}/`: Tekton Pipelines operator per managed cluster
- `operators/vault/global/`: Vault secret management system

**Deployment targets:**
- `global/`: Hub cluster deployments (shared infrastructure)
- `{cluster-name}/`: Managed cluster-specific deployments (e.g., `ocp-02/`, `eks-01/`)

## Key Technologies

- **OpenShift GitOps (ArgoCD)**: Continuous deployment and cluster management
- **Red Hat Advanced Cluster Management (ACM)**: Multi-cluster management with CAPI integration
- **Cluster API (CAPI)**: Kubernetes-native cluster lifecycle management
- **Hive**: OpenShift cluster provisioning operator (for OCP clusters)
- **Infrastructure Providers**: AWS, Azure, GCP, vSphere, OpenStack, BareMetal (via ACM)
- **Kustomize**: YAML configuration management and templating
- **Tekton Pipelines**: CI/CD workflows

## Claude Memories

- Don't run `bin/bootstrap.sh` from a Claude session
- When provisioning or managing OpenShift, always use `oc` client
- Critical! Always use smart semantic naming for maximum usability and comprehensive

## Documentation Navigation

For detailed information, see:
- **[Architecture](./docs/architecture/ARCHITECTURE.md)** - Visual diagrams and technical architecture
- **[Installation](./docs/getting-started/production-installation.md)** - Complete setup guide
- **[Cluster Creation](./guides/cluster-creation.md)** - End-to-end cluster deployment
- **[Monitoring](./guides/monitoring.md)** - Status checking and troubleshooting
- **[Documentation Index](./docs/INDEX.md)** - Complete documentation reference