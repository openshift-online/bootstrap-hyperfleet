# Regional Cluster Specification Design

## Core Principle: One Pattern, One Purpose

**Successfully Implemented**: Simple regional specifications eliminate complex Kustomize patches.

## Current Directory Structure

```
bases/
├── clusters/                  # Base templates (Kustomize pattern)
│   ├── clusterdeployment.yaml # Hive ClusterDeployment template
│   ├── managedcluster.yaml    # ACM ManagedCluster template
│   ├── machinepool.yaml       # Hive MachinePool template
│   ├── install-config.yaml    # OpenShift install config template
│   └── eks/                   # EKS cluster templates  
│       ├── cluster.yaml       # CAPI Cluster template
│       └── workers.yaml       # AWSManagedMachinePool template
│
clusters/
├── global/                    # Hub cluster configuration
│   ├── operators/
│   ├── pipelines/
│   └── gitops/
│
├── ocp-02/                    # Managed cluster
│   ├── ocp-02.yaml            # Cluster specification
│   ├── cluster/               # OCP: Hive resources
│   ├── operators/
│   ├── pipelines/
│   ├── deployments/
│   └── gitops/
│
├── ocp-03/                    # Another managed cluster
│   ├── ocp-03.yaml
│   ├── cluster/
│   ├── operators/
│   ├── pipelines/
│   ├── deployments/
│   └── gitops/
│
├── ocp-05/                    # EKS cluster
│   ├── ocp-05.yaml
│   ├── cluster/               # EKS: CAPI resources
│   ├── operators/
│   ├── pipelines/
│   ├── deployments/
│   └── gitops/
│
└── eks-02/                    # Another EKS cluster
    ├── eks-02.yaml
    ├── cluster/
    ├── operators/
    ├── pipelines/
    ├── deployments/
    └── gitops/
```

## Minimal Regional Cluster Spec

**Single file defines entire cluster** ({cluster-name}.yaml) - **Simplified Format**:

```yaml
# clusters/ocp-02/ocp-02.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-02
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: bootstrap.red-chesterfield.com
  
  # Minimal compute config
  compute:
    instanceType: m5.xlarge
    replicas: 1
```

```yaml
# clusters/eks-01-mturansk-test/eks-01-mturansk-test.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: eks-01-mturansk-test
  namespace: us-east-2
spec:
  type: eks
  region: us-east-2
  domain: bootstrap.red-chesterfield.com
  
  # Minimal compute config
  compute:
    instanceType: m5.large
    replicas: 3
    
  # Type-specific configuration
  kubernetes:
    version: "1.28"
```

**Workers file only when different from defaults** (currently not implemented, but reserved):

```yaml  
# clusters/ocp-02/workers.yaml (optional)
instanceType: m5.2xlarge           # Different from default
replicas: 5                        # Different from default
minSize: 2
maxSize: 10
```

## Benefits of This Design

1. **Single Source of Truth**: One file per cluster with all essential config
2. **Human Readable**: No JSON patches, no base template hunting
3. **Region-Organized**: Physical location obvious from directory structure  
4. **Type Agnostic**: OCP and EKS use same specification format
5. **Minimal**: Only specify what's different from sensible defaults
6. **Discoverable**: `ls regions/` shows all regions, `ls regions/us-east-1/` shows all clusters

## Implementation Strategy ✅ COMPLETED

**Phase 1**: Create converter tool ✅
```bash
# Convert existing cluster to new format
./bin/convert-cluster clusters/ocp-02 > clusters/ocp-02/ocp-02.yaml
```

**Phase 2**: Generate traditional Kustomize resources ✅
```bash  
# Generate cluster overlays from cluster specs
./bin/cluster-generate clusters/ocp-02/ocp-02.yaml clusters/ocp-02/cluster/
```

**Phase 3**: Consolidated structure ✅ ACTIVE
- ✅ Cluster specifications in `clusters/{cluster-name}/{cluster-name}.yaml`
- ✅ Cluster provisioning resources in `clusters/{cluster-name}/cluster/`
- ✅ Converter tools available for generation
- ✅ Templates maintained in `bases/clusters/`

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
      
  domain: bootstrap.red-chesterfield.com  # Global default
```

## Before/After Comparison

**Before** (ocp-02 traditional):
- 4 files in clusters/ocp-02/cluster/: namespace.yaml, kustomization.yaml, install-config.yaml, klusterletaddonconfig.yaml
- Complex install-config with nested OpenShift configuration
- Need to read multiple files to understand cluster config

**After** (ocp-02 consolidated):  
- 1 file: clusters/ocp-02/ocp-02.yaml
- 7 lines total with key-value pairs
- All essential configuration visible at once
- Zero patches, zero template hunting

**Example Simplification**:
```yaml
# Before: install-config.yaml (46+ lines)
apiVersion: v1
baseDomain: bootstrap.red-chesterfield.com
metadata:
  name: ocp-02
compute:
- name: worker
  platform:
    aws:
      type: m5.xlarge
  replicas: 1
# ... 40+ more lines

# After: ocp-02.yaml (19 lines)
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-02
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: bootstrap.red-chesterfield.com
  
  # Minimal compute config
  compute:
    instanceType: m5.xlarge
    replicas: 1
