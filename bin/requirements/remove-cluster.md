# bin/remove-cluster Requirements

## Requirements

### Semantic Naming Requirements
- **MANDATORY**: Accept cluster names using semantic naming format: `{type}-{number}` or `{type}-{number}-{suffix}`
- **MANDATORY**: Support three cluster types: `ocp`, `eks`, `hcp`
- **MANDATORY**: Validate cluster name exists before attempting removal
- **MANDATORY**: Handle zero-padded numbering: `01`, `02`, `03`, etc.
- **MANDATORY**: Support name suffixes (e.g., `hcp-01-mytest`, `ocp-02-dev`)

### Cluster Name Validation Logic
1. **Input**: User provides cluster name to remove
2. **Validate**: Verify cluster exists in regions/ directory structure
3. **Confirm**: Prompt user for confirmation before removal
4. **Remove**: Delete all references systematically
5. **Verify**: Confirm complete removal

### Examples
- Remove first OCP cluster: `ocp-01`
- Remove second OCP cluster with suffix: `ocp-02-dev`  
- Remove first EKS cluster: `eks-01`
- Remove HCP cluster with suffix: `hcp-01-test`

## Features

### 1. Interactive Input Collection and Validation
The tool prompts for required input with validation:

- **Cluster Name** (string, required)
  - Validates cluster name format matches semantic naming
  - Checks cluster exists in `regions/` directory
  - Confirms cluster has generated files to remove
  - Shows what will be removed before proceeding

### 2. Comprehensive Removal Process
Systematically removes all cluster references in order:

#### Phase 1: Hub Cluster Resources (if connected)
- **ManagedCluster**: `oc delete managedcluster [cluster-name]`
- **ClusterDeployment**: `oc delete clusterdeployment [cluster-name] -n [cluster-name]`
- **Namespace**: `oc delete namespace [cluster-name]`
- **ArgoCD Applications**: `oc delete application -n openshift-gitops -l app.kubernetes.io/name=[cluster-name]`
- **ApplicationSets**: `oc delete applicationset [cluster-name]-applications -n openshift-gitops`

#### Phase 2: Repository File Removal
Remove all generated files in dependency order:

1. **Regional Specification**: `regions/[region]/[cluster-name]/`
   - Remove the source `region.yaml` file
   - Remove parent directory if empty
2. **GitOps Applications**: `gitops-applications/[cluster-name].yaml`
   - Remove ArgoCD ApplicationSet configuration
   - Update `gitops-applications/kustomization.yaml` to remove reference
3. **Cluster Manifests**: `clusters/[cluster-name]/`
   - Remove all cluster deployment manifests
   - Includes install-config, namespace, deployment resources
4. **Operator Deployments**: `operators/openshift-pipelines/[cluster-name]/`
   - Remove cluster-specific operator configurations
5. **Pipeline Configurations**: 
   - `pipelines/hello-world/[cluster-name]/`
   - `pipelines/cloud-infrastructure-provisioning/[cluster-name]/`
6. **Service Deployments**: `deployments/ocm/[cluster-name]/`
   - Remove cluster-specific service configurations

### 3. Safety Features and Validation
- **Pre-removal validation**: Verify all expected files exist
- **Confirmation prompt**: Show complete removal plan and require user confirmation
- **Dry-run mode**: Optional flag to show what would be removed without actually removing
- **Error handling**: Continue removal even if some hub cluster resources don't exist
- **Rollback protection**: Warn if cluster appears to be running/active

### 4. Status Reporting and Feedback
- **Phase indicators**: Show progress through removal phases
- **File tracking**: List each file/resource as it's removed
- **Success confirmation**: Verify complete removal
- **Summary report**: Show what was successfully removed
- **Error reporting**: Clear messages for any failures

## Command Line Interface

### Basic Usage
```bash
./bin/remove-cluster [cluster-name]
```

### Optional Flags
```bash
./bin/remove-cluster [cluster-name] --dry-run    # Show what would be removed
./bin/remove-cluster [cluster-name] --force      # Skip confirmation prompts
./bin/remove-cluster [cluster-name] --hub-only   # Only remove hub cluster resources
./bin/remove-cluster [cluster-name] --files-only # Only remove repository files
```

## Error Handling

### Validation Errors
- Cluster name format validation with helpful examples
- Missing cluster detection with suggestions for similar names
- Empty regions directory handling

### Removal Errors
- Hub cluster connection failures (continue with file removal)
- Missing hub resources (warn but continue)
- File permission errors with clear messages
- Partial removal tracking and recovery suggestions

### Recovery Scenarios
- **Interrupted removal**: Resume from last successful phase
- **Hub cluster unavailable**: Skip hub cleanup, remove files only
- **Partial file removal**: Report what remains and provide cleanup commands

## Safety Considerations

### Confirmation Requirements
- **Default behavior**: Always prompt for confirmation
- **Cluster status warning**: Check if cluster appears active before removal
- **Dependency warning**: Alert if cluster has dependent resources

### Data Protection
- **No data removal**: Script only removes configuration, not cluster data
- **Backup suggestion**: Recommend backing up important configurations
- **Audit trail**: Log all removal actions for recovery if needed

## Integration with Existing Tools

### Complementary Tools
- **[new-cluster.md](./new-cluster.md)** - Creates clusters that this tool removes
- **[health-check.md](./health-check.md)** - Can verify cluster removal completion
- **[bootstrap.md](./bootstrap.md)** - May need re-run after cluster removal

### Validation Tools
- **[validate-docs.md](./validate-docs.md)** - Ensure documentation consistency after removal
- Uses same validation logic as new-cluster for consistency

## Future Improvements

Potential enhancements for future iterations:

1. **Bulk Removal**: Remove multiple clusters with pattern matching
2. **Archive Mode**: Move configurations to archive directory instead of deletion
3. **Dependency Analysis**: Check for cross-cluster dependencies before removal
4. **Backup Integration**: Automatic configuration backup before removal
5. **Audit Logging**: Detailed logs of all removal actions
6. **Recovery Tools**: Restore accidentally removed clusters from backups

## Related Tools

### Prerequisites
- Hub cluster access for complete removal (optional)
- File system write permissions for repository cleanup

### Direct Dependencies
- None - standalone removal tool

### Complementary Workflows
- **[health-check.md](./health-check.md)** - Verify removal completion
- **[new-cluster.md](./new-cluster.md)** - Recreate clusters if needed
- **[bootstrap.md](./bootstrap.md)** - Re-initialize GitOps after bulk removals