# EKS Cluster Provisioning Guide

## Overview

This document provides comprehensive technical documentation for Amazon Elastic Kubernetes Service (EKS) cluster provisioning within the bootstrap GitOps system. EKS clusters are provisioned using Cluster API (CAPI) and automatically integrated with Red Hat Advanced Cluster Management (ACM) for centralized management.

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                        Hub Cluster                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   ArgoCD    │  │     ACM     │  │   Tekton    │            │
│  │   GitOps    │  │     Hub     │  │  Pipelines  │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │    CAPI     │  │   Vault     │  │ External    │            │
│  │ Controller  │  │ Secrets     │  │ Secrets     │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼ Provisions & Manages
┌─────────────────────────────────────────────────────────────────┐
│                       EKS Cluster                              │
│  ┌─────────────┐                    ┌─────────────┐            │
│  │     EKS     │                    │   Worker    │            │
│  │ Control     │◄──────────────────►│   Nodes     │            │
│  │   Plane     │                    │ (EC2 Inst)  │            │
│  └─────────────┘                    └─────────────┘            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ Klusterlet  │  │   Tekton    │  │  Workload   │            │
│  │   Agent     │  │ Pipelines   │  │Applications │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

### Key Technologies

- **Cluster API (CAPI)**: Kubernetes-native cluster lifecycle management
- **CAPI AWS Provider (CAPA)**: AWS-specific infrastructure provisioning
- **Red Hat ACM**: Multi-cluster management and policy enforcement
- **ArgoCD**: GitOps continuous deployment
- **Tekton Pipelines**: CI/CD automation for cluster integration
- **External Secrets Operator**: Vault integration for credential management
- **Kustomize**: Configuration management and templating

## Prerequisites

### Infrastructure Requirements

1. **AWS Account Configuration**
   - AWS credentials with EKS cluster creation permissions
   - VPC and subnet planning (or use auto-created VPC)
   - IAM roles for EKS cluster and node groups
   - Service quotas verification (especially Elastic IPs)

2. **Hub Cluster Components**
   - Red Hat Advanced Cluster Management (ACM) operator installed
   - Cluster API controllers (multicluster-engine)
   - CAPI AWS provider (CAPA) configured
   - Tekton Pipelines operator installed
   - ArgoCD/OpenShift GitOps operator installed
   - External Secrets Operator installed

3. **Secret Management**
   - Vault instance accessible from hub cluster
   - AWS credentials stored in Vault at `secret/data/aws/credentials`
   - Pull secrets for Red Hat registry at `secret/data/clusters/pull-secret`
   - ClusterSecretStore configured and validated

### AWS Service Quotas

Critical quotas that commonly cause failures:

| Service | Quota | Requirement | Reason |
|---------|--------|-------------|---------|
| EC2 | Elastic IPs | 5+ per region | NAT gateways for private subnets |
| EKS | Clusters | 100+ per region | Default limit usually sufficient |
| EC2 | Running instances | 20+ per instance type | Worker nodes |
| VPC | VPCs per region | 5+ | If creating new VPCs |

**Critical**: Elastic IP quota is the most common failure point. Each EKS cluster requires 3-6 Elastic IPs for NAT gateways.

## Kubernetes Resource Specifications

### Regional Cluster Specification

**File**: `regions/{region}/{cluster-name}/region.yaml`

```yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: {cluster-name}
  namespace: {region}
spec:
  type: eks
  region: {aws-region}
  domain: {base-domain}
  
  compute:
    instanceType: {ec2-instance-type}  # e.g., m5.large
    replicas: {worker-count}           # e.g., 3
    
  kubernetes:
    version: "{major.minor}"           # e.g., "1.28"
```

### Generated Kubernetes Resources

The `bin/cluster-generate` script creates the following resources in order:

#### 1. Namespace and Basic Resources

```yaml
# clusters/{cluster-name}/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {cluster-name}
```

#### 2. External Secrets (Credentials)

