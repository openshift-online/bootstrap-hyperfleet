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

**GitOps Layer Issues**:
1. **ORPHANED_MC**: ManagedCluster exists but no repository configuration
2. **MISSING_MC**: Repository configuration exists but no ManagedCluster
3. **STUCK_NS**: Namespace stuck in Terminating phase
4. **STUCK_FINALIZERS**: ManagedCluster has finalizers but unavailable
5. **TAINTED**: ManagedCluster has taints applied

**ArgoCD Issues** (NEW):
6. **STUCK_ARGOCD_APPS**: Applications stuck in deletion with finalizers
7. **STUCK_ARGOCD_APPSETS**: ApplicationSets stuck in deletion with finalizers
8. **ARGOCD_SYNC_FAILED**: Applications failing to sync or out of sync
9. **ARGOCD_FINALIZER_STUCK**: ArgoCD finalizers preventing resource cleanup

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

**ArgoCD Issues** (NEW):

**Stuck ArgoCD Applications** (Hierarchical Cleanup - Dependencies First):
```bash
# STEP 1: Identify stuck applications and their dependencies
oc get applications.argoproj.io -A -o json | jq '.items[] | select(.metadata.deletionTimestamp) | {name: .metadata.name, namespace: .metadata.namespace, finalizers: .metadata.finalizers, ownerReferences: .metadata.ownerReferences}'

# STEP 2: Remove dependent resources first (if any)
# Check for dependent resources created by the application
oc get all -l app.kubernetes.io/instance=<cluster-name> -A --show-kind --ignore-not-found

# STEP 3: Remove ArgoCD finalizer to allow application deletion
oc patch application <cluster-name>-cluster -n openshift-gitops --type=json -p='[{"op": "remove", "path": "/metadata/finalizers", "value": ["resources-finalizer.argocd.argoproj.io"]}]'

# STEP 4: Force delete if still stuck
oc delete application <cluster-name>-cluster -n openshift-gitops --force --grace-period=0
```

**Stuck ArgoCD ApplicationSets** (Parent Resource Cleanup):
```bash
# STEP 1: List applications created by the ApplicationSet
oc get applications.argoproj.io -A -l argocd.argoproj.io/application-set-name=<cluster-name>-applications

# STEP 2: Delete all dependent applications first
oc delete applications.argoproj.io -A -l argocd.argoproj.io/application-set-name=<cluster-name>-applications --wait=false

# STEP 3: Remove finalizers from stuck applications
for app in $(oc get applications.argoproj.io -A -l argocd.argoproj.io/application-set-name=<cluster-name>-applications -o name); do
    oc patch $app --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
done

# STEP 4: Remove ApplicationSet finalizers
oc patch applicationset <cluster-name>-applications -n openshift-gitops --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'

# STEP 5: Force delete ApplicationSet
oc delete applicationset <cluster-name>-applications -n openshift-gitops --force --grace-period=0
```

**Complete ArgoCD Cleanup Procedure** (Full Dependency Chain):
```bash
#!/bin/bash
# Complete ArgoCD resource cleanup for stuck cluster
CLUSTER_NAME="$1"

echo "🔍 Step 1: Identifying ArgoCD resources for $CLUSTER_NAME..."
oc get applications.argoproj.io,applicationsets.argoproj.io -A | grep $CLUSTER_NAME

echo "🧹 Step 2: Stopping ApplicationSet reconciliation..."
oc patch applicationset ${CLUSTER_NAME}-applications -n openshift-gitops --type=merge -p '{"spec":{"syncPolicy":{"preserveResourcesOnDeletion":true}}}'

echo "🗑️  Step 3: Removing dependent applications..."
for app in $(oc get applications.argoproj.io -A -o json | jq -r --arg cluster "$CLUSTER_NAME" '.items[] | select(.metadata.name | contains($cluster)) | "\(.metadata.namespace)/\(.metadata.name)"'); do
    namespace=$(echo $app | cut -d'/' -f1)
    name=$(echo $app | cut -d'/' -f2)
    echo "  Removing application: $name"
    oc patch application $name -n $namespace --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]' 2>/dev/null || true
    oc delete application $name -n $namespace --ignore-not-found=true --wait=false
done

echo "🗑️  Step 4: Removing ApplicationSet..."
oc patch applicationset ${CLUSTER_NAME}-applications -n openshift-gitops --type=json -p='[{"op": "remove", "path": "/metadata/finalizers"}]' 2>/dev/null || true
oc delete applicationset ${CLUSTER_NAME}-applications -n openshift-gitops --ignore-not-found=true --wait=false

echo "✅ Step 5: Verifying cleanup..."
sleep 5
remaining=$(oc get applications.argoproj.io,applicationsets.argoproj.io -A | grep $CLUSTER_NAME | wc -l)
if [[ $remaining -eq 0 ]]; then
    echo "✅ All ArgoCD resources cleaned up successfully"
else
    echo "⚠️  $remaining ArgoCD resources still remain - may need manual intervention"
    oc get applications.argoproj.io,applicationsets.argoproj.io -A | grep $CLUSTER_NAME
fi
```

