# OpenShift Bootstrap Cluster Provisioning Metrics & Dashboard

## Overview

This solution provides comprehensive end-to-end observability for the OpenShift Bootstrap cluster provisioning process, exposing 100% of success metrics from regional specification validation through complete cluster readiness.

## Architecture

### Metrics Collection Pipeline
```
Regional Specs → Kustomize Generation → GitOps Deployment → Sync Waves → Infrastructure → Platform Health
      ↓                ↓                      ↓              ↓             ↓              ↓
   File System    Validation Logs      ArgoCD APIs    Wave Progression   CAPI/Hive    Cluster APIs
      ↓                ↓                      ↓              ↓             ↓              ↓
                            Custom Metrics Exporter
                                       ↓
                                 Prometheus
                                       ↓
                                  Grafana Dashboard
```

### Components

1. **Custom Metrics Exporter** - Python application that collects metrics from multiple sources
2. **Prometheus Integration** - ServiceMonitor for automatic metrics scraping
3. **Alerting Rules** - PrometheusRule for automated failure detection
4. **Grafana Dashboard** - Comprehensive installation flow visualization

## Metrics Coverage

### 📊 End-to-End Provisioning Metrics
- **Total provisioning duration** (regional spec → fully operational cluster)
- **Phase-wise timing breakdown** (sync waves, infrastructure, platform)
- **Active provisioning cluster count**
- **Success rate trending** (1h, 24h, 7d windows)

### 🌊 ArgoCD Sync Wave Progression
- **Wave status tracking** (not_started, in_progress, completed, failed)
- **Wave duration measurements**
- **Application count by status** (total, synced, healthy per wave)
- **Cross-wave dependency tracking**

### 🏗️ Infrastructure Provisioning
- **CAPI cluster status** (EKS provisioning via Cluster API)
- **Hive deployment status** (OCP provisioning via Hive)
- **Worker node capacity** (expected vs actual vs ready)
- **Cluster API responsiveness**

### 🎯 Platform Health & Readiness
- **Cluster operator status** (available, progressing, degraded counts)
- **Core service health** (API server, etcd availability)
- **Overall readiness score** (composite 0-100 metric)

### 🔗 ACM Management Integration
- **ManagedCluster availability status**
- **Klusterlet agent connectivity**
- **Multi-cluster observability health**
- **Policy compliance tracking**

### 📈 Performance Analytics
- **Provisioning duration histograms**
- **Failure categorization** (infrastructure, platform, application)
- **Resource utilization during provisioning**
- **Geographic performance comparison**

### 🔍 External Dependencies
- **External Secrets synchronization status**
- **Vault connectivity health**
- **AWS credentials validation**
- **Network connectivity checks**

## Installation

### 1. Deploy the Metrics Exporter

```bash
# Apply the metrics collection components
oc apply -k operators/cluster-provisioning-metrics/global/

# Verify deployment
oc get pods -n openshift-gitops -l app=cluster-provisioning-metrics
oc get servicemonitor cluster-provisioning-metrics -n openshift-gitops
```

### 2. Verify Prometheus Scraping

```bash
# Check if Prometheus is scraping the metrics
oc exec -n openshift-monitoring prometheus-k8s-0 -c prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/targets' | \
  jq '.data.activeTargets[] | select(.labels.job=="cluster-provisioning-metrics")'

# Verify metrics are available
oc exec -n openshift-monitoring prometheus-k8s-0 -c prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/query?query=cluster_provisioning_active_total' | \
  jq '.data.result'
```

### 3. Import Grafana Dashboard

#### Option A: Via OpenShift Console
1. Navigate to **Observe → Dashboards**
2. Click **Import Dashboard**
3. Copy contents of `grafana-dashboard.json`
4. Paste and click **Import**

#### Option B: Via Grafana ConfigMap
```bash
# Create dashboard ConfigMap
oc create configmap cluster-provisioning-dashboard \
  --from-file=dashboard.json=operators/cluster-provisioning-metrics/global/grafana-dashboard.json \
  -n openshift-config-managed

# Label for auto-discovery
oc label configmap cluster-provisioning-dashboard \
  console.openshift.io/dashboard=true \
  -n openshift-config-managed
```

### 4. Configure Alerting (Optional)

```bash
# Alerts are automatically created with the PrometheusRule
# Verify alert rules are loaded
oc get prometheusrule cluster-provisioning-alerts -n openshift-gitops

# Check alert status in Prometheus
oc port-forward -n openshift-monitoring prometheus-k8s-0 9090:9090 &
# Navigate to http://localhost:9090/alerts
```

## Dashboard Guide

### 📊 Provisioning Overview
- **Active Provisioning Clusters** - Current clusters being provisioned
- **Success Rate (24h)** - Overall success percentage  
- **Mean Provisioning Time** - Average duration for successful clusters
- **Total Clusters by Status** - Distribution across all managed clusters

### 🚀 End-to-End Provisioning Flow
- **Cluster Provisioning Timeline** - Duration tracking for each cluster
- **Phase Duration Breakdown** - Time spent in each major phase

### 🌊 ArgoCD Sync Wave Progression
- **Sync Wave Status Matrix** - Heatmap showing wave progression across clusters
- **Sync Wave Applications Status** - Application sync/health status per wave