```yaml
# clusters/{cluster-name}/external-secrets.yaml
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: aws-credentials
  namespace: {cluster-name}
spec:
  secretStoreRef:
    name: vault-cluster-store
    kind: ClusterSecretStore
  target:
    name: aws-credentials
    creationPolicy: Owner
  data:
  - secretKey: credentials
    remoteRef:
      key: secret/data/aws/credentials
      property: credentials
---
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: pull-secret
  namespace: {cluster-name}
spec:
  secretStoreRef:
    name: vault-cluster-store
    kind: ClusterSecretStore
  target:
    name: pull-secret
    creationPolicy: Owner
  data:
  - secretKey: .dockerconfigjson
    remoteRef:
      key: secret/data/clusters/pull-secret
      property: .dockerconfigjson
```

#### 3. CAPI Cluster Resources

```yaml
# clusters/{cluster-name}/cluster.yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
spec:
  clusterNetwork:
    pods:
      cidrBlocks: ["192.168.0.0/16"]
  infrastructureRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1beta2
    kind: AWSManagedControlPlane
    name: {cluster-name}
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1beta2
    kind: AWSManagedControlPlane
    name: {cluster-name}
```

```yaml
# clusters/{cluster-name}/awsmanagedcontrolplane.yaml
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: AWSManagedControlPlane
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
spec:
  region: {aws-region}
  sshKeyName: ""
  version: "v{major.minor}"  # EKS API format: v1.28
```

```yaml
# clusters/{cluster-name}/awsmanagedmachinepool.yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
kind: AWSManagedMachinePool
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
spec:
  instanceType: {ec2-instance-type}
  scaling:
    minSize: {worker-count}
    maxSize: {worker-count}
```

#### 4. CAPI MachinePool (Critical Linking Resource)

```yaml
# clusters/{cluster-name}/machinepool.yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachinePool
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
spec:
  clusterName: {cluster-name}
  replicas: {worker-count}
  template:
    spec:
      bootstrap:
        dataSecretName: ""
      clusterName: {cluster-name}
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
        kind: AWSManagedMachinePool
        name: {cluster-name}
      version: "{major.minor.patch}"  # Semantic version: 1.28.0
```

**Critical Note**: The MachinePool resource links the CAPI Cluster to the AWSManagedMachinePool. Without this resource, the EKS control plane will be created but no worker nodes will be provisioned.

#### 5. ACM Integration Resources

```yaml
# clusters/{cluster-name}/managedcluster.yaml
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
  labels:
    name: {cluster-name}
    cloud: Amazon
    region: {aws-region}
    vendor: EKS
spec:
  hubAcceptsClient: true
```

```yaml
# clusters/{cluster-name}/klusterletaddonconfig.yaml
apiVersion: agent.open-cluster-management.io/v1
kind: KlusterletAddonConfig
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
spec:
  clusterName: {cluster-name}
  clusterNamespace: {cluster-name}
  clusterLabels:
    name: {cluster-name}
    cloud: Amazon
    vendor: EKS
  applicationManager:
    enabled: true
  policyController:
    enabled: true
  searchCollector:
    enabled: true
  certPolicyController:
    enabled: true
  iamPolicyController:
    enabled: true
```

#### 6. Automated ACM Integration Pipeline

```yaml
# clusters/{cluster-name}/acm-integration-pipeline.yaml
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: klusterlet-crd-{cluster-name}
  namespace: {cluster-name}
data:
  klusterlet-crd.yaml: |
    # Clean Klusterlet CRD extracted from hub cluster
    apiVersion: apiextensions.k8s.io/v1
    kind: CustomResourceDefinition
    metadata:
      name: klusterlets.operator.open-cluster-management.io
    # ... full CRD specification
---
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: eks-acm-integration-{cluster-name}
  namespace: {cluster-name}
spec:
  params:
  - name: cluster-name
    default: "{cluster-name}"
  - name: region
    default: "{aws-region}"
  steps:
  - name: install-tools
    # Install kubectl, aws cli, oc
  - name: wait-for-eks-cluster
    # Wait for EKS cluster to become ACTIVE
  - name: configure-managed-cluster
    # Install Klusterlet CRD and configure access
  - name: apply-acm-import
    # Apply ACM import manifest
  - name: fix-pull-secret
    # Fix Red Hat registry pull secrets if needed
  - name: verify-acm-integration
    # Verify successful ACM integration
---
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: {cluster-name}-acm-integration
  namespace: {cluster-name}
  annotations:
    argocd.argoproj.io/sync-wave: "2"
spec:
  pipelineSpec:
    tasks:
    - name: integrate-with-acm
      taskRef:
        name: eks-acm-integration-{cluster-name}
  serviceAccountName: cluster-provisioner
```

