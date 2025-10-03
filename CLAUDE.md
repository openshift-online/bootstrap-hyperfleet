# CLAUDE.md - Project Overview

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Red Hat OpenShift bootstrap repository that contains GitOps infrastructure for deploying and managing OpenShift clusters across multiple regions. The project uses OpenShift GitOps (ArgoCD), Red Hat Advanced Cluster Management (ACM), and Hive for cluster lifecycle management.

## Architecture

The codebase is organized into several key components:

### Core Components
- **Bootstrap Control Plane**: Uses OpenShift GitOps to manage the initial cluster setup
- **Cluster Provisioning**: Uses CAPI (Cluster API) for automated cluster creation
- **Regional Management**: Uses ACM for multi-cluster management across regions
- **Configuration Management**: Pure Kustomize-based approach for generating cluster manifests

### Directory Structure

The project uses semantic directory organization with consistent patterns:

**Top-level "things":**
- `clusters/`: Cluster deployment configurations (auto-generated from regions/)
- `operators/`: Operator/application deployments following {operator-name}/{deployment-target} pattern
- `pipelines/`: Tekton pipeline configurations following {pipeline-name}/{cluster-name} pattern
- `deployments/`: Service deployments following {service-name}/{cluster-name} pattern
- `regions/`: Regional cluster specifications (input for generation)
- `bases/`: Reusable Kustomize base components
- `gitops-applications/`: ArgoCD ApplicationSets for GitOps automation
- `prereqs/`: Prerequisites for bootstrap process

**Operator deployments organized semantically:**
- `operators/advanced-cluster-management/global/`: ACM hub cluster deployment
- `operators/gitops-integration/global/`: GitOps integration policies and configurations
- `operators/openshift-pipelines/global/`: Pipelines hub cluster deployment  
- `operators/openshift-pipelines/{cluster-name}/`: Tekton Pipelines operator per managed cluster
- `operators/vault/global/`: Vault secret management system

**Deployment targets:**
- `global/`: Hub cluster deployments (shared infrastructure)
- `{cluster-name}/`: Managed cluster-specific deployments (e.g., `ocp-02/`, `eks-01/`)

## Key Technologies

- **OpenShift GitOps (ArgoCD)**: Continuous deployment and cluster management
- **Red Hat Advanced Cluster Management (ACM)**: Multi-cluster management with CAPI integration
- **Cluster API (CAPI)**: Kubernetes-native cluster lifecycle management
- **Hive**: OpenShift cluster provisioning operator (for OCP clusters)
- **Infrastructure Providers**: AWS, Azure, GCP, vSphere, OpenStack, BareMetal (via ACM)
- **Kustomize**: YAML configuration management and templating
- **Tekton Pipelines**: CI/CD workflows

## Claude Memories

- Don't run `bin/bootstrap` from a Claude session
- When provisioning or managing OpenShift, always use `oc` client
- Critical! Always use smart semantic naming for maximum usability and comprehensive

## AWS Infrastructure Knowledge

### DNS and Hosted Zones
- **Primary Domain**: `bootstrap.red-chesterfield.com` (public hosted zone ID: Z10440443GZJQIRRN54G5)
- **Parent Domain**: `red-chesterfield.com` is in another AWS account with NS delegation to our child account
- **DNS Resolution**: Public hosted zone resolves via NS delegation from parent domain
- **Previous Issue**: Bootstrap failures were caused by VPC DNS configuration, not missing DNS records
- **Critical**: Always ensure VPC has `enableDnsSupport=true` and `enableDnsHostnames=true` for proper Route53 resolution

### Current US-West-2 Subnets
**VPC**: `vpc-02f36017bd9f8e468` (10.0.0.0/16)
- **Public Subnet 1**: `subnet-0a9938c38050df215` (10.0.7.0/24 in us-west-2b)
- **Public Subnet 2**: `subnet-00e750cbfb46344f4` (10.0.8.0/24 in us-west-2a) 
- **Private Subnet**: `subnet-0243a19e2c2589b0b` (10.0.6.0/24 in us-west-2a)

