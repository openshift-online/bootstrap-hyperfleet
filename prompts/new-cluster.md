# bin/new-cluster - Interactive Cluster Generator

## Overview
The `bin/new-cluster` tool is an interactive CLI that generates complete OpenShift/EKS cluster configurations for the bootstrap GitOps repository.

## Implementation Status
✅ **COMPLETED** - Tool has been implemented and tested successfully.

## Features

### 1. Interactive Input Collection
The tool prompts for 6 required inputs with validation:

- **Cluster Name** (string, required)
  - Validates uniqueness against existing clusters
  - Checks for existing cluster directories and GitOps applications
- **Type** (string, required)
  - Only accepts "ocp" or "eks"
  - Validates input before proceeding
- **Region** (string, default: "us-west-2")
  - AWS region for cluster deployment
- **Domain** (string, default: "rosa.mturansk-test.csu2.i3.devshift.org")
  - Base domain for cluster endpoints
- **Instance Type** (string, default: "m5.2xlarge")
  - EC2 instance type for cluster nodes
- **Replicas** (int, default: "2")
  - Number of worker nodes

### 2. Regional Specification Generation
Creates `regions/[Region]/[Cluster Name]/region.yaml` with the following template:

```yaml
name: [Cluster Name]
type: [Type]  # "ocp" or "eks"
region: [Region]
domain: [Domain]
instanceType: [Instance Type]
replicas: [Replicas]
```

### 3. Complete Cluster Configuration
Automatically calls `bin/generate-cluster` to create:

- **Cluster manifests**: `clusters/[Cluster Name]/`
  - Namespace, CAPI resources (EKS) or Hive resources (OCP)
  - ManagedCluster and KlusterletAddonConfig
- **Operators**: `operators/openshift-pipelines/[Cluster Name]/`
  - OpenShift Pipelines operator deployment
- **Pipelines**: Multiple pipeline overlays
  - `pipelines/hello-world/[Cluster Name]/`
  - `pipelines/cloud-infrastructure-provisioning/[Cluster Name]/`
- **Deployments**: `deployments/ocm/[Cluster Name]/`
  - OCM service deployments
- **GitOps**: `gitops-applications/[Cluster Name].yaml`
  - ArgoCD ApplicationSet for cluster management
  - Automatic update to `gitops-applications/kustomization.yaml`

### 4. Automatic Validation and Feedback
- Shows configuration summary before proceeding
- Requires user confirmation
- **Automatically runs validation checks** after generation:
  - `oc kustomize clusters/[cluster-name]/` - validates cluster configuration
  - `oc kustomize deployments/ocm/[cluster-name]/` - validates deployments configuration
  - `oc kustomize gitops-applications/` - validates GitOps applications
- **Clear status reporting** with ✅/❌ indicators for each validation check
- **Error handling** with debug commands if validation fails
- Lists all generated files and provides next steps

## Usage

```bash
./bin/new-cluster
```

The tool will interactively prompt for all required information and generate a complete cluster configuration ready for GitOps deployment.

## Example Output

```
OpenShift Regional Cluster Generator
===================================

Please provide the following information for your new cluster:

Cluster Name: my-new-cluster
Cluster Type (ocp/eks) [ocp]: eks
Region [us-west-2]: us-east-1
Base Domain [rosa.mturansk-test.csu2.i3.devshift.org]: 
Instance Type [m5.2xlarge]: m5.large
Number of Replicas [2]: 3

Configuration Summary:
=====================
Cluster Name: my-new-cluster
Type: eks
Region: us-east-1
Domain: rosa.mturansk-test.csu2.i3.devshift.org
Instance Type: m5.large
Replicas: 3

Proceed with cluster generation? (y/N): y

Generated regional specification at: regions/us-east-1/my-new-cluster/region.yaml
Running bin/generate-cluster to create cluster configuration...
[...generation output...]
🎉 Cluster generation completed successfully!

Validating generated configuration...

Validating cluster configuration...
✅ Cluster configuration is valid
Validating deployments configuration...
✅ Deployments configuration is valid
Validating GitOps applications...
✅ GitOps applications configuration is valid

✅ All validation checks passed!
```

## Automatic Validation

The tool now automatically runs validation checks after generating the cluster configuration:

- **Cluster Configuration**: `oc kustomize clusters/[cluster-name]/`
- **Deployments Configuration**: `oc kustomize deployments/ocm/[cluster-name]/`
- **GitOps Applications**: `oc kustomize gitops-applications/`

Each validation check shows a clear ✅ or ❌ status indicator. If validation fails, the tool provides debug commands to investigate the issues.

## Integration with Bootstrap Workflow

1. **Generation**: Use `bin/new-cluster` to create cluster configuration
2. **Review**: Validate generated files and configuration
3. **Commit**: Add changes to git repository
4. **Deploy**: Run `./bootstrap.sh` to provision cluster via GitOps

## Error Handling

- Validates cluster name uniqueness
- Checks cluster type input
- Provides clear error messages
- Allows cancellation at confirmation step
- **Automatic validation** with clear status indicators
- **Debug commands** provided when validation fails
- Cleans up on failure

## Future Improvements

Potential enhancements for future iterations:

1. **Batch Mode**: Support for non-interactive mode with CLI flags
2. **Configuration Templates**: Pre-defined cluster templates
3. **Advanced Validation**: Network and resource validation
4. **Rollback Support**: Automatic cleanup on generation failure
5. **Enhanced Validation**: Additional resource validation beyond kustomize
6. **Multi-Region Support**: Cross-region cluster configurations
