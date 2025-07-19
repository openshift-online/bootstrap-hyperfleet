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
│  │ │• ocp-02 │ │    │ │• AWS (EKS)  │ │    │ │• Pipelines  │ │                 │
│  │ │• ocp-03 │ │    │ │• Azure(AKS) │ │    │ │  Operator   │ │                 │
│  │ │• ocp-04 │ │    │ │• GCP (GKE)  │ │    │ │             │ │                 │
│  │ │• eks-02 │ │    │ │• vSphere    │ │    │ │             │ │                 │
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
│  │ocp-02 │  │              │  │ocp-04 │  │              │  │eks-02 │  │
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
│  │ocp-03 │  │              │                 │              │                 │
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
│   │   ├── ocp-02/         # region.yaml (type: ocp)
│   │   └── ocp-03/         # region.yaml (type: ocp)
│   ├── us-west-2/
│   │   └── ocp-04/         # region.yaml (type: ocp)
│   └── ap-southeast-1/
│       └── eks-02/         # region.yaml (type: eks)
├── clusters/                   # Generated cluster configs
│   ├── ocp-02/            # OCP: Hive resources
│   ├── ocp-03/            # OCP: Hive resources  
│   ├── ocp-04/            # OCP: Hive resources
│   └── eks-02/            # EKS: CAPI resources
├── pipelines/                 # Tekton pipelines per cluster
│   ├── hello-world/
│   └── cloud-infrastructure-provisioning/
├── deployments/ocm/           # Service deployments per cluster
├── gitops-applications/       # ArgoCD ApplicationSets
│   ├── ocp-02.yaml
│   ├── ocp-03.yaml
│   ├── ocp-04.yaml
│   └── eks-02.yaml
└── operators/                 # Operator installations per cluster
    ├── ocp-02/
    ├── ocp-03/
    ├── ocp-04/
    └── openshift-pipelines/
```

## Key Architecture Features:

1. **Single Hub Cluster**: Manages all regional clusters via GitOps
2. **Multi-Provider Support**: OCP (Hive) + EKS (CAPI) + AKS/GKE (planned)
3. **Automatic CRD Management**: ACM infrastructure providers handle CAPI CRDs
4. **Sync Wave Ordering**: Ensures proper deployment sequence
5. **Regional Isolation**: Each region contains independent clusters
6. **Unified GitOps**: Single ArgoCD manages all cluster types and regions