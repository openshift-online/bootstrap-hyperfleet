# Status Checking Knowledge Base

## Overview

This document consolidates all status checking knowledge for the OpenShift Bootstrap multi-cluster environment, including tools, techniques, and troubleshooting procedures for monitoring GitOps deployments, cluster provisioning, and fleet management.

## Status Checking Tools

### 1. Built-in Status Scripts

#### `status.sh` - CRD Establishment Monitor
```bash
Usage: ./status.sh <crd-name> [timeout-in-seconds]
```

**Purpose**: Waits for Kubernetes CustomResourceDefinitions to be established
**Default timeout**: 120 seconds
**Check interval**: 5 seconds

**Examples**:
```bash
./status.sh applications.argoproj.io 300
./status.sh clusters.cluster.x-k8s.io 180
```

**What it checks**:
- CRD existence: `kubectl get crd <crd-name>`
- Establishment status: `'{.status.conditions[?(@.type=="Established")].status}'`
- Returns success when status is "True"

#### `wait.kube.sh` - Generic Resource Condition Waiter
```bash
Usage: ./wait.kube.sh <type> <name> <namespace> <jsonpath> <expected-value> [timeout]
```

**Purpose**: Waits for any Kubernetes resource to meet a specific condition
**Default timeout**: 1800 seconds (30 minutes)
**Check interval**: 60 seconds

**Examples**:
```bash
# Wait for ArgoCD server route
./wait.kube.sh route openshift-gitops-server openshift-gitops '{.kind}' Route

# Wait for ACM MultiClusterHub completion
./wait.kube.sh mch multiclusterhub open-cluster-management \
  '{.status.conditions[?(@.type=="Complete")].message}' "All hub components ready."

# Wait for cluster provisioning
./wait.kube.sh cd cluster-40 cluster-40 \
  '{.status.conditions[?(@.type=="Provisioned")].message}' "Cluster is provisioned"
```

### 2. CAPI Monitoring Tools

#### `bin/capi-status` - CAPI Cluster Status Monitor
```bash
Usage: ./bin/capi-status [--format=dashboard|json|table] [--cluster=name] [--watch]
```

**Formats**:
- `dashboard`: Interactive visual progress display
- `table`: Tabular output for scripting
- `json`: Machine-readable output
- `watch`: Continuous monitoring

**Monitors**:
- Cluster API (CAPI) resources
- Control plane provisioning progress
- Worker node status
- Failure detection and messages

**Examples**:
```bash
# Dashboard view
./bin/capi-status --format=dashboard

# Monitor specific cluster
./bin/capi-status --cluster=cluster-43 --watch

# JSON output for automation
./bin/capi-status --format=json | jq '.clusters[] | select(.progress < 100)'
```

#### `bin/capi-dashboard` - Live CAPI Dashboard
```bash
Usage: ./bin/capi-dashboard [--cluster=name] [--interval=seconds]
```

**Features**:
- Real-time progress bars
- Interactive controls (q to quit, r to refresh)
- Color-coded status indicators
- Cluster details (region, instance type, age, nodes)

**Controls**:
- `q` or `Ctrl+C`: Quit
- `r`: Refresh immediately

#### `bin/cluster-health` - Multi-Cluster Health Monitor
```bash
Usage: ./bin/cluster-health [--summary] [--cluster=name]
```

**Monitors**:
- CAPI clusters (EKS, AKS, GKE)
- OpenShift ClusterDeployments
- ACM ManagedClusters

**Status Types**:
- ✓ **Healthy**: Fully provisioned and ready
- ⚠ **Degraded**: Exists but components not ready
- 🔄 **Provisioning**: CAPI cluster being created
- 📦 **Installing**: OpenShift cluster being installed
- ⏳ **Pending**: Created but installation not started
- 😴 **Hibernating**: Powered down to save costs
- ❌ **Failed**: Provisioning/installation failed

