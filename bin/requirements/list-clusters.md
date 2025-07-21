# bin/list-clusters Requirements

## Requirements

### Primary Function
- **MANDATORY**: Walk the `regions/` directory structure to discover all cluster configurations
- **MANDATORY**: Parse regional specification files to extract cluster metadata
- **MANDATORY**: Present information in a clear, formatted table view
- **MANDATORY**: Support multiple output formats for different use cases

### Semantic Naming Requirements
- **MANDATORY**: Recognize and parse semantic naming format: `{type}-{number}` or `{type}-{number}-{suffix}`
- **MANDATORY**: Support three cluster types: `ocp`, `eks`, `hcp`
- **MANDATORY**: Handle zero-padded numbering: `01`, `02`, `03`, etc.
- **MANDATORY**: Display cluster type, number, and suffix separately when applicable

### Discovery Logic
1. **Scan**: Recursively search `regions/*/` directories for cluster specifications
2. **Parse**: Extract metadata from each `region.yaml` file
3. **Validate**: Check for required fields and report any issues
4. **Sort**: Order clusters by type, then by number for consistent display
5. **Format**: Present in requested output format

## Features

### 1. Directory Structure Discovery
The tool scans the regions directory structure:

- **Regions Directory**: `regions/[region-name]/[cluster-name]/region.yaml`
- **Pattern Recognition**: Automatically identifies cluster directories
- **Error Handling**: Reports missing or malformed specifications
- **Recursive Search**: Handles nested directory structures

### 2. Metadata Extraction
Parses regional specification files to extract:

- **Basic Information**:
  - Cluster name (semantic format)
  - Cluster type (`ocp`, `eks`, `hcp`)
  - Region location
  - Base domain
- **Compute Configuration**:
  - Instance type
  - Number of replicas
- **Type-Specific Settings**:
  - OpenShift version/channel (OCP)
  - Kubernetes version (EKS)
  - HyperShift configuration (HCP)

### 3. Multiple Output Formats

#### Default Table Format
```
CLUSTER NAME        TYPE  REGION     DOMAIN                                    INSTANCE TYPE  REPLICAS  VERSION
=================== ===== ========== ========================================= ============== ======== =========
ocp-01              ocp   us-west-2  rosa.mturansk-test.csu2.i3.devshift.org  m5.2xlarge     2        4.15
eks-01-test         eks   us-east-1  rosa.mturansk-test.csu2.i3.devshift.org  m5.large       3        1.28
hcp-01              hcp   us-west-2  rosa.mturansk-test.csu2.i3.devshift.org  m5.xlarge      2        4.15
```

#### Compact Format (`--compact`)
```
ocp-01 (ocp/us-west-2/m5.2xlarge/2)
eks-01-test (eks/us-east-1/m5.large/3) 
hcp-01 (hcp/us-west-2/m5.xlarge/2)
```

#### JSON Format (`--json`)
```json
[
  {
    "name": "ocp-01",
    "type": "ocp",
    "region": "us-west-2",
    "domain": "rosa.mturansk-test.csu2.i3.devshift.org",
    "instanceType": "m5.2xlarge",
    "replicas": 2,
    "version": "4.15",
    "channel": "stable",
    "path": "regions/us-west-2/ocp-01/region.yaml"
  }
]
```

#### CSV Format (`--csv`)
```csv
Name,Type,Region,Domain,InstanceType,Replicas,Version,Path
ocp-01,ocp,us-west-2,rosa.mturansk-test.csu2.i3.devshift.org,m5.2xlarge,2,4.15,regions/us-west-2/ocp-01/region.yaml
```

### 4. Filtering and Sorting Options

#### Filter by Type
```bash
./bin/list-clusters --type ocp        # Only OCP clusters
./bin/list-clusters --type eks,hcp    # Only EKS and HCP clusters
```

#### Filter by Region
```bash
./bin/list-clusters --region us-west-2    # Only us-west-2 clusters
./bin/list-clusters --region us-east-1,us-west-2  # Multiple regions
```

#### Sort Options
```bash
./bin/list-clusters --sort name        # Sort by cluster name (default)
./bin/list-clusters --sort type        # Sort by cluster type
./bin/list-clusters --sort region      # Sort by region
./bin/list-clusters --sort replicas    # Sort by replica count
```

### 5. Status Integration
Show deployment status when connected to hub cluster:

```
CLUSTER NAME        TYPE  REGION     STATUS      LAST SEEN
=================== ===== ========== =========== =================
ocp-01              ocp   us-west-2  Available   2025-01-20 14:30
eks-01-test         eks   us-east-1  Pending     2025-01-20 14:25
hcp-01              hcp   us-west-2  NotFound    Never
```

Status meanings:
- **Available**: ManagedCluster exists and is ready
- **Pending**: ManagedCluster exists but not ready
- **NotFound**: No corresponding ManagedCluster found
- **Error**: ManagedCluster exists but has errors

