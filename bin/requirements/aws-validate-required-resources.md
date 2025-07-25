# aws-validate-required-resources

## Purpose
Validates that an AWS account has sufficient quotas and permissions to successfully provision OpenShift clusters.

## Synopsis
```bash
aws-validate-required-resources [OPTIONS] [cluster-requirements]
```

## Description
This tool validates AWS account quotas against OpenShift cluster requirements to prevent deployment failures. It can automatically generate resource requirements based on cluster type or use provided specification files. The tool checks service quotas, current resource usage, and calculates whether the account can support the requested cluster configuration.

The tool leverages quota validation patterns extracted from the UHC clusters service preflight checks to provide comprehensive validation across multiple AWS services.

## Arguments
- `[cluster-requirements]`: Optional JSON or YAML file containing cluster specification. If not provided, the tool generates requirements automatically based on cluster type.

## Options
- `--region REGION`: AWS region to validate (default: us-west-2)
- `--cluster-type TYPE`: Cluster type - 'ocp', 'hcp', or 'eks' (default: ocp)
- `--instance-type TYPE`: Override default instance types (format: master:m5.xlarge,worker:m5.large)
- `--buffer-vcpu NUM`: vCPU quota buffer for safety margin (default: 10)
- `--output FORMAT`: Output format - 'text', 'json', or 'table' (default: text)
- `--check TYPE`: Specific check to run - 'vcpu', 'storage', 'network', 'iam' or 'all' (default: all)
- `--verbose`: Enable detailed validation output
- `--mock`: Use mock AWS data for testing (no real API calls)
- `--non-interactive`: Disable interactive prompts (use command-line args or defaults)

## Cluster Requirements File Format

### JSON Format (Legacy)
```json
{
  "region": "us-east-1",
  "cluster_type": "ocp",
  "master_nodes": {
    "instance_type": "m5.xlarge",
    "replicas": 3
  },
  "infra_nodes": {
    "instance_type": "m5.large", 
    "replicas": 3
  },
  "compute_nodes": {
    "instance_type": "m5.large",
    "replicas": 2
  },
  "autoscale": {
    "max_replicas": 10
  }
}
```

### YAML Format (Regional Cluster Specification)
```yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-test
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
  # Compute configuration (always required)
  compute:
    instanceType: m5.large
    replicas: 2
    
  # Master nodes (ocp only) - optional, uses defaults if not specified
  master:
    instanceType: m5.xlarge  
    replicas: 3
    
  # Infra nodes (ocp only) - optional, uses defaults if not specified  
  infra:
    instanceType: m5.large
    replicas: 3
    
  # Autoscaling configuration - optional
  autoscale:
    enabled: true
    maxReplicas: 10
    minReplicas: 1
    
  # Type-specific configuration - OCP
  openshift:
    version: "4.14"
    channel: stable
```

## Validation Checks

### vCPU Quota Validation
- **Quota Code**: L-1216C47A (Running On-Demand Standard instances)
- **Validates**: Current usage + required vCPUs ≤ quota limit - buffer
- **Calculates**: Master + Infra + Compute + Autoscale requirements
- **OCP Clusters**: Standard OpenShift with master + infra + compute nodes
- **HCP Clusters**: Hypershift with managed control plane (compute nodes only)  
- **EKS Clusters**: Amazon EKS with managed control plane (compute nodes only)

### Storage Quota Validation  
- **EBS Volume Quota**: General Purpose SSD (gp3) volumes
- **Snapshot Quota**: EBS snapshot limits
- **Validates**: Storage requirements for OS disks and persistent volumes

### Network Quota Validation
- **VPC Limits**: VPCs per region, subnets per VPC
- **Security Groups**: Rules per security group, groups per VPC
- **Load Balancers**: Application and Network Load Balancer limits
- **Elastic IPs**: Address allocation limits

