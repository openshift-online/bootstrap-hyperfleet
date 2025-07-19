# Documentation Index

**Quick reference to all documentation with clear purpose statements**

## 🚀 **Getting Started**

| Document | Purpose | Audience |
|----------|---------|----------|
| [README.md](../README.md) | **Main entry point** - Project overview with navigation patterns | All users |
| [getting-started/README.md](./getting-started/README.md) | **Progressive learning paths** - Guided documentation journey | New users |
| [getting-started/decision-tree.md](./getting-started/decision-tree.md) | **Interactive guidance** - Choose your path with decision trees | All users |
| [getting-started/quickstart.md](./getting-started/quickstart.md) | **5-minute overview** - Understand the big picture quickly | New users |
| [getting-started/concepts.md](./getting-started/concepts.md) | **Core concepts** - Architecture, GitOps, and cluster types | New users |
| [getting-started/first-cluster.md](./getting-started/first-cluster.md) | **Hands-on deployment** - Complete walkthrough | New users |
| [INSTALL.md](../INSTALL.md) | **Complete installation guide** - Hub setup + adding regions | Administrators |

## 🏗️ **Architecture & Technical**

| Document | Purpose | Audience |
|----------|---------|----------|
| [ARCHITECTURE.md](../ARCHITECTURE.md) | **Visual architecture** - Diagrams with GitOps sync wave flows | Developers |
| [KUSTOMIZATION.md](../KUSTOMIZATION.md) | **Configuration management** - Kustomize patterns and structure | Developers |
| [eks-aws-auth-setup.md](./eks-aws-auth-setup.md) | **EKS authentication** - aws-auth ConfigMap setup procedures | Operators |

## 🔧 **Operations & Tools**

| Document | Purpose | Audience |
|----------|---------|----------|
| [operations/cluster-management.md](./operations/cluster-management.md) | **Day-to-day operations** - Managing existing clusters, scaling, upgrades | Operators |
| [guides/monitoring.md](../guides/monitoring.md) | **Complete monitoring guide** - Status checking, health monitoring, troubleshooting | Operators |
| [bin/new-cluster.md](../bin/new-cluster.md) | **Interactive cluster generator** - Tool documentation | Operators |
| [bin/clean-aws.md](../bin/clean-aws.md) | **AWS cleanup procedures** - Resource cleanup automation | Administrators |

## 🔐 **Security & Integration**

| Document | Purpose | Audience |
|----------|---------|----------|
| [VAULT-SETUP.md](../VAULT-SETUP.md) | **Vault + ESO integration** - Complete secret management guide | Administrators |
| [REGIONALSPEC.md](../REGIONALSPEC.md) | **Regional specifications** - Cluster configuration details | Developers |

## 🔄 **Utilities & Conversion**

| Document | Purpose | Audience |
|----------|---------|----------|
| [CONVERTER.md](../CONVERTER.md) | **Conversion utilities** - Migration and transformation tools | Developers |

## 📋 **Component Documentation**

### Advanced Cluster Management (ACM)
| Document | Purpose |
|----------|---------|
| [operators/advanced-cluster-management/global/README.md](../operators/advanced-cluster-management/global/README.md) | ACM deployment and configuration |
| [operators/advanced-cluster-management/global/INFO.md](../operators/advanced-cluster-management/global/INFO.md) | ACM component information |

### OpenShift Pipelines
| Document | Purpose |
|----------|---------|
| [operators/openshift-pipelines/global/README.md](../operators/openshift-pipelines/global/README.md) | Pipelines operator documentation |
| [operators/openshift-pipelines/global/components/README.md](../operators/openshift-pipelines/global/components/README.md) | Pipeline components overview |
| [operators/openshift-pipelines/global/components/enable-console-plugin/README.md](../operators/openshift-pipelines/global/components/enable-console-plugin/README.md) | Console plugin enablement |

## 🎯 **Architectural Decisions**

| Document | Purpose |
|----------|---------|
| [plugin-decisions/README.md](./plugin-decisions/README.md) | Architectural decision records overview |
| [plugin-decisions/template.md](./plugin-decisions/template.md) | ADR template for new decisions |
| [plugin-decisions/rosa-hcp/README.md](./plugin-decisions/rosa-hcp/README.md) | ROSA HCP decisions |
| [plugin-decisions/rosa-hcp/regional-clusters.md](./plugin-decisions/rosa-hcp/regional-clusters.md) | Regional cluster strategy |

## 🤖 **Claude Code Integration**

| Document | Purpose | Audience |
|----------|---------|----------|
| [CLAUDE.md](../CLAUDE.md) | **Project overview** - Focused guidance for Claude Code | Claude Code |
| [CLAUDE-FULL.md](../CLAUDE-FULL.md) | **Complete context** - Full project details and session history | Claude Code |
| [prompts/docs.md](../prompts/docs.md) | **Documentation improvement plan** - This implementation guide | Developers |

---

## 📍 Navigation Tips

### New Users
1. **Start here**: [Getting Started Overview](./getting-started/README.md)
2. **Choose your path**: [Interactive Decision Tree](./getting-started/decision-tree.md)
3. **Quick understanding**: [5-minute Quickstart](./getting-started/quickstart.md)
4. **Deep learning**: [Core Concepts](./getting-started/concepts.md)
5. **Hands-on**: [Deploy First Cluster](./getting-started/first-cluster.md)

### Experienced Users
- **Quick setup**: [Installation Guide](../INSTALL.md)
- **Daily operations**: [Cluster Management](./operations/cluster-management.md)
- **Health monitoring**: [Monitoring Guide](../guides/monitoring.md)
- **Command reference**: [Quick Commands](./reference/commands.md)

## 🔗 External References

- **OpenShift GitOps**: [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- **Advanced Cluster Management**: [ACM Documentation](https://access.redhat.com/documentation/en-us/red_hat_advanced_cluster_management_for_kubernetes/)
- **Cluster API**: [CAPI Documentation](https://cluster-api.sigs.k8s.io/)
- **Kustomize**: [Kustomize Documentation](https://kustomize.io/)

---

*Last updated: Auto-generated from docs/INDEX.md*