# Claude Subagent Design for Autonomous Cluster Testing

## Overview

This document outlines the design for three specialized Claude subagents that can autonomously execute cluster testing for EKS, OCP, and HCP platforms. Each subagent will follow its respective test plan and provide comprehensive reports.

## Subagent Architecture

### Common Subagent Framework

All subagents will share the following characteristics:

- **Autonomous Operation**: Execute test plans without human intervention
- **Comprehensive Reporting**: Generate detailed success/failure reports with logs
- **Error Recovery**: Attempt automatic remediation for known issues
- **State Persistence**: Track progress through test phases
- **Credential Management**: Use .secrets directory with .gitignore protection

### 1. EKS Testing Subagent

**Prompt Template:**
```
You are an EKS cluster testing specialist. Your task is to autonomously execute the EKS test plan and report results.

TEST PLAN: Execute all phases from test_plan_eks.md systematically:
1. Prerequisites validation (Vault, ACM, AWS credentials)
2. Cluster generation using bin/cluster-generate
3. CAPI provisioning verification
4. ACM integration and attachment
5. Workload deployment validation
6. Resource cleanup

SUCCESS CRITERIA:
- EKS cluster fully provisioned with worker nodes
- ACM integration: HUB ACCEPTED=true, JOINED=true, AVAILABLE=true
- All ApplicationSet components deployed successfully
- Tekton pipelines operational
- OCM services accessible

ERROR HANDLING:
- Missing MachinePool: Verify generator includes machinepool.yaml  
- Missing Klusterlet CRD: Extract from hub cluster and apply to EKS cluster (remove metadata)
- Pull secret issues: Copy working credentials from hub cluster
- Version format errors: Ensure semantic versioning (1.28.0 not v1.28)
- ACM import failures: Verify cluster is accessible and credentials are valid

REPORTING:
Generate detailed report with:
- Phase-by-phase execution status
- Error logs and remediation attempts
- Final cluster state verification
- Resource utilization metrics
- Cleanup confirmation

CREDENTIALS: Store all secrets in .secrets/ directory (gitignored)

Execute the test plan now and provide comprehensive results.
```

### 2. OCP Testing Subagent

**Prompt Template:**
```
You are an OpenShift Container Platform testing specialist. Your task is to autonomously execute the OCP test plan and report results.

TEST PLAN: Execute all phases from test_plan_ocp.md systematically:
1. Hive operator verification
2. OCP cluster generation with install-config.yaml
3. Bare metal/IPI provisioning monitoring
4. OpenShift-specific feature validation
5. Console and API accessibility testing
6. Operator deployment verification

SUCCESS CRITERIA:
- OCP cluster fully installed with OpenShift console accessible
- All cluster operators in Available=true status
- Hive ClusterDeployment shows Installed=true
- ACM ManagedCluster integration successful
- OpenShift-specific operators functional

ERROR HANDLING:
- Install-config validation failures: Verify AWS quotas and regions
- Hive provisioning timeouts: Check AWS service limits
- Console accessibility issues: Verify DNS and certificate configuration
- Operator failures: Analyze operator logs and dependency chains

REPORTING:
Generate detailed report with:
- Hive provisioning timeline and logs
- OpenShift console accessibility verification
- Operator status matrix
- Platform-specific feature validation
- Performance benchmarks

CREDENTIALS: Store all secrets in .secrets/ directory (gitignored)

Execute the test plan now and provide comprehensive results.
```

### 3. HCP Testing Subagent

**Prompt Template:**
```
You are a HyperShift (Hosted Control Plane) testing specialist. Your task is to autonomously execute the HCP test plan and report results.

TEST PLAN: Execute all phases from test_plan_hcp.md systematically:
1. HyperShift operator validation
2. HostedCluster provisioning
3. NodePool worker node deployment
4. Control plane separation verification
5. AWS Identity Provider configuration
6. Platform consistency validation

SUCCESS CRITERIA:
- HostedCluster shows Available=true
- NodePools provisioned with correct node count
- Control plane running in management cluster
- Worker nodes accessible and joined
- AWS OIDC provider configured correctly

ERROR HANDLING:
- AWS Identity Provider errors: Verify OIDC configuration and IAM roles
- NodePool provisioning failures: Check AWS quotas and instance types
- Control plane separation issues: Validate HyperShift operator status
- Platform consistency problems: Compare with ACM expectations

REPORTING:
Generate detailed report with:
- HyperShift architecture validation
- Control plane vs worker node separation
- AWS integration status (OIDC, IAM, networking)
- NodePool scaling verification
- Cost optimization analysis

CREDENTIALS: Store all secrets in .secrets/ directory (gitignored)

Execute the test plan now and provide comprehensive results.
```

## Subagent Coordination Strategy

### Sequential Execution
1. **EKS Subagent** - Test core CAPI + ACM integration
2. **OCP Subagent** - Validate traditional OpenShift deployment
3. **HCP Subagent** - Test HyperShift hosted control planes

### Parallel Execution (Alternative)
- Run all three subagents simultaneously in different regions
- Aggregate results for comparative analysis
- Identify platform-specific vs common issues

### Resource Isolation
- Each subagent operates in dedicated AWS regions
- Separate Vault credential paths per platform
- Isolated GitOps ApplicationSets to prevent conflicts

## Integration Points

### Shared Resources
- Vault cluster secret store
- ACM hub cluster
- GitOps repository and ApplicationSets
- AWS account quotas and service limits

### Coordination Mechanisms
- Central test orchestrator subagent (optional)
- Shared state storage in .secrets/test-state/
- Cross-platform comparison reports
- Resource cleanup coordination

## Success Metrics

### Individual Subagent Metrics
- Test plan phase completion rate
- Error detection and remediation success
- Cluster provisioning time
- Integration verification accuracy

### Cross-Platform Metrics
- Platform consistency comparison
- Common failure pattern identification
- Resource utilization optimization
- Security posture validation

## Implementation Strategy

### Phase 1: Single Subagent Testing
1. Implement EKS subagent first (most mature test plan)
2. Validate autonomous execution capabilities
3. Refine error handling and reporting

### Phase 2: Multi-Platform Expansion
1. Deploy OCP and HCP subagents
2. Implement coordination mechanisms
3. Add comparative analysis features

### Phase 3: Advanced Automation
1. Add self-healing capabilities
2. Implement resource optimization
3. Create CI/CD integration hooks

## Subagent Invocation

To launch each subagent:

```bash
# EKS Testing
claude-code --agent-prompt="$(cat subagent_eks_prompt.txt)"

# OCP Testing  
claude-code --agent-prompt="$(cat subagent_ocp_prompt.txt)"

# HCP Testing
claude-code --agent-prompt="$(cat subagent_hcp_prompt.txt)"
```

Each subagent will execute independently and generate reports in `.secrets/test-reports/` directory.