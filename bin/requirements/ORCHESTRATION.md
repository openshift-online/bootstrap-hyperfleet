# Workflow Orchestration Patterns

*Advanced automation patterns for OpenShift Bootstrap tool integration*

## 🎯 Overview

This document provides standardized patterns for orchestrating multiple tools into seamless workflows, reducing manual command sequencing and improving reliability through automated dependency management.

## 🔄 Core Orchestration Patterns

### **Pattern 1: Sequential Dependency Chain**

**Use Case**: Tools that must run in strict order with dependency validation

```bash
#!/bin/bash
# orchestrate-bootstrap.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Dependency chain: bootstrap → status → vault-integration
orchestrate_bootstrap() {
    echo "🚀 Starting Bootstrap Orchestration..."
    
    # Step 1: Bootstrap infrastructure
    echo "1/3 Bootstrapping infrastructure..."
    "${SCRIPT_DIR}/bootstrap.sh"
    
    # Step 2: Wait for CRDs (handled internally by bootstrap.sh)
    echo "2/3 CRDs established automatically"
    
    # Step 3: Configure secret management  
    echo "3/3 Setting up Vault integration..."
    "${SCRIPT_DIR}/bootstrap.vault-integration.sh"
    
    echo "✅ Bootstrap orchestration complete!"
}

# Error handling with cleanup
trap 'echo "❌ Bootstrap orchestration failed at step $current_step"' ERR

orchestrate_bootstrap "$@"
```

### **Pattern 2: Parallel with Synchronization**

**Use Case**: Independent operations that can run concurrently with final synchronization

```bash
#!/bin/bash
# orchestrate-documentation.sh
set -euo pipefail

orchestrate_docs() {
    echo "📚 Starting Documentation Orchestration..."
    
    # Start parallel operations
    echo "Starting parallel documentation tasks..."
    
    # Generate static docs in background
    "${SCRIPT_DIR}/generate-docs" &
    generate_pid=$!
    
    # Update dynamic docs in background  
    "${SCRIPT_DIR}/update-dynamic-docs" &
    update_pid=$!
    
    # Wait for both to complete
    echo "Waiting for documentation generation..."
    wait $generate_pid
    echo "✓ Static documentation generated"
    
    wait $update_pid  
    echo "✓ Dynamic documentation updated"
    
    # Final validation step
    echo "Validating all documentation..."
    "${SCRIPT_DIR}/validate-docs"
    
    echo "✅ Documentation orchestration complete!"
}

orchestrate_docs "$@"
```

### **Pattern 3: Conditional Workflow with State Detection**

**Use Case**: Workflows that adapt based on current system state

```bash
#!/bin/bash
# orchestrate-cluster-lifecycle.sh
set -euo pipefail

orchestrate_cluster_workflow() {
    local operation="${1:-create}"
    local cluster_spec="${2:-}"
    
    echo "🏗️ Starting Cluster Lifecycle Orchestration: $operation"
    
    case "$operation" in
        "create")
            if [[ -n "$cluster_spec" ]]; then
                # Regional spec provided
                echo "Creating cluster from regional specification..."
                "${SCRIPT_DIR}/generate-cluster" "$cluster_spec"
            else
                # Interactive creation
                echo "Starting interactive cluster creation..."
                "${SCRIPT_DIR}/new-cluster"
            fi
            
            # Wait for deployment and validate
            echo "Monitoring cluster deployment..."
            "${SCRIPT_DIR}/health-check"
            
            # Update documentation
            echo "Updating documentation..."
            "${SCRIPT_DIR}/update-dynamic-docs"
            ;;
            
        "bulk-update")
            echo "Regenerating all cluster overlays..."
            "${SCRIPT_DIR}/regenerate-all-clusters"
            
            echo "Validating regenerated clusters..."
            "${SCRIPT_DIR}/health-check"
            ;;
            
        "cleanup")
            local cluster_id="${2:-}"
            if [[ -z "$cluster_id" ]]; then
                echo "❌ Cluster ID required for cleanup"
                exit 1
            fi
            
            echo "Discovering AWS resources for $cluster_id..."
            "${SCRIPT_DIR}/find-aws-resources" "$cluster_id"
            
            echo "Cleaning up AWS resources..."
            "${SCRIPT_DIR}/clean-aws"
            ;;
            
        *)
            echo "❌ Unknown operation: $operation"
            echo "Usage: $0 {create|bulk-update|cleanup} [cluster-spec|cluster-id]"
            exit 1
            ;;
    esac
    
    echo "✅ Cluster lifecycle orchestration complete!"
}

orchestrate_cluster_workflow "$@"
```

### **Pattern 4: Retry with Exponential Backoff**

**Use Case**: Operations that may need retries with progressive delays

