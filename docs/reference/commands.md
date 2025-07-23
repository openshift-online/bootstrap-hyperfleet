# Command Reference

**Audience**: All users  
**Complexity**: Reference  
**Estimated Time**: Quick lookup  
**Prerequisites**: Basic understanding of the system

## 🔧 Cluster Management Commands

### Quick Status
```bash
# Complete environment health check
./bin/monitor-health

# Hub cluster status
oc get managedcluster
oc get applications -n openshift-gitops

# Individual cluster status  
oc get clusterdeployment -A      # OCP clusters
oc get cluster.cluster.x-k8s.io -A  # EKS clusters
```

### Cluster Creation
```bash
# Interactive cluster creation
./bin/cluster-create

# Generate cluster from existing specification
./bin/cluster-generate regions/us-east-1/cluster-name/

# Regenerate all clusters (after base template changes)
./bin/regenerate-all-clusters
```

### Deployment Management
```bash
# Deploy all GitOps applications
./bin/bootstrap.sh

# Force application sync
oc patch application APP-NAME -n openshift-gitops --type='merge' \
  -p='{"operation":{"sync":{"syncStrategy":{"hook":{"force":true}}}}}'

# Suspend/resume ApplicationSet
oc patch applicationset APPSET-NAME -n openshift-gitops --type='merge' \
  -p='{"spec":{"syncPolicy":{"suspend":true}}}'   # suspend
oc patch applicationset APPSET-NAME -n openshift-gitops --type='merge' \
  -p='{"spec":{"syncPolicy":{"suspend":false}}}'  # resume
```

## 📊 Monitoring Commands

### Health Checks
```bash
# Comprehensive status report
./bin/monitor-health

# ArgoCD application status
oc get applications -n openshift-gitops
oc get applications -n openshift-gitops | grep -E "(OutOfSync|Unknown|Degraded)"

# Cluster connectivity
oc get managedcluster
oc describe managedcluster CLUSTER-NAME
```

### Resource Monitoring
```bash
# CRD establishment
./status.sh applications.argoproj.io 300
./status.sh clusters.cluster.x-k8s.io 180

# Resource condition waiting
./wait.kube.sh clusterdeployment CLUSTER-NAME NAMESPACE '{.status.webConsoleURL}' 
./wait.kube.sh awsmanagedcontrolplane CLUSTER-NAME NAMESPACE '{.status.ready}' true

# Node resource usage
oc adm top nodes --use-protocol-buffers
```

### Troubleshooting
```bash
# Application details
oc describe application APP-NAME -n openshift-gitops
oc logs -n openshift-gitops deployment/openshift-gitops-application-controller

# Cluster provisioning issues
oc describe clusterdeployment CLUSTER-NAME -n CLUSTER-NAME     # OCP
oc describe awsmanagedcontrolplane CLUSTER-NAME -n CLUSTER-NAME # EKS

# ACM issues
oc describe managedcluster CLUSTER-NAME
oc logs -n open-cluster-management deployment/multicluster-operators-hub-registration
```

## 🔐 Secret Management Commands

### Vault Operations
```bash
# Check Vault status
oc get pods -n vault
oc exec vault-0 -n vault -- vault status

# List secrets in Vault
oc exec vault-0 -n vault -- vault kv list secret/

# Read specific secret
oc exec vault-0 -n vault -- vault kv get secret/aws-credentials

# Update secret in Vault
oc exec vault-0 -n vault -- vault kv put secret/aws-credentials \
  aws_access_key_id="new-key" \
  aws_secret_access_key="new-secret"
```

### External Secrets Operator
```bash
# Check ESO status
oc get pods -n external-secrets
oc get clustersecretstore vault-cluster-store

# Check ExternalSecret status
oc get externalsecret -A
oc describe externalsecret aws-credentials -n CLUSTER-NAME

# Force secret refresh
oc annotate externalsecret aws-credentials -n CLUSTER-NAME force-sync="$(date +%s)"
```

### Traditional Secret Management
```bash
# Apply secrets to namespace
oc apply -f secrets/aws-credentials.yaml -n CLUSTER-NAME
oc apply -f secrets/pull-secret.yaml -n CLUSTER-NAME

# Run vault bootstrap script
./bin/bootstrap.vault.sh
```

## 🔧 Cluster Operations

### Scaling
```bash
# EKS cluster scaling
oc patch awsmanagedmachinepool CLUSTER-NAME -n CLUSTER-NAME --type='merge' \
  -p='{"spec":{"scaling":{"maxSize":10,"minSize":3}}}'

# OCP cluster scaling  
oc patch machinepool CLUSTER-NAME-worker -n CLUSTER-NAME --type='merge' \
  -p='{"spec":{"replicas":5}}'
```

