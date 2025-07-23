# HyperShift (HCP) Cluster Provisioning Guide

## Overview

This document provides comprehensive technical documentation for HyperShift Hosted Control Plane (HCP) cluster provisioning within the bootstrap GitOps system. HyperShift clusters provide a hosted control plane architecture where the Kubernetes control plane runs as pods on a management cluster, while worker nodes run separately in the target infrastructure.

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                    Management Cluster (Hub)                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   ArgoCD    │  │     ACM     │  │   Tekton    │            │
│  │   GitOps    │  │     Hub     │  │  Pipelines  │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ HyperShift  │  │   Vault     │  │ External    │            │
│  │  Operator   │  │ Secrets     │  │ Secrets     │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │            Hosted Control Plane Pods                   │  │
│  │  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐   │  │
│  │  │ kube- │ │ etcd  │ │ kube- │ │ kube- │ │  ...  │   │  │
│  │  │ api   │ │       │ │ ctrl  │ │ sched │ │       │   │  │
│  │  └───────┘ └───────┘ └───────┘ └───────┘ └───────┘   │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼ Manages Worker Nodes
┌─────────────────────────────────────────────────────────────────┐
│                   Target Infrastructure                        │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   Worker    │  │   Worker    │  │   Worker    │            │
│  │   Node 1    │  │   Node 2    │  │   Node N    │            │
│  │             │  │             │  │             │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ Node Agent  │  │   Tekton    │  │  Workload   │            │
│  │(HyperShift) │  │ Pipelines   │  │Applications │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

### Key Technologies

- **HyperShift Operator**: Hosted control plane lifecycle management
- **Cluster API (CAPI)**: Infrastructure provisioning for worker nodes
- **Red Hat ACM**: Multi-cluster management integration
- **ArgoCD**: GitOps continuous deployment
- **Tekton Pipelines**: CI/CD automation for cluster workflows
- **External Secrets Operator**: Vault integration for credential management
- **Kustomize**: Configuration management and templating

### HyperShift Benefits

- **Resource Efficiency**: Multiple control planes share management cluster resources
- **Faster Provisioning**: Control plane starts in ~5 minutes vs 45+ minutes for full clusters
- **Cost Optimization**: Reduced infrastructure costs for multiple clusters
- **Centralized Management**: All control planes managed from single location
- **Isolation**: Each hosted cluster has isolated control plane components
- **Scalability**: Support for hundreds of hosted clusters per management cluster

## Prerequisites

### Infrastructure Requirements

1. **Management Cluster Requirements**
   - OpenShift 4.14+ with sufficient resources for hosted control planes
   - HyperShift operator installed
   - Red Hat Advanced Cluster Management (ACM) operator installed
   - Cluster API controllers (multicluster-engine)
   - Minimum 32 GB RAM, 16 vCPUs per management cluster

2. **Platform Support**
   - AWS (most mature)
   - Azure (supported)
   - Agent-based (bare metal/on-premises)
   - KubeVirt (nested virtualization)
   - None platform (bring your own infrastructure)

3. **Network Requirements**
   - Connectivity between management cluster and worker node infrastructure
   - Load balancer for API server access (AWS ALB, Azure Load Balancer, etc.)
   - DNS resolution for hosted cluster API endpoints
   - Certificate management for TLS termination

### Secret Management

- **Vault Integration**: Credentials stored securely in Vault
- **Cloud Credentials**: Platform-specific credentials for worker node provisioning
- **Pull Secrets**: Red Hat registry credentials for image pulls
- **SSH Keys**: Optional access keys for worker nodes
- **TLS Certificates**: API server and ingress certificates

### AWS-Specific Requirements

