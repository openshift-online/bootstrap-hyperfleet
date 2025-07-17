# OpenShift Bootstrap Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                 HUB CLUSTER                                          │
│                            (OpenShift + ArgoCD + ACM)                               │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                 │
│  │   ArgoCD        │    │      ACM        │    │   Tekton        │                 │
│  │   GitOps        │    │ MultiClusterHub │    │   Pipelines     │                 │
│  │                 │    │                 │    │                 │                 │
│  │ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │                 │
│  │ │Applications │ │    │ │Infrastructur│ │    │ │   Global    │ │                 │
│  │ │    ets      │ │    │ │  Providers  │ │    │ │  Operators  │ │                 │
│  │ │             │ │    │ │             │ │    │ │             │ │                 │
│  │ │• cluster-10 │ │    │ │• AWS (EKS)  │ │    │ │• Pipelines  │ │                 │
│  │ │• cluster-20 │ │    │ │• Azure(AKS) │ │    │ │  Operator   │ │                 │
│  │ │• cluster-30 │ │    │ │• GCP (GKE)  │ │    │ │             │ │                 │
│  │ │• cluster-40 │ │    │ │• vSphere    │ │    │ │             │ │                 │
│  │ └─────────────┘ │    │ │• OpenStack  │ │    │ └─────────────┘ │                 │
│  └─────────────────┘    │ │• BareMetal  │ │    └─────────────────┘                 │
│                         │ └─────────────┘ │                                        │
│                         └─────────────────┘                                        │
│                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                      CAPI CRDs & Resources                                  │   │
│  │                                                                             │   │
│  │  OCP Clusters (Hive)           EKS Clusters (CAPI)                        │   │
│  │  ┌─────────────────┐            ┌─────────────────┐                        │   │
│  │  │ClusterDeployment│            │AWSManagedControl│                        │   │
│  │  │MachinePool      │            │Plane            │                        │   │
│  │  │InstallConfig    │            │AWSManagedMachine│                        │   │
│  │  │                 │            │Pool             │                        │   │
│  │  └─────────────────┘            └─────────────────┘                        │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                          │
                                          │ GitOps Sync
                                          │
      ┌───────────────────────────────────┼───────────────────────────────────┐
      │                                   │                                   │
      ▼                                   ▼                                   ▼
┌─────────────────┐              ┌─────────────────┐              ┌─────────────────┐
│   us-east-1     │              │   us-west-2     │              │ ap-southeast-1  │
│                 │              │                 │              │                 │
│  ┌───────────┐  │              │  ┌───────────┐  │              │  ┌───────────┐  │
│  │cluster-10 │  │              │  │cluster-30 │  │              │  │cluster-40 │  │
│  │   (OCP)   │  │              │  │   (OCP)   │  │              │  │   (EKS)   │  │
│  │           │  │              │  │           │  │              │  │           │  │
│  │┌─────────┐│  │              │  │┌─────────┐│  │              │  │┌─────────┐│  │
│  ││Pipelines││  │              │  ││Pipelines││  │              │  ││Pipelines││  │
│  ││• Hello  ││  │              │  ││• Hello  ││  │              │  ││• Hello  ││  │
│  ││• Cloud  ││  │              │  ││• Cloud  ││  │              │  ││• Cloud  ││  │
│  ││  Infra  ││  │              │  ││  Infra  ││  │              │  ││  Infra  ││  │
│  │└─────────┘│  │              │  │└─────────┘│  │              │  │└─────────┘│  │
│  │           │  │              │  │           │  │              │  │           │  │
│  │┌─────────┐│  │              │  │┌─────────┐│  │              │  │┌─────────┐│  │
│  ││OCM      ││  │              │  ││OCM      ││  │              │  ││OCM      ││  │
│  ││Services ││  │              │  ││Services ││  │              │  ││Services ││  │
│  ││• AMS-DB ││  │              │  ││• AMS-DB ││  │              │  ││• AMS-DB ││  │
│  ││• OSL-DB ││  │              │  ││• OSL-DB ││  │              │  ││• OSL-DB ││  │
│  ││• CS-DB  ││  │              │  ││• CS-DB  ││  │              │  ││• CS-DB  ││  │
│  │└─────────┘│  │              │  │└─────────┘│  │              │  │└─────────┘│  │
│  └───────────┘  │              │  └───────────┘  │              │  └───────────┘  │
│                 │              │                 │              │                 │
│  ┌───────────┐  │              │                 │              │                 │
│  │cluster-20 │  │              │                 │              │                 │
│  │   (OCP)   │  │              │                 │              │                 │
│  │           │  │              │                 │              │                 │
│  │┌─────────┐│  │              │                 │              │                 │
│  ││Pipelines││  │              │                 │              │                 │
│  ││• Hello  ││  │              │                 │              │                 │
│  ││• Cloud  ││  │              │                 │              │                 │
│  ││  Infra  ││  │              │                 │              │                 │
│  │└─────────┘│  │              │                 │              │                 │
│  │           │  │              │                 │              │                 │
│  │┌─────────┐│  │              │                 │              │                 │
│  ││OCM      ││  │              │                 │              │                 │
│  ││Services ││  │              │                 │              │                 │
│  ││• AMS-DB ││  │              │                 │              │                 │
│  ││• OSL-DB ││  │              │                 │              │                 │
│  ││• CS-DB  ││  │              │                 │              │                 │
│  │└─────────┘│  │              │                 │              │                 │
│  └───────────┘  │              │                 │              │                 │
└─────────────────┘              └─────────────────┘              └─────────────────┘

