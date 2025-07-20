# bin/convert-cluster Requirements

## Purpose

The `convert-cluster` tool implements bidirectional conversion from the current complex Kustomize-based cluster configurations to the new simplified Regional Cluster specification format.

## Functional Requirements

### Input/Output Requirements
- **Input**: Kustomize overlay directory (`clusters/overlay/cluster-XX/`)
- **Output**: Minimal regional specification (YAML to stdout)

### Usage Patterns
```bash
# Convert OCP cluster
./bin/convert-cluster clusters/overlay/ocp-02 > regions/us-east-1/ocp-prod/region.yaml

# Convert EKS cluster  
./bin/convert-cluster clusters/overlay/eks-02 > regions/ap-southeast-1/eks-dev/region.yaml
```

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

The converter assumes intelligent defaults to minimize required configuration:

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

### EKS Cluster Conversion

**From Complex to Simple**:
1. Parse `awsmanagedcontrolplane.yaml` for region, version, domain
2. Parse `awsmanagedmachinepool.yaml` for compute configuration  
3. Extract cluster name from metadata
4. Apply defaults and output minimal specification

## Error Handling Requirements

### Converter Errors
- **Missing files**: Graceful degradation with warnings
- **Invalid YAML**: Clear error messages with file/line numbers
- **Unsupported patterns**: Warnings about non-standard configurations

## Validation Requirements

```bash
# Test conversion round-trip
./bin/convert-cluster clusters/overlay/ocp-02 | \
./bin/generate-cluster - /tmp/test-cluster/
kubectl kustomize /tmp/test-cluster/ > /tmp/generated.yaml

kubectl kustomize clusters/overlay/ocp-02/ > /tmp/original.yaml
diff /tmp/original.yaml /tmp/generated.yaml
```

## Benefits Requirements

### Before (Complex)
- **ocp-02**: 7 files, 200+ lines, 84-line kustomization with complex JSON patches
- **Cognitive Load**: Need to read 3+ files to understand cluster configuration
- **Maintenance**: JSON patches are error-prone and hard to debug

### After (Simple)
- **ocp-prod**: 1-2 files, 25 lines total
- **Cognitive Load**: All configuration visible in single file
- **Maintenance**: Direct YAML, no patches, no base template hunting

## Related Tools

### Prerequisites
- **[generate-cluster.md](./generate-cluster.md)** - Processes regional specifications created by this conversion tool

### Alternative Workflows
- **[new-cluster.md](./new-cluster.md)** - Creates new regional specifications from scratch

### Bulk Operations
- **[regenerate-all-clusters.md](./regenerate-all-clusters.md)** - Bulk generation from converted specifications

## Design Principles

*This design prioritizes **cognitive simplicity** - any developer can understand a cluster configuration in 30 seconds by reading one file.*