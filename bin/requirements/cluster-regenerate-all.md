# bin/regenerate-all-clusters Requirements

## Purpose

The `regenerate-all-clusters` script automates the regeneration of all cluster overlays from regional specifications, providing bulk cluster configuration updates with comprehensive validation.

## Functional Requirements

### Discovery and Processing

#### Regional Specification Discovery
- **File Pattern**: Must find all `region.yaml` files in `regions/` directory
- **Sorted Processing**: Process specifications in sorted order for consistency
- **Empty Handling**: Exit with error code 1 if no specifications found

#### Cluster Generation Process
- **Directory Extraction**: Extract cluster name from specification directory
- **Generator Invocation**: Call `./bin/cluster-generate` for each specification
- **Error Handling**: Continue processing remaining specs if one fails
- **Progress Reporting**: Clear status for each cluster processed

### Validation Requirements

#### Multi-Layer Validation
Must validate all generated overlays:
1. **Cluster Overlays**: `clusters/cluster-*` directories
2. **Pipeline Overlays**: `pipelines/cluster-*` directories  
3. **Deployment Overlays**: `deployments/ocm/cluster-*` directories

#### Validation Method
- **Dry-run Application**: Use `oc apply --dry-run=client` for validation
- **Silent Mode**: Suppress output during validation checks
- **Status Indicators**: Use ✅ for valid, ❌ for invalid overlays
- **Failure Tracking**: Track validation failures across all categories

### Output Requirements

#### Progress Reporting
```
=== Regenerating All Clusters from Regional Specifications ===

Found X regional specifications:
regions/us-east-1/cluster-name/region.yaml
...

Processing regions/us-east-1/cluster-name/ (cluster-name)...
✅ Generated cluster overlay for cluster-name

=== Regeneration Complete ===
```

#### Validation Reporting
```
Validating cluster overlays...
  cluster-10: ✅ Valid
  cluster-20: ❌ Invalid

Validating pipeline overlays...
  cluster-10: ✅ Valid

Validating deployment overlays...
  cluster-10: ✅ Valid
```

#### Summary Statistics
```
=== Summary ===
Clusters: X
Pipelines: X
Deployments: X
GitOps Apps: X
```

### Error Handling Requirements

#### Generation Failures
- **Continue Processing**: Don't stop on individual cluster generation failures
- **Error Reporting**: Log specific failures with cluster identification
- **Exit Status**: Complete all processing before reporting final status

#### Validation Failures
- **Comprehensive Checking**: Validate all overlays regardless of individual failures
- **Failure Tracking**: Maintain boolean flag for any validation failures
- **Exit Code**: Return exit code 1 if any validations failed

#### Missing Resources
- **Directory Existence**: Check directory existence before validation
- **File Permissions**: Handle permission issues gracefully
- **Command Availability**: Assume `oc` command is available

### Dependencies

#### External Scripts
- **generate-cluster**: Must exist in `bin/` directory for cluster generation
- **oc command**: OpenShift CLI for dry-run validation

#### Directory Structure
- **regions/**: Source directory for regional specifications
- **clusters/**: Target directory for cluster overlays
- **pipelines/**: Target directory for pipeline configurations
- **deployments/ocm/**: Target directory for service deployments
- **gitops-applications/**: Directory containing GitOps application files

#### File System Access
- **Read Access**: Must read regional specification files
- **Write Access**: Must write generated overlay files
- **Directory Traversal**: Must navigate repository directory structure

### Usage Patterns

```bash
# Regenerate all clusters from regional specs
./bin/regenerate-all-clusters

# Typical workflow
# 1. Update regional specifications in regions/
# 2. Run regeneration script
# 3. Review validation results
# 4. Commit changes if all validations pass
```

### Performance Requirements

#### Bulk Processing
- **Sequential Processing**: Process specifications one at a time
- **Progress Feedback**: Provide status updates during processing
- **Resource Usage**: Minimize concurrent resource usage

#### Validation Efficiency
- **Batch Validation**: Group validation by overlay type
- **Silent Operation**: Suppress unnecessary output during validation
- **Quick Feedback**: Provide immediate pass/fail status per overlay

### Integration Requirements

#### Git Workflow Integration
- **Commit Preparation**: Generate all overlays before validation
- **Change Detection**: Allow Git to detect configuration changes
- **Review Process**: Enable bulk review of generated changes

#### CI/CD Integration
- **Exit Codes**: Provide appropriate exit codes for automation
- **Machine Readable**: Output suitable for CI/CD parsing
- **Error Propagation**: Fail builds on validation errors

### File Path Patterns

#### Input Patterns
- **Regional Specs**: `regions/*/region.yaml`
- **Cluster Names**: Extracted from directory structure

#### Output Patterns
- **Clusters**: `clusters/cluster-*`
- **Pipelines**: `pipelines/cluster-*`
- **Deployments**: `deployments/ocm/cluster-*`
- **GitOps Apps**: `gitops-applications/cluster-*.yaml`

### Validation Coverage

#### Overlay Types
Must validate these configuration types:
- **Kustomize Overlays**: All cluster, pipeline, and deployment overlays
- **Kubernetes Resources**: Ensure generated YAML is valid
- **Cross-references**: Verify resource references are correct

#### Error Categories
- **Syntax Errors**: Invalid YAML or Kubernetes resource syntax
- **Reference Errors**: Missing base resources or invalid references
- **Schema Errors**: Resources that don't match Kubernetes schemas

## Related Tools

### Prerequisites
- **[new-cluster.md](./new-cluster.md)** - Creates regional specifications that this tool processes in bulk
- **[convert-cluster.md](./convert-cluster.md)** - Converts existing clusters to regional specifications

### Direct Dependencies
- **[generate-cluster.md](./generate-cluster.md)** - Core tool called for each regional specification

### Validation and Monitoring
- **[status.md](./status.md)** - Monitor the regenerated cluster configurations
- **[health-check.md](./health-check.md)** - Validate cluster health after regeneration

## Design Principles

*This tool enables **configuration consistency** - ensuring all cluster overlays are regenerated from authoritative regional specifications with comprehensive validation before deployment.*