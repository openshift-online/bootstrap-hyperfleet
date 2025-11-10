# OpenShift Bootstrap Architecture

## Bootstrap Architecture

### Phase 1: Initial Bootstrap
```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                          EXTERNAL BOOTSTRAP (GitHub)                               │
│                            oc apply -k clusters/global/gitops/                     │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                  │
│  │   OpenShift     │    │      ACM        │    │     Vault       │                  │
│  │   GitOps        │    │ApplicationSet   │    │   + ESO         │                  │
│  │                 │    │                 │    │                 │                  │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │                  │
│  │ │Self-Managing│ │    │ │Wave 2: Oper │ │    │ │Secret Mgmt  │ │                  │
│  │ │  ArgoCD     │ │    │ │Wave 3: Hub  │ │    │ │Integration  │ │                  │
│  │ │             │ │    │ │Wave 4: Pol. │ │    │ │             │ │                  │
│  │ │Applications │ │    │ │             │ │    │ │External Sec │ │                  │
│  │ │ApplicationS │ │    │ │Multi-Cluster│ │    │ │ Operator    │ │                  │
│  │ │    ets      │ │    │ │  Management │ │    │ │             │ │                  │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │                  │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘                  │
│                                                                                     │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                  │
│  │    Tekton       │    │ Hub Provisioner │    │  Cluster        │                  │
│  │   Pipelines     │    │   Pipelines     │    │   Management    │                  │
│  │                 │    │                 │    │                 │                  │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │                  │
│  │ │Hub Cluster  │ │    │ │Cluster      │ │    │ │ApplicationS │ │                  │
│  │ │ Pipelines   │ │    │ │Provisioning │ │    │ │    ets      │ │                  │
│  │ │             │ │    │ │             │ │    │ │             │ │                  │
│  │ │Global Oper. │ │    │ │Automated    │ │    │ │Multi-Cluster│ │                  │
│  │ │Installation │ │    │ │Workflows    │ │    │ │  Lifecycle  │ │                  │
│  │ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │                  │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                          │
                                          │ GitOps Reconciliation
                                          │
### Phase 2: Cluster Lifecycle Management
                                          ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                       ARGOCD CLUSTER LIFECYCLE MANAGEMENT                          │
│                     Declarative GitOps-based Provisioning                          │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                    Cluster Provisioning & Management                       │    │
│  │                                                                             │    │
│  │  OCP Clusters (Hive)           EKS Clusters (CAPI)                          │    │
│  │  ┌─────────────────┐            ┌─────────────────┐                         │    │
│  │  │ClusterDeployment│            │AWSManagedControl│                         │    │
│  │  │MachinePool      │            │Plane            │                         │    │
│  │  │InstallConfig    │            │AWSManagedMachine│                         │    │
│  │  │ManagedCluster   │            │Pool             │                         │    │
│  │  └─────────────────┘            │ManagedCluster   │                         │    │
│  │                                 └─────────────────┘                         │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                          │
                                          │ Regional Cluster Deployment
                                          │
      ┌───────────────────────────────────┼───────────────────────────────────┐
      │                                   │                                   │
      ▼                                   ▼                                   ▼
┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
│   us-east-1     │              │   us-west-2     │              │ ap-southeast-1  │
│                 │              │                 │              │                 │
│  ┌───────────┐  │              │  ┌───────────┐  │              │  ┌───────────┐  │
│  │my-cluster │  │              │  │prod-api   │  │              │  │eks-cluster │  │
│  │   (OCP)   │  │              │  │   (OCP)   │  │              │  │   (EKS)   │  │
│  │           │  │              │  │           │  │              │  │           │  │
│  │┌─────────┐│  │              │  │┌─────────┐│  │              │  │┌─────────┐│  │
│  ││Pipelines││  │              │  ││Pipelines││  │              │  ││Pipelines││  │
│  ││• Cluster││  │              │  ││• Cluster││  │              │  ││• Cluster││  │
│  ││  Bootstr││  │              │  ││  Bootstr││  │              │  ││  Bootstr││  │
│  ││• Hub    ││  │              │  ││• Hub    ││  │              │  ││• Hub    ││  │
│  ││  Provis ││  │              │  ││  Provis ││  │              │  ││  Provis ││  │
│  │└─────────┘│  │              │  │└─────────┘│  │              │  │└─────────┘│  │
│  │           │  │              │  │           │  │              │  │           │  │
│  │┌─────────┐│  │              │  │┌─────────┐│  │              │  │┌─────────┐│  │
│  ││Regional ││  │              │  ││Regional ││  │              │  ││Regional ││  │
│  ││Services ││  │              │  ││Services ││  │              │  ││Services ││  │
│  ││• Config ││  │              │  ││• Config ││  │              │  ││• Config ││  │
│  ││• Deploy ││  │              │  ││• Deploy ││  │              │  ││• Deploy ││  │
│  │└─────────┘│  │              │  │└─────────┘│  │              │  │└─────────┘│  │
│  └───────────┘  │              │  └───────────┘  │              │  └───────────┘  │
└─────────────────┘              └─────────────────┘              └─────────────────┘

───────────────────────────────────────────────────────────────────────────────────────

## Current GitOps Sync Wave Flow

### Application-Level Sync Wave Orchestration
```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                          GitOps Applications Deployment Order                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

