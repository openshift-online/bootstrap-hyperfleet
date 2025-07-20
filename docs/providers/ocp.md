# OpenShift Container Platform (OCP) Cluster Provisioning Guide

## Overview

This document provides comprehensive technical documentation for Red Hat OpenShift Container Platform (OCP) cluster provisioning within the bootstrap GitOps system. OCP clusters are provisioned using the Hive operator and automatically integrated with Red Hat Advanced Cluster Management (ACM) for centralized management.

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
│  │    Hive     │  │   Vault     │  │ External    │            │
│  │  Operator   │  │ Secrets     │  │ Secrets     │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼ Provisions & Manages
┌─────────────────────────────────────────────────────────────────┐
│                    OpenShift Cluster                           │
│  ┌─────────────┐                    ┌─────────────┐            │
│  │ OpenShift   │                    │   Worker    │            │
│  │ Control     │◄──────────────────►│   Nodes     │            │
│  │ Plane       │                    │             │            │
│  └─────────────┘                    └─────────────┘            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ Klusterlet  │  │   Tekton    │  │  OpenShift  │            │
│  │   Agent     │  │ Pipelines   │  │ Workloads   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

### Key Technologies

- **Hive Operator**: OpenShift cluster lifecycle management
- **Red Hat ACM**: Multi-cluster management and policy enforcement
- **ArgoCD**: GitOps continuous deployment
- **Tekton Pipelines**: CI/CD automation for cluster workflows
- **External Secrets Operator**: Vault integration for credential management
- **OpenShift Assisted Service**: Bare metal and infrastructure installation
- **Kustomize**: Configuration management and templating

## Prerequisites

### Infrastructure Requirements

1. **Platform Support**
   - AWS (most common)
   - Azure
   - Google Cloud Platform (GCP)
   - VMware vSphere
   - OpenStack
   - Bare Metal (via Assisted Service)

2. **Hub Cluster Components**
   - Red Hat Advanced Cluster Management (ACM) operator installed
   - Hive operator installed (part of ACM)
   - Tekton Pipelines operator installed
   - ArgoCD/OpenShift GitOps operator installed
   - External Secrets Operator installed

3. **Secret Management**
   - Vault instance accessible from hub cluster
   - Cloud provider credentials stored in Vault
   - OpenShift pull secret stored in Vault at `secret/data/clusters/pull-secret`
   - SSH keys for cluster access (if required)
   - ClusterSecretStore configured and validated

### Cloud Provider Quotas (AWS Example)

| Service | Quota | Requirement | Reason |
|---------|--------|-------------|---------|
| EC2 | Running instances | 20+ per instance type | Control plane + worker nodes |
| EC2 | Elastic IPs | 10+ per region | Load balancers and NAT gateways |
| ELB | Application Load Balancers | 10+ | Ingress and API server access |
| Route53 | Hosted zones | 5+ | DNS management |
| IAM | Roles | 100+ | Service accounts and operators |

### OpenShift Version Support

- **Supported Versions**: 4.12+ (Long Term Support)
- **Recommended**: 4.14+ for latest features and security updates
- **Version Strategy**: Use even-numbered releases for production (4.12, 4.14, 4.16)
- **Upgrade Path**: Hive supports automated upgrades between compatible versions

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
  type: ocp
  region: {cloud-region}
  domain: {base-domain}
  
  compute:
    instanceType: {instance-type}     # e.g., m5.xlarge
    replicas: {worker-count}          # e.g., 3
    
  openshift:
    version: "{major.minor}"          # e.g., "4.14"
    channel: "{channel}"              # stable, fast, candidate
    
  platform:
    aws:                              # Platform-specific config
      region: {aws-region}
      instanceType: {instance-type}
      zones: ["{zone1}", "{zone2}", "{zone3}"]
