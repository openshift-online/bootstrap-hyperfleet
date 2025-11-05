# bin/cluster-generate Requirements

## Purpose

The `generate-cluster` tool converts simplified Regional Cluster specifications into complete Kustomize overlay directories ready for GitOps deployment.

## Functional Requirements

### Input/Output Requirements
- **Input**: Cluster specification file (`clusters/{cluster-name}/{cluster-name}.yaml`)
- **Output**: Complete cluster directory structure

### Usage Patterns
```bash
# Generate cluster from specification
./bin/cluster-generate clusters/ocp-01/ocp-01.yaml

# Generate and validate
./bin/cluster-generate clusters/eks-01/eks-01.yaml
kubectl kustomize clusters/eks-01/cluster/

# Generate HCP cluster
./bin/cluster-generate clusters/hcp-01/hcp-01.yaml
```

## Generation Logic Requirements

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

## Generated File Structure Requirements

### OCP Cluster Output
```
clusters/ocp-XX/
├── ocp-XX.yaml                      # Cluster specification
├── cluster/
│   ├── namespace.yaml                    # 4 lines
│   ├── clusterdeployment.yaml           # 25 lines - direct config
│   ├── managedcluster.yaml              # 12 lines - direct config  
│   ├── machinepool.yaml                 # 20 lines - direct config
│   ├── install-config.yaml              # 46 lines - OpenShift config
│   ├── klusterletaddonconfig.yaml       # 21 lines - ACM config
│   └── kustomization.yaml               # 8 lines - simple resource list
├── operators/
├── pipelines/
├── deployments/
└── gitops/
```

### EKS Cluster Output
```
clusters/eks-XX/
├── eks-XX.yaml                      # Cluster specification
├── cluster/
│   ├── namespace.yaml                    # 4 lines
│   ├── cluster.yaml                     # 18 lines - CAPI binding
│   ├── awsmanagedcontrolplane.yaml      # 18 lines - EKS control plane
│   ├── awsmanagedmachinepool.yaml       # 20 lines - EKS workers
│   ├── managedcluster.yaml              # 14 lines - ACM registration
│   ├── klusterletaddonconfig.yaml       # 21 lines - ACM config
│   ├── external-secrets.yaml            # External secrets config
│   ├── acm-integration-pipeline.yaml    # Pipeline integration
│   └── kustomization.yaml               # 15 lines - resource list
├── operators/
├── pipelines/
├── deployments/
└── gitops/

# Note: klusterlet-crd.yaml NOT generated - CRD managed by ACM hub
```

### HCP Cluster Output
```
clusters/hcp-XX/
├── hcp-XX.yaml                      # Cluster specification
├── cluster/
│   ├── namespace.yaml                    # 4 lines
│   ├── hostedcluster.yaml               # 45 lines - HyperShift config
│   ├── ssh-key-secret.yaml              # 8 lines - SSH key secret
│   ├── klusterletaddonconfig.yaml       # 21 lines - ACM config
│   └── kustomization.yaml               # 25 lines - resource list with patches
├── operators/
├── pipelines/
├── deployments/
└── gitops/
```

## Default Values Requirements

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
  domain: bootstrap.red-chesterfield.com
  aws:
    rootVolume:
      size: 100
      type: io1
      iops: 2000
```

## Input Format Requirements

### Required Cluster Specification

```yaml
# clusters/ocp-01/ocp-01.yaml
apiVersion: cluster.openshift.io/v1
kind: ClusterSpec
metadata:
  name: ocp-01
spec:
  type: ocp                           # or 'eks', 'hcp'
  region: us-east-1
  domain: bootstrap.red-chesterfield.com
  
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
# clusters/ocp-01/workers.yaml (optional)
apiVersion: cluster.openshift.io/v1  
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

## Error Handling Requirements

### Generator Errors  
- **Invalid specs**: Schema validation with helpful error messages
- **Missing defaults**: Clear indication of required vs optional fields
- **Template errors**: Detailed context for generation failures

## Validation Requirements

### Integration Tests
```bash
# Convert existing cluster to new spec format
./bin/convert-cluster clusters/ocp-01 > /tmp/cluster.yaml

# Generate back to directory structure
./bin/cluster-generate /tmp/cluster.yaml /tmp/test-cluster/

# Compare outputs
kubectl kustomize clusters/ocp-01/cluster/ > /tmp/original.yaml
kubectl kustomize /tmp/test-cluster/cluster/ > /tmp/generated.yaml
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

## Migration Strategy Requirements

### Phase 1: Validation
```bash
# Convert existing cluster to new format
./bin/convert-cluster clusters/ocp-02 > clusters/ocp-01/ocp-01.yaml

# Generate cluster directory structure
./bin/cluster-generate clusters/ocp-01/ocp-01.yaml

# Compare outputs
diff -r clusters/ocp-02/ clusters/ocp-01/
```

### Phase 2: Parallel Operation
- Keep existing `clusters/{cluster-name}/` flat structure temporarily
- Add new consolidated structure with subdirectories
- Use generator to create new cluster directory structure
- Validate both produce identical results

### Phase 3: Migration
- Convert all existing clusters to semantic naming
- Update GitOps applications to use semantic names
- Remove old cluster-XX numbering scheme

## Benefits Requirements

### Simplified Configuration Management
- **Before**: 7 files, 200+ lines, complex JSON patches
- **After**: 1-2 files, 25 lines total, direct YAML
- **Maintenance**: No patches, no base template hunting
- **Cognitive Load**: All configuration visible in single file

### Consistency
- **Standardized Defaults**: Consistent across all clusters
- **Schema Validation**: Prevent configuration errors
- **Template-Based**: Ensures proper resource generation

## Related Tools

### Prerequisites
- **[new-cluster.md](./new-cluster.md)** - Creates regional specifications that this tool processes

### Alternative Workflows
- **[convert-cluster.md](./convert-cluster.md)** - Converts existing clusters to regional specs

### Bulk Operations
- **[regenerate-all-clusters.md](./regenerate-all-clusters.md)** - Uses this tool for bulk cluster generation

## Design Principles

*This tool enables **configuration as code** - cluster specifications are version-controlled, reviewable, and auditable.*