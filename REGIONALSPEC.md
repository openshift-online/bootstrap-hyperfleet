# Regional Cluster Specification Design

## Core Principle: One Pattern, One Purpose

**Current Problem**: Two different patterns (Hive patches vs direct YAML) create unnecessary complexity.

## Proposed Directory Structure

```
regions/
├── templates/                 # Minimal base templates
│   ├── ocp/                  # OpenShift cluster templates
│   │   ├── cluster.yaml      # ClusterDeployment + ManagedCluster
│   │   └── workers.yaml      # MachinePool
│   └── eks/                  # EKS cluster templates  
│       ├── cluster.yaml      # CAPI Cluster + ManagedCluster
│       └── workers.yaml      # AWSManagedMachinePool + ControlPlane
│
├── us-east-1/                # Region-based organization
│   ├── ocp-prod/             # cluster-10 (simple name)
│   │   ├── region.yaml       # 15 lines - ALL cluster config
│   │   └── workers.yaml      # 10 lines - worker config only
│   └── eks-dev/              # cluster-40 (simple name)  
│       ├── region.yaml       # 12 lines - ALL cluster config
│       └── workers.yaml      # 8 lines - worker config only
│
└── ap-southeast-1/           # Different region
    └── eks-stage/            # cluster-41
        ├── region.yaml
        └── workers.yaml
```

## Minimal Regional Cluster Spec

**Single file defines entire cluster** (region.yaml):

```yaml
# regions/us-east-1/ocp-prod/region.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-prod
  namespace: us-east-1
spec:
  type: ocp                           # or 'eks'
  region: us-east-1
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
  # Minimal compute config
  compute:
    instanceType: m5.xlarge
    replicas: 3
    
  # Only when different from defaults
  kubernetes:
    version: "1.28"                   # EKS only
    
  openshift:                          # OCP only
    version: "4.14"
    channel: stable
```

**Workers file only when different from defaults**:

```yaml  
# regions/us-east-1/ocp-prod/workers.yaml (optional)
apiVersion: regional.openshift.io/v1  
kind: WorkerPool
metadata:
  name: compute
spec:
  instanceType: m5.2xlarge           # Different from default
  replicas: 5                        # Different from default
  scaling:
    min: 2
    max: 10
```

## Benefits of This Design

1. **Single Source of Truth**: One file per cluster with all essential config
2. **Human Readable**: No JSON patches, no base template hunting
3. **Region-Organized**: Physical location obvious from directory structure  
4. **Type Agnostic**: OCP and EKS use same specification format
5. **Minimal**: Only specify what's different from sensible defaults
6. **Discoverable**: `ls regions/` shows all regions, `ls regions/us-east-1/` shows all clusters

## Implementation Strategy

**Phase 1**: Create converter tool
```bash
# Convert existing cluster-10 to new format
./bin/convert-cluster clusters/overlay/cluster-10 > regions/us-east-1/ocp-prod/region.yaml
```

**Phase 2**: Generate traditional Kustomize resources
```bash  
# Generate current format from simple spec
./bin/generate-cluster regions/us-east-1/ocp-prod/ > clusters/overlay/cluster-10/
```

**Phase 3**: Replace current structure once validated

## Default Assumptions

The minimal spec works because we assume smart defaults:

```yaml
# Implicit defaults (not written in files)
defaults:
  ocp:
    version: "4.14"
    channel: stable
    compute:
      instanceType: m5.xlarge
      replicas: 3
      
  eks:  
    version: "1.28"
    compute:
      instanceType: m5.large
      replicas: 3
      scaling: {min: 1, max: 10}
      
  domain: rosa.mturansk-test.csu2.i3.devshift.org  # Global default
```

## Before/After Comparison

**Current** (cluster-10):
- 7 files, 200+ lines
- 84-line kustomization with complex patches
- Need to read 3 files to understand cluster config

**Proposed** (ocp-prod):  
- 1-2 files, 25 lines total
- All configuration visible at once
- Zero patches, zero base template hunting

This design prioritizes **cognitive simplicity** - a developer can understand any cluster in 30 seconds by reading one file.

## Current Structure Analysis

### Directory Structure Overview

The repository uses a **dual-pattern architecture** that separates cluster provisioning from regional service deployments:

```
clusters/
├── base/                     # Hive/OCP cluster templates
│   ├── clusterdeployment.yaml
│   ├── managedcluster.yaml
│   └── machinepool.yaml
└── overlay/
    ├── cluster-10/          # OCP cluster (Hive-based)
    ├── cluster-20/          # OCP cluster (Hive-based)  
    ├── cluster-30/          # OCP cluster (Hive-based)
    └── cluster-40/          # EKS cluster (CAPI-based)

regional-deployments/
├── base/                    # Database services (AMS, CS, OSL)
└── overlays/
    ├── cluster-10/         # Regional services for cluster-10
    ├── cluster-20/         # Regional services for cluster-20
    ├── cluster-30/         # Regional services for cluster-30
    └── cluster-40/         # Regional services for cluster-40

gitops-applications/
├── regional-clusters.cluster-XX.application.yaml      # Cluster provisioning
└── regional-deployments.cluster-XX.application.yaml   # Service deployment
```