```

### Generated Kubernetes Resources

The `bin/generate-cluster` script creates the following resources for OCP clusters:

#### 1. Namespace and Basic Resources

```yaml
# clusters/{cluster-name}/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {cluster-name}
  labels:
    name: {cluster-name}
    cluster.open-cluster-management.io/managedCluster: {cluster-name}
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
  - secretKey: aws_access_key_id
    remoteRef:
      key: secret/data/aws/credentials
      property: aws_access_key_id
  - secretKey: aws_secret_access_key
    remoteRef:
      key: secret/data/aws/credentials
      property: aws_secret_access_key
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

#### 3. Install Config Secret

```yaml
# clusters/{cluster-name}/install-config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: {cluster-name}-install-config
  namespace: {cluster-name}
type: Opaque
data:
  install-config.yaml: |
    apiVersion: v1
    metadata:
      name: '{cluster-name}'
    baseDomain: {base-domain}
    controlPlane:
      architecture: amd64
      hyperthreading: Enabled
      name: master
      replicas: 3
      platform:
        aws:
          zones:
            - {region}a
            - {region}b
            - {region}c
          rootVolume:
            iops: 4000
            size: 100
            type: io1
          type: {control-plane-instance-type}
    compute:
    - hyperthreading: Enabled
      architecture: amd64
      name: 'worker'
      replicas: {worker-count}
      platform:
        aws:
          rootVolume:
            iops: 2000
            size: 100
            type: io1
          type: {worker-instance-type}
          zones:
            - {region}a
            - {region}b
            - {region}c
    networking:
      networkType: OVNKubernetes
      clusterNetwork:
      - cidr: 10.128.0.0/14
        hostPrefix: 23
      serviceNetwork:
      - 172.30.0.0/16
    platform:
      aws:
        region: {region}
        userTags:
          openshift-cluster: {cluster-name}
          environment: {environment}
    pullSecret: ""  # Injected by Hive from pull-secret Secret
    sshKey: ""      # Optional SSH key for node access
```

#### 4. Hive ClusterDeployment

```yaml
# clusters/{cluster-name}/clusterdeployment.yaml
apiVersion: hive.openshift.io/v1
kind: ClusterDeployment
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
  labels:
    cloud: AWS
    region: {region}
    vendor: OpenShift
    cluster.open-cluster-management.io/clusterset: default
spec:
  baseDomain: {base-domain}
  clusterName: {cluster-name}
  controlPlaneConfig:
    servingCertificates: {}
  installAttemptsLimit: 1
  installed: false
  platform:
    aws:
      credentialsSecretRef:
        name: aws-credentials
      region: {region}
  provisioning:
    installConfigSecretRef:
      name: {cluster-name}-install-config
    sshPrivateKeySecretRef:
      name: {cluster-name}-ssh-private-key  # Optional
    imageSetRef:
      name: openshift-v{major.minor.patch}
    manifestsConfigMapRef:
      name: {cluster-name}-manifests       # Optional custom manifests
  pullSecretRef:
    name: pull-secret
```

#### 5. Hive ClusterImageSet

```yaml
# clusters/{cluster-name}/clusterimageset.yaml
apiVersion: hive.openshift.io/v1
kind: ClusterImageSet
metadata:
  name: openshift-v{major.minor.patch}
spec:
  releaseImage: quay.io/openshift-release-dev/ocp-release:{version}-{arch}
```

#### 6. ACM Integration Resources

```yaml
# clusters/{cluster-name}/managedcluster.yaml
apiVersion: cluster.open-cluster-management.io/v1
kind: ManagedCluster
metadata:
  name: {cluster-name}
  labels:
    name: {cluster-name}
    cloud: AWS
    region: {region}
    vendor: OpenShift
    cluster.open-cluster-management.io/clusterset: default
    openshiftVersion: "{major.minor}"
spec:
  hubAcceptsClient: true
  leaseDurationSeconds: 60
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
    cloud: AWS
    vendor: OpenShift
    openshiftVersion: "{major.minor}"
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
  observabilityController:
    enabled: true
```

#### 7. OpenShift-Specific Manifests (Optional)

