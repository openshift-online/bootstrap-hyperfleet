# bin/list-clusters Requirements

## Requirements

### Primary Function
- **MANDATORY**: Walk the `regions/` directory structure to discover all cluster configurations
- **MANDATORY**: Parse regional specification files to extract cluster metadata
- **MANDATORY**: Present information in a clear, formatted table view

### Semantic Naming Requirements
- **MANDATORY**: Recognize and parse semantic naming format: `{type}-{number}` or `{type}-{number}-{suffix}`
- **MANDATORY**: Support three cluster types: `ocp`, `eks`, `hcp`
- **MANDATORY**: Handle zero-padded numbering: `01`, `02`, `03`, etc.

### Discovery Logic
1. **Scan**: Search `regions/*/` directories for cluster specifications
2. **Parse**: Extract metadata from each `region.yaml` file
3. **Display**: Present in simple table format

## Features

### Directory Structure Discovery
The tool scans the regions directory structure:

- **Regions Directory**: `regions/[region-name]/[cluster-name]/region.yaml`
- **Pattern Recognition**: Automatically identifies cluster directories
- **Error Handling**: Reports missing or malformed specifications

### Metadata Extraction
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

### Simple Table Output
```
CLUSTER NAME        TYPE  REGION     DOMAIN                                    INSTANCE TYPE  REPLICAS  VERSION
=================== ===== ========== ========================================= ============== ======== =========
ocp-01-mturansk-a   ocp   us-west-2  rosa.mturansk-test.csu2.i3.devshift.org  m5.2xlarge     2        4.15
ocp-01-mturansk-test ocp  us-west-2  rosa.mturansk-test.csu2.i3.devshift.org  m5.large       3        4.15
```

## Command Line Interface

### Basic Usage
```bash
./bin/list-clusters                    # Simple table format
```

## Error Handling

### Specification Validation
- **Missing Files**: Report clusters with missing `region.yaml`
- **Parse Errors**: Show YAML parsing failures
- **Required Fields**: Validate presence of mandatory fields

### Directory Structure Issues
- **Empty Regions**: Report regions with no clusters
- **Invalid Paths**: Handle malformed directory structures

## Integration with Existing Tools

### Complementary Tools
- **[remove-cluster.md](./remove-cluster.md)** - Removes clusters that this tool discovers
- Uses same regional specification parsing logic as other tools