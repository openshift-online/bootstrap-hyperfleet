# Regional Cluster plugin implementation strategy - AWS

**Plugin/Component Name:** Regional Cluster Management

**Description:** Strategy for deploying and managing regional OpenShift Container Platform (OCP) clusters on AWS infrastructure

**Requirements:** 
- Automated cluster provisioning across multiple AWS regions
- Integration with AWS services (Route53, IAM, VPC)
- Support for OCP 4.x deployment
- GitOps-based cluster configuration management
- Multi-cluster monitoring and observability

**Evaluation Date:** 2024-01-15

**Evaluator(s):** Platform Architecture Team

**Platform Context:** AWS-specific (OCP on AWS)

---

## Step 1: Reuse of Existing Code
**Priority: High**

### Available Options:
- OSD Fleet Manager regional cluster deployment patterns
- Application Interface OSD deployments

### Assessment:
- **Suitability:** High - OSD Fleet Manager has proven patterns for AWS OCP deployment
- **Modification Required:** Medium - Is it even possible to provision a cluster with this tech stack that isn't OCM based (chicken/egg)
- **Maintenance Considerations:** Red Hat maintains OSD Fleet Manager, but we want to move more toward the tech stack we sell
- **Support and Roadmap:** Active development, no clear outcome with a bespoke solution

### Decision:
☐ **Selected** - Leverage OSD Fleet Manager patterns with MCE Hive for cluster deployment
☐ **Not Suitable**

---

## Step 2: Red Hat Products or Promoted Open Source Alternatives
**Priority: High**

### Available Options:
- MCE (Multi-Cluster Engine) with Hive
- ArgoCD for GitOps deployment
- Red Hat Advanced Cluster Management (RHACM)
- OCP Pipelines

### Assessment:
- **Product/Project Name:** MCE Hive + ArgoCD
- **Alignment with Requirements:** Excellent - MCE Hive handles AWS OCP deployment, ArgoCD manages GitOps
- **Integration Complexity:** Low - Both are well-integrated with AWS
- **Licensing Considerations:** Red Hat products
- **Support and Roadmap:** Strong Red Hat support, active development

### Decision:
☑ **Selected** - MCE Hive for cluster deployment, ArgoCD for GitOps management
☐ **Not Suitable**

---

## Step 3: Upstream or Community Options
**Priority: Medium**

### Available Options:
- Cluster API (CAPI) with AWS provider
- Terraform-based cluster provisioning
- Crossplane for cloud resource management

### Assessment:
- **Project Name:** Cluster API (CAPI)
- **Maturity Level:** High - Well-established, widely adopted
- **Community Health:** Excellent - Strong community, active development
- **Technical Fit:** LOW - Can handle AWS cluster deployment, but OpenShift HCP is not natively supported on EKS
- **Strategic Alignment:** Aligns with how ARO-HCP was implemented
- **Risk Assessment:** High - Requires Management Clusters to support EKS

### Decision:
☑ **Selected** - This is an eventual long term goal
☑ **Not Suitable** - While CAPI is excellent, MCE Hive provides better OCP-specific integration.

---

## Step 4: Net New Build
**Priority: Lowest (Last Resort)**

### Justification:
Not required - MCE Hive and ArgoCD provide all needed functionality

### Decision:
☐ **Selected**
☑ **Not Suitable** - Existing solutions meet all requirements

---

## Final Decision Summary

**Selected Option:** Step 2 - MCE Hive + ArgoCD

**Final Rationale:** 
MCE Hive provides proven AWS OCP deployment capabilities with excellent Red Hat support. ArgoCD handles GitOps-based configuration management. This combination leverages existing Red Hat technologies while providing the automation and management capabilities required for regional cluster deployment.

**Next Steps:**
1. Set up MCE Hive infrastructure for AWS
2. Configure ArgoCD for cluster configuration management
3. Create GitOps repositories for cluster configurations
4. Implement monitoring and observability stack
5. Document deployment procedures

**Stakeholder Approval:** Platform Architecture Team - Pending

**Cross-Platform Impact:** This decision is AWS-specific. Other implementations will need similar but platform-specific solutions (GKE for GCP, OCI for Oracle).

---

**Template Version:** 0.1  
**Last Updated:** See Git