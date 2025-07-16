# CAPI Provisioning Dashboard

## Overview

The CAPI Provisioning Dashboard provides real-time visibility into cluster provisioning status for the OpenShift Bootstrap multi-cluster environment. It monitors CAPI (Cluster API) resources to track EKS cluster creation progress.

## Features

- **Real-time Progress Tracking**: Visual progress bars showing cluster provisioning stages
- **Multi-Format Output**: Dashboard, table, and JSON formats
- **Cluster Health Monitoring**: Status of control planes, worker nodes, and overall cluster health
- **Failure Detection**: Automatic detection of stuck or failed provisioning
- **CLI-Based Interface**: No external dependencies, runs entirely in terminal

## Installation

The dashboard scripts are included in the `bin/` directory:

```bash
# Make scripts executable (if not already)
chmod +x bin/capi-status bin/capi-dashboard bin/cluster-health
```

## Usage

### 1. Live Dashboard

Start the interactive dashboard with real-time updates:

```bash
./bin/capi-dashboard
```

**Options:**
- `--cluster=name`: Filter by specific cluster
- `--interval=seconds`: Refresh interval (default: 5 seconds)

**Controls:**
- `q` or `Ctrl+C`: Quit dashboard
- `r`: Refresh immediately

### 2. Static Status Check

Get current cluster status in various formats:

```bash
# Dashboard format (single snapshot)
./bin/capi-status --format=dashboard

# Table format
./bin/capi-status --format=table

# JSON format (for scripting)
./bin/capi-status --format=json

# Filter specific cluster
./bin/capi-status --cluster=cluster-43

# Continuous monitoring
./bin/capi-status --watch
```

### 3. Quick Health Check

Fast summary of cluster health:

```bash
# Detailed health check
./bin/cluster-health

# Summary only
./bin/cluster-health --summary

# Specific cluster
./bin/cluster-health --cluster=cluster-43
```

## Dashboard Output

### Live Dashboard View

```
┌─ CAPI Cluster Provisioning Dashboard ────────────────────────────────────┐
│ 2025-07-16 14:30:15                                                       │
│ Press 'q' to quit, 'r' to refresh                                        │
└───────────────────────────────────────────────────────────────────────────┘

🔄 cluster-43 (cluster-43)
  [████████░░░░░░░░░░░░] 40% - Creating worker nodes
  📍 us-west-2 | 💻 m5.xlarge | 🕐 5m | 🖥️ 1/3 nodes

✓ cluster-41 (cluster-41)
  [████████████████████] 100% - Ready
  📍 us-west-2 | 💻 m5.large | 🕐 2d | 🖥️ 3/3 nodes

❌ cluster-44 (cluster-44)
  [XXXXXXXXXXXXXXXXXXXXXX] Failed - Invalid subnet configuration
  📍 us-east-1 | 💻 m5.large | 🕐 30m | 🖥️ 0/3 nodes

────────────────────────────────────────────────────────────────────────────
Next refresh in 5s | Press 'q' to quit, 'r' to refresh now
```

### Table Format

```
CLUSTER      NAMESPACE    PROGRESS  STATUS                 AGE   NODES  REGION       INSTANCE
────────────────────────────────────────────────────────────────────────────────────────
cluster-41   cluster-41   100%      Ready                  2d    3/3    us-west-2    m5.large
cluster-42   cluster-42   100%      Ready                  1d    5/5    ap-se-1      m5.xlarge
cluster-43   cluster-43   40%       Creating worker nodes  5m    1/3    us-west-2    m5.xlarge
cluster-44   cluster-44   Failed    Invalid subnet config  30m   0/3    us-east-1    m5.large
```

### Health Check Output

```
Checking cluster health...

CLUSTER         NAMESPACE       STATUS       DETAILS
─────────────────────────────────────────────────────────────────
cluster-41      cluster-41      healthy      3/3 nodes, us-west-2
cluster-42      cluster-42      healthy      5/5 nodes, ap-southeast-1
cluster-43      cluster-43      provisioning 1/3 nodes, us-west-2
cluster-44      cluster-44      failed       0/3 nodes, us-east-1

Summary:
  ✓ Healthy:      2
  ⚠ Degraded:     0
  🔄 Provisioning: 1
  ❌ Failed:       1
  📊 Total:        4
```

## Monitored Resources

The dashboard monitors these CAPI resources:

### Core Resources
- **Cluster** (`cluster.x-k8s.io/v1beta1`): Overall cluster coordination
- **AWSManagedControlPlane** (`controlplane.cluster.x-k8s.io/v1beta1`): EKS control plane
- **AWSManagedMachinePool** (`infrastructure.cluster.x-k8s.io/v1beta1`): Worker node pools

### Status Fields
- `status.phase`: Cluster lifecycle phase (Provisioning, Provisioned, Failed, Deleting)
- `status.ready`: Resource readiness indicators
- `status.conditions`: Detailed condition information
- `status.failureMessage`: Error messages for failed provisioning
- `status.replicas`: Current vs desired node counts

## Progress Calculation

The dashboard calculates provisioning progress based on:

1. **0-10%**: Cluster resource created
2. **10-30%**: Infrastructure provisioning started
3. **30-60%**: Control plane ready
4. **60-80%**: Control plane accessible
5. **80-95%**: Worker nodes provisioning
6. **95-100%**: All nodes ready and cluster operational

## Error Detection

The dashboard automatically detects:

- **Stuck Provisioning**: Clusters not progressing for extended periods
- **Failed Clusters**: Clusters with failure messages or Failed phase
- **Resource Issues**: Node quota exceeded, invalid configurations
- **Network Problems**: Subnet, security group, or VPC issues

## Integration

### With Existing Scripts

The dashboard integrates with existing bootstrap scripts:

```bash
# Use with wait scripts
./wait.kube.sh cluster cluster-43 cluster-43 '{.status.phase}' "Provisioned" 1800

# Monitor during bootstrap
./bootstrap.sh &
./bin/capi-dashboard --cluster=new-cluster
```

### With CI/CD

JSON output format enables scripting integration:

```bash
# Check if all clusters are healthy
./bin/cluster-health --summary
exit_code=$?

if [[ $exit_code -eq 0 ]]; then
    echo "All clusters healthy"
elif [[ $exit_code -eq 1 ]]; then
    echo "Some clusters degraded"
else
    echo "Critical failures detected"
fi
```

### With Monitoring Systems

Export metrics for external monitoring:

```bash
# Export to monitoring system
./bin/capi-status --format=json | jq '.clusters[] | select(.progress < 100)' | monitoring-system-import
```

## Troubleshooting

### Common Issues

1. **No clusters found**: Check if CAPI resources exist
   ```bash
   kubectl get clusters --all-namespaces
   ```

2. **Permission errors**: Ensure proper RBAC access
   ```bash
   kubectl auth can-i get clusters
   ```

3. **Script not found**: Run from bootstrap repository root
   ```bash
   ls -la bin/capi-*
   ```

### Debug Mode

For troubleshooting, run with verbose output:

```bash
# Debug specific cluster
kubectl get cluster cluster-43 -o yaml
kubectl get awsmanagedcontrolplane cluster-43 -o yaml
kubectl get awsmanagedmachinepool cluster-43 -o yaml
```

## Customization

### Add New Cluster Types

To support additional cluster types (AKS, GKE), modify the `collect_cluster_data()` function in `bin/capi-status`:

```bash
# Add new resource types
local gke_ready=$(kubectl get gkemanagedcontrolplane "$name" -n "$namespace" -o jsonpath='{.status.ready}' 2>/dev/null || echo "false")
```

### Custom Progress Stages

Modify the `get_cluster_progress()` function to add custom provisioning stages:

```bash
# Add custom condition checks
local custom_ready=$(kubectl get customresource "$name" -n "$namespace" -o jsonpath='{.status.ready}' 2>/dev/null || echo "false")
```

## Future Enhancements

1. **Historical Tracking**: Store provisioning times and success rates
2. **Alerting Integration**: Send notifications on failures
3. **Cost Tracking**: Monitor cluster costs during provisioning
4. **Predictive Analytics**: Estimate completion times based on history
5. **Web Interface**: Browser-based dashboard for teams

## Security Considerations

- Dashboard requires read-only access to CAPI resources
- No sensitive data is logged or displayed
- Cluster credentials are not accessed or stored
- Safe to run in production environments

## Performance

- Minimal resource usage (bash scripts only)
- Efficient kubectl queries with specific JSONPath
- Configurable refresh intervals
- Scales to 100+ clusters

The CAPI Provisioning Dashboard provides essential visibility for managing multi-cluster EKS deployments with minimal overhead and maximum operational insight.