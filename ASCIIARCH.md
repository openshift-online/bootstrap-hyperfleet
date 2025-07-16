# OpenShift Bootstrap Architecture

```
                    ┌─ OpenShift Bootstrap Architecture ─┐
                    │                                     │
                    │  ┌─────────────────────────────────┐│
                    │  │     EKS Hub Cluster             ││
                    │  │   (acme-test-001 us-east-1)     ││
                    │  │                                 ││
                    │  │  ┌─────────────────────────────┐││
                    │  │  │    OpenShift GitOps         │││
                    │  │  │      (ArgoCD)               │││
                    │  │  │                             │││
                    │  │  │  ┌─────────────────────────┐│││
                    │  │  │  │   Regional Clusters     ││││
                    │  │  │  │    Applications         ││││
                    │  │  │  └─────────────────────────┘│││
                    │  │  └─────────────────────────────┘││
                    │  │                                 ││
                    │  │  ┌─────────────────────────────┐││
                    │  │  │   ACM MultiClusterHub       │││
                    │  │  │                             │││
                    │  │  │  ┌─────────┐ ┌─────────────┐│││
                    │  │  │  │  CAPI   │ │ Infrastructure││││
                    │  │  │  │ CRDs    │ │  Providers    ││││
                    │  │  │  │         │ │ AWS|Azure|GCP ││││
                    │  │  │  └─────────┘ └─────────────┘│││
                    │  │  └─────────────────────────────┘││
                    │  └─────────────────────────────────┘│
                    └─────────────────────────────────────┘
                                     │
                        ┌────────────┼────────────┐
                        │            │            │
                   ┌─────▼─────┐ ┌────▼────┐ ┌────▼────┐
                   │ Regional  │ │Regional │ │Regional │
                   │Cluster-41 │ │Cluster-42│ │Cluster-43│
                   │(us-west-2)│ │(ap-se-1) │ │(us-west-2)│
                   │  EKS      │ │  EKS     │ │  EKS     │
                   │  Stage    │ │  Prod    │ │  Stage   │
                   └───────────┘ └─────────┘ └─────────┘

       ┌─── Configuration Flow ───┐
       │                          │
       │  regions/us-west-2/      │         bin/
       │    eks-stage/            │    ┌─────────────┐
       │      region.yaml         │────│ Converter   │
       │     (12 lines)           │    │   Tools     │
       │                          │    │             │
       │           │              │    │ convert-    │
       │           ▼              │    │ cluster     │
       │  bin/generate-cluster    │    │             │
       │           │              │    │ generate-   │
       │           ▼              │    │ cluster     │
       │  clusters/overlay/       │    └─────────────┘
       │    cluster-43/           │
       │      (7 files, 110 lines)│
       │                          │
       │           │              │
       │           ▼              │    ┌─────────────┐
       │  gitops-applications/    │    │   GitOps    │
       │    regional-clusters.    │────│ Applications│
       │      cluster-43.yaml    │    │             │
       │    regional-deployments. │    │ ArgoCD      │
       │      cluster-43.yaml    │    │ Deployment  │
       └──────────────────────────┘    └─────────────┘

Technology Stack:
┌──────────────────────────────────────────────────────────────┐
│ Cluster Types: │ OCP (Hive)          │ EKS (CAPI)           │
│ Provisioning:  │ ClusterDeployment   │ AWSManagedControl    │
│               │ MachinePool         │ AWSManagedMachine    │
│ Management:    │ ACM ManagedCluster  │ ACM ManagedCluster   │
│ GitOps:        │ ArgoCD Applications │ ArgoCD Applications  │
│ Config:        │ Pure Kustomize      │ Pure Kustomize       │
└──────────────────────────────────────────────────────────────┘

Complexity Reduction:
   Before: 200+ lines, 7 files, JSON patches
      ████████████████████████████████████████
   After:  12 lines, 1 file, direct YAML  
      ████
```

## Architecture Components

### Hub Cluster (EKS)
- **Name**: acme-test-001
- **Region**: us-east-1
- **Purpose**: Central control plane for multi-cluster management
- **Components**:
  - OpenShift GitOps (ArgoCD)
  - Red Hat Advanced Cluster Management (ACM)
  - Cluster API (CAPI) controllers
  - Infrastructure providers (AWS, Azure, GCP)

### Regional Clusters
- **cluster-41**: EKS stage (us-west-2, m5.large, 1-10 nodes)
- **cluster-42**: EKS prod (ap-southeast-1, m5.xlarge, 2-20 nodes)
- **cluster-43**: EKS stage (us-west-2, m5.xlarge, 3 nodes)

### Configuration Management
- **Regional Specifications**: Minimal 12-line YAML files
- **Converter Tools**: Transform complex overlays to simple specs
- **Generator Tools**: Create complete Kustomize overlays from specs
- **GitOps Integration**: Automatic ArgoCD application generation

### Technology Stack
- **OpenShift GitOps**: Continuous deployment
- **ACM**: Multi-cluster management with CAPI integration
- **CAPI**: Kubernetes-native cluster lifecycle
- **Kustomize**: Configuration templating
- **Infrastructure Providers**: Cloud-specific cluster provisioning

### Workflow
1. Create minimal regional specification (12 lines)
2. Generate complete overlay with converter tools (110+ lines)
3. Deploy via ArgoCD applications
4. Provision clusters via CAPI/ACM
5. Manage with ACM governance