## Provisioning Flow

### Phase 1: Resource Deployment (ArgoCD Sync Wave 1)

1. **Namespace Creation**: Creates dedicated namespace for cluster resources
2. **External Secrets Sync**: Vault credentials synchronized to cluster namespace
3. **CAPI Resource Creation**: Cluster API begins EKS cluster provisioning
4. **ACM Resource Creation**: ManagedCluster and addon configs created

### Phase 2: Infrastructure Provisioning (AWS)

1. **VPC and Networking**: AWS creates VPC, subnets, NAT gateways, security groups
2. **EKS Control Plane**: AWS provisions managed Kubernetes control plane
3. **IAM Resources**: AWS creates necessary service roles and policies
4. **Worker Node Groups**: EC2 instances launched and joined to cluster

**Typical Timeline**: 15-25 minutes for complete infrastructure provisioning

### Phase 3: ACM Integration (ArgoCD Sync Wave 2)

1. **Pipeline Execution**: Tekton Pipeline runs automated integration tasks
2. **Klusterlet CRD Installation**: Required CRD installed on managed cluster
3. **ACM Import**: Cluster imported into ACM hub for management
4. **Agent Deployment**: Klusterlet agent pods deployed and configured
5. **Pull Secret Fix**: Red Hat registry credentials applied if needed

**Typical Timeline**: 5-10 minutes for ACM integration

### Phase 4: Workload Deployment (ArgoCD Sync Waves 3+)

1. **Operator Installation**: OpenShift Pipelines and other operators
2. **Pipeline Deployment**: Tekton pipelines for CI/CD workflows
3. **Application Deployment**: Business applications via GitOps

## Version Format Requirements

**Critical**: EKS clusters require different version formats for different resources:

| Resource | Format | Example | Reason |
|----------|--------|---------|---------|
| AWSManagedControlPlane | `v{major.minor}` | `v1.28` | EKS API requirement |
| MachinePool | `{major.minor.patch}` | `1.28.0` | CAPI semantic versioning |
| Regional Specification | `{major.minor}` | `1.28` | User-friendly input |

The generator automatically converts between these formats.

## Common Issues and Solutions

### 1. MachinePool Version Format Error

**Error**: `json: cannot unmarshal number into Go struct field MachineSpec.spec.template.spec.version of type string`

**Cause**: MachinePool version specified as number instead of string
**Solution**: Ensure version is quoted string with semantic format: `"1.28.0"`

### 2. Klusterlet CRD Already Exists

**Error**: `AlreadyExists: customresourcedefinitions.apiextensions.k8s.io "klusterlets.operator.open-cluster-management.io" already exists`

**Cause**: Previous cluster import left CRD on managed cluster
**Solution**: Automated pipeline handles this by applying clean CRD

### 3. Klusterlet ImagePullBackOff

**Error**: Klusterlet pods fail with `ImagePullBackOff` for Red Hat registry images

**Cause**: Missing or invalid Red Hat registry credentials
**Solution**: Pipeline automatically detects and fixes pull secret issues

### 4. AWS Elastic IP Quota Exceeded

**Error**: `AddressLimitExceeded` during NAT gateway creation

**Cause**: Insufficient Elastic IP quota in AWS region
**Solution**: Request quota increase or clean up unused Elastic IPs

### 5. ApplicationSet Destination Server Not Found

**Error**: ArgoCD cannot connect to EKS cluster for workload deployment

**Cause**: Cluster not properly imported to ACM or wrong endpoint URL
**Solution**: Verify ACM integration completed successfully

### 6. CAPI Controllers Not Ready

**Error**: Cluster stuck in "Provisioning" state

