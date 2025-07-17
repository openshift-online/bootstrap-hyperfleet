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

### EKS Global Hub Connection
```bash
# Connect to EKS hub cluster (acme-test-001)
aws eks update-kubeconfig --region us-east-1 --name acme-test-001 --profile default

# Hub cluster details:
# - Name: acme-test-001
# - Region: us-east-1
# - Endpoint: https://7CE7E6372FDBCCC16A73A03435D729C3.gr7.us-east-1.eks.amazonaws.com
# - OIDC Provider: https://oidc.eks.us-east-1.amazonaws.com/id/7CE7E6372FDBCCC16A73A03435D729C3
# - IAM Role: arn:aws:iam::765374464689:role/AmazonEKSAutoClusterRole

# Note: User must be added to aws-auth ConfigMap for cluster access
```

## Current Session Status

### **Session State (as of 2025-07-17)**
- **Hub Cluster**: OpenShift cluster with ArgoCD, ACM, and Tekton Pipelines
- **Kubeconfig**: Connected to hub cluster
- **Status**: ✅ Active development on OCM-16599/capi_poc branch
- **Tools Available**: kubectl, oc, git, kustomize

### **Completed Work**
1. ✅ **ArgoCD Tekton Integration**: Fixed resource exclusions to allow Pipeline/PipelineRun resources
2. ✅ **Pipeline Deployment**: Resolved Pipeline resource deployment issues on managed clusters
3. ✅ **Regional Deployments**: Fixed OpenShift Pipelines operator deployment to all clusters
4. ✅ **GitOps Sync Waves**: Added proper ordering (cluster → pipelines → deployments)
5. ✅ **Cluster Health Monitoring**: Fixed cluster health monitoring script
6. ✅ **Multi-Cluster Setup**: Established proper GitOps structure for multiple clusters

### **Key Fixes Applied**
- **ArgoCD Resource Exclusions**: Patch ArgoCD CR (not ConfigMap) to allow Tekton resources
- **Tekton CRDs**: Added OpenShift Pipelines operator to regional-deployments/base
- **GitOps Applications**: Created proper applications for cluster-20, cluster-30 pipeline deployment
- **Sync Ordering**: Applied sync waves to ensure cluster provisioning before pipeline deployment
- **Automation**: Created `bin/patch-argocd-tekton` script and `prereqs/argocd-tekton-exclusions.yaml`

### **Current Architecture**
- **Hub Cluster**: Manages ArgoCD applications and cluster provisioning
- **Spoke Clusters**: cluster-10, cluster-20, cluster-30 (OpenShift) + cluster-40+ (EKS)
- **Pipeline Integration**: Tekton Pipelines deployed to all managed clusters
- **Regional Deployments**: Operators and services deployed per cluster
- **GitOps Flow**: Hub → Cluster Provisioning → Pipeline Deployment → Service Deployment

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
- **Red Hat Advanced Cluster Management (ACM)**: Multi-cluster management with CAPI integration
- **Cluster API (CAPI)**: Kubernetes-native cluster lifecycle management
- **Hive**: OpenShift cluster provisioning operator (for OCP clusters)
- **Infrastructure Providers**: AWS, Azure, GCP, vSphere, OpenStack, BareMetal (via ACM)
- **Kustomize**: YAML configuration management and templating
- **Tekton Pipelines**: CI/CD workflows

## Cluster Management Workflow

1. **Bootstrap**: Run `./bootstrap.sh` to set up the control plane
2. **Provision**: Clusters are automatically provisioned via:
   - **OCP Clusters**: Hive ClusterDeployments
   - **EKS Clusters**: CAPI with ACM infrastructure providers
3. **Import**: ACM imports managed clusters for governance
4. **Deploy**: ArgoCD deploys applications to target clusters
5. **Monitor**: Status monitoring via custom wait scripts

## Infrastructure Provider Integration

ACM is configured with infrastructure providers that automatically install and manage CAPI controllers:

### Enabled Providers
- **AWS**: EKS clusters via AWSManagedControlPlane and AWSManagedMachinePool
- **Azure**: AKS clusters via Azure infrastructure provider
- **GCP**: GKE clusters via GCP infrastructure provider
- **vSphere**: On-premises clusters via vSphere provider
- **OpenStack**: OpenStack-based clusters
- **BareMetal**: Physical machine clusters

### CAPI Integration
- ACM MultiClusterHub automatically installs CAPI CRDs for enabled providers
- No need for standalone CAPI operators
- Infrastructure providers managed through ACM lifecycle
- Seamless integration with ACM's cluster governance and policies

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

## Claude Memories

- Don't run `bootstrap.sh` from a Claude session
- When provisioning or managing OpenShift, always use `oc` client

## Adding New Clusters

### Current Implementation: cluster-40 (EKS)
**Status**: Implemented and ready for deployment
- **Type**: EKS cluster using CAPI v1beta1 resources (via ACM infrastructure providers)
- **Region**: ap-southeast-1
- **Compute**: m5.large instances (3 nodes, scaling 1-10)
- **Base Domain**: rosa.mturansk-test.csu2.i3.devshift.org
- **Resources**: AWSManagedControlPlane, AWSManagedMachinePool, ArgoCD applications
- **GitOps**: Configured with ArgoCD applications for cluster + regional deployments
- **CAPI CRDs**: Automatically installed by ACM infrastructure providers

### General Process for Adding New Clusters

#### For OCP Clusters (Hive-based):
1. **Copy existing overlay**: Copy `./clusters/overlay/cluster-20` to `./clusters/overlay/cluster-XX`
2. **Update cluster references**: Find/Replace cluster names in all files
3. **Configure region and compute**: Update install-config.yaml for target region/instance type
4. **Uses**: ClusterDeployment + MachinePool resources

#### For EKS Clusters (CAPI-based):
1. **Create overlay directory**: `mkdir -p ./clusters/overlay/cluster-XX`
2. **Create CAPI resources**: AWSManagedControlPlane + AWSManagedMachinePool
3. **Configure region and compute**: Set AWS region and instance type in CAPI resources
4. **Set base domain**: Add baseDomain to AWSManagedControlPlane
5. **Uses**: CAPI v1beta1 API versions (compatible with ACM infrastructure providers)
6. **CRDs**: Automatically installed by ACM when infrastructure providers are enabled

#### Common Steps:
1. **Create regional deployment overlay**: Copy and update `./regional-deployments/overlays/`
2. **Create ArgoCD applications**: Copy and update cluster + regional deployment apps
3. **Update kustomization**: Add applications to `./gitops-applications/kustomization.yaml`
4. **Deploy**: Run `./bootstrap.sh` to provision cluster (not from Claude session)

### GitOps Workflow
- ArgoCD applies the new cluster via regional clusters application
- **OCP Clusters**: Hive handles cluster provisioning
- **EKS Clusters**: CAPI (via ACM infrastructure providers) handles cluster provisioning
- ACM imports and manages the new cluster for governance
- Infrastructure providers automatically install required CRDs

### Secret Management
Currently uses manual secret management (Vault integration planned):
```bash
# Retrieve secrets from ACM for each cluster namespace
oc get secret aws-credentials -n $cluster_namespace -o yaml > secrets/aws-credentials.yaml
oc get secret pull-secret -n $cluster_namespace -o yaml > secrets/pull-secret.yaml
```

## ACM GitOps Integration

The project uses ACM's native GitOps integration with infrastructure providers for comprehensive cluster management:

### Infrastructure Provider Components
- **MultiClusterHub**: Configured with infrastructure providers for AWS, Azure, GCP, vSphere, OpenStack, BareMetal
- **CAPI Controllers**: Automatically installed by ACM for enabled infrastructure providers
- **Provider-Specific CRDs**: EKS, AKS, GKE, and other cloud-native cluster resources
- **Unified Management**: Single interface for managing diverse cluster types