**Infrastructure Configuration**:
- Internet Gateway: `igw-083792a85a29601af` (attached to VPC)
- NAT Gateway: `nat-0e4f1428c36190422` (in public subnet for private subnet internet access)
- Public Route Table: `rtb-01e5bcb606c7c7f89` (routes 0.0.0.0/0 to IGW)
- Private Route Table: `rtb-01967c73fd21b58f3` (routes 0.0.0.0/0 to NAT Gateway)
- All public subnets have auto-assign public IP enabled
- VPC DNS settings properly configured for Route53 resolution

**OpenShift Requirements**:
- Multi-AZ: Public subnets in multiple availability zones (production)
- Single-AZ: One public + one private subnet (testing/development)
- Public subnets must have internet gateway routing
- Private subnets for master/worker nodes (**CRITICAL: MUST have NAT Gateway for internet access**)
- **Bootstrap Failure**: Without NAT Gateway, bootstrap times out waiting for control plane

### Subnet Creation Scripts

**Create VPC and Subnets for OpenShift:**
```bash
# Variables
VPC_ID="vpc-02f36017bd9f8e468"  # Use existing or create new
REGION="us-west-2"

# Step 1: Enable VPC DNS settings (required for Route53 resolution)
aws ec2 modify-vpc-attribute --region $REGION --vpc-id $VPC_ID --enable-dns-support
aws ec2 modify-vpc-attribute --region $REGION --vpc-id $VPC_ID --enable-dns-hostnames

# Step 2: Check for existing internet gateway
IGW_ID=$(aws ec2 describe-internet-gateways --region $REGION \
  --filters "Name=attachment.vpc-id,Values=$VPC_ID" \
  --query "InternetGateways[0].InternetGatewayId" --output text)

# Create IGW if none exists
if [ "$IGW_ID" = "None" ] || [ "$IGW_ID" = "null" ]; then
  IGW_ID=$(aws ec2 create-internet-gateway --region $REGION \
    --tag-specifications "ResourceType=internet-gateway,Tags=[{Key=Name,Value=bootstrap-igw}]" \
    --query 'InternetGateway.InternetGatewayId' --output text)
  aws ec2 attach-internet-gateway --region $REGION --vpc-id $VPC_ID --internet-gateway-id $IGW_ID
fi

# Step 3: Check existing subnets to avoid CIDR conflicts
aws ec2 describe-subnets --region $REGION --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "Subnets[].CidrBlock" --output table

# Step 4: Create subnets with available CIDR blocks
PRIVATE_SUBNET=$(aws ec2 create-subnet --region $REGION \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.6.0/24 \
  --availability-zone ${REGION}a \
  --tag-specifications "ResourceType=subnet,Tags=[{Key=Name,Value=bootstrap-private-${REGION}a},{Key=Type,Value=Private}]" \
  --query 'Subnet.SubnetId' --output text)

PUBLIC_SUBNET_B=$(aws ec2 create-subnet --region $REGION \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.7.0/24 \
  --availability-zone ${REGION}b \
  --tag-specifications "ResourceType=subnet,Tags=[{Key=Name,Value=bootstrap-public-${REGION}b},{Key=Type,Value=Public}]" \
  --query 'Subnet.SubnetId' --output text)

PUBLIC_SUBNET_A=$(aws ec2 create-subnet --region $REGION \
  --vpc-id $VPC_ID \
  --cidr-block 10.0.8.0/24 \
  --availability-zone ${REGION}a \
  --tag-specifications "ResourceType=subnet,Tags=[{Key=Name,Value=bootstrap-public-${REGION}a},{Key=Type,Value=Public}]" \
  --query 'Subnet.SubnetId' --output text)

# Step 5: Create public route table
PUBLIC_RT=$(aws ec2 create-route-table --region $REGION --vpc-id $VPC_ID \
  --tag-specifications "ResourceType=route-table,Tags=[{Key=Name,Value=bootstrap-public-rt}]" \
  --query 'RouteTable.RouteTableId' --output text)

# Step 6: Add internet route and associate public subnets
aws ec2 create-route --region $REGION --route-table-id $PUBLIC_RT \
  --destination-cidr-block 0.0.0.0/0 --gateway-id $IGW_ID

aws ec2 associate-route-table --region $REGION \
  --subnet-id $PUBLIC_SUBNET_A --route-table-id $PUBLIC_RT
aws ec2 associate-route-table --region $REGION \
  --subnet-id $PUBLIC_SUBNET_B --route-table-id $PUBLIC_RT

# Step 7: Enable auto-assign public IP for public subnets
aws ec2 modify-subnet-attribute --region $REGION \
  --subnet-id $PUBLIC_SUBNET_A --map-public-ip-on-launch
aws ec2 modify-subnet-attribute --region $REGION \
  --subnet-id $PUBLIC_SUBNET_B --map-public-ip-on-launch

# Step 8: Create NAT Gateway (CRITICAL for OpenShift bootstrap)
EIP_ALLOC=$(aws ec2 allocate-address --region $REGION --domain vpc \
  --tag-specifications "ResourceType=elastic-ip,Tags=[{Key=Name,Value=openshift-nat-eip}]" \
  --query 'AllocationId' --output text)

NAT_GW=$(aws ec2 create-nat-gateway --region $REGION \
  --subnet-id $PUBLIC_SUBNET_A \
  --allocation-id $EIP_ALLOC \
  --tag-specifications "ResourceType=natgateway,Tags=[{Key=Name,Value=openshift-nat-gw}]" \
  --query 'NatGateway.NatGatewayId' --output text)

# Wait for NAT Gateway to be available
aws ec2 wait nat-gateway-available --region $REGION --nat-gateway-ids $NAT_GW

# Step 9: Create private route table with NAT Gateway route
PRIVATE_RT=$(aws ec2 create-route-table --region $REGION --vpc-id $VPC_ID \
  --tag-specifications "ResourceType=route-table,Tags=[{Key=Name,Value=openshift-private-rt}]" \
  --query 'RouteTable.RouteTableId' --output text)

aws ec2 create-route --region $REGION --route-table-id $PRIVATE_RT \
  --destination-cidr-block 0.0.0.0/0 --nat-gateway-id $NAT_GW

aws ec2 associate-route-table --region $REGION \
  --subnet-id $PRIVATE_SUBNET --route-table-id $PRIVATE_RT

echo "OpenShift VPC Setup Complete:"
echo "  Private: $PRIVATE_SUBNET (10.0.6.0/24 in ${REGION}a)"
echo "  Public:  $PUBLIC_SUBNET_A (10.0.8.0/24 in ${REGION}a)"
echo "  Public:  $PUBLIC_SUBNET_B (10.0.7.0/24 in ${REGION}b)"
echo "  NAT Gateway: $NAT_GW (Elastic IP: $EIP_ALLOC)"
```

