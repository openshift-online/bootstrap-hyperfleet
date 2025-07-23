# Lead Agent Multi-Cluster Testing Plan

## Overview

This plan orchestrates testing of all three cluster types (EKS, OCP, HCP) using a lead agent architecture. One lead agent spawns and coordinates three specialized subagents, each responsible for testing a specific cluster provider while maintaining regular status communication.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Lead Agent                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Command & Control                          │   │
│  │  • Spawns subagents                                     │   │
│  │  • Receives status reports                              │   │
│  │  • Available for user instructions                     │   │
│  │  • Coordinates testing phases                          │   │
│  │  • Aggregates results                                   │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   EKS Subagent  │ │   OCP Subagent  │ │   HCP Subagent  │
│                 │ │                 │ │                 │
│ Tests:          │ │ Tests:          │ │ Tests:          │
│ • CAPI/CAPA     │ │ • Hive Operator │ │ • HyperShift    │
│ • ACM Integration│ │ • ACM Integration│ │ • NodePools     │
│ • Pipeline ACM  │ │ • OpenShift     │ │ • Control Plane │
│ • Worker Nodes  │ │ • Full Install  │ │ • Hosted Arch   │
│                 │ │                 │ │                 │
│ Reports every   │ │ Reports every   │ │ Reports every   │
│ 5 minutes       │ │ 10 minutes      │ │ 5 minutes       │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

## Lead Agent Responsibilities

### Core Functions
1. **Subagent Management**
   - Spawn three specialized testing subagents
   - Monitor subagent health and communication
   - Restart failed subagents automatically
   - Aggregate status reports from all subagents

2. **User Interface**
   - Remain available for user commands during testing
   - Provide real-time status of all testing activities
   - Accept new instructions while tests are running
   - Handle emergency stop commands

3. **Coordination & Reporting**
   - Synchronize testing phases across subagents
   - Generate consolidated status reports
   - Detect and report cross-cluster issues
   - Maintain testing timeline and progress tracking

### Status Collection
```yaml
# Status Report Format (received from subagents every 5-10 minutes)
agent_id: "eks-subagent-001"
cluster_type: "eks"
timestamp: "2025-01-20T15:30:00Z"
phase: "infrastructure-provisioning"
status: "in-progress"
progress: 60
details:
  cluster_name: "eks-test-20250120-153000"
  region: "us-west-2"
  current_step: "waiting-for-worker-nodes"
  estimated_completion: "2025-01-20T15:45:00Z"
issues: []
logs_location: ".test/eks-test-logs-20250120-153000"
```

## EKS Subagent Specification

### Test Plan: `test_plan_eks.md`
### Unique ID: `eks-subagent-{timestamp}`
### Reporting Interval: Every 5 minutes

**Test Phases:**
1. **Prerequisites** (5 min)
   - Verify AWS quotas and credentials
   - Check CAPI controllers
   - Validate External Secrets

2. **Cluster Generation** (2 min)
   - Run `bin/cluster-generate` with test specification
   - Validate generated manifests
   - Verify resource structure

3. **Infrastructure Provisioning** (15-25 min)
   - Monitor CAPI cluster creation
   - Track AWS EKS cluster status
   - Verify worker node provisioning

4. **ACM Integration** (5-10 min)
   - Execute Tekton Pipeline for automatic integration
   - Monitor Klusterlet CRD installation
   - Verify pull secret configuration

5. **Validation & Cleanup** (5 min)
   - Test cluster connectivity
   - Verify ArgoCD ApplicationSet deployment
   - Clean up test cluster

**Status Reporting:**
- Report cluster provisioning progress
- Monitor AWS resource creation
- Track ACM integration pipeline status
- Alert on quota or permission issues

## OCP Subagent Specification

### Test Plan: `test_plan_ocp.md`
### Unique ID: `ocp-subagent-{timestamp}`
### Reporting Interval: Every 10 minutes (longer due to install time)

**Test Phases:**
1. **Prerequisites** (5 min)
   - Verify Hive operator status
   - Check cloud provider quotas
   - Validate install-config template

