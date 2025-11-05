# Cluster Directory Structure Proposal

## Current Structure (Scattered)
```
regions/us-west-2/ocp-foo/
  └── region.yaml                           # Cluster spec

clusters/ocp-foo/
  ├── namespace.yaml
  ├── install-config.yaml
  ├── klusterletaddonconfig.yaml
  └── kustomization.yaml

operators/openshift-pipelines/ocp-foo/
  ├── namespace.yaml
  └── kustomization.yaml

pipelines/cloud-infrastructure-provisioning/ocp-foo/
  ├── kustomization.yaml
  └── cloud-infrastructure-provisioning.pipelinerun.yaml

deployments/ocm/ocp-foo/
  ├── namespace.yaml
  └── kustomization.yaml

gitops-applications/clusters/
  └── ocp-foo.yaml                          # ApplicationSets
```

---

## Proposed Structure (Consolidated)

### Top-Level Directory Structure
```
bootstrap-hyperfleet/
  ├── clusters/                   # All cluster definitions
  │   ├── global/                 # Hub cluster resources
  │   ├── ocp-foo/                # Full cluster definition
  │   ├── ocp-test/               # Another cluster
  │   ├── eks-01/                 # EKS cluster
  │   ├── hcp-dev/                # HCP cluster
  │   └── kustomization.yaml      # References all clusters
  │
  ├── bases/                      # Shared base resources
  │   ├── clusters/
  │   ├── pipelines/
  │   ├── ocm/
  │   └── applications/
  │
  └── bin/                        # Tools
      ├── cluster-create
      ├── cluster-generate
      ├── cluster-remove
      └── bootstrap
```

---

### Global Directory (Hub Cluster Resources)
```
clusters/global/
  ├── cluster/                              # Hub cluster configuration
  │   └── kustomization.yaml
  │
  ├── operators/                            # Hub operators
  │   ├── advanced-cluster-management/
  │   │   ├── namespace.yaml
  │   │   ├── subscription.yaml
  │   │   ├── multiclusterhub.yaml
  │   │   └── kustomization.yaml
  │   ├── openshift-gitops/
  │   │   └── kustomization.yaml
  │   ├── openshift-pipelines/
  │   │   └── kustomization.yaml
  │   ├── vault/
  │   │   ├── vault-deployment.yaml
  │   │   ├── cluster-secret-store.yaml
  │   │   └── kustomization.yaml
  │   ├── external-secrets/
  │   │   └── kustomization.yaml
  │   └── kustomization.yaml
  │
  ├── pipelines/                            # Hub pipelines
  │   ├── hub-provisioner/
  │   │   ├── namespace.yaml
  │   │   ├── hub-provisioner.pipeline.yaml
  │   │   ├── bootstrap.pipeline.yaml
  │   │   ├── github-credentials.yaml
  │   │   └── kustomization.yaml
  │   └── kustomization.yaml
  │
  ├── deployments/                          # Hub services
  │   └── kustomization.yaml
  │
  ├── gitops/                               # Global GitOps ApplicationSets
  │   ├── cluster-bootstrap.applicationset.yaml
  │   ├── eso.application.yaml
  │   ├── ocp-foo.yaml                      # Per-cluster ApplicationSets
  │   ├── ocp-test.yaml
  │   └── kustomization.yaml
  │
  └── kustomization.yaml                    # Root kustomization for global
```

---