| Service | Requirement | Reason |
|---------|-------------|---------|
| EC2 | Instance quotas for worker nodes | NodePool provisioning |
| ELB | Application Load Balancer quota | API server and ingress access |
| Route53 | Hosted zone management | DNS for cluster endpoints |
| IAM | Service roles and policies | Worker node permissions |
| VPC | Subnets and security groups | Network isolation |

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
  type: hcp
  region: {cloud-region}
  domain: {base-domain}
  
  compute:
    instanceType: {instance-type}     # e.g., m5.large
    replicas: {worker-count}          # e.g., 3
    
  hypershift:
    version: "{major.minor}"          # e.g., "4.14"
    releaseImage: "quay.io/openshift-release-dev/ocp-release:{version}"
    controlPlaneAvailabilityPolicy: "HighlyAvailable"  # or SingleReplica
    infrastructureAvailabilityPolicy: "HighlyAvailable"
    
  platform:
    aws:
      region: {aws-region}
      instanceType: {instance-type}
      zones: ["{zone1}", "{zone2}", "{zone3}"]
      endpointAccess: "Public"       # Public, Private, PublicAndPrivate
```

### Generated Kubernetes Resources

The `bin/generate-cluster` script creates the following resources for HCP clusters:

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
    hypershift.openshift.io/hosted-cluster: {cluster-name}
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

#### 3. HyperShift HostedCluster

```yaml
# clusters/{cluster-name}/hostedcluster.yaml
apiVersion: hypershift.openshift.io/v1beta1
kind: HostedCluster
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
  annotations:
    cluster.open-cluster-management.io/managedcluster-name: {cluster-name}
    cluster.open-cluster-management.io/hypershiftdeployment: {cluster-name}
spec:
  release:
    image: "quay.io/openshift-release-dev/ocp-release:{version}-{arch}"
  
  pullSecret:
    name: pull-secret
  
  sshKey:
    name: {cluster-name}-ssh-key  # Optional
  
  networking:
    clusterNetwork:
    - cidr: 10.132.0.0/14
    serviceNetwork:
    - cidr: 172.31.0.0/16
    networkType: OVNKubernetes
    machineNetwork:
    - cidr: 10.0.0.0/16  # Platform network CIDR
  
  controllerAvailabilityPolicy: HighlyAvailable  # or SingleReplica
  infrastructureAvailabilityPolicy: HighlyAvailable
  
  platform:
    type: AWS
    aws:
      region: {aws-region}
      cloudProviderConfig:
        vpc: {vpc-id}                    # Optional: existing VPC
        subnet:
          id: {subnet-id}                # Optional: existing subnet
        zone: {availability-zone}
      endpointAccess: Public             # Public, Private, PublicAndPrivate
      resourceTags:
      - key: "kubernetes.io/cluster/{cluster-name}"
        value: "owned"
      - key: "Environment"
        value: "{environment}"
  
  services:
  - service: APIServer
    servicePublishingStrategy:
      type: LoadBalancer                 # LoadBalancer, Route, NodePort
      loadBalancer:
        hostname: api.{cluster-name}.{base-domain}
  - service: OAuthServer
    servicePublishingStrategy:
      type: Route
  - service: Konnectivity
    servicePublishingStrategy:
      type: Route
  - service: Ignition
    servicePublishingStrategy:
      type: Route
  
  dns:
    baseDomain: {base-domain}
    publicZoneID: {route53-zone-id}      # Optional: existing Route53 zone
    privateZoneID: {private-zone-id}     # Optional: existing private zone
  
  etcd:
    managementType: Managed              # Managed or Unmanaged
    managed:
      storage:
        persistentVolume:
          size: 8Gi
          storageClassName: gp3-csi
        type: PersistentVolume
```

#### 4. HyperShift NodePool

```yaml
# clusters/{cluster-name}/nodepool.yaml
apiVersion: hypershift.openshift.io/v1beta1
kind: NodePool
metadata:
  name: {cluster-name}
  namespace: {cluster-name}
spec:
  clusterName: {cluster-name}
  replicas: {worker-count}
  
  config:
  - name: {cluster-name}-config
  
  management:
    autoRepair: true
    upgradeType: Replace                 # Replace or InPlace
    replace:
      strategy: RollingUpdate
      rollingUpdate:
        maxUnavailable: 1
        maxSurge: 1
  
  platform:
    type: AWS
    aws:
      instanceType: {instance-type}
      instanceProfile: {cluster-name}-worker  # Created by HyperShift
      subnet:
        id: {subnet-id}                  # Optional: specific subnet
      securityGroups:
      - id: {security-group-id}          # Optional: additional security groups
      rootVolume:
        type: gp3
        size: 120
        iops: 3000
        encrypted: true
      userData: |                        # Optional: additional user data
        #!/bin/bash
        echo "Additional configuration" >> /var/log/user-data.log
      tags:
        Environment: "{environment}"
        Team: "{team}"
  
  release:
    image: "quay.io/openshift-release-dev/ocp-release:{version}-{arch}"
  
  nodeClassRef:                          # Optional: custom MachineConfig
    name: {cluster-name}-worker-config
