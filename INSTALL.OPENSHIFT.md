# OpenShift Bootstrap Installation Guide

This guide provides comprehensive instructions for bootstrapping Red Hat OpenShift clusters using this GitOps repository.

## Architecture Overview

The OpenShift bootstrap process uses OpenShift's native operators and GitOps capabilities:
- **OpenShift GitOps (ArgoCD)**: Continuous deployment and cluster management
- **Red Hat Advanced Cluster Management (ACM)**: Multi-cluster management with CAPI integration
- **Hive**: OpenShift cluster provisioning operator
- **Kustomize**: YAML configuration management

## Prerequisites

### OpenShift Cluster Requirements
- OpenShift 4.12+ cluster with cluster-admin permissions
- Minimum 16GB RAM, 4 vCPUs for control plane workloads
- Network connectivity to target regions for cluster provisioning

### Authentication
```bash
# Log in to your OpenShift cluster
oc login https://api.your-cluster.example.com:6443 --token=your-token
```

### Repository Setup
```bash
git clone https://github.com/openshift-online/bootstrap.git
cd bootstrap
```

## Installation Process

### 1. Bootstrap Control Plane
```bash
# Run the automated bootstrap script
./bootstrap.sh
```

**What this script does:**
1. **Deploys Prerequisites** (`oc apply -k ./prereqs`):
   - OpenShift GitOps Operator subscription
   - Cluster role bindings for ArgoCD
   - Service accounts for cluster import

2. **Creates Secrets** (`./bootstrap.vault.sh`):
   - AWS credentials for cluster provisioning
   - Pull secrets for OpenShift installations
   - SSH keys for cluster access

3. **Deploys GitOps Applications** (`oc apply -k ./gitops-applications`):
   - ACM operator and MultiClusterHub instance
   - Regional cluster applications
   - Tekton pipelines for CI/CD

4. **Waits for Completion**:
   - OpenShift GitOps route availability
   - ACM hub components readiness
   - Regional cluster provisioning status

### 2. Access Control Plane
```bash
# Get the ArgoCD admin password
oc extract secret/openshift-gitops-cluster -n openshift-gitops --to=-

# Access the ArgoCD UI
oc get route openshift-gitops-server -n openshift-gitops
```

### 3. Monitor Cluster Provisioning
```bash
# Watch cluster deployment status
watch oc get clusterdeployments -A

# Check specific cluster status
oc describe clusterdeployment cluster-10 -n cluster-10
```

## Regional Cluster Management

### Existing Clusters
The repository is configured with these regional clusters:
- **cluster-10**: Production OCP cluster (existing)
- **cluster-20**: Staging OCP cluster (existing)
- **cluster-30**: Development OCP cluster (existing)

### Adding New OCP Clusters

#### 1. Create Cluster Overlay
```bash
# Copy existing cluster configuration
cp -r ./clusters/overlay/cluster-20 ./clusters/overlay/cluster-XX

# Update cluster references
find ./clusters/overlay/cluster-XX -type f -exec sed -i 's/cluster-20/cluster-XX/g' {} \;
```

#### 2. Configure Region and Compute
Edit `./clusters/overlay/cluster-XX/install-config.yaml`:
```yaml
metadata:
  name: cluster-XX
baseDomain: your-domain.com
platform:
  aws:
    region: us-west-2  # Update target region
compute:
- name: worker
  platform:
    aws:
      type: m5.xlarge  # Update instance type
  replicas: 3
```

#### 3. Create Regional Deployment
```bash
# Copy regional deployment configuration
cp -r ./regional-deployments/overlays/cluster-20 ./regional-deployments/overlays/cluster-XX

# Update references
find ./regional-deployments/overlays/cluster-XX -type f -exec sed -i 's/cluster-20/cluster-XX/g' {} \;
```

#### 4. Create ArgoCD Applications
```bash
# Copy application manifests
cp gitops-applications/regional-clusters.cluster-20.application.yaml \
   gitops-applications/regional-clusters.cluster-XX.application.yaml

cp gitops-applications/regional-deployments.cluster-20.application.yaml \
   gitops-applications/regional-deployments.cluster-XX.application.yaml

# Update cluster references in both files
sed -i 's/cluster-20/cluster-XX/g' gitops-applications/regional-clusters.cluster-XX.application.yaml
sed -i 's/cluster-20/cluster-XX/g' gitops-applications/regional-deployments.cluster-XX.application.yaml
```

#### 5. Update Kustomization
Add new applications to `./gitops-applications/kustomization.yaml`:
```yaml
resources:
# ... existing resources
- regional-clusters.cluster-XX.application.yaml
- regional-deployments.cluster-XX.application.yaml
```

#### 6. Deploy New Cluster
```bash
# Apply the updated configuration
oc apply -k ./gitops-applications

# Monitor deployment
./wait.kube.sh cd cluster-XX cluster-XX '{.status.conditions[?(@.type=="Provisioned")].message}' "Cluster is provisioned"
```

## Key Components

### Operators Deployed
- **OpenShift GitOps**: `/prereqs/openshift-gitops.subscription.yaml`
- **Advanced Cluster Management**: `/operators/advanced-cluster-management/`
- **OpenShift Pipelines**: `/operators/openshift-pipelines/`

### Cluster Provisioning
- **Base Configurations**: `/clusters/base/` - Common ClusterDeployment templates
- **Overlays**: `/clusters/overlay/` - Environment-specific customizations
- **Install Configs**: Embedded in ClusterDeployment secrets for Hive

### GitOps Applications
- **Application of Applications**: Root ArgoCD applications in `/gitops-applications/`
- **Regional Clusters**: Provision OCP clusters via Hive
- **Regional Deployments**: Deploy services to provisioned clusters

## Infrastructure Provider Integration

ACM automatically configures infrastructure providers for:
- **AWS**: Uses Hive for OCP cluster provisioning
- **Azure, GCP, vSphere**: Available via ACM infrastructure providers
- **OpenStack, BareMetal**: Supported through ACM

Configuration located at:
```
operators/advanced-cluster-management/instance/base/multiclusterhub.yaml
```

## Troubleshooting

### Common Issues

#### Bootstrap Script Fails
```bash
# Check cluster connectivity
oc cluster-info

# Verify prerequisites
oc get subscription openshift-gitops-operator -n openshift-operators

# Check operator status
oc get csv -n openshift-operators | grep gitops
```

#### Cluster Provisioning Stuck
```bash
# Check cluster deployment events
oc describe clusterdeployment cluster-XX -n cluster-XX

# Verify AWS credentials
oc get secret aws-creds -n cluster-XX -o yaml

# Check Hive controller logs
oc logs -n hive deployment/hive-controllers
```

#### ArgoCD Application Sync Issues
```bash
# Check application status
oc get applications -n openshift-gitops

# View application details
oc describe application regional-clusters -n openshift-gitops

# Check ArgoCD server logs
oc logs -n openshift-gitops deployment/openshift-gitops-server
```

### Validation Commands
```bash
# Test kustomize builds
kustomize build ./clusters/overlay/cluster-XX
kustomize build ./gitops-applications

# Verify cluster access
oc get managedclusters

# Check ACM status
oc get mch -n open-cluster-management
```

## Security Considerations

- All secrets are managed through OpenShift's native secret management
- Cluster access is controlled via RBAC policies
- Network policies isolate cluster management workloads
- Pull secrets and SSH keys are rotated regularly

## Advanced Configuration

### Custom Compute Configuration
Modify `install-config.yaml` in cluster overlays for:
- Instance types and sizes
- Availability zones
- Network CIDR ranges
- Storage configurations

### Policy Management
ACM policies are applied automatically for:
- Security compliance
- Configuration drift detection
- Application deployment governance

### Monitoring and Observability
Enable ACM observability:
```bash
oc apply -k operators/advanced-cluster-management/instance/observability/
```

## Support

For issues with this bootstrap process:
1. Check the troubleshooting section above
2. Review ArgoCD and ACM operator logs
3. Validate kustomize configurations
4. Ensure proper RBAC permissions

This installation method provides a production-ready OpenShift cluster management platform with full GitOps automation.