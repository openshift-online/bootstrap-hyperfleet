# Cluster Creation Guide

**Audience**: Operators  
**Complexity**: Intermediate  
**Estimated Time**: 30 minutes setup + 1-2 hours for complete deployment  
**Prerequisites**: Running hub cluster, AWS credentials, GitOps workflow understanding

## Quick Start (3 Steps)

### 1. Generate Cluster Configuration
```bash
./bin/new-cluster
```
The interactive tool will prompt for:
- **Cluster Name** (validates uniqueness)
- **Type** ("ocp" or "eks") 
- **Region** (default: "us-west-2")
- **Domain** (default: "rosa.mturansk-test.csu2.i3.devshift.org")
- **Instance Type** (default: "m5.2xlarge")
- **Replicas** (default: "2")

### 2. Review Generated Files
The tool automatically creates:
- `regions/[Region]/[Cluster Name]/region.yaml` - Regional specification
- `clusters/[Cluster Name]/` - Cluster provisioning manifests
- `operators/openshift-pipelines/[Cluster Name]/` - Operator deployments
- `pipelines/*/[Cluster Name]/` - Pipeline configurations
- `deployments/ocm/[Cluster Name]/` - Service deployments
- `gitops-applications/[Cluster Name].yaml` - ArgoCD ApplicationSet

### 3. Deploy via GitOps
```bash
./bootstrap.sh
```

## Detailed Workflow & Validation

### Interactive Configuration
The `bin/new-cluster` tool provides guided input collection:

```
OpenShift Regional Cluster Generator
===================================

Please provide the following information for your new cluster:

Cluster Name: my-new-cluster
Cluster Type (ocp/eks) [ocp]: eks
Region [us-west-2]: us-east-1
Base Domain [rosa.mturansk-test.csu2.i3.devshift.org]: 
Instance Type [m5.2xlarge]: m5.large
Number of Replicas [2]: 3

Configuration Summary:
=====================
Cluster Name: my-new-cluster
Type: eks
Region: us-east-1
Domain: rosa.mturansk-test.csu2.i3.devshift.org
Instance Type: m5.large
Replicas: 3

Proceed with cluster generation? (y/N): y
```

### Automated Generation Process
1. **Validation**: Checks cluster name uniqueness and input validity
2. **Regional Spec**: Creates `regions/[Region]/[Cluster Name]/region.yaml`
3. **Full Generation**: Calls `bin/generate-cluster` for complete overlay creation
4. **Validation**: Automatically runs validation checks:
   - `oc kustomize clusters/[cluster-name]/`
   - `oc kustomize deployments/ocm/[cluster-name]/`
   - `oc kustomize gitops-applications/`

### Generated Regional Specification
```yaml
# regions/[Region]/[Cluster Name]/region.yaml
name: [Cluster Name]
type: [Type]  # "ocp" or "eks"
region: [Region]
domain: [Domain]
instanceType: [Instance Type]
replicas: [Replicas]
```

## Comprehensive Test Plan

### Rules
1. **Passing Tests Required**: Always stop on any error. Reset and start again.

### Test Scenarios

#### Scenario 1: Create Regional Specification
**Objective**: Create regional specification directory and configuration file

**Steps**:
1. ✅ Create regional spec directory: `mkdir -p regions/[AWS_REGION]/[CLUSTER_NAME]/`
2. ✅ Create region.yaml specification file (automated by tool)
3. ✅ Validate specification file syntax: `cat regions/[AWS_REGION]/[CLUSTER_NAME]/region.yaml`

#### Scenario 2: Generate Complete Cluster Overlay
**Objective**: Use automated tool to generate all required overlays and GitOps resources

**Steps**:
1. ✅ Run cluster generator: `./bin/generate-cluster regions/[AWS_REGION]/[CLUSTER_NAME]/`
2. ✅ Verify cluster overlay creation: `ls -la clusters/[CLUSTER_NAME]/`
3. ✅ Verify pipeline overlays creation
4. ✅ Verify operators overlay creation
5. ✅ Verify deployments overlay creation
6. ✅ Verify ApplicationSet creation

#### Scenario 3: Validate Generated Manifests
**Objective**: Ensure all generated Kustomize overlays build successfully

**Steps**:
1. ✅ Test cluster overlay build: `kustomize build clusters/[CLUSTER_NAME]/`
2. ✅ Test pipeline builds for both hello-world and cloud-infrastructure
3. ✅ Test operators overlay build
4. ✅ Test deployments overlay build
5. ✅ Test GitOps applications build
6. ✅ Dry-run validation: `oc --dry-run=client apply -k clusters/[CLUSTER_NAME]/`

