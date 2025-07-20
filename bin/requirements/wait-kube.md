# bin/wait.kube.sh Requirements

## Purpose

The `wait.kube.sh` script provides flexible waiting for any Kubernetes resource to meet a specific condition defined by JSONPath expressions, enabling reliable synchronization in automation workflows.

## Functional Requirements

### Resource Monitoring
- **Generic Resource Support**: Wait for any Kubernetes resource type (pods, deployments, CRDs, etc.)
- **JSONPath Conditions**: Support arbitrary JSONPath expressions for condition checking
- **Expected Value Matching**: Compare JSONPath results against expected values
- **Namespace Flexibility**: Support both namespaced and cluster-scoped resources

### Input Requirements
- **Resource Type**: Kubernetes resource type (e.g., pod, deployment, service, crd)
- **Resource Name**: Specific resource instance name
- **Namespace**: Target namespace (empty string "" for cluster-scoped resources)
- **JSONPath Condition**: JSONPath expression to query resource status
- **Expected Value**: Value that JSONPath query must return for success
- **Timeout**: Optional timeout in seconds (default: 1800 seconds / 30 minutes)

### Timing and Polling
- **Default Timeout**: 1800 seconds (30 minutes) if not specified
- **Sleep Interval**: 60 seconds between status checks
- **Elapsed Time Tracking**: Calculate and display elapsed time during polling
- **Timeout Handling**: Exit with error code 1 when timeout exceeded

### Output Requirements

#### Progress Reporting
```
Waiting for resource 'deployment/my-app' in namespace 'production'...
Condition: JSONPath '{.status.readyReplicas}' must be '3'.
Resource not ready yet. Current status: '2'. Retrying in 60s... (Elapsed: 120s)
Success: Resource 'deployment/my-app' has met the condition.
```

#### Error Messages
```
Error: Missing required arguments.
Usage: ./wait.kube.sh <type> <name> <namespace> <jsonpath> <expected-value> [timeout]
Note: For cluster-scoped resources, provide an empty string for the namespace: ""

Error: Timeout of 1800s reached. Resource condition not met.
```

### JSONPath Expression Support

#### Common Patterns
- **Pod Readiness**: `{.status.conditions[?(@.type=="Ready")].status}` → "True"
- **Deployment Rollout**: `{.status.readyReplicas}` → desired replica count
- **CRD Establishment**: `{.status.conditions[?(@.type=="Established")].status}` → "True"
- **Service Endpoints**: `{.status.loadBalancer.ingress[0].ip}` → IP address

#### Advanced Conditions
- **Multi-condition Checks**: Support complex JSONPath expressions
- **Array Access**: Handle indexed array elements
- **Conditional Filtering**: Support condition-based filtering

### Usage Patterns

```bash
# Wait for Pod to be Ready
./bin/wait.kube.sh pod my-pod default '{.status.conditions[?(@.type=="Ready")].status}' "True"

# Wait for Deployment rollout completion
DESIRED=$(oc get deployment my-app -n web -o jsonpath='{.spec.replicas}')
./bin/wait.kube.sh deployment my-app web '{.status.readyReplicas}' "$DESIRED" 300

# Wait for CRD establishment (cluster-scoped)
./bin/wait.kube.sh crd applications.argoproj.io "" '{.status.conditions[?(@.type=="Established")].status}' "True" 120

# Wait for Route to exist (simple existence check)
./bin/wait.kube.sh route openshift-gitops-server openshift-gitops '{.kind}' "Route"

# Wait for MultiClusterHub completion
./bin/wait.kube.sh mch multiclusterhub open-cluster-management '{.status.conditions[?(@.type=="Complete")].message}' "All hub components ready."
```

### Error Handling Requirements

#### Argument Validation
- **Required Parameters**: Validate all required arguments are provided
- **Usage Display**: Show proper usage format on parameter errors
- **Parameter Order**: Ensure correct parameter order and types

#### Resource Access
- **Resource Existence**: Handle resources that don't exist yet
- **Permission Issues**: Handle insufficient permissions gracefully
- **API Availability**: Handle temporary API server unavailability

#### Timeout Management
- **Precise Timing**: Use epoch seconds for accurate timeout calculation
- **Timeout Message**: Clear indication when timeout exceeded
- **Exit Codes**: Return appropriate exit codes for automation

### Namespace Handling

#### Namespaced Resources
- **Namespace Flag**: Use `-n <namespace>` for namespaced resources
- **Namespace Validation**: Accept valid namespace names
- **Resource Scoping**: Limit queries to specified namespace

#### Cluster-Scoped Resources
- **Empty Namespace**: Accept "" for cluster-scoped resources
- **No Namespace Flag**: Omit namespace flag for cluster-scoped queries
- **Global Resources**: Handle CRDs, ClusterRoles, Nodes, etc.

### Integration Requirements

#### Bootstrap Workflow Integration
- **Sequential Execution**: Block execution until conditions met
- **Dependency Management**: Ensure resources ready before proceeding
- **Automation Support**: Reliable behavior in automated scripts

#### OpenShift Integration
- **oc Command**: Use OpenShift CLI for all resource queries
- **JSONPath Support**: Leverage oc's JSONPath implementation
- **Resource Types**: Support all OpenShift/Kubernetes resource types

### Performance Requirements

#### Efficient Polling
- **Reasonable Intervals**: 60-second intervals balance responsiveness with resource usage
- **Minimal Resource Usage**: Lightweight queries using JSONPath
- **Silent Errors**: Suppress expected errors (resource not found initially)

#### Responsive Feedback
- **Progress Updates**: Show current status during waiting
- **Immediate Success**: Exit immediately when condition met
- **Clear Status**: Display current vs expected values

### Dependencies

#### External Commands
- **oc**: OpenShift CLI for resource queries
- **date**: System date command for timing calculations
- **sleep**: System sleep command for polling intervals

#### Kubernetes Access
- **Cluster Connection**: Must be connected to target Kubernetes cluster
- **Resource Permissions**: Must have permission to read specified resource types
- **API Availability**: Kubernetes API server must be accessible

### Advanced Features

#### Flexible Condition Matching
- **String Comparison**: Direct string matching for status values
- **Numeric Comparison**: Support numeric comparisons via string matching
- **Complex JSONPath**: Support sophisticated JSONPath expressions

#### Debug Output
- **Current Status Display**: Show actual JSONPath query results
- **Resource Not Found**: Indicate when resource doesn't exist yet
- **Condition Tracking**: Display progress toward meeting condition

### Script Configuration

#### Configurable Parameters
- **Sleep Interval**: 60 seconds between checks (configurable in script)
- **Default Timeout**: 1800 seconds if not provided
- **Error Suppression**: Silent handling of expected errors

#### Documentation
- **Comprehensive Examples**: Cover common use cases
- **Usage Instructions**: Clear parameter descriptions
- **JSONPath Guidance**: Example JSONPath expressions for common scenarios

## Related Tools

### Called By
- **[bootstrap.md](./bootstrap.md)** - Uses this tool for resource readiness monitoring

### Similar Function
- **[status.md](./status.md)** - Specialized for CRD establishment waiting

### Workflow Integration
- **[bootstrap-vault-integration.md](./bootstrap-vault-integration.md)** - Uses generic resource waiting patterns

## Design Principles

*This tool enables **reliable synchronization** - providing flexible, generic waiting capabilities for any Kubernetes resource condition, ensuring automation workflows can reliably wait for resource readiness before proceeding.*