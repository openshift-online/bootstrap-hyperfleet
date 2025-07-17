# New Regional Deployment Test Plan

This document outlines the test plan for defining and implementing a new regional deployment in the OpenShift bootstrap project.

## Rules

1. **Passing Tests Required**: Always stop the test plan on any error. Reset and start again.

## Overview

This test plan validates the process of creating a new regional deployment that includes:
1. Cluster provisioning via CAPI (Cluster API)
2. Regional service deployments
3. GitOps integration with ArgoCD
4. ACM multi-cluster management

## Interactive Configuration

Before proceeding with the test plan, please provide the following information:

1. **Cluster Type**: 
   - [ ] OCP (OpenShift Container Platform) - uses ClusterDeployment + MachinePool
   - [ ] EKS (Amazon Elastic Kubernetes Service) - uses AWSManagedControlPlane + AWSManagedMachinePool

2. **AWS Region**: _________________ (e.g., us-east-1, us-west-2)

3. **Compute Type**: _________________ (e.g., m5.large, t3.medium)

4. **Cluster Name**: _________________ (e.g., cluster-40, cluster-50)

## Prerequisites

- [ ] Bootstrap control plane cluster running with cluster-admin access
- [ ] AWS credentials configured and available
- [ ] OpenShift pull secrets configured (if OCP cluster type)
- [ ] ACM and GitOps operators installed
- [ ] CAPI AWS provider installed
- [ ] Secrets stored in `secrets/aws-credentials.yaml` and `secrets/pull-secret.yaml`

## Test Scenarios

### Scenario 1: Create New Cluster Overlay

**Objective**: Create a new cluster overlay configuration based on user input

**Steps**:

#### For OCP Cluster Type:
1. [ ] Copy existing OCP overlay: `cp -r ./clusters/overlay/cluster-20 ./clusters/overlay/[CLUSTER_NAME]`
2. [ ] Update cluster references: Find/Replace 'cluster-20' with '[CLUSTER_NAME]' in all files
3. [ ] Update AWS region in install-config.yaml to '[AWS_REGION]'
4. [ ] Update compute instance type to '[COMPUTE_TYPE]' in install-config.yaml
5. [ ] Verify ClusterDeployment and MachinePool patches reference correct cluster name
6. [ ] Test kustomize build: `kustomize build clusters/overlay/[CLUSTER_NAME]`

#### For EKS Cluster Type:
1. [ ] Create new EKS overlay directory: `mkdir -p ./clusters/overlay/[CLUSTER_NAME]`
2. [ ] Create base EKS manifests with AWSManagedControlPlane and AWSManagedMachinePool
3. [ ] Configure AWS region '[AWS_REGION]' in AWSManagedControlPlane
4. [ ] Set compute instance type '[COMPUTE_TYPE]' in AWSManagedMachinePool
5. [ ] Create kustomization.yaml with EKS-specific patches
6. [ ] Test kustomize build: `kustomize build clusters/overlay/[CLUSTER_NAME]`

**Expected Results**:
- [ ] New cluster overlay directory created
- [ ] All cluster references updated correctly
- [ ] AWS region and compute type configured
- [ ] Kustomize build produces valid manifests without errors
- [ ] Generated manifests contain correct cluster name, region, and compute type

### Scenario 2: Create Regional Deployment Overlay

**Objective**: Create regional services overlay for the new cluster

**Steps**:
1. [ ] Copy existing regional overlay: `cp -r ./regional-deployments/overlays/region-01 ./regional-deployments/overlays/[CLUSTER_NAME]`
2. [ ] Update namespace in kustomization.yaml to `ocm-[CLUSTER_NAME]`
3. [ ] Update namespace.yaml to create `ocm-[CLUSTER_NAME]` namespace
4. [ ] Test kustomize build: `kustomize build regional-deployments/overlays/[CLUSTER_NAME]`

**Expected Results**:
- [ ] New regional deployment overlay created
- [ ] Namespace correctly set to `ocm-[CLUSTER_NAME]`
- [ ] Kustomize build produces valid regional service manifests
- [ ] AMS, CS, and OSL database configurations included

### Scenario 3: Create ArgoCD Applications

**Objective**: Create ArgoCD applications for cluster and regional deployments