```

#### 5. NodePool Configuration (Optional)

```yaml
# clusters/{cluster-name}/nodepool-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {cluster-name}-config
  namespace: {cluster-name}
data:
  config.yaml: |
    apiVersion: machineconfiguration.openshift.io/v1
    kind: MachineConfig
    metadata:
      labels:
        machineconfiguration.openshift.io/role: worker
      name: 99-worker-custom
    spec:
      config:
        ignition:
          version: 3.2.0
        storage:
          files:
          - path: /etc/custom-config
            mode: 0644
            contents:
              source: data:text/plain;base64,Y3VzdG9tIGNvbmZpZ3VyYXRpb24=
        systemd:
          units:
          - name: custom-service.service
            enabled: true
            contents: |
              [Unit]
              Description=Custom Service
              After=network.target
              
              [Service]
              Type=oneshot
              ExecStart=/bin/echo "Custom configuration applied"
              
              [Install]
              WantedBy=multi-user.target
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
    region: {aws-region}
    vendor: HyperShift
    cluster.open-cluster-management.io/clusterset: default
    hypershift.openshift.io/hosted-cluster: {cluster-name}
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
    vendor: HyperShift
    hypershift: "true"
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
    enabled: false  # Often disabled for hosted clusters
```

## Provisioning Flow

### Phase 1: Resource Deployment (ArgoCD Sync Wave 1)

1. **Namespace Creation**: Creates dedicated namespace for cluster resources
2. **External Secrets Sync**: Vault credentials synchronized to cluster namespace
3. **HostedCluster Creation**: HyperShift operator begins control plane provisioning
4. **NodePool Creation**: Worker node specification defined
5. **ACM Resource Creation**: ManagedCluster and addon configs created

### Phase 2: Control Plane Provisioning (HyperShift Operator)

1. **Control Plane Pods**: Kubernetes control plane components deployed as pods
2. **etcd Deployment**: Persistent etcd storage configured
3. **API Server Setup**: Load balancer and TLS certificates configured
4. **Networking Setup**: CNI and service networking configured
5. **Control Plane Ready**: API server accessible and functional

**Typical Timeline**: 5-10 minutes for control plane provisioning

### Phase 3: Worker Node Provisioning (CAPI)

1. **Infrastructure Creation**: Cloud resources for worker nodes (VPC, subnets, etc.)
2. **Machine Deployment**: Worker node instances launched
3. **Node Bootstrap**: Worker nodes configured and joined to cluster
4. **Node Readiness**: All worker nodes report Ready status

**Typical Timeline**: 10-15 minutes for worker node provisioning

### Phase 4: ACM Integration (Automatic)

1. **Automatic Import**: HostedCluster automatically registered with ACM
2. **Klusterlet Deployment**: ACM agent deployed to hosted cluster
3. **Addon Configuration**: ACM addons configured per KlusterletAddonConfig
4. **Management Ready**: Cluster available for GitOps workload deployment

**Typical Timeline**: 2-5 minutes for ACM integration

### Phase 5: Workload Deployment (ArgoCD Sync Waves 2+)

1. **Operator Installation**: Additional operators via GitOps
2. **Application Deployment**: Business applications via GitOps

## HyperShift States and Monitoring

### HostedCluster Status Conditions

| Condition | Status | Description |
|-----------|---------|-------------|
| `Available` | `True` | Control plane is running and accessible |
| `Progressing` | `False` | No ongoing provisioning operations |
| `Degraded` | `False` | All control plane components healthy |
| `ReconciliationActive` | `True` | HyperShift operator actively managing cluster |
| `ValidConfiguration` | `True` | Cluster configuration is valid |
| `SupportedHostedCluster` | `True` | Configuration is supported by HyperShift |

### NodePool Status Conditions

| Condition | Status | Description |
|-----------|---------|-------------|
| `Ready` | `True` | All replicas are ready |
| `Progressing` | `False` | No ongoing node operations |
| `AutoscalerEnabled` | `True/False` | Cluster autoscaler status |
| `UpdatingVersion` | `False` | No version update in progress |
| `UpdatingConfig` | `False` | No configuration update in progress |

### Key Monitoring Commands

```bash
# Check HostedCluster status
oc get hostedcluster {cluster-name} -n {cluster-name}

