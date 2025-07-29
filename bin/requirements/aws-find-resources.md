# bin/aws-find-resources Requirements

## Functional Requirements for bin/aws-find-resources

### Primary Objective
Create comprehensive AWS resource discovery for OpenShift clusters using **relationship-based discovery** that follows AWS resource dependencies through IDs rather than relying solely on tags. This ensures **zero orphaned resources** are missed, including those with missing or incorrect cluster tags.

### Discovery Strategy: Comprehensive Relationship Following

**Problem Solved:**
Traditional tag-based discovery misses resources that:
- Lack proper cluster identification tags (common with ELB-managed network interfaces)
- Have non-standard naming conventions (AWS service-generated names)
- Are created by AWS services that don't inherit cluster tags
- Exist as dependencies of properly-tagged resources

**Solution: 4-Level Dependency Hierarchy**

The script follows AWS resource relationships through their IDs in a structured dependency tree:

```
LEVEL 0 (ROOT): VPC + Global Resources
├── VPCs (tagged with cluster) ← Primary entry point
├── IAM Roles/Policies (tagged with cluster) 
└── S3 Buckets (tagged with cluster)

LEVEL 1 (VPC-SCOPED): Core Infrastructure  
├── Subnets → ALL subnets in cluster VPCs
├── Security Groups → ALL groups in cluster VPCs
├── Route Tables → ALL tables in cluster VPCs  
├── Internet/NAT Gateways → attached to cluster VPCs
└── VPC Endpoints → deployed in cluster VPCs

LEVEL 2 (SUBNET-SCOPED): Deployed Services
├── EC2 Instances → deployed in cluster subnets
├── RDS Instances → deployed in cluster subnets  
├── ELB/ALB → deployed in cluster subnets
├── EFS Mount Targets → deployed in cluster subnets
└── Network Interfaces → deployed in cluster subnets

LEVEL 3 (INSTANCE-SCOPED): Attached Resources
├── EBS Volumes → attached to cluster EC2 instances
├── EBS Snapshots → from cluster EBS volumes
├── Elastic IPs → attached to cluster instances/ENIs
└── Auto Scaling Groups → managing cluster instances

LEVEL 4 (SERVICE-SCOPED): Service Resources
├── Target Groups → attached to cluster ELBs
├── Launch Templates → used by cluster ASGs
├── EFS File Systems → mounted in cluster
└── CloudFormation Stacks → created for cluster
```

### Key Discovery Functions

**Level 0 Discovery (Entry Points):**
```bash
find_cluster_vpcs()           # Primary entry point - tagged VPCs
find_cluster_iam_roles()      # Global IAM roles with cluster name
find_cluster_iam_policies()   # Global IAM policies with cluster name
```

**Level 1 Discovery (VPC-Scoped):**
```bash
find_subnets_in_vpcs()                # ALL subnets in cluster VPCs
find_security_groups_in_vpcs()        # ALL security groups in cluster VPCs
find_route_tables_in_vpcs()           # ALL route tables in cluster VPCs
find_internet_gateways_for_vpcs()     # Gateways attached to cluster VPCs
find_nat_gateways_in_vpcs()           # NAT gateways in cluster VPCs
find_vpc_endpoints_in_vpcs()          # VPC endpoints in cluster VPCs
find_classic_elbs_in_vpcs()           # Classic ELBs deployed in cluster VPCs
find_application_elbs_in_vpcs()       # Application ELBs deployed in cluster VPCs
```

**Level 2 Discovery (Subnet-Scoped):**
```bash
find_network_interfaces_in_subnets()  # ALL ENIs in cluster subnets
find_ec2_instances_in_subnets()       # ALL EC2 instances in cluster subnets
find_rds_instances_in_subnets()       # RDS instances in cluster DB subnet groups
find_efs_mount_targets_in_subnets()   # EFS mount targets in cluster subnets
```

**Level 3 Discovery (Instance-Scoped):**
```bash
find_ebs_volumes_for_instances()      # Volumes attached to cluster instances
find_ebs_snapshots_for_volumes()      # Snapshots from cluster volumes
find_elastic_ips_for_instances()      # EIPs attached to cluster instances/ENIs
find_auto_scaling_groups_for_instances() # ASGs managing cluster instances
```

**Level 4 Discovery (Service-Scoped):**
```bash
find_target_groups_for_elbs()         # Target groups attached to cluster ELBs
find_launch_templates_for_asgs()      # Launch templates used by cluster ASGs
find_efs_filesystems_for_mount_targets() # EFS filesystems from mount targets
find_cloudformation_stacks_for_cluster() # CloudFormation stacks for cluster
find_ecr_repositories_for_cluster()   # ECR repositories for cluster
```

### Comprehensive Resource Coverage