### IAM Quota Validation
- **Roles**: IAM roles per account
- **Instance Profiles**: Instance profiles per account  
- **Policies**: Customer managed policies per account

## Exit Codes
- **0**: All validations passed - cluster can be provisioned
- **1**: Quota validation failed - insufficient resources
- **2**: Permission denied - missing AWS API access
- **3**: Configuration error - invalid parameters or cluster spec
- **4**: AWS API error - service unavailable or authentication failure

## Examples

### Automatic Generation (No File Required)
```bash
# Interactive prompts for region and cluster type - generates requirements automatically
aws-validate-required-resources
# Will prompt: Enter AWS region (default: us-west-2):
# Will prompt: Enter cluster type (default: ocp):

# Non-interactive with automatic generation
aws-validate-required-resources --non-interactive
aws-validate-required-resources --non-interactive --region us-east-1 --cluster-type hcp

# Custom instance types with automatic generation
aws-validate-required-resources --cluster-type ocp --instance-type master:m5.2xlarge,worker:m5.xlarge
```

### File-Based Validation
```bash
# Use existing cluster specification file
aws-validate-required-resources cluster-spec.json

# Interactive prompts with file override
aws-validate-required-resources cluster-spec.json
# Will still prompt for region/cluster type to override file values

# Non-interactive with file
aws-validate-required-resources --non-interactive cluster-spec.json
```

### Basic Validation
```bash
# Validate with JSON format (legacy)
aws-validate-required-resources --region us-west-2 --cluster-type ocp cluster-spec.json

# Validate with YAML regional spec format (recommended)
aws-validate-required-resources regions/us-west-2/ocp-03/region.yaml

# Validate Hypershift cluster
aws-validate-required-resources --cluster-type hcp cluster-spec.json

# Validate EKS cluster  
aws-validate-required-resources --cluster-type eks cluster-spec.json
```

### Advanced Usage
```bash
# Check only vCPU quotas with custom buffer (automatic generation)
aws-validate-required-resources --check vcpu --buffer-vcpu 20 --cluster-type ocp

# Generate JSON report for automation (automatic generation)
aws-validate-required-resources --output json --cluster-type hcp --region us-west-2 > validation-report.json

# Test with mock data (no AWS credentials required)
aws-validate-required-resources --mock --cluster-type eks

# Combine automatic generation with file-based validation
aws-validate-required-resources --check vcpu --buffer-vcpu 20 cluster-spec.json
```

### Smart Defaults by Cluster Type

The tool automatically generates sensible defaults based on cluster type:

**OCP (OpenShift Container Platform)**:
- Master nodes: 3 × m5.xlarge (12 vCPUs)
- Infra nodes: 3 × m5.large (6 vCPUs)  
- Compute nodes: 3 × m5.large (6 vCPUs)
- Max autoscale: 10 nodes (20 vCPUs)
- **Total**: 44 vCPUs

**HCP (Hosted Control Planes/Hypershift)**:
- Compute nodes: 3 × m5.large (6 vCPUs)
- Max autoscale: 15 nodes (30 vCPUs)
- **Total**: 36 vCPUs (no master/infra - managed control plane)

**EKS (Amazon Elastic Kubernetes Service)**:
- Compute nodes: 3 × m5.large (6 vCPUs)
- Max autoscale: 12 nodes (24 vCPUs)
- **Total**: 30 vCPUs (no master/infra - managed control plane)

### Inline Cluster Specification
```bash
# Override instance types for automatic generation
aws-validate-required-resources \
  --region us-east-1 \
  --cluster-type ocp \
  --instance-type master:m5.2xlarge,worker:m5.xlarge

# You can still override with files
aws-validate-required-resources \
  --region us-east-1 \
  --cluster-type standard \
  --instance-type master:m5.2xlarge,worker:m5.xlarge \
  cluster-spec.json
```

## Output Formats

