# Which Path Should I Take?

**Interactive decision tree to guide your documentation journey**

## 🤔 What's Your Situation?

### New to This Project?
```mermaid
flowchart TD
    A[New to this project?] --> B{Familiar with GitOps?}
    B -->|Yes| C[Start with Quickstart]
    B -->|No| D[Start with Core Concepts]
    C --> E[Then: First Cluster]
    D --> F[Then: Quickstart]
    F --> E
    E --> G[Then: Operations Docs]
```

**Recommended Path:**
- **GitOps Familiar**: [Quickstart](./quickstart.md) → [First Cluster](./first-cluster.md) → [Operations](../operations/cluster-management.md)
- **GitOps New**: [Core Concepts](./concepts.md) → [Quickstart](./quickstart.md) → [First Cluster](./first-cluster.md)

### What Cluster Type Should I Choose?

```mermaid
flowchart TD
    A[Need a new cluster?] --> B{Existing OpenShift<br/>licensing?}
    B -->|Yes| C[Use OCP]
    B -->|No| D{Cost sensitive?}
    D -->|Yes| E[Use EKS]
    D -->|No| F{Advanced OpenShift<br/>features needed?}
    F -->|Yes| C
    F -->|No| E
    
    C --> G[ClusterDeployment<br/>via Hive]
    E --> H[AWSManagedControlPlane<br/>via CAPI]
    
    G --> I[Full OpenShift capabilities<br/>Higher cost<br/>30-45 min provision]
    H --> J[Kubernetes + some features<br/>Lower cost<br/>15-20 min provision]
```

**Decision Factors:**
- **Choose OCP if**: You need full OpenShift features, have existing licensing, or require advanced operators
- **Choose EKS if**: You want cost optimization, faster provisioning, or basic Kubernetes is sufficient

### I Have Issues - Where Should I Go?

```mermaid
flowchart TD
    A[Having problems?] --> B{What type of issue?}
    B -->|Cluster won't provision| C[Cluster Provisioning Issues]
    B -->|Applications not syncing| D[GitOps Sync Issues]
    B -->|Can't access cluster| E[Access/Auth Issues]
    B -->|Want to clean up resources| F[Cleanup Issues]
    B -->|Need to monitor health| G[Monitoring Issues]
    
    C --> C1[Check AWS credentials<br/>Check quotas<br/>Check CAPI/Hive logs]
    D --> D1[Check ArgoCD applications<br/>Force sync<br/>Check target cluster connectivity]
    E --> E1[Check aws-auth ConfigMap<br/>Verify RBAC<br/>Check kubeconfig]
    F --> F1[Use bin/clean-aws script<br/>Follow rollback procedures]
    G --> G1[Run bin/health-check<br/>Check monitoring guide]
```

**Quick Solutions:**
- **Provisioning**: [Troubleshooting Guide](../operations/troubleshooting.md#cluster-provisioning)
- **GitOps**: [Monitoring Guide](../operations/monitoring.md#troubleshooting-workflow)
- **Access**: [EKS Auth Setup](../eks-aws-auth-setup.md)
- **Cleanup**: [AWS Cleanup Guide](../../bin/clean-aws.md)

### What Role Am I In?

```mermaid
flowchart TD
    A[What's your role?] --> B{Administrator}
    A --> C{Operator}
    A --> D{Developer}
    A --> E{New User}
    
    B --> B1[Focus on:<br/>- Installation Guide<br/>- Security & Vault<br/>- Compliance & Governance]
    C --> C1[Focus on:<br/>- Cluster Management<br/>- Monitoring<br/>- Day-to-day Operations]
    D --> D1[Focus on:<br/>- Architecture<br/>- GitOps Flows<br/>- Customization]
    E --> E1[Start with:<br/>- Getting Started<br/>- Core Concepts<br/>- First Cluster]
```

**Role-Based Paths:**
- **Administrator**: [Installation](../../INSTALL.md) → [Vault Setup](../../VAULT-SETUP.md) → [Operations](../operations/cluster-management.md)
- **Operator**: [Quickstart](./quickstart.md) → [Cluster Management](../operations/cluster-management.md) → [Monitoring](../operations/monitoring.md)
- **Developer**: [Architecture](../../ARCHITECTURE.md) → [Core Concepts](./concepts.md) → [GitOps Flow](../architecture/gitops-flow.md)
- **New User**: [Core Concepts](./concepts.md) → [Quickstart](./quickstart.md) → [First Cluster](./first-cluster.md)

## 📋 Quick Checklists

### Pre-Deployment Checklist
- [ ] Hub cluster running and accessible
- [ ] AWS credentials configured (`aws sts get-caller-identity`)
- [ ] Pull secrets available (for OCP clusters)
- [ ] Required tools installed (`oc`, `kubectl`, `kustomize`)
- [ ] Understanding of cluster type choice (OCP vs EKS)

### Post-Deployment Validation
- [ ] Cluster provisioned successfully
- [ ] ACM shows cluster as "Available"
- [ ] ArgoCD applications all "Synced/Healthy"
- [ ] Services running on target cluster
- [ ] Health check shows no errors

### Troubleshooting Checklist
- [ ] Checked recent logs for errors
- [ ] Verified AWS credentials and quotas
- [ ] Confirmed network connectivity
- [ ] Reviewed ArgoCD application status
- [ ] Checked ACM ManagedCluster status
- [ ] Validated target cluster access

## 🎯 Task-Based Navigation

### "I want to..."

#### Deploy a new cluster
→ [First Cluster Guide](./first-cluster.md) or [Cluster Creation Guide](../../guides/cluster-creation.md)

#### Understand the architecture
→ [Core Concepts](./concepts.md) → [Architecture Overview](../../ARCHITECTURE.md)

#### Monitor cluster health
→ [Monitoring Guide](../operations/monitoring.md)

#### Troubleshoot issues
→ [Troubleshooting Guide](../operations/troubleshooting.md)

#### Clean up AWS resources
→ [AWS Cleanup Guide](../../bin/clean-aws.md)

#### Set up Vault integration
→ [Vault Setup Guide](../../VAULT-SETUP.md)

#### Manage day-to-day operations
→ [Cluster Management Guide](../operations/cluster-management.md)

#### Get quick command reference
→ [Command Reference](../reference/commands.md)

#### Understand GitOps workflow
→ [GitOps Flow Documentation](../architecture/gitops-flow.md)

## 🔄 Learning Progression

### Beginner → Intermediate → Advanced

**Beginner Path:**
1. [Core Concepts](./concepts.md) - Understand the basics
2. [Quickstart](./quickstart.md) - See the workflow
3. [First Cluster](./first-cluster.md) - Hands-on experience

**Intermediate Path:**
4. [Cluster Management](../operations/cluster-management.md) - Day-to-day operations
5. [Monitoring](../operations/monitoring.md) - Health checking
6. [Architecture Deep Dive](../architecture/gitops-flow.md) - Technical details

**Advanced Path:**
7. [Vault Integration](../../VAULT-SETUP.md) - Advanced security
8. [Customization](../architecture/customization.md) - Extending the system
9. [Troubleshooting Mastery](../operations/troubleshooting.md) - Expert-level problem solving

## 🚀 Quick Start Based on Time Available

### 5 minutes
→ [Quickstart Guide](./quickstart.md)

### 15 minutes  
→ [Core Concepts](./concepts.md)

### 45 minutes
→ [First Cluster Deployment](./first-cluster.md)

### 2 hours
→ Complete [Getting Started](./README.md) section

### Half day
→ Getting Started + [Operations](../operations/cluster-management.md)

### Full day
→ Complete documentation review + hands-on deployment