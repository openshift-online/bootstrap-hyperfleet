# bin/test-find-aws-resources Requirements

## Purpose

The `test-find-aws-resources` script provides comprehensive testing for the `find-aws-resources` script, validating functionality, structure, and project-specific configurations without requiring AWS credentials.

## Functional Requirements

### Test Coverage Categories

#### Command Line Interface Tests
1. **Help Functionality**: Verify `--help` option displays usage information
2. **Default Discovery**: Test default cluster discovery behavior
3. **Verbose Output**: Test `--verbose` flag functionality
4. **All Regions**: Test `--all-regions` flag behavior
5. **JSON Output**: Test `--output-json` format option

#### Script Structure Validation
Must verify presence of required functions:
- **discover_ec2_instances**: EC2 resource discovery function
- **discover_networking**: VPC/subnet discovery function
- **discover_containers**: EKS/ECS container discovery function
- **discover_iam**: IAM resource discovery function

#### Resource Type Coverage
Must validate coverage for all required AWS resource types:
- **EC2_INSTANCES**: EC2 instance discovery
- **EBS_VOLUMES**: EBS volume discovery
- **LOAD_BALANCERS**: Load balancer discovery
- **AUTO_SCALING_GROUPS**: ASG discovery
- **VPCS**: VPC discovery
- **SUBNETS**: Subnet discovery
- **EKS_CLUSTERS**: EKS cluster discovery
- **RDS_INSTANCES**: RDS instance discovery
- **IAM_ROLES**: IAM role discovery
- **CLOUDFORMATION_STACKS**: CloudFormation stack discovery

### Testing Methodology

#### Timeout-Based Testing
- **Command Timeout**: Use 10-second timeout for AWS credential-dependent tests
- **Expected Failures**: Accept timeout/failure when AWS credentials unavailable
- **Dry Run Approach**: Test command structure without actual AWS calls

#### Output Validation
- **Head Limiting**: Use `head` to limit output during testing
- **Pattern Matching**: Use `grep` to verify specific content patterns
- **Silent Errors**: Redirect stderr to `/dev/null` for clean test output

### Project-Specific Validation

#### Instance Type Patterns
Must verify project-specific instance types:
- **m5.xlarge**: Primary instance type for OCP clusters
- **m5.large**: Standard instance type for EKS clusters  
- **c5.4xlarge**: Compute-optimized instance type

#### Regional Coverage
Must verify project-specific regions:
- **us-east-1**: Primary US East region
- **us-west-2**: Primary US West region
- **eu-west-1**: Primary European region
- **ap-southeast-1**: Primary Asia Pacific region

#### Kubernetes Integration
Must verify Kubernetes-specific patterns:
- **kubernetes.io/cluster**: Cluster tag pattern for resource identification

### Output Requirements

#### Test Progress Reporting
```
=== Testing find-aws-resources script ===

Test 1: Help functionality
==========================
[help output preview]
✅ Help test passed

Test 2: Default cluster discovery (dry run)
===========================================
[discovery output preview]
✅ Default discovery test passed
```

#### Validation Results
```
Test 6: Script structure validation
===================================
✅ EC2 discovery function found
✅ Networking discovery function found
✅ Container discovery function found
✅ IAM discovery function found

Test 7: Resource type coverage validation
========================================
✅ EC2_INSTANCES coverage found
✅ EBS_VOLUMES coverage found
❌ SOME_RESOURCE coverage missing
```

#### Usage Examples
```
The find-aws-resources script is ready for use.

Usage examples:
  ./bin/find-aws-resources ocp-02                    # Discover ocp-02 resources
  ./bin/find-aws-resources eks-02 --all-regions     # Search all regions for eks-02
  ./bin/find-aws-resources ocp-03 --verbose         # Detailed resource information
  ./bin/find-aws-resources hcp-01 --output-json     # JSON formatted output
```

### Error Handling Requirements

#### Missing Script
- **Script Availability**: Verify `./bin/find-aws-resources` exists before testing
- **Permission Issues**: Handle script execution permission problems

#### AWS Credential Absence
- **Expected Failures**: Handle timeouts and credential errors gracefully
- **Dry Run Validation**: Validate script structure without AWS access
- **Silent Operation**: Suppress AWS-related error messages during testing

### Test Execution Requirements

#### Safe Testing
- **No AWS Calls**: Use timeouts to prevent actual AWS API calls
- **Structure Only**: Test script structure and command parsing only
- **No Side Effects**: Ensure testing doesn't modify any resources

#### Comprehensive Coverage
- **All CLI Options**: Test every command line option and flag
- **Function Presence**: Verify all required discovery functions exist
- **Resource Coverage**: Validate all AWS resource types are handled
- **Project Patterns**: Verify project-specific configurations

### Dependencies

#### External Commands
- **timeout**: Command timeout utility
- **grep**: Pattern matching for function/resource validation
- **head**: Output limiting for clean test results

#### Target Script
- **find-aws-resources**: Must exist in `bin/` directory
- **Executable Permissions**: Script must be executable
- **Function Structure**: Script must contain required discovery functions

### Usage Patterns

```bash
# Run complete test suite
./bin/test-find-aws-resources

# Verify script is ready for production use
./bin/test-find-aws-resources && echo "Script validated"

# Integration with CI/CD
if ./bin/test-find-aws-resources; then
    echo "find-aws-resources validation passed"
else
    echo "find-aws-resources validation failed"
    exit 1
fi
```

### Integration Requirements

#### Development Workflow
- **Pre-commit Testing**: Run tests before committing find-aws-resources changes
- **Validation Gateway**: Ensure script structure is correct before deployment
- **Documentation Sync**: Verify usage examples match actual capabilities

#### CI/CD Integration
- **Automated Testing**: Include in automated test suites
- **Regression Prevention**: Catch structural changes in find-aws-resources
- **Quality Assurance**: Ensure comprehensive AWS resource coverage

## Related Tools

### Primary Target
- **[find-aws-resources.md](./find-aws-resources.md)** - The primary tool that this testing script validates

### Related Validation
- **[clean-aws.md](./clean-aws.md)** - Uses the same discovery patterns that this tool validates

### Testing Infrastructure
- **[validate-docs.md](./validate-docs.md)** - Validates documentation quality for testing guides
- **[status.md](./status.md)** - Monitors the clusters that generate AWS resources for testing

## Design Principles

*This tool enables **test-driven validation** - ensuring the find-aws-resources script maintains comprehensive AWS resource discovery capabilities and project-specific configurations through automated structural testing.*