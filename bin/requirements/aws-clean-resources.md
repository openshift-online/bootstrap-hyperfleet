# bin/aws-clean-resources Requirements

## Functional Requirements for bin/aws-clean-resources

### Primary Objective
Create an interactive AWS resource cleanup tool that processes JSON resource files (from aws-find-resources), queries AWS for additional details, presents resources for selection, generates a deletion manifest file, and optionally executes the deletions with proper dependency ordering.

### Key Features
- **JSON Input Processing**: Accepts resource files from aws-find-resources tool
- **AWS Detail Enrichment**: Queries AWS for current resource details and status
- **Interactive Selection**: Present each resource with y/N prompts (N is default)
- **Deletion Manifest**: Creates {input-file}-delete-me.json with selected resources
- **Execution Phase**: Optional immediate deletion execution with y/N prompt (N is default)
- **Dependency-Aware Deletion**: Sorts resources by dependency order to prevent AWS violations
- **Comprehensive Coverage**: Supports 25+ resource types with proper cleanup logic
- **Debug Mode**: `--debug` for comprehensive logging
- **Skip Checks Mode**: `--skip-checks` to bypass all y/N prompts and force delete all resources

### Command Line Interface
```bash
# Process a resource file interactively
./bin/aws-clean-resources .tmp/aws-resources.json

# With debug logging
./bin/aws-clean-resources cluster-resources.json --debug

# Skip all prompts and force delete all resources
./bin/aws-clean-resources cluster-resources.json --skip-checks

# Show help
./bin/aws-clean-resources --help
```

### Implementation Reference

#### Core Functions
The script implements these key functions (see `bin/aws-clean-resources` for full implementation):

- **`get_resource_details()`**: Query AWS for current resource details
- **`process_resource_type()`**: Process each resource type with interactive prompts
- **`debug_log()`**: Timestamped debug logging

#### Architecture Overview
```bash
# 1. Argument parsing (input file, --debug, --skip-checks, --help)
# 2. JSON file validation and region extraction
# 3. Resource type processing with AWS detail enrichment
# 4. Interactive resource selection (y/N prompts, unless --skip-checks)
# 5. Deletion manifest generation
# 6. Optional execution phase (with y/N prompt unless --skip-checks)
```

#### JSON Input Processing

**Input Format**: Structured JSON from aws-find-resources tool
```json
{
  "REGION": "us-west-2",
  "EC2_INSTANCES": [["i-123", "m5.large", "running", "vpc-123", ...]],
  "VPCS": [["vpc-123", "10.0.0.0/16", "available", false, "cluster-vpc"]],
  "GLOBAL_IAM_RESOURCES": {
    "IAM_ROLES": [["role-name", "arn", "2025-01-01", null]]
  }
}
```

**Resource Types Processed** (25+ types):
1. EC2 Instances
2. EBS Volumes & Snapshots
3. Load Balancers (ALB/NLB + Classic)
4. Target Groups
5. Auto Scaling Groups & Launch Templates
6. VPCs, Subnets, Security Groups
7. Network Interfaces, Route Tables, NAT Gateways
8. Internet Gateways, Elastic IPs
9. VPC Endpoints, Network ACLs
10. EFS File Systems & Mount Targets
11. RDS Instances & Clusters
12. EKS Clusters, ECR Repositories
13. CloudFormation Stacks
14. IAM Roles, Policies, Instance Profiles

**AWS Detail Enrichment**:
```bash
# Example: EC2 Instance details
aws ec2 describe-instances --instance-ids "$resource_id" \
  --query 'Reservations[0].Instances[0].[InstanceType,State.Name,LaunchTime,PrivateIpAddress,PublicIpAddress,Tags[?Key==`Name`].Value|[0]]'

# Example: VPC details  
aws ec2 describe-vpcs --vpc-ids "$resource_id" \
  --query 'Vpcs[0].[CidrBlock,State,IsDefault,Tags[?Key==`Name`].Value|[0]]'
```

#### Interactive Resource Selection

