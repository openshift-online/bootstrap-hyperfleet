# bin/bootstrap.vault-integration.sh Requirements

## Purpose

The `bootstrap.vault-integration.sh` script automates Vault-based secret management setup for all discovered cluster configurations, creating ExternalSecrets that enable secure credential distribution to cluster namespaces.

## Functional Requirements

### Prerequisites Validation
- **Vault Pod Verification**: Must verify `vault-0` pod exists in `vault` namespace
- **ClusterSecretStore Check**: Must verify `vault-cluster-store` ClusterSecretStore exists
- **Secret Validation**: Must verify required secrets exist in Vault:
  - `secret/aws-credentials` - AWS access credentials
  - `secret/pull-secret` - Container registry pull secrets
- **Exit Conditions**: Must exit with error code 1 if any prerequisite fails

### Discovery Requirements
- **Cluster Detection**: Must dynamically discover all cluster directories in `clusters/`
- **Name Extraction**: Must extract semantic cluster names from directory structure
- **Empty Handling**: Must gracefully handle empty clusters directory with warning

### Secret Management Setup

#### Per-Cluster Configuration
For each discovered cluster, must:
1. **Namespace Creation**: Create cluster namespace (ignore if exists)
2. **Service Account**: Create `vault-secret-reader` service account
3. **Token Secret**: Create service account token secret for Vault authentication
4. **ExternalSecrets**: Apply namespace-specific external secret configurations

#### ExternalSecret Requirements
- **Template Processing**: Must substitute `CLUSTER_NAMESPACE` placeholder with actual cluster name
- **Source Template**: Must use `bases/ocm/external-secrets-template.yaml`
- **Resource Creation**: Must create ExternalSecret resources for each cluster namespace

### Output Requirements

#### Progress Reporting
- **Color-coded Output**: Use consistent color scheme (green=success, red=error, yellow=warning)
- **Step-by-step**: Clear messaging for each setup phase per cluster
- **Discovery Results**: List all discovered clusters before processing

#### Status Information
- **Sync Conditions**: Explain when secrets will automatically sync
- **Monitoring Commands**: Provide specific commands for status checking
- **Troubleshooting**: Include debugging commands for common issues

### Error Handling Requirements

#### Prerequisite Failures
- **Missing Vault**: Clear error message with pod status check
- **Missing ClusterSecretStore**: Specific instructions for manual setup
- **Missing Secrets**: Instructions for storing secrets in Vault

#### Resource Creation Errors
- **Namespace Conflicts**: Use `2>/dev/null || true` pattern for idempotent operations
- **Service Account Exists**: Handle existing service accounts gracefully
- **YAML Application**: Provide clear feedback on ExternalSecret creation

### Integration Requirements

#### Vault Integration
- **Authentication**: Use Kubernetes service account tokens for Vault auth
- **Secret Paths**: Follow standardized Vault secret paths (`secret/aws-credentials`, `secret/pull-secret`)
- **Refresh Interval**: Configure 1-hour secret refresh interval

#### External Secrets Operator (ESO)
- **Deferred Sync**: Secrets sync only when ESO is ready and target namespaces exist
- **ClusterSecretStore**: Rely on `vault-cluster-store` for Vault connectivity
- **Status Monitoring**: Provide commands for ExternalSecret status checking

### Usage Patterns

```bash
# Called automatically by bootstrap.sh
./bin/bootstrap.vault-integration.sh

# Manual execution for new clusters
./bin/bootstrap.vault-integration.sh

# Verify prerequisites first
oc get pod vault-0 -n vault && ./bin/bootstrap.vault-integration.sh
```

## Dependencies

### External Resources
- **Vault Pod**: `vault-0` in `vault` namespace must be running
- **ClusterSecretStore**: `vault-cluster-store` must be configured
- **Template File**: `bases/ocm/external-secrets-template.yaml`

### Required Secrets in Vault
- `secret/aws-credentials` - AWS access keys for cluster provisioning
- `secret/pull-secret` - Container registry authentication

### Kubernetes Permissions
- **Namespace Management**: Create and manage cluster namespaces
- **Service Account Creation**: Create vault-secret-reader service accounts
- **Secret Management**: Create token secrets for Vault authentication
- **ExternalSecret CRDs**: Apply ExternalSecret custom resources

## Output Format

### Success Summary
```
🎉 All Vault-based cluster secrets setup complete!

📋 Summary:
  • Vault ClusterSecretStore: vault-cluster-store
  • Secret refresh interval: 1 hour
  • Managed secrets: aws-credentials, pull-secret

🔍 Monitoring:
  • Check ExternalSecret status: oc get externalsecret -A
  • Check secret sync status: oc describe externalsecret <name> -n <namespace>
  • View Vault secrets: oc exec vault-0 -n vault -- vault kv list secret/

🔧 Troubleshooting:
  • If secrets fail to sync, check Vault auth: oc logs -n external-secrets deployment/eso-external-secrets
  • Vault UI access: oc port-forward -n vault vault-0 8200:8200 (token: root)
```

## Related Tools

### Prerequisites
- **[bootstrap.md](./bootstrap.md)** - Must run before this tool to establish core components

### Dependencies
- **[status.md](./status.md)** - May use for CRD readiness verification
- **[wait-kube.md](./wait-kube.md)** - May use for resource readiness verification

### Workflow Integration
- **[health-check.md](./health-check.md)** - Can verify secret setup status after this tool runs

## Design Principles

*This script enables **secure credential distribution** - all sensitive data flows through Vault with automatic rotation and namespace isolation.*