### Per-Cluster Directory
```
clusters/ocp-foo/
  ├── ocp-foo.yaml                          # Cluster spec (from region.yaml)
  │
  ├── cluster/                              # Cluster provisioning resources
  │   ├── namespace.yaml
  │   ├── install-config.yaml               # OCP: install-config
  │   ├── cluster.yaml                      # EKS: CAPI Cluster
  │   ├── awsmanagedcontrolplane.yaml       # EKS: Control plane
  │   ├── awsmanagedmachinepool.yaml        # EKS: Worker nodes
  │   ├── machinepool.yaml                  # EKS: MachinePool link
  │   ├── hostedcluster.yaml                # HCP: HostedCluster
  │   ├── nodepool.yaml                     # HCP: NodePool
  │   ├── managedcluster.yaml               # ACM ManagedCluster
  │   ├── klusterletaddonconfig.yaml        # ACM addons
  │   ├── external-secrets.yaml             # Vault credentials
  │   ├── acm-integration-pipeline.yaml     # EKS: ACM integration
  │   └── kustomization.yaml
  │
  ├── operators/                            # Operator deployments for managed cluster
  │   ├── namespace.yaml                    # ocm-ocp-foo namespace
  │   ├── pipelines-operator/               # OpenShift Pipelines
  │   │   └── kustomization.yaml
  │   └── kustomization.yaml
  │
  ├── pipelines/                            # Pipeline configurations
  │   ├── cloud-infrastructure/             # Cloud infra pipelines
  │   │   ├── cloud-infrastructure-provisioning.pipelinerun.yaml
  │   │   └── kustomization.yaml
  │   └── kustomization.yaml
  │
  ├── deployments/                          # Service deployments
  │   ├── namespace.yaml                    # ocm-ocp-foo namespace
  │   ├── ocm/                              # OCM services
  │   │   └── kustomization.yaml
  │   └── kustomization.yaml
  │
  ├── gitops/                               # GitOps ApplicationSets
  │   ├── provisioning.applicationset.yaml  # Provisions cluster on hub
  │   ├── content.applicationset.yaml       # Deploys to managed cluster
  │   └── kustomization.yaml
  │
  └── kustomization.yaml                    # Root kustomization for entire cluster
```

---

## Benefits

1. **Three top-level directories**: Clean separation - `clusters/`, `bases/`, `bin/`
2. **Consistent structure**: `global/` and all clusters use same layout
3. **Single source of truth**: All `ocp-foo` resources in `clusters/ocp-foo/`
4. **Easy navigation**: `cd clusters/ocp-foo` to see everything
5. **Clear separation**: Hub resources in `clusters/global/`, cluster resources in `clusters/{cluster-name}/`
6. **Simple cleanup**: `rm -rf clusters/ocp-foo` removes entire cluster
7. **Self-documenting**: Directory structure shows what's deployed
8. **Eliminated confusion**: No more scattered `operators/`, `pipelines/`, `deployments/`, `gitops-applications/`

---

## Root Kustomization Examples

### Top-Level Clusters
```yaml
# clusters/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - global/
  - ocp-foo/
  - ocp-test/
  - eks-01/
```

### Per-Cluster Root
```yaml
# clusters/ocp-foo/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - cluster/
  - operators/
  - pipelines/
  - deployments/
  - gitops/
```

### Global Root
```yaml
# clusters/global/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - operators/
  - pipelines/
  - gitops/
```

---

## Migration Path

### Directory Moves
```bash
# Hub resources
mkdir -p clusters/global
mv operators/advanced-cluster-management/global/     → clusters/global/operators/advanced-cluster-management/
mv operators/openshift-gitops/global/                → clusters/global/operators/openshift-gitops/
mv operators/openshift-pipelines/global/             → clusters/global/operators/openshift-pipelines/
mv operators/vault/global/                           → clusters/global/operators/vault/
mv operators/external-secrets/global/                → clusters/global/operators/external-secrets/
mv pipelines/hub-provisioner/                        → clusters/global/pipelines/hub-provisioner/
mv gitops-applications/                              → clusters/global/gitops/

# Cluster resources (ocp-foo example)
mkdir -p clusters/ocp-foo
mv clusters/ocp-foo/                                 → clusters/ocp-foo/cluster/
mv operators/openshift-pipelines/ocp-foo/            → clusters/ocp-foo/operators/
mv pipelines/cloud-infrastructure-provisioning/ocp-foo/ → clusters/ocp-foo/pipelines/cloud-infrastructure/
mv deployments/ocm/ocp-foo/                          → clusters/ocp-foo/deployments/ocm/
mv regions/us-west-2/ocp-foo/region.yaml             → clusters/ocp-foo/ocp-foo.yaml

# Create new ApplicationSets location
mkdir -p clusters/ocp-foo/gitops
mv clusters/global/gitops/ocp-foo.yaml               → clusters/ocp-foo/gitops/ (split into provisioning + content)
```

### Changed Paths in ApplicationSets
```yaml
# OLD
path: clusters/ocp-foo
path: operators/openshift-pipelines/ocp-foo
path: pipelines/cloud-infrastructure-provisioning/ocp-foo
path: deployments/ocm/ocp-foo

# NEW
path: clusters/ocp-foo/cluster
path: clusters/ocp-foo/operators
path: clusters/ocp-foo/pipelines
path: clusters/ocp-foo/deployments
```

