# ROSA-HCP plugin Implementation Decisions

This directory contains plugin implementation decisions specific to AWS platform implementations.

## Context

The AWS platform focuses on:
- **OpenShift Hosted Control Planes (HCP)** as the primary Kubernetes distribution
- **AWS services integration** (Route53, IAM, VPC, etc.)
- **Red Hat ecosystem** tools and patterns
- **Multi-cluster management** using OpenShift GitOps, Pipelines & Advanced Cluster Management

## Key Decisions

### Regional Clusters
- **File:** `regional-clusters.md`
- **Decision:** TBD

### Managment Clusters
- **File:** `management-clusters.md`
- **Decision:** TBD

### Relational Database Service
- **File:** `rds.md`
- **Decision:** TBD

## Contributing

When adding new AWS-specific decisions:
1. Use the template from `../template.md`
2. Focus on AWS-specific (ROSA-HCP) considerations and constraints
3. Reference related cross-platform decisions where applicable
4. Consider how the decision might differ from ARO-HCP, GCP-HCP or possible Bare Metal implementations 