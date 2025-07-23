# bin/find-all-aws-resources Requirements

## Purpose

Create a comprehensive AWS resource inventory tool that automatically discovers and catalogs **ALL AWS resources** in target regions, including both cluster-managed resources and orphaned/untracked resources. Cross-references with this OpenShift Bootstrap repository to identify resources that may need cleanup or management.

## Functional Requirements

### Primary Objective

Automate the discovery of AWS resources for all clusters defined in the repository by:
1. **Auto-discovering clusters** from `regions/` and `clusters/` directories
2. **Running `find-aws-resources`** for each cluster and region combination
3. **Discovering ALL resources** in target regions (orphan detection mode)
4. **Cross-referencing** to identify untracked/orphaned resources
5. **Aggregating results** into comprehensive tables showing resource usage patterns
6. **Providing cost analysis** and resource utilization insights

### Repository Analysis

Based on current repository structure:

**Active Clusters:**
- `regions/us-west-2/ocp-01-mturansk-test/` - OCP cluster in us-west-2
- `clusters/ocp-01-mturansk-test/` - Deployed OCP cluster configuration

**Discovery Strategy:**
1. **Scan regions/** directory for cluster specifications  
2. **Scan clusters/** directory for deployed clusters
3. **Extract cluster names and regions** from configurations
4. **Run discovery** for each cluster-region combination

### Core Functionality

#### 1. Cluster Discovery
```bash
# Auto-discover all clusters from repository structure
find_repository_clusters() {
    # Scan regions/ directory for regional specifications
    for region_dir in regions/*/; do
        region_name=$(basename "$(dirname "$region_dir")")
        for cluster_dir in "$region_dir"*/; do
            cluster_name=$(basename "$cluster_dir")
            echo "$cluster_name:$region_name:regional_spec"
        done
    done
    
    # Scan clusters/ directory for deployed clusters  
    for cluster_dir in clusters/*/; do
        cluster_name=$(basename "$cluster_dir")
        # Extract region from cluster configuration
        region=$(grep -r "region:" "$cluster_dir" | head -1 | awk '{print $2}')
        echo "$cluster_name:${region:-unknown}:deployed"
    done
}
```

#### 2. Resource Discovery Execution
```bash
# Run find-aws-resources for each discovered cluster
discover_all_cluster_resources() {
    local cluster_list="$1"
    local output_format="$2"  # table|csv|json
    
    echo "=== Multi-Cluster AWS Resource Discovery ==="
    echo "Discovered clusters:"
    echo "$cluster_list"
    echo ""
    
    while IFS=':' read -r cluster region status; do
        echo "Processing: $cluster in $region ($status)"
        
        # Run find-aws-resources for this cluster
        if ./bin/find-aws-resources "$cluster" "$region" > "/tmp/resources-$cluster-$region.txt" 2>&1; then
            echo "✅ Resources discovered for $cluster"
        else
            echo "❌ Failed to discover resources for $cluster"
        fi
    done <<< "$cluster_list"
}
```

#### 3. Aggregated Table Generation

**Master Resource Table:**
```bash
# Generate comprehensive resource summary table
generate_master_table() {
    echo "| Cluster | Region | Status | EC2 Instances | EBS Volumes | Load Balancers | VPCs | Subnets | Total Resources |"
    echo "|---------|--------|--------|---------------|-------------|----------------|------|---------|-----------------|"
    
    for resource_file in /tmp/resources-*.txt; do
        cluster=$(echo "$resource_file" | cut -d'-' -f2)
        region=$(echo "$resource_file" | cut -d'-' -f3 | cut -d'.' -f1)
        
        # Parse resource counts from discovery output
        ec2_count=$(grep -c "EC2 Instance:" "$resource_file" || echo "0")
        ebs_count=$(grep -c "EBS Volume:" "$resource_file" || echo "0") 
        elb_count=$(grep -c "Load Balancer:" "$resource_file" || echo "0")
        vpc_count=$(grep -c "VPC:" "$resource_file" || echo "0")
        subnet_count=$(grep -c "Subnet:" "$resource_file" || echo "0")
        total=$((ec2_count + ebs_count + elb_count + vpc_count + subnet_count))
        
        # Determine cluster status
        if [[ -f "clusters/$cluster/kustomization.yaml" ]]; then
            status="Deployed"
        elif [[ -d "regions/$region/$cluster" ]]; then
            status="Spec Only"
        else
            status="Unknown"
        fi
        
        echo "| $cluster | $region | $status | $ec2_count | $ebs_count | $elb_count | $vpc_count | $subnet_count | $total |"
    done
}
```

**Detailed Resource Breakdown:**
```bash
# Generate detailed resource type breakdown
generate_detailed_breakdown() {
    echo ""
    echo "=== Detailed Resource Breakdown by Cluster ==="
    echo ""
    
    for resource_file in /tmp/resources-*.txt; do
        cluster=$(echo "$resource_file" | cut -d'-' -f2)
        region=$(echo "$resource_file" | cut -d'-' -f3 | cut -d'.' -f1)
        
        echo "## $cluster ($region)"
        echo ""
        
        # EC2 Instances
        echo "### EC2 Instances"
        echo "| Instance ID | Type | State | VPC ID | Private IP |"
        echo "|-------------|------|-------|--------|------------|"
        grep "EC2 Instance:" "$resource_file" | while read -r line; do
            # Parse instance details and format as table row
            instance_id=$(echo "$line" | grep -o 'i-[a-zA-Z0-9]*')
            instance_type=$(echo "$line" | grep -o 'm[0-9]*\.[a-zA-Z]*\|c[0-9]*\.[a-zA-Z]*')
            echo "| $instance_id | $instance_type | running | vpc-xxx | 10.x.x.x |"
        done
        echo ""
        
        # EBS Volumes
        echo "### EBS Volumes"  
        echo "| Volume ID | Size (GB) | Type | State | Attached To |"
        echo "|-----------|-----------|------|-------|-------------|"
        grep "EBS Volume:" "$resource_file" | while read -r line; do
            volume_id=$(echo "$line" | grep -o 'vol-[a-zA-Z0-9]*')
            echo "| $volume_id | 100 | gp3 | in-use | i-xxx |"
        done
        echo ""
        
        # Continue for other resource types...
    done
}
```

#### 4. Cost Analysis Integration

```bash
# Generate cost estimates based on resource inventory
generate_cost_analysis() {
    echo ""
    echo "=== Estimated Monthly Costs by Cluster ==="
    echo ""
    echo "| Cluster | Region | EC2 Cost | EBS Cost | Load Balancer Cost | Total Est. Cost |"
    echo "|---------|--------|----------|----------|-------------------|-----------------|"
    
    # Use AWS Pricing API or predefined cost estimates
    local ec2_hourly_cost=0.20  # Average cost per hour for m5.xlarge
    local ebs_monthly_cost=0.10  # Cost per GB per month for gp3
    local elb_monthly_cost=18.00  # Cost per ALB per month
    
    for resource_file in /tmp/resources-*.txt; do
        cluster=$(echo "$resource_file" | cut -d'-' -f2)
        region=$(echo "$resource_file" | cut -d'-' -f3 | cut -d'.' -f1)
        
        ec2_count=$(grep -c "EC2 Instance:" "$resource_file" || echo "0")
        ebs_count=$(grep -c "EBS Volume:" "$resource_file" || echo "0")
        elb_count=$(grep -c "Load Balancer:" "$resource_file" || echo "0")
        
        ec2_cost=$(echo "$ec2_count * $ec2_hourly_cost * 24 * 30" | bc -l)
        ebs_cost=$(echo "$ebs_count * 100 * $ebs_monthly_cost" | bc -l)  # Assume 100GB per volume
        elb_cost=$(echo "$elb_count * $elb_monthly_cost" | bc -l)
        
        total_cost=$(echo "$ec2_cost + $ebs_cost + $elb_cost" | bc -l)
        
        printf "| %s | %s | \$%.2f | \$%.2f | \$%.2f | \$%.2f |\n" \
            "$cluster" "$region" "$ec2_cost" "$ebs_cost" "$elb_cost" "$total_cost"
    done
}
```

### Usage Patterns

#### Basic Discovery
```bash
# Discover all resources for all clusters
./bin/find-all-aws-resources

# Output to file
./bin/find-all-aws-resources > cluster-resource-inventory.md

# Generate CSV format
./bin/find-all-aws-resources --format csv > resources.csv
```

#### Advanced Options
```bash
# Include cost analysis
./bin/find-all-aws-resources --include-costs

# Filter by cluster pattern
./bin/find-all-aws-resources --cluster-pattern "ocp-*"

# Filter by region
./bin/find-all-aws-resources --region us-west-2

# Debug mode with detailed output
./bin/find-all-aws-resources --debug

# Refresh cache (re-run discovery)
./bin/find-all-aws-resources --refresh
```

#### Integration with Existing Tools
```bash
# Compare with current resources vs. expected from configs
./bin/find-all-aws-resources --compare-with-configs

# Generate cleanup recommendations  
./bin/find-all-aws-resources --recommend-cleanup

# Export for cost optimization analysis
./bin/find-all-aws-resources --format json | jq '.clusters[] | select(.estimated_cost > 100)'
```

### Output Format Examples

#### Master Summary Table
```
=== Multi-Cluster AWS Resource Inventory ===
Generated: 2025-07-20 15:30:00 UTC

| Cluster | Region | Status | EC2 Instances | EBS Volumes | Load Balancers | VPCs | Subnets | Total Resources |
|---------|--------|--------|---------------|-------------|----------------|------|---------|-----------------|
| ocp-01-mturansk-test | us-west-2 | Deployed | 3 | 6 | 2 | 1 | 6 | 18 |

Total Clusters: 1
Total Resources: 18
Total Estimated Monthly Cost: $456.78
```

#### Regional Distribution
```
=== Resource Distribution by Region ===

us-west-2:
  - Clusters: 1
  - Total Resources: 18
  - Estimated Cost: $456.78
```

#### Resource Type Summary
```
=== Resource Type Summary Across All Clusters ===

| Resource Type | Count | Percentage | Est. Monthly Cost |
|---------------|-------|------------|-------------------|
| EC2 Instances | 3 | 16.7% | $432.00 |
| EBS Volumes | 6 | 33.3% | $60.00 |
| Load Balancers | 2 | 11.1% | $36.00 |
| VPCs | 1 | 5.6% | $0.00 |
| Subnets | 6 | 33.3% | $0.00 |
```

### Error Handling Requirements

#### Cluster Discovery Failures
```bash
# Handle missing cluster configurations
if [[ ! -d "regions/" && ! -d "clusters/" ]]; then
    echo "❌ Error: No cluster configurations found"
    echo "   Expected: regions/ or clusters/ directories"
    exit 1
fi

# Handle AWS credential issues
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ Error: AWS credentials not configured"
    echo "   Run: aws configure"
    exit 1
fi
```

#### Resource Discovery Failures
```bash
# Handle find-aws-resources failures
if ! ./bin/find-aws-resources "$cluster" "$region" >/dev/null 2>&1; then
    echo "⚠️  Warning: Failed to discover resources for $cluster in $region"
    echo "   Possible causes:"
    echo "   - Cluster not deployed to AWS"
    echo "   - AWS credentials lack required permissions"
    echo "   - Region not accessible"
    continue  # Continue with other clusters
fi
```

#### Output Generation Failures
```bash
# Handle table generation issues
if [[ ! -s "/tmp/resources-$cluster-$region.txt" ]]; then
    echo "⚠️  Warning: No resources found for $cluster in $region"
    echo "| $cluster | $region | No Resources | 0 | 0 | 0 | 0 | 0 | 0 |"
fi
```

### Integration Requirements

#### Dependencies
```bash
# Required tools
command -v aws >/dev/null 2>&1 || { echo "aws CLI required"; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq required for JSON parsing"; exit 1; }
command -v bc >/dev/null 2>&1 || { echo "bc required for cost calculations"; exit 1; }

# Required permissions
# - EC2: Describe* permissions for all resource types
# - ELB: Describe* permissions  
# - IAM: List* permissions (optional, for role discovery)
# - CloudFormation: Describe* permissions (optional)
```

#### File System Requirements
```bash
# Temporary file space for resource discovery results
TEMP_DIR="/tmp/find-all-aws-resources-$$"
mkdir -p "$TEMP_DIR"
trap "rm -rf $TEMP_DIR" EXIT

# Optional cache directory for repeated runs
CACHE_DIR="${HOME}/.cache/bootstrap-aws-resources"
mkdir -p "$CACHE_DIR"
```

### Orphan Resource Discovery

#### Core Concept
Orphan resources are AWS resources that exist in target regions but are not tracked by any cluster configuration in this repository. They may be:
- Resources from deleted clusters that weren't properly cleaned up
- Manually created resources outside GitOps workflows
- Failed cluster deployments that left infrastructure behind
- Shared infrastructure not tagged with cluster identifiers
- Resources created by previous tools or processes

#### Implementation Strategy
```bash
#!/bin/bash
# Orphan Resource Discovery Implementation

discover_all_resources() {
    local region="$1"
    local output_file=".tmp/aws_resources.md"
    
    echo "=== Comprehensive AWS Resource Discovery ===" >> "$output_file"
    echo "Region: $region" >> "$output_file"
    echo "Generated: $(date)" >> "$output_file"
    echo "" >> "$output_file"
    
    # Get ALL resources by type (no filtering)
    discover_ec2_instances "$region" >> "$output_file"
    discover_ebs_volumes "$region" >> "$output_file"
    discover_load_balancers "$region" >> "$output_file"
    discover_vpcs_subnets "$region" >> "$output_file"
    discover_security_groups "$region" >> "$output_file"
    discover_iam_resources >> "$output_file"
    
    # Cross-reference with repository
    cross_reference_resources "$output_file"
}

cross_reference_resources() {
    local output_file="$1"
    
    echo "" >> "$output_file"
    echo "## Cross-Reference Analysis" >> "$output_file"
    echo "" >> "$output_file"
    
    # Get known clusters from repository
    echo "### Known Clusters (from repository):" >> "$output_file"
    find regions/ clusters/ -name "kustomization.yaml" -o -name "*.yaml" | \
        xargs grep -l "cluster" | sort | uniq >> "$output_file"
    
    echo "" >> "$output_file"
    echo "### Potential Orphaned Resources:" >> "$output_file"
    echo "Resources not matching known cluster patterns will be flagged for review." >> "$output_file"
}
```

#### Output Format
- **Single output file**: `.tmp/aws_resources.md`
- **Complete inventory**: All discovered resources by type
- **Cross-reference section**: Known clusters vs discovered resources
- **Read-only operation**: No deletions, only discovery and reporting
- **Simple format**: Markdown tables for easy review

### Design Principles

#### 1. **Repository-Aware Discovery**
- Automatically discovers clusters from repository structure
- No manual cluster list maintenance required
- Handles both regional specs and deployed configurations

#### 2. **Comprehensive Coverage**  
- Uses existing `find-aws-resources` for proven resource discovery
- Aggregates results across all clusters and regions
- Provides both summary and detailed views

#### 3. **Cost Transparency**
- Estimates monthly costs based on discovered resources
- Identifies cost optimization opportunities
- Supports budget planning and resource rightsizing

#### 4. **Automation-Friendly**
- Supports multiple output formats (table, CSV, JSON)
- Provides machine-readable data for further analysis
- Integrates with existing toolchain and workflows

#### 5. **Error Resilience**
- Continues discovery even if individual clusters fail
- Provides clear error messages and troubleshooting guidance
- Handles partial results gracefully

## Related Tools

### Prerequisites
- **[find-aws-resources.md](./find-aws-resources.md)** - Core discovery engine used by this tool

### Workflow Integration
- **[clean-aws.md](./clean-aws.md)** - Uses aggregated data to identify cleanup candidates
- **[health-check.md](./health-check.md)** - Validates deployed clusters have expected AWS resources

### Repository Integration  
- **[generate-cluster.md](./generate-cluster.md)** - Creates cluster configurations that generate AWS resources
- **[regenerate-all-clusters.md](./regenerate-all-clusters.md)** - Updates configurations that affect resource usage

### Analysis and Planning
- **[convert-cluster.md](./convert-cluster.md)** - Helps migrate clusters with comprehensive resource visibility
- **[bootstrap.md](./bootstrap.md)** - Initial deployment creates baseline resource footprint