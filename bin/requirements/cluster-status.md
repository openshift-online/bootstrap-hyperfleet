# bin/cluster-status Requirements

## Purpose

The `cluster-status` tool compares ACM ManagedClusters with repository cluster configurations to identify mismatches, stuck terminations, orphaned resources, and ArgoCD Application inconsistencies across the OpenShift Bootstrap environment.

## Functional Requirements

### Primary Objective

Perform comprehensive cross-validation between multiple cluster management systems:
1. **Repository Configuration**: Discover clusters from `clusters/` and `regions/` directories
2. **ACM ManagedClusters**: Query OpenShift ACM for managed cluster state
3. **ArgoCD Applications**: Check GitOps application deployment status
4. **Namespace State**: Validate cluster namespace health and termination status

### Discovery Strategy

#### Repository Cluster Discovery
```bash
# Scan clusters/ directory for deployed clusters
find clusters/ -mindepth 1 -maxdepth 1 -type d

# Scan regions/ directory for regional specifications  
find regions/ -mindepth 2 -maxdepth 2 -type d
```

#### ACM ManagedCluster Discovery
```bash
# Get all managed clusters except local-cluster
oc get managedclusters -o json | jq -r '.items[] | select(.metadata.name != "local-cluster") | .metadata.name'
```

#### ArgoCD Application Discovery
```bash
# Count applications related to each cluster
oc get applications.argoproj.io -A -o json | jq -r --arg cluster "$cluster_name" '.items[] | select(.metadata.name | contains($cluster)) | .metadata.name'
```

### Status Analysis Requirements

#### Cluster State Matrix
For each discovered cluster, collect and analyze:

**Repository State**:
- Configuration exists in `clusters/` directory
- Regional specification exists in `regions/` directory

**ACM State**:
- ManagedCluster resource exists
- ManagedCluster availability status (`True`/`False`/`Unknown`)
- Finalizers present on ManagedCluster
- Taints applied to ManagedCluster

**Namespace State**:
- Cluster namespace exists
- Namespace phase (`Active`/`Terminating`)
- Stuck namespace termination detection

**ArgoCD State**:
- Count of related ArgoCD applications
- Application sync and health status

### Issue Detection Logic

#### Issue Categories
1. **ORPHANED_MC**: ManagedCluster exists but no repository configuration
2. **MISSING_MC**: Repository configuration exists but no ManagedCluster
3. **STUCK_NS**: Namespace stuck in Terminating phase
4. **STUCK_FINALIZERS**: ManagedCluster has finalizers but unavailable
5. **TAINTED**: ManagedCluster has taints applied

#### Issue Detection Rules
```bash
# Orphaned ManagedCluster
if [[ "$mc_status" == "Exists" && "$repo_config" == "No" ]]; then
    issues+=("ORPHANED_MC")
fi

# Missing ManagedCluster
if [[ "$repo_config" == "Yes" && "$mc_status" == "Not Found" ]]; then
    issues+=("MISSING_MC")
fi

# Stuck namespace termination
if [[ "$ns_status" == "Terminating" ]]; then
    issues+=("STUCK_NS")
fi

# Stuck finalizers
if [[ "$mc_finalizers" == "Yes" && "$mc_available" != "True" ]]; then
    issues+=("STUCK_FINALIZERS")
fi

# Tainted clusters
if [[ "$mc_taint" != "None" ]]; then
    issues+=("TAINTED")
fi
```

### Output Format Requirements

#### Table Output (Default)
```
| Cluster | Repo Config | ManagedCluster | Available | Finalizers | Taints | Namespace | Argo Apps | Issues |
|---------|-------------|----------------|-----------|------------|--------|-----------|-----------|--------|
| eks-01  | Yes         | Exists         | True      | No         | None   | Active    | 3         | OK     |
| ocp-02  | Yes         | Exists         | Unknown   | Yes        | NoSelect | Terminating | 1       | STUCK_NS,STUCK_FINALIZERS,TAINTED |
```

#### JSON Output
```json
{
  "clusters": [
    {
      "name": "eks-01",
      "repository_config": true,
      "managed_cluster": {
        "exists": true,
        "available": "True",
        "has_finalizers": false,
        "taints": "None"
      },
      "namespace_status": "Active",
      "argocd_applications": 3,
      "issues": "OK"
    }
  ]
}
```

#### CSV Output
```csv
Cluster,RepoConfig,ManagedCluster,Available,Finalizers,Taints,Namespace,ArgoApps,Issues
eks-01,Yes,Exists,True,No,None,Active,3,OK
```

