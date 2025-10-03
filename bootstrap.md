# Bootstrap Walkthrough

## Prerequisites
- OpenShift cluster (4.12+), `oc` CLI, cluster admin privileges

## Simple Bootstrap Process

### 1. Verify Cluster Access
```bash
oc cluster-info
oc whoami
```

### 2. Install OpenShift GitOps Operator
```bash
oc apply -k operators/openshift-gitops/global
./bin/monitor-status applications.argoproj.io
```

### 3. Deploy All GitOps Applications
```bash
oc apply -k gitops-applications/
```

This single command creates all ArgoCD applications with proper ordering:
- **Wave -1**: OpenShift GitOps (self-management)
- **Wave 1**: Pipelines (Tekton operator)
- **Wave 2**: Vault (secret management)
- **Wave 2**: ESO (External Secrets Operator)
- **Wave 3**: ACM (Advanced Cluster Management with ordered deployment)
  - Operator → Hub → Policies
- **Wave 4**: GitOps Integration (cluster import)
- **Wave 5**: Gitea (internal Git service)
- **Wave 5**: Cluster Bootstrap (self-referential ApplicationSets)

### 4. Wait for Core Infrastructure
```bash
./bin/wait-kube route openshift-gitops-server openshift-gitops '{.metadata.name}' openshift-gitops-server
./bin/wait-kube mch multiclusterhub open-cluster-management '{.status.conditions[?(@.type=="Complete")].message}' "All hub components ready."
```

### 5. Verify Deployment
```bash
oc get applications -n openshift-gitops    # All "Synced" and "Healthy"
oc get multiclusterhub -n open-cluster-management
oc get pods -n vault
```

### 6. Configure Vault (Automated)
```bash
./bin/bootstrap-vault
```

### 7. Access Management Interfaces
```bash
echo "OpenShift: $(oc whoami --show-console)"
echo "ArgoCD: https://$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}')"
echo "ACM: https://$(oc get route multicloud-console -n open-cluster-management -o jsonpath='{.spec.host}')"
echo "Gitea: https://$(oc get route gitea -n gitea-system -o jsonpath='{.spec.host}')"
```

## What Happens During Bootstrap

1. **GitOps Installation**: ArgoCD operator installs and configures itself
2. **Ordered Application Deployment**: Sync waves ensure proper dependency order
3. **Self-Referential Setup**: Cluster prepares to manage additional clusters
4. **Internal Git Service**: Gitea deployed for cluster-specific configurations
5. **Secret Management**: Vault and ESO configured for secure credential handling
6. **Multi-Cluster Ready**: ACM hub ready to manage regional clusters

## Next Steps

After bootstrap completes:

```bash
# Add your first cluster
./bin/cluster-create

# Monitor cluster provisioning
oc get clusterdeployments -A     # OpenShift clusters
oc get clusters -A               # EKS clusters

# Check overall health
./bin/monitor-health
```

The cluster is now ready to provision and manage additional regional clusters through GitOps automation.