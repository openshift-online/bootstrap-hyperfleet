# Deploy Your First Cluster

**Audience**: New users  
**Complexity**: Beginner  
**Estimated Time**: 45 minutes  
**Prerequisites**: Hub cluster running, AWS credentials configured

## Overview

This hands-on guide walks you through deploying your first regional cluster using the automated tools, from initial specification to running services.

## Prerequisites Check

Before starting, verify you have:

```bash
# 1. Access to hub cluster
oc whoami

# 2. AWS credentials available  
aws sts get-caller-identity

# 3. Required tools installed
which kustomize
which kubectl
which oc
```

## Step 1: Create Cluster Specification

Use the interactive tool to create your cluster configuration:

```bash
./bin/new-cluster
```

**Example interaction:**
```
OpenShift Regional Cluster Generator
===================================

Please provide the following information for your new cluster:

Cluster Name: my-first-cluster
Cluster Type (ocp/eks) [ocp]: eks
Region [us-west-2]: us-east-1  
Base Domain [rosa.mturansk-test.csu2.i3.devshift.org]: ✓
Instance Type [m5.2xlarge]: m5.large
Number of Replicas [2]: 3

Configuration Summary:
=====================
Cluster Name: my-first-cluster
Type: eks
Region: us-east-1
Domain: rosa.mturansk-test.csu2.i3.devshift.org  
Instance Type: m5.large
Replicas: 3

Proceed with cluster generation? (y/N): y
```

**What this creates:**
- Regional specification: `regions/us-east-1/my-first-cluster/region.yaml`
- Complete cluster overlays in multiple directories
- ArgoCD ApplicationSet for GitOps deployment

## Step 2: Review Generated Configuration

Examine what was created:

```bash
# View the regional specification
cat regions/us-east-1/my-first-cluster/region.yaml

# Check generated cluster resources
ls -la clusters/my-first-cluster/

# Check pipeline configurations  
ls -la pipelines/hello-world/my-first-cluster/
ls -la pipelines/cloud-infrastructure-provisioning/my-first-cluster/

# Check operator deployment
ls -la operators/openshift-pipelines/my-first-cluster/

# Check service deployments
ls -la deployments/ocm/my-first-cluster/

# View the ApplicationSet
cat gitops-applications/my-first-cluster.yaml
```

## Step 3: Validate Configuration

The tool automatically validates during generation, but you can run additional checks:

```bash
# Validate cluster overlay builds correctly
kustomize build clusters/my-first-cluster/

# Validate deployments overlay  
kustomize build deployments/ocm/my-first-cluster/

# Validate GitOps applications
kustomize build gitops-applications/

# Dry-run validation
oc --dry-run=client apply -k clusters/my-first-cluster/
```

## Step 4: Deploy via GitOps

Launch the deployment:

```bash
./bin/bootstrap.sh
```

**This triggers:**
1. ApplicationSet creation in ArgoCD
2. Automatic generation of 5 applications for your cluster
3. Ordered deployment via sync waves

## Step 5: Monitor Deployment Progress

Track the deployment as it progresses through sync waves:

```bash
# Overall environment status
./bin/health-check

# Watch ApplicationSet creation
oc get applicationset my-first-cluster-applications -n openshift-gitops -w

# Monitor individual applications
oc get applications -n openshift-gitops | grep my-first-cluster

# Check sync wave progression
oc get application my-first-cluster-cluster -n openshift-gitops -o yaml | grep wave
oc get application my-first-cluster-operators -n openshift-gitops -o yaml | grep wave
```

## Step 6: Verify Cluster Provisioning

### For EKS Clusters:
```bash
# Check CAPI resources
oc get awsmanagedcontrolplane my-first-cluster -n my-first-cluster
oc get awsmanagedmachinepool my-first-cluster -n my-first-cluster

# Monitor provisioning progress
oc describe awsmanagedcontrolplane my-first-cluster -n my-first-cluster

# Wait for cluster to be ready (this takes ~15-20 minutes)
./wait.kube.sh awsmanagedcontrolplane my-first-cluster my-first-cluster '{.status.ready}' true 1800
```

### For OCP Clusters:
```bash
# Check Hive resources
oc get clusterdeployment my-first-cluster -n my-first-cluster
oc get machinepool my-first-cluster-worker -n my-first-cluster

# Monitor provisioning progress  
oc describe clusterdeployment my-first-cluster -n my-first-cluster

# Wait for web console URL (this takes ~30-45 minutes)
./wait.kube.sh clusterdeployment my-first-cluster my-first-cluster '{.status.webConsoleURL}' 
```

