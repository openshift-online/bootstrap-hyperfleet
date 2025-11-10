================================================================================
KUSTOMIZE RESOURCE RECONCILIATION CHAIN - Bootstrap Hyperfleet
================================================================================

LEGEND:
  [Base]     = Pure template with placeholders
  (Overlay)  = References base + patches
  {Instance} = Cluster-specific configuration
  >>>        = Kustomize build flow
  ===        = Resource dependency
  ...        = Applied to cluster


┌─────────────────────────────────────────────────────────────────────────────┐
│                      PHASE 1: HUB CLUSTER BOOTSTRAP                         │
└─────────────────────────────────────────────────────────────────────────────┘

User runs: bin/bootstrap
    │
    └──> kustomize build clusters/
              │
              ├──> clusters/kustomization.yaml
              │         │
              │         └──> resources:
              │               └──> global/
              │                     │
              ├─────────────────────┴────────────────────────┐
              │                                               │
              ▼                                               ▼
    ┌─────────────────────┐                        ┌──────────────────┐
    │ OPERATORS BOOTSTRAP │                        │ GITOPS BOOTSTRAP │
    └─────────────────────┘                        └──────────────────┘
              │                                               │
              ▼                                               ▼
    clusters/global/operators/                    clusters/global/gitops/
              │                                               │
              ├──> advanced-cluster-management/              ├──> global/
              │         │                                     │      │
              │         └──> (Overlay)                        │      └──> Applications for:
              │               base: bases/operators/          │            ├─ openshift-gitops
              │               acm/overlays/release-2.14/      │            ├─ acm
              │                     │                         │            ├─ pipelines
              │                     └──> [Base]               │            ├─ vault
              │                           bases/operators/    │            └─ eso
              │                           acm/base/           │
              │                                               ├──> clusters/
              │                                               │      │
              ├──> openshift-gitops/ (Standalone)            │      └──> clusters-applicationset.yaml
              │                                               │            (Meta ApplicationSet)
              ├──> openshift-pipelines/                      │
              │         │                                     ├──> ../../ocp-456/gitops/
              │         └──> (Overlay)                        │      │
              │               base: bases/operators/          │      └──> {Instance Overlay}
              │               pipelines/overlays/1.18/        │            base: bases/clusters/ocp/
              │                     │                         │                  provisioning/
              │                     ├──> [Base]               │                        │
              │                     │     bases/operators/    │                        └──> [Templates]
              │                     │     pipelines/base/     │                              ├─ provisioning.applicationset.yaml
              │                     │                         │                              └─ content.applicationset.yaml
              │                     └──> [Components]         │
              │                           ├─ console-plugin   │            patches: JSON patches
              │                           └─ pipeline-rbac    │            ├─ CLUSTER_NAME → ocp-456
              │                                               │            ├─ paths → clusters/ocp-456/*
              ├──> vault/ (Standalone)                        │            └─ destination → https://api.ocp-456...
              │      └─ configMapGenerator:                   │
              │         vault-policies.hcl                    └──> repo-url-patch.yaml
              │                                                      │
              └──> gitops-integration/ (Standalone)                 └──> replacements:
                     ├─ ManagedClusterSetBinding                          ConfigMap.repo-config.repoURL
                     ├─ Placement                                          ─┬─> Application.spec.source.repoURL
                     ├─ GitOpsCluster                                       └─> ApplicationSet.spec.template.spec.source.repoURL
                     └─ Policies
                                                                              (Dynamic repo URL injection)
              │
              ▼
    ┌────────────────────────────────────┐
    │ Resources Applied to Hub Cluster:  │
    │                                    │
    │  ✓ Namespaces                      │
    │  ✓ Operator Subscriptions          │
    │  ✓ ArgoCD Application (self)       │
    │  ✓ ArgoCD Applications (operators) │
    │  ✓ Vault + ClusterSecretStore      │
    │  ✓ ACM MultiClusterHub             │
    │  ✓ Hub Provisioner Pipelines       │
    └────────────────────────────────────┘
              │
              │ ArgoCD reconciles Applications
              │
              ▼
    ┌─────────────────────────────────────────────┐
    │ ArgoCD auto-syncs all Applications/AppSets │
    └─────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│              PHASE 2: MANAGED CLUSTER PROVISIONING (OCP-456)                │
└─────────────────────────────────────────────────────────────────────────────┘

ArgoCD reads: clusters/global/gitops/kustomization.yaml
    │
    └──> resources:
          └──> ../../ocp-456/gitops/
                    │
                    └──> {Instance Overlay}
                          base: bases/clusters/ocp/provisioning/
                                │
                                ├──> [Template] provisioning.applicationset.yaml
                                │         │
                                │         └──> generators:
                                │               - path: clusters/CLUSTER_NAME/cluster
                                │               - destination: https://k8s.default.svc
                                │
                                └──> [Template] content.applicationset.yaml
                                          │
                                          └──> generators:
                                                - path: clusters/CLUSTER_NAME/operators
                                                - path: clusters/CLUSTER_NAME/pipelines
                                                - path: clusters/CLUSTER_NAME/deployments
                                                - destination: https://api.CLUSTER_NAME.DOMAIN:6443
                          │
                          └──> patches: (JSON Patch)
                                - CLUSTER_NAME → ocp-456
                                - DOMAIN → bootstrap.red-chesterfield.com
                                - paths → clusters/ocp-456/*
                    │
                    ▼
          ┌─────────────────────────────────────┐
          │ ArgoCD ApplicationSets Created:     │
          │                                     │
          │  • ocp-456-provisioning             │
          │  • ocp-456-content                  │
          └─────────────────────────────────────┘
                    │
                    │ ApplicationSet generates Applications
                    │
                    ├──────────────────────┬──────────────────────┐
                    ▼                      ▼                      ▼
          ┌───────────────────┐  ┌──────────────────┐  ┌─────────────────┐
          │ ocp-456-cluster   │  │ ocp-456-operators│  │ ocp-456-pipelines│
          │ (sync-wave: 10)   │  │ (sync-wave: 10)  │  │ (sync-wave: 20)  │
          └───────────────────┘  └──────────────────┘  └─────────────────┘
                    │                      │                      │
                    │ Builds:              │ Builds:              │ Builds:
                    ▼                      ▼                      ▼
          clusters/ocp-456/    clusters/ocp-456/    clusters/ocp-456/
          cluster/              operators/           pipelines/cloud-infrastructure/
                    │                      │                      │
                    └──> {Instance}        └──> (Overlay)         └──> (Overlay + Component)
                          base: bases/            base: bases/           component: bases/pipelines/
                          clusters/ocp/           operators/pipelines/   cloud-infra-provisioning/
                                │                  overlays/operator-only/      │
                                │                           │                   └──> [Component]
                                └──> [Base]                 └──> (Overlay)           cloud-infra.pipeline.yaml
                                      ├─ clusterdeployment       base: ../../base/
                                      ├─ managedcluster                │
                                      ├─ machinepool                   └──> [Base]
                                      └─ external-secrets                    subscription.yaml
                                │
                                └──> local resources:
                                      ├─ namespace.yaml
                                      ├─ klusterletaddonconfig.yaml
                                      └─ install-config.yaml
                                │
                                └──> secretGenerator:
                                      install-config (from install-config.yaml)
                                │
                                └──> patches: (JSON Patch)
                                      ├─ ClusterDeployment:
                                      │   ├─ namespace → ocp-456
                                      │   ├─ name → ocp-456
                                      │   ├─ clusterName → ocp-456
                                      │   └─ region → us-west-2
                                      │
                                      ├─ ManagedCluster:
                                      │   ├─ namespace → ocp-456
                                      │   ├─ name → ocp-456
                                      │   └─ labels.region → us-west-2
                                      │
                                      ├─ MachinePool:
                                      │   ├─ namespace → ocp-456
                                      │   ├─ name → ocp-456-worker
                                      │   └─ clusterDeploymentRef → ocp-456
                                      │
                                      ├─ KlusterletAddonConfig:
                                      │   ├─ namespace → ocp-456
                                      │   ├─ name → ocp-456
                                      │   └─ cluster labels
                                      │
                                      └─ ExternalSecret:
                                          ├─ namespace → ocp-456
                                          └─ template.type → dockerconfigjson
                    │
                    ▼
          ┌─────────────────────────────────────────────────┐
          │ Resources Applied to Hub Cluster:               │
          │                                                 │
          │  ✓ Namespace: ocp-456                          │
          │  ✓ Secret: install-config (generated)          │
          │  ✓ ClusterDeployment (Hive)                    │
          │     ├─ name: ocp-456                           │
          │     ├─ namespace: ocp-456                      │
          │     ├─ region: us-west-2                       │
          │     └─ installConfigSecretRef: install-config  │
          │  ✓ ManagedCluster (ACM)                        │
          │  ✓ MachinePool (Hive)                          │
          │  ✓ KlusterletAddonConfig (ACM)                 │
          │  ✓ ExternalSecret: aws-credentials             │
          │  ✓ ExternalSecret: pull-secret                 │
          └─────────────────────────────────────────────────┘
                    │
                    │ Hive ClusterDeployment Controller
                    │
                    ▼
          ┌─────────────────────────────────────┐
          │ Hive Provisions OpenShift Cluster:  │
          │                                     │
          │  1. Reads install-config Secret     │
          │  2. Generates ignition configs      │
          │  3. Creates AWS resources:          │
          │     ├─ VPC, Subnets, IGW, NAT       │
          │     ├─ ELBs, Security Groups        │
          │     └─ EC2 instances                │
          │  4. Bootstraps OpenShift            │
          │  5. Updates ClusterDeployment:      │
          │     ├─ infraID                      │
          │     ├─ clusterID                    │
          │     ├─ adminKubeconfigSecretRef     │
          │     └─ status: Installed            │
          └─────────────────────────────────────┘
                    │
                    │ Cluster becomes available
                    │
                    ▼
          ┌─────────────────────────────────────┐
          │ ACM ManagedCluster Controller:      │
          │                                     │
          │  1. Detects ClusterDeployment ready │
          │  2. Generates import manifest       │
          │  3. Applies to managed cluster      │
          │  4. Klusterlet installed            │
          │  5. ManagedCluster: Available=True  │
          └─────────────────────────────────────┘
                    │
                    │ ApplicationSet detects cluster API available
                    │
                    ▼
          ┌─────────────────────────────────────────────┐
          │ ocp-456-content Applications Deploy:        │
          │                                             │
          │  • ocp-456-operators                        │
          │    └─> Destination: https://api.ocp-456... │
          │        ├─ OpenShift Pipelines operator      │
          │        └─ namespace: ocm-ocp-456            │
          │                                             │
          │  • ocp-456-pipelines                        │
          │    └─> cloud-infrastructure pipeline        │
          │                                             │
          │  • ocp-456-deployments                      │
          │    └─> OCM namespace                        │
          └─────────────────────────────────────────────┘
                    │
                    ▼
          ┌─────────────────────────────────────┐
          │ Managed Cluster (ocp-456) Running: │
          │                                     │
          │  ✓ OpenShift 4.x                    │
          │  ✓ Pipelines operator installed     │
          │  ✓ ACM agent (klusterlet) running   │
          │  ✓ Managed from hub cluster         │
          └─────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│                   PHASE 3: CLUSTER DEPROVISIONING                           │
└─────────────────────────────────────────────────────────────────────────────┘

User runs: bin/cluster-deprovision ocp-456
    │
    ├──> Updates clusters/global/gitops/kustomization.yaml
    │    └──> Adds: - ../../ocp-456/deprovisioning/
    │
    └──> Stages: clusters/ocp-456/deprovisioning/
                   │
                   ├──> deprovision.applicationset.yaml
                   │         (sync-wave: 1)
                   │
                   └──> cleanup.applicationset.yaml
                            (sync-wave: 2)

User commits and pushes:
    │
    └──> ArgoCD detects new ApplicationSets
              │
              ├───────────────────┬────────────────────┐
              ▼                   ▼                    ▼
    ┌─────────────────┐ ┌──────────────────┐ ┌───────────────────┐
    │ Wave 1: Deprov  │ │ Wait for AWS     │ │ Wave 2: Cleanup   │
    │                 │ │ cleanup complete │ │                   │
    │ ApplicationSet  │ │                  │ │ ApplicationSet    │
    │ Creates:        │ │                  │ │ Creates:          │
    │                 │ │                  │ │                   │
    │ Application →   │ │                  │ │ Application →     │
    │   Deploys:      │ │                  │ │   Triggers:       │
    │   ClusterDe-    │ │   Hive deletes:  │ │   PipelineRun     │
    │   provision     ├─┼─> • ELBs         │ │   (cluster-remove)│
    │   (in hive ns)  │ │   • EC2s         │ │                   │
    │                 │ │   • VPCs         │ │   Deletes:        │
    │                 │ │   • DNS records  ├─┼─> • Namespace     │
    │                 │ │                  │ │   • Git files     │
    │                 │ │   Finalizer      │ │   • ApplicationSets│
    │                 │ │   completes      │ │                   │
    └─────────────────┘ └──────────────────┘ └───────────────────┘
              │                   │                    │
              │                   │                    │
              └───────────────────┴────────────────────┘
                                  │
                                  ▼
                  ┌────────────────────────────────────┐
                  │ Cluster Fully Removed:             │
                  │                                    │
                  │  ✓ AWS resources deleted           │
                  │  ✓ Namespace deleted               │
                  │  ✓ Git repo cleaned                │
                  │  ✓ ApplicationSets removed         │
                  └────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────┐
│                        KUSTOMIZE BUILD MECHANICS                            │
└─────────────────────────────────────────────────────────────────────────────┘

When: kustomize build clusters/ocp-456/cluster/

Step 1: Load kustomization.yaml
    │
    └──> resources:
          ├─ namespace.yaml (local)
          ├─ klusterletaddonconfig.yaml (local)
          └─ ../../../bases/clusters/ocp/ (base reference)

Step 2: Load Base
    │
    └──> bases/clusters/ocp/kustomization.yaml
          │
          └──> resources:
                ├─ clusterdeployment.yaml
                ├─ managedcluster.yaml
                ├─ machinepool.yaml
                └─ external-secrets.yaml

Step 3: Merge Resources
    │
    └──> Combined resource list:
          ├─ namespace.yaml (from overlay)
          ├─ klusterletaddonconfig.yaml (from overlay)
          ├─ clusterdeployment.yaml (from base)
          ├─ managedcluster.yaml (from base)
          ├─ machinepool.yaml (from base)
          └─ external-secrets.yaml (from base)

Step 4: Run secretGenerator
    │
    └──> secretGenerator:
          - name: install-config
            namespace: ocp-456
            files:
              - install-config.yaml
          │
          └──> Creates Secret manifest with base64 encoded install-config.yaml

Step 5: Apply JSON Patches (in order)
    │
    ├──> Patch 1: ClusterDeployment
    │    ├─ /metadata/namespace: ocp-123 → ocp-456
    │    ├─ /metadata/name: CLUSTER_NAME → ocp-456
    │    ├─ /spec/clusterName: CLUSTER_NAME → ocp-456
    │    └─ /spec/platform/aws/region: REGION → us-west-2
    │
    ├──> Patch 2: ManagedCluster
    │    ├─ /metadata/namespace: ocp-123 → ocp-456
    │    ├─ /metadata/name: CLUSTER_NAME → ocp-456
    │    ├─ /metadata/labels/name: CLUSTER_NAME → ocp-456
    │    └─ /metadata/labels/region: REGION → us-west-2
    │
    ├──> Patch 3: MachinePool
    │    ├─ /metadata/namespace: ocp-123 → ocp-456
    │    ├─ /metadata/name: CLUSTER_NAME → ocp-456-worker
    │    └─ /spec/clusterDeploymentRef/name: CLUSTER_NAME → ocp-456
    │
    ├──> Patch 4: KlusterletAddonConfig
    │    ├─ /metadata/namespace: ocp-123 → ocp-456
    │    ├─ /metadata/name: CLUSTER_NAME → ocp-456
    │    ├─ /spec/clusterLabels/name: CLUSTER_NAME → ocp-456
    │    ├─ /spec/clusterNamespace: CLUSTER_NAME → ocp-456
    │    └─ /spec/clusterName: CLUSTER_NAME → ocp-456
    │
    └──> Patch 5-6: ExternalSecrets
         ├─ aws-credentials: /metadata/namespace → ocp-456
         └─ pull-secret: /metadata/namespace → ocp-456
                         /spec/target/template/type → kubernetes.io/dockerconfigjson

Step 6: Apply generatorOptions
    │
    └──> disableNameSuffixHash: true
          │
          └──> Secret name: install-config (NOT install-config-abc123xyz)

Step 7: Output Final Manifests
    │
    └──> YAML stream with all resources:
          ├─ Namespace/ocp-456
          ├─ Secret/install-config (with install-config.yaml content)
          ├─ ClusterDeployment/ocp-456 (namespace: ocp-456, region: us-west-2)
          ├─ ManagedCluster/ocp-456
          ├─ MachinePool/ocp-456-worker
          ├─ KlusterletAddonConfig/ocp-456
          ├─ ExternalSecret/aws-credentials (namespace: ocp-456)
          └─ ExternalSecret/pull-secret (namespace: ocp-456)


┌─────────────────────────────────────────────────────────────────────────────┐
│                   REPLACEMENTS MECHANISM (Advanced)                         │
└─────────────────────────────────────────────────────────────────────────────┘

In: clusters/global/gitops/kustomization.yaml

Step 1: Load repo-url-patch.yaml
    │
    └──> apiVersion: v1
          kind: ConfigMap
          metadata:
            name: repo-config
          data:
            repoURL: https://github.com/openshift-online/bootstrap-hyperfleet

Step 2: Process replacements field
    │
    └──> replacements:
          - source:
              kind: ConfigMap
              name: repo-config
              fieldPath: data.repoURL
            targets:
            - select:
                kind: Application
              fieldPaths:
              - spec.source.repoURL
              reject:
              - annotationSelector: external-repo=true
            - select:
                kind: ApplicationSet
              fieldPaths:
              - spec.template.spec.source.repoURL

Step 3: Find Source Value
    │
    └──> Extract: ConfigMap.repo-config.data.repoURL
          = "https://github.com/openshift-online/bootstrap-hyperfleet"

Step 4: Find Target Resources
    │
    ├──> Scan all resources for kind: Application
    │    └──> Skip if annotation: external-repo=true
    │         (e.g., Helm chart repos)
    │
    └──> Scan all resources for kind: ApplicationSet

Step 5: Apply Replacement
    │
    ├──> Application (openshift-gitops)
    │    spec:
    │      source:
    │        repoURL: <placeholder> → https://github.com/.../bootstrap-hyperfleet
    │
    ├──> ApplicationSet (ocp-456-provisioning)
    │    spec:
    │      template:
    │        spec:
    │          source:
    │            repoURL: <placeholder> → https://github.com/.../bootstrap-hyperfleet
    │
    └──> ApplicationSet (ocp-456-content)
         spec:
           template:
             spec:
               source:
                 repoURL: <placeholder> → https://github.com/.../bootstrap-hyperfleet

Result: All Applications/ApplicationSets reference the same dynamic repo URL


┌─────────────────────────────────────────────────────────────────────────────┐
│                        COMPONENT COMPOSITION                                │
└─────────────────────────────────────────────────────────────────────────────┘

In: bases/operators/openshift-pipelines/overlays/pipelines-1.18/kustomization.yaml

Step 1: Load base
    │
    └──> resources:
          - ../../base/
               │
               └──> subscription.yaml (pipelines operator)

Step 2: Load components
    │
    ├──> components:
    │     - ../../components/enable-console-plugin/
    │            │
    │            └──> configMapGenerator:
    │                  - name: job-tekton-console-plugin-scripts
    │                    files:
    │                      - enable-console-plugin.sh
    │                 resources:
    │                  - console-plugin-job.yaml
    │                  - console-plugin.yaml
    │
    └──> components:
          - ../../components/pipeline-rbac/
                 │
                 └──> resources:
                       - pipeline-rbac.yaml

Step 3: Merge all resources
    │
    └──> Combined:
          ├─ subscription.yaml (from base)
          ├─ console-plugin-job.yaml (from component 1)
          ├─ console-plugin.yaml (from component 1)
          ├─ pipeline-rbac.yaml (from component 2)
          └─ ConfigMap/job-tekton-console-plugin-scripts (generated)

Step 4: Apply patches
    │
    └──> patches:
          - patch-channel.yaml
               │
               └──> Strategic merge patch:
                     apiVersion: operators.coreos.com/v1alpha1
                     kind: Subscription
                     metadata:
                       name: openshift-pipelines-operator
                     spec:
                       channel: pipelines-1.18

Result: Full-featured pipelines operator with console plugin + RBAC


================================================================================
                              END OF DIAGRAM
================================================================================