### Upgrades
```bash
# EKS Kubernetes version upgrade
oc patch awsmanagedcontrolplane CLUSTER-NAME -n CLUSTER-NAME --type='merge' \
  -p='{"spec":{"version":"1.28"}}'

# OCP version upgrade
oc patch clusterdeployment CLUSTER-NAME -n CLUSTER-NAME --type='merge' \
  -p='{"spec":{"provisioning":{"imageSetRef":{"name":"openshift-v4.14.0-img"}}}}'
```

### Configuration Updates
```bash
# Validate configuration changes
kustomize build clusters/CLUSTER-NAME/
oc --dry-run=client apply -k clusters/CLUSTER-NAME/

# Apply changes via GitOps
git add . && git commit -m "Update cluster config" && git push
```

## 🧹 Cleanup Commands

### AWS Resource Cleanup
```bash
# Interactive cleanup
./bin/clean-aws

# Automated cleanup (no prompts)
./bin/clean-aws --disable-prompts

# Debug mode cleanup
./bin/clean-aws --debug --disable-prompts
```

### Cluster Removal
```bash
# Remove cluster ApplicationSet and applications
oc delete applicationset CLUSTER-NAME-applications -n openshift-gitops

# Remove cluster namespace
oc delete namespace CLUSTER-NAME

# Remove cluster overlays
rm -rf clusters/CLUSTER-NAME
rm -rf pipelines/*/CLUSTER-NAME  
rm -rf operators/openshift-pipelines/CLUSTER-NAME
rm -rf deployments/ocm/CLUSTER-NAME
rm -f gitops-applications/CLUSTER-NAME.yaml

# Update GitOps kustomization
# Remove line: - ./CLUSTER-NAME.yaml from gitops-applications/kustomization.yaml
```

## 🔧 Development Commands

### Validation
```bash
# Kustomize builds
kustomize build clusters/CLUSTER-NAME/
kustomize build deployments/ocm/CLUSTER-NAME/
kustomize build gitops-applications/

# Dry-run validation
oc --dry-run=client apply -k clusters/CLUSTER-NAME/
oc --dry-run=client apply -k gitops-applications/
```

### Container Operations
```bash
# Build container images
make podman-build

# Run containers
make podman-run
```

## 📋 Information Commands

### Cluster Access
```bash
# EKS cluster access
aws eks update-kubeconfig --region REGION --name CLUSTER-NAME

# OCP cluster access
oc extract secret/CLUSTER-NAME-admin-kubeconfig -n CLUSTER-NAME --to=- > cluster-kubeconfig
export KUBECONFIG=cluster-kubeconfig
```

### Resource Inventory
```bash
# Cluster inventory
oc get managedcluster -o custom-columns=NAME:.metadata.name,STATUS:.status.conditions[-1].type,VERSION:.status.version.kubernetes

# Application inventory
oc get applications -n openshift-gitops -o custom-columns=NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status

# Resource usage
oc adm top nodes --use-protocol-buffers
oc adm top pods -A --use-protocol-buffers
```

## 🔗 Useful Aliases

Add these to your shell profile for convenience:

```bash
# Health and status
alias hc='./bin/monitor-health'
alias apps='oc get applications -n openshift-gitops'
alias clusters='oc get managedcluster'

# Common operations
alias new-cluster='./bin/cluster-create'
alias bootstrap='./bin/bootstrap.sh'
alias clean-aws='./bin/clean-aws'

# ArgoCD shortcuts
alias argocd-login='oc -n openshift-gitops get secret openshift-gitops-cluster -o jsonpath="{.data.admin\.password}" | base64 -d'
alias argocd-url='oc -n openshift-gitops get route openshift-gitops-server -o jsonpath="{.spec.host}"'

# Vault shortcuts
alias vault-ui='oc port-forward -n vault vault-0 8200:8200'
alias vault-status='oc exec vault-0 -n vault -- vault status'
```

## 📚 Quick References

### Important Namespaces
- `openshift-gitops` - ArgoCD and GitOps applications
- `open-cluster-management` - ACM hub components
- `vault` - HashiCorp Vault deployment
- `external-secrets` - External Secrets Operator
- `CLUSTER-NAME` - Individual cluster resources
- `ocm-CLUSTER-NAME` - Services deployed to managed clusters

### Key Resource Types
- `applicationset` - ArgoCD ApplicationSets
- `application` - ArgoCD Applications  
- `managedcluster` - ACM managed clusters
- `clusterdeployment` - Hive OCP cluster provisioning
- `awsmanagedcontrolplane` - CAPI EKS cluster provisioning
- `externalsecret` - ESO secret synchronization

### Default Timeouts
- Cluster provisioning: 30-45 minutes
- Application sync: 5 minutes
- Secret refresh: 1 hour (ESO)
- Health check interval: 15 minutes (recommended)