# Check NodePool status
oc get nodepool {cluster-name} -n {cluster-name}

# Get hosted cluster kubeconfig
oc get secret {cluster-name}-admin-kubeconfig -n {cluster-name} -o jsonpath='{.data.kubeconfig}' | base64 -d

# Check control plane pods
oc get pods -n {cluster-name} | grep {cluster-name}

# Check worker nodes (from hosted cluster)
oc --kubeconfig={cluster-kubeconfig} get nodes

# Check ManagedCluster status
oc get managedcluster {cluster-name}
```

### Control Plane Pod Examples

```bash
# Example control plane pods for cluster 'hcp-prod-01'
NAME                                          READY   STATUS    RESTARTS
hcp-prod-01-kube-apiserver-6b8f7d8c4d-xyz12   2/2     Running   0
hcp-prod-01-etcd-0                            2/2     Running   0
hcp-prod-01-etcd-1                            2/2     Running   0
hcp-prod-01-etcd-2                            2/2     Running   0
hcp-prod-01-kube-controller-manager-abc34     1/1     Running   0
hcp-prod-01-kube-scheduler-def56              1/1     Running   0
hcp-prod-01-cluster-version-operator-ghi78    1/1     Running   0
```

## Platform-Specific Configurations

### AWS Platform Configuration

```yaml
platform:
  type: AWS
  aws:
    region: us-east-1
    cloudProviderConfig:
      vpc: vpc-12345678
      subnet:
        id: subnet-87654321
    endpointAccess: PublicAndPrivate
    resourceTags:
    - key: "Environment"
      value: "production"
    - key: "Team"
      value: "platform"
    multiArch: false
    rolesRef:
      controlPlaneOperatorARN: ""     # Optional: custom IAM role
      imageRegistryARN: ""            # Optional: custom registry role
      ingressARN: ""                  # Optional: custom ingress role
      kubeCloudControllerARN: ""      # Optional: custom cloud controller role
      networkARN: ""                  # Optional: custom network role
      nodePoolManagementARN: ""       # Optional: custom node pool role
      storageARN: ""                  # Optional: custom storage role
```

### Azure Platform Configuration

```yaml
platform:
  type: Azure
  azure:
    location: eastus
    resourceGroupName: {resource-group}
    vnetName: {vnet-name}
    vnetID: {vnet-id}
    subnetName: {subnet-name}
    subscriptionID: {subscription-id}
    machineIdentityID: {identity-id}
    securityGroupName: {security-group}
    credentials:
      name: azure-credentials
```

### Agent Platform (Bare Metal/On-Premises)

```yaml
platform:
  type: Agent
  agent:
    agentNamespace: {agent-namespace}
```

### KubeVirt Platform (Nested Virtualization)

```yaml
platform:
  type: KubeVirt
  kubevirt:
    baseDomainPassthrough: true
    generateID: {cluster-name}
    credentials:
      infraKubeConfigSecret:
        name: kubevirt-credentials
        key: kubeconfig
```

## Networking Configurations

### Service Publishing Strategies

```yaml
services:
- service: APIServer
  servicePublishingStrategy:
    type: LoadBalancer
    loadBalancer:
      hostname: api.{cluster-name}.{domain}
      
- service: OAuthServer
  servicePublishingStrategy:
    type: Route
    route:
      hostname: oauth.{cluster-name}.{domain}
      
