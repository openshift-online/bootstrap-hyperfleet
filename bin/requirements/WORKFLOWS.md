# Common End-to-End Workflows

*Comprehensive scenarios for OpenShift Bootstrap tool usage*

## 🚀 Complete Environment Setup

### **Scenario**: Setting up a new multi-cluster OpenShift environment from scratch

**Prerequisites**:
- OpenShift cluster with cluster-admin access
- AWS credentials configured
- Vault secrets populated

**Workflow**:
```bash
# 1. Initialize GitOps infrastructure
./bin/bootstrap.sh

# 2. Wait for core components (automatically handled by bootstrap)
# - ArgoCD GitOps operator deployment
# - ACM MultiClusterHub installation
# - External Secrets Operator setup

# 3. Configure Vault-based secret management
./bin/bootstrap.vault-integration.sh

# 4. Create first managed cluster
./bin/cluster-create

# 5. Verify deployment
./bin/monitor-health

# 6. Generate documentation
./bin/generate-docs
```

**Expected Timeline**:
- Bootstrap: 10-15 minutes
- EKS cluster provisioning: ~15 minutes
- OCP cluster provisioning: ~45 minutes
- HCP cluster provisioning: ~10 minutes

---

## 🏗️ Adding New Clusters

### **Scenario**: Adding managed clusters to existing infrastructure

**Prerequisites**:
- Existing bootstrap environment
- Regional specifications or interactive creation

**Option A: Interactive Creation**
```bash
# 1. Create cluster interactively
./bin/cluster-create

# 2. Monitor deployment
./bin/monitor-health

# 3. Update documentation
./bin/update-dynamic-docs
```

**Option B: From Regional Specification**
```bash
# 1. Generate cluster overlay from regional spec
./bin/cluster-generate regions/us-west-2/ocp-03/

# 2. Commit and push changes (triggers GitOps deployment)
git add clusters/ocp-03/
git commit -m "Add ocp-03 cluster"
git push

# 3. Monitor deployment
./bin/monitor-health

# 4. Update live documentation
./bin/update-dynamic-docs
```

**Option C: Converting Existing Cluster**
```bash
# 1. Convert existing cluster to regional specification
./bin/convert-cluster clusters/legacy-cluster-01 > regions/us-east-1/ocp-04/region.yaml

# 2. Generate new semantic overlay
./bin/cluster-generate regions/us-east-1/ocp-04/

# 3. Validate and deploy
git add regions/us-east-1/ocp-04/ clusters/ocp-04/
./bin/monitor-health
```

---

## 🔧 Maintenance and Operations

### **Scenario**: Regular operational tasks

**Daily Operations**
```bash
# 1. Check overall environment health
./bin/monitor-health

# 2. Update dynamic documentation with latest status
./bin/update-dynamic-docs

# 3. Validate documentation quality
./bin/validate-docs
```

**Bulk Cluster Updates**
```bash
# 1. Regenerate all cluster overlays from regional specs
./bin/regenerate-all-clusters

# 2. Review changes
git diff

# 3. Commit and deploy
git add clusters/
git commit -m "Regenerate cluster overlays"
git push

# 4. Monitor deployment
./bin/monitor-health
```

**Troubleshooting Workflow**
```bash
# 1. Check overall status
./bin/monitor-health

# 2. Wait for specific CRDs if needed
./bin/status.sh applications.argoproj.io

# 3. Wait for specific resources
./bin/wait.kube.sh deployment argocd-server openshift-gitops '{.status.readyReplicas}' "1"

# 4. Check component-specific logs
oc logs -n openshift-gitops deployment/argocd-server
```

---

## 🧹 Resource Cleanup

### **Scenario**: Decommissioning clusters and cleaning up AWS resources

**Planned Cluster Removal**
```bash
# 1. Discover AWS resources for specific cluster
./bin/find-aws-resources my-test-cluster

# 2. Review resources to be deleted
# (script output shows all discovered resources)

# 3. Clean up AWS resources
./bin/clean-aws

# 4. Remove cluster from GitOps
git rm -r clusters/my-test-cluster/
git rm -r regions/us-west-2/my-test-cluster/
git commit -m "Remove my-test-cluster"
git push

# 5. Update documentation
./bin/monitor-health
./bin/update-dynamic-docs
```

**Emergency Cleanup**
```bash
# 1. Test AWS discovery tool
./bin/test-find-aws-resources

# 2. Find all resources for cluster
./bin/find-aws-resources problematic-cluster

# 3. Automated cleanup
./bin/clean-aws --disable-prompts

# 4. Verify cleanup
./bin/find-aws-resources problematic-cluster
```