**Update install-config.yaml:**

**Multi-AZ (Production):**
```yaml
platform:
  aws:
    region: us-west-2
    subnets:
      - subnet-0a9938c38050df215  # Public subnet us-west-2b
      - subnet-00e750cbfb46344f4  # Public subnet us-west-2a
      - subnet-0243a19e2c2589b0b  # Private subnet us-west-2a
```

**Single-AZ (Testing/Development - Cost Optimized):**
```yaml
platform:
  aws:
    region: us-west-2
    subnets:
      - subnet-00e750cbfb46344f4  # Public subnet us-west-2a
      - subnet-0243a19e2c2589b0b  # Private subnet us-west-2a
```

**Single-AZ Benefits:**
- Cost savings: No cross-AZ data transfer charges
- Simplified networking: Only one AZ to manage
- Faster deployment: Fewer resources to create
- Adequate for testing and development workloads

**Single-AZ Limitations:**
- No high availability across AZs
- AWS Load Balancer Controller requires multi-AZ
- Some AWS services prefer multi-AZ deployment

## Complete Working VPC Configuration (Corrected)

**IMPORTANT: Use mturansk-vpc for OpenShift deployments, NOT ROSA VPCs**

**Current Production Configuration:**
- **VPC**: `vpc-0502e366b9ea976b0` (mturansk-vpc, 10.0.0.0/20)
- **DNS**: `enableDnsSupport=true`, `enableDnsHostnames=true` ✓
- **Internet Gateway**: `igw-0a57132a1769d9115` ✓
- **NAT Gateway**: `nat-0626f94c257ac8156` (Public IP: `44.242.28.216`) ✓