### 🏗️ Infrastructure Provisioning  
- **Infrastructure Status by Type** - CAPI vs Hive provisioning status
- **Worker Nodes Status** - Expected vs actual vs ready node counts
- **Cluster API Response Time** - API server responsiveness

### 🎯 Platform Health & Readiness
- **Cluster Operators Status** - OpenShift operator health tracking
- **Cluster Readiness Score** - Composite health score (0-100)

### 🔗 ACM Management & Dependencies
- **ManagedCluster Conditions** - ACM availability conditions
- **External Dependencies Health** - Vault, secrets, connectivity status

### 📈 Performance Analytics
- **Provisioning Duration Distribution** - Histogram of provisioning times
- **Failure Categories** - Breakdown of failure types and causes

### 🔍 Detailed Status Table
- **Multi-dimensional cluster status** - Readiness, ACM, infrastructure in tabular format

## Key Metrics Reference

### Critical Success Metrics

| Metric | Description | Success Criteria |
|--------|-------------|------------------|
| `cluster_provisioning_duration_seconds` | End-to-end provisioning time | < 45 minutes (green), < 75 minutes (yellow) |
| `cluster_sync_wave_status` | Sync wave progression | All waves = 2 (completed) |
| `cluster_infrastructure_status` | Infrastructure readiness | Status = "ready" |
| `cluster_readiness_score` | Overall health score | Score > 90 (green), > 70 (yellow) |
| `managed_cluster_status` | ACM management status | Status = "available" |

### Alert Conditions

| Alert | Trigger | Severity |
|-------|---------|----------|
| `ClusterProvisioningTakingTooLong` | > 75 minutes | Warning |
| `ClusterProvisioningFailed` | Any provisioning failure | Critical |
| `SyncWaveStuck` | Wave in progress > 30 minutes | Warning |
| `InfrastructureProvisioningFailed` | Infrastructure status = failed | Critical |
| `ClusterReadinessLow` | Readiness score < 80 | Warning |

## Troubleshooting

### Metrics Not Appearing

```bash
# Check exporter logs
oc logs -n openshift-gitops deployment/cluster-provisioning-metrics-exporter

# Verify ServiceMonitor target
oc get servicemonitor cluster-provisioning-metrics -n openshift-gitops -o yaml

# Check Prometheus targets
oc port-forward -n openshift-monitoring prometheus-k8s-0 9090:9090 &
# Navigate to http://localhost:9090/targets
```

### Dashboard Not Loading

```bash
# Verify Grafana can access Prometheus
oc get route grafana -n openshift-monitoring

# Check dashboard ConfigMap
oc get configmap cluster-provisioning-dashboard -n openshift-config-managed

# Verify dashboard labels
oc get configmap cluster-provisioning-dashboard -n openshift-config-managed -o yaml | grep labels -A 5
```

### Missing Cluster Data

```bash
# Check cluster discovery
oc get applications.argoproj.io -n openshift-gitops | grep -E "(ocp-|eks-)"

# Verify RBAC permissions
oc auth can-i get managedclusters --as=system:serviceaccount:openshift-gitops:cluster-provisioning-metrics

# Check custom resources availability
oc api-resources | grep -E "(cluster\.x-k8s\.io|hive\.openshift\.io|cluster\.open-cluster-management\.io)"
```

## Performance Baselines

### Expected Provisioning Times

| Cluster Type | Fast Path | Normal Path | Slow Path |
|--------------|-----------|-------------|-----------|
| **OCP (Hive)** | 30-40 min | 45-60 min | 60-90 min |
| **EKS (CAPI)** | 25-35 min | 40-55 min | 55-85 min |

### Sync Wave Expectations

| Wave | Component | Expected Duration |
|------|-----------|-------------------|
| **1** | Cluster Provisioning | 25-50 minutes |
| **2** | Operator Installation | 3-8 minutes |
| **3** | Pipeline Deployment | 2-5 minutes |
| **4** | Service Deployment | 3-8 minutes |

### Performance Indicators

- **🟢 Green**: < 45 minutes total, all waves sequential
- **🟡 Yellow**: 45-75 minutes, acceptable wave delays  
- **🔴 Red**: > 75 minutes or wave failures requiring intervention

## Extension Points

### Adding Custom Metrics

1. **Extend the Python exporter** - Add new collection methods in `metrics_exporter.py`
2. **Add new Prometheus metrics** - Define metrics in `_init_metrics()`
3. **Create collection logic** - Implement Kubernetes API queries
4. **Update dashboard** - Add new panels for custom metrics

### Integration with External Systems

```python
# Example: Adding custom dependency checks
def collect_custom_dependencies(self):
    """Collect metrics from external systems"""
    # Check custom registry health
    # Validate custom networking
    # Monitor custom storage systems
    pass
```

### Advanced Analytics

The collected metrics support advanced use cases:
- **Predictive failure analysis** using duration trends
- **Capacity planning** based on resource utilization  
- **Cost optimization** through efficiency metrics
- **SLA monitoring** with success rate tracking

## Related Tools

- **[cluster-status](../../bin/cluster-status)** - Manual cluster health checking
- **[aws-find-resources](../../bin/aws-find-resources)** - Infrastructure resource discovery
- **[argocd-cleanup](../../bin/argocd-cleanup)** - ArgoCD resource cleanup

This comprehensive observability solution provides complete visibility into the OpenShift Bootstrap cluster provisioning process, enabling data-driven optimization and reliable operations at scale.