```

This design prioritizes **cognitive simplicity** - a developer can understand any cluster in 10 seconds by reading one file.

## Current Structure Analysis

### Directory Structure Overview

The repository successfully implements **regional cluster specifications** alongside traditional Kustomize patterns:

```
bases/
├── clusters/                # Base templates
│   ├── clusterdeployment.yaml
│   ├── managedcluster.yaml
│   ├── machinepool.yaml
│   └── eks/
│       ├── cluster.yaml
│       └── workers.yaml
│
clusters/
├── global/                  # Hub cluster configuration
│   ├── operators/
│   ├── pipelines/
│   └── gitops/
│
├── ocp-02/                  # Managed cluster
│   ├── ocp-02.yaml          # Cluster specification
│   ├── cluster/             # OCP: Hive-based provisioning resources
│   ├── operators/           # Cluster-specific operators
│   ├── pipelines/           # Cluster-specific pipelines
│   ├── deployments/         # OCM services for ocp-02
│   └── gitops/              # Cluster ApplicationSets
│
├── ocp-03/
│   ├── ocp-03.yaml
│   ├── cluster/
│   ├── operators/
│   ├── pipelines/
│   ├── deployments/
│   └── gitops/
│
├── ocp-04/
│   ├── ocp-04.yaml
│   ├── cluster/
│   ├── operators/
│   ├── pipelines/
│   ├── deployments/
│   └── gitops/
│
├── eks-02/                  # EKS cluster (CAPI-based)
│   ├── eks-02.yaml
│   ├── cluster/
│   ├── operators/
│   ├── pipelines/
│   ├── deployments/
│   └── gitops/
│
└── ocp-05/
    ├── ocp-05.yaml
    ├── cluster/
    ├── operators/
    ├── pipelines/
    ├── deployments/
    └── gitops/
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
# Current complex approach (ocp-02)
patches:
  - target:
      kind: ClusterDeployment
    patch: |
      - op: replace
        path: /metadata/namespace
        value: ocp-02
      - op: replace
        path: /metadata/name
        value: ocp-02
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

#### **2. Consolidate Base Templates** ✅ IMPLEMENTED
Minimal, parameterizable base templates are implemented:

```yaml
# Current implementation
bases/clusters/              # ✅ Base templates implemented
├── clusterdeployment.yaml  # Hive ClusterDeployment template
├── managedcluster.yaml     # ACM ManagedCluster template
├── machinepool.yaml        # Hive MachinePool template
├── install-config.yaml     # OpenShift install config template
└── eks/                    # ✅ EKS templates separated
    ├── cluster.yaml        # CAPI Cluster template
    └── workers.yaml        # AWSManagedMachinePool template
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

#### **For OCP Clusters** ✅ SIMPLIFIED:
```
clusters/cluster-XX/cluster/
├── namespace.yaml                    # 4 lines
├── install-config.yaml              # 46 lines - direct OpenShift config
├── klusterletaddonconfig.yaml       # 21 lines - ACM config
└── kustomization.yaml               # 10 lines - simple resource list
```
**Total**: ~81 lines (simplified from complex base+patch pattern)

**Cluster Specification**:
```
clusters/cluster-XX/
└── cluster-XX.yaml                  # 19 lines - ALL config
```
**Total**: ~19 lines (77% reduction in complexity)

#### **For EKS Clusters** (Pattern is optimal):
```
clusters/cluster-XX/cluster/
├── namespace.yaml                    # 4 lines
├── cluster.yaml                     # 18 lines
├── awsmanagedcontrolplane.yaml      # 18 lines
├── awsmanagedmachinepool.yaml       # 20 lines
├── managedcluster.yaml              # 14 lines
├── klusterletaddonconfig.yaml       # 21 lines
└── kustomization.yaml               # 15 lines
```
**Total**: ~110 lines (already streamlined)

**Cluster Specification**:
```
clusters/cluster-XX/
└── cluster-XX.yaml                  # 19 lines - ALL config
```
**Total**: ~19 lines (83% reduction in complexity)

### Implementation Status ✅

1. **✅ Eliminate Base+Patch Pattern**: OCP clusters simplified to direct configuration files
2. **✅ Cluster Specifications**: Implemented simple key-value format in `clusters/{cluster-name}/{cluster-name}.yaml`
3. **✅ Template-based Generation**: Converter tools (`bin/convert-cluster`, `bin/cluster-generate`) implemented
4. **✅ Minimal File Set**: Cluster specs reduced to 1 file per cluster
5. **✅ Base Templates**: Consolidated in `bases/clusters/` following Kustomize conventions

### Current Benefits Achieved

- **Cognitive Simplicity**: Cluster specs readable in 10 seconds
- **Bidirectional Conversion**: Can convert between formats as needed
- **Template Reuse**: Base templates in `bases/clusters/` eliminate duplication
- **Consolidated Organization**: All cluster resources in one location
- **Tool Integration**: Generation tools enable automation and validation

### Related Documentation

- **[Convert Cluster Tool](./bin/convert-cluster.md)** - Convert overlays to cluster specs
- **[Generate Cluster Tool](./bin/cluster-generate.md)** - Generate overlays from cluster specs
- **[Cluster Creation Guide](./guides/cluster-creation.md)** - End-to-end deployment workflow

The cluster specification design has been **successfully implemented** with practical adaptations that maintain the core principles while following established Kustomize conventions.