## Step 7: Verify ACM Integration

Check that ACM properly imports and manages your cluster:

```bash
# Check ManagedCluster creation
oc get managedcluster my-first-cluster

# Verify KlusterletAddonConfig (enables ApplicationManager)
oc get klusterletaddonconfig my-first-cluster -n my-first-cluster

# Check GitOpsCluster for automatic ArgoCD registration
oc get gitopscluster -n openshift-gitops

# Verify ArgoCD can see the cluster
oc get secret -n openshift-gitops -l argocd.argoproj.io/secret-type=cluster | grep my-first-cluster
```

## Step 8: Verify Application Deployment

Once the cluster is provisioned, check that applications deploy:

```bash
# Monitor application sync status
oc get applications -n openshift-gitops | grep my-first-cluster

# Check operators deployment (Wave 2)
oc get application my-first-cluster-operators -n openshift-gitops

# Check pipeline deployment (Wave 3)
oc get application my-first-cluster-pipelines-hello-world -n openshift-gitops  
oc get application my-first-cluster-pipelines-cloud-infrastructure-provisioning -n openshift-gitops

# Check service deployment (Wave 4)
oc get application my-first-cluster-deployments-ocm -n openshift-gitops
```

## Step 9: Access Your Cluster

### Get Cluster Access:
```bash
# For EKS clusters
aws eks update-kubeconfig --region us-east-1 --name my-first-cluster

# For OCP clusters  
oc extract secret/my-first-cluster-admin-kubeconfig -n my-first-cluster --to=- > my-first-cluster-kubeconfig
export KUBECONFIG=my-first-cluster-kubeconfig
```

### Verify Deployed Services:
```bash
# Check OpenShift Pipelines operator
oc get subscription openshift-pipelines-operator-rh -n openshift-operators

# Check deployed pipelines
oc get pipeline -n ocm-my-first-cluster

# Check deployed services  
oc get deployment -n ocm-my-first-cluster

# Check configmaps for database configurations
oc get configmap -n ocm-my-first-cluster
```

## Step 10: Run Health Check

Generate a complete status report:

```bash
# Switch back to hub cluster context
oc config use-context <hub-cluster-context>

# Run comprehensive health check
./bin/health-check

# Check specific cluster status
grep -A 20 "my-first-cluster" STATUS.md
```

## Success Criteria

Your deployment is successful when:

- ✅ **Cluster Provisioned**: Ready status within 45 minutes
- ✅ **ACM Import**: ManagedCluster shows "Available"  
- ✅ **ArgoCD Registration**: Cluster appears in ArgoCD
- ✅ **Applications Synced**: All 5 applications show "Synced/Healthy"
- ✅ **Services Running**: Pipelines and deployments active on target cluster
- ✅ **Health Check**: No errors in status report

## Troubleshooting

### Common Issues:

**Cluster Provisioning Stuck:**
```bash
# Check AWS credentials and quotas
aws sts get-caller-identity
aws servicequotas get-service-quota --service-code eks --quota-code L-1194D53C

# Check CAPI controller logs
oc logs -n capi-aws-system deployment/capa-controller-manager
```

**Application Sync Failures:**
```bash
# Check application status and events
oc describe application my-first-cluster-cluster -n openshift-gitops

# Check ArgoCD controller logs
oc logs -n openshift-gitops deployment/openshift-gitops-application-controller
```

**ACM Import Issues:**
```bash
# Check ManagedCluster status
oc describe managedcluster my-first-cluster

# Check ACM hub controller logs
oc logs -n open-cluster-management deployment/multicluster-operators-hub-registration
```

## Next Steps

Now that you have a working cluster:

1. **Explore Services**: Connect to your cluster and explore deployed applications
2. **Add More Clusters**: Repeat process for additional regions
3. **Customize**: Modify base templates for your specific needs
4. **Monitor**: Set up regular health checking and alerting
5. **Scale**: Add additional services and pipelines

## Related Documentation

- **[Monitoring Guide](../operations/monitoring.md)** - Ongoing cluster management
- **[Cluster Management](../operations/cluster-management.md)** - Day-to-day operations  
- **[Troubleshooting](../operations/troubleshooting.md)** - Common issues and solutions
- **[Architecture Deep Dive](../architecture/gitops-flow.md)** - Technical details