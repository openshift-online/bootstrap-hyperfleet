# GitOps and Kustomize Architecture

## Overview

This document describes the current GitOps and Kustomize patterns used in the OpenShift Bootstrap repository. The architecture follows a **hub-spoke model** with **Application-level sync wave orchestration** and **declarative cluster lifecycle management**.

## Current Directory Structure

```
bootstrap/
├── clusters/
│   ├── global/                     # Hub cluster configuration
│   │   ├── operators/              # Hub cluster operators
│   │   │   ├── openshift-gitops/   # GitOps operator installation
│   │   │   ├── advanced-cluster-management/  # ACM operator and hub
│   │   │   ├── openshift-pipelines/# Pipelines operator (hub)
│   │   │   ├── vault/              # Vault deployment
│   │   │   └── external-secrets/   # ESO deployment
│   │   ├── pipelines/              # Hub cluster pipelines
│   │   │   ├── cluster-bootstrap/  # Cluster preparation pipelines
│   │   │   └── hub-provisioner/    # Cluster creation workflows
│   │   └── gitops/                 # Hub cluster GitOps applications
│   │       ├── openshift-gitops/   # Self-managing GitOps
│   │       ├── advanced-cluster-management/  # ACM ApplicationSet
│   │       ├── openshift-pipelines-operator/ # Tekton operator
│   │       ├── vault/              # Secret management
│   │       ├── eso/                # External Secrets Operator
│   │       └── cluster-bootstrap/  # Bootstrap coordination
│   │
│   ├── my-cluster/                 # Managed cluster
│   │   ├── my-cluster.yaml         # Cluster specification
│   │   ├── cluster/                # OCP cluster provisioning resources
│   │   ├── operators/              # Cluster-specific operators
│   │   ├── pipelines/              # Cluster-specific pipelines
│   │   ├── deployments/            # Cluster-specific deployments (OCM services)
│   │   └── gitops/                 # Cluster-specific GitOps applications
│   │
│   └── eks-cluster/                # Another managed cluster
│       ├── eks-cluster.yaml        # Cluster specification
│       ├── cluster/                # EKS cluster provisioning resources
│       ├── operators/
│       ├── pipelines/
│       ├── deployments/
│       └── gitops/
│
└── bases/                          # Reusable Kustomize templates
    ├── clusters/                   # Cluster provisioning templates
    └── pipelines/                  # Pipeline templates
```

## GitOps Applications Layer

### Main Kustomization

The root GitOps kustomization (`clusters/global/gitops/kustomization.yaml`) manages hub cluster applications:

**Key Applications:**
- **OpenShift GitOps** (Wave -1): Self-managing ArgoCD
- **OpenShift Pipelines** (Wave 1): Tekton operator
- **Vault + ESO** (Wave 2): Secret management
- **ACM ApplicationSet** (Wave 3): Multi-cluster management with internal ordering
- **GitOps Integration** (Wave 4): Cluster integration and metrics
- **Hub Provisioner** (Wave 5): Cluster lifecycle automation pipelines

### Application-Level Sync Wave Strategy

```yaml
# Example: OpenShift GitOps (self-managing)
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: openshift-gitops
  annotations:
    argocd.argoproj.io/sync-wave: "-1"
spec:
  destination:
    name: in-cluster
  source:
    path: clusters/global/operators/openshift-gitops

# Example: ACM ApplicationSet (ordered internal waves)
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: advanced-cluster-management-set
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  generators:
  - list:
      elements:
      - component: acm-operator
        syncWave: "2"    # Internal ordering
      - component: acm-hub  
        syncWave: "3"    # Internal ordering
      - component: acm-policies
        syncWave: "4"    # Internal ordering
```

### Cluster ApplicationSets

Cluster-specific ApplicationSets manage cluster lifecycle:

```yaml
# Example: Provisioning ApplicationSet
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: ocp-456-provisioning
  annotations:
    argocd.argoproj.io/sync-wave: "10"
spec:
  generators:
  - list:
      elements:
      - component: cluster
        path: clusters/ocp-456/cluster
  template:
    spec:
      source:
        repoURL: 'https://github.com/openshift-online/bootstrap-hyperfleet'
        path: '{{path}}'
```

## Cluster Provisioning Layer

### Cluster Specification to Overlay Generation

**Input**: Simple cluster specifications
```yaml
# clusters/my-cluster/my-cluster.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: my-cluster
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: bootstrap.red-chesterfield.com
  compute:
    instanceType: m5.xlarge
    replicas: 2
```

**Output**: Generated cluster overlay
```
clusters/my-cluster/cluster/
├── namespace.yaml                 # Cluster namespace
├── install-config.yaml           # OpenShift installation config
├── klusterletaddonconfig.yaml    # ACM agent configuration
└── kustomization.yaml            # Resource list (no patches)
```

### Base Templates

`bases/clusters/kustomization.yaml` provides reusable templates:
- ClusterDeployment (Hive)
- ManagedCluster (ACM)
- MachinePool (Hive)
- InstallConfig (OpenShift)
- EKS-specific resources (CAPI)

### Simplified Overlay Pattern

**Current approach** eliminates complex JSON patches in favor of direct configuration:

```yaml
# clusters/my-cluster/cluster/kustomization.yaml (simplified)
resources:
  - namespace.yaml
  - install-config.yaml  
  - klusterletaddonconfig.yaml
  - ../../../bases/clusters

# No patches - configuration is direct
```

## Operator Management Layer

### Hub Cluster Operators

**Structure**: `clusters/global/operators/{operator-name}/`
- **OpenShift GitOps**: Self-managing operator installation
- **ACM**: ApplicationSet with ordered deployment (Operator → Hub → Policies)
- **Vault**: Secret management deployment
- **ESO**: External secret synchronization
- **Pipelines**: Tekton operator for hub cluster