### GitOps Integration Components
- **GitOpsCluster CR**: Automatically registers clusters with ArgoCD based on Placement selection
- **ManagedClusterSetBinding**: Binds the global ManagedClusterSet to openshift-gitops namespace
- **Placement**: Selects clusters based on labels (vendor=OpenShift/EKS, cloud=Amazon)
- **Policy**: Automates the creation of GitOps resources across clusters

### Features
- **Automated Cluster Registration**: No manual ArgoCD secret management required
- **ApplicationManager Integration**: KlusterletAddonConfig enables ArgoCD permissions on target clusters
- **Policy-Driven**: ACM policies ensure consistent GitOps configuration across all clusters
- **Multi-Provider Support**: Seamless management of OCP, EKS, AKS, GKE, and other cluster types
- **Infrastructure Provider Lifecycle**: ACM manages CAPI controllers and CRDs automatically

### ACM Configuration Location
The ACM MultiClusterHub configuration is located at:
```
operators/advanced-cluster-management/instance/base/multiclusterhub.yaml
```

Key configuration sections:
- `infrastructureProviders`: Enables AWS, Azure, GCP, vSphere, OpenStack, BareMetal
- `overrides.components`: Configures all ACM components including hypershift
- `availabilityConfig`: Set to High for production deployment

## ArgoCD Architecture and Exclusions

### ArgoCD Deployment Model
- **Centralized GitOps**: Only the hub cluster runs ArgoCD (in `openshift-gitops` namespace)
- **No ArgoCD on Managed Clusters**: Managed clusters don't have ArgoCD installed
- **ApplicationManager**: ACM's ApplicationManager (via `KlusterletAddonConfig`) handles ArgoCD application deployment on managed clusters
- **Single Source of Truth**: All GitOps operations are controlled from the hub cluster

### ArgoCD Resource Exclusions
ArgoCD exclusions are configured **only on the hub cluster** and apply to all managed cluster deployments:

#### Current Exclusions (configured via `prereqs/argocd-tekton-exclusions.yaml`):
- **Tekton TaskRuns**: Excluded to prevent ArgoCD from managing transient pipeline runs
- **ACM-Managed Secrets**: Excluded to prevent ArgoCD from pruning ACM-created secrets
- **Allowed Tekton Resources**: Pipeline and PipelineRun resources are allowed for deployment

#### Exclusion Management:
- **Configuration**: `prereqs/argocd-tekton-exclusions.yaml` - Job that patches ArgoCD CR during bootstrap
- **Manual Script**: `bin/patch-argocd-tekton` - Manual patching script for ArgoCD CR
- **Target**: ArgoCD CR (`openshift-gitops` resource in `openshift-gitops` namespace)
- **Not ConfigMap**: ArgoCD operator manages ConfigMap based on CR spec - always patch the CR

### Key Architecture Points:
1. **Hub-Spoke Model**: Hub cluster ArgoCD deploys to all managed clusters
2. **ACM Integration**: GitOpsCluster CR automatically registers managed clusters with ArgoCD
3. **Unified Exclusions**: Single set of exclusions applies to all managed cluster deployments
4. **No Local ArgoCD**: Managed clusters use ApplicationManager for ArgoCD integration without local ArgoCD instance

## NEWREGION.md Test Plan

An interactive test plan is available at `NEWREGION.md` that guides through creating new regional deployments:
- **Interactive Configuration**: Prompts for cluster type (OCP/EKS), region, compute type, and cluster name
- **Type-Specific Instructions**: Different steps for OCP vs EKS cluster creation
- **Validation Steps**: Includes kustomize build testing and GitOps integration verification
- **Rollback Procedures**: Instructions for cleaning up failed deployments

## Current Cluster Status

### Deployed Clusters
- **cluster-10**: OCP cluster (existing)
- **cluster-20**: OCP cluster (existing) 
- **cluster-30**: OCP cluster (existing)
- **cluster-40**: EKS cluster (implemented, ready for deployment)

## Development Best Practices

- Use `kustomize build` to validate configuration changes
- Test overlays before applying to clusters
- Follow GitOps principles for all cluster modifications
- Reference existing cluster overlays (cluster-10, region-02, region-03) as templates
```