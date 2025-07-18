# New Regional Deployment Test Plan

This document outlines the test plan for defining and implementing a new regional deployment in the OpenShift bootstrap project using the automated `bin/generate-cluster` tool.

## Rules

1. **Passing Tests Required**: Always stop the test plan on any error. Reset and start again.

## Overview

This test plan validates the automated process of creating a new regional deployment that includes:
1. Cluster provisioning via CAPI (Cluster API) or Hive
2. Regional service deployments
3. GitOps integration with ArgoCD
4. ACM multi-cluster management
5. Pipeline deployments (Hello World and Cloud Infrastructure)
6. Operator deployments (OpenShift Pipelines)

## Interactive Configuration

Before proceeding with the test plan, please provide the following information:

1. **Cluster Type**: 
   - [ ] OCP (OpenShift Container Platform) - uses ClusterDeployment + MachinePool
   - [ ] EKS (Amazon Elastic Kubernetes Service) - uses AWSManagedControlPlane + AWSManagedMachinePool

2. **AWS Region**: _________________ (e.g., us-east-1, us-west-2)

3. **Compute Type**: _________________ (e.g., m5.large, t3.medium)

4. **Cluster Name**: _________________ (e.g., cluster-40, cluster-50)

5. **Node Count**: _________________ (e.g., 3, 5)

## Prerequisites

- [ ] Bootstrap control plane cluster running with cluster-admin access
- [ ] AWS credentials configured and available
- [ ] OpenShift pull secrets configured (if OCP cluster type)
- [ ] ACM and GitOps operators installed
- [ ] CAPI AWS provider installed
- [ ] Secrets stored in `secrets/aws-credentials.yaml` and `secrets/pull-secret.yaml`

## Test Scenarios

### Scenario 1: Create Regional Specification

**Objective**: Create regional specification directory and configuration file

**Steps**:
1. [ ] Create regional spec directory: `mkdir -p regions/[AWS_REGION]/[CLUSTER_NAME]/`
2. [ ] Create region.yaml specification file:
```yaml
# regions/[AWS_REGION]/[CLUSTER_NAME]/region.yaml
type: [CLUSTER_TYPE]  # "ocp" or "eks"
name: [CLUSTER_NAME]
region: [AWS_REGION]
domain: rosa.mturansk-test.csu2.i3.devshift.org
instanceType: [COMPUTE_TYPE]
replicas: [NODE_COUNT]
```
3. [ ] Validate specification file syntax: `cat regions/[AWS_REGION]/[CLUSTER_NAME]/region.yaml`

**Expected Results**:
- [ ] Regional specification directory created
- [ ] region.yaml contains correct cluster configuration
- [ ] All required fields populated with user-provided values

### Scenario 2: Generate Complete Cluster Overlay

**Objective**: Use automated tool to generate all required overlays and GitOps resources

**Steps**:
1. [ ] Run cluster generator: `./bin/generate-cluster regions/[AWS_REGION]/[CLUSTER_NAME]/`
2. [ ] Verify cluster overlay creation: `ls -la clusters/[CLUSTER_NAME]/`
3. [ ] Verify pipeline overlays creation: 
   - [ ] `ls -la pipelines/hello-world/[CLUSTER_NAME]/`
   - [ ] `ls -la pipelines/cloud-infrastructure-provisioning/[CLUSTER_NAME]/`
4. [ ] Verify operators overlay creation: `ls -la operators/openshift-pipelines/[CLUSTER_NAME]/`
5. [ ] Verify deployments overlay creation: `ls -la deployments/ocm/[CLUSTER_NAME]/`
6. [ ] Verify ApplicationSet creation: `ls -la gitops-applications/[CLUSTER_NAME].yaml`

**Expected Results**:
- [ ] All overlay directories created successfully
- [ ] Cluster overlay contains type-appropriate resources (OCP: install-config.yaml + patches, EKS: CAPI resources)
- [ ] Pipeline overlays contain PipelineRun resources with correct parameters
- [ ] Operators overlay configured for OpenShift Pipelines deployment
- [ ] Deployments overlay configured for OCM services in `ocm-[CLUSTER_NAME]` namespace
- [ ] ApplicationSet contains all components with proper sync waves
- [ ] GitOps kustomization.yaml updated with new ApplicationSet

