# OpenShift Bootstrap Observability Architecture

## Overview

The OpenShift Bootstrap observability architecture provides comprehensive end-to-end monitoring across the entire cluster provisioning lifecycle, from regional specification validation through complete cluster readiness. This architecture captures 100% of success metrics and enables data-driven optimization of the provisioning process.

## Architecture Principles

### Design Goals
- **Complete Visibility**: 100% coverage of provisioning lifecycle phases
- **Real-time Monitoring**: Sub-minute detection of issues and progress
- **Actionable Insights**: Precise failure location and root cause analysis
- **Predictive Capabilities**: Early warning systems for potential failures
- **Operational Excellence**: Data-driven SLA management and optimization

### Core Components
```
┌─────────────────────────────────────────────────────────────────┐
│                    Observability Architecture                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────────┐ │
│  │   Data Sources  │    │  Metrics Layer  │    │ Presentation │ │
│  │                 │    │                 │    │              │ │
│  │ • ArgoCD APIs   │────│ Custom Metrics  │────│   Grafana    │ │
│  │ • CAPI/Hive     │    │   Exporter      │    │  Dashboard   │ │
│  │ • ACM APIs      │    │                 │    │              │ │
│  │ • OpenShift     │    │ ┌─────────────┐ │    │ ┌──────────┐ │ │
│  │ • External Deps │    │ │ Prometheus  │ │    │ │ Alerting │ │ │
│  └─────────────────┘    │ │ Integration │ │    │ │  Rules   │ │ │
│                         │ └─────────────┘ │    │ └──────────┘ │ │
│                         └─────────────────┘    └──────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Provisioning Flow Observability

### Phase 1: Regional Specification Creation
**Duration: 1-2 minutes**

```yaml
Observability Focus:
  - Configuration validation success rate
  - Specification generation time
  - Input validation errors
  - Template rendering success

Metrics Exposed:
  - spec_validation_duration_seconds
  - spec_validation_success_total
  - spec_validation_failures_total{reason}
```

### Phase 2: Configuration Generation
**Duration: 2-5 minutes**

```yaml
Observability Focus:
  - Kustomize overlay generation
  - YAML validation success
  - Resource dependency resolution
  - GitOps application creation

Metrics Exposed:
  - config_generation_duration_seconds
  - kustomize_validation_success_total
  - gitops_application_creation_time
```

### Phase 3: GitOps Sync Wave Deployment
**Duration: 45-90 minutes**

#### Wave 1: Cluster Provisioning (30-60 minutes)
```yaml
CAPI Clusters (EKS):
  Observability:
    - cluster.x-k8s.io/Cluster readiness
    - AWSManagedControlPlane status
    - AWSManagedMachinePool scaling
    - AWS resource provisioning time
  
  Metrics:
    - cluster_infrastructure_status{type="capi"}
    - cluster_provisioning_duration_seconds{phase="infrastructure"}
    - aws_resource_provisioning_time{resource_type}

Hive Clusters (OCP):
  Observability:
    - ClusterDeployment conditions
    - InstallConfig validation
    - MachinePool replica status
    - Installation log analysis
  
  Metrics:
    - cluster_infrastructure_status{type="hive"}
    - cluster_install_progress_percent
    - cluster_provisioning_failures_total{reason}
```

#### Wave 2: Operator Installation (5-10 minutes)
```yaml
Observability Focus:
  - OpenShift Pipelines operator deployment
  - CRD availability and readiness
  - Operator pod startup time
  - Network connectivity validation

Metrics Exposed:
  - operator_installation_duration_seconds
  - operator_readiness_status
  - crd_availability_status
```

#### Wave 3: Pipeline Deployment (3-5 minutes)
```yaml
Observability Focus:
  - Tekton pipeline configuration
  - Pipeline run execution status
  - Task completion rates
  - Resource availability

Metrics Exposed:
  - pipeline_deployment_duration_seconds
  - pipeline_execution_success_rate
  - task_completion_time_seconds
```

#### Wave 4: Service Deployment (5-10 minutes)
```yaml
Observability Focus:
  - OCM service health (AMS-DB, OSL-DB, CS-DB, TREX)
  - Persistent volume provisioning
  - External secret synchronization
  - Service mesh connectivity

Metrics Exposed:
  - service_deployment_duration_seconds
  - service_health_status
  - pvc_provisioning_time_seconds
  - external_secret_sync_status
