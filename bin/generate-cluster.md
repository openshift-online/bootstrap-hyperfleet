# Generate Cluster Tool

**Generates complete Kustomize overlay from minimal regional specification**

## Purpose

The `generate-cluster` tool converts simplified Regional Cluster specifications into complete Kustomize overlay directories ready for GitOps deployment.

## Usage

```bash
# Generate cluster overlay from regional spec
./bin/generate-cluster regions/us-east-1/ocp-01/

# Generate and validate
./bin/generate-cluster regions/ap-southeast-1/eks-01/
kubectl kustomize clusters/eks-01/

# Generate HCP cluster
./bin/generate-cluster regions/us-east-1/hcp-01/
```

**Input**: Regional specification directory (`regions/region/cluster/`)
**Output**: Complete Kustomize overlay directory

## Generation Logic

### OCP Cluster Generation

**From Simple to Complex**:
1. Apply defaults to minimal specification
2. Generate `install-config.yaml` with full OpenShift structure
3. Generate direct YAML files (no base+patches approach)
4. Create simple `kustomization.yaml` with resource list

### EKS Cluster Generation

**From Simple to Complex**:
1. Apply defaults to minimal specification
2. Generate `cluster.yaml` (CAPI Cluster binding)
3. Generate `awsmanagedcontrolplane.yaml` (EKS control plane)
4. Generate `awsmanagedmachinepool.yaml` (EKS worker nodes)
5. Generate `managedcluster.yaml` (ACM registration)

## Generated File Structure

### OCP Cluster Output
```
clusters/ocp-XX/
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
clusters/eks-XX/
├── namespace.yaml                    # 4 lines
├── cluster.yaml                     # 18 lines - CAPI binding
├── awsmanagedcontrolplane.yaml      # 18 lines - EKS control plane
├── awsmanagedmachinepool.yaml       # 20 lines - EKS workers
├── managedcluster.yaml              # 14 lines - ACM registration
├── klusterletaddonconfig.yaml       # 21 lines - ACM config
└── kustomization.yaml               # 15 lines - resource list
```

### HCP Cluster Output
```
clusters/hcp-XX/
├── namespace.yaml                    # 4 lines
├── hostedcluster.yaml               # 45 lines - HyperShift config
├── ssh-key-secret.yaml              # 8 lines - SSH key secret
├── klusterletaddonconfig.yaml       # 21 lines - ACM config
└── kustomization.yaml               # 25 lines - resource list with patches
```

## Default Values Applied

The generator applies intelligent defaults to minimize required configuration:

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

## Input Format

### Required Regional Specification

```yaml
# regions/us-east-1/ocp-01/region.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-01
  namespace: us-east-1
spec:
  type: ocp                           # or 'eks', 'hcp'
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
# regions/us-east-1/ocp-01/workers.yaml (optional)
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

## Error Handling

### Generator Errors  
- **Invalid specs**: Schema validation with helpful error messages
- **Missing defaults**: Clear indication of required vs optional fields
- **Template errors**: Detailed context for generation failures

## Validation Testing

### Integration Tests
```bash
# Convert existing cluster to regional spec
./bin/convert-cluster clusters/ocp-01 > /tmp/region.yaml

# Generate back to overlay format
./bin/generate-cluster /tmp/region.yaml /tmp/test-cluster/

# Compare outputs
kubectl kustomize clusters/ocp-01/ > /tmp/original.yaml
kubectl kustomize /tmp/test-cluster/ > /tmp/generated.yaml
diff /tmp/original.yaml /tmp/generated.yaml
```

### ArgoCD Validation
```bash
# Test with ArgoCD dry-run
argocd app create test-cluster \
  --repo https://github.com/your-org/bootstrap \
  --path /tmp/test-cluster \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace default \
  --dry-run
```

## Migration Strategy

### Phase 1: Validation
```bash
# Convert existing cluster to new format
./bin/convert-cluster clusters/ocp-02 > regions/us-east-1/ocp-01/region.yaml

# Generate back to semantic naming
./bin/generate-cluster regions/us-east-1/ocp-01/

# Compare outputs
diff -r clusters/ocp-02/ clusters/ocp-01/
```

### Phase 2: Parallel Operation
- Keep existing `clusters/cluster-XX/` structure temporarily
- Add new `regions/` structure with semantic naming 
- Use generator to create semantic cluster directories
- Validate both produce identical results

### Phase 3: Migration
- Convert all existing clusters to semantic naming
- Update GitOps applications to use semantic names
- Remove old cluster-XX numbering scheme

## Benefits

### Simplified Configuration Management
- **Before**: 7 files, 200+ lines, complex JSON patches
- **After**: 1-2 files, 25 lines total, direct YAML
- **Maintenance**: No patches, no base template hunting
- **Cognitive Load**: All configuration visible in single file

### Consistency
- **Standardized Defaults**: Consistent across all clusters
- **Schema Validation**: Prevent configuration errors
- **Template-Based**: Ensures proper resource generation

## Related Documentation

- **[Convert Cluster Tool](./convert-cluster.md)** - Convert Kustomize overlays to regional specs
- **[Regional Specification](../REGIONALSPEC.md)** - Complete regional cluster specification format
- **[New Cluster Tool](./new-cluster.md)** - Interactive cluster creation wizard
- **[Cluster Creation Guide](../guides/cluster-creation.md)** - End-to-end cluster deployment workflow

---

*This tool enables **configuration as code** - cluster specifications are version-controlled, reviewable, and auditable.*