### Required vs Optional Files Analysis

#### **Absolutely Required (Core)**
1. **Per Cluster Overlay**:
   - `namespace.yaml` - Cluster namespace
   - `kustomization.yaml` - Kustomize configuration
   - `klusterletaddonconfig.yaml` - ACM cluster management

2. **OCP Clusters (Hive-based)**:
   - `install-config.yaml` - OpenShift installation config
   - Base templates: `clusterdeployment.yaml`, `managedcluster.yaml`, `machinepool.yaml`

3. **EKS Clusters (CAPI-based)**:
   - `awsmanagedcontrolplane.yaml` - EKS control plane
   - `awsmanagedmachinepool.yaml` - EKS worker nodes
   - `cluster.yaml` - CAPI cluster binding
   - `managedcluster.yaml` - ACM cluster registration

4. **GitOps Applications**:
   - `regional-clusters.cluster-XX.application.yaml` - Cluster provisioning
   - `regional-deployments.cluster-XX.application.yaml` - Service deployment

#### **Optional (Convenience/Features)**
- Regional deployments (database services) - Only needed if regional services required
- Additional machine pools
- Custom network configurations
- Extended ACM configurations

### Redundant/Complex Patterns Identified

#### **1. Excessive JSON Patch Complexity**
OCP clusters use 84-line kustomization files with extensive JSON patches for simple name substitutions:

```yaml
# Current complex approach (cluster-10)
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
      # ... 20+ more patches
```

#### **2. Inconsistent Base Template Usage**
- **OCP clusters**: Use shared base templates + patches (complex)
- **EKS clusters**: Use individual YAML files (simple)
- **Result**: Two completely different patterns for same outcome

#### **3. Redundant Namespace Patterns**
- Cluster namespace: `cluster-XX`
- Regional namespace: `ocm-cluster-XX`
- Creates unnecessary separation for regional services

#### **4. Over-Engineering of Regional Deployments**
The regional-deployments structure deploys identical PostgreSQL databases (AMS, CS, OSL) to every cluster:
- Same 3 databases deployed everywhere
- Same credentials (hardcoded "foobar" passwords)
- Minimal customization between clusters

### Simplification Opportunities

#### **1. Standardize on Single Pattern**
**Recommendation**: Adopt EKS-style individual files approach for all clusters
- **Current OCP**: Base + 84-line patches → **Simplified**: Direct YAML files
- **Benefits**: Easier to read, modify, and debug
- **Trade-off**: Slight duplication vs significantly reduced complexity

#### **2. Consolidate Base Templates**
Create minimal, parameterizable base templates:

```yaml
# Simplified base approach
clusters/base/
├── ocp/                    # Hive-based cluster templates
│   ├── clusterdeployment.yaml
│   ├── managedcluster.yaml
│   └── machinepool.yaml
└── eks/                    # CAPI-based cluster templates
    ├── cluster.yaml
    ├── awsmanagedcontrolplane.yaml
    └── awsmanagedmachinepool.yaml
```

#### **3. Eliminate Regional Deployment Redundancy**
**Current**: Every cluster gets identical database stack
**Simplified Options**:
- **Option A**: Single shared regional service per region (not per cluster)
- **Option B**: Make regional deployments truly optional
- **Option C**: Template-based regional services with meaningful differentiation

#### **4. Unified Namespace Strategy**
- Use single namespace pattern: `cluster-XX` for both cluster and regional resources
- Eliminates `ocm-cluster-XX` vs `cluster-XX` confusion

### Minimal Cluster Requirements

#### **For OCP Clusters** (Simplified):
```
clusters/overlay/cluster-XX/
├── namespace.yaml                    # 4 lines
├── clusterdeployment.yaml           # 25 lines - direct config
├── managedcluster.yaml              # 12 lines - direct config  
├── machinepool.yaml                 # 20 lines - direct config
├── install-config.yaml              # 46 lines - OpenShift config
├── klusterletaddonconfig.yaml       # 21 lines - ACM config
└── kustomization.yaml               # 8 lines - simple resource list
```
**Total**: ~136 lines vs current 200+ lines with complex patches

#### **For EKS Clusters** (Current pattern is already optimal):
```
clusters/overlay/cluster-XX/
├── namespace.yaml                    # 4 lines
├── cluster.yaml                     # 18 lines
├── awsmanagedcontrolplane.yaml      # 18 lines
├── awsmanagedmachinepool.yaml       # 20 lines
├── managedcluster.yaml              # 14 lines
├── klusterletaddonconfig.yaml       # 21 lines
└── kustomization.yaml               # 15 lines
```
**Total**: ~110 lines (already streamlined)

### Key Recommendations

1. **Eliminate Base+Patch Pattern**: Move to direct YAML files like EKS clusters
2. **Rationalize Regional Deployments**: Question necessity of per-cluster identical services
3. **Standardize Namespace Strategy**: Single `cluster-XX` pattern
4. **Minimal File Set**: 7 files per cluster maximum
5. **Template-based Generation**: Consider cluster generation tools vs manual duplication

The current structure shows signs of **organic growth** with **two competing patterns** (Hive vs CAPI) that could benefit from **unification and simplification** while maintaining the same functional capabilities.