**Exit Codes**:
- `0`: All clusters healthy
- `1`: Some clusters degraded
- `2`: Some clusters failed

### 3. ArgoCD Status Checking

#### Application Status Commands
```bash
# Check all applications
oc get applications -n openshift-gitops

# Check specific application status
oc get application advanced-cluster-management -n openshift-gitops -o yaml

# Check application sync status
oc get application advanced-cluster-management -n openshift-gitops \
  -o jsonpath='{.status.sync.status}'

# Check application health
oc get application advanced-cluster-management -n openshift-gitops \
  -o jsonpath='{.status.health.status}'
```

#### ArgoCD Server Status
```bash
# Check ArgoCD pods
oc get pods -n openshift-gitops

# Check ArgoCD server route
oc get route -n openshift-gitops openshift-gitops-server

# Check ArgoCD application controller logs
oc logs -n openshift-gitops openshift-gitops-application-controller-0 --tail=20
```

### 4. ACM Status Checking

#### MultiClusterHub Status
```bash
# Check MultiClusterHub status
oc get mch multiclusterhub -n open-cluster-management

# Check detailed status
oc get mch multiclusterhub -n open-cluster-management -o yaml

# Check MCH conditions
oc get mch multiclusterhub -n open-cluster-management \
  -o jsonpath='{.status.conditions[?(@.type=="Complete")].message}'
```

#### Managed Cluster Status
```bash
# List all managed clusters
oc get managedclusters

# Check specific cluster status
oc get managedcluster cluster-40 -o yaml

# Check cluster conditions
oc get managedcluster cluster-40 \
  -o jsonpath='{.status.conditions[?(@.type=="ManagedClusterConditionAvailable")].status}'
```

### 5. Operator Status Checking

#### Subscription Status
```bash
# Check GitOps operator subscription
oc get subscription openshift-gitops-operator -n openshift-operators

# Check all operator subscriptions
oc get subscriptions --all-namespaces

# Check CSV (ClusterServiceVersion) status
oc get csv --all-namespaces | grep gitops
```

#### Operator Pod Status
```bash
# Check GitOps operator pods
oc get pods -n openshift-gitops-operator

# Check ACM operator pods
oc get pods -n open-cluster-management

# Check specific operator logs
oc logs -n openshift-gitops-operator deployment/openshift-gitops-operator-controller-manager
```

## Status Checking Workflows

### 1. Bootstrap Process Status

#### Step 1: Prerequisites Check
```bash
# Check cluster connection
oc whoami

# Check GitOps operator subscription
oc get subscription openshift-gitops-operator -n openshift-operators

# Wait for GitOps CRDs
./status.sh applications.argoproj.io 300
```

#### Step 2: GitOps Deployment Status
```bash
# Check ArgoCD pods
oc get pods -n openshift-gitops

# Wait for ArgoCD server
./wait.kube.sh route openshift-gitops-server openshift-gitops '{.kind}' Route

# Check applications deployment
oc get applications -n openshift-gitops
```

#### Step 3: ACM Deployment Status
```bash
# Wait for ACM installation
./wait.kube.sh mch multiclusterhub open-cluster-management \
  '{.status.conditions[?(@.type=="Complete")].message}' "All hub components ready."

# Check ACM pods
oc get pods -n open-cluster-management
```

### 2. Cluster Provisioning Status

#### CAPI Cluster Monitoring
```bash
# Start live dashboard
./bin/capi-dashboard

# Check specific cluster
./bin/capi-status --cluster=cluster-43

# Monitor provisioning progress
./bin/capi-status --watch
```

#### OpenShift Cluster Monitoring
```bash
# Check cluster deployments
oc get clusterdeployments --all-namespaces

# Monitor specific cluster
oc get clusterdeployment cluster-10 -n cluster-10 -o yaml

# Check install job status
oc get job -n cluster-10 | grep install
```

### 3. Application Sync Status

