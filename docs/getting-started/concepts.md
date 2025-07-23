# Core Concepts

**Audience**: New users  
**Complexity**: Beginner  
**Estimated Time**: 10 minutes  
**Prerequisites**: Basic Kubernetes understanding

## Hub-Spoke Architecture

### Hub Cluster
**Central control plane** that manages all regional clusters:
- **OpenShift**: Main platform running GitOps automation
- **ArgoCD**: GitOps engine for continuous deployment
- **ACM**: Multi-cluster management and governance
- **CAPI**: Cluster API for automated provisioning

### Spoke Clusters (Managed Clusters)
**Regional deployments** provisioned and managed by the hub:
- **OpenShift (OCP)**: Full-featured OpenShift clusters via Hive
- **EKS**: Cost-effective Kubernetes clusters via CAPI
- **Applications**: Services deployed via GitOps from hub

## GitOps Workflow

### Sync Waves (Deployment Ordering)
ArgoCD deploys resources in ordered waves to ensure dependencies:

1. **Wave 1**: Cluster provisioning (infrastructure)
2. **Wave 2**: Operator installation (capabilities) 
3. **Wave 3**: Pipeline deployment (automation)
4. **Wave 4**: Service deployment (applications)

### Resource Flow
```
Regional Specification → Generation Tool → Git Repository → ArgoCD → Target Cluster
```

## Cluster Types

### OpenShift (OCP) Clusters
- **Provisioner**: Hive operator
- **Resources**: ClusterDeployment + MachinePool + InstallConfig
- **Features**: Full OpenShift capabilities, advanced operators
- **Use Case**: Production workloads, complex applications

### EKS Clusters  
- **Provisioner**: Cluster API (CAPI) with AWS provider
- **Resources**: AWSManagedControlPlane + AWSManagedMachinePool
- **Features**: Kubernetes-native, cost-effective, AWS-managed
- **Use Case**: Development, cost-sensitive workloads

## Configuration Management

### Kustomize-Based
- **Base Templates**: Shared configurations in `bases/`
- **Overlays**: Environment-specific customizations in cluster directories
- **Patches**: Modifications applied via Kustomize patches
- **Generators**: ConfigMap and Secret generators for dynamic content

### Regional Specifications
Simple YAML files that drive complete cluster generation:
```yaml
name: my-cluster
type: eks  # or "ocp"
region: us-west-2
domain: rosa.mturansk-test.csu2.i3.devshift.org
instanceType: m5.large
replicas: 3
```

## Application Deployment

### ApplicationSets
Single ArgoCD ApplicationSet per cluster generates multiple applications:
- **cluster-XX-cluster**: Cluster provisioning
- **cluster-XX-operators**: Operator installations
- **cluster-XX-pipelines-***: Pipeline deployments  
- **cluster-XX-deployments-ocm**: Service deployments

### Service Structure
Each cluster gets standardized service deployments:
- **OCM Services**: AMS-DB, CS-DB, OSL-DB
- **Pipeline Templates**: Hello World, Cloud Infrastructure
- **Operators**: OpenShift Pipelines for CI/CD automation

## Secret Management

### Traditional (Current)
Manual secret files applied during bootstrap:
- `secrets/aws-credentials.yaml` 
- `secrets/pull-secret.yaml`

### Vault Integration (Available)
Automated secret management with External Secrets Operator:
- **Centralized Storage**: HashiCorp Vault KV store
- **Automated Sync**: ESO pulls secrets to clusters
- **Access Control**: Vault policies control secret access
- **Audit Trail**: Complete secret access logging

## Multi-Cluster Management

### ACM Integration
- **ManagedCluster**: Represents each cluster in ACM
- **Policies**: Governance and compliance across clusters
- **Observability**: Centralized monitoring and metrics
- **GitOps Integration**: Automatic ArgoCD cluster registration

### Infrastructure Providers
ACM automatically manages CAPI providers for different cloud platforms:
- **AWS**: EKS clusters with native AWS integration
- **Azure**: AKS clusters (planned)
- **GCP**: GKE clusters (planned)
- **vSphere**: On-premises clusters (planned)

## Automation Tools

### Generation Tools
- **`bin/cluster-create`**: Interactive cluster specification creation
- **`bin/cluster-generate`**: Complete overlay generation from specifications
- **`bin/regenerate-all-clusters`**: Bulk regeneration for template updates

### Monitoring Tools
- **`bin/monitor-health`**: Comprehensive environment status
- **`status.sh`**: CRD establishment monitoring
- **`wait.kube.sh`**: Resource condition waiting

### Cleanup Tools
- **`bin/clean-aws`**: Automated AWS resource cleanup
- **Manual procedures**: Documented rollback processes

## Key Benefits

### For Operators
- **Simplified Deployment**: Single command cluster creation
- **Standardized Configuration**: Consistent patterns across clusters
- **Automated Validation**: Built-in testing and verification
- **Centralized Monitoring**: Hub-based status visibility

### For Developers
- **Multi-Environment**: Easy testing across regions and cluster types
- **GitOps Native**: Infrastructure as code with Git workflows
- **Extensible**: Easy addition of new services and pipelines
- **Consistent APIs**: Same tools work across OCP and EKS

### For Administrators
- **Governance**: ACM policies across all clusters
- **Cost Management**: Mix of OCP and EKS for optimal pricing
- **Security**: Centralized secret management and access control
- **Compliance**: Audit trails and automated policy enforcement

## Next Steps

- **Deep Dive**: [Architecture Overview](../architecture/gitops-flow.md)
- **Hands-On**: [First Cluster Deployment](./first-cluster.md)
- **Operations**: [Cluster Management](../operations/cluster-management.md)