**Cause**: CAPI or CAPA controllers not running properly
**Solution**: Check multicluster-engine pod status and logs

## Monitoring and Troubleshooting

### Key Resources to Monitor

1. **CAPI Cluster Status**:
   ```bash
   oc get cluster {cluster-name} -n {cluster-name}
   ```

2. **AWS EKS Cluster Status**:
   ```bash
   aws eks describe-cluster --name {cluster-name} --region {aws-region}
   ```

3. **ACM ManagedCluster Status**:
   ```bash
   oc get managedcluster {cluster-name}
   ```

4. **Pipeline Execution Status**:
   ```bash
   oc get pipelinerun {cluster-name}-acm-integration -n {cluster-name}
   ```

5. **ArgoCD Application Status**:
   ```bash
   oc get applications -n openshift-gitops | grep {cluster-name}
   ```

### Expected Status Progression

1. **Initial**: CAPI Cluster shows "Provisioning"
2. **Infrastructure Ready**: AWS EKS shows "ACTIVE"
3. **CAPI Complete**: CAPI Cluster shows "Provisioned"
4. **ACM Integration**: ManagedCluster shows "HUB ACCEPTED=true, JOINED=True, AVAILABLE=True"
5. **GitOps Ready**: ArgoCD Applications show "Synced" and "Healthy"

## Security Considerations

### Credential Management

- AWS credentials stored in Vault, never in Git
- Pull secrets managed through External Secrets Operator
- Service account tokens for pipeline automation
- RBAC permissions scoped to minimum required access

### Network Security

- EKS clusters deployed in private subnets by default
- NAT gateways provide outbound internet access
- Security groups restrict inbound traffic
- AWS VPC CNI provides pod-level networking

### Cluster Access

- EKS API server accessible via AWS IAM authentication
- kubectl access requires proper AWS credentials and RBAC
- ACM provides centralized RBAC policy enforcement

## Performance and Scaling

### Cluster Sizing Guidelines

| Use Case | Instance Type | Node Count | Notes |
|----------|---------------|------------|-------|
| Development | m5.large | 3 | Cost-effective testing |
| Production | m5.xlarge+ | 3-10 | Based on workload requirements |
| CI/CD | c5.large | 3-5 | CPU-optimized for builds |

### Scaling Considerations

- EKS clusters can scale to 1000+ nodes
- Consider AWS service limits (EC2 instances, Elastic IPs)
- Use Cluster Autoscaler for dynamic scaling
- Monitor costs with AWS Cost Explorer

## Maintenance and Lifecycle

### Updates and Patches

- EKS control plane updates managed by AWS
- Worker node updates via CAPI MachinePool rolling updates
- Kubernetes version upgrades supported
- ACM policies can enforce compliance

### Backup and Disaster Recovery

- ETCD backups managed by AWS for control plane
- Application data backups via Velero or similar
- GitOps provides configuration disaster recovery
- Multi-region deployments for high availability

### Decommissioning

1. Remove from ArgoCD ApplicationSet
2. Delete ManagedCluster from ACM
3. Delete CAPI cluster resources
4. Verify AWS resources cleaned up
5. Remove cluster directories from Git

## Best Practices

### Resource Naming

- Use descriptive, consistent naming conventions
- Include environment and purpose in names
- Follow DNS naming rules (lowercase, hyphens)
- Keep names under 63 characters for Kubernetes

### Configuration Management

- Use regional specifications for environment-specific config
- Leverage Kustomize for configuration variants
- Store sensitive data in Vault, not Git
- Use Git tags for release management

### Monitoring and Alerting

- Monitor cluster provisioning pipelines
- Set up alerts for failed cluster deployments
- Track AWS cost and quota usage
- Monitor ACM managed cluster health

### Security

- Regularly rotate AWS credentials
- Keep Kubernetes versions up to date
- Use Pod Security Standards
- Implement network policies
- Regular security scanning of container images

This documentation provides the complete technical foundation for supporting EKS clusters within the bootstrap GitOps system. For additional support, refer to the vendor documentation for CAPI, ACM, and AWS EKS.