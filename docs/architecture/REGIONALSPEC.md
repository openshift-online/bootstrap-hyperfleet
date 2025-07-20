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
regions/
├── us-east-1/                 # Region-based organization
│   ├── ocp-02/            # OCP cluster (maintains cluster-XX naming)
│   │   └── region.yaml        # 7 lines - ALL cluster config
│   └── ocp-03/            # Another OCP cluster
│       └── region.yaml        # Simple key-value format
│
├── eu-west-1/                 # Different region
│   └── ocp-05/            # EKS cluster
│       └── region.yaml        # Same simple format for EKS
│
└── ap-southeast-1/            # Asia Pacific region
    └── eks-02/            # EKS cluster
        └── region.yaml        # 6 lines - minimal config
```

## Minimal Regional Cluster Spec

**Single file defines entire cluster** (region.yaml) - **Simplified Format**:

```yaml
# regions/us-east-1/ocp-02/region.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-02
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
  # Minimal compute config
  compute:
    instanceType: m5.xlarge
    replicas: 1
```

```yaml
# regions/us-east-2/eks-01-mturansk-test/region.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: eks-01-mturansk-test
  namespace: us-east-2
spec:
  type: eks
  region: us-east-2
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
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
# regions/us-east-1/ocp-02/workers.yaml (optional)
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
# Convert existing cluster to regional format
./bin/convert-cluster clusters/ocp-02 > regions/us-east-1/ocp-02/region.yaml
```

**Phase 2**: Generate traditional Kustomize resources ✅
```bash  
# Generate cluster overlays from regional specs
./bin/generate-cluster regions/us-east-1/ocp-02/ clusters/ocp-02/
```

**Phase 3**: Parallel operation ✅ ACTIVE
- ✅ Regional specifications implemented in `regions/`
- ✅ Traditional cluster overlays maintained in `clusters/`
- ✅ Converter tools available for bidirectional conversion
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
      
  domain: rosa.mturansk-test.csu2.i3.devshift.org  # Global default
```

## Before/After Comparison

**Before** (ocp-02 traditional):
- 4 files in clusters/ocp-02/: namespace.yaml, kustomization.yaml, install-config.yaml, klusterletaddonconfig.yaml
- Complex install-config with nested OpenShift configuration
- Need to read multiple files to understand cluster config

**After** (ocp-02 regional):  
- 1 file: regions/us-east-1/ocp-02/region.yaml
- 7 lines total with key-value pairs
- All essential configuration visible at once
- Zero patches, zero template hunting

**Example Simplification**:
```yaml
# Before: install-config.yaml (46+ lines)
apiVersion: v1
baseDomain: rosa.mturansk-test.csu2.i3.devshift.org
metadata:
  name: ocp-02
compute:
- name: worker
  platform:
    aws:
      type: m5.xlarge
  replicas: 1
# ... 40+ more lines

# After: region.yaml (19 lines)
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-02
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
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
├── clusters/                # Base templates (replaces regions/templates/)
│   ├── clusterdeployment.yaml
│   ├── managedcluster.yaml
│   ├── machinepool.yaml
│   └── eks/
│       ├── cluster.yaml
│       └── workers.yaml
│
clusters/                    # Traditional cluster overlays
├── ocp-02/              # OCP cluster (Hive-based)
├── ocp-03/              # OCP cluster (Hive-based)
├── ocp-04/              # OCP cluster (Hive-based)
├── eks-02/              # EKS cluster (CAPI-based)
└── ocp-05/              # EKS cluster (CAPI-based)

regions/                     # ✅ IMPLEMENTED: Regional specifications
├── us-east-1/
│   ├── ocp-02/          # Simple regional spec
│   └── ocp-03/
├── us-west-2/
│   └── ocp-04/
├── ap-southeast-1/
│   └── eks-02/
└── eu-west-1/
    └── ocp-05/

deployments/ocm/             # Service deployments per cluster
├── ocp-02/              # OCM services for ocp-02
├── ocp-03/              # OCM services for ocp-03
├── ocp-04/              # OCM services for ocp-04
├── eks-02/              # OCM services for eks-02
└── ocp-05/              # OCM services for ocp-05

gitops-applications/
├── ocp-02.yaml          # Cluster + services ApplicationSet
├── ocp-03.yaml          # Cluster + services ApplicationSet
├── ocp-04.yaml          # Cluster + services ApplicationSet
└── ocp-05.yaml          # Cluster + services ApplicationSet
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
clusters/cluster-XX/
├── namespace.yaml                    # 4 lines
├── install-config.yaml              # 46 lines - direct OpenShift config
├── klusterletaddonconfig.yaml       # 21 lines - ACM config
└── kustomization.yaml               # 10 lines - simple resource list
```
**Total**: ~81 lines (simplified from complex base+patch pattern)

**Regional Specification**:
```
regions/region-name/cluster-XX/
└── region.yaml                      # 19 lines - ALL config
```
**Total**: ~19 lines (77% reduction in complexity)

#### **For EKS Clusters** (Pattern is optimal):
```
clusters/cluster-XX/
├── namespace.yaml                    # 4 lines
├── cluster.yaml                     # 18 lines
├── awsmanagedcontrolplane.yaml      # 18 lines
├── awsmanagedmachinepool.yaml       # 20 lines
├── managedcluster.yaml              # 14 lines
├── klusterletaddonconfig.yaml       # 21 lines
└── kustomization.yaml               # 15 lines
```
**Total**: ~110 lines (already streamlined)

**Regional Specification**:
```
regions/region-name/cluster-XX/
└── region.yaml                      # 19 lines - ALL config
```
**Total**: ~19 lines (83% reduction in complexity)

### Implementation Status ✅

1. **✅ Eliminate Base+Patch Pattern**: OCP clusters simplified to direct configuration files
2. **✅ Regional Specifications**: Implemented simple key-value format in `regions/`
3. **✅ Template-based Generation**: Converter tools (`bin/convert-cluster`, `bin/generate-cluster`) implemented
4. **✅ Minimal File Set**: Regional specs reduced to 1 file per cluster
5. **✅ Base Templates**: Consolidated in `bases/clusters/` following Kustomize conventions

### Current Benefits Achieved

- **Cognitive Simplicity**: Regional specs readable in 10 seconds
- **Bidirectional Conversion**: Can convert between formats as needed
- **Template Reuse**: Base templates in `bases/clusters/` eliminate duplication
- **Regional Organization**: Physical location obvious from directory structure
- **Tool Integration**: Generation tools enable automation and validation

### Related Documentation

- **[Convert Cluster Tool](./bin/convert-cluster.md)** - Convert overlays to regional specs
- **[Generate Cluster Tool](./bin/generate-cluster.md)** - Generate overlays from regional specs
- **[Cluster Creation Guide](./guides/cluster-creation.md)** - End-to-end deployment workflow

The regional cluster specification design has been **successfully implemented** with practical adaptations that maintain the core principles while following established Kustomize conventions.