# bin/bootstrap.sh Requirements

## Purpose

The `bootstrap.sh` script orchestrates the complete initialization of an OpenShift GitOps-managed multi-cluster control plane, deploying all prerequisites, operators, and GitOps applications required for automated cluster lifecycle management.

## Functional Requirements

### Prerequisites Validation
- **Cluster Authentication**: Must verify user is logged into an OpenShift cluster via `oc cluster-info`
- **Permission Check**: Must have cluster-admin privileges to deploy operators and CRDs
- **Exit Condition**: Must exit with error code 1 if authentication fails

### Deployment Sequence
1. **Prerequisites Deployment**: Deploy OpenShift GitOps Operator via `oc apply -k ./prereqs`
2. **Operator Readiness**: Wait for GitOps CRDs to be available via `status.sh applications.argoproj.io`
3. **GitOps Applications**: Deploy all GitOps applications via `oc apply -k ./gitops-applications`
4. **Component Readiness**: Wait for core components (GitOps, ACM) to be fully operational
5. **Vault Integration**: Initialize vault-based secret management for cluster namespaces

### Component Readiness Requirements

#### OpenShift GitOps (ArgoCD)
- **Wait Condition**: Route `openshift-gitops-server` in `openshift-gitops` namespace exists
- **Validation**: Route kind must equal "Route"
- **Timeout**: Must handle reasonable timeout for operator deployment

#### Advanced Cluster Management (ACM)
- **Wait Condition**: MultiClusterHub `multiclusterhub` in `open-cluster-management` namespace
- **Status Check**: `status.conditions[?(@.type=="Complete")].message` equals "All hub components ready."
- **Validation**: All ACM hub components must be operational

### Output Requirements

#### Progress Reporting
- **Cluster Info**: Display target cluster information before deployment
- **Step Notifications**: Clear messaging for each deployment phase
- **Console Access**: Provide OpenShift console URL upon completion

#### Monitoring Guidance
- **Application Status**: `oc get applications -n openshift-gitops`
- **OCP Clusters**: `oc get clusterdeployments -A`
- **EKS Clusters**: `oc get clusters -A`
- **HCP Clusters**: `oc get hostedclusters -A`

### Integration Requirements

#### Vault Secret Management
- **Post-Deployment**: Execute `bootstrap.vault-integration.sh` after core components
- **ExternalSecrets**: Create external secret configurations for cluster namespaces
- **Deferred Sync**: ExternalSecrets won't sync until target clusters are provisioned

#### GitOps Automation
- **Cluster Provisioning**: Automated via GitOps ApplicationSets
- **Timing Expectations**: 
  - EKS clusters: ~15 minutes
  - OCP via Hive: ~45 minutes  
  - HCP clusters: ~10 minutes

### Error Handling Requirements

#### Authentication Failures
- **Clear Messaging**: "Please log in to an OpenShift cluster using 'oc login'"
- **Exit Code**: Must exit with code 1 for script automation compatibility

#### Deployment Failures
- **Component Timeouts**: Graceful handling of component readiness timeouts
- **Operator Issues**: Clear error reporting for operator deployment failures
- **Resource Conflicts**: Handle existing resource conflicts appropriately

### Usage Patterns

```bash
# Standard bootstrap deployment
./bin/bootstrap.sh

# Verify cluster access first
oc cluster-info && ./bin/bootstrap.sh
```

## Dependencies

### External Scripts
- `status.sh` - CRD readiness validation
- `wait.kube.sh` - Resource readiness waiting
- `bootstrap.vault-integration.sh` - Vault secret management setup

### Kustomize Configurations
- `./prereqs` - Operator subscriptions and prerequisites
- `./gitops-applications` - GitOps ApplicationSets and Applications

### Required Permissions
- **cluster-admin** role for operator deployment
- **Persistent storage** access for GitOps and ACM components
- **Network policies** configuration for multi-cluster communication

## Related Tools

### Prerequisites
- **[status.md](./status.md)** - Used for CRD establishment validation
- **[wait-kube.md](./wait-kube.md)** - Used for resource readiness monitoring

### Workflow Integration  
- **[bootstrap-vault-integration.md](./bootstrap-vault-integration.md)** - Called after core components are ready

### Monitoring and Validation
- **[health-check.md](./health-check.md)** - Comprehensive status checking after bootstrap

## Design Principles

*This script enables **GitOps-first infrastructure** - all cluster lifecycle management is automated through declarative configuration and continuous deployment.*