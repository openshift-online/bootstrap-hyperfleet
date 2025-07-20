# bin/new-cluster - Interactive Cluster Generator

## Overview
The `bin/new-cluster` tool is an interactive CLI that generates complete OpenShift/EKS cluster configurations for the bootstrap GitOps repository.

## Implementation Status
✅ **COMPLETED** - Tool has been implemented and tested successfully.

## Requirements

### Semantic Naming Requirements
- **MANDATORY**: All cluster names MUST use semantic naming format: `{type}-{number}` or `{type}-{number}-{suffix}`
- **MANDATORY**: Do NOT prompt user for cluster name - generate automatically
- **MANDATORY**: Support three cluster types: `ocp`, `eks`, `hcp`
- **MANDATORY**: Use zero-padded numbering: `01`, `02`, `03`, etc.
- **MANDATORY**: Find next available number for the specified type
- **OPTIONAL**: Allow name suffix for differentiation (e.g., `hcp-01-mytest`, `ocp-02-dev`)

### Cluster Name Generation Logic
1. **Input**: User selects cluster type (`ocp`, `eks`, or `hcp`) and optional name suffix (defaults to empty)
2. **Scan**: Check existing clusters and regions for pattern `{type}-XX`
3. **Generate**: Find next available number in sequence
4. **Suffix**: Append optional suffix if provided (e.g., `-mturansk-test`)
5. **Validate**: Ensure generated name doesn't conflict with existing clusters
6. **Output**: Use generated name throughout configuration

### Examples
- First OCP cluster: `ocp-01`
- Second OCP cluster: `ocp-02`  
- First EKS cluster: `eks-01`
- First HCP cluster: `hcp-01`
- If `hcp-01` exists, next would be `hcp-02`

## Features

### 1. Interactive Input Collection
The tool prompts for 5 required inputs with validation:

- **Type** (string, required)
  - Accepts "ocp", "eks", or "hcp"
  - Validates input before proceeding
- **Cluster Name** (auto-generated)
  - **REQUIREMENT**: Uses semantic naming: `{type}-{number}`
  - **REQUIREMENT**: Automatically generated based on type and existing clusters
  - Examples: `ocp-01`, `eks-01`, `hcp-01`, `ocp-02`, etc.
  - Validates uniqueness against existing clusters
  - Checks for existing cluster directories and GitOps applications
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
name: [Auto-Generated Name]  # e.g., "ocp-01", "eks-01", "hcp-01"
type: [Type]  # "ocp", "eks", or "hcp"
region: [Region]
domain: [Domain]
instanceType: [Instance Type]
replicas: [Replicas]
```

### 3. Complete Cluster Configuration
Automatically calls `bin/generate-cluster` to create:

- **Cluster manifests**: `clusters/[Semantic Name]/`
  - **OCP**: Namespace, Hive ClusterDeployment, MachinePool, install-config
  - **EKS**: Namespace, CAPI Cluster, AWSManagedControlPlane, AWSManagedMachinePool
  - **HCP**: Namespace, HostedCluster, SSH key secret
  - All types: ManagedCluster and KlusterletAddonConfig for ACM integration
- **Operators**: `operators/openshift-pipelines/[Semantic Name]/`
  - OpenShift Pipelines operator deployment
- **Pipelines**: Multiple pipeline overlays
  - `pipelines/hello-world/[Semantic Name]/`
  - `pipelines/cloud-infrastructure-provisioning/[Semantic Name]/`
- **Deployments**: `deployments/ocm/[Semantic Name]/`
  - OCM service deployments
- **GitOps**: `gitops-applications/[Semantic Name].yaml`
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

Cluster Type (ocp/eks/hcp) [ocp]: hcp
Region [us-west-2]: us-east-1
Base Domain [rosa.mturansk-test.csu2.i3.devshift.org]: 
Instance Type [m5.2xlarge]: m5.large
Number of Replicas [2]: 3

Generating cluster name...
Next available cluster name: hcp-01

Configuration Summary:
=====================
Cluster Name: hcp-01 (auto-generated)
Type: hcp
Region: us-east-1
Domain: rosa.mturansk-test.csu2.i3.devshift.org
Instance Type: m5.large
Replicas: 3

Proceed with cluster generation? (y/N): y

Generated regional specification at: regions/us-east-1/hcp-01/region.yaml
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
4. **Deploy**: Run `./bin/bootstrap.sh` to provision cluster via GitOps

## Error Handling

- Automatically generates semantic cluster names
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