```yaml
# clusters/{cluster-name}/manifests-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {cluster-name}-manifests
  namespace: {cluster-name}
data:
  # Custom manifests applied during installation
  cluster-monitoring-config.yaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: cluster-monitoring-config
      namespace: openshift-monitoring
    data:
      config.yaml: |
        enableUserWorkload: true
        prometheusK8s:
          retention: 7d
  
  cluster-network-operator.yaml: |
    apiVersion: operator.openshift.io/v1
    kind: Network
    metadata:
      name: cluster
    spec:
      defaultNetwork:
        type: OVNKubernetes
        ovnKubernetesConfig:
          mtu: 1450
          genevePort: 6081
```

## Provisioning Flow

### Phase 1: Resource Deployment (ArgoCD Sync Wave 1)

1. **Namespace Creation**: Creates dedicated namespace for cluster resources
2. **External Secrets Sync**: Vault credentials synchronized to cluster namespace
3. **Install Config Creation**: OpenShift installation configuration prepared
4. **Hive Resource Creation**: ClusterDeployment triggers installation process
5. **ACM Resource Creation**: ManagedCluster and addon configs created

### Phase 2: OpenShift Installation (Hive Operator)

1. **Infrastructure Provisioning**: Cloud provider resources created (VPC, instances, etc.)
2. **Bootstrap Installation**: Temporary bootstrap node provisions control plane
3. **Control Plane Deployment**: Master nodes installed and configured
4. **Bootstrap Destruction**: Bootstrap resources cleaned up automatically
5. **Worker Node Installation**: Worker nodes joined to cluster
6. **Operator Installation**: Core OpenShift operators deployed
7. **Cluster Operators**: Platform operators reach available state

**Typical Timeline**: 45-60 minutes for complete cluster installation

### Phase 3: ACM Integration (Automatic)

1. **Automatic Import**: Hive automatically imports cluster to ACM upon completion
2. **Klusterlet Deployment**: ACM agent automatically deployed via import process
3. **Addon Configuration**: ACM addons configured based on KlusterletAddonConfig
4. **Policy Enforcement**: ACM policies applied to manage cluster compliance

**Typical Timeline**: 5-10 minutes for ACM integration

### Phase 4: Workload Deployment (ArgoCD Sync Waves 2+)

1. **Operator Installation**: Additional operators via GitOps
2. **Platform Configuration**: Cluster-wide configuration and policies
3. **Application Deployment**: Business applications via GitOps

## Installation States and Monitoring

### Hive ClusterDeployment States

| State | Description | Next State | Typical Duration |
|-------|-------------|------------|------------------|
| `Installing` | Installation in progress | `Installed` or `InstallLaunchError` | 45-60 min |
| `Installed` | Installation completed successfully | Stable | - |
| `InstallLaunchError` | Failed to start installation | Manual intervention | - |
| `ProvisionStopped` | Installation stopped/cancelled | Manual restart | - |

### Key Monitoring Commands

```bash
# Check ClusterDeployment status
oc get clusterdeployment {cluster-name} -n {cluster-name}

# Check installation progress
oc describe clusterdeployment {cluster-name} -n {cluster-name}

# View installation logs
oc logs -n {cluster-name} deployment/hive-controllers -f

# Check ManagedCluster status
oc get managedcluster {cluster-name}

# Get cluster credentials
oc get secret {cluster-name}-admin-kubeconfig -n {cluster-name} -o jsonpath='{.data.kubeconfig}' | base64 -d
```

### Installation Progress Indicators

1. **Infrastructure Ready**: AWS resources created, nodes launching
2. **Bootstrap Complete**: Control plane accessible via API
3. **Control Plane Ready**: All master nodes joined and ready
4. **Workers Joined**: Worker nodes successfully joined cluster
5. **Operators Available**: All cluster operators reporting available
6. **Installation Complete**: ClusterDeployment shows `installed: true`

## Platform-Specific Configurations

### AWS Platform