**ArgoCD Sync Failures**:
```bash
# Check application sync status and errors
oc get application <cluster-name>-cluster -n openshift-gitops -o yaml | grep -A 20 "status:"

# Force application refresh and sync
oc patch application <cluster-name>-cluster -n openshift-gitops --type=json -p='[{"op": "replace", "path": "/operation", "value": {"sync": {"revision": "HEAD", "prune": true}}}]'

# Reset application if stuck
oc patch application <cluster-name>-cluster -n openshift-gitops --type=json -p='[{"op": "remove", "path": "/operation"}]'
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

### Health Check Implementation Requirements

#### Infrastructure Health Functions
```bash
check_infrastructure_health() {
    # Compare actual vs expected worker nodes
    local expected_workers=$(oc get machinepool -n $cluster $cluster-worker -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    local actual_workers=$(oc --kubeconfig=$cluster get nodes --no-headers 2>/dev/null | grep -c worker || echo "0")
    
    # Check node readiness
    local ready_workers=$(oc --kubeconfig=$cluster get nodes --no-headers 2>/dev/null | grep worker | grep -c Ready || echo "0")
    
    # CAPI cluster status for EKS
    local capi_ready=$(oc get cluster.cluster.x-k8s.io $cluster -n $cluster -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "N/A")
    
    # Hive cluster status for OCP
    local hive_ready=$(oc get clusterdeployment $cluster -n $cluster -o jsonpath='{.status.conditions[?(@.type=="ClusterReadyCondition")].status}' 2>/dev/null || echo "N/A")
    
    echo "$expected_workers/$actual_workers/$ready_workers/$capi_ready/$hive_ready"
}

check_platform_health() {
    # Count degraded cluster operators
    local degraded_cos=$(oc --kubeconfig=$cluster get co --no-headers 2>/dev/null | awk '$3 != "True" || $4 != "False" || $5 != "False" {count++} END {print count+0}' || echo "N/A")
    
    # Check critical services
    local api_health=$(oc --kubeconfig=$cluster get co kube-apiserver -o jsonpath='{.status.conditions[?(@.type=="Available")].status}' 2>/dev/null || echo "Unknown")
    local etcd_health=$(oc --kubeconfig=$cluster get co etcd -o jsonpath='{.status.conditions[?(@.type=="Available")].status}' 2>/dev/null || echo "Unknown")
    
    echo "$degraded_cos/$api_health/$etcd_health"
}

check_sync_wave_health() {
    # Find highest sync wave with applications
    local max_wave=0
    local current_wave=0
    
    for wave in $(seq 1 10); do
        local wave_apps=$(oc get applications.argoproj.io -A -o json 2>/dev/null | jq -r --arg cluster "$cluster" --arg wave "$wave" '.items[] | select(.metadata.name | contains($cluster)) | select(.metadata.annotations["argocd.argoproj.io/sync-wave"] == $wave) | .metadata.name' 2>/dev/null | wc -l || echo "0")
        if [[ $wave_apps -gt 0 ]]; then
            max_wave=$wave
            local synced_apps=$(oc get applications.argoproj.io -A -o json 2>/dev/null | jq -r --arg cluster "$cluster" --arg wave "$wave" '.items[] | select(.metadata.name | contains($cluster)) | select(.metadata.annotations["argocd.argoproj.io/sync-wave"] == $wave) | select(.status.sync.status == "Synced") | .metadata.name' 2>/dev/null | wc -l || echo "0")
            if [[ $synced_apps -eq $wave_apps ]]; then
                current_wave=$wave
            fi
        fi
    done
    
    echo "$current_wave/$max_wave"
}

check_workload_health() {
    # Count pending pods
    local pending_pods=$(oc --kubeconfig=$cluster get pods -A --field-selector=status.phase=Pending --no-headers 2>/dev/null | wc -l || echo "N/A")
    
    # Check PVC binding issues
    local pending_pvcs=$(oc --kubeconfig=$cluster get pvc -A --no-headers 2>/dev/null | grep -c Pending || echo "0")
    
    echo "$pending_pods/$pending_pvcs"
}

check_external_dependencies() {
    # External Secrets health
    local external_secrets_failed=$(oc --kubeconfig=$cluster get externalsecrets -A -o json 2>/dev/null | jq '.items[] | select(.status.conditions[] | select(.type == "Ready" and .status != "True"))' 2>/dev/null | jq length || echo "N/A")
    
    # Test basic connectivity (simplified)
    local connectivity="OK"
    if ! oc --kubeconfig=$cluster get nodes >/dev/null 2>&1; then
        connectivity="FAILED"
    fi
    
    echo "$external_secrets_failed/$connectivity"
}
```

### Cluster Management
- **[generate-cluster.md](./generate-cluster.md)** - Creates cluster configurations tracked by this tool
- **[list-clusters.md](./list-clusters.md)** - Basic cluster enumeration functionality