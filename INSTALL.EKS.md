# EKS Bootstrap Installation Guide

This guide provides comprehensive instructions for bootstrapping Amazon EKS clusters using this GitOps repository with Helm-based operator deployment.

## Architecture Overview

The EKS bootstrap process adapts the OpenShift GitOps approach for EKS using:
- **ArgoCD (Helm)**: Continuous deployment and cluster management
- **Red Hat Advanced Cluster Management (Upstream)**: Multi-cluster management with CAPI integration
- **Cluster API (CAPI)**: EKS cluster provisioning via AWS infrastructure provider
- **Kustomize**: YAML configuration management adapted for EKS

## Prerequisites

### EKS Cluster Requirements
- EKS 1.31+ cluster with admin permissions
- Minimum 8GB RAM, 2 vCPUs for control plane workloads
- AWS IAM permissions for cluster provisioning and management

### Tools Required
```bash
# Install required tools
brew install aws-cli kubectl helm eksctl
# or
sudo apt-get install awscli kubectl helm
```

### AWS Authentication
```bash
# Configure AWS credentials
aws configure --profile default

# Update kubeconfig for existing EKS cluster
aws eks update-kubeconfig --region us-east-1 --name your-cluster-name --profile default
```

### Repository Setup
```bash
git clone https://github.com/openshift-online/bootstrap.git
cd bootstrap
```

## Current EKS Hub Status

### Test Environment
- **Cluster**: `acme-test-001` (us-east-1)
- **Kubeconfig**: `/home/mturansk/projects/secrets/eks.kubeconfig`
- **Status**: ✅ Cluster access verified (1 node ready)
- **Version**: v1.33.1-eks-b9364f6

### Completed Setup
```bash
# Generate kubeconfig
aws eks update-kubeconfig --region us-east-1 --name acme-test-001 --profile default --kubeconfig /path/to/eks.kubeconfig

# Verify access
export KUBECONFIG=/path/to/eks.kubeconfig
kubectl get nodes
```

## Installation Process

### 1. Install ArgoCD via Helm

#### Add ArgoCD Helm Repository
```bash
# Add official ArgoCD Helm repo
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update
```

#### Install ArgoCD
```bash
# Create namespace and install ArgoCD
kubectl create namespace argocd
helm install argocd argo/argo-cd -n argocd --create-namespace

# Wait for deployment
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
```

#### Configure ArgoCD Access
```bash
# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to access UI (or configure ingress)
kubectl port-forward service/argocd-server -n argocd 8080:443

# Access ArgoCD at https://localhost:8080
# Username: admin, Password: from secret above
```

### 2. Install Advanced Cluster Management (Upstream)

**⚠️ Issue Found**: ACM/OCM installation on EKS requires manual adaptation. Current upstream URLs and Helm repos are not available.

#### Current Status
```bash
# Create ACM namespace
kubectl create namespace open-cluster-management

# NOTE: Standard OCM installation methods not working:
# - helm repo add ocm https://openclustermanagement.io/helm-charts/ (invalid repo)
# - GitHub raw URLs return 404 errors
# - Requires manual manifest compilation from source
```

**Recommended Alternatives for EKS**:
- **✅ Cluster API**: Direct CAPI controller installation (working)
- **Flux**: Alternative GitOps solution  
- **Amazon EKS Anywhere**: AWS's multi-cluster solution
- **Manual OCM**: Build from source for advanced users

**For this installation**: Continue with CAPI-only approach for cluster provisioning

### 3. Install Cluster API for EKS

#### Initialize CAPI Management Cluster
```bash
# Install clusterctl (adjust path as needed - avoid sudo requirements)
mkdir -p ~/projects/bin
curl -L https://github.com/kubernetes-sigs/cluster-api/releases/latest/download/clusterctl-linux-amd64 -o ~/projects/bin/clusterctl
chmod +x ~/projects/bin/clusterctl
export PATH=$PATH:~/projects/bin

# Configure AWS credentials for CAPI (required environment variables)
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id --profile default)
export AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key --profile default)
export AWS_B64ENCODED_CREDENTIALS=$(echo -n "$AWS_ACCESS_KEY_ID:$AWS_SECRET_ACCESS_KEY" | base64 -w0)

# Initialize CAPI with AWS infrastructure provider
clusterctl init --infrastructure aws
```

#### Configure AWS Infrastructure Provider
```bash
# Create AWS credentials secret
kubectl create secret generic capa-manager-bootstrap-credentials \
  --from-literal=AccessKeyID="$AWS_ACCESS_KEY_ID" \
  --from-literal=SecretAccessKey="$AWS_SECRET_ACCESS_KEY" \
  --namespace capa-system
```

## Regional Fleet Management

### Current Fleet Configuration
Based on the repository structure, these EKS clusters are ready for deployment:
- **cluster-41**: EKS stage (us-west-2, m5.large, 1-10 nodes)
- **cluster-42**: EKS prod (ap-southeast-1, m5.xlarge, 2-20 nodes)

### EKS Cluster Templates
The repository includes EKS-specific templates in:
```
regions/templates/eks/
regions/us-west-2/eks-stage/
regions/ap-southeast-1/eks-prod/
clusters/overlay/cluster-41/
clusters/overlay/cluster-42/
```

### Deploy EKS Fleet via ArgoCD

#### 1. Convert GitOps Applications

**⚠️ Issue Found**: ArgoCD applications need namespace and syncPolicy format updates for EKS.

Required changes:
```yaml
# Change namespace from openshift-gitops to argocd
metadata:
  namespace: argocd  # was: openshift-gitops

# Fix syncPolicy format
syncPolicy:
  automated:
    selfHeal: true
    allowEmpty: false
  syncOptions:
  - Prune=false  # was: prune: false
```

```bash
# Apply EKS-compatible ArgoCD applications (after fixes)
kubectl apply -f gitops-applications/regional-clusters.cluster-41.application.yaml
kubectl apply -f gitops-applications/regional-clusters.cluster-42.application.yaml
```

#### 2. Monitor Cluster Provisioning
```bash
# Watch CAPI cluster creation
kubectl get clusters -A

# Check EKS-specific resources
kubectl get awsmanagedcontrolplane -A
kubectl get awsmanagedmachinepool -A
```

### Adding New EKS Clusters

#### 1. Use Regional Template System
```bash
# Create new region specification
mkdir -p regions/eu-west-1/eks-prod

# Copy from existing template
cp -r regions/templates/eks/* regions/eu-west-1/eks-prod/

# Customize for region and environment
vi regions/eu-west-1/eks-prod/cluster-spec.yaml
```

#### 2. Generate Traditional Overlays
```bash
# Create kustomize overlay for compatibility
./scripts/generate-cluster-overlay.sh cluster-43 eu-west-1 eks-prod
```

#### 3. Create ArgoCD Applications
```bash
# Copy and customize application manifests
cp gitops-applications/regional-clusters.cluster-41.application.yaml \
   gitops-applications/regional-clusters.cluster-43.application.yaml

# Update cluster references
sed -i 's/cluster-41/cluster-43/g' gitops-applications/regional-clusters.cluster-43.application.yaml
sed -i 's/us-west-2/eu-west-1/g' gitops-applications/regional-clusters.cluster-43.application.yaml
```

## Helm Values Configuration

### ArgoCD Helm Values for Multi-Cluster
Create `helm-values/argocd-values.yaml`:
```yaml
global:
  image:
    tag: v2.12.7

server:
  service:
    type: LoadBalancer  # For AWS NLB
  ingress:
    enabled: true
    annotations:
      alb.ingress.kubernetes.io/scheme: internet-facing
    hosts:
      - argocd.your-domain.com

configs:
  repositories:
    bootstrap-repo:
      url: https://github.com/openshift-online/bootstrap.git
      type: git
      name: bootstrap

  cluster:
    # Enable cluster-level RBAC
    admin.enabled: true

rbac:
  policy.default: role:readonly
  policy.csv: |
    p, role:cluster-admin, clusters, *, *, allow
    p, role:cluster-admin, applications, *, *, allow
    g, system:cluster-admins, role:cluster-admin

applicationSet:
  enabled: true
```

### Install with Custom Values
```bash
helm upgrade --install argocd argo/argo-cd \
  -n argocd \
  -f helm-values/argocd-values.yaml \
  --create-namespace
```

## EKS-Specific Configurations

### AWS Load Balancer Controller
```bash
# Install AWS Load Balancer Controller for ingress
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=your-cluster-name \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

### EBS CSI Driver
```bash
# Install EBS CSI driver for persistent volumes
helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
helm install aws-ebs-csi-driver aws-ebs-csi-driver/aws-ebs-csi-driver \
  -n kube-system
```

### Cluster Autoscaler
```bash
# Install cluster autoscaler
helm repo add autoscaler https://kubernetes.github.io/autoscaler
helm install cluster-autoscaler autoscaler/cluster-autoscaler \
  -n kube-system \
  --set autoDiscovery.clusterName=your-cluster-name \
  --set awsRegion=us-east-1
```

## GitOps Workflow Adaptation

### Application Structure
The existing GitOps applications work with EKS by:
1. **ArgoCD Applications**: Point to EKS-compatible manifests
2. **CAPI Resources**: Use AWSManagedControlPlane instead of ClusterDeployment
3. **Regional Deployments**: Deploy to EKS clusters via ArgoCD

### Secrets Management
```bash
# Create AWS credentials for cluster provisioning
kubectl create secret generic aws-credentials \
  --from-literal=AccessKeyID="$AWS_ACCESS_KEY_ID" \
  --from-literal=SecretAccessKey="$AWS_SECRET_ACCESS_KEY" \
  -n cluster-41

# Apply to each cluster namespace
for cluster in cluster-41 cluster-42; do
  kubectl create namespace $cluster
  kubectl create secret generic aws-credentials \
    --from-literal=AccessKeyID="$AWS_ACCESS_KEY_ID" \
    --from-literal=SecretAccessKey="$AWS_SECRET_ACCESS_KEY" \
    -n $cluster
done
```

## Monitoring and Troubleshooting

### ArgoCD Status
```bash
# Check ArgoCD applications
kubectl get applications -n argocd

# View application sync status
kubectl describe application regional-clusters -n argocd
```

### CAPI Cluster Status
```bash
# Monitor cluster provisioning
kubectl get clusters -A -w

# Check specific cluster
kubectl describe cluster cluster-41 -n cluster-41

# View CAPI controller logs
kubectl logs -n capa-system deployment/capa-controller-manager
```

### EKS Cluster Access
```bash
# Update kubeconfig for new clusters
aws eks update-kubeconfig --region us-west-2 --name cluster-41

# Test access
kubectl get nodes --context=arn:aws:eks:us-west-2:account:cluster/cluster-41
```

## Differences from OpenShift

### Key Adaptations
1. **Operators**: Helm charts instead of OLM subscriptions
2. **Routes**: Ingress controllers instead of OpenShift routes  
3. **Security**: Pod Security Standards instead of SCCs
4. **Registry**: ECR or external registries instead of internal registry
5. **Networking**: CNI plugins instead of SDN/OVN

### Feature Mapping
| OpenShift | EKS Equivalent |
|-----------|----------------|
| OpenShift GitOps | ArgoCD Helm Chart |
| OLM Operators | Helm Charts |
| Routes | ALB/NLB Ingress |
| SCCs | Pod Security Standards |
| Hive | Cluster API |
| Internal Registry | Amazon ECR |

## Production Considerations

### High Availability
- Deploy ArgoCD with HA configuration
- Use multiple availability zones for EKS clusters
- Configure backup and disaster recovery

### Security
- Enable EKS cluster encryption
- Use IAM roles for service accounts (IRSA)
- Implement network policies
- Regular security updates

### Scaling
- Configure cluster autoscaling
- Use horizontal pod autoscaling
- Monitor resource utilization

This EKS installation provides GitOps automation similar to the OpenShift approach, adapted for EKS-specific requirements and AWS services.