### Operator Deployment Pattern

Each operator follows consistent structure:
```
clusters/global/operators/{operator-name}/
├── operator/                      # Operator installation
│   ├── namespace.yaml
│   └── subscription.yaml
├── configuration/                 # Operator configuration
│   └── instance.yaml
└── kustomization.yaml            # Resource aggregation
```

## Pipeline Management Layer

### Base Pipeline Templates

`bases/pipelines/` contains reusable Tekton pipeline definitions:
- **Cluster Bootstrap**: Automated cluster preparation
- **Hub Provisioner**: Centralized cluster creation workflows

### Pipeline Deployment Pattern

Pipelines are deployed per target:
```
clusters/global/pipelines/{pipeline-name}/
├── {pipeline-name}.pipeline.yaml
├── {pipeline-name}.pipelinerun.yaml
└── kustomization.yaml
```

## Secret Management Integration

### Vault + External Secrets Operator

**Architecture**:
1. **Vault** (Wave 2): Secure credential storage
2. **ESO** (Wave 2): Automatic secret synchronization  
3. **ExternalSecret** resources: Sync specific secrets to cluster namespaces

**Example ExternalSecret**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: aws-credentials
  namespace: my-cluster
spec:
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: aws-credentials
    creationPolicy: Owner
  data:
  - secretKey: aws-access-key-id
    remoteRef:
      key: secret/aws-credentials
      property: aws-access-key-id
```

## Design Patterns and Benefits

### 1. **Application-Level Orchestration**
- **Sync waves at Application level**: Clear dependency ordering
- **No resource-level sync waves**: Simplifies individual resources
- **ApplicationSet for complex deployments**: ACM uses internal wave ordering

### 2. **Cluster Specification Simplicity**
- **Single file per cluster**: All configuration in {cluster-name}.yaml
- **Auto-generation**: Complex overlays generated from simple specs
- **Template reuse**: Base templates eliminate duplication

### 3. **GitOps Cluster Management**
- **Declarative**: All cluster config in Git
- **ArgoCD-driven**: Continuous reconciliation
- **Reuse-friendly**: Same base repo, cluster-specific configurations

### 4. **Secure Secret Management**
- **No secrets in Git**: All credentials via Vault + ESO
- **Automatic synchronization**: Secrets available where needed
- **Credential rotation**: Vault enables secret rotation

### 5. **Consistent Operator Management**
- **Semantic naming**: `clusters/{target}/operators/{operator-name}` pattern
- **Ordered deployment**: ApplicationSets handle complex dependencies
- **Self-managing**: GitOps operator manages itself

## Validation and Testing

### Kustomize Validation

```bash
# Validate all GitOps applications
oc kustomize clusters/global/gitops/

# Validate cluster overlays
oc kustomize clusters/my-cluster/cluster/

# Validate operator configurations  
oc kustomize clusters/global/operators/advanced-cluster-management/

# Validate pipeline configurations
oc kustomize clusters/global/pipelines/cluster-bootstrap/
```

### Dry-Run Testing

```bash
# Test cluster provisioning without deployment
oc --dry-run=client apply -k clusters/my-cluster/cluster/

# Test operator deployment
oc --dry-run=client apply -k clusters/global/operators/vault/

# Test complete GitOps application deployment
oc --dry-run=client apply -k clusters/global/gitops/
```

## Operational Procedures

### Bootstrap Deployment

```bash
# 1. Install GitOps operator
oc apply -k clusters/global/operators/openshift-gitops

# 2. Deploy all applications with sync wave ordering
oc apply -k clusters/global/gitops/

# 3. Monitor deployment progress
oc get applications -n openshift-gitops
```

### Adding New Clusters

```bash
# 1. Create cluster specification
./bin/cluster-create

# 2. Generate cluster overlay (automatic)
./bin/cluster-generate clusters/new-cluster/new-cluster.yaml

# 3. Commit and deploy via GitOps
git add clusters/
git commit -m "Add new-cluster"
git push origin main
```

### Updating Operator Configurations

```bash
# 1. Modify operator configuration
vim clusters/global/operators/vault/configuration/vault.yaml

# 2. Validate changes
oc kustomize clusters/global/operators/vault/

# 3. Deploy via GitOps
git add clusters/global/operators/vault/
git commit -m "Update Vault configuration"
git push origin main
```

## Migration from Complex Patterns

### Eliminated Patterns

1. **Complex JSON Patches**: Replaced with direct configuration files
2. **Base + Patch Overlays**: Simplified to base + resource lists
3. **Resource-Level Sync Waves**: Moved to Application-level orchestration
4. **Hardcoded Secrets**: Replaced with Vault + ESO integration
5. **Manual ApplicationSet Management**: Automated via cluster generation

### Benefits Achieved

- **77% reduction** in configuration complexity (cluster specs vs overlays)
- **Eliminated patch debugging** through direct configuration
- **Simplified dependency management** via Application sync waves
- **Secure secret handling** with no credentials in Git
- **Self-referential reuse** enabling multi-tenant scenarios

## Related Documentation

- **[Architecture Overview](./ARCHITECTURE.md)** - Complete system architecture
- **[Cluster Specifications](./REGIONALSPEC.md)** - Cluster definition patterns
- **[Namespace Architecture](./NAMESPACE.md)** - Multi-cluster namespace strategy
- **[Bootstrap Walkthrough](../../BOOTSTRAP.md)** - Step-by-step deployment guide
- **[Cluster Creation Guide](../../guides/cluster-creation.md)** - End-to-end workflow

This Kustomize architecture provides a foundation for scalable, secure, and maintainable multi-cluster GitOps operations while maintaining simplicity for day-to-day cluster management.