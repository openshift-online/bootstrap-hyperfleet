# Kustomization Structure Analysis

## Overview

This document provides a comprehensive analysis of the Kustomize-based GitOps infrastructure in the OpenShift bootstrap project. The project implements a hub-spoke architecture for managing OpenShift clusters using ArgoCD, ACM, and Tekton Pipelines.

## Directory Structure

```
bootstrap/
├── gitops-applications/          # ArgoCD Applications (Hub cluster)
├── clusters/                     # Cluster provisioning manifests
│   ├── base/                     # Base cluster templates
│   └── overlay/                  # Cluster-specific configurations
├── regional-pipelines/           # Tekton Pipelines per cluster
│   ├── base/                     # Base pipeline definitions
│   └── overlays/                 # Cluster-specific pipeline configs
├── regional-deployments/         # Regional service deployments
│   ├── base/                     # Base service configurations
│   └── overlays/                 # Cluster-specific service configs
└── operators/                    # Operator deployments
    ├── advanced-cluster-management/
    └── openshift-pipelines/
```

## GitOps Applications Layer

**Location**: `gitops-applications/`

### Main Kustomization

The root kustomization (`gitops-applications/kustomization.yaml`) manages:

- **Namespace**: `openshift-gitops` (hub cluster)
- **Core Applications**:
  - Advanced Cluster Management (ACM)
  - OpenShift Pipelines Operator
  - ACM GitOps Integration
- **Cluster-Specific Applications** (per cluster):
  - Cluster provisioning (sync-wave: 1)
  - Pipeline deployment (sync-wave: 2) 
  - Service deployment (sync-wave: 3)

### Sync Wave Strategy

```yaml
# gitops-applications/cluster-10.cluster.yaml
annotations:
  argocd.argoproj.io/sync-wave: "1"    # Cluster provisioning first

# gitops-applications/cluster-10.pipelines.yaml  
annotations:
  argocd.argoproj.io/sync-wave: "2"    # Pipelines after cluster ready

# gitops-applications/cluster-10.deployments.yaml
annotations:
  argocd.argoproj.io/sync-wave: "3"    # Services after pipelines ready
```

### Application Targeting

- **Hub Cluster**: `destination.name: in-cluster`
- **Managed Clusters**: `destination.server: https://api.cluster-X.domain`

## Cluster Provisioning Layer

**Location**: `clusters/`

### Base Templates

`clusters/base/kustomization.yaml` provides:
- ClusterDeployment (Hive)
- ManagedCluster (ACM)
- MachinePool (Hive)

### Overlay Pattern

Each cluster overlay (`clusters/overlay/cluster-X/`) includes:

1. **Namespace**: Cluster-specific namespace
2. **Install Config**: OpenShift installation configuration
3. **KlusterletAddonConfig**: ACM agent configuration
4. **Patches**: Cluster-specific customizations

**Example**: `clusters/overlay/cluster-10/kustomization.yaml`

```yaml
resources:
  - namespace.yaml
  - klusterletaddonconfig.yaml  
  - ../../base

secretGenerator:
  - name: install-config
    namespace: cluster-10
    files:
      - install-config.yaml

patches:
  - target:
      kind: ClusterDeployment
    patch: |
      - op: replace
        path: /metadata/namespace
        value: cluster-10
```

## Regional Pipelines Layer

**Location**: `regional-pipelines/`

### Base Configuration

`regional-pipelines/base/kustomization.yaml`:
- References OpenShift Pipelines operator (pipelines-operator-only overlay)
- Includes base Pipeline definitions
- Common annotations for versioning

### Overlay Pattern

Each overlay (`regional-pipelines/overlays/cluster-X/`) provides:
- **Namespace**: Cluster-specific pipeline namespace (`ocm-cluster-X`)
- **Cluster Annotations**: Cluster type and metadata
- **Pipeline Resources**: Cluster-specific PipelineRun configurations

**Key Features**:
- Operator deployment to `openshift-operators` namespace
- Pipeline resources deployed to cluster-specific namespaces
- Separation prevents SharedResourceWarning conflicts

## Regional Deployments Layer

**Location**: `regional-deployments/`

### Base Services

`regional-deployments/base/kustomization.yaml` includes:
- Database services (AMS, CS, OSL)
- SecretGenerator for database credentials
- Common service templates

### Overlay Pattern

Each overlay (`regional-deployments/overlays/cluster-X/`) provides:
- **Namespace**: Cluster-specific service namespace (`ocm-cluster-X`)
- **Base Reference**: Inherits from base services
- **Local Resources**: Cluster-specific configurations

## Operator Management

### Advanced Cluster Management