**Steps**:
1. [ ] Copy and modify cluster application: `cp ./gitops-applications/regional-clusters.cluster-10.application.yaml ./gitops-applications/regional-clusters.[CLUSTER_NAME].application.yaml`
2. [ ] Update application name to `regional-[CLUSTER_NAME]`
3. [ ] Update source path to `clusters/overlay/[CLUSTER_NAME]`
4. [ ] Copy and modify regional deployment application: `cp ./gitops-applications/regional-deployments.cluster-10.application.yaml ./gitops-applications/regional-deployments.[CLUSTER_NAME].application.yaml`
5. [ ] Update application name to `regional-deployments-[CLUSTER_NAME]`
6. [ ] Update destination server URL for [CLUSTER_NAME] (will be populated after cluster creation)
7. [ ] Update source path to `regional-deployments/overlays/[CLUSTER_NAME]`
8. [ ] Add new applications to `gitops-applications/kustomization.yaml`

**Expected Results**:
- [ ] Two new ArgoCD applications created
- [ ] Applications reference correct source paths
- [ ] Regional deployment targets correct cluster API endpoint
- [ ] Applications included in main kustomization

### Scenario 4: Test GitOps Integration

**Objective**: Verify GitOps workflow deployment

**Steps**:
1. [ ] Run bootstrap: `./bootstrap.sh`
2. [ ] Check ArgoCD application status: `oc get applications -n openshift-gitops`

#### For OCP Cluster Type:
3. [ ] Verify cluster provisioning: `oc get clusterdeployment [CLUSTER_NAME] -n [CLUSTER_NAME]`
4. [ ] Check MachinePool status: `oc get machinepool [CLUSTER_NAME]-worker -n [CLUSTER_NAME]`
5. [ ] Verify ACM import: `oc get klusterletaddonconfig [CLUSTER_NAME] -n [CLUSTER_NAME]`

#### For EKS Cluster Type:
3. [ ] Verify cluster provisioning: `oc get awsmanagedcontrolplane [CLUSTER_NAME] -n [CLUSTER_NAME]`
4. [ ] Check MachinePool status: `oc get awsmanagedmachinepool [CLUSTER_NAME]-worker -n [CLUSTER_NAME]`
5. [ ] Verify ACM import: `oc get klusterletaddonconfig [CLUSTER_NAME] -n [CLUSTER_NAME]`

**Common Steps**:
6. [ ] Check ManagedCluster creation: `oc get managedcluster [CLUSTER_NAME]`

**Expected Results**:
- [ ] ArgoCD applications created and syncing
- [ ] Cluster resources in provisioning state (ClusterDeployment for OCP, AWSManagedControlPlane for EKS)
- [ ] ManagedCluster resource created
- [ ] KlusterletAddonConfig with applicationManager enabled
- [ ] Cluster appears in ACM console

### Scenario 5: Test ACM GitOps Integration

**Objective**: Verify automated ArgoCD cluster registration

**Steps**:
1. [ ] Check GitOpsCluster status: `oc get gitopscluster -n openshift-gitops`
2. [ ] Verify Placement selection: `oc get placementdecision -n openshift-gitops`
3. [ ] Check ArgoCD cluster secret creation: `oc get secret -n openshift-gitops -l argocd.argoproj.io/secret-type=cluster`
4. [ ] Verify cluster appears in ArgoCD UI

**Expected Results**:
- [ ] GitOpsCluster shows healthy status
- [ ] Placement selects [CLUSTER_NAME] based on labels
- [ ] ArgoCD cluster secret automatically created
- [ ] [CLUSTER_NAME] visible in ArgoCD clusters list

### Scenario 6: Test Regional Service Deployment

**Objective**: Verify regional services deploy to new cluster

**Steps**:
1. [ ] Wait for cluster provisioning completion
2. [ ] Check regional deployment application sync: `oc get application regional-deployments-[CLUSTER_NAME] -n openshift-gitops`
3. [ ] Verify services deployed: `oc get deployment -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig`
4. [ ] Check database configurations: `oc get configmap -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig`

**Expected Results**:
- [ ] Regional deployment application synced successfully
- [ ] AMS, CS, and OSL services deployed
- [ ] Database configurations applied
- [ ] Services healthy and running

### Scenario 7: Test Status Monitoring

**Objective**: Verify monitoring and status scripts work

**Steps**:
1. [ ] Add [CLUSTER_NAME] to `./bootstrap.sh` monitoring
2. [ ] Run status check: `./status.sh applications.argoproj.io`

#### For OCP Cluster Type:
3. [ ] Check cluster-specific status: `./status.sh clusterdeployment [CLUSTER_NAME] [CLUSTER_NAME]`
4. [ ] Verify wait script: `./wait.kube.sh clusterdeployment [CLUSTER_NAME] [CLUSTER_NAME] {.status.webConsoleURL}`