**User Experience**: Each resource is presented with context and details
```bash
=== VPCS (1 found) ===

Resource: vpc-0cad6af52ce766e7e
  Raw data: ["vpc-0cad6af52ce766e7e","10.0.0.0/16","available",false,"ocp-01-mturansk-t3-24qq2-vpc"]
  AWS details: 10.0.0.0/16 available false ocp-01-mturansk-t3-24qq2-vpc

Delete this VPCS resource? (y/N): n
  Skipped: vpc-0cad6af52ce766e7e
```

**Selection Logic**:
- **Default Response**: N (No) - conservative approach prevents accidental deletions
- **User Confirmation**: Only resources explicitly marked 'y' or 'Y' are selected
- **Skip Checks Mode**: With `--skip-checks`, all resources are automatically selected without prompts
- **Detail Enrichment**: Shows both raw JSON data and current AWS status
- **Resource Context**: Groups resources by type with counts

#### Output Manifest Generation

**Deletion Manifest Format**: Structured JSON with metadata
```json
{
  "metadata": {
    "source_file": ".tmp/aws-resources.json",
    "region": "us-west-2", 
    "timestamp": "2025-07-25 10:30:00 UTC",
    "total_selected": 3
  },
  "selected_resources": [
    {
      "type": "VPCS",
      "id": "vpc-0cad6af52ce766e7e", 
      "raw_data": ["vpc-0cad6af52ce766e7e", "10.0.0.0/16", "available", false, "cluster-vpc"],
      "aws_details": "10.0.0.0/16 available false cluster-vpc"
    }
  ]
}
```

**File Naming**: `{input-file}-delete-me.json`
- Input: `.tmp/aws-resources.json` → Output: `.tmp/aws-resources-delete-me.json`
- Input: `cluster-resources.json` → Output: `cluster-resources-delete-me.json`

#### Execution Phase

**Post-Selection Execution**: After creating the deletion manifest, the tool prompts for immediate execution
```bash
=== Summary ===
Source file: .tmp/aws-resources.json
Resources selected for deletion: 5
Output file: .tmp/aws-resources-delete-me.json

Next steps:
1. Review the deletion list: jq . '.tmp/aws-resources-delete-me.json'

Please review the deletion manifest above or run: jq . '.tmp/aws-resources-delete-me.json'

Do you want to execute the deletion of these 5 resources now? (y/N): y

=== Executing Resource Deletions ===
Region: us-west-2
Total resources to delete: 5

# With --skip-checks flag, execution proceeds automatically:
=== Executing Resource Deletions ===
Region: us-west-2
Total resources to delete: 5
(no execution prompt - proceeds automatically)
```

**Two-Phase Dependency-Aware Deletion Order**: Resources are deleted in a carefully orchestrated two-phase approach to avoid AWS dependency violations:

**Phase 1: Service Termination** (Disconnect and stop all services)
1. **EC2 Instances** → Terminate compute workloads and release network interfaces
2. **EBS Snapshots** → No dependencies, safe to delete early
3. **Load Balancers** (ALB/NLB/Classic) → Wait for complete deletion including network interfaces
4. **Target Groups, Auto Scaling Groups, Launch Templates** → Application layer resources
5. **EKS/RDS Clusters** → Managed service resources
6. **VPC Endpoints** → Service endpoints
7. **NAT Gateways** → Wait for complete deletion and network interface cleanup
8. **ECR, CloudFormation, EBS Volumes** → Remaining services and storage

*60-second wait period for all services to fully terminate and release dependencies*

**Phase 2: Infrastructure Cleanup** (Clean up networking after services are gone)
9. **Route Tables** → Clear all routes, disassociate subnets, then delete tables
10. **Internet Gateways** → Detach from VPC, then delete with retry logic (moved before subnets)
11. **Network Interfaces** → Delete any remaining orphaned interfaces
12. **Security Groups** → Clear all rules with retry logic, extensive dependency cleanup
13. **Network ACLs, EFS resources** → Network and storage layer
14. **Subnets** → Delete with retry logic after ensuring route tables and gateways are clear
15. **Elastic IPs** → Release IP addresses
16. **VPCs** → Container for all resources, deleted last with comprehensive dependency cleanup
17. **IAM Resources** → Identity layer, cleaned up last