#### ArgoCD Application Health
```bash
# Check all application sync status
oc get applications -n openshift-gitops \
  -o jsonpath='{.items[*].status.sync.status}'

# Check application health
oc get applications -n openshift-gitops \
  -o jsonpath='{.items[*].status.health.status}'

# Check failed applications
oc get applications -n openshift-gitops \
  -o jsonpath='{.items[?(@.status.sync.status=="Failed")].metadata.name}'
```

#### Sync Error Analysis
```bash
# Check application controller logs for errors
oc logs -n openshift-gitops openshift-gitops-application-controller-0 \
  | grep -i error

# Check specific application sync result
oc get application advanced-cluster-management -n openshift-gitops \
  -o jsonpath='{.status.operationState.message}'
```

## Common Status Checking Patterns

### 1. JSONPath Expressions

#### Resource Status
```bash
# Check readiness
'{.status.conditions[?(@.type=="Ready")].status}'

# Check completion
'{.status.conditions[?(@.type=="Complete")].message}'

# Check provisioning
'{.status.conditions[?(@.type=="Provisioned")].status}'

# Check available
'{.status.conditions[?(@.type=="Available")].status}'
```

#### Count Resources
```bash
# Count ready replicas
'{.status.readyReplicas}'

# Count desired replicas
'{.spec.replicas}'

# Count nodes
'{.status.replicas}'
```

### 2. Condition Checking

#### Standard Kubernetes Conditions
- `Ready`: Resource is ready for use
- `Available`: Resource is available
- `Progressing`: Resource is being updated
- `Failure`: Resource has failed

#### ACM Specific Conditions
- `Complete`: MultiClusterHub installation complete
- `ManagedClusterConditionAvailable`: Managed cluster is available
- `ManagedClusterJoined`: Cluster joined ACM

#### CAPI Specific Conditions
- `InfrastructureReady`: Infrastructure provisioned
- `ControlPlaneReady`: Control plane operational
- `NodesReady`: Worker nodes ready

### 3. Timeout Strategies

#### Progressive Timeouts
```bash
# Quick check (30s)
timeout 30s kubectl get pods

# Standard wait (5 minutes)
./wait.kube.sh deployment my-app default '{.status.readyReplicas}' "3" 300

# Long provisioning (30 minutes)
./wait.kube.sh cd cluster-40 cluster-40 \
  '{.status.conditions[?(@.type=="Provisioned")].message}' "Cluster is provisioned" 1800
```

## Troubleshooting Common Issues

### 1. CRD Not Found Errors

**Symptom**: `the server doesn't have a resource type "applications"`

**Diagnosis**:
```bash
# Check if CRD exists
oc get crd applications.argoproj.io

# Check CRD status
oc get crd applications.argoproj.io -o yaml
```

**Solution**:
```bash
# Wait for CRD establishment
./status.sh applications.argoproj.io 300
```

### 2. Application Sync Failures

**Symptom**: ArgoCD applications stuck in "Failed" state

**Diagnosis**:
```bash
# Check application status
oc get application advanced-cluster-management -n openshift-gitops -o yaml

# Check controller logs
oc logs -n openshift-gitops openshift-gitops-application-controller-0 --tail=50
```

**Common Causes**:
- Invalid resource configurations
- Missing CRDs
- Webhook validation failures
- Resource conflicts

### 3. Cluster Provisioning Failures

**Symptom**: Clusters stuck in "Provisioning" state

**Diagnosis**:
```bash
# Check cluster status
./bin/cluster-health

# Check CAPI resources
oc get cluster cluster-40 -n cluster-40 -o yaml
oc get awsmanagedcontrolplane cluster-40 -n cluster-40 -o yaml
```

**Common Causes**:
- AWS credential issues
- Resource quota limits
- Network configuration problems
- Invalid instance types

### 4. ACM Installation Issues

**Symptom**: MultiClusterHub fails to install