#### Scenario 4: GitOps Integration Test
**Objective**: Verify GitOps workflow deployment using ApplicationSet

**Steps**:
1. ✅ Run bootstrap: `./bootstrap.sh`
2. ✅ Check ApplicationSet status
3. ✅ Check generated applications with proper sync wave ordering
4. ✅ Verify cluster provisioning (OCP: ClusterDeployment, EKS: AWSManagedControlPlane)
5. ✅ Check ACM import and ManagedCluster creation

#### Scenario 5: ACM GitOps Integration
**Objective**: Verify automated ArgoCD cluster registration

**Steps**:
1. ✅ Check GitOpsCluster status
2. ✅ Verify Placement selection
3. ✅ Check ArgoCD cluster secret creation
4. ✅ Verify cluster appears in ArgoCD UI

#### Scenario 6: Pipeline and Service Deployment
**Objective**: Verify pipelines and regional services deploy to new cluster

**Steps**:
1. ✅ Wait for cluster provisioning completion
2. ✅ Check operators deployment
3. ✅ Check pipeline deployments
4. ✅ Check regional deployment application
5. ✅ Verify OpenShift Pipelines operator installation
6. ✅ Verify pipelines and services deployed

## Cluster Types

### OCP Clusters (Hive-based)
- **Resources**: ClusterDeployment + MachinePool + InstallConfig
- **Provisioning**: Hive operator handles cluster creation
- **Features**: Full OpenShift capabilities, advanced operators

### EKS Clusters (CAPI-based)
- **Resources**: AWSManagedControlPlane + AWSManagedMachinePool (v1beta2)
- **Provisioning**: CAPI with ACM infrastructure providers
- **Features**: Kubernetes-native, cost-effective, AWS-managed control plane

## GitOps Sync Waves

ApplicationSets deploy resources in ordered waves:
1. **Wave 1**: Cluster provisioning (CAPI/Hive resources)
2. **Wave 2**: Operator installation (OpenShift Pipelines)
3. **Wave 3**: Pipeline deployment (Tekton resources)
4. **Wave 4**: Service deployment (OCM services)

## Rollback Procedures

### Remove Failed Cluster
```bash
# Delete ApplicationSet and all generated applications
oc delete applicationset [CLUSTER_NAME]-applications -n openshift-gitops

# Delete cluster namespace
oc delete namespace [CLUSTER_NAME]

# Remove overlay directories
rm -rf ./clusters/[CLUSTER_NAME]
rm -rf ./pipelines/*/[CLUSTER_NAME]
rm -rf ./operators/openshift-pipelines/[CLUSTER_NAME]
rm -rf ./deployments/ocm/[CLUSTER_NAME]

# Remove ApplicationSet file
rm ./gitops-applications/[CLUSTER_NAME].yaml

# Remove regional specification
rm -rf ./regions/[AWS_REGION]/[CLUSTER_NAME]

# Update gitops-applications/kustomization.yaml
# Remove line: - ./[CLUSTER_NAME].yaml
```

## Success Criteria

A cluster deployment is successful when:
- ✅ Cluster provisions within 45 minutes
- ✅ ACM successfully imports cluster
- ✅ ArgoCD automatically registers cluster
- ✅ Regional services deploy and are healthy
- ✅ All validation checks pass

## Error Handling & Troubleshooting

### Common Issues

**Issue 1: Cluster Provisioning Fails**
- **Symptoms**: ClusterDeployment/AWSManagedControlPlane stuck
- **Resolution**: Check AWS credentials, quotas, CAPI provider logs

**Issue 2: ArgoCD Sync Fails**
- **Symptoms**: Application shows out-of-sync or error state
- **Resolution**: Verify manifests, check permissions, review controller logs

**Issue 3: ACM Import Fails**
- **Symptoms**: ManagedCluster not created or unknown state
- **Resolution**: Check KlusterletAddonConfig, verify connectivity, review ACM logs

**Issue 4: Validation Failures**
- **Symptoms**: `kustomize build` or `oc --dry-run` fails
- **Resolution**: Check generated manifests, verify base templates, review error output

## Next Steps

After successful cluster creation:
1. **Monitor**: Use `./bin/health-check` for status monitoring
2. **Customize**: Add additional services or pipelines as needed
3. **Scale**: Repeat process for additional regions/clusters
4. **Maintain**: Regular health checks and updates

## Related Documentation

- [Architecture Overview](../ARCHITECTURE.md) - Visual architecture diagrams
- [Installation Guide](../INSTALL.md) - Hub cluster setup
- [Monitoring Guide](./monitoring.md) - Status checking procedures
- [AWS Cleanup Guide](../bin/clean-aws.md) - Resource cleanup procedures