# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Red Hat OpenShift bootstrap repository that contains GitOps infrastructure for deploying and managing OpenShift clusters across multiple regions. The project uses OpenShift GitOps (ArgoCD), Red Hat Advanced Cluster Management (ACM), and Hive for cluster lifecycle management.

## Architecture

The codebase is organized into several key components:

### Core Components
- **Bootstrap Control Plane**: Uses OpenShift GitOps to manage the initial cluster setup
- **Cluster Provisioning**: Leverages Hive ClusterDeployments for automated cluster creation
- **Regional Management**: Uses ACM for multi-cluster management across regions
- **ACME Tool**: Go-based CLI tool for generating cluster configuration manifests

### Directory Structure
- `acme/`: Go CLI tool for cluster configuration generation
- `clusters/`: Cluster deployment configurations (base + overlays)
- `gitops-applications/`: ArgoCD Application manifests
- `operators/`: Operator deployments (ACM, Pipelines, etc.)
- `regional-deployments/`: Regional service configurations
- `prereqs/`: Prerequisites for bootstrap process
- `pipelines/`: Tekton pipeline configurations

## Common Commands

### Bootstrap Process
```bash
# Initial cluster bootstrap (requires cluster-admin and kubeconfig)
./bootstrap.sh

# Check cluster status
./status.sh applications.argoproj.io

# Wait for specific components
./wait.kube.sh route openshift-gitops-server openshift-gitops {.kind} Route
```

### ACME Tool (Go CLI)
```bash
# Build the ACME tool
cd acme && make build

# Generate cluster configurations
cd acme && make run

# Run tests
cd acme && make test

# Install dependencies
cd acme && make install-deps
```

### Development Commands
```bash
# Root level (no specific commands defined)
# Build Docker images
make podman-build

# Run with Docker
make podman-run
```

## Key Technologies

- **OpenShift GitOps (ArgoCD)**: Continuous deployment and cluster management
- **Red Hat Advanced Cluster Management (ACM)**: Multi-cluster management
- **Hive**: OpenShift cluster provisioning operator
- **Kustomize**: YAML configuration management
- **Tekton Pipelines**: CI/CD workflows
- **Go**: ACME tool implementation with Kubernetes client libraries

## Cluster Management Workflow

1. **Bootstrap**: Run `./bootstrap.sh` to set up the control plane
2. **Provision**: Clusters are automatically provisioned via Hive ClusterDeployments
3. **Import**: ACM imports managed clusters for governance
4. **Deploy**: ArgoCD deploys applications to target clusters
5. **Monitor**: Status monitoring via custom wait scripts

## Configuration Management

- Base configurations in `clusters/base/`
- Environment-specific overlays in `clusters/overlay/`
- Regional configurations support multiple availability zones
- Kustomize for YAML templating and patching

## ACME Tool Details

The ACME tool is a Go CLI application that:
- Generates ClusterDeployment, InstallConfig, MachinePool, and ManagedCluster manifests
- Reads cluster specifications from code
- Outputs JSON configurations for GitOps consumption
- Supports multiple cluster regions and configurations

### Data Model

The ACME tool implements a structured data model in `acme/pkg/api/` that represents the relationships between cluster entities:

**Core Entity:**
- `CentralControlPlane` - Top-level entity representing the bootstrap control plane that orchestrates regional cluster management

**Entity Relationships:**
- `CentralControlPlane` 1:Many `RegionalCluster` - One control plane manages multiple regional clusters
- `CentralControlPlane` 1:1 `ClusterDeploymentConfig` - Control plane has its own configuration
- `RegionalCluster` 1:1 `ClusterDeploymentConfig` - Each regional cluster has configuration parameters
- `RegionalCluster` 1:1 `ClusterDeployment` - Hive cluster provisioning resource
- `RegionalCluster` 1:1 `InstallConfig` - OpenShift installation configuration
- `RegionalCluster` 1:1 `MachinePool` - Hive worker node pool definition  
- `RegionalCluster` 1:1 `ManagedCluster` - ACM cluster management resource

**Constructor Pattern:**
- `NewRegionalCluster(config)` creates all related entities from a single ClusterDeploymentConfig
- Individual constructors available for each entity type (NewClusterDeployment, NewInstallConfig, etc.)

**Testing:**
```bash
# Test the data model and entity generation
cd acme && go run cmd/main.go clusters
```

## Important Notes

- This project manages production OpenShift infrastructure
- All cluster changes go through GitOps workflows
- Secrets management is handled through external processes (Vault integration planned)
- The bootstrap process requires cluster-admin permissions
- Regional clusters are provisioned automatically via ACM and Hive