**Diagnosis**:
```bash
# Check MCH status
oc get mch multiclusterhub -n open-cluster-management -o yaml

# Check MCH operator logs
oc logs -n open-cluster-management deployment/multiclusterhub-operator
```

**Common Causes**:
- Invalid component configurations
- Resource constraints
- Network policies
- Storage issues

## Monitoring Best Practices

### 1. Automated Status Checks

#### Health Check Scripts
```bash
#!/bin/bash
# comprehensive-health-check.sh

echo "=== GitOps Status ==="
oc get applications -n openshift-gitops

echo "=== Cluster Health ==="
./bin/cluster-health --summary

echo "=== ACM Status ==="
oc get mch multiclusterhub -n open-cluster-management \
  -o jsonpath='{.status.phase}'

echo "=== Managed Clusters ==="
oc get managedclusters
```

#### Continuous Monitoring
```bash
# Watch applications continuously
watch -n 30 'oc get applications -n openshift-gitops'

# Monitor cluster provisioning
./bin/capi-dashboard --interval=10
```

### 2. Alert Thresholds

#### Critical Alerts
- Application sync failures > 5 minutes
- Cluster provisioning stuck > 30 minutes
- ACM MultiClusterHub not ready > 10 minutes

#### Warning Alerts
- High resource utilization > 80%
- Long sync times > 2 minutes
- Certificate expiration < 30 days

### 3. Status Reporting

#### Daily Reports
```bash
# Generate daily status report
{
  echo "# Daily Fleet Status Report - $(date)"
  echo
  echo "## Cluster Health"
  ./bin/cluster-health --summary
  echo
  echo "## Application Status"
  oc get applications -n openshift-gitops
  echo
  echo "## ACM Status"
  oc get mch multiclusterhub -n open-cluster-management
} > daily-report-$(date +%Y%m%d).md
```

#### Metrics Export
```bash
# Export metrics for monitoring systems
./bin/capi-status --format=json | jq -r '.clusters[] | 
  "\(.name),\(.status),\(.progress),\(.region)"' > cluster-metrics.csv
```

## Integration with External Systems

### 1. Prometheus Metrics

#### Custom Metrics
```bash
# Cluster health metrics
cluster_health_status{cluster="cluster-40",type="eks"} 1

# Application sync metrics
argocd_app_sync_total{name="advanced-cluster-management",status="success"} 1
```

### 2. Alert Manager

#### Alert Rules
```yaml
- alert: ClusterProvisioningStuck
  expr: cluster_provisioning_duration_seconds > 1800
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Cluster {{ $labels.cluster }} stuck provisioning"

- alert: ApplicationSyncFailure
  expr: argocd_app_sync_total{status="failed"} > 0
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Application {{ $labels.name }} sync failed"
```

### 3. Dashboard Integration

#### Grafana Dashboards
- Cluster provisioning progress
- Application sync status
- Resource utilization metrics
- Error rate trends

## Status Checking Reference

### Quick Commands
```bash
# Overall health
./bin/cluster-health --summary

# GitOps status
oc get applications -n openshift-gitops

# ACM status
oc get mch multiclusterhub -n open-cluster-management

# CAPI status
./bin/capi-status --format=table

# Cluster provisioning
./bin/capi-dashboard
```

### Emergency Procedures
```bash
# Reset failed application
oc patch application advanced-cluster-management -n openshift-gitops \
  --type merge -p '{"operation":null}'

# Force sync application
oc patch application advanced-cluster-management -n openshift-gitops \
  --type merge -p '{"spec":{"syncPolicy":{"automated":{"prune":true}}}}'

# Restart ArgoCD
oc rollout restart deployment/openshift-gitops-server -n openshift-gitops
```

This comprehensive status checking knowledge base provides all the tools, techniques, and troubleshooting procedures needed to effectively monitor and maintain the OpenShift Bootstrap multi-cluster environment.