Wave -1: Self-Managing GitOps
┌─────────────────────────────────────────────────────────────────────────────────┐
│ OpenShift GitOps (self-referential)                                            │
│ • ArgoCD manages its own configuration                                         │
│ • Self-managing Application for GitOps operator                                │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
Wave 1: Platform Operators
┌─────────────────────────────────────────────────────────────────────────────────┐
│ OpenShift Pipelines Operator                                                   │
│ • Tekton operator installation on hub cluster                                  │
│ • CRDs: Pipeline, PipelineRun, Task, TaskRun                                   │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
Wave 2: Secret Management
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Vault + External Secrets Operator                                              │
│ • Vault deployment for secure credential storage                               │
│ • ESO deployment for secret synchronization                                    │
│ • Integration with AWS credentials and pull secrets                            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
Wave 3: Advanced Cluster Management (Ordered ApplicationSet)
┌─────────────────────────────────────────────────────────────────────────────────┐
│ ACM ApplicationSet with Internal Ordering:                                     │
│ • Sub-Wave 2: ACM Operator (installs MCH CRD)                                  │
│ • Sub-Wave 3: ACM Hub (creates MultiClusterHub instance)                       │
│ • Sub-Wave 4: ACM Policies (GitOps integration policies)                       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
Wave 4: GitOps Integration & Configuration
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Cluster Provisioning Metrics + Integration                                     │
│ • ACM GitOps cluster integration (automatic cluster registration)              │
│ • Cluster provisioning monitoring and metrics                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
Wave 5: Hub Provisioner Pipelines
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Hub Provisioner Pipeline Infrastructure                                        │
│ • Tekton pipelines for cluster provisioning automation                         │
│ • Cluster lifecycle management workflows                                       │
│ • Automated cluster creation and deletion                                      │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Regional Cluster Provisioning Flow
Once the hub cluster is bootstrapped, regional clusters are provisioned automatically:

