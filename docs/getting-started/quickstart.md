# 5-Minute Quickstart Guide

**Audience**: New users  
**Complexity**: Beginner  
**Estimated Time**: 5 minutes to understand  
**Prerequisites**: Basic Kubernetes knowledge

## What You'll Learn

In 5 minutes, you'll understand how to deploy and manage multi-cluster OpenShift environments using this GitOps bootstrap repository.

## The Big Picture

```
Hub Cluster (OpenShift + ArgoCD + ACM) 
    ↓ GitOps automation
Regional Clusters (OCP or EKS)
    ↓ Applications deployed
Running services across regions
```

## 3-Step Workflow

### 1. 🚀 Create New Cluster
```bash
./bin/new-cluster
# Interactive prompts guide you through configuration
```

### 2. 🔄 Deploy via GitOps  
```bash
./bootstrap.sh
# ArgoCD automatically provisions cluster and deploys applications
```

### 3. 📊 Monitor Status
```bash
./bin/health-check
# View real-time status of all clusters and applications
```

## What Gets Created

**Automatically generated for each cluster:**
- ✅ **Cluster provisioning** (OCP via Hive, EKS via CAPI)
- ✅ **Operator installations** (OpenShift Pipelines)
- ✅ **Pipeline deployments** (Hello World, Cloud Infrastructure)
- ✅ **Service deployments** (OCM database services)
- ✅ **GitOps integration** (ArgoCD ApplicationSets with sync waves)

## Key Features

- **Multi-Provider**: OpenShift (OCP) and EKS clusters
- **GitOps Automation**: ArgoCD manages everything
- **Regional Deployment**: Services deployed per cluster
- **Centralized Management**: Single hub cluster controls all
- **Automated Validation**: Built-in testing and health checks

## Directory Structure (Simplified)

```bash
├── regions/           # Input: Regional specifications
├── clusters/          # Generated: Cluster provisioning configs  
├── operators/         # Generated: Operator deployments
├── pipelines/         # Generated: Pipeline configurations
├── deployments/       # Generated: Service deployments
├── gitops-applications/ # Generated: ArgoCD ApplicationSets
└── bin/               # Tools: new-cluster, health-check, etc.
```

## Next Steps

Ready to dive deeper?

1. **Full Setup**: [Complete Installation Guide](../../INSTALL.md)
2. **Understand Architecture**: [Architecture Overview](../../ARCHITECTURE.md)  
3. **Deploy Your First Cluster**: [Cluster Creation Guide](../../guides/cluster-creation.md)
4. **Monitor & Troubleshoot**: [Monitoring Guide](../../guides/monitoring.md)

## Common Use Cases

- **Development**: Multi-environment testing across regions
- **Production**: Regional deployments with centralized management
- **Hybrid Cloud**: Mix of OpenShift and EKS clusters
- **CI/CD**: Automated pipeline deployment across clusters

## Support

- **Documentation**: [Complete Index](../INDEX.md)
- **Troubleshooting**: [Monitoring Guide](../../guides/monitoring.md)
- **AWS Cleanup**: [Resource Cleanup](../../bin/clean-aws.md)