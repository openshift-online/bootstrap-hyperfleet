# OpenShift Bootstrap - Production Installation Guide

**Audience**: System Administrators, Platform Engineers  
**Complexity**: Advanced  
**Estimated Time**: 3-4 hours for complete production setup  
**Prerequisites**: Cluster-admin access, AWS credentials, production environment planning

This guide covers production-grade installation and management of multi-cluster OpenShift deployments using GitOps automation.

> **New Users**: For basic setup and learning, start with [Getting Started Installation Guide](./installation.md)

## Production Planning

### Infrastructure Requirements

**Hub Cluster Specifications:**
- OpenShift 4.12+ with cluster-admin permissions
- Minimum 32GB RAM, 8 vCPUs for production workloads
- High-availability control plane (3+ masters)
- Persistent storage for GitOps and ACM data
- Network connectivity to all target regions

**Managed Cluster Capacity:**
- Hub cluster can manage 100+ regional clusters
- Each managed cluster requires ~200MB memory on hub
- Network bandwidth scales with cluster count and sync frequency

**AWS Resource Planning:**
- VPC quotas in target regions
- EC2 instance limits for cluster nodes
- EBS volume quotas for persistent storage
- Route53 hosted zone for cluster domains

### Security Considerations

**Secret Management:**
- Production environments should use Vault or External Secrets Operator
- Avoid storing secrets in Git repository
- Implement secret rotation policies
- Use separate AWS IAM roles per region/environment

**Network Security:**
- Private subnets for cluster nodes
- VPC peering or transit gateway for hub-spoke communication
- Network policies for workload isolation
- TLS certificates for all external endpoints

**Access Control:**
- RBAC policies for GitOps operations
- Separate service accounts per cluster
- ACM cluster access controls
- ArgoCD application-level permissions

## Production Bootstrap Process

### 1. Environment Preparation

```bash
# Clone repository
git clone https://github.com/openshift-online/bootstrap.git
cd bootstrap

# Verify production cluster access
oc login https://api.production-hub.example.com:6443 --token=your-token
oc whoami
oc cluster-info
```

### 2. Secret Management Setup

**Option A: External Secrets Operator (Recommended)**

Follow the [Vault Setup Guide](../../operators/vault/global/VAULT-SETUP.md) to configure enterprise secret management:

```bash
# Deploy External Secrets Operator
oc apply -k operators/external-secrets/

# Configure Vault integration
./bin/bootstrap.vault-integration.sh
```

**Option B: Manual Secret Creation**

For non-production or testing environments:

```bash
./bin/bootstrap.vault.sh
```

### 3. Production Bootstrap

```bash
# Deploy core infrastructure
./bin/bootstrap.sh

# Verify all components
oc get csv -n openshift-operators | grep -E "(gitops|advanced-cluster-management|pipelines)"
oc get mch -n open-cluster-management
oc get applications -n openshift-gitops
```

**Production Verification:**
- All operators show "Succeeded" phase
- ACM MultiClusterHub shows "Running" status
- ArgoCD applications sync successfully
- No error events in system namespaces

## Production Cluster Management

### Regional Deployment Strategy

**Geographic Distribution:**
```bash
# Production cluster layout example
regions/
├── us-east-1/          # Primary production region
│   ├── prod-api/       # API services cluster
│   └── prod-web/       # Web services cluster
├── us-west-2/          # DR region
│   ├── dr-api/         # Disaster recovery
│   └── dr-web/
├── eu-west-1/          # European region
│   ├── eu-api/         # EU data residency
│   └── eu-web/
└── ap-southeast-1/     # Asia Pacific
    ├── ap-api/
    └── ap-web/
```

### Automated Cluster Provisioning

**Production Cluster Creation:**
```bash
# Use standardized naming convention
./bin/cluster-generate regions/us-east-1/prod-api/

# Validate before deployment
kustomize build clusters/prod-api/ | oc apply --dry-run=client -f -

# Deploy through GitOps
git add .
git commit -m "Add prod-api cluster"
git push origin main
```

**Cluster Specifications for Production:**
```yaml
# regions/us-east-1/prod-api/region.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: prod-api
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: production.example.com
  
  # Production compute config
  compute:
    instanceType: m5.2xlarge    # Larger instances
    replicas: 6                 # Higher replica count
    
  # Production-specific settings
  controlPlane:
    instanceType: m5.xlarge
    replicas: 3
    
  networking:
    networkType: OVNKubernetes
    machineNetwork:
      cidr: 10.0.0.0/16
```

### Production Monitoring

**Comprehensive Health Checking:**
```bash
# Automated monitoring setup
./bin/monitor-health --production --export-metrics

# Check cluster provisioning status
oc get clusterdeployments -A
oc get awsmanagedcontrolplane -A

# Monitor ACM managed clusters
oc get managedclusters
oc get placementrules -A

# ArgoCD application health
oc get applications -n openshift-gitops
argocd app list --grpc-web
```

**Production Alerting:**
```bash
# Set up monitoring dashboards
oc apply -k monitoring/production/

# Configure alert rules
oc apply -k monitoring/alerts/
```

### Backup and Disaster Recovery

**GitOps Repository Backup:**
```bash
# Automated backup of cluster configurations
./bin/backup-configs --s3-bucket production-backup
```

**ACM Configuration Backup:**
```bash
# Export ACM policies and placements
oc get managedclusters -o yaml > backup/managedclusters.yaml
oc get placementbindings -A -o yaml > backup/placementbindings.yaml
```

**Hub Cluster Recovery:**
```bash
# Restore from backup
./bin/restore-hub --from-backup backup/hub-cluster-backup.tar.gz
```

## Advanced Configuration

### Multi-Tenancy Setup

**Namespace Isolation:**
```yaml
# Setup per-team namespaces
apiVersion: v1
kind: Namespace
metadata:
  name: team-alpha-clusters
  labels:
    team: alpha
---
# Team-specific RBAC
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: team-alpha-cluster-admin
  namespace: team-alpha-clusters
subjects:
- kind: Group
  name: team-alpha
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
```

**ApplicationSet Per Team:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: team-alpha-clusters
  namespace: openshift-gitops
spec:
  generators:
  - clusters:
      selector:
        matchLabels:
          team: alpha
  template:
    metadata:
      name: '{{name}}-{{team}}'
    spec:
      source:
        path: 'clusters/{{team}}/{{name}}'
```

### Performance Optimization

**Hub Cluster Tuning:**
```bash
# Increase ArgoCD controller replicas
oc patch argocd openshift-gitops -n openshift-gitops --type='merge' -p='{"spec":{"controller":{"replicas":3}}}'

# Tune ACM for large scale
oc patch mch multiclusterhub -n open-cluster-management --type='merge' -p='{"spec":{"availabilityConfig":"High"}}'
```

**Network Optimization:**
```bash
# Configure cluster networking for performance
oc apply -f configs/production/network-performance.yaml
```

## Production Troubleshooting

### Large Scale Issues

**ApplicationSet Performance:**
```bash
# Check ApplicationSet controller logs
oc logs -n openshift-gitops deployment/openshift-gitops-applicationset-controller

# Monitor resource usage
oc top pods -n openshift-gitops
```

**ACM Hub Performance:**
```bash
# Check ACM controller resource usage
oc top pods -n open-cluster-management

# Review hub cluster resource limits
oc describe mch multiclusterhub -n open-cluster-management
```

### Security Incident Response

**Cluster Isolation:**
```bash
# Temporarily isolate cluster from hub
oc label managedcluster problem-cluster cluster.open-cluster-management.io/unreachable=true

# Remove cluster from GitOps management
oc delete application problem-cluster-cluster -n openshift-gitops
```

**Secret Rotation:**
```bash
# Rotate AWS credentials
./bin/rotate-aws-credentials --cluster problem-cluster

# Update GitOps secrets
./bin/update-gitops-secrets
```

### Compliance and Auditing

**Configuration Auditing:**
```bash
# Export all cluster configurations
./bin/audit-clusters --export-format yaml > audit/cluster-configs.yaml

# Check policy compliance
oc get configurationpolicy -A
```

**Change Tracking:**
```bash
# Git history for cluster changes
git log --oneline --grep="cluster" --since="1 month ago"

# ArgoCD deployment history
argocd app history production-cluster --grpc-web
```

## Maintenance Procedures

### Cluster Upgrades

**Rolling Upgrade Strategy:**
```bash
# Upgrade clusters by region
for region in us-east-1 us-west-2 eu-west-1; do
  ./bin/upgrade-region-clusters $region --version 4.14.1
  ./bin/verify-region-health $region
done
```

### GitOps Repository Maintenance

**Repository Cleanup:**
```bash
# Remove decommissioned clusters
./bin/cleanup-decommissioned-clusters

# Optimize repository size
git gc --aggressive
```

**Configuration Validation:**
```bash
# Validate all configurations
./bin/validate-all-configs --strict
```

## Related Documentation

- **[Getting Started](./installation.md)** - Basic setup for new users
- **[Architecture](../architecture/ARCHITECTURE.md)** - System design and components
- **[Vault Setup](../../operators/vault/global/VAULT-SETUP.md)** - Enterprise secret management
- **[Monitoring Guide](../../guides/monitoring.md)** - Health checking and observability
- **[Regional Specifications](../architecture/REGIONALSPEC.md)** - Cluster configuration details

This production installation guide provides enterprise-grade deployment and management capabilities for multi-cluster OpenShift environments.