```
Hub Cluster ApplicationSet
           │
           ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Cluster-Specific Deployment (per cluster via ApplicationSet)                   │
│                                                                                 │
│ 1. Cluster Provisioning (Wave 1)                                              │
│    • OCP: ClusterDeployment + MachinePool + InstallConfig                      │
│    • EKS: Cluster + AWSManagedControlPlane + AWSManagedMachinePool             │
│    • Target: Hub cluster → cluster-specific namespace                          │
│                                                                                 │
│ 2. Operator Installation (Wave 2)                                              │
│    • OpenShift Pipelines operator to managed cluster                           │
│    • Target: Managed cluster once provisioned                                  │
│                                                                                 │
│ 3. Pipeline Deployment (Wave 3)                                                │
│    • Cluster Bootstrap pipelines                                               │
│    • Hub Provisioner pipelines                                                 │
│    • Target: Managed cluster                                                   │
│                                                                                 │
│ 4. Service Deployment (Wave 4)                                                 │
│    • Regional services and applications                                        │
│    • Target: Managed cluster                                                   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Repository Structure

### Current Directory Organization
```
bootstrap/
├── clusters/
│   ├── global/                     # Hub cluster configuration
│   │   ├── operators/              # Hub cluster operators
│   │   │   ├── openshift-gitops/   # Self-managing GitOps operator
│   │   │   ├── advanced-cluster-management/  # ACM ApplicationSet
│   │   │   ├── openshift-pipelines/# Pipelines hub deployment
│   │   │   ├── vault/              # Vault secret management
│   │   │   └── external-secrets/   # ESO for secret sync
│   │   ├── pipelines/              # Hub cluster pipelines
│   │   │   ├── cluster-bootstrap/  # Bootstrap pipelines
│   │   │   └── hub-provisioner/    # Cluster provisioning pipelines
│   │   └── gitops/                 # Hub cluster GitOps applications
│   │       ├── openshift-gitops/   # Self-managing GitOps
│   │       ├── advanced-cluster-management/  # ACM ApplicationSet
│   │       ├── vault/              # Vault application
│   │       ├── eso/                # ESO application
│   │       └── cluster-bootstrap/  # Bootstrap application
│   │
│   ├── my-cluster/                 # Managed cluster configuration
│   │   ├── my-cluster.yaml         # Cluster specification
│   │   ├── cluster/                # Cluster provisioning resources
│   │   ├── operators/              # Cluster-specific operators
│   │   ├── pipelines/              # Cluster-specific pipelines
│   │   ├── deployments/            # Cluster-specific deployments
│   │   └── gitops/                 # Cluster-specific GitOps applications
│   │
│   └── eks-cluster/                # Another managed cluster
│       ├── eks-cluster.yaml        # Cluster specification
│       ├── cluster/                # EKS: CAPI resources
│       ├── operators/
│       ├── pipelines/
│       ├── deployments/
│       └── gitops/
│
└── bases/                          # Reusable templates
    ├── clusters/                   # Cluster provisioning templates
    └── pipelines/                  # Pipeline templates
```
## Key Architecture Features

### 1. **GitOps-Based Management**
- **Declarative**: All infrastructure defined in Git
- **ArgoCD**: Continuous reconciliation of desired state
- **ApplicationSets**: Automated cluster lifecycle management

### 2. **Application-Level Sync Wave Orchestration**
- **8 sync waves** with proper dependency ordering
- **ApplicationSet approach** for ACM with internal wave ordering
- **Self-managing ArgoCD** that configures itself

### 3. **Multi-Provider Cluster Support**
- **OpenShift (OCP)**: Via Hive ClusterDeployment
- **EKS**: Via CAPI AWSManagedControlPlane
- **Unified Management**: Same ApplicationSet pattern for all cluster types

### 4. **Regional Specification Simplicity**
- **Single file per cluster**: Simple region.yaml format
- **Auto-generation**: Complex Kustomize overlays generated from simple specs
- **Regional organization**: Physical location obvious from directory structure

### 5. **Secret Management Integration**
- **Vault**: Secure credential storage on hub cluster
- **External Secrets Operator**: Automatic secret synchronization
- **No secrets in Git**: All credentials managed through Vault

### 6. **Pipeline-Driven Provisioning**
- **Cluster Bootstrap**: Automated cluster preparation pipelines
- **Hub Provisioner**: Centralized cluster creation workflows
- **Self-Service**: Interactive cluster creation tools

### 7. **GitOps-Native Operations**
- **Single command bootstrap**: `oc apply -k gitops-applications/`
- **Automatic cluster registration**: ACM integration with ArgoCD
- **Declarative management**: All cluster state maintained in Git