### Command Line Interface

#### Usage Pattern
```bash
./bin/cluster-status [OPTIONS]
```

#### Options
- `--format FORMAT`: Output format (table, json, csv)
- `--issues-only`: Show only clusters with problems
- `--debug`: Enable debug output
- `--help`: Show usage information

#### Examples
```bash
# Basic cluster status comparison
./bin/cluster-status

# Show only problematic clusters
./bin/cluster-status --issues-only

# JSON output for automation
./bin/cluster-status --format json

# CSV export
./bin/cluster-status --format csv > cluster-status.csv
```

### Dependency Requirements

#### Required Tools
- `oc`: OpenShift CLI for cluster access
- `jq`: JSON processing for API responses

#### Required Permissions
- Read access to ManagedClusters
- Read access to ArgoCD Applications
- Read access to Namespaces

#### Connection Requirements
- Must be authenticated to hub cluster (`oc whoami`)
- Hub cluster must have ACM installed
- ArgoCD must be accessible (optional, graceful degradation)

### Error Handling Requirements

#### Connection Failures
```bash
# Check OpenShift authentication
if ! oc whoami >/dev/null 2>&1; then
    echo "❌ Error: Not connected to OpenShift cluster"
    echo "   Run: oc login"
    exit 1
fi
```

#### ACM Availability
```bash
# Graceful handling of missing ACM
if ! oc get managedclusters >/dev/null 2>&1; then
    echo "Warning: Cannot access ManagedClusters (ACM not available?)"
    # Continue with repository-only analysis
fi
```

#### ArgoCD Availability
```bash
# Optional ArgoCD integration
if ! oc get applications.argoproj.io -A >/dev/null 2>&1; then
    # Set ArgoCD app count to "N/A" instead of failing
    argo_apps="N/A"
fi
```

### Remediation Guidance

#### Cleanup Recommendations
The tool must provide actionable cleanup commands for each issue type:

**Orphaned ManagedClusters**:
```bash
oc delete managedcluster <cluster-name>
```

**Stuck Finalizers**:
```bash
oc patch managedcluster <cluster-name> --type=merge -p '{"metadata":{"finalizers":[]}}'
```

**Stuck Namespaces**:
```bash
# Check for remaining resources with finalizers
oc get all,rolebindings,secrets -n <cluster-name>
```

**Tainted Clusters**:
```bash
# Review cluster health and remove taints if appropriate
oc patch managedcluster <cluster-name> --type=json -p='[{"op": "remove", "path": "/spec/taints"}]'
```

### Integration Requirements

#### Status Summary
Generate summary statistics:
- Total clusters discovered
- Clusters with issues count
- Healthy clusters count
- Most common issue types

#### Automation Support
- JSON output for CI/CD pipeline integration
- CSV output for reporting and analysis
- Exit codes for automation (0 = healthy, 1 = issues found)

#### Monitoring Integration
- Structured output suitable for monitoring system ingestion
- Consistent issue classification for alerting rules
- Timestamp information for trend analysis

### Performance Requirements

#### Execution Time
- Must complete analysis within 60 seconds for typical deployments
- Parallel API queries where possible to reduce latency
- Caching of repeated API calls within single execution

#### Resource Usage
- Minimal memory footprint (shell script with standard tools)
- No persistent storage requirements
- Safe for automated execution

### Security Requirements

#### Credential Handling
- Use existing `oc` authentication (no credential storage)
- Respect RBAC permissions (graceful degradation for denied access)
- No sensitive information in output logs

#### Cluster Access
- Read-only operations only (no cluster modifications)
- Safe execution on production hub clusters
- No interference with cluster operations

## Related Tools

### Prerequisites
- **OpenShift CLI**: Authentication and cluster access
- **ACM**: ManagedCluster resource availability

### Workflow Integration
- **[health-check.md](./health-check.md)** - Complementary cluster health monitoring
- **[status.md](./status.md)** - Overall environment status tracking

### Remediation Tools
- **[clean-aws.md](./clean-aws.md)** - AWS resource cleanup for orphaned clusters
- **[remove-cluster.md](./remove-cluster.md)** - Proper cluster decommissioning

### Cluster Management
- **[generate-cluster.md](./generate-cluster.md)** - Creates cluster configurations tracked by this tool
- **[list-clusters.md](./list-clusters.md)** - Basic cluster enumeration functionality