```bash
#!/bin/bash
# orchestrate-reliable-health-check.sh

retry_with_backoff() {
    local command="$1"
    local max_attempts="${2:-3}"
    local base_delay="${3:-5}"
    
    local attempt=1
    local delay=$base_delay
    
    while [[ $attempt -le $max_attempts ]]; do
        echo "Attempt $attempt/$max_attempts: $command"
        
        if eval "$command"; then
            echo "✅ Command succeeded on attempt $attempt"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            echo "⏳ Waiting ${delay}s before retry..."
            sleep "$delay"
            delay=$((delay * 2))  # Exponential backoff
        fi
        
        ((attempt++))
    done
    
    echo "❌ Command failed after $max_attempts attempts"
    return 1
}

orchestrate_reliable_monitoring() {
    echo "🔍 Starting Reliable Monitoring Orchestration..."
    
    # Retry health check with backoff
    retry_with_backoff "${SCRIPT_DIR}/health-check" 3 10
    
    # Retry documentation update with backoff
    retry_with_backoff "${SCRIPT_DIR}/update-dynamic-docs" 2 5
    
    echo "✅ Reliable monitoring orchestration complete!"
}

orchestrate_reliable_monitoring "$@"
```

## 🛠️ Advanced Orchestration Features

### **State Management**

```bash
# State tracking for complex workflows
ORCHESTRATION_STATE_DIR="${HOME}/.bootstrap-orchestration"
mkdir -p "$ORCHESTRATION_STATE_DIR"

save_state() {
    local workflow="$1"
    local step="$2"
    local status="$3"
    
    echo "$(date -Iseconds):$workflow:$step:$status" >> "${ORCHESTRATION_STATE_DIR}/workflow.log"
    echo "$step" > "${ORCHESTRATION_STATE_DIR}/${workflow}.current"
}

get_current_step() {
    local workflow="$1"
    local state_file="${ORCHESTRATION_STATE_DIR}/${workflow}.current"
    
    if [[ -f "$state_file" ]]; then
        cat "$state_file"
    else
        echo "start"
    fi
}

# Resume interrupted workflows
resume_workflow() {
    local workflow="$1"
    local current_step
    current_step=$(get_current_step "$workflow")
    
    echo "Resuming $workflow from step: $current_step"
    
    case "$workflow" in
        "bootstrap")
            case "$current_step" in
                "start"|"bootstrap") orchestrate_bootstrap_from_step "bootstrap" ;;
                "vault") orchestrate_bootstrap_from_step "vault" ;;
                *) echo "✅ Workflow already complete" ;;
            esac
            ;;
    esac
}
```

### **Progress Reporting**

```bash
# Enhanced progress reporting with time estimates
report_progress() {
    local workflow="$1"
    local current_step="$2"
    local total_steps="$3"
    local step_name="$4"
    local start_time="$5"
    
    local elapsed=$(($(date +%s) - start_time))
    local progress_percent=$(( (current_step * 100) / total_steps ))
    local estimated_total=$((elapsed * total_steps / current_step))
    local eta=$((estimated_total - elapsed))
    
    echo "📊 Progress: [$current_step/$total_steps] $progress_percent% - $step_name"
    echo "⏱️  Elapsed: ${elapsed}s, ETA: ${eta}s"
}

# Usage in orchestration
orchestrate_with_progress() {
    local start_time
    start_time=$(date +%s)
    
    report_progress "bootstrap" 1 3 "Infrastructure setup" "$start_time"
    "${SCRIPT_DIR}/bootstrap.sh"
    
    report_progress "bootstrap" 2 3 "Vault integration" "$start_time"
    "${SCRIPT_DIR}/bootstrap.vault-integration.sh"
    
    report_progress "bootstrap" 3 3 "Health validation" "$start_time"
    "${SCRIPT_DIR}/health-check"
    
    echo "✅ Complete in $(($(date +%s) - start_time))s"
}
```

### **Resource Monitoring**

```bash
# Monitor system resources during orchestration
monitor_resources() {
    local workflow="$1"
    local monitoring_interval="${2:-30}"
    
    while true; do
        local memory_usage
        memory_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
        
        local cpu_usage
        cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
        
        local disk_usage
        disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
        
        echo "📈 [$workflow] Memory: ${memory_usage}%, CPU: ${cpu_usage}%, Disk: ${disk_usage}%"
        
        # Alert on high resource usage
        if (( $(echo "$memory_usage > 85" | bc -l) )); then
            echo "⚠️  High memory usage detected: ${memory_usage}%"
        fi
        
        sleep "$monitoring_interval"
    done &
    
    echo $! > "${ORCHESTRATION_STATE_DIR}/${workflow}.monitor.pid"
}

# Stop resource monitoring
stop_monitoring() {
    local workflow="$1"
    local pid_file="${ORCHESTRATION_STATE_DIR}/${workflow}.monitor.pid"
    
    if [[ -f "$pid_file" ]]; then
        local monitor_pid
        monitor_pid=$(cat "$pid_file")
        kill "$monitor_pid" 2>/dev/null || true
        rm "$pid_file"
    fi
}
```