```yaml
platform:
  aws:
    region: us-east-1
    userTags:
      Environment: production
      Team: platform
    subnets:
    - subnet-12345678  # Optional: use existing subnets
    - subnet-87654321
    hostedZone: Z1234567890ABC  # Optional: existing Route53 zone
    amiID: ami-12345678         # Optional: custom AMI
    serviceEndpoints:           # Optional: VPC endpoints
    - name: ec2
      url: https://vpce-12345678.ec2.us-east-1.vpce.amazonaws.com
```

### Azure Platform

```yaml
platform:
  azure:
    region: eastus
    baseDomainResourceGroupName: {resource-group}
    resourceGroupName: {cluster-resource-group}
    virtualNetwork: {vnet-name}
    controlPlaneSubnet: {control-plane-subnet}
    computeSubnet: {worker-subnet}
    outboundType: Loadbalancer
    cloudName: AzurePublicCloud
```

### Google Cloud Platform

```yaml
platform:
  gcp:
    projectID: {project-id}
    region: us-central1
    network: {vpc-network}
    controlPlaneSubnet: {control-plane-subnet}
    computeSubnet: {worker-subnet}
```

### VMware vSphere

```yaml
platform:
  vsphere:
    vcenter: {vcenter-server}
    username: {username}
    password: {password}
    datacenter: {datacenter}
    defaultDatastore: {datastore}
    cluster: {cluster}
    network: {network}
    apiVIP: {api-vip}
    ingressVIP: {ingress-vip}
```

### Bare Metal (Assisted Service)

```yaml
platform:
  baremetal:
    apiVIP: {api-vip}
    ingressVIP: {ingress-vip}
    provisioningNetwork: Unmanaged
    externalBridge: {bridge-name}
    provisioningBridge: {provisioning-bridge}
    hosts:
    - name: {hostname}
      role: master
      bmc:
        address: ipmi://{bmc-ip}
        username: {bmc-username}
        password: {bmc-password}
      bootMACAddress: {mac-address}
      hardwareProfile: default
```

## OpenShift Version Management

### Version Selection Strategy

```yaml
# Stable channel (recommended for production)
spec:
  channel: stable-4.14
  version: 4.14.15

# Fast channel (early access to updates)
spec:
  channel: fast-4.15
  version: 4.15.2

# Candidate channel (release candidates)
spec:
  channel: candidate-4.16
  version: 4.16.0-rc.1
```

### Upgrade Management

```yaml
# Automatic upgrades via ClusterVersion
apiVersion: config.openshift.io/v1
kind: ClusterVersion
metadata:
  name: version
spec:
  channel: stable-4.14
  upstream: https://api.openshift.com/api/upgrades_info/v1/graph
  desiredUpdate:
    version: 4.14.16
    force: false
```

## Common Issues and Solutions

### 1. Installation Timeout

**Error**: ClusterDeployment stuck in "Installing" state beyond 60 minutes
**Causes**: 
- Insufficient cloud quotas
- Network connectivity issues
- DNS resolution problems
**Solution**: Check cloud provider quotas, verify network configuration

### 2. Bootstrap Failure

**Error**: Installation fails during bootstrap phase
**Causes**:
- Incorrect install-config.yaml
- Missing cloud credentials
- Platform-specific configuration errors
**Solution**: Validate install-config, check credentials, review platform requirements

### 3. Worker Nodes Not Joining

**Error**: Control plane ready but worker nodes fail to join
**Causes**:
- Security group misconfigurations
- Certificate issues
- Machine config problems
**Solution**: Check security groups, verify certificates, review MachineConfigPool status

### 4. Cluster Operators Degraded

**Error**: Some cluster operators remain degraded after installation
**Causes**:
- Resource constraints
- Network policies blocking communication
- Storage issues
**Solution**: Check resource utilization, review network policies, verify storage classes

### 5. DNS Resolution Issues

**Error**: Internal DNS resolution failing
**Causes**:
- Incorrect DNS configuration
- CoreDNS operator issues
- Network policy restrictions
**Solution**: Verify DNS operator status, check CoreDNS configuration

### 6. Certificate Authority Issues

