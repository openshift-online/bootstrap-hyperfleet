# AWS Resource Discovery

## Functional Requirements for bin/find-aws-resources

### Primary Objective
Create comprehensive AWS CLI commands to discover all resources consumed by OpenShift clusters managed through this GitOps bootstrap project. Uses cluster configuration analysis to provide accurate resource discovery across all supported regions and instance types.

### Project Analysis Summary

Based on analysis of the current cluster configurations in this repository:

**Active Regions:**
- `us-east-1` (primary) - 3 OCP clusters (ocp-02, ocp-03, o1 HCP cluster)
- `us-west-2` - 1 OCP cluster (ocp-04)  
- `eu-west-1` - 1 OCP cluster (ocp-05)
- `ap-southeast-1` - 2 EKS clusters (eks-01, eks-02)

**Instance Types in Use:**
- `m5.xlarge` - OCP master and worker nodes (primary)
- `m5.large` - EKS worker nodes and some OCP worker nodes
- `c5.4xlarge` - OCP worker nodes (base template)

**Cluster Types:**
- **OpenShift (OCP)** - Uses Hive ClusterDeployment resources
- **Amazon EKS** - Uses CAPI AWSManagedControlPlane/AWSManagedMachinePool resources  
- **HyperShift (HCP)** - Hosted control plane clusters

### AWS Resource Discovery Commands

#### Core Discovery Strategy
Build on the proven two-pass discovery approach from `bin/clean-aws`:
1. **Tag-Based Search** - Find resources with cluster ID in tag values
2. **VPC-Based Search** - Find additional resources by VPC association

#### Complete Resource Discovery Commands

**1. EC2 Instances**
```bash
# Find all instances by cluster tag
aws ec2 describe-instances --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" "Name=instance-state-name,Values=running,stopped,stopping,pending" \
  --query 'Reservations[*].Instances[*].[InstanceId,InstanceType,State.Name,VpcId,SubnetId,PrivateIpAddress,Tags[?Key==`Name`].Value|[0]]' \
  --output table

# Find instances by specific instance types used in project
aws ec2 describe-instances --region $region \
  --filters "Name=instance-type,Values=m5.xlarge,m5.large,c5.4xlarge" "Name=instance-state-name,Values=running,stopped,stopping,pending" \
  --query 'Reservations[*].Instances[*].[InstanceId,InstanceType,State.Name,VpcId,Tags[?Key==`kubernetes.io/cluster`] | [0].Value || `no-cluster-tag`]' \
  --output table
```

**2. EBS Volumes and Snapshots**
```bash
# EBS Volumes (including root volumes for instances)
aws ec2 describe-volumes --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'Volumes[*].[VolumeId,Size,VolumeType,State,Encrypted,Attachments[0].InstanceId || `unattached`]' \
  --output table

# Find volumes by instance attachment (for untagged volumes)
aws ec2 describe-volumes --region $region \
  --filters "Name=attachment.instance-id,Values=${instance_id}" \
  --query 'Volumes[*].[VolumeId,Size,VolumeType,Iops,State]' \
  --output table

# EBS Snapshots  
aws ec2 describe-snapshots --region $region \
  --owner-ids self \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'Snapshots[*].[SnapshotId,VolumeSize,State,StartTime,Description]' \
  --output table
```

**3. Load Balancers (All Types)**
```bash
# Application/Network Load Balancers (ALB/NLB)
aws elbv2 describe-load-balancers --region $region \
  --query "LoadBalancers[?contains(LoadBalancerName, '${cluster_id}')].[LoadBalancerArn,LoadBalancerName,Type,State.Code,VpcId]" \
  --output table

# Classic Load Balancers
aws elb describe-load-balancers --region $region \
  --query "LoadBalancerDescriptions[?contains(LoadBalancerName, '${cluster_id}')].[LoadBalancerName,VPCId,Scheme,CreatedTime]" \
  --output table

# Target Groups
aws elbv2 describe-target-groups --region $region \
  --query "TargetGroups[?contains(TargetGroupName, '${cluster_id}')].[TargetGroupArn,TargetGroupName,Protocol,Port,VpcId]" \
  --output table
```

**4. Auto Scaling Groups**  
```bash
# Auto Scaling Groups (used by EKS managed node groups and OCP machine sets)
aws autoscaling describe-auto-scaling-groups --region $region \
  --query "AutoScalingGroups[?contains(AutoScalingGroupName, '${cluster_id}')].[AutoScalingGroupName,DesiredCapacity,MinSize,MaxSize,VPCZoneIdentifier]" \
  --output table

# Launch Templates (used by auto scaling groups)
aws ec2 describe-launch-templates --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'LaunchTemplates[*].[LaunchTemplateId,LaunchTemplateName,LatestVersionNumber,CreatedBy]' \
  --output table
```