---

## Directory Removal

**These top-level directories will be eliminated:**
- `operators/` → split into `clusters/global/operators/` and `clusters/{cluster-name}/operators/`
- `pipelines/` → split into `clusters/global/pipelines/` and `clusters/{cluster-name}/pipelines/`
- `deployments/` → moved to `clusters/{cluster-name}/deployments/`
- `gitops-applications/` → moved to `clusters/global/gitops/`
- `regions/` → specs moved to `clusters/{cluster-name}/{cluster-name}.yaml`
- `prereqs/` → eliminated

**Final top-level structure:**
```
bootstrap-hyperfleet/
  ├── clusters/
  ├── bases/
  └── bin/
```

---

## Resources Created by `bin/cluster-generate`

### Common Resources (All Cluster Types)

| Resource | Current Path | Proposed Path |
|----------|--------------|---------------|
| Cluster spec | `regions/{region}/{cluster-name}/region.yaml` | `clusters/{cluster-name}/{cluster-name}.yaml` |
| Namespace | `clusters/{cluster-name}/namespace.yaml` | `clusters/{cluster-name}/cluster/namespace.yaml` |
| KlusterletAddonConfig | `clusters/{cluster-name}/klusterletaddonconfig.yaml` | `clusters/{cluster-name}/cluster/klusterletaddonconfig.yaml` |
| Kustomization (cluster) | `clusters/{cluster-name}/kustomization.yaml` | `clusters/{cluster-name}/cluster/kustomization.yaml` |
| Kustomization (pipelines) | `pipelines/cloud-infrastructure-provisioning/{cluster-name}/kustomization.yaml` | `clusters/{cluster-name}/pipelines/cloud-infrastructure/kustomization.yaml` |
| PipelineRun | `pipelines/cloud-infrastructure-provisioning/{cluster-name}/cloud-infrastructure-provisioning.pipelinerun.yaml` | `clusters/{cluster-name}/pipelines/cloud-infrastructure/cloud-infrastructure-provisioning.pipelinerun.yaml` |
| Kustomization (operators) | `operators/openshift-pipelines/{cluster-name}/kustomization.yaml` | `clusters/{cluster-name}/operators/kustomization.yaml` |
| Namespace (operators) | `operators/openshift-pipelines/{cluster-name}/namespace.yaml` | `clusters/{cluster-name}/operators/namespace.yaml` |
| Kustomization (deployments) | `deployments/ocm/{cluster-name}/kustomization.yaml` | `clusters/{cluster-name}/deployments/ocm/kustomization.yaml` |
| Namespace (deployments) | `deployments/ocm/{cluster-name}/namespace.yaml` | `clusters/{cluster-name}/deployments/ocm/namespace.yaml` |
| ApplicationSet (provisioning) | `gitops-applications/clusters/{cluster-name}.yaml` (1st doc) | `clusters/{cluster-name}/gitops/provisioning.applicationset.yaml` |
| ApplicationSet (content) | `gitops-applications/clusters/{cluster-name}.yaml` (2nd doc) | `clusters/{cluster-name}/gitops/content.applicationset.yaml` |
| Root kustomization | N/A | `clusters/{cluster-name}/kustomization.yaml` |

---

### OCP-Specific Resources

| Resource | Current Path | Proposed Path |
|----------|--------------|---------------|
| install-config.yaml | `clusters/{cluster-name}/install-config.yaml` | `clusters/{cluster-name}/cluster/install-config.yaml` |
| ClusterDeployment* | `clusters/{cluster-name}/kustomization.yaml` (patched from `bases/clusters`) | `clusters/{cluster-name}/cluster/kustomization.yaml` (patched from `bases/clusters`) |
| ManagedCluster* | `clusters/{cluster-name}/kustomization.yaml` (patched from `bases/clusters`) | `clusters/{cluster-name}/cluster/kustomization.yaml` (patched from `bases/clusters`) |
| MachinePool* | `clusters/{cluster-name}/kustomization.yaml` (patched from `bases/clusters`) | `clusters/{cluster-name}/cluster/kustomization.yaml` (patched from `bases/clusters`) |
| ExternalSecret (aws-credentials)* | `clusters/{cluster-name}/kustomization.yaml` (patched from `bases/clusters`) | `clusters/{cluster-name}/cluster/kustomization.yaml` (patched from `bases/clusters`) |
| ExternalSecret (pull-secret)* | `clusters/{cluster-name}/kustomization.yaml` (patched from `bases/clusters`) | `clusters/{cluster-name}/cluster/kustomization.yaml` (patched from `bases/clusters`) |

