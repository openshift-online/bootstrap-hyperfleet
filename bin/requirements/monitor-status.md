# bin/status.sh Requirements

## Purpose

The `status.sh` script waits for specific Kubernetes CustomResourceDefinitions (CRDs) to be established, providing reliable synchronization during operator deployment and bootstrap processes.

## Functional Requirements

### CRD Monitoring
- **CRD Status Checking**: Monitor CRD establishment status via JSONPath query
- **Established Condition**: Check for `status.conditions[?(@.type=="Established")].status == "True"`
- **Continuous Polling**: Poll CRD status at regular intervals until established

### Input Requirements
- **CRD Name**: First argument must specify the CRD name to monitor
- **Timeout**: Second argument optionally specifies timeout in seconds (default: 120)
- **Required Parameter**: Must exit with error if CRD name not provided

### Timing and Intervals
- **Default Timeout**: 120 seconds (2 minutes) if not specified
- **Sleep Interval**: 5 seconds between status checks
- **Elapsed Time Tracking**: Calculate and display elapsed time during polling

### Output Requirements

#### Progress Reporting
```
Waiting for CRD 'applications.argoproj.io' to be established...
Waiting for 'applications.argoproj.io' to be established. Retrying in 5s... (Elapsed: 10s)
Found 'applications.argoproj.io'
```

#### Error Messages
```
Error: No CRD name provided.
Usage: ./status.sh <crd-name> [timeout-in-seconds]

Timeout of 120s reached. CRD 'example.com' not found.
```

### Success/Failure Conditions

#### Success Criteria
- **CRD Established**: CRD exists and has `Established: True` condition
- **Within Timeout**: CRD becomes available before timeout expires
- **Exit Code 0**: Return success exit code when CRD found

#### Failure Criteria
- **Missing CRD Name**: Exit code 1 if no CRD name provided
- **Timeout Exceeded**: Exit code 1 if timeout reached before CRD established
- **Command Availability**: Assume `kubectl` command is available

### Usage Patterns

```bash
# Wait for ArgoCD Applications CRD (default 120s timeout)
./bin/status.sh applications.argoproj.io

# Wait for cert-manager CRD with custom timeout
./bin/status.sh certificates.cert-manager.io 300

# Check ACM PolicyReport CRD
./bin/status.sh policyreports.wgpolicyk8s.io

# Usage in bootstrap sequence
./bin/bootstrap.sh
./bin/status.sh applications.argoproj.io
# Continue with ArgoCD application deployment
```

### Error Handling Requirements

#### Command Line Validation
- **Parameter Count**: Validate required CRD name parameter
- **Usage Display**: Show proper usage format on parameter errors

#### Kubernetes API Access
- **Silent Errors**: Use `2>/dev/null` to suppress kubectl error output
- **Connection Issues**: Handle API server unavailability gracefully
- **Permission Issues**: Handle insufficient permissions to query CRDs

#### Timeout Management
- **Precise Timing**: Use epoch seconds for accurate timeout calculation
- **Timeout Message**: Clear indication when timeout exceeded
- **Cleanup**: Ensure script exits cleanly on timeout

### Integration Requirements

#### Bootstrap Process Integration
- **Operator Deployment**: Wait for operator CRDs after subscription creation
- **Dependency Management**: Ensure CRDs available before deploying dependent resources
- **Sequential Execution**: Block execution until CRDs ready

#### Common CRD Patterns
Support for typical OpenShift Bootstrap CRDs:
- **ArgoCD**: `applications.argoproj.io`, `applicationsets.argoproj.io`
- **ACM**: `managedclusters.cluster.open-cluster-management.io`
- **CAPI**: `clusters.cluster.x-k8s.io`
- **Hive**: `clusterdeployments.hive.openshift.io`
- **Pipelines**: `pipelines.tekton.dev`, `tasks.tekton.dev`

### Performance Requirements

#### Efficient Polling
- **Minimal Resource Usage**: 5-second intervals balance responsiveness with resource usage
- **JSONPath Queries**: Use efficient JSONPath for condition checking
- **Silent Operation**: Suppress unnecessary kubectl output

#### Responsive Feedback
- **Progress Updates**: Show elapsed time during waiting
- **Immediate Success**: Exit immediately when CRD established
- **Clear Status**: Indicate when CRD found vs still waiting

### Dependencies

#### External Commands
- **kubectl**: Kubernetes CLI for CRD status queries
- **date**: System date command for timing calculations
- **sleep**: System sleep command for polling intervals

#### Kubernetes Access
- **Cluster Connection**: Must be connected to target Kubernetes cluster
- **CRD Permissions**: Must have permission to read CustomResourceDefinitions
- **API Availability**: Kubernetes API server must be accessible

### Command Line Interface

#### Parameter Format
```bash
./status.sh <crd-name> [timeout-in-seconds]
```

#### Parameter Validation
- **CRD Name**: Required first parameter, must be valid CRD name
- **Timeout**: Optional second parameter, must be positive integer
- **Help Display**: Show usage on invalid parameters

### Script Metadata

#### Header Information
- **Purpose**: CRD establishment monitoring
- **Usage Examples**: Clear command line examples
- **Configuration**: Documented timeout and interval settings

## Related Tools

### Called By
- **[bootstrap.md](./bootstrap.md)** - Uses this tool for CRD readiness validation

### Similar Function
- **[wait-kube.md](./wait-kube.md)** - More generic resource waiting with JSONPath conditions

### Workflow Integration
- **[bootstrap-vault-integration.md](./bootstrap-vault-integration.md)** - Runs after CRDs are established

## Design Principles

*This tool enables **dependency synchronization** - ensuring CustomResourceDefinitions are available before deploying resources that depend on them, preventing race conditions in operator deployment sequences.*