# Namespace Architecture

## Overview

The OpenShift Bootstrap project implements a **semantic namespace architecture** that provides clear separation between cluster provisioning resources and deployment services while maintaining consistency across hub and managed clusters.

## Core Principles

1. **Semantic Naming**: Namespace names clearly indicate purpose and target cluster
2. **Multi-Cluster Consistency**: Same namespace names across hub and managed clusters  
3. **Service Isolation**: Each cluster's services are isolated by dedicated namespaces
4. **GitOps Integration**: Namespace pattern aligns with ArgoCD sync wave orchestration

## Namespace Patterns

### 1. Cluster Provisioning Namespaces

**Pattern**: `{cluster-name}`  
**Location**: Hub cluster only  
**Purpose**: Contains cluster provisioning CRDs and resources

**Examples**:
- `my-cluster` - Hive ClusterDeployment, ManagedCluster, ExternalSecrets
- `prod-api` - OpenShift cluster provisioning resources  
- `eks-cluster` - CAPI Cluster, AWSManagedControlPlane, AWSManagedMachinePool

**Resources Included**:
```yaml
# OCP Clusters (Hive-based)
- ClusterDeployment (hive.openshift.io)
- MachinePool (hive.openshift.io)  
- ManagedCluster (cluster.open-cluster-management.io)
- KlusterletAddonConfig (agent.open-cluster-management.io)
- ExternalSecret (external-secrets.io) - AWS credentials, pull secrets

# EKS Clusters (CAPI-based)  
- Cluster (cluster.x-k8s.io)
- AWSManagedControlPlane (controlplane.cluster.x-k8s.io)
- AWSManagedMachinePool (infrastructure.cluster.x-k8s.io)
- ManagedCluster (cluster.open-cluster-management.io)
- ExternalSecret (external-secrets.io) - AWS credentials
```

### 2. Deployment Service Namespaces

**Pattern**: `{deployment}-{cluster-name}`  
**Location**: Hub cluster initially, then managed cluster post-provisioning  
**Purpose**: Contains services and applications related to specific clusters

#### OCM Service Namespaces

**Pattern**: `ocm-{cluster-name}`  
**Purpose**: OpenShift Cluster Manager services for cluster lifecycle management

**Examples**:
- `ocm-my-cluster`
- `ocm-prod-api` 
- `ocm-eks-cluster`

**Resources Included**:
- Database services (PostgreSQL instances for AMS, CS, OSL)
- OCM API components
- Cluster configuration management services

#### Pipeline Service Namespaces

**Pattern**: `pipelines-{cluster-name}` (implied from Tekton deployment pattern)  
**Purpose**: Tekton pipelines for infrastructure automation

**Examples**:
- `pipelines-my-cluster`
- `pipelines-eks-cluster`

**Resources Included**:
- Tekton Pipeline definitions
- PipelineRun instances
- Pipeline ServiceAccounts and RBAC
- Pipeline workspaces and persistent volumes

## Multi-Cluster Deployment Flow

### Phase 1: Hub Cluster Provisioning (Sync Wave 1)

**Target**: Hub cluster (`https://kubernetes.default.svc`)
**Namespaces Created**:
```yaml
# Cluster provisioning namespace
{cluster-name}:
  - ClusterDeployment/ManagedCluster resources
  - ExternalSecrets for credentials
  - Hive or CAPI provisioning resources

# Service preparation namespaces  
ocm-{cluster-name}:
  - OCM service configurations
  - Database initialization
  - Service preparation resources
```

### Phase 2: Managed Cluster Deployment (Sync Wave 2+)

**Target**: Managed cluster (`{cluster-name}`)
**Namespaces Replicated**:
```yaml
# Same namespace names, different cluster
ocm-{cluster-name}:
  - OCM services deployed locally
  - Database connections to hub or local instances
  - Cluster-local service configurations

pipelines-{cluster-name}:
  - Tekton operators and pipelines
  - Pipeline execution environments
  - Local CI/CD automation
```

## Namespace Lifecycle

### Creation Order

1. **Cluster Provisioning** (Wave 1)
   ```yaml
   # Created on hub cluster
   namespace: my-cluster
   # Contains: ClusterDeployment, ManagedCluster, etc.
   ```

2. **Service Preparation** (Wave 2)  
   ```yaml
   # Created on hub cluster
   namespace: ocm-my-cluster
   # Contains: OCM service definitions
   ```

3. **Managed Cluster Services** (Wave 3+)
   ```yaml
   # Created on managed cluster my-cluster
   namespace: ocm-my-cluster
   # Contains: Local OCM services, pipelines, applications
   ```

### Namespace Ownership

| Namespace Pattern | Hub Cluster | Managed Cluster | Ownership |
|------------------|-------------|-----------------|-----------|
| `{cluster-name}` | ✅ Provisioning | ❌ Not created | Hub-exclusive |
| `ocm-{cluster-name}` | ✅ Preparation | ✅ Runtime | Dual-deployment |
| `pipelines-{cluster-name}` | ✅ Definitions | ✅ Execution | Dual-deployment |

## ArgoCD Integration

### ApplicationSet Destination Mapping

```yaml
# Sync Wave 1: Cluster provisioning (to hub)
- component: cluster
  path: clusters/my-cluster
  destination: https://kubernetes.default.svc  # Hub cluster
  syncWave: "1"
  # Creates namespace: my-cluster

# Sync Wave 2+: Service deployment (to managed cluster)  
- component: operators
  path: operators/openshift-pipelines/my-cluster
  destination: my-cluster  # Managed cluster
  syncWave: "2"  
  # Creates namespace: ocm-my-cluster

- component: deployments-ocm
  path: deployments/ocm/my-cluster
  destination: my-cluster  # Managed cluster  
  syncWave: "4"
  # Deploys to namespace: ocm-my-cluster
```

## Current Implementation Examples

### Cluster: `my-cluster`

**Hub Cluster Namespaces**:
```bash
# Provisioning namespace
my-cluster/
├── ClusterDeployment/my-cluster
├── ManagedCluster/my-cluster  
├── MachinePool/my-cluster-worker
├── ExternalSecret/aws-credentials
└── ExternalSecret/pull-secret

# OCM services namespace  
ocm-my-cluster/
├── PostgreSQL databases (AMS, CS, OSL)
├── Service configurations
└── Initialization jobs
```

**Managed Cluster Namespaces** (post-provisioning):
```bash
# OCM services (replicated)
ocm-my-cluster/
├── Local OCM services
├── Database connections
└── Cluster-specific configurations

# Pipeline services  
pipelines-my-cluster/
├── Tekton Pipeline definitions
├── PipelineRun executions
└── Pipeline ServiceAccounts
```

### Cluster: `eks-cluster`

**Hub Cluster Namespaces**:
```bash
# EKS provisioning namespace
eks-cluster/
├── Cluster/eks-cluster (CAPI)
├── AWSManagedControlPlane/eks-cluster
├── AWSManagedMachinePool/eks-cluster-workers
├── ManagedCluster/eks-cluster
└── ExternalSecret/aws-credentials

# OCM services namespace
ocm-eks-cluster/
├── EKS-specific OCM configurations  
├── Service mesh integration
└── Cross-cluster networking
```

## Advantages of This Architecture

### 1. **Clear Resource Separation**
- Provisioning resources isolated from runtime services
- Each cluster's services clearly identified by namespace name
- No resource conflicts between different clusters

### 2. **GitOps Orchestration Alignment**  
- Namespace pattern supports ArgoCD sync wave ordering
- Clear distinction between hub-provisioned and cluster-deployed resources
- Predictable deployment destinations for ApplicationSets

### 3. **Multi-Cluster Service Continuity**
- Same namespace names across hub and managed clusters
- Services can reference consistent namespace regardless of execution location
- Simplified configuration templates and service discovery

### 4. **Operational Clarity**
- Namespace name immediately identifies cluster and service purpose
- Easy troubleshooting with semantic naming
- Clear ownership and lifecycle management

### 5. **Scalability**
- Pattern scales to hundreds of clusters without conflicts
- Consistent naming enables automation and tooling
- Clear separation supports independent cluster lifecycle management

## Troubleshooting Guide

### Common Namespace Issues

#### Provisioning Namespace Stuck
```bash
# Check cluster provisioning status
oc get clusterdeployment -n my-cluster
oc get managedcluster my-cluster  
oc get externalsecrets -n my-cluster
```

#### Service Namespace Missing on Managed Cluster
```bash
# Verify ArgoCD sync status for service deployments
oc get application -n openshift-gitops | grep my-cluster
oc describe application my-cluster-operators -n openshift-gitops
```

#### Namespace Permission Issues
```bash
# Check ServiceAccount permissions in service namespaces
oc get sa -n ocm-my-cluster
oc describe clusterrolebinding | grep my-cluster
```

## Related Documentation

- **[Architecture Overview](./ARCHITECTURE.md)** - Complete system architecture
- **[Regional Specifications](./REGIONALSPEC.md)** - Cluster definition patterns
- **[Kustomization Patterns](./KUSTOMIZATION.md)** - Resource templating approaches
- **[Getting Started](../getting-started/first-cluster.md)** - End-to-end cluster creation
- **[Operations Guide](../operations/cluster-management.md)** - Cluster lifecycle management

## References

- **Cluster Examples**: `regions/us-west-2/*/region.yaml`
- **Namespace Templates**: `operators/openshift-pipelines/*/namespace.yaml`
- **ApplicationSet Configs**: `gitops-applications/*.yaml`
- **Deployment Configs**: `deployments/ocm/*/namespace.yaml`