```

### Phase 4: ACM Integration & Management
**Duration: 10-20 minutes**

```yaml
Observability Focus:
  - ManagedCluster registration
  - Klusterlet agent deployment
  - Policy compliance evaluation
  - Multi-cluster observability setup

Metrics Exposed:
  - managed_cluster_registration_time
  - klusterlet_agent_connectivity
  - policy_compliance_score
  - multicluster_observability_status
```

## Metrics Architecture

### Custom Metrics Exporter Design

```python
# Core Architecture Components
class ClusterProvisioningMetrics:
    """
    Comprehensive metrics collector with multi-source integration
    
    Data Sources:
    - Kubernetes APIs (core, apps, custom resources)
    - ArgoCD Applications and ApplicationSets
    - CAPI cluster.x-k8s.io resources
    - Hive hive.openshift.io resources  
    - ACM cluster.open-cluster-management.io resources
    - OpenShift config.openshift.io resources
    
    Collection Strategy:
    - 30-second scrape interval
    - Parallel API queries for performance
    - Error isolation (failed queries don't stop collection)
    - State tracking for duration calculations
    """
```

### Prometheus Integration

```yaml
# ServiceMonitor Configuration
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cluster-provisioning-metrics
  namespace: openshift-gitops
spec:
  selector:
    matchLabels:
      app: cluster-provisioning-metrics
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    scheme: http
```

### Metric Categories

#### 1. End-to-End Provisioning Metrics
```prometheus
# Total provisioning duration from start to finish
cluster_provisioning_duration_seconds{cluster_name, cluster_type, region, result}

# Phase-wise duration breakdown
cluster_provisioning_phase_duration_seconds{cluster_name, phase, result}

# Active provisioning tracking
cluster_provisioning_active_total

# Success rate calculations
cluster_provisioning_success_rate{cluster_type, region, time_window}
```

#### 2. Sync Wave Progression Metrics
```prometheus
# Wave status tracking (0=not_started, 1=in_progress, 2=completed, 3=failed)
cluster_sync_wave_status{cluster_name, wave_number}

# Wave duration measurements
cluster_sync_wave_duration_seconds{cluster_name, wave_number, result}

# Application counts per wave
cluster_sync_wave_applications_total{cluster_name, wave_number, status}
```

#### 3. Infrastructure Provisioning Metrics
```prometheus
# Infrastructure readiness status
cluster_infrastructure_status{cluster_name, infrastructure_type}

# Worker node capacity tracking
cluster_worker_nodes_status{cluster_name, status_type}

# API responsiveness
cluster_api_response_time_seconds{cluster_name}
```

#### 4. Platform Health Metrics
```prometheus
# Cluster operator health
cluster_operators_status_total{cluster_name, status}

# Overall readiness score (0-100)
cluster_readiness_score{cluster_name}

# Core service availability
cluster_core_services_health{cluster_name, service}
```

#### 5. ACM Management Metrics
```prometheus
# ManagedCluster status
managed_cluster_status{cluster_name}

# Condition tracking
managed_cluster_conditions{cluster_name, condition_type}
```

#### 6. Failure Analysis Metrics
```prometheus
# Categorized failure tracking
cluster_provisioning_failures_total{cluster_name, failure_category, failure_reason}

# Retry success rates
cluster_provisioning_retry_success_rate{cluster_name, failure_type}
```

## Dashboard Architecture

### Multi-Layer Visualization Strategy

#### Layer 1: Executive Overview
```yaml
Purpose: High-level operational status
Audience: Management, SRE leadership
Refresh: 30 seconds
Content:
  - Active provisioning clusters count
  - 24-hour success rate percentage
  - Mean provisioning time trending
  - Geographic distribution of clusters
```

#### Layer 2: Operational Details
```yaml
Purpose: Real-time provisioning monitoring
Audience: SRE engineers, platform operators
Refresh: 30 seconds
Content:
  - End-to-end provisioning timeline
  - Sync wave progression matrix
  - Infrastructure health status
  - Platform readiness scores
```

#### Layer 3: Performance Analytics
```yaml
Purpose: Trend analysis and optimization
Audience: Platform engineers, capacity planners
Refresh: 5 minutes
Content:
  - Provisioning duration histograms
  - Failure categorization analysis
  - Resource utilization trends
  - Success rate trending by region/type
```

#### Layer 4: Troubleshooting Details
```yaml
Purpose: Deep dive investigation
Audience: SRE engineers during incidents
Refresh: 15 seconds
Content:
  - Individual cluster status table
  - Detailed failure logs integration
  - Dependency health matrix
  - Real-time log correlation