#### For EKS Cluster Type:
3. [ ] Check cluster-specific status: `./status.sh awsmanagedcontrolplane [CLUSTER_NAME] [CLUSTER_NAME]`
4. [ ] Verify wait script: `./wait.kube.sh awsmanagedcontrolplane [CLUSTER_NAME] [CLUSTER_NAME] {.status.ready}`

**Expected Results**:
- [ ] Bootstrap script monitors [CLUSTER_NAME]
- [ ] Status scripts report [CLUSTER_NAME] status
- [ ] Wait scripts can monitor [CLUSTER_NAME] resources
- [ ] All monitoring commands work without errors

## Rollback Procedures

### Rollback Scenario 1: Remove Failed Cluster

**Steps**:
1. [ ] Delete ArgoCD applications: `oc delete application regional-[CLUSTER_NAME] regional-deployments-[CLUSTER_NAME] -n openshift-gitops`
2. [ ] Delete cluster namespace: `oc delete namespace [CLUSTER_NAME]`
3. [ ] Remove overlay directories: `rm -rf ./clusters/overlay/[CLUSTER_NAME] ./regional-deployments/overlays/[CLUSTER_NAME]`
4. [ ] Remove application files: `rm ./gitops-applications/regional-*[CLUSTER_NAME].application.yaml`
5. [ ] Update kustomization.yaml to remove references

**Expected Results**:
- [ ] All [CLUSTER_NAME] resources cleaned up
- [ ] No residual configurations remain
- [ ] System returns to previous state

## Validation Checklist

### Pre-Deployment Validation
- [ ] Kustomize builds succeed for all overlays
- [ ] YAML manifests are valid
- [ ] No naming conflicts with existing resources
- [ ] All required secrets are available

### Post-Deployment Validation
- [ ] Cluster provisioning completes successfully
- [ ] ACM imports cluster for management
- [ ] ArgoCD automatically discovers cluster
- [ ] Regional services deploy and are healthy
- [ ] Monitoring scripts work correctly

### Security Validation
- [ ] No secrets exposed in manifests
- [ ] RBAC permissions are appropriate
- [ ] TLS certificates are valid
- [ ] Network policies are enforced

## Performance Considerations

### Resource Requirements
- [ ] Cluster provisioning time: < 45 minutes
- [ ] ArgoCD sync time: < 5 minutes
- [ ] Regional service deployment: < 10 minutes
- [ ] ACM import time: < 2 minutes

### Scaling Limits
- [ ] Maximum clusters per region: 10
- [ ] Maximum regional deployments: 50
- [ ] ArgoCD application limits: 100

## Common Issues and Troubleshooting

### Issue 1: Cluster Provisioning Fails
**Symptoms**: 
- OCP: ClusterDeployment stuck in provisioning
- EKS: AWSManagedControlPlane stuck in provisioning
**Resolution**: Check AWS credentials, quotas, and CAPI provider logs

### Issue 2: ArgoCD Sync Fails
**Symptoms**: Application shows out-of-sync or error state
**Resolution**: Verify manifests, check permissions, review ArgoCD controller logs

### Issue 3: ACM Import Fails
**Symptoms**: ManagedCluster not created or in unknown state
**Resolution**: Check KlusterletAddonConfig, verify network connectivity, review ACM operator logs

### Issue 4: Regional Services Not Deploying
**Symptoms**: Services not appearing in target cluster
**Resolution**: Check ArgoCD cluster secret, verify RBAC, review application logs

## Success Criteria

A new regional deployment is considered successful when:
- [ ] Cluster provisions within 45 minutes
- [ ] ACM successfully imports cluster
- [ ] ArgoCD automatically registers cluster
- [ ] Regional services deploy and are healthy
- [ ] Monitoring and status scripts work
- [ ] All validation checks pass
- [ ] Documentation is updated

## Notes

- This test plan assumes AWS as the cloud provider
- Replace [CLUSTER_NAME], [AWS_REGION], and [COMPUTE_TYPE] with actual values from Interactive Configuration
- OCP clusters use Hive ClusterDeployment + MachinePool resources
- EKS clusters use CAPI AWSManagedControlPlane + AWSManagedMachinePool resources
- Always test in non-production environment first
- Keep backups of working configurations
- Document any deviations from the plan

## Example Overlay Creation Command

Based on your configuration, the overlay creation command would be:

**For OCP**: `cp -r ./clusters/overlay/cluster-20 ./clusters/overlay/[CLUSTER_NAME]`
**For EKS**: `mkdir -p ./clusters/overlay/[CLUSTER_NAME]` (then create EKS-specific manifests)