**Structure**: `operators/advanced-cluster-management/`
- **Operator**: Version-specific overlays (2.8 - 2.13)
- **Instance**: MultiClusterHub configuration
- **Observability**: Optional observability features

### OpenShift Pipelines

**Structure**: `operators/openshift-pipelines/operator/`
- **Base**: Core subscription
- **Components**: Console plugin enablement
- **Overlays**: 
  - Version-specific (pipelines-1.18)
  - **pipelines-operator-only**: Operator without console plugin

## Key Design Patterns

### 1. Hub-Spoke Architecture

- **Hub Cluster**: Runs ArgoCD, ACM, manages all clusters
- **Spoke Clusters**: Managed clusters receive deployments via GitOps

### 2. Layered Deployment

1. **Infrastructure Layer**: Cluster provisioning via Hive/ACM
2. **Platform Layer**: Operator deployments (Pipelines, monitoring)
3. **Application Layer**: Regional services and databases

### 3. Namespace Isolation

- **Hub namespaces**: `openshift-gitops`, `open-cluster-management`
- **Cluster namespaces**: `cluster-X` (for cluster resources)
- **Service namespaces**: `ocm-cluster-X` (for applications)

### 4. Resource Separation

- **Operator-only overlays**: Prevent resource conflicts
- **Base + Overlay pattern**: Promotes reusability
- **Patch-based customization**: Minimal duplication

## Identified Issues and Improvements

### Current Issues

#### 1. Hardcoded Cluster URLs
**Problem**: ArgoCD Application manifests contain static server URLs like `https://api.cluster-10.rosa.mturansk-test.csu2.i3.devshift.org:6443`

**Explanation**: Each cluster application hardcodes its destination server URL, requiring manual updates when:
- Adding new clusters
- Changing cluster domains
- Migrating to different regions

**Impact**: 
- Manual scaling process for new clusters
- Error-prone configuration updates
- Tight coupling between cluster infrastructure and GitOps applications

#### 2. Commented Resources
**Problem**: Multiple commented cluster applications in `gitops-applications/kustomization.yaml`:
```yaml
#- ./regional-clusters.cluster-40.application.yaml
#- ./regional-deployments.cluster-40.application.yaml
```

**Explanation**: Commented resources indicate manual cluster activation/deactivation rather than automated lifecycle management. This suggests:
- Incomplete deployment automation
- Manual intervention required for cluster provisioning
- Inconsistent cluster state management

**Impact**:
- Incomplete automation pipeline
- Manual operational overhead
- Risk of configuration drift

#### 3. Secret Management
**Problem**: Database passwords hardcoded in `regional-deployments/base/kustomization.yaml`:
```yaml
secretGenerator:
  - name: ams-db
    literals:
      - db.password="foobar"
```

**Explanation**: Secrets are stored as plain text in the Git repository, violating security best practices:
- Passwords visible in version control
- No secret rotation capability
- Shared secrets across environments

**Impact**:
- **Critical security vulnerability**
- Compliance violations
- Operational risk from credential exposure

#### 4. Duplicate Patch Logic
**Problem**: Repetitive JSON patches across cluster overlays in `clusters/overlay/cluster-X/kustomization.yaml`:
```yaml
patches:
  - target:
      kind: ClusterDeployment
    patch: |
      - op: replace
        path: /metadata/namespace
        value: cluster-10
      - op: replace
        path: /metadata/name
        value: cluster-10
```

**Explanation**: Each cluster overlay contains nearly identical patch operations with only cluster names differing:
- Same patch structure repeated for ClusterDeployment, ManagedCluster, MachinePool
- Manual copy-paste pattern for new clusters
- No abstraction for common patterns

**Impact**:
- High maintenance burden
- Error-prone cluster creation
- Inconsistent configurations

#### 5. Missing Validation
**Problem**: No kustomize build validation in CI/CD pipeline

**Explanation**: The project lacks automated validation to ensure:
- Kustomization files are syntactically correct
- Generated manifests are valid Kubernetes resources
- Overlays properly reference base resources

**Impact**:
- Runtime failures during deployment
- Broken clusters due to invalid configurations
- No early feedback on configuration errors

### Recommended Improvements

#### 1. Dynamic Cluster Registration
**Solution**: Replace hardcoded cluster URLs with ACM GitOpsCluster for automatic registration

**Implementation**:
```yaml
# Use ACM GitOpsCluster for automatic registration
apiVersion: apps.open-cluster-management.io/v1beta1
kind: GitOpsCluster
metadata:
  name: gitops-cluster-registration
  namespace: openshift-gitops
spec:
  argoServer:
    cluster: local-cluster
    argoNamespace: openshift-gitops
  placementRef:
    kind: Placement
    name: all-openshift-clusters
```