*Referenced from `bases/clusters`, patched with cluster-specific values

---

### EKS-Specific Resources

| Resource | Current Path | Proposed Path |
|----------|--------------|---------------|
| Cluster (CAPI) | `clusters/{cluster-name}/cluster.yaml` | `clusters/{cluster-name}/cluster/cluster.yaml` |
| AWSManagedControlPlane | `clusters/{cluster-name}/awsmanagedcontrolplane.yaml` | `clusters/{cluster-name}/cluster/awsmanagedcontrolplane.yaml` |
| AWSManagedMachinePool | `clusters/{cluster-name}/awsmanagedmachinepool.yaml` | `clusters/{cluster-name}/cluster/awsmanagedmachinepool.yaml` |
| MachinePool | `clusters/{cluster-name}/machinepool.yaml` | `clusters/{cluster-name}/cluster/machinepool.yaml` |
| ManagedCluster | `clusters/{cluster-name}/managedcluster.yaml` | `clusters/{cluster-name}/cluster/managedcluster.yaml` |
| ExternalSecret (aws-credentials) | `clusters/{cluster-name}/external-secrets.yaml` (1st doc) | `clusters/{cluster-name}/cluster/external-secrets.yaml` (1st doc) |
| ExternalSecret (pull-secret) | `clusters/{cluster-name}/external-secrets.yaml` (2nd doc) | `clusters/{cluster-name}/cluster/external-secrets.yaml` (2nd doc) |
| Task (ACM integration) | `clusters/{cluster-name}/acm-integration-pipeline.yaml` (1st doc) | `clusters/{cluster-name}/cluster/acm-integration-pipeline.yaml` (1st doc) |
| PipelineRun (ACM integration) | `clusters/{cluster-name}/acm-integration-pipeline.yaml` (2nd doc) | `clusters/{cluster-name}/cluster/acm-integration-pipeline.yaml` (2nd doc) |
| ServiceAccount | `clusters/{cluster-name}/acm-integration-pipeline.yaml` (3rd doc) | `clusters/{cluster-name}/cluster/acm-integration-pipeline.yaml` (3rd doc) |
| ClusterRole | `clusters/{cluster-name}/acm-integration-pipeline.yaml` (4th doc) | `clusters/{cluster-name}/cluster/acm-integration-pipeline.yaml` (4th doc) |
| ClusterRoleBinding | `clusters/{cluster-name}/acm-integration-pipeline.yaml` (5th doc) | `clusters/{cluster-name}/cluster/acm-integration-pipeline.yaml` (5th doc) |

---

### HCP-Specific Resources

| Resource | Current Path | Proposed Path |
|----------|--------------|---------------|
| HostedCluster | `clusters/{cluster-name}/hostedcluster.yaml` | `clusters/{cluster-name}/cluster/hostedcluster.yaml` |
| NodePool | `clusters/{cluster-name}/nodepool.yaml` | `clusters/{cluster-name}/cluster/nodepool.yaml` |
| Secret (SSH key) | `clusters/{cluster-name}/ssh-key-secret.yaml` | `clusters/{cluster-name}/cluster/ssh-key-secret.yaml` |
| ExternalSecret (aws-credentials) | `clusters/{cluster-name}/external-secrets.yaml` (1st doc) | `clusters/{cluster-name}/cluster/external-secrets.yaml` (1st doc) |
| ExternalSecret (pull-secret) | `clusters/{cluster-name}/external-secrets.yaml` (2nd doc) | `clusters/{cluster-name}/cluster/external-secrets.yaml` (2nd doc) |

---

## Modified Files

| File | Current Modification | Proposed Modification |
|------|---------------------|----------------------|
| `clusters/kustomization.yaml` | Adds `- {cluster-name}/` to resources | Adds `- {cluster-name}/` to resources (unchanged) |
| `gitops-applications/clusters/kustomization.yaml` | Adds `- ./{cluster-name}.yaml` after `# ADD CLUSTERS HERE` marker | **Eliminated** - ApplicationSets moved to `clusters/{cluster-name}/gitops/` |