### Text Output (Default)
```
🌩️  AWS Account Validation for OpenShift Cluster

📊 Cluster Requirements:
   Region: us-east-1
   Type: standard
   Master nodes: 3 × m5.xlarge (12 vCPUs)
   Infra nodes: 3 × m5.large (6 vCPUs)  
   Compute nodes: 2 × m5.large (4 vCPUs)
   Max autoscale: 10 × m5.large (20 vCPUs)
   Total required: 42 vCPUs

✅ vCPU Quota Validation
   Available: 358 vCPUs (limit: 400, usage: 32, buffer: 10)
   Required: 42 vCPUs
   Status: PASSED

✅ Storage Quota Validation  
   Available: 450 volumes (limit: 500, usage: 50)
   Required: 15 volumes
   Status: PASSED

🎉 All validations passed! Cluster can be provisioned.
```

### JSON Output
```json
{
  "validation_result": "PASSED",
  "timestamp": "2025-07-25T15:30:00Z",
  "region": "us-east-1",
  "cluster_requirements": {
    "total_vcpus": 42,
    "total_volumes": 15,
    "cluster_type": "standard"
  },
  "validations": {
    "vcpu_quota": {
      "status": "PASSED",
      "available": 358,
      "required": 42,
      "quota_limit": 400,
      "current_usage": 32
    },
    "storage_quota": {
      "status": "PASSED", 
      "available": 450,
      "required": 15
    }
  }
}
```

## Implementation Details

### AWS Service Integrations
- **Service Quotas API**: Real-time quota limits (requires `servicequotas:GetServiceQuota`)
- **EC2 API**: Current resource usage (requires `ec2:Describe*`)
- **IAM API**: Role and policy validation (requires `iam:List*`)
- **ELB API**: Load balancer quota checks (requires `elasticloadbalancing:Describe*`)

### Quota Calculation Logic
Uses the same calculation patterns as UHC clusters service:
1. **Standard Clusters**: Master + Infra + Compute + Autoscale vCPUs
2. **Hypershift Clusters**: Compute + Autoscale only (managed control plane)
3. **Autoscaler Limits**: Respects cluster autoscaler maximum core constraints
4. **Safety Buffer**: Configurable headroom for quota fluctuations

### Error Handling
- **Graceful Degradation**: Continues validation if some APIs are unavailable
- **Permission Fallbacks**: Skips checks requiring unavailable permissions
- **Detailed Suggestions**: Provides actionable remediation steps
- **Retry Logic**: Handles transient AWS API failures

### Instance Type Mappings
Comprehensive vCPU mappings for all OpenShift-supported instance families:
- T3/T3a series (burstable performance)
- M5/M5a/M6i series (general purpose)  
- C5/C5a/C6i series (compute optimized)
- R5/R5a/R6i series (memory optimized)
- Auto-detection from EC2 API for unknown types

## Dependencies
- **AWS CLI**: AWS SDK authentication and region configuration
- **jq**: JSON processing for cluster requirements and output formatting
- **yq** or **python3 with PyYAML**: YAML processing for regional cluster specifications
- **bash 4.0+**: Advanced array and arithmetic operations

## Error Recovery
- **Missing Cluster Spec**: Generates sample specification file
- **Invalid JSON**: Validates and suggests corrections
- **AWS Auth Failure**: Provides credential configuration guidance  
- **Quota API Errors**: Falls back to estimated limits from documentation

## Integration Points
- **cluster-create**: Pre-flight validation before cluster provisioning
- **monitor-health**: Ongoing quota monitoring for existing clusters
- **aws-find-resources**: Resource discovery for capacity planning
- **Tekton Pipelines**: Automated validation in CI/CD workflows

## Related Tools
- `aws-find-resources`: Discovers existing AWS resources
- `aws-clean-resources`: Cleans up AWS resources after validation failures
- `cluster-create`: Uses validation results for cluster sizing decisions
- `monitor-health`: Tracks ongoing resource consumption