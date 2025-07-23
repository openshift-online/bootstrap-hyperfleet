# bin/clean-aws Requirements

## Functional Requirements for bin/clean-aws

### Primary Objective
Create a single, comprehensive bash script that discovers all orphaned cluster resources and deletes them with proper dependency ordering. Supports both interactive and automated execution.

### Key Features
- **Default Values**: `mt-test` cluster, `us-east-1` region
- **Automated Mode**: `--disable-prompts` for non-interactive execution  
- **Debug Mode**: `--debug` for comprehensive logging
- **Dependency-Aware Deletion**: Proper order to avoid AWS dependency violations
- **Comprehensive Discovery**: Two-pass discovery (tag-based + VPC-based)

### Command Line Interface
```bash
# Interactive mode with defaults
./bin/clean-aws

# Fully automated execution  
./bin/clean-aws --disable-prompts

# Debug mode
./bin/clean-aws --debug --disable-prompts
```

### Implementation Reference

#### Core Functions
The script implements these key functions (see `bin/clean-aws` for full implementation):

- **`find_resources()`**: Two-pass resource discovery
- **`display_resources()`**: Resource listing with counts
- **`delete_resource()`**: Type-specific deletion logic
- **`debug_log()`**: Timestamped debug logging

#### Architecture Overview
```bash
# 1. Argument parsing (--debug, --disable-prompts)
# 2. Input handling (defaults: mt-test, us-east-1)  
# 3. Two-pass resource discovery
# 4. Dependency-ordered deletion
# 5. Comprehensive error handling
```

#### Resource Discovery Strategy

**Two-Pass Discovery**:
1. **Tag-Based Search**: Find resources with cluster ID in tag values
2. **VPC-Based Search**: Find additional resources by VPC association (for untagged resources)

**Resource Types Discovered** (12 types):
1. Load Balancers (ALB/NLB + Classic)
2. RDS Instances  
3. EC2 Instances
4. VPC Endpoints
5. NAT Gateways
6. Network Interfaces
7. Route Tables (non-main)
8. Security Groups (non-default)
9. Network ACLs (non-default) 
10. Subnets
11. Internet Gateways
12. VPCs

**Key AWS CLI Patterns**:
```bash
# Tag-based search
--filters "Name=tag-value,Values=*${cluster_id}*"

# VPC-based search (second pass)
--filters "Name=vpc-id,Values=$vpc_id"

# Name-based search (for load balancers)
--query "LoadBalancers[?contains(LoadBalancerName, '${cluster_id}')]"

# Non-default filtering
--query 'RouteTables[?length(Associations[?Main == `true`]) == `0`]'
--query 'NetworkAcls[?IsDefault == `false`]'
```

#### Dependency-Aware Deletion Order

**Critical**: Resources must be deleted in this exact order to avoid AWS dependency violations:

```bash
# Correct deletion sequence (from bin/clean-aws)
1.  Load Balancers (ALB/NLB + Classic) - nothing depends on them
2.  RDS Instances - nothing depends on them  
3.  EC2 Instances - release network interfaces
4.  VPC Endpoints - block route deletion
5.  NAT Gateways - use ENIs, block route deletion
6.  Network Interfaces - attach to subnets, use security groups  
7.  Route Tables - clear routes first, then delete tables
8.  Security Groups - clear rules first, then delete groups
9.  Network ACLs - associated with subnets
10. Subnets - contain ENIs, use route tables
11. Internet Gateways - detach first, then delete
12. VPCs - last, everything else is inside
```

#### Special Deletion Logic

**Security Groups**: Clear all rules before deletion
```bash
# Remove inbound rules to break circular dependencies
aws ec2 revoke-security-group-ingress --group-id "$id" --ip-permissions "$inbound_rules"

# Remove outbound rules (except default allow-all)  
aws ec2 revoke-security-group-egress --group-id "$id" --ip-permissions "$outbound_rules"
```

**Route Tables**: Clear routes before deletion
```bash
# Delete all non-local routes first
aws ec2 delete-route --route-table-id "$id" --destination-cidr-block "$destination"
aws ec2 delete-route --route-table-id "$id" --destination-prefix-list-id "$destination"
```

**Network Interfaces**: Detach before deletion
```bash
# Check attachment and force detach if needed
aws ec2 detach-network-interface --attachment-id "$attachment" --force
```

#### Status and Validation

**Current Status** (2025-07-19):
- ✅ **Script Location**: `bin/clean-aws` (executable)
- ✅ **Working Implementation**: All core features functional
- ✅ **Default Values**: `mt-test` cluster, `us-east-1` region
- ✅ **Automated Mode**: `--disable-prompts` flag implemented
- ✅ **Debug Mode**: `--debug` flag with timestamped logging
- ✅ **Dependency Ordering**: Fixed JMESPath queries and deletion sequence
- ✅ **Two-Pass Discovery**: Tag-based + VPC-based resource discovery
- ✅ **Special Deletion Logic**: Security group rule clearing, route clearing, ENI detachment
- ✅ **Comprehensive Coverage**: 12 resource types discovered and deleted

**Test Commands**:
```bash
# Basic functionality test
./bin/clean-aws
# Interactive mode with defaults

# Fully automated execution
./bin/clean-aws --disable-prompts
# Uses defaults: mt-test, us-east-1, auto-deletes all resources

# Debug mode
./bin/clean-aws --debug --disable-prompts
# Full logging with automated execution
```

**Recent Validation** (2025-07-19):
- ✅ Successfully discovered 12 route tables via VPC-based second pass
- ✅ Successfully cleared routes before route table deletion  
- ✅ Successfully cleared security group rules before deletion
- ✅ Fixed JMESPath query syntax errors (`length(Associations[?Main == 'true']) == '0'`)
- ✅ VPC endpoint discovery and deletion working
- ✅ Enhanced VPC-based discovery to include security groups (fixes untagged resource detection)
- ✅ Successfully deleted previously blocked VPC after security group cleanup

### Implementation Notes

**Key Technical Achievements**:
1. **Fixed Dependency Violations**: Proper deletion order prevents AWS errors
2. **Rule Clearing**: Security groups and route tables cleared before deletion
3. **VPC-Based Discovery**: Second pass finds untagged resources
4. **Error Handling**: Comprehensive error handling with `set -euo pipefail`
5. **Debug Logging**: Timestamped logs for troubleshooting

**Architecture Summary**:
- **Modular Functions**: `find_resources()`, `display_resources()`, `delete_resource()`
- **Two-Pass Discovery**: Tag search + VPC association search
- **Dependency-Aware**: 12-step deletion sequence
- **User Experience**: Interactive prompts with defaults + automated mode

The script successfully eliminates code duplication by implementing a working, comprehensive solution in `bin/clean-aws` while maintaining concise requirements documentation that references the implementation rather than duplicating code.

## Related Tools

### Discovery Dependencies
- **[find-aws-resources.md](./find-aws-resources.md)** - Provides the discovery patterns and resource identification used by this cleanup tool

### Validation and Testing
- **[test-find-aws-resources.md](./test-find-aws-resources.md)** - Validates resource discovery patterns used in cleanup

### Cluster Lifecycle
- **[generate-cluster.md](./generate-cluster.md)** - Creates clusters that eventually require cleanup
- **[status.md](./status.md)** - Monitor cluster status before cleanup decisions