### Scenario 3: Validate Generated Manifests

**Objective**: Ensure all generated Kustomize overlays build successfully

**Steps**:
1. [ ] Test cluster overlay build: `kustomize build clusters/[CLUSTER_NAME]/`
2. [ ] Test hello-world pipeline build: `kustomize build pipelines/hello-world/[CLUSTER_NAME]/`
3. [ ] Test cloud-infrastructure pipeline build: `kustomize build pipelines/cloud-infrastructure-provisioning/[CLUSTER_NAME]/`
4. [ ] Test operators overlay build: `kustomize build operators/openshift-pipelines/[CLUSTER_NAME]/`
5. [ ] Test deployments overlay build: `kustomize build deployments/ocm/[CLUSTER_NAME]/`
6. [ ] Test GitOps applications build: `kustomize build gitops-applications/`
7. [ ] Dry-run validation: `oc --dry-run=client apply -k clusters/[CLUSTER_NAME]/`

**Expected Results**:
- [ ] All kustomize builds succeed without errors
- [ ] Generated manifests contain correct cluster name, region, and compute type
- [ ] Resource names and namespaces follow expected patterns
- [ ] ApplicationSet sync waves configured correctly (1: cluster, 2: operators, 3: pipelines, 4: deployments)
- [ ] Dry-run validation passes for all overlays

### Scenario 4: Test GitOps Integration

**Objective**: Verify GitOps workflow deployment using ApplicationSet

**Steps**:
1. [ ] Run bootstrap: `./bootstrap.sh`
2. [ ] Check ApplicationSet status: `oc get applicationset [CLUSTER_NAME]-applications -n openshift-gitops`
3. [ ] Check generated applications: `oc get applications -n openshift-gitops | grep [CLUSTER_NAME]`
4. [ ] Verify sync wave ordering:
   - [ ] Wave 1 (cluster): `oc get application [CLUSTER_NAME]-cluster -n openshift-gitops`
   - [ ] Wave 2 (operators): `oc get application [CLUSTER_NAME]-operators -n openshift-gitops`
   - [ ] Wave 3 (pipelines): `oc get application [CLUSTER_NAME]-pipelines-* -n openshift-gitops`
   - [ ] Wave 4 (deployments): `oc get application [CLUSTER_NAME]-deployments-ocm -n openshift-gitops`

#### For OCP Cluster Type:
5. [ ] Verify cluster provisioning: `oc get clusterdeployment [CLUSTER_NAME] -n [CLUSTER_NAME]`
6. [ ] Check MachinePool status: `oc get machinepool [CLUSTER_NAME]-worker -n [CLUSTER_NAME]`
7. [ ] Verify ACM import: `oc get klusterletaddonconfig [CLUSTER_NAME] -n [CLUSTER_NAME]`

#### For EKS Cluster Type:
5. [ ] Verify cluster provisioning: `oc get awsmanagedcontrolplane [CLUSTER_NAME] -n [CLUSTER_NAME]`
6. [ ] Check MachinePool status: `oc get awsmanagedmachinepool [CLUSTER_NAME] -n [CLUSTER_NAME]`
7. [ ] Verify ACM import: `oc get klusterletaddonconfig [CLUSTER_NAME] -n [CLUSTER_NAME]`

**Common Steps**:
8. [ ] Check ManagedCluster creation: `oc get managedcluster [CLUSTER_NAME]`

**Expected Results**:
- [ ] ApplicationSet creates all required applications with proper sync waves
- [ ] Cluster resources in provisioning state (ClusterDeployment for OCP, AWSManagedControlPlane for EKS)
- [ ] ManagedCluster resource created
- [ ] KlusterletAddonConfig with applicationManager enabled
- [ ] All applications syncing according to wave ordering
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

### Scenario 6: Test Pipeline and Service Deployment

**Objective**: Verify pipelines and regional services deploy to new cluster

**Steps**:
1. [ ] Wait for cluster provisioning completion
2. [ ] Check operators deployment: `oc get application [CLUSTER_NAME]-operators -n openshift-gitops`
3. [ ] Check pipeline deployments: 
   - [ ] `oc get application [CLUSTER_NAME]-pipelines-hello-world -n openshift-gitops`
   - [ ] `oc get application [CLUSTER_NAME]-pipelines-cloud-infrastructure-provisioning -n openshift-gitops`
