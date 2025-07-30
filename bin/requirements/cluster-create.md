# bin/cluster-create Requirements

## Requirements

### Semantic Naming Requirements
- **MANDATORY**: All cluster names MUST use semantic naming format: `{type}-{number}` or `{type}-{number}-{suffix}`
- **MANDATORY**: Do NOT prompt user for cluster name - generate automatically
- **MANDATORY**: Support three cluster types: `ocp`, `eks`, `hcp`
- **MANDATORY**: Use zero-padded numbering: `01`, `02`, `03`, etc.
- **MANDATORY**: Find next available number for the specified type
- **MANDATORY**: Cluster names MUST NOT exceed 19 characters to prevent Kubernetes label length violations
- **MANDATORY**: Cluster names must contain only lowercase letters, numbers, and hyphens
- **MANDATORY**: Cluster names must start and end with alphanumeric characters
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

### Cluster Name Length Validation

**Problem**: ApplicationSet generates Kubernetes application names by combining cluster names with component names:
- Pattern: `{cluster-name}-{component}`
- Longest component: `pipelines-cloud-infrastructure-provisioning` (43 characters)
- Kubernetes label length limit: 63 characters

**Solution**: Maximum cluster name length = 63 - 43 - 1 (hyphen) = **19 characters**

**Examples of generated ApplicationSet names**:
- `ocp-01-cluster` (14 chars) ✅
- `ocp-01-pipelines-cloud-infrastructure-provisioning` (50 chars) ✅
- `ocp-01-mturansk-jul29-pipelines-cloud-infrastructure-provisioning` (67 chars) ❌ Exceeds limit

**Validation prevents**:
- Kubernetes label length violations
- ApplicationSet generation failures
- ArgoCD sync issues

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
Automatically calls `bin/cluster-generate` to create:

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

## Error Handling

- Automatically generates semantic cluster names
- **Validates cluster name length** (maximum 19 characters for Kubernetes compatibility)
- **Validates cluster name format** (lowercase alphanumeric and hyphens only)
- Validates cluster name uniqueness
- Checks cluster type input
- Provides clear error messages with detailed explanations
- Allows cancellation at confirmation step
- **Automatic validation** with clear status indicators
- **Debug commands** provided when validation fails
- Cleans up on failure

### Cluster Name Validation Errors

The tool provides detailed error messages for common validation failures:

```bash
# Length validation error
Error: Cluster name 'ocp-01-mturansk-jul29' is 20 characters long
Maximum allowed length is 19 characters to avoid Kubernetes label length violations
Generated ApplicationSet names like 'ocp-01-mturansk-jul29-pipelines-cloud-infrastructure-provisioning' would exceed 63 character limit

# Format validation error  
Error: Cluster name 'OCP-01-Test_Cluster' contains invalid characters
Cluster names must contain only lowercase letters, numbers, and hyphens
Must start and end with an alphanumeric character
```

## Future Improvements

Potential enhancements for future iterations:

1. **Batch Mode**: Support for non-interactive mode with CLI flags
2. **Configuration Templates**: Pre-defined cluster templates
3. **Advanced Validation**: Network and resource validation
4. **Rollback Support**: Automatic cleanup on generation failure
5. **Enhanced Validation**: Additional resource validation beyond kustomize
6. **Multi-Region Support**: Cross-region cluster configurations

## Related Tools

### Prerequisites
- **[bootstrap.md](./bootstrap.md)** - Sets up the GitOps infrastructure for cluster management

### Direct Dependencies
- **[generate-cluster.md](./generate-cluster.md)** - Automatically called to create complete cluster configurations

### Bulk Operations
- **[regenerate-all-clusters.md](./regenerate-all-clusters.md)** - Regenerates all clusters including those created by this tool

### Alternative Workflows
- **[convert-cluster.md](./convert-cluster.md)** - Converts existing clusters to the regional specification format