**Working Subnets:**
- **Public us-west-2a**: `subnet-06720003300d9b5c2` (10.0.0.0/24) → IGW, Auto-assign public IP ✓
- **Private us-west-2a**: `subnet-0e2d3bfa9373adb69` (10.0.1.0/24) → NAT Gateway
- **Public us-west-2b**: `subnet-096c3e5793294c6d5` (10.0.2.0/24) → IGW, Auto-assign public IP ✓  
- **Private us-west-2b**: `subnet-02a2393848aa586c5` (10.0.3.0/24) → NAT Gateway

**Route Tables:**
- **Public RT**: `rtb-02602bcd0de9da0d5` (0.0.0.0/0 → IGW)
- **Private RT**: `rtb-0daf5b09c485f8d0b` (0.0.0.0/0 → NAT Gateway)

**Verified Working install-config.yaml:**
```yaml
platform:
  aws:
    region: us-west-2
    subnets:
    - subnet-06720003300d9b5c2  # Public subnet us-west-2a
    - subnet-0e2d3bfa9373adb69  # Private subnet us-west-2a  
    - subnet-096c3e5793294c6d5  # Public subnet us-west-2b
    - subnet-02a2393848aa586c5  # Private subnet us-west-2b
```

**Critical**: Previous issues were caused by using ROSA HCP VPC (`vpc-02f36017bd9f8e468`) instead of dedicated OpenShift VPC.

## SRE Tool Categories

**Cluster Operations** (cluster-*):
- `cluster-create` - Generate new cluster configurations
- `cluster-remove` - Clean cluster removal
- `cluster-convert` - Convert cluster types
- `cluster-list` - List available clusters
- `cluster-status` - Compare ACM vs repository state
- `cluster-regenerate-all` - Update all cluster configurations

**AWS Resource Management** (aws-*):
- `aws-find-resources` - Discover AWS resources for specific cluster
- `aws-find-all-resources` - Comprehensive resource discovery with orphan detection
- `aws-clean-resources` - Clean up AWS resources
- `aws-test-find-resources` - Test resource discovery functionality

**Monitoring & Health** (monitor-*):
- `monitor-health` - Comprehensive cluster health checks
- `monitor-status` - Overall environment status

**Documentation** (docs-*):
- `docs-generate` - Generate documentation
- `docs-validate` - Validate documentation consistency
- `docs-update` - Update dynamic documentation

**Bootstrap Operations**:
- `bootstrap` - Initial environment setup
- `bootstrap-vault` - Vault integration setup

## Documentation Navigation

For detailed information, see:
- **[Architecture](./docs/architecture/ARCHITECTURE.md)** - Visual diagrams and technical architecture
- **[Installation](./docs/getting-started/production-installation.md)** - Complete setup guide
- **[Cluster Creation](./guides/cluster-creation.md)** - End-to-end cluster deployment
- **[Monitoring](./guides/monitoring.md)** - Status checking and troubleshooting
- **[Documentation Index](./docs/INDEX.md)** - Complete documentation reference