```

### Dashboard Panel Design

#### Sync Wave Visualization
```yaml
Panel Type: Heatmap
Purpose: Visual representation of wave progression across clusters
Data Source: cluster_sync_wave_status
Visualization:
  - X-axis: Clusters
  - Y-axis: Sync wave numbers (1-4)
  - Color coding: Green (completed), Yellow (in progress), Red (failed)
  - Tooltips: Wave duration, application counts
```

#### Infrastructure Status Matrix
```yaml
Panel Type: Stat panels with conditional coloring
Purpose: Infrastructure provisioning health
Data Sources:
  - cluster_infrastructure_status
  - cluster_worker_nodes_status
  - cluster_api_response_time_seconds
Visualization:
  - Green: Ready/Available
  - Yellow: Provisioning/Progressing
  - Red: Failed/Degraded
```

#### Provisioning Timeline
```yaml
Panel Type: Time series graph
Purpose: Duration tracking over time
Data Source: cluster_provisioning_duration_seconds
Features:
  - Multiple series per cluster
  - Success/failure annotation
  - Performance baseline overlays
  - Drill-down to phase details
```

## Alerting Architecture

### Alert Classification

#### Severity: Critical
```yaml
Alerts:
  - ClusterProvisioningFailed: Immediate notification of provisioning failures
  - InfrastructureProvisioningFailed: Infrastructure layer failures
  - SyncWaveFailed: ArgoCD sync wave failures
  - CoreServicesDegraded: API server, etcd unavailability

Response SLA: 5 minutes
Escalation: Page on-call engineer
```

#### Severity: Warning  
```yaml
Alerts:
  - ClusterProvisioningTakingTooLong: Duration > 75 minutes
  - SyncWaveStuck: Wave in progress > 30 minutes
  - WorkerNodesInsufficient: Actual < expected worker nodes
  - ClusterReadinessLow: Readiness score < 80%

Response SLA: 30 minutes
Escalation: Slack notification to SRE team
```

#### Severity: Info
```yaml
Alerts:
  - ProvisioningSuccessRateLow: Success rate < 95% over 24h
  - ManagedClusterUnreachable: ACM connectivity issues
  - ExternalDependencyDegraded: Vault, secrets, networking issues

Response SLA: 4 hours
Escalation: Daily standup discussion
```

### Alert Context Enhancement

```yaml
Alert Annotations:
  summary: Human-readable description
  description: Detailed context with cluster information
  runbook_url: Link to troubleshooting procedures
  dashboard_url: Direct link to relevant dashboard section
  grafana_orgId: Organization context for Grafana links
```

## Performance Baselines

### Provisioning Duration SLAs

| **Cluster Type** | **Target (95th %ile)** | **Warning Threshold** | **Critical Threshold** |
|------------------|------------------------|------------------------|------------------------|
| **OCP (Hive)**   | 45 minutes            | 60 minutes            | 90 minutes            |
| **EKS (CAPI)**   | 40 minutes            | 55 minutes            | 85 minutes            |

### Sync Wave Performance Expectations

| **Wave** | **Component** | **Target Duration** | **Warning Threshold** |
|----------|---------------|-------------------- |----------------------|
| **1**    | Infrastructure | 35 minutes         | 50 minutes           |
| **2**    | Operators     | 5 minutes          | 10 minutes           |
| **3**    | Pipelines     | 3 minutes          | 8 minutes            |
| **4**    | Services      | 5 minutes          | 12 minutes           |

### Success Rate Targets

| **Time Window** | **Target Success Rate** | **Warning Threshold** |
|-----------------|------------------------|----------------------|
| **1 hour**      | 98%                   | 90%                  |
| **24 hours**    | 95%                   | 85%                  |
| **7 days**      | 95%                   | 90%                  |

## Integration Points

### GitOps Integration
```yaml
Metrics Collection Triggers:
  - ArgoCD Application sync events
  - ApplicationSet reconciliation
  - Sync wave progression updates
  - Health status changes

Data Flow:
  ArgoCD Webhook → Metrics Exporter → Prometheus → Grafana
```

### Infrastructure Integration
```yaml
CAPI Integration:
  - cluster.x-k8s.io/Cluster status
  - Machine and MachineSet health
  - Cloud provider resource status

Hive Integration:
  - ClusterDeployment conditions
  - Installation progress tracking
  - ClusterImageSet and InstallConfig validation