**Special Deletion Logic**: Each resource type implements proper cleanup procedures:
- **Security Groups**: Revokes all inbound/outbound rules before deletion
- **Route Tables**: Deletes all non-local routes before table deletion
- **Internet Gateways**: Detaches from VPC before deletion
- **Network Interfaces**: Detaches from instances with force flag
- **VPCs**: Comprehensive dependency cleanup including VPC peering connections, remaining VPC endpoints, custom DHCP options sets, network ACLs, and final network interface scan
- **IAM Roles**: Detaches managed policies and deletes inline policies first
- **IAM Instance Profiles**: Removes associated roles before deletion

**Error Handling and Reporting**:
```bash
=== Deletion Summary ===
Successfully processed: 4
Failed: 1
Total: 5

⚠ Some deletions failed. Check AWS console for remaining resources.
Resources may have dependencies that prevent deletion or may no longer exist.
```

**Critical Implementation Notes**:
1. **Safe Arithmetic**: All arithmetic operations use the safe increment pattern `variable=$((variable + 1))` instead of `((variable++))` to avoid issues under `set -euo pipefail` where the increment operator returns non-zero when the result is 0.
2. **Two-Phase Execution**: Phase 1 terminates all services and waits 60 seconds. Phase 2 cleans up networking infrastructure with extensive dependency handling.
3. **Comprehensive Wait Logic**: Load balancers and NAT gateways wait for complete deletion (up to 5-7.5 minutes) to ensure all network interfaces are released.
4. **Extensive Retry Logic**: Security groups (5 attempts), subnets (3 attempts), Internet gateways (3 attempts), route tables (3 attempts), and VPCs (5 attempts) all include retry logic with backoff.
5. **Route Table Association Cleanup**: Route tables are explicitly disassociated from subnets before deletion to prevent dependency violations.
6. **Orphaned Resource Detection**: The script automatically finds and cleans up orphaned network interfaces before attempting security group deletion.
7. **Internet Gateway Detachment**: Internet gateways are properly detached from VPCs before deletion attempts.
8. **Resource ID Validation**: AWS resource IDs are validated using regex patterns to prevent processing malformed data (e.g., instance types instead of instance IDs). Invalid resources are skipped with detailed warnings.

#### Status and Validation

**Current Status** (2025-07-25):
- ✅ **Script Location**: `bin/aws-clean-resources` (executable)
- ✅ **JSON Input Processing**: Accepts and validates resource files from aws-find-resources
- ✅ **AWS Detail Enrichment**: Queries AWS for current resource status and details
- ✅ **Interactive Selection**: y/N prompts with N as default for safe operation
- ✅ **Deletion Manifest**: Creates structured JSON output with metadata
- ✅ **Execution Phase**: Optional immediate deletion with dependency-aware ordering
- ✅ **Comprehensive Deletion**: 25+ resource types with proper cleanup logic
- ✅ **Dependency Ordering**: Prevents AWS dependency violations during deletion
- ✅ **Error Resilience**: Failed deletions don't stop the process, detailed reporting
- ✅ **Debug Mode**: `--debug` flag with timestamped logging
- ✅ **Region Detection**: Automatically extracts region from input JSON file
- ✅ **Safety Features**: Conservative defaults (N) for both selection and execution
- ✅ **Skip Checks Mode**: `--skip-checks` flag to bypass all prompts for automated use

**Test Commands**:
```bash
# Process resource file interactively
./bin/aws-clean-resources .tmp/aws-resources.json

# With debug logging
./bin/aws-clean-resources .tmp/aws-resources.json --debug

# Skip all prompts for automated deletion
./bin/aws-clean-resources .tmp/aws-resources.json --skip-checks

# Show help and usage
./bin/aws-clean-resources --help
```