- service: Konnectivity
  servicePublishingStrategy:
    type: Route
    route:
      hostname: konnectivity.{cluster-name}.{domain}
      
- service: Ignition
  servicePublishingStrategy:
    type: Route
    route:
      hostname: ignition.{cluster-name}.{domain}
```

### Advanced Networking

```yaml
networking:
  clusterNetwork:
  - cidr: 10.132.0.0/14
    hostPrefix: 23
  serviceNetwork:
  - cidr: 172.31.0.0/16
  machineNetwork:
  - cidr: 10.0.0.0/16
  networkType: OVNKubernetes
  
  # Advanced OVN configuration
  ovnKubernetesConfig:
    mtu: 1450
    genevePort: 6081
    hybridOverlayConfig:
      hybridClusterNetwork:
      - cidr: 10.132.0.0/14
        hostPrefix: 23
      hybridOverlayVXLANPort: 9898
```

## Common Issues and Solutions

### 1. Control Plane Pods CrashLoopBackOff

**Error**: Control plane pods failing to start or crashing repeatedly
**Causes**:
- Insufficient resources on management cluster
- Invalid configuration in HostedCluster
- Network connectivity issues
**Solution**: Check resource limits, validate configuration, verify network connectivity

### 2. NodePool Stuck in Progressing

**Error**: NodePool remains in "Progressing" state, nodes not joining
**Causes**:
- AWS quota limits (instance types, Elastic IPs)
- Invalid instance type or AMI
- Network security group issues
- IAM permission problems
**Solution**: Check AWS quotas, verify instance type availability, review security groups and IAM

### 3. API Server Unreachable

**Error**: Cannot connect to hosted cluster API server
**Causes**:
- Load balancer misconfiguration
- DNS resolution issues
- Certificate problems
- Service publishing strategy issues
**Solution**: Verify load balancer status, check DNS records, validate certificates

### 4. NodePool Upgrade Failures

**Error**: NodePool upgrade gets stuck or fails
**Causes**:
- Incompatible OpenShift versions
- Resource constraints during rolling update
- Custom MachineConfig conflicts
**Solution**: Check version compatibility, ensure adequate resources, review MachineConfig

### 5. etcd Storage Issues

**Error**: etcd pods failing due to storage problems
**Causes**:
- Insufficient storage space
- Storage class issues
- Persistent volume problems
**Solution**: Check storage capacity, verify storage class, review PV status

### 6. Worker Node Registration Issues

**Error**: Worker nodes fail to register with hosted cluster
**Causes**:
- Network connectivity between management and worker infrastructure
- Certificate authority issues
- Ignition service problems
**Solution**: Verify network connectivity, check CA certificates, review ignition service logs

## Security Considerations

### Control Plane Isolation

- **Namespace Isolation**: Each hosted cluster runs in dedicated namespace
- **Network Policies**: Control plane components isolated by network policies
- **RBAC**: Strict RBAC controls access to hosted cluster resources
- **Secret Management**: Pull secrets and credentials properly isolated

### Worker Node Security

```yaml
# Security Context for worker nodes
spec:
  config:
  - name: security-config
data:
  config.yaml: |
    apiVersion: machineconfiguration.openshift.io/v1
    kind: MachineConfig
    metadata:
      labels:
        machineconfiguration.openshift.io/role: worker
      name: 99-worker-security
    spec:
      config:
        ignition:
          version: 3.2.0
        systemd:
          units:
          - name: kubelet.service
            dropins:
            - name: 99-kubelet-security.conf
              contents: |
                [Service]
                Environment="KUBELET_EXTRA_ARGS=--protect-kernel-defaults=true"
```

### Network Security

```yaml
# Network policies for hosted cluster
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: control-plane-isolation
  namespace: {cluster-name}
spec:
  podSelector:
    matchLabels:
      app: {cluster-name}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: hypershift
    ports:
    - protocol: TCP
      port: 6443
