# CLAUDE.md

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
- `gitops-applications/`: ArgoCD Application manifests
- `operators/`: Operator deployments (ACM, Pipelines, etc.)
- `regional-deployments/`: Regional service configurations
- `prereqs/`: Prerequisites for bootstrap process
- `pipelines/`: Tekton pipeline configurations
- `acm-gitops/`: ACM GitOps integration with automated ArgoCD cluster registration

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
- **Kustomize**: YAML configuration management and templating
- **Tekton Pipelines**: CI/CD workflows

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

## Kustomize Configuration Management

The project uses pure Kustomize for generating cluster manifests:
- ClusterDeployment manifests for Hive cluster provisioning
- InstallConfig Secrets for OpenShift installation configuration
- MachinePool manifests for worker node pool definitions
- ManagedCluster manifests for ACM cluster management

### Configuration Structure

- **Base configurations**: Common templates in `clusters/base/`
- **Overlays**: Environment-specific customizations in `clusters/overlay/`
- **Patches**: Kustomize patches for region-specific modifications
- **Generators**: ConfigMap and Secret generators for cluster-specific data

## Important Notes

- This project manages production OpenShift infrastructure
- All cluster changes go through GitOps workflows
- Secrets management is handled through external processes (Vault integration planned)
- The bootstrap process requires cluster-admin permissions
- Regional clusters are provisioned automatically via ACM and Hive

## Adding New Clusters

To add a new cluster (e.g., `cluster-40`), follow this manual process:

### Prerequisites
- Bootstrap control plane cluster running with cluster-admin access
- AWS credentials and pull secrets configured
- Secrets stored in `secrets/aws-creds.yaml` and `secrets/pull-secret.yaml`

### Step-by-Step Process
1. **Copy existing overlay**: Copy `./clusters/overlay/region-02` to `./clusters/overlay/region-04`
2. **Update cluster references**: Find/Replace 'cluster-10' with 'cluster-40' in all files within the new overlay
3. **Add to kustomization**: Add the new cluster overlay to `./regional-clusters/kustomization.yaml`
4. **Update bootstrap script**: Add the new cluster to `./bootstrap.sh` for status monitoring
5. **Create Pull Request**: Submit changes to the repository
6. **Deploy**: Run `./bootstrap.sh` to provision cluster-40 or monitor via ACM console

### GitOps Workflow
- ArgoCD applies the new cluster via the [regional clusters](./gitops-applications/regional-clusters.application.yaml) application
- Hive ClusterDeployments handle automated cluster provisioning
- ACM imports and manages the new cluster for governance

### Secret Management
Currently uses manual secret management (Vault integration planned):
```bash
# Retrieve secrets from ACM for each cluster namespace
oc get secret aws-creds -n $cluster_namespace -o yaml > secrets/aws-creds.yaml
oc get secret pull-secret -n $cluster_namespace -o yaml > secrets/pull-secret.yaml
```

## ACM GitOps Integration

The project uses ACM's native GitOps integration to automatically register ManagedClusters with ArgoCD:

### Components
- **GitOpsCluster CR**: Automatically registers clusters with ArgoCD based on Placement selection
- **ManagedClusterSetBinding**: Binds the global ManagedClusterSet to openshift-gitops namespace
- **Placement**: Selects clusters based on labels (OpenShift + Amazon)
- **Policy**: Automates the creation of GitOps resources across clusters

### Features
- **Automated Cluster Registration**: No manual ArgoCD secret management required
- **ApplicationManager Integration**: KlusterletAddonConfig enables ArgoCD permissions on target clusters
- **Policy-Driven**: ACM policies ensure consistent GitOps configuration across all clusters
- **Label-Based Selection**: Clusters are automatically included based on vendor=OpenShift, cloud=Amazon labels

## Development Best Practices

- Use `kustomize build` to validate configuration changes
- Test overlays before applying to clusters
- Follow GitOps principles for all cluster modifications
- Reference existing cluster overlays (cluster-10, region-02, region-03) as templates