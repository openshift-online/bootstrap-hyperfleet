# bin/update-dynamic-docs Requirements

## Purpose

The `update-dynamic-docs` script automates the update of dynamic documentation content, synchronizing repository documentation with live cluster state and current configurations.

## Functional Requirements

### Dynamic Content Categories

#### STATUS.md Updates
- **Live Cluster Status**: ArgoCD applications, managed clusters, provisioning status
- **Health Summary**: Overall system health with issue identification
- **Quick Commands**: Operational commands for status checking
- **Configuration Overview**: Repository-based cluster and regional distribution

#### Inventory Documentation
- **Cluster Inventory**: Live cluster status combined with repository configurations
- **Application Inventory**: GitOps applications and service deployments
- **Regional Distribution**: Cluster organization by geographic region
- **Deployment Configuration**: Operator and pipeline deployment statistics

#### Documentation Metrics
- **File Statistics**: Total markdown files, lines, words, averages
- **Coverage Analysis**: Documentation coverage of cluster configurations
- **Size Distribution**: Largest documentation files for maintenance insight

### Cluster Access Management

#### Connection Validation
- **oc Command**: Verify OpenShift CLI availability
- **Authentication**: Check cluster authentication via `oc auth can-i`
- **Graceful Degradation**: Continue with static content when cluster unavailable
- **Offline Mode**: Support --offline flag to skip cluster data collection

#### Live Data Integration
When cluster access available, must collect:
- **ArgoCD Applications**: Sync status, health status, target clusters
- **Managed Clusters**: Status, version, creation timestamps
- **Cluster Provisioning**: Hive ClusterDeployments and CAPI clusters
- **ApplicationSets**: Generator configurations and creation times

### Output Requirements

#### Auto-generated File Headers
```markdown
# [Document Title]

**Auto-generated [description] from [sources]**

*Last updated: [timestamp]*
```

#### Status Indicators
- ✅ **All systems operational**: No issues detected
- ⚠️ **Issues detected**: Specific problem counts
- Clear problem identification with counts

#### Data Formatting
- **Code Blocks**: Use triple backticks for command output
- **Custom Columns**: Structured data presentation via oc custom-columns
- **Sorted Output**: Consistent ordering for readability

### Command Line Interface

#### Usage Patterns
```bash
# Update all dynamic content
./bin/update-dynamic-docs

# Update specific content types
./bin/update-dynamic-docs --status-only
./bin/update-dynamic-docs --inventory-only
./bin/update-dynamic-docs --metrics-only

# Control timestamp updates
./bin/update-dynamic-docs --no-timestamps

# Run without cluster access
./bin/update-dynamic-docs --offline
```

#### Option Requirements
- `--status-only`: Update only STATUS.md file
- `--inventory-only`: Update only inventory files in docs/reference/
- `--metrics-only`: Generate only documentation metrics
- `--no-timestamps`: Skip timestamp updates in existing files
- `--offline`: Skip live cluster data collection
- `-h, --help`: Display usage information

### File Management

#### Target Files
Must create/update these files:
- **STATUS.md**: Root-level system status
- **docs/reference/cluster-inventory.md**: Cluster inventory
- **docs/reference/application-inventory.md**: Application inventory  
- **docs/reference/documentation-metrics.md**: Documentation statistics

#### Directory Creation
- **Auto-creation**: Create docs/reference/ directory if missing
- **Permissions**: Handle file system permission issues
- **Atomic Updates**: Use temporary files for safe updates

### Error Handling Requirements

#### Missing Dependencies
- **oc Command**: Continue without live data if oc unavailable
- **Cluster Access**: Graceful handling when cluster unreachable
- **Permission Issues**: Handle insufficient cluster permissions

#### File System Operations
- **Directory Creation**: Create missing directories automatically
- **Write Permissions**: Handle filesystem permission issues
- **Temporary Files**: Clean up temporary files on errors

### Integration Requirements

#### Repository Structure Awareness
Must discover and report:
- **Cluster Configurations**: Count and list clusters/ directory contents
- **Regional Organization**: Analyze regions/ directory structure
- **GitOps Applications**: Count gitops-applications/ configurations
- **Operator Deployments**: Analyze operators/ directory structure
- **Pipeline Configurations**: Count pipelines/ directory contents

#### Live System Integration
- **Health Assessment**: Analyze application sync and health status
- **Cluster Status**: Monitor managed cluster availability
- **Resource Counting**: Count provisioned vs configured resources

### Advanced Features

#### Architecture Diagram Support
- **Mermaid Detection**: Check for mermaid-cli (mmdc) availability
- **Diagram Discovery**: Find markdown files with mermaid diagrams
- **Future Enhancement**: Placeholder for diagram generation

#### Timestamp Management
- **Document Timestamps**: Update "Last updated" timestamps in docs
- **Selective Updates**: Optional timestamp updating via --no-timestamps
- **Pattern Matching**: Use sed for consistent timestamp formatting

### Performance Requirements

#### Efficient Data Collection
- **Batch Queries**: Use custom-columns for structured data collection
- **Limited Output**: Use head to limit large result sets
- **Silent Operations**: Suppress unnecessary command output

#### Resource Usage
- **Temporary Files**: Use mktemp for safe temporary file creation
- **Memory Efficiency**: Stream processing for large data sets
- **Quick Execution**: Minimize total execution time

### Dependencies

#### External Commands
- **oc**: OpenShift CLI for cluster data collection
- **jq**: JSON processing for health summary calculations
- **find**: Repository structure discovery
- **sed**: Timestamp updates in existing files

#### Optional Dependencies
- **mmdc**: Mermaid CLI for diagram generation (optional)
- **wc**: Word/line counting for metrics
- **sort**: Data sorting for consistent output

### Quality Assurance

#### Data Validation
- **Error Handling**: Handle missing resources gracefully
- **Fallback Content**: Provide meaningful content when data unavailable
- **Consistency**: Ensure consistent formatting across all outputs

#### Documentation Standards
- **Auto-generation Markers**: Clear indication of auto-generated content
- **Cross-references**: Maintain links to related documentation
- **Editing Warnings**: Warn against manual editing of auto-generated files

## Related Tools

### Documentation Pipeline
- **[generate-docs.md](./generate-docs.md)** - Creates base documentation that this tool updates with dynamic content
- **[validate-docs.md](./validate-docs.md)** - Validates the dynamically updated documentation

### Data Sources
- **[status.md](./status.md)** - Provides cluster status data for documentation updates
- **[health-check.md](./health-check.md)** - Provides health data for system status documentation

### Infrastructure Integration
- **[bootstrap.md](./bootstrap.md)** - Documents the infrastructure that provides data for dynamic updates

## Design Principles

*This tool enables **living documentation** - automatically synchronizing documentation with live system state and repository configuration changes, ensuring documentation accuracy and reducing maintenance overhead.*