**Networking Resources:**
- VPCs, Subnets, Security Groups, Route Tables
- Internet Gateways, NAT Gateways, VPC Endpoints
- Network Interfaces (including ELB-managed ENIs)
- Elastic IPs attached to instances or ENIs

**Compute Resources:**
- EC2 Instances in cluster subnets (regardless of tags)
- Auto Scaling Groups managing cluster instances
- Launch Templates used by cluster ASGs
- EBS Volumes attached to cluster instances
- EBS Snapshots from cluster volumes

**Load Balancing:**
- Classic Load Balancers deployed in cluster VPCs
- Application/Network Load Balancers deployed in cluster VPCs  
- Target Groups attached to cluster ELBs

**Storage Resources:**
- EFS File Systems mounted in cluster
- EFS Mount Targets deployed in cluster subnets
- RDS Instances in DB subnet groups using cluster subnets

**Service Resources:**
- CloudFormation Stacks containing cluster name
- ECR Repositories containing cluster name
- IAM Roles and Policies containing cluster name

### Usage Examples

**Basic Discovery:**
```bash
# Discover all resources for cluster (JSON output)
./bin/aws-find-resources ocp-02

# Verbose discovery with progress information
./bin/aws-find-resources ocp-03 --verbose

# Multi-region discovery
./bin/aws-find-resources eks-02 --all-regions
```

**Integration with aws-clean-resources:**
```bash
# Generate deletion file from comprehensive discovery
./bin/aws-find-resources ocp-01-mturansk-a3 > cluster-resources.json

# Process deletion file with relationship-aware cleanup
./bin/aws-clean-resources cluster-resources.json
```

### Output Format Compatibility

**JSON Structure:**
The script outputs AWS resources in the exact format expected by `aws-clean-resources`:

```json
{
  "REGION": "us-west-2",
  "EC2_INSTANCES": [...],
  "EBS_VOLUMES": [...],
  "CLASSIC_LOAD_BALANCERS": [...],
  "SUBNETS": [...],
  "SECURITY_GROUPS": [...],
  "NETWORK_INTERFACES": [...],
  ...
}
```

**Cluster Identification:**
Each resource includes cluster identification as the **last field**:
- **Tagged resources**: Shows the tag value containing the cluster ID
- **Relationship-discovered**: Shows descriptive identifiers like `vpc-relationship`, `subnet-relationship`, `instance-attachment`
- **Service-managed**: Shows `elb-managed`, `aws-managed` for service-created resources

### Project Integration

**Supported Regions:**
- `us-east-1`, `us-east-2`, `us-west-2` (primary regions)
- `eu-west-1`, `ap-southeast-1` (additional regions)

**Cluster Types Supported:**
- **OpenShift (OCP)**: Full resource discovery including Hive-created infrastructure
- **Amazon EKS**: Comprehensive CAPI resource discovery
- **HyperShift (HCP)**: Hosted control plane resource discovery

**Workflow Integration:**
```bash
# Complete cluster lifecycle management
1. ./bin/aws-find-resources cluster-name > resources.json
2. ./bin/aws-clean-resources resources.json
3. # All cluster resources deleted including orphaned ones
```

### Advantages Over Tag-Based Discovery

**Zero False Negatives:**
- Discovers ELB-managed network interfaces without cluster tags
- Finds AWS service-created resources with non-standard naming
- Locates orphaned resources through dependency relationships

**High Precision:**
- VPC-scoping prevents false positives from unrelated infrastructure
- Relationship following ensures discovered resources actually belong to cluster
- Structured hierarchy prevents over-broad discovery

**Operational Excellence:**
- Compatible with existing aws-clean-resources workflow
- Verbose mode provides clear discovery progress
- Comprehensive coverage eliminates manual resource hunting

### Relationship Discovery Success Cases

**Case 1: Orphaned ELB Network Interface**
- **Problem**: ELB deleted manually, network interface `eni-0c6c39d4ccb19ae47` remained without cluster tags
- **Tag-based discovery**: Missed the ENI completely
- **Relationship discovery**: Found ENI through subnet → network interface relationship
- **Result**: Successful subnet deletion after ENI cleanup

**Case 2: VPC Dependencies Without Tags**
- **Problem**: Route tables and security groups created by AWS services lacked cluster tags
- **Tag-based discovery**: Found 0 route tables, 0 security groups
- **Relationship discovery**: Found 1 route table, 2 security groups through VPC relationship
- **Result**: Complete VPC dependency cleanup

**Case 3: Service-Managed Resources**
- **Problem**: Classic ELB `a21c1256fc1a34d7fa093e492262d24d` had no cluster name in its identifier
- **Tag-based discovery**: Never found the ELB
- **Relationship discovery**: Found ELB through VPC deployment relationship
- **Result**: Proper load balancer cleanup before infrastructure deletion

This comprehensive relationship-based approach ensures **zero orphaned resources** and provides complete visibility into cluster AWS consumption without the limitations of tag-based discovery.