```

## Performance and Sizing

### Management Cluster Sizing

| Hosted Clusters | vCPU | Memory | Storage | Notes |
|-----------------|------|--------|---------|-------|
| 1-10 | 16 | 64 GB | 500 GB | Small scale |
| 10-50 | 32 | 128 GB | 1 TB | Medium scale |
| 50-100 | 64 | 256 GB | 2 TB | Large scale |
| 100+ | 128+ | 512 GB+ | 4 TB+ | Enterprise scale |

### Control Plane Resource Requests

```yaml
# Example resource requirements per hosted cluster
resources:
  controlPlane:
    cpu: "2"
    memory: "8Gi"
  etcd:
    cpu: "1"
    memory: "4Gi"
    storage: "8Gi"
```

### Worker Node Sizing

| Use Case | Instance Type | Node Count | Notes |
|----------|---------------|------------|-------|
| Development | t3.medium | 2-3 | Cost-effective |
| Production | m5.large+ | 3-6 | Balanced workloads |
| Compute Intensive | c5.xlarge+ | 3-9 | CPU-bound applications |
| Memory Intensive | r5.large+ | 3-6 | Memory-bound applications |

## Maintenance and Lifecycle

### Cluster Updates

```yaml
# Updating HostedCluster version
spec:
  release:
    image: "quay.io/openshift-release-dev/ocp-release:4.14.15-x86_64"
```

```yaml
# Updating NodePool configuration
spec:
  release:
    image: "quay.io/openshift-release-dev/ocp-release:4.14.15-x86_64"
  management:
    upgradeType: Replace
    replace:
      strategy: RollingUpdate
      rollingUpdate:
        maxUnavailable: 1
        maxSurge: 1
```

### Scaling Operations

```bash
# Scale NodePool replicas
oc patch nodepool {cluster-name} -n {cluster-name} --type='merge' -p='{"spec":{"replicas":5}}'

# Add additional NodePool
oc apply -f - <<EOF
apiVersion: hypershift.openshift.io/v1beta1
kind: NodePool
metadata:
  name: {cluster-name}-compute
  namespace: {cluster-name}
spec:
  clusterName: {cluster-name}
  replicas: 3
  platform:
    type: AWS
    aws:
      instanceType: c5.2xlarge
EOF
```

### Backup and Disaster Recovery

```yaml
# etcd backup configuration
spec:
  etcd:
    managed:
      storage:
        persistentVolume:
          size: 8Gi
          storageClassName: gp3-csi
        backupPolicy:
          schedule: "0 2 * * *"  # Daily at 2 AM
          retention: "7d"
```

### Decommissioning

1. **Workload Migration**: Move applications to other clusters
2. **Data Backup**: Ensure all data is backed up
3. **ACM Detach**: Remove from ACM management if needed
4. **NodePool Deletion**: Delete NodePools (worker nodes cleaned up)
5. **HostedCluster Deletion**: Delete HostedCluster (control plane cleaned up)
6. **Resource Cleanup**: Verify cloud resources are cleaned up
7. **DNS Cleanup**: Remove DNS records

## Best Practices

### Resource Management

- **Right-sizing**: Use appropriate instance types for workloads
- **Resource Limits**: Set resource limits on hosted cluster components
- **Node Pools**: Use multiple NodePools for different workload types
- **Monitoring**: Monitor resource utilization across all hosted clusters

### Security

- **Network Segmentation**: Implement network policies for isolation
- **Least Privilege**: Use minimal required permissions
- **Secret Rotation**: Regular rotation of credentials and certificates
- **Security Scanning**: Regular scanning of container images

### Operational Excellence

- **GitOps**: Manage all cluster configuration through Git
- **Monitoring**: Implement comprehensive monitoring for hosted clusters
- **Automation**: Automate cluster lifecycle operations
- **Documentation**: Maintain updated runbooks and procedures

### Cost Optimization

- **Cluster Consolidation**: Use hosted clusters to reduce infrastructure costs
- **Right-sizing**: Monitor and adjust worker node sizes
- **Lifecycle Management**: Automatically scale or shutdown dev/test clusters
- **Resource Sharing**: Share management cluster resources across teams

This documentation provides complete technical coverage for HyperShift Hosted Control Plane cluster provisioning and management within the bootstrap GitOps system. For additional support, refer to the official HyperShift documentation and Red Hat Advanced Cluster Management guides.