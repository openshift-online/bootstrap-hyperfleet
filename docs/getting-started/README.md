# Getting Started

**Progressive learning path for new users**

## 🚀 Learning Journey

### 1. **Quick Overview** (5 minutes)
**[Quickstart Guide](./quickstart.md)** - Understand the big picture and 3-step workflow

### 2. **Core Understanding** (10 minutes)  
**[Core Concepts](./concepts.md)** - Learn hub-spoke architecture, GitOps workflow, and cluster types

### 3. **Basic Setup** (1-2 hours)
**[Installation Guide](./installation.md)** - Set up hub cluster and deploy your first regional cluster

### 4. **Hands-On Experience** (45 minutes)
**[Deploy Your First Cluster](./first-cluster.md)** - Detailed walkthrough from specification to running services

### 5. **Production Setup** (3-4 hours)
**[Production Installation](./production-installation.md)** - Enterprise-grade deployment and management

## 📍 Choose Your Path

### New to GitOps?
Start with → [Core Concepts](./concepts.md) → [Quickstart](./quickstart.md) → [Installation Guide](./installation.md)

### Want Quick Results?
Start with → [Quickstart](./quickstart.md) → [Installation Guide](./installation.md) → [First Cluster](./first-cluster.md)

### Need Deep Understanding?
Start with → [Core Concepts](./concepts.md) → [Architecture Overview](../architecture/ARCHITECTURE.md) → [Installation Guide](./installation.md)

### Setting Up Production?
Start with → [Installation Guide](./installation.md) → [Production Installation](./production-installation.md) → [Operations](../operations/cluster-management.md)

## 🎯 What You'll Learn

After completing this section, you'll be able to:
- ✅ Understand the hub-spoke architecture
- ✅ Deploy clusters using the automated tools
- ✅ Monitor cluster health and troubleshoot issues
- ✅ Customize deployments for your environment

## 🔗 Next Steps

Once you've completed the getting started section:

- **Operations**: [Day-to-day cluster management](../operations/cluster-management.md)
- **Deep Dive**: [Technical architecture details](../architecture/gitops-flow.md)  
- **Reference**: [Command cheat sheet](../reference/commands.md)
- **Troubleshooting**: [Common issues and solutions](../operations/troubleshooting.md)

## 🧭 Understanding Repository Navigation

This repository uses **semantic directory organization** - each directory level tells you exactly what you're looking at:

### **Quick Navigation Guide**
```bash
# See all available operators/applications
ls operators/                    # → advanced-cluster-management, gitops-integration, openshift-pipelines, vault

# See where a specific operator is deployed  
ls operators/vault/              # → global/ (hub cluster deployment)
ls operators/openshift-pipelines/ # → global/, ocp-02/, ocp-03/, eks-01/ (all deployment targets)

# See all pipeline types
ls pipelines/                    # → cloud-infrastructure-provisioning, hello-world

# See which clusters run a specific pipeline
ls pipelines/hello-world/        # → ocp-02/, ocp-03/, eks-01/ (all target clusters)

# See all available clusters
ls clusters/                     # → ocp-02/, ocp-03/, eks-01/ (provisioned clusters)

# See services deployed to clusters
ls deployments/                  # → ocm/ (service type)
ls deployments/ocm/              # → ocp-02/, ocp-03/, eks-01/ (deployment targets)
```

### **Pattern Recognition**
- **operators/{name}/{target}/** - Application deployments
- **pipelines/{type}/{cluster}/** - Pipeline runs  
- **deployments/{service}/{cluster}/** - Service deployments
- **global/** - Always means "hub cluster"
- **{cluster-name}/** - Always means "managed cluster"

This pattern makes it easy to find exactly what you're looking for without memorizing complex paths.

## 🛠️ Advanced Tools

### Internal Git Server (Gitea)
For advanced workflows and eliminating external dependencies:

**[Gitea Installation and Usage Guide](../../operators/gitea/global/GITEA.md)** - Deploy internal git server for immediate ArgoCD sync without GitHub dependency

**Use Case**: When you need to provision clusters without waiting for manual git pushes to GitHub. The enhanced `bin/cluster-generate --push-to-gitea` workflow automatically pushes configurations to internal Gitea, enabling immediate ArgoCD synchronization.