```

### ACM Integration
```yaml
Multi-Cluster Management:
  - ManagedCluster lifecycle events
  - Klusterlet agent health
  - Policy compliance scoring
  - Multi-cluster observability data
```

### External Dependencies
```yaml
Monitoring Integration:
  - Vault connectivity and secret refresh
  - External Secrets Operator status
  - AWS credentials validation
  - Network connectivity checks
  - Container registry health
```

## Deployment Architecture

### High Availability Design
```yaml
Metrics Exporter:
  Replicas: 1 (single instance with leader election)
  Resource Limits: 256Mi memory, 200m CPU
  Restart Policy: Always
  Health Checks: /health endpoint

Prometheus Integration:
  Scrape Interval: 30 seconds
  Retention: 15 days (cluster metrics)
  Storage: Persistent volume for long-term retention

Grafana Dashboard:
  Auto-refresh: 30 seconds for operational views
  Auto-refresh: 5 minutes for analytical views
  Data retention: Linked to Prometheus retention
```

### Scaling Considerations
```yaml
Performance Characteristics:
  - Handles 50+ clusters simultaneously
  - Sub-second metric collection per cluster
  - Minimal API server impact (<1% additional load)
  - Memory usage scales linearly with cluster count

Optimization Strategies:
  - Parallel API queries for performance
  - Caching for repeated resource lookups
  - Incremental state tracking
  - Efficient Prometheus metric cardinality
```

## Security Considerations

### RBAC Requirements
```yaml
ServiceAccount: cluster-provisioning-metrics
ClusterRole Permissions:
  - get, list, watch: namespaces, pods, services
  - get, list, watch: applications.argoproj.io
  - get, list, watch: managedclusters.cluster.open-cluster-management.io
  - get, list, watch: clusters.cluster.x-k8s.io
  - get, list, watch: clusterdeployments.hive.openshift.io
  - get, list, watch: clusteroperators.config.openshift.io
```

### Data Privacy
```yaml
Metrics Collection:
  - No sensitive data in metric labels
  - Cluster names and types only
  - No secrets or configuration data
  - Aggregated statistics only

Data Retention:
  - Metrics: 15 days (Prometheus default)
  - Logs: 7 days (container logs)
  - Dashboards: No local data storage
```

## Operational Procedures

### Day 1 Operations
```yaml
Deployment Checklist:
  1. Deploy metrics exporter to openshift-gitops namespace
  2. Verify ServiceMonitor is discovered by Prometheus
  3. Import Grafana dashboard
  4. Configure alert notification channels
  5. Validate metric collection for existing clusters
  6. Test alert firing and resolution
```

### Day 2 Operations
```yaml
Maintenance Procedures:
  - Weekly dashboard review for performance trends
  - Monthly alert threshold tuning based on SLA data
  - Quarterly capacity planning using historical data
  - Continuous improvement of failure categorization

Troubleshooting Procedures:
  - Metrics collection failures → Check exporter logs
  - Dashboard not loading → Verify Prometheus connectivity
  - Missing cluster data → Validate RBAC permissions
  - High alert noise → Review and adjust thresholds
```

### Disaster Recovery
```yaml
Recovery Procedures:
  - Metrics exporter redeploy: < 5 minutes
  - Dashboard reimport: < 2 minutes
  - Historical data recovery: Dependent on Prometheus backup
  - Alert rule restoration: < 1 minute (via GitOps)

Backup Strategy:
  - Dashboard configuration: Stored in Git
  - Alert rules: Stored in Git
  - Prometheus data: Cluster backup strategy
  - Metrics exporter config: GitOps managed
```

## Future Enhancements

### Predictive Analytics
```yaml
Planned Features:
  - Machine learning-based failure prediction
  - Capacity planning automation
  - Performance trend analysis
  - Cost optimization recommendations

Implementation Timeline:
  - Q1: Historical data collection and analysis
  - Q2: Predictive model development
  - Q3: Integration with alerting system
  - Q4: Automated remediation workflows
```

### Extended Integration
```yaml
Additional Data Sources:
  - Cost tracking integration (AWS Cost Explorer)
  - Security scanning results (compliance scoring)
  - User experience metrics (provisioning request to cluster ready)
  - Network performance metrics (latency, throughput)

External System Integration:
  - ServiceNow incident creation
  - Slack workflow automation
  - Jira ticket creation for failures
  - Email reporting for management
```

This observability architecture provides a comprehensive foundation for monitoring, analyzing, and optimizing the OpenShift Bootstrap cluster provisioning process, enabling data-driven operational excellence and continuous improvement.