# Regional Cluster Converter Tools

This document describes the converter and generator tools that implement the Regional Cluster Specification defined in [REGIONALSPEC.md](REGIONALSPEC.md).

## Overview

The tools provide bidirectional conversion between the current complex Kustomize-based cluster configurations and the new simplified Regional Cluster specification format.

## Tools

### 1. Converter Tool (`./bin/convert-cluster`)

**Purpose**: Converts existing complex cluster overlays into minimal regional specifications.

**Usage**:
```bash
# Convert OCP cluster
./bin/convert-cluster clusters/overlay/cluster-10 > regions/us-east-1/ocp-prod/region.yaml

# Convert EKS cluster  
./bin/convert-cluster clusters/overlay/cluster-40 > regions/ap-southeast-1/eks-dev/region.yaml
```

**Input**: Kustomize overlay directory (`clusters/overlay/cluster-XX/`)
**Output**: Minimal regional specification (YAML to stdout)

### 2. Generator Tool (`./bin/generate-cluster`)

**Purpose**: Generates complete Kustomize overlay from minimal regional specification.

**Usage**:
```bash
# Generate cluster overlay from regional spec
./bin/generate-cluster regions/us-east-1/ocp-prod/ clusters/overlay/cluster-10/

# Generate and validate
./bin/generate-cluster regions/ap-southeast-1/eks-dev/ clusters/overlay/cluster-40/
kubectl kustomize clusters/overlay/cluster-40/
```

**Input**: Regional specification directory (`regions/region/cluster/`)
**Output**: Complete Kustomize overlay directory

## Regional Specification Format

### Minimal Cluster Specification

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
    
  # Type-specific configuration
  openshift:                          # OCP only
    version: "4.14"
    channel: stable
    
  kubernetes:                         # EKS only
    version: "1.28"
```

### Optional Worker Pool Configuration

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

## Default Values

The tools assume intelligent defaults to minimize required configuration:

### OCP Defaults
```yaml
defaults:
  ocp:
    version: "4.14"
    channel: stable
    compute:
      instanceType: m5.xlarge
      replicas: 3
    networking:
      networkType: OVNKubernetes
      clusterNetwork: "10.128.0.0/14"
      serviceNetwork: "172.30.0.0/16"
```

### EKS Defaults  
```yaml
defaults:
  eks:
    version: "1.28"
    compute:
      instanceType: m5.large
      replicas: 3
      scaling: {min: 1, max: 10}
    vpc:
      availabilityZoneUsageLimit: 2
      availabilityZoneSelection: Ordered
```

### Global Defaults
```yaml
defaults:
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  aws:
    rootVolume:
      size: 100
      type: io1
      iops: 2000
```

## Conversion Logic

### OCP Cluster Conversion

**From Complex to Simple**:
1. Parse `kustomization.yaml` to identify cluster type
2. Extract cluster name from JSON patches
3. Parse `install-config.yaml` for region, compute, versions
4. Apply defaults and output minimal specification

**From Simple to Complex**:
1. Apply defaults to minimal specification
2. Generate `install-config.yaml` with full OpenShift structure
3. Generate direct YAML files (no base+patches approach)
4. Create simple `kustomization.yaml` with resource list

### EKS Cluster Conversion

**From Complex to Simple**:
1. Parse `awsmanagedcontrolplane.yaml` for region, version, domain
2. Parse `awsmanagedmachinepool.yaml` for compute configuration  
3. Extract cluster name from metadata
4. Apply defaults and output minimal specification

**From Simple to Complex**:
1. Apply defaults to minimal specification
2. Generate `cluster.yaml` (CAPI Cluster binding)
3. Generate `awsmanagedcontrolplane.yaml` (EKS control plane)
4. Generate `awsmanagedmachinepool.yaml` (EKS worker nodes)
5. Generate `managedcluster.yaml` (ACM registration)

## Generated File Structure

### OCP Cluster Output
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

### EKS Cluster Output
```
clusters/overlay/cluster-XX/
├── namespace.yaml                    # 4 lines
├── cluster.yaml                     # 18 lines - CAPI binding
├── awsmanagedcontrolplane.yaml      # 18 lines - EKS control plane
├── awsmanagedmachinepool.yaml       # 20 lines - EKS workers
├── managedcluster.yaml              # 14 lines - ACM registration
├── klusterletaddonconfig.yaml       # 21 lines - ACM config
└── kustomization.yaml               # 15 lines - resource list
```

## Benefits

### Before (Complex)
- **cluster-10**: 7 files, 200+ lines, 84-line kustomization with complex JSON patches
- **Cognitive Load**: Need to read 3+ files to understand cluster configuration
- **Maintenance**: JSON patches are error-prone and hard to debug

### After (Simple)
- **ocp-prod**: 1-2 files, 25 lines total
- **Cognitive Load**: All configuration visible in single file
- **Maintenance**: Direct YAML, no patches, no base template hunting

## Migration Strategy

### Phase 1: Validation
```bash
# Convert existing cluster to new format
./bin/convert-cluster clusters/overlay/cluster-10 > regions/us-east-1/ocp-prod/region.yaml

# Generate back to old format
./bin/generate-cluster regions/us-east-1/ocp-prod/ clusters/overlay/cluster-10-new/

# Compare outputs
diff -r clusters/overlay/cluster-10/ clusters/overlay/cluster-10-new/
```

### Phase 2: Parallel Operation
- Keep existing `clusters/overlay/` structure
- Add new `regions/` structure  
- Use generator to create overlays from regional specs
- Validate both produce identical results

### Phase 3: Migration
- Convert all existing clusters to regional format
- Update GitOps applications to use generator
- Remove old complex overlay structure

## Error Handling

### Converter Errors
- **Missing files**: Graceful degradation with warnings
- **Invalid YAML**: Clear error messages with file/line numbers
- **Unsupported patterns**: Warnings about non-standard configurations

### Generator Errors  
- **Invalid specs**: Schema validation with helpful error messages
- **Missing defaults**: Clear indication of required vs optional fields
- **Template errors**: Detailed context for generation failures

## Testing

### Validation Tests
```bash
# Test conversion round-trip
./bin/convert-cluster clusters/overlay/cluster-10 | \
./bin/generate-cluster - /tmp/test-cluster/
kubectl kustomize /tmp/test-cluster/ > /tmp/generated.yaml

kubectl kustomize clusters/overlay/cluster-10/ > /tmp/original.yaml
diff /tmp/original.yaml /tmp/generated.yaml
```

### Integration Tests
- Convert all existing clusters
- Generate overlays from converted specs
- Validate against existing manifests
- Test with kustomize build
- Test with ArgoCD dry-run

This design prioritizes **cognitive simplicity** - any developer can understand a cluster configuration in 30 seconds by reading one file.