**5. Networking Resources**
```bash
# VPCs
aws ec2 describe-vpcs --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'Vpcs[*].[VpcId,CidrBlock,State,IsDefault,Tags[?Key==`Name`].Value|[0]]' \
  --output table

# Subnets 
aws ec2 describe-subnets --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'Subnets[*].[SubnetId,VpcId,CidrBlock,AvailabilityZone,MapPublicIpOnLaunch,Tags[?Key==`Name`].Value|[0]]' \
  --output table

# Internet Gateways
aws ec2 describe-internet-gateways --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'InternetGateways[*].[InternetGatewayId,Attachments[0].VpcId,Attachments[0].State,Tags[?Key==`Name`].Value|[0]]' \
  --output table

# NAT Gateways
aws ec2 describe-nat-gateways --region $region \
  --filter "Name=tag-value,Values=*${cluster_id}*" \
  --query 'NatGateways[*].[NatGatewayId,VpcId,SubnetId,State,NatGatewayAddresses[0].PublicIp]' \
  --output table

# Route Tables
aws ec2 describe-route-tables --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'RouteTables[*].[RouteTableId,VpcId,Associations[0].Main || `false`,Tags[?Key==`Name`].Value|[0]]' \
  --output table

# Security Groups
aws ec2 describe-security-groups --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'SecurityGroups[*].[GroupId,GroupName,VpcId,Description]' \
  --output table

# Network ACLs
aws ec2 describe-network-acls --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'NetworkAcls[*].[NetworkAclId,VpcId,IsDefault,Associations[0].SubnetId || `none`]' \
  --output table

# Network Interfaces (ENIs)
aws ec2 describe-network-interfaces --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'NetworkInterfaces[*].[NetworkInterfaceId,InterfaceType,Status,VpcId,SubnetId,PrivateIpAddress,Attachment.InstanceId || `unattached`]' \
  --output table

# VPC Endpoints
aws ec2 describe-vpc-endpoints --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'VpcEndpoints[*].[VpcEndpointId,VpcId,ServiceName,State,VpcEndpointType]' \
  --output table

# Elastic IPs
aws ec2 describe-addresses --region $region \
  --filters "Name=tag-value,Values=*${cluster_id}*" \
  --query 'Addresses[*].[AllocationId,PublicIp,InstanceId || `unassociated`,NetworkInterfaceId || `none`,Domain]' \
  --output table
```

**6. Storage Resources**
```bash
# EFS File Systems (if used for persistent storage)
aws efs describe-file-systems --region $region \
  --query "FileSystems[?contains(Name, '${cluster_id}')].[FileSystemId,Name,LifeCycleState,SizeInBytes.Value,CreationTime]" \
  --output table

# EFS Mount Targets  
aws efs describe-mount-targets --region $region \
  --query 'MountTargets[*].[MountTargetId,FileSystemId,SubnetId,LifeCycleState,IpAddress]' \
  --output table
```

**7. Database Resources**
```bash
# RDS Instances 
aws rds describe-db-instances --region $region \
  --query "DBInstances[?contains(DBInstanceIdentifier, '${cluster_id}')].[DBInstanceIdentifier,DBInstanceStatus,DBInstanceClass,Engine,DBSubnetGroup.VpcId]" \
  --output table

# RDS Cluster (Aurora)
aws rds describe-db-clusters --region $region \
  --query "DBClusters[?contains(DBClusterIdentifier, '${cluster_id}')].[DBClusterIdentifier,Status,Engine,VpcSecurityGroups[0].VpcId]" \
  --output table
```

**8. Container and Kubernetes Resources**
```bash
# EKS Clusters
aws eks describe-cluster --region $region --name ${cluster_id} \
  --query 'cluster.[name,status,version,platformVersion,endpoint,resourcesVpcConfig.vpcId]' \
  --output table

# EKS Node Groups
aws eks describe-nodegroup --region $region --cluster-name ${cluster_id} --nodegroup-name ${nodegroup_name} \
  --query 'nodegroup.[nodegroupName,status,instanceTypes[0],scalingConfig,subnets]' \
  --output table

# ECR Repositories (if using private registries)
aws ecr describe-repositories --region $region \
  --query "repositories[?contains(repositoryName, '${cluster_id}')].[repositoryName,repositoryUri,createdAt]" \
  --output table
```