4. [ ] Check regional deployment application: `oc get application [CLUSTER_NAME]-deployments-ocm -n openshift-gitops`
5. [ ] Verify OpenShift Pipelines operator: `oc get subscription openshift-pipelines-operator-rh -n openshift-operators --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig`
6. [ ] Verify pipelines deployed: `oc get pipeline -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig`
7. [ ] Verify services deployed: `oc get deployment -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig`
8. [ ] Check database configurations: `oc get configmap -n ocm-[CLUSTER_NAME] --kubeconfig=/path/to/[CLUSTER_NAME]/kubeconfig`

**Expected Results**:
- [ ] All ApplicationSet-generated applications synced successfully
- [ ] OpenShift Pipelines operator installed and healthy
- [ ] Hello World and Cloud Infrastructure pipelines deployed
- [ ] AMS, CS, and OSL services deployed in ocm-[CLUSTER_NAME] namespace
- [ ] Database configurations applied
- [ ] All services healthy and running

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
1. [ ] Delete ApplicationSet and all generated applications: `oc delete applicationset [CLUSTER_NAME]-applications -n openshift-gitops`
2. [ ] Delete cluster namespace: `oc delete namespace [CLUSTER_NAME]`
3. [ ] Remove overlay directories: 
   - [ ] `rm -rf ./clusters/[CLUSTER_NAME]`
   - [ ] `rm -rf ./pipelines/hello-world/[CLUSTER_NAME]`
   - [ ] `rm -rf ./pipelines/cloud-infrastructure-provisioning/[CLUSTER_NAME]`
   - [ ] `rm -rf ./operators/openshift-pipelines/[CLUSTER_NAME]`
   - [ ] `rm -rf ./deployments/ocm/[CLUSTER_NAME]`
4. [ ] Remove ApplicationSet file: `rm ./gitops-applications/[CLUSTER_NAME].yaml`
5. [ ] Remove regional specification: `rm -rf ./regions/[AWS_REGION]/[CLUSTER_NAME]`
6. [ ] Update gitops-applications/kustomization.yaml to remove ApplicationSet reference:
   ```bash
   # Remove line: - ./[CLUSTER_NAME].yaml
   ```

**Expected Results**:
- [ ] All [CLUSTER_NAME] resources cleaned up
- [ ] All generated overlay directories removed
- [ ] ApplicationSet and generated applications deleted
- [ ] Regional specification directory removed
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

- This test plan uses the automated `bin/generate-cluster` tool for consistent overlay generation
- Replace [CLUSTER_NAME], [AWS_REGION], [COMPUTE_TYPE], and [NODE_COUNT] with actual values from Interactive Configuration
- Regional specifications are stored in `regions/[AWS_REGION]/[CLUSTER_NAME]/region.yaml`
- OCP clusters use Hive ClusterDeployment + MachinePool resources with Kustomize patches
- EKS clusters use CAPI AWSManagedControlPlane + AWSManagedMachinePool resources (v1beta2)
- ApplicationSets replace individual ArgoCD applications for better management
- Sync waves ensure proper deployment ordering: cluster → operators → pipelines → deployments
- All pipelines and services deploy to the `ocm-[CLUSTER_NAME]` namespace
- Always test in non-production environment first
- Keep backups of working configurations
- Document any deviations from the plan

## Example Usage

Based on your Interactive Configuration, the complete process would be:

### Step 1: Create Regional Specification
```bash
# Example: EKS cluster in us-west-2
mkdir -p regions/us-west-2/cluster-50/
cat > regions/us-west-2/cluster-50/region.yaml << EOF
type: eks
name: cluster-50
region: us-west-2
domain: rosa.mturansk-test.csu2.i3.devshift.org
instanceType: m5.large
replicas: 3
EOF
```

### Step 2: Generate All Overlays and GitOps Resources
```bash
./bin/generate-cluster regions/us-west-2/cluster-50/
```

### Step 3: Deploy via GitOps
```bash
./bootstrap.sh
```

This automated approach replaces all manual copying and updating steps, ensuring consistency and reducing errors.