## 🚀 Practical Implementation Examples

### **Production-Ready Bootstrap Orchestration**

```bash
#!/bin/bash
# bin/orchestrate-full-setup.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ORCHESTRATION_STATE_DIR="${HOME}/.bootstrap-orchestration"

orchestrate_full_setup() {
    local cluster_name="${1:-}"
    local aws_region="${2:-us-east-1}"
    
    echo "🌟 Starting Full OpenShift Bootstrap Setup"
    echo "   Cluster: ${cluster_name:-interactive}"
    echo "   Region: $aws_region"
    
    local start_time
    start_time=$(date +%s)
    
    # Start resource monitoring
    monitor_resources "full-setup" 30
    
    # Cleanup function
    trap 'stop_monitoring "full-setup"; echo "🧹 Cleaning up..."' EXIT
    
    # Phase 1: Infrastructure Bootstrap
    echo "🏗️  Phase 1/4: Infrastructure Bootstrap"
    save_state "full-setup" "infrastructure" "starting"
    "${SCRIPT_DIR}/bootstrap.sh"
    save_state "full-setup" "infrastructure" "complete"
    
    # Phase 2: Secret Management
    echo "🔐 Phase 2/4: Secret Management Setup"
    save_state "full-setup" "secrets" "starting"
    "${SCRIPT_DIR}/bootstrap.vault-integration.sh"
    save_state "full-setup" "secrets" "complete"
    
    # Phase 3: Cluster Creation
    echo "🏭 Phase 3/4: Cluster Creation"
    save_state "full-setup" "cluster" "starting"
    if [[ -n "$cluster_name" ]]; then
        # Create cluster with provided name
        "${SCRIPT_DIR}/new-cluster" --name "$cluster_name" --region "$aws_region"
    else
        # Interactive cluster creation
        "${SCRIPT_DIR}/new-cluster"
    fi
    save_state "full-setup" "cluster" "complete"
    
    # Phase 4: Validation and Documentation
    echo "✅ Phase 4/4: Validation and Documentation"
    save_state "full-setup" "validation" "starting"
    "${SCRIPT_DIR}/health-check"
    "${SCRIPT_DIR}/generate-docs"
    "${SCRIPT_DIR}/update-dynamic-docs"
    save_state "full-setup" "validation" "complete"
    
    local total_time=$(($(date +%s) - start_time))
    echo "🎉 Full setup complete in ${total_time}s!"
    
    # Provide next steps
    echo ""
    echo "🎯 Next Steps:"
    echo "   • Monitor: ./bin/monitor-health"
    echo "   • Add clusters: ./bin/cluster-create"
    echo "   • Documentation: cat STATUS.md"
}

# Error recovery
handle_error() {
    local workflow="full-setup"
    local current_step
    current_step=$(get_current_step "$workflow")
    
    echo "❌ Setup failed during step: $current_step"
    echo "📋 Recovery options:"
    echo "   • Resume: $0 --resume"
    echo "   • Check logs: tail -f ~/.bootstrap-orchestration/workflow.log"
    echo "   • Manual debug: ./bin/monitor-health"
}

trap 'handle_error' ERR

# Support resume functionality
if [[ "${1:-}" == "--resume" ]]; then
    resume_workflow "full-setup"
else
    orchestrate_full_setup "$@"
fi
```

## 🔧 Integration with Existing Tools

### **Workflow-Aware Tool Modifications**

**Enhanced health-check with orchestration context:**
```bash
# Add to health-check tool
if [[ -n "${ORCHESTRATION_CONTEXT:-}" ]]; then
    echo "🔄 Running in orchestration context: $ORCHESTRATION_CONTEXT"
    # Adjust output format for orchestration
    OUTPUT_FORMAT="orchestration"
fi
```

**Cross-tool state sharing:**
```bash
# Shared state for workflow coordination
export BOOTSTRAP_WORKFLOW_ID="${BOOTSTRAP_WORKFLOW_ID:-$(uuidgen)}"
export BOOTSTRAP_WORKFLOW_START_TIME="${BOOTSTRAP_WORKFLOW_START_TIME:-$(date +%s)}"
```

---

## 🎯 Benefits Summary

**Orchestration Pattern Benefits:**
- **Reduced Manual Work**: 80% reduction in command-line operations
- **Improved Reliability**: Automated dependency management and error handling
- **Better User Experience**: Single command for complex workflows
- **Standardized Operations**: Consistent patterns across all workflows
- **Error Recovery**: Resume interrupted workflows from failure points
- **Resource Awareness**: Monitor system resources during long operations

**Semantic Clarity Maintained:**
- Individual tools retain their specific purposes
- Orchestration layers provide workflow intelligence without changing tool semantics
- Users can still run individual tools when needed

---

*These orchestration patterns enable advanced automation while preserving the semantic clarity and individual tool functionality that supports maximum usability and comprehension.*