**9. IAM Resources**
```bash
# IAM Roles
aws iam list-roles \
  --query "Roles[?contains(RoleName, '${cluster_id}')].[RoleName,Arn,CreateDate,Description]" \
  --output table

# IAM Policies  
aws iam list-policies --scope Local \
  --query "Policies[?contains(PolicyName, '${cluster_id}')].[PolicyName,Arn,CreateDate,Description]" \
  --output table

# IAM Instance Profiles
aws iam list-instance-profiles \
  --query "InstanceProfiles[?contains(InstanceProfileName, '${cluster_id}')].[InstanceProfileName,Arn,CreateDate]" \
  --output table
```

**10. CloudFormation Stacks**
```bash
# CloudFormation Stacks (used by OpenShift installer and EKS)
aws cloudformation describe-stacks --region $region \
  --query "Stacks[?contains(StackName, '${cluster_id}')].[StackName,StackStatus,CreationTime,Description]" \
  --output table

# Stack Resources
aws cloudformation describe-stack-resources --region $region --stack-name ${stack_name} \
  --query 'StackResources[*].[LogicalResourceId,PhysicalResourceId,ResourceType,ResourceStatus]' \
  --output table
```

### Resource Discovery Execution Strategy

**Multi-Region Discovery Script:**
```bash
#!/bin/bash
# Example usage: ./discover-cluster-resources.sh ocp-02

CLUSTER_ID=${1:-"mt-test"}
REGIONS=("us-east-1" "us-west-2" "eu-west-1" "ap-southeast-1")

echo "=== AWS Resource Discovery for Cluster: $CLUSTER_ID ==="
echo ""

for region in "${REGIONS[@]}"; do
    echo "Region: $region"
    echo "=================="
    
    # Run discovery commands for each resource type
    # [Include all the commands above with proper error handling]
    
    echo ""
done
```

### Project-Specific Discovery Patterns

**Cluster Naming Conventions:**
- OpenShift clusters: `ocp-02`, `ocp-03`, `ocp-04`, `ocp-05`
- EKS clusters: `eks-01`, `eks-02`
- HyperShift clusters: `hcp-01`
- Legacy pattern: `cluster-XX` (in backup configurations)

**Common Tag Patterns:**
- `kubernetes.io/cluster/${cluster_name}=owned`
- `kubernetes.io/cluster/${cluster_name}=shared`
- `Name=${cluster_name}-*`
- `sigs.k8s.io/cluster-api-provider-aws/cluster/${cluster_name}=owned` (for CAPI EKS)

**Resource Naming Patterns:**
- Load Balancers: `${cluster_name}-*-elb`
- Auto Scaling Groups: `${cluster_name}-*-asg`
- Launch Templates: `${cluster_name}-*-template`
- Security Groups: `${cluster_name}-*-sg`

### Integration with Existing Tools

**Extends bin/clean-aws:**
- Uses the same two-pass discovery strategy
- Adds resource types not covered in cleanup (EFS, ECR, CloudFormation)
- Provides detailed resource information vs. cleanup focus
- Compatible with the same cluster ID and region inputs

**Usage Patterns:**
```bash
# Discover resources for specific cluster
./bin/find-aws-resources.sh ocp-02

# Discover across all regions for cluster
./bin/find-aws-resources.sh eks-02 --all-regions

# Discover with detailed output
./bin/find-aws-resources.sh ocp-03 --verbose

# Output to file for analysis
./bin/find-aws-resources.sh ocp-04 > cluster-resources.txt
```

### Expected Resource Footprint per Cluster

**OpenShift (OCP) Cluster:**
- 3 EC2 instances (masters) + 1-3 worker instances
- 4-6 EBS volumes (root + data volumes)
- 1 VPC with 6 subnets (3 public, 3 private)
- 2 load balancers (API + ingress)
- 1 NAT gateway per AZ
- 5-8 security groups
- 2-3 route tables
- Multiple network interfaces
- CloudFormation stacks for infrastructure

**EKS Cluster:**
- 1 EKS control plane (managed)
- 1-3 managed node group instances
- 1-3 EBS volumes (worker node storage)
- 1 VPC with 4-6 subnets
- 1-2 load balancers (ingress)
- Auto scaling group and launch template
- 3-5 security groups
- IAM roles and policies for node groups

This comprehensive discovery approach ensures complete visibility into AWS resource consumption across all cluster types and regions used in this OpenShift bootstrap project.