**Error**: Certificate validation failures
**Causes**:
- Clock skew between nodes
- CA certificate problems
- Certificate rotation issues
**Solution**: Sync node clocks, verify CA certificates, check certificate rotation

## Security Considerations

### Cluster Security Defaults

- **Pod Security Standards**: Restricted profile enforced by default
- **Network Policies**: Default deny-all, explicit allow required
- **RBAC**: Role-based access control enforced
- **Image Security**: Image scanning and admission control
- **Secrets Management**: Encrypted at rest, External Secrets integration

### Platform Hardening

```yaml
# Security Context Constraints
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: restricted-custom
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowPrivilegedContainer: false
allowedCapabilities: null
defaultAddCapabilities: null
fsGroup:
  type: MustRunAs
runAsUser:
  type: MustRunAsRange
  uidRangeMin: 1000000000
  uidRangeMax: 2000000000
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
```

### Network Security

```yaml
# Default Network Policy - Deny All
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: {namespace}
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

## Performance and Sizing

### Node Sizing Guidelines

| Use Case | Control Plane | Worker Nodes | Storage |
|----------|---------------|--------------|---------|
| Development | 3x m5.xlarge | 3x m5.large | gp3 |
| Production | 3x m5.2xlarge | 6x m5.xlarge | io1/io2 |
| Compute Intensive | 3x c5.2xlarge | 6x c5.4xlarge | gp3/io1 |
| Memory Intensive | 3x r5.2xlarge | 6x r5.2xlarge | gp3 |

### Storage Classes

```yaml
# Fast SSD storage for databases
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: ebs.csi.aws.com
parameters:
  type: io1
  iopsPerGB: "50"
  encrypted: "true"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true

# General purpose storage
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: general-purpose
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  encrypted: "true"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

## Maintenance and Lifecycle

### Cluster Updates

1. **Automated Updates**: Configure ClusterVersion for automatic updates
2. **Scheduled Maintenance**: Use maintenance windows for updates
3. **Update Approval**: Manual approval for critical updates
4. **Rollback Support**: Automated rollback on update failures

### Backup Strategy

```yaml
# ETCD backup configuration
apiVersion: config.openshift.io/v1
kind: Scheduler
metadata:
  name: cluster
spec:
  mastersSchedulable: false
  policy:
    name: ""
```

### Monitoring and Alerting

- **Built-in Monitoring**: Prometheus, Grafana, AlertManager included
- **Custom Metrics**: User workload monitoring enabled
- **External Integration**: Forward metrics to external systems
- **Log Aggregation**: Cluster Logging Operator for centralized logs

### Decommissioning

1. **Workload Migration**: Move applications to other clusters
2. **Data Backup**: Ensure all data is backed up
3. **ACM Detach**: Remove from ACM management
4. **Hive Cleanup**: Delete ClusterDeployment (triggers cloud resource cleanup)
5. **DNS Cleanup**: Remove DNS records
6. **Certificate Cleanup**: Revoke certificates if needed

## Best Practices

### Resource Management

- Use ResourceQuotas and LimitRanges
- Implement proper node selectors and taints
- Configure Horizontal Pod Autoscaling
- Monitor resource utilization regularly

### Security

- Regular security scanning of container images
- Keep OpenShift version current with security patches
- Implement network segmentation with NetworkPolicies
- Use Service Mesh for advanced traffic management
- Regular RBAC audits

### Operational Excellence

- Implement comprehensive monitoring and alerting
- Use GitOps for all configuration changes
- Maintain disaster recovery procedures
- Regular cluster health checks
- Capacity planning and growth management

### Development Workflow

- Use separate clusters for dev/test/prod environments
- Implement CI/CD pipelines with Tekton
- Use OpenShift Templates and Helm for application deployment
- Implement proper secret management practices

This documentation provides complete technical coverage for OpenShift Container Platform cluster provisioning and management within the bootstrap GitOps system. For additional support, refer to the official Red Hat OpenShift documentation and Hive operator guides.