**Benefits**:
- Automatic cluster discovery and registration
- Eliminates manual URL management
- Scales automatically with new clusters
- Reduces configuration errors

#### 2. Templated Applications
**Solution**: Use ApplicationSet to generate cluster applications from templates

**Implementation**:
```yaml
# Use ApplicationSet for scalable cluster applications
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: regional-clusters
spec:
  generators:
  - clusters:
      selector:
        matchLabels:
          vendor: OpenShift
  template:
    metadata:
      name: 'regional-cluster-{{name}}'
    spec:
      source:
        path: 'clusters/overlay/{{name}}'
```

**Benefits**:
- Eliminates commented resources
- Automatic application generation for new clusters
- Consistent application configurations
- Reduces manual intervention

#### 3. External Secret Management
**Solution**: Implement External Secrets Operator for secure credential management

**Implementation**:
```yaml
# Use External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "clusters"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "bootstrap-reader"
```

**Benefits**:
- Removes secrets from Git repository
- Enables secret rotation
- Centralized credential management
- Compliance with security standards

#### 4. Kustomize Components
**Solution**: Use Kustomize components to eliminate duplicate patch logic

**Implementation**:
```yaml
# Create reusable component: components/cluster-base/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component

resources:
  - cluster-patches.yaml

replacements:
  - source:
      kind: ConfigMap
      name: cluster-info
      fieldPath: data.clusterName
    targets:
      - select:
          kind: ClusterDeployment
        fieldPaths:
          - metadata.name
          - metadata.namespace
          - spec.clusterName
```

**Benefits**:
- Eliminates duplicate patch code
- Parameterized cluster creation
- Consistent configurations
- Easier maintenance

#### 5. CI/CD Validation
**Solution**: Add comprehensive validation pipeline for Kustomize configurations

**Implementation**:
```yaml
# Add validation step to CI: .github/workflows/validate.yml
name: validate-kustomization
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
    - name: Validate Kustomizations
      run: |
        # Validate all cluster overlays
        for overlay in clusters/overlay/*/; do
          echo "Validating $overlay"
          kustomize build "$overlay" --dry-run
        done
        
        # Validate pipeline overlays
        for overlay in regional-pipelines/overlays/*/; do
          echo "Validating $overlay"
          kustomize build "$overlay" --dry-run
        done
        
        # Validate deployment overlays
        for overlay in regional-deployments/overlays/*/; do
          echo "Validating $overlay"
          kustomize build "$overlay" --dry-run
        done
```

**Benefits**:
- Early detection of configuration errors
- Prevents runtime failures
- Automated quality assurance
- Faster feedback loop for developers

### Scaling Recommendations

1. **Cluster Factory Pattern**: 
   - Create cluster templates with parameter substitution
   - Use Kustomize replacements for cluster-specific values

2. **Policy-Based Management**:
   - Implement ACM policies for cluster compliance
   - Use Gatekeeper for resource validation

3. **Multi-Tenancy**:
   - Separate namespaces per team/environment
   - Implement RBAC boundaries

4. **Observability**:
   - Add OpenTelemetry for deployment tracking
   - Implement GitOps metrics collection

## Current Cluster Status

### Deployed Clusters (OpenShift)
- **cluster-10**: Active (us-east-1)
- **cluster-20**: Active (region-02)
- **cluster-30**: Active (region-03)

### Architecture Benefits

1. **Declarative Management**: All cluster state in Git
2. **Automated Provisioning**: Hive + ACM integration
3. **Pipeline Integration**: Tekton workflows per cluster
4. **Centralized GitOps**: Single ArgoCD instance manages all
5. **Observability**: ACM provides multi-cluster monitoring

## Validation Commands

```bash
# Validate cluster overlays
kustomize build clusters/overlay/cluster-10/
kustomize build clusters/overlay/cluster-20/
kustomize build clusters/overlay/cluster-30/

# Validate pipeline overlays
kustomize build regional-pipelines/overlays/cluster-10/
kustomize build regional-pipelines/overlays/cluster-20/
kustomize build regional-pipelines/overlays/cluster-30/

# Validate deployment overlays
kustomize build regional-deployments/overlays/cluster-10/
kustomize build regional-deployments/overlays/cluster-20/
kustomize build regional-deployments/overlays/cluster-30/

# Dry-run validation
oc --dry-run=client apply -k clusters/overlay/cluster-10/
oc --dry-run=client apply -k regional-pipelines/overlays/cluster-10/
oc --dry-run=client apply -k regional-deployments/overlays/cluster-10/
```

This analysis provides a foundation for understanding the current Kustomize structure and implementing improvements for better scalability, security, and maintainability.