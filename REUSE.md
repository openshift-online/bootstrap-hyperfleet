# How to Reuse This Repository

This repository is designed for **reuse**. You clone it, bootstrap your hub cluster, and it automatically sets up internal infrastructure for managing additional clusters.

## Two-Phase Reuse Pattern

### Phase 1: Bootstrap from GitHub
```bash
# 1. Clone this repository
git clone https://github.com/openshift-online/bootstrap.git
cd bootstrap

# 2. Log into your OpenShift cluster
oc login https://api.your-hub-cluster.example.com:6443

# 3. Bootstrap the hub cluster
oc apply -k clusters/global/
```

**What this does:**
- Installs OpenShift GitOps (ArgoCD)
- Deploys Advanced Cluster Management (ACM) 
- Sets up Vault for secret management
- Installs internal Gitea service
- Creates pipeline infrastructure

### Phase 2: Self-Referential Management
After bootstrap completes, the cluster becomes **self-managing**:

- **External GitHub**: Used only for initial bootstrap deployment
- **Internal Gitea**: Used for ongoing cluster-specific configuration
- **Self-Referential**: New clusters reference their own internal Git service

The cluster automatically:
1. Clones this repository to internal Gitea
2. Switches ArgoCD to use internal Gitea for cluster-specific configs
3. Provisions new clusters using internal Git as source

## Directory Structure for Reuse

```
bootstrap/
├── clusters/
│   ├── global/                  # Hub cluster resources
│   └── {cluster-name}/          # Managed cluster resources
├── bases/                       # Reusable templates
└── bin/                         # Management tools
```

## Adding Your First Cluster

Once bootstrap is complete:

```bash
# 1. Create cluster specification
./bin/cluster-create

# Follow prompts to specify:
# - Cluster name: my-first-cluster
# - Type: ocp (OpenShift) or eks (EKS)
# - Region: us-east-1
# - Instance type: m5.2xlarge

# 2. Generate cluster configuration
# (automatically called by cluster-create)

# 3. Commit and push changes
git add .
git commit -m "Add my-first-cluster"
git push origin main
```

The system automatically:
- ✅ Creates cluster provisioning resources
- ✅ Generates pipeline deployments  
- ✅ Sets up operator installations
- ✅ Configures service deployments
- ✅ Creates ArgoCD applications with proper ordering

## How Self-Reference Works

**Initial Bootstrap** (GitHub):
```yaml
# gitops-applications use GitHub
source:
  repoURL: 'https://github.com/openshift-online/bootstrap'
```

**Ongoing Management** (Internal Gitea):
```yaml
# Cluster-specific configs use internal Gitea
source:
  repoURL: 'http://gitea.gitea-system.svc.cluster.local:3000/myadmin/bootstrap.git'
```

This allows:
- **Reuse**: Multiple teams can use the same base GitHub repo
- **Isolation**: Each cluster has its own internal Git with specific configs
- **Independence**: Clusters manage themselves without external dependencies

## Monitoring Your Deployment

```bash
# Check ArgoCD applications
oc get applications -n openshift-gitops

# Monitor cluster provisioning
oc get clusterdeployments -A     # OpenShift clusters
oc get clusters -A               # EKS clusters

# Check hub cluster health
./bin/monitor-health
```

## Access Management Interfaces

```bash
# ArgoCD console
echo "ArgoCD: https://$(oc get route openshift-gitops-server -n openshift-gitops -o jsonpath='{.spec.host}')"

# ACM console  
echo "ACM: https://$(oc get route multicloud-console -n open-cluster-management -o jsonpath='{.spec.host}')"

# Gitea console
echo "Gitea: https://$(oc get route gitea -n gitea-system -o jsonpath='{.spec.host}')"
```

## Customization

The repository supports customization through:

- **Cluster specifications**: Define cluster requirements in `clusters/{cluster-name}/{cluster-name}.yaml`
- **Base templates**: Modify shared components in `bases/`
- **Hub operators**: Customize deployments in `clusters/global/operators/`
- **Cluster resources**: Customize per-cluster resources in `clusters/{cluster-name}/`

## Support

- **Quick Start**: [docs/getting-started/QUICKSTART.md](./docs/getting-started/QUICKSTART.md)
- **Architecture**: [docs/architecture/ARCHITECTURE.md](./docs/architecture/ARCHITECTURE.md)
- **Bootstrap Details**: [BOOTSTRAP.md](./BOOTSTRAP.md)
- **Navigation Guide**: [NAVIGATION.md](./NAVIGATION.md)