---

## 📚 Documentation Workflows

### **Scenario**: Maintaining comprehensive documentation

**Regular Documentation Updates**
```bash
# 1. Generate static documentation
./bin/generate-docs

# 2. Update dynamic content (STATUS.md, inventories)
./bin/update-dynamic-docs

# 3. Validate all documentation
./bin/validate-docs

# 4. Commit changes
git add docs/ STATUS.md
git commit -m "Update documentation"
git push
```

**Documentation Quality Assurance**
```bash
# 1. Full documentation validation
./bin/validate-docs

# 2. Fix any reported issues
# (Address broken links, syntax errors, etc.)

# 3. Re-validate
./bin/validate-docs

# 4. Generate fresh documentation
./bin/generate-docs
./bin/update-dynamic-docs
```

---

## 🔄 Migration Scenarios

### **Scenario**: Migrating from legacy cluster naming to semantic naming

**Legacy to Semantic Migration**
```bash
# 1. Convert existing cluster-01 to semantic naming
./bin/convert-cluster clusters/cluster-01 > regions/us-east-1/ocp-primary/region.yaml

# 2. Generate new semantic overlay
./bin/cluster-generate regions/us-east-1/ocp-primary/

# 3. Compare outputs to ensure consistency
kubectl kustomize clusters/cluster-01/ > /tmp/legacy.yaml
kubectl kustomize clusters/ocp-primary/ > /tmp/semantic.yaml
diff /tmp/legacy.yaml /tmp/semantic.yaml

# 4. Deploy new semantic cluster
git add regions/us-east-1/ocp-primary/ clusters/ocp-primary/
git commit -m "Add semantic ocp-primary cluster"
git push

# 5. Verify deployment
./bin/monitor-health

# 6. Remove legacy cluster
git rm -r clusters/cluster-01/
git commit -m "Remove legacy cluster-01"
git push
```

**Bulk Legacy Migration**
```bash
# 1. Convert all existing clusters
for cluster in clusters/cluster-*; do
  cluster_name=$(basename "$cluster")
  ./bin/convert-cluster "$cluster" > "regions/converted/${cluster_name}/region.yaml"
done

# 2. Generate semantic overlays
for region_spec in regions/converted/*/region.yaml; do
  ./bin/cluster-generate "$(dirname "$region_spec")"
done

# 3. Validate all conversions
./bin/regenerate-all-clusters

# 4. Comprehensive health check
./bin/monitor-health
```

---

## ⚡ Advanced Automation

### **Scenario**: Streamlined workflows for power users

**One-Command Environment Setup**
```bash
# Custom orchestration script
#!/bin/bash
set -euo pipefail

echo "🚀 Setting up complete OpenShift Bootstrap environment..."

./bin/bootstrap.sh
./bin/bootstrap.vault-integration.sh
./bin/cluster-create
./bin/monitor-health
./bin/generate-docs

echo "✅ Environment setup complete!"
```

**Continuous Integration Workflow**
```bash
# CI/CD pipeline integration
#!/bin/bash
set -euo pipefail

# Validate all documentation
./bin/validate-docs

# Regenerate all cluster overlays
./bin/regenerate-all-clusters

# Update dynamic documentation
./bin/update-dynamic-docs

# Final health check
./bin/monitor-health

echo "✅ CI validation complete"
```

---

## 🎯 Workflow Selection Guide

### **Choose Your Workflow**

| **Goal** | **Primary Tool** | **Typical Workflow** |
|----------|------------------|---------------------|
| **First-time setup** | `bootstrap` | Complete Environment Setup |
| **Add clusters** | `new-cluster` | Adding New Clusters |
| **Monitor health** | `health-check` | Maintenance and Operations |
| **Clean resources** | `clean-aws` | Resource Cleanup |
| **Update docs** | `generate-docs` | Documentation Workflows |
| **Migrate clusters** | `convert-cluster` | Migration Scenarios |

### **By Complexity**

- **Beginner**: Complete Environment Setup → Adding New Clusters → Daily Operations
- **Intermediate**: Bulk operations → Migration scenarios → Documentation workflows  
- **Advanced**: Custom automation → CI/CD integration → Multi-environment management

---

*This workflow documentation supports the maximum usability principle by providing clear, actionable scenarios for all common OpenShift Bootstrap operations.*