## Command Line Interface

### Basic Usage
```bash
./bin/list-clusters                    # Default table format
```

### Output Format Options
```bash
./bin/list-clusters --format table    # Default table format
./bin/list-clusters --format compact  # Compact one-line format
./bin/list-clusters --format json     # JSON format
./bin/list-clusters --format csv      # CSV format
```

### Filtering Options
```bash
./bin/list-clusters --type ocp                    # Filter by cluster type
./bin/list-clusters --region us-west-2            # Filter by region
./bin/list-clusters --name "*-test"               # Filter by name pattern
```

### Sorting and Display Options
```bash
./bin/list-clusters --sort type,name              # Sort by type, then name
./bin/list-clusters --no-header                   # Suppress table header
./bin/list-clusters --show-path                   # Include file paths
./bin/list-clusters --show-status                 # Include deployment status
```

### Validation Options
```bash
./bin/list-clusters --validate                    # Validate all specifications
./bin/list-clusters --show-errors                 # Show parsing errors
./bin/list-clusters --check-files                 # Verify generated files exist
```

## Error Handling

### Specification Validation
- **Missing Files**: Report clusters with missing `region.yaml`
- **Parse Errors**: Show line numbers for YAML parsing failures
- **Required Fields**: Validate presence of mandatory fields
- **Format Validation**: Check semantic naming compliance

### Directory Structure Issues
- **Empty Regions**: Report regions with no clusters
- **Invalid Paths**: Handle malformed directory structures
- **Permission Issues**: Clear error messages for access problems

### Hub Cluster Integration
- **Connection Failures**: Continue with local data if hub unavailable
- **Authentication Issues**: Clear guidance on login requirements
- **API Errors**: Graceful degradation when ManagedCluster API unavailable

## Integration with Existing Tools

### Complementary Tools
- **[new-cluster.md](./new-cluster.md)** - Creates clusters that this tool lists
- **[remove-cluster.md](./remove-cluster.md)** - Removes clusters that this tool discovers
- **[health-check.md](./health-check.md)** - Uses similar status checking logic

### Validation Integration
- **[validate-docs.md](./validate-docs.md)** - Can use cluster list for documentation validation
- Uses same regional specification parsing logic as other tools

### Automation Support
- **Machine-readable formats**: JSON and CSV for scripting
- **Exit codes**: 0 for success, non-zero for errors
- **Consistent output**: Stable format for parsing by other scripts

## Advanced Features

### Performance Optimization
- **Parallel Processing**: Parse multiple region.yaml files concurrently
- **Caching**: Cache parsed results for repeated calls
- **Lazy Loading**: Only fetch hub cluster status when requested

### Extensibility
- **Plugin System**: Support for custom output formatters
- **Custom Fields**: Allow additional metadata extraction
- **External Integrations**: API for other tools to use cluster discovery

### Monitoring Integration
- **Metrics Export**: Prometheus metrics for cluster counts
- **Health Reporting**: Integration with monitoring systems
- **Alerting**: Detect configuration drift or missing clusters

## Usage Examples

### Daily Operations
```bash
# Quick overview of all clusters
./bin/list-clusters --compact

# Check status of OCP clusters
./bin/list-clusters --type ocp --show-status

# Find test clusters across all regions
./bin/list-clusters --name "*test*" --show-path
```

### Automation Scripts
```bash
# Export cluster inventory to CSV
./bin/list-clusters --csv > cluster-inventory.csv

# Get cluster count by type
./bin/list-clusters --json | jq 'group_by(.type) | map({type: .[0].type, count: length})'

# Validate all cluster specifications
./bin/list-clusters --validate --show-errors
```

### Troubleshooting
```bash
# Find clusters with missing generated files
./bin/list-clusters --check-files --show-errors

# Compare regional specs with hub cluster state
./bin/list-clusters --show-status --format table

# Export detailed cluster information for support
./bin/list-clusters --json --show-path --validate > cluster-debug.json
```

## Future Improvements

Potential enhancements for future iterations:

1. **Interactive Mode**: Terminal UI for browsing clusters
2. **Diff Mode**: Compare regional specifications with deployed state  
3. **History Tracking**: Show cluster creation and modification dates
4. **Dependency Analysis**: Show relationships between clusters and services
5. **Cost Analysis**: Integrate with AWS cost data for resource planning
6. **Template Analysis**: Identify clusters using non-standard configurations

## Related Documentation

### Prerequisites
- Regional specifications exist in `regions/` directory structure
- Optional: Hub cluster access for status information

### Direct Dependencies
- None - standalone discovery and reporting tool

### Complementary Workflows
- **[health-check.md](./health-check.md)** - Comprehensive cluster health analysis
- **[new-cluster.md](./new-cluster.md)** - Create new clusters to be listed
- **[remove-cluster.md](./remove-cluster.md)** - Remove clusters from listings