2. **Cluster Generation** (2 min)
   - Generate ClusterDeployment and supporting resources
   - Validate OpenShift version compatibility
   - Check install-config.yaml structure

3. **OpenShift Installation** (45-60 min)
   - Monitor Hive ClusterDeployment progress
   - Track installation phases (bootstrap, control plane, workers)
   - Monitor cluster operator deployment

4. **ACM Integration** (5-10 min)
   - Verify automatic ACM import
   - Check Klusterlet agent deployment
   - Validate addon configuration

5. **Validation & Cleanup** (10 min)
   - Test OpenShift console access
   - Verify cluster operator health
   - Clean up test cluster

**Status Reporting:**
- Report installation phase progress
- Monitor Hive operator logs
- Track cloud resource provisioning
- Alert on installation failures

## HCP Subagent Specification

### Test Plan: `test_plan_hcp.md`
### Unique ID: `hcp-subagent-{timestamp}`
### Reporting Interval: Every 5 minutes

**Test Phases:**
1. **Prerequisites** (5 min)
   - Verify HyperShift operator status
   - Check management cluster capacity
   - Validate platform configuration

2. **Cluster Generation** (2 min)
   - Generate HostedCluster and NodePool manifests
   - Validate HyperShift configuration
   - Check resource requirements

3. **Control Plane Provisioning** (5-10 min)
   - Monitor HostedCluster deployment
   - Track control plane pod startup
   - Verify API server accessibility

4. **Worker Node Provisioning** (10-15 min)
   - Monitor NodePool scaling
   - Track worker node registration
   - Verify cluster networking

5. **Validation & Cleanup** (5 min)
   - Test hosted cluster functionality
   - Verify pod scheduling
   - Clean up test cluster

**Status Reporting:**
- Report control plane provisioning progress
- Monitor worker node scaling
- Track resource utilization on management cluster
- Alert on capacity or networking issues

## Subagent Communication Protocol

### Status Report Structure
```json
{
  "agent_id": "string",
  "cluster_type": "eks|ocp|hcp",
  "timestamp": "ISO-8601",
  "phase": "string",
  "status": "pending|in-progress|completed|failed|blocked",
  "progress": "0-100",
  "cluster_details": {
    "name": "string",
    "region": "string",
    "phase": "string",
    "estimated_completion": "ISO-8601"
  },
  "current_step": "string",
  "issues": [
    {
      "severity": "warning|error|critical",
      "message": "string",
      "timestamp": "ISO-8601",
      "remediation": "string"
    }
  ],
  "metrics": {
    "duration_minutes": "number",
    "resources_created": "number",
    "tests_passed": "number",
    "tests_failed": "number"
  },
  "logs": {
    "location": "string",
    "key_entries": ["string"]
  }
}
```

### Communication Methods
1. **File-based Status**: Write status to `./.test/agent-status-{agent_id}.json`
2. **Heartbeat**: Update timestamp every minute to indicate agent health
3. **Emergency Signals**: Use signal files for immediate attention requests

## Lead Agent Implementation

### Subagent Spawning
```bash
# Lead agent spawns each subagent as background Task
spawn_subagent() {
    local cluster_type=$1
    local test_plan="test_plan_${cluster_type}.md"
    local agent_id="${cluster_type}-subagent-$(date +%Y%m%d-%H%M%S)"
    
    # Launch subagent with unique task instance
    Task description="Test ${cluster_type} cluster" prompt="
    You are a specialized ${cluster_type} cluster testing subagent.
    Your agent ID is: ${agent_id}
    
    Execute the test plan in ${test_plan} with these requirements:
    1. Report status every $(get_reporting_interval ${cluster_type}) minutes
    2. Write status to .test/agent-status-${agent_id}.json
    3. Update heartbeat file .test/heartbeat-${agent_id} every minute
    4. Include detailed error information in status reports
    5. Clean up resources on completion or failure
    6. Log all activities to .test/${agent_id}-test.log
    
    Test cluster naming: ${cluster_type}-test-$(date +%Y%m%d-%H%M%S)
    
    Begin testing immediately and report initial status.
    "
}
```

### Status Monitoring Loop
```bash
# Lead agent monitoring loop
monitor_subagents() {
    while [[ $testing_active == true ]]; do
        for agent_id in "${active_agents[@]}"; do
            # Check heartbeat
            if [[ ! -f ".test/heartbeat-${agent_id}" ]] || 
               [[ $(($(date +%s) - $(stat -c %Y ".test/heartbeat-${agent_id}"))) -gt 120 ]]; then
                alert_unresponsive_agent "$agent_id"
            fi
            
            # Collect status
            if [[ -f ".test/agent-status-${agent_id}.json" ]]; then
                process_status_report "$agent_id"
            fi
        done
        
        # Generate consolidated report
        generate_lead_agent_report
        
        sleep 30  # Check every 30 seconds
    done
}
```

## Testing Execution Plan

### Phase 1: Initialization (Lead Agent)
1. Validate hub cluster prerequisites
2. Check all testing tools and dependencies
3. Verify access to test plan files
4. Initialize status tracking systems

### Phase 2: Subagent Deployment
1. Spawn EKS subagent with `test_plan_eks.md`
2. Spawn OCP subagent with `test_plan_ocp.md`
3. Spawn HCP subagent with `test_plan_hcp.md`
4. Verify all subagents initialize successfully

### Phase 3: Concurrent Testing
1. All three subagents execute their test plans simultaneously
2. Lead agent monitors status from all subagents
3. Lead agent remains available for user commands
4. Real-time status updates provided to user

### Phase 4: Results Aggregation
1. Collect final status from all subagents
2. Generate comprehensive test report
3. Clean up any remaining test resources
4. Provide recommendations for improvements

## Emergency Procedures

### Subagent Failure Recovery
- **Unresponsive Agent**: Restart with same test parameters
- **Test Failure**: Analyze logs and attempt remediation
- **Resource Conflicts**: Coordinate cleanup and retry

### User Override Commands
- **Stop All**: Immediate halt of all testing
- **Stop Agent**: Halt specific subagent type
- **Status**: Immediate status report from all agents
- **Restart**: Restart failed or stuck agents

### Resource Management
- **Quota Monitoring**: Track AWS/cloud resource usage
- **Cleanup Automation**: Automatic cleanup of failed tests
- **Conflict Resolution**: Handle resource naming conflicts

## Success Criteria

### Individual Subagent Success
- **EKS**: Cluster provisioned, ACM integrated, ArgoCD applications deployed
- **OCP**: Full OpenShift installation completed, all operators available
- **HCP**: Control plane running, worker nodes joined, workloads schedulable

### Overall Test Success
- All three cluster types provision successfully
- ACM integration works for all cluster types
- GitOps applications deploy correctly
- All test resources cleaned up properly
- Comprehensive test report generated

## Expected Timeline

| Phase | Duration | Parallel Activities |
|-------|----------|-------------------|
| Lead Agent Setup | 2 min | Validation and initialization |
| Subagent Spawn | 1 min | All three agents start |
| EKS Testing | 30-45 min | CAPI provisioning + ACM integration |
| OCP Testing | 60-75 min | Hive installation + validation |
| HCP Testing | 20-30 min | HyperShift provisioning + NodePools |
| Results & Cleanup | 5 min | Aggregation and cleanup |

**Total Estimated Time**: 75-90 minutes (limited by OCP installation time)

## Output Deliverables

1. **Real-time Status Dashboard**: Live updates from all subagents
2. **Detailed Test Logs**: Complete logs from each cluster type test
3. **Comprehensive Report**: Success/failure status, issues found, recommendations
4. **Updated Test Plans**: Improvements based on testing results
5. **Resource Cleanup Verification**: Confirmation all test resources removed

This lead agent architecture ensures comprehensive testing of all cluster types while maintaining user availability and providing detailed status tracking throughout the entire testing process.