**Example Usage Flow**:
```bash
# 1. Discover resources
./bin/aws-find-resources ocp-01-mturansk-t3 > .tmp/cluster-resources.json

# 2. Interactively select and optionally execute deletions
./bin/aws-clean-resources .tmp/cluster-resources.json

# Alternative: Skip all prompts and delete everything
./bin/aws-clean-resources .tmp/cluster-resources.json --skip-checks

# Alternative: Review deletion manifest separately (if execution was skipped)
jq . .tmp/cluster-resources-delete-me.json

# Alternative: Execute deletions later using the manifest (external process)
# The manifest file contains all necessary information for batch deletion
```

**Complete Workflow Example**:
```bash
# Step 1: Resource Discovery → JSON file
./bin/aws-find-resources ocp-01-mturansk-t3 > .tmp/cluster-resources.json

# Step 2: Interactive Selection → Deletion manifest + Optional execution
./bin/aws-clean-resources .tmp/cluster-resources.json
# User selects resources interactively (y/N for each)
# Creates: .tmp/cluster-resources-delete-me.json
# Prompts: "Execute deletion now? (y/N)"
# If yes: Executes deletions in dependency order
# If no: Exits with manifest file for later use

# The tool now provides complete end-to-end cleanup capability
```

### Implementation Notes

**Key Technical Achievements**:
1. **JSON Processing Pipeline**: Seamless integration with aws-find-resources output
2. **AWS Detail Enrichment**: Real-time resource status and configuration queries
3. **Safe Interaction Design**: Default 'N' response prevents accidental deletions
4. **Structured Output**: Machine-readable deletion manifests with metadata
5. **Comprehensive Coverage**: Supports all resource types from aws-find-resources
6. **Dependency-Aware Execution**: Implements proper AWS resource deletion ordering
7. **Robust Error Handling**: Failed deletions don't abort the entire process
8. **Complete Workflow**: End-to-end resource cleanup with human oversight

**Architecture Summary**:
- **Input Validation**: JSON structure validation and region extraction
- **Resource Processing**: Type-specific AWS detail queries and user prompts
- **Selection Management**: Tracks user choices in structured JSON format
- **Manifest Generation**: Creates deletion manifest with source traceability
- **Execution Engine**: Dependency-ordered deletion with comprehensive AWS API coverage
- **Error Resilience**: Individual failures tracked but don't stop overall process

**Workflow Integration**:
The tool now provides a complete resource cleanup solution: discover → select → review → execute. It serves as both an interactive selection tool and a comprehensive deletion executor, with proper dependency management and error handling.

**Design Principles**:
- **Conservative by Default**: N is default response for both selection and execution
- **Transparent Operation**: Shows both raw data and current AWS details
- **Traceability**: Output manifest includes source file and timestamp metadata
- **Dependency Awareness**: Respects AWS resource dependencies during deletion
- **Error Isolation**: Failed deletions are reported but don't affect other resources
- **Human Oversight**: Multiple confirmation points prevent accidental bulk deletions
- **Automation Support**: `--skip-checks` flag enables unattended operation when needed

## Related Tools

### Discovery Dependencies
- **[aws-find-resources.md](./aws-find-resources.md)** - Generates the JSON input files processed by this tool

### Workflow Sequence  
1. **aws-find-resources**: Discovers and catalogs AWS resources → JSON file
2. **aws-clean-resources**: Processes JSON, user selects resources → deletion manifest + optional immediate execution
3. **Optional**: External tools can process deletion manifest for batch operations or auditing

### Validation and Testing
- **[test-find-aws-resources.md](./test-find-aws-resources.md)** - Validates resource discovery patterns
- **[aws-find-all-resources.md](./aws-find-all-resources.md)** - Batch resource discovery across multiple clusters

### Cluster Lifecycle
- **[generate-cluster.md](./generate-cluster.md)** - Creates clusters that eventually require cleanup
- **[status.md](./status.md)** - Monitor cluster status before cleanup decisions