───────────────────────────────────────────────────────────────────────────────────────

GitOps Sync Wave Flow:
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                     │
│  Wave 1: Cluster Provisioning                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │ Hub Cluster → cluster-XX namespace → CAPI/Hive Resources                   │   │
│  │ • OCP: ClusterDeployment + MachinePool + InstallConfig                     │   │
│  │ • EKS: Cluster + AWSManagedControlPlane + AWSManagedMachinePool            │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                               │
│                                    ▼                                               │
│  Wave 2: Operators Installation                                                    │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │ Hub Cluster → Managed Cluster                                               │   │
│  │ • OpenShift Pipelines Operator                                             │   │
│  │ • CRDs: Pipeline, PipelineRun, Task, TaskRun                              │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                               │
│                                    ▼                                               │
│  Wave 3: Pipeline Deployment                                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │ Hub Cluster → Managed Cluster                                               │   │
│  │ • Hello World Pipeline + PipelineRun                                       │   │
│  │ • Cloud Infrastructure Pipeline + PipelineRun                              │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                               │
│                                    ▼                                               │
│  Wave 4: Service Deployment                                                        │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │ Hub Cluster → Managed Cluster                                               │   │
│  │ • OCM Services: AMS-DB, OSL-DB, CS-DB                                      │   │
│  │ • Persistent Volumes + Services + Deployments                              │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘

Repository Structure:
├── regions/                     # Regional specifications
│   ├── us-east-1/
│   │   ├── cluster-10/         # region.yaml (type: ocp)
│   │   └── cluster-20/         # region.yaml (type: ocp)
│   ├── us-west-2/
│   │   └── cluster-30/         # region.yaml (type: ocp)
│   └── ap-southeast-1/
│       └── cluster-40/         # region.yaml (type: eks)
├── clusters/                   # Generated cluster configs
│   ├── cluster-10/            # OCP: Hive resources
│   ├── cluster-20/            # OCP: Hive resources  
│   ├── cluster-30/            # OCP: Hive resources
│   └── cluster-40/            # EKS: CAPI resources
├── pipelines/                 # Tekton pipelines per cluster
│   ├── hello-world/
│   └── cloud-infrastructure-provisioning/
├── deployments/ocm/           # Service deployments per cluster
├── gitops-applications/       # ArgoCD ApplicationSets
│   ├── cluster-10.yaml
│   ├── cluster-20.yaml
│   ├── cluster-30.yaml
│   └── cluster-40.yaml
└── operators/                 # Operator installations per cluster
    ├── cluster-10/
    ├── cluster-20/
    ├── cluster-30/
    └── openshift-pipelines/
```

## Key Architecture Features:

1. **Single Hub Cluster**: Manages all regional clusters via GitOps
2. **Multi-Provider Support**: OCP (Hive) + EKS (CAPI) + AKS/GKE (planned)
3. **Automatic CRD Management**: ACM infrastructure providers handle CAPI CRDs
4. **Sync Wave Ordering**: Ensures proper deployment sequence
5. **Regional Isolation**: Each region contains independent clusters
6. **Unified GitOps**: Single ArgoCD manages all cluster types and regions