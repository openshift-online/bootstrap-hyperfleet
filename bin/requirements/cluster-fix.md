# bin/cluster-fix Requirements

## Purpose

The `cluster-fix` tool provides an interactive remediation workflow that discovers cluster issues using `bin/cluster-status` and guides users through fixing each problem with actionable commands and confirmation prompts.

## Functional Requirements

### Primary Objective

Create an automated cluster issue resolution system that:
1. **Discovers Issues**: Uses `bin/cluster-status --format json` to identify cluster problems
2. **Interactive Remediation**: Prompts user for each issue with specific fix commands
3. **Selective Resolution**: Allows users to accept/decline each fix individually
4. **Safety Validation**: Confirms each action before execution
5. **Progress Tracking**: Shows completion status throughout the process

### Issue Detection Integration

#### Cluster Status Analysis
```bash
# Retrieve structured issue data
cluster_issues=$(./bin/cluster-status --format json | jq '.clusters[] | select(.issues != "OK")')
```

#### Issue Categories Handled
- **ORPHANED_MC**: ManagedCluster exists without repository configuration
- **MISSING_MC**: Repository configuration exists without ManagedCluster
- **STUCK_NS**: Namespace stuck in Terminating phase
- **STUCK_FINALIZERS**: ManagedCluster has finalizers but unavailable
- **TAINTED**: ManagedCluster has taints applied

### Interactive Remediation Workflow

#### Issue Presentation Format
```bash
# Example prompt for each issue
echo "🔧 ISSUE FOUND: STUCK_FINALIZERS"
echo "   Cluster: ocp-01-mturansk-t10"
echo "   Problem: ManagedCluster has finalizers but is unavailable"
echo ""
echo "   Recommended Fix:"
echo "   oc patch managedcluster ocp-01-mturansk-t10 --type=merge -p '{\"metadata\":{\"finalizers\":[]}}'"
echo ""
echo "   This will remove finalizers to allow cluster cleanup to proceed."
echo ""
read -p "Apply this fix? [y/N]: " response
```

#### User Interaction Logic
```bash
case "$response" in
    [yY]|[yY][eE][sS])
        echo "✅ Applying fix..."
        execute_fix_command
        verify_fix_result
        ;;
    *)
        echo "⏭️  Skipped - moving to next issue"
        ;;
esac
```

## Fix Command Templates

### ORPHANED_MC Remediation
```bash
# Issue: ManagedCluster exists without repository configuration
fix_command="oc delete managedcluster $cluster_name"
description="Remove orphaned ManagedCluster that has no corresponding repository configuration"
warning="This will permanently remove the cluster from ACM management"
```

### MISSING_MC Remediation
```bash
# Issue: Repository configuration exists without ManagedCluster
fix_command="./bin/cluster-generate $cluster_name && oc apply -k clusters/$cluster_name/"
description="Generate and apply cluster manifests from repository configuration"
warning="This will create new cluster resources - ensure cluster is intended to exist"
```

### STUCK_NS Remediation
```bash
# Issue: Namespace stuck in Terminating phase
fix_command="kubectl patch namespace $cluster_name --type=merge -p '{\"metadata\":{\"finalizers\":[]}}'"
description="Remove finalizers from stuck namespace to allow termination"
warning="This will force namespace deletion - ensure no important resources remain"
```

### STUCK_FINALIZERS Remediation
```bash
# Issue: ManagedCluster has finalizers but unavailable
fix_command="oc patch managedcluster $cluster_name --type=merge -p '{\"metadata\":{\"finalizers\":[]}}'"
description="Remove finalizers from unavailable ManagedCluster"
warning="This will allow cluster cleanup to proceed but may leave orphaned resources"
```

### TAINTED Remediation
```bash
# Issue: ManagedCluster has taints applied
fix_command="oc patch managedcluster $cluster_name --type=json -p='[{\"op\": \"remove\", \"path\": \"/spec/taints\"}]'"
description="Remove taints from ManagedCluster to restore normal operations"
warning="Only remove taints if cluster health issues have been resolved"
```

## Command Line Interface

### Basic Usage
```bash
./bin/cluster-fix                    # Interactive mode for all issues
```

### Optional Flags
```bash
./bin/cluster-fix --cluster NAME     # Fix issues for specific cluster only
./bin/cluster-fix --issue-type TYPE  # Fix specific issue type only
./bin/cluster-fix --auto-yes         # Auto-accept all fixes (dangerous)
./bin/cluster-fix --dry-run          # Show fixes without executing
./bin/cluster-fix --verbose          # Detailed output and explanations
```

### Examples
```bash
# Fix all cluster issues interactively
./bin/cluster-fix

# Fix only stuck finalizer issues
./bin/cluster-fix --issue-type STUCK_FINALIZERS

# Preview all fixes without executing
./bin/cluster-fix --dry-run

# Fix issues for specific cluster
./bin/cluster-fix --cluster ocp-01-mturansk-t10
```

## Safety Features

### Pre-execution Validation
```bash
# Verify cluster-status tool availability
if ! command -v "./bin/cluster-status" >/dev/null 2>&1; then
    echo "❌ Error: cluster-status tool not found"
    echo "   Ensure bin/cluster-status exists and is executable"
    exit 1
fi

# Verify OpenShift connectivity
if ! oc whoami >/dev/null 2>&1; then
    echo "❌ Error: Not connected to OpenShift cluster"
    echo "   Run: oc login <cluster-url>"
    exit 1
fi
```

### Command Verification
```bash
# Before executing each fix command
echo "📋 About to execute:"
echo "   $fix_command"
echo ""
echo "⚠️  Warning: $warning"
echo ""
read -p "Confirm execution? [y/N]: " confirm
```

### Results Validation
```bash
# After each fix, verify resolution
post_fix_status=$(./bin/cluster-status --cluster "$cluster_name" --format json)
if [[ $(echo "$post_fix_status" | jq -r '.clusters[0].issues') == "OK" ]]; then
    echo "✅ Fix successful - issue resolved"
else
    echo "⚠️  Fix completed but issue may persist"
fi
```

## Error Handling

### Fix Command Failures
```bash
if ! eval "$fix_command"; then
    echo "❌ Fix command failed"
    echo "   Command: $fix_command"
    echo "   Check cluster connectivity and permissions"
    read -p "Continue with remaining issues? [y/N]: " continue_response
    [[ "$continue_response" =~ ^[yY] ]] || exit 1
fi
```

### Cluster Status Failures
```bash
# Handle cluster-status tool errors gracefully
if ! cluster_data=$(./bin/cluster-status --format json 2>/dev/null); then
    echo "❌ Unable to retrieve cluster status"
    echo "   Ensure cluster-status tool is working properly"
    exit 1
fi
```

### No Issues Found
```bash
if [[ $(echo "$cluster_data" | jq '.clusters[] | select(.issues != "OK")' | wc -l) -eq 0 ]]; then
    echo "✅ No cluster issues found - all clusters healthy"
    exit 0
fi
```

## Progress and Reporting

### Session Summary
```bash
# Track fixes applied during session
echo ""
echo "📊 SESSION SUMMARY"
echo "=================="
echo "Issues Found: $total_issues"
echo "Fixes Applied: $fixes_applied"
echo "Fixes Skipped: $fixes_skipped"
echo "Fixes Failed: $fixes_failed"
echo ""

# Show remaining issues
remaining_issues=$(./bin/cluster-status --issues-only --format json | jq '.clusters[].issues' | grep -v "OK" | wc -l)
if [[ $remaining_issues -gt 0 ]]; then
    echo "⚠️  $remaining_issues issues remain - run cluster-fix again if needed"
else
    echo "✅ All issues resolved!"
fi
```

### Verbose Mode Output
```bash
if [[ "$verbose" == "true" ]]; then
    echo "🔍 ISSUE ANALYSIS:"
    echo "   Type: $issue_type"
    echo "   Cluster: $cluster_name"  
    echo "   Root Cause: $root_cause_explanation"
    echo "   Fix Strategy: $fix_strategy_explanation"
    echo ""
fi
```

## Integration Requirements

### Dependency Tools
- **cluster-status**: Must be available and executable
- **OpenShift CLI (oc)**: Required for cluster operations
- **jq**: JSON processing for cluster-status output

### Workflow Integration
```bash
# Common workflow after cluster operations
./bin/cluster-status --issues-only  # Discover problems
./bin/cluster-fix                    # Interactively resolve issues
./bin/monitor-health                 # Verify overall cluster health
```

### Automation Support
```bash
# CI/CD pipeline integration
./bin/cluster-fix --auto-yes --verbose > cluster-fix-report.log
exit_code=$?
if [[ $exit_code -ne 0 ]]; then
    echo "Cluster fix process encountered errors - manual intervention required"
fi
```

## Security Requirements

### Read-Only Verification
- Never execute destructive commands without explicit user confirmation
- Show complete command before execution
- Provide clear warnings about potential impact

### Permissions Validation
```bash
# Verify required permissions before proceeding
if ! oc auth can-i delete managedclusters 2>/dev/null; then
    echo "⚠️  Warning: Insufficient permissions for some fix operations"
    echo "   Some fixes may fail due to RBAC restrictions"
fi
```

## Related Tools

### Prerequisites
- **[cluster-status.md](./cluster-status.md)** - Issue discovery engine
- **OpenShift CLI** - Cluster resource management

### Complementary Tools
- **[monitor-health.md](./monitor-health.md)** - Post-fix health verification
- **[cluster-remove.md](./cluster-remove.md)** - Clean removal of problematic clusters

### Workflow Tools
- **[cluster-generate.md](./cluster-generate.md)** - Recreate missing cluster configurations
- **[aws-find-resources.md](./aws-find-resources.md)** - Find orphaned AWS resources

## Performance Requirements

### Execution Time
- Interactive prompts should not timeout
- Each fix command should complete within 60 seconds
- Overall session should handle 50+ issues efficiently

### Resource Usage
- Minimal memory footprint (shell script with JSON processing)
- No persistent state between executions
- Safe for concurrent execution on different clusters

## Future Enhancements

### Batch Operations
- Fix multiple issues of same type simultaneously
- Cluster-specific fix sessions
- Automated fix scheduling

### Advanced Features
- Fix command customization
- Issue priority sorting
- Fix rollback capabilities
- Integration with monitoring systems