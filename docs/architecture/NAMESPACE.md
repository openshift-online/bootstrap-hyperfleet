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
- `ocp-01-mturansk-t3` - Hive ClusterDeployment, ManagedCluster, ExternalSecrets
- `ocp-01-mturansk-t10` - OpenShift cluster provisioning resources  
- `eks-01-mturansk-t2` - CAPI Cluster, AWSManagedControlPlane, AWSManagedMachinePool

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
- `ocm-ocp-01-mturansk-t3`
- `ocm-ocp-01-mturansk-t10` 
- `ocm-eks-01-mturansk-t2`

**Resources Included**:
- Database services (PostgreSQL instances for AMS, CS, OSL)
- OCM API components
- Cluster configuration management services

#### Pipeline Service Namespaces

**Pattern**: `pipelines-{cluster-name}` (implied from Tekton deployment pattern)  
**Purpose**: Tekton pipelines for infrastructure automation

**Examples**:
- `pipelines-ocp-01-mturansk-t3`
- `pipelines-eks-01-mturansk-t2`

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
   namespace: ocp-01-mturansk-t3
   # Contains: ClusterDeployment, ManagedCluster, etc.
   ```

2. **Service Preparation** (Wave 2)  
   ```yaml
   # Created on hub cluster
   namespace: ocm-ocp-01-mturansk-t3
   # Contains: OCM service definitions
   ```

3. **Managed Cluster Services** (Wave 3+)
   ```yaml
   # Created on managed cluster ocp-01-mturansk-t3
   namespace: ocm-ocp-01-mturansk-t3
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
  path: clusters/ocp-01-mturansk-t3
  destination: https://kubernetes.default.svc  # Hub cluster
  syncWave: "1"
  # Creates namespace: ocp-01-mturansk-t3

# Sync Wave 2+: Service deployment (to managed cluster)  
- component: operators
  path: operators/openshift-pipelines/ocp-01-mturansk-t3
  destination: ocp-01-mturansk-t3  # Managed cluster
  syncWave: "2"  
  # Creates namespace: ocm-ocp-01-mturansk-t3

- component: deployments-ocm
  path: deployments/ocm/ocp-01-mturansk-t3
  destination: ocp-01-mturansk-t3  # Managed cluster  
  syncWave: "4"
  # Deploys to namespace: ocm-ocp-01-mturansk-t3
```

## Current Implementation Examples

### Cluster: `ocp-01-mturansk-t3`

**Hub Cluster Namespaces**:
```bash
# Provisioning namespace
ocp-01-mturansk-t3/
├── ClusterDeployment/ocp-01-mturansk-t3
├── ManagedCluster/ocp-01-mturansk-t3  
├── MachinePool/ocp-01-mturansk-t3-worker
├── ExternalSecret/aws-credentials
└── ExternalSecret/pull-secret

# OCM services namespace  
ocm-ocp-01-mturansk-t3/
├── PostgreSQL databases (AMS, CS, OSL)
├── Service configurations
└── Initialization jobs
```

**Managed Cluster Namespaces** (post-provisioning):
```bash
# OCM services (replicated)
ocm-ocp-01-mturansk-t3/
├── Local OCM services
├── Database connections
└── Cluster-specific configurations

# Pipeline services  
pipelines-ocp-01-mturansk-t3/
├── Tekton Pipeline definitions
├── PipelineRun executions
└── Pipeline ServiceAccounts
```

### Cluster: `eks-01-mturansk-t2`

**Hub Cluster Namespaces**:
```bash
# EKS provisioning namespace
eks-01-mturansk-t2/
├── Cluster/eks-01-mturansk-t2 (CAPI)
├── AWSManagedControlPlane/eks-01-mturansk-t2
├── AWSManagedMachinePool/eks-01-mturansk-t2-workers
├── ManagedCluster/eks-01-mturansk-t2
└── ExternalSecret/aws-credentials

# OCM services namespace
ocm-eks-01-mturansk-t2/
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
oc get clusterdeployment -n ocp-01-mturansk-t3
oc get managedcluster ocp-01-mturansk-t3  
oc get externalsecrets -n ocp-01-mturansk-t3
```

#### Service Namespace Missing on Managed Cluster
```bash
# Verify ArgoCD sync status for service deployments
oc get application -n openshift-gitops | grep ocp-01-mturansk-t3
oc describe application ocp-01-mturansk-t3-operators -n openshift-gitops
```

#### Namespace Permission Issues
```bash
# Check ServiceAccount permissions in service namespaces
oc get sa -n ocm-ocp-01-mturansk-t3
oc describe clusterrolebinding | grep ocp-01-mturansk-t3
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