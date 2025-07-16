# EKS aws-auth ConfigMap Setup

## Overview
The aws-auth ConfigMap in the kube-system namespace controls access to the EKS cluster by mapping IAM users and roles to Kubernetes users and groups.

## Current EKS Cluster Details
- **Cluster Name**: acme-test-001
- **Region**: us-east-1
- **Account**: 765374464689
- **Current User**: arn:aws:iam::765374464689:user/mturansk

## Method 1: Create aws-auth ConfigMap (If Missing)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-auth
  namespace: kube-system
data:
  mapRoles: |
    - rolearn: arn:aws:iam::765374464689:role/AmazonEKSAutoClusterRole
      username: eks-cluster-admin
      groups:
        - system:bootstrappers
        - system:nodes
        - system:masters
    - rolearn: arn:aws:iam::765374464689:role/eksctl-acme-test-001-nodegroup-NodeInstanceRole
      username: system:node:{{EC2PrivateDNSName}}
      groups:
        - system:bootstrappers
        - system:nodes
  mapUsers: |
    - userarn: arn:aws:iam::765374464689:user/mturansk
      username: mturansk
      groups:
        - system:masters
```

## Method 2: Using eksctl (Recommended)

```bash
# Add current user as cluster admin
eksctl create iamidentitymapping \
  --cluster acme-test-001 \
  --region us-east-1 \
  --arn arn:aws:iam::765374464689:user/mturansk \
  --group system:masters \
  --username mturansk

# Add the auto cluster role
eksctl create iamidentitymapping \
  --cluster acme-test-001 \
  --region us-east-1 \
  --arn arn:aws:iam::765374464689:role/AmazonEKSAutoClusterRole \
  --group system:masters \
  --username eks-cluster-admin
```

## Method 3: Using AWS CLI

```bash
# Get current aws-auth ConfigMap (if exists)
kubectl get configmap aws-auth -n kube-system -o yaml > aws-auth-backup.yaml

# Create or update aws-auth ConfigMap
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-auth
  namespace: kube-system
data:
  mapRoles: |
    - rolearn: arn:aws:iam::765374464689:role/AmazonEKSAutoClusterRole
      username: eks-cluster-admin
      groups:
        - system:masters
  mapUsers: |
    - userarn: arn:aws:iam::765374464689:user/mturansk
      username: mturansk
      groups:
        - system:masters
EOF
```

## Method 4: Manual kubectl edit

```bash
# Edit existing ConfigMap
kubectl edit configmap aws-auth -n kube-system

# Or create new one
kubectl create configmap aws-auth -n kube-system --from-literal=mapUsers='...'
```

## Verification Steps

```bash
# Test cluster access
kubectl get nodes

# Check current user
kubectl auth whoami

# Verify permissions
kubectl auth can-i '*' '*' --all-namespaces

# List ConfigMap
kubectl get configmap aws-auth -n kube-system -o yaml
```

## Common IAM Roles for EKS

- **Node Group Role**: Usually named like `eksctl-{cluster-name}-nodegroup-*-NodeInstanceRole`
- **Cluster Service Role**: Usually named like `eksServiceRole` or `AmazonEKSClusterServiceRole`
- **Auto Cluster Role**: `AmazonEKSAutoClusterRole` (as specified)

## Bootstrap User Setup

For the bootstrap repository, ensure the user has:
- **system:masters** group (full cluster admin)
- Ability to create namespaces, deploy operators, manage GitOps

## Security Notes

- Always backup existing aws-auth before modifying
- Minimize users in system:masters group
- Use specific RBAC roles for non-admin users
- Monitor access with CloudTrail

## Troubleshooting

### Access Denied
- Verify IAM user ARN is correct
- Check aws-auth ConfigMap syntax
- Ensure kubectl context is correct

### eksctl Command Fails
```bash
# Install eksctl if missing
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin
```

### ConfigMap Not Found
- Create it using Method 1 above
- Verify you're in correct cluster context
- Check if cluster was created with different tools