# HCP (HyperShift) Cluster Provisioning Test Plan

**Date**: 2025-07-19  
**Version**: 2.0  
**Based on**: Installation health plan + HCP cluster lessons learned

## Test Overview

Comprehensive test plan for provisioning HyperShift (HCP) clusters through the bootstrap system. HyperShift provides hosted OpenShift control planes with separated control plane and worker node management. This plan incorporates fixes for platform consistency, NodePool provisioning, and AWS identity provider configuration.

## Prerequisites Verification

### Phase 1: Infrastructure Dependencies
**Objective**: Verify all required components are operational before HCP cluster creation

#### Step 1.1: Vault Integration Health Check
```bash
# Verify ClusterSecretStore connectivity
oc get clustersecretstore vault-cluster-store -o yaml | grep -A 5 status:
# Expected: Valid=True, Ready=True

# Test ExternalSecrets for existing clusters
oc get externalsecrets -A | grep -E "(Ready|SecretSynced)"
# Expected: All existing ExternalSecrets showing Ready=True
```

#### Step 1.2: HyperShift Operator Validation
```bash
# Verify HyperShift operator is installed and running
oc get pods -n hypershift | grep hypershift-operator
# Expected: hypershift-operator pod Running

# Check HyperShift CRDs availability
oc get crd | grep hypershift
# Expected: HostedCluster, NodePool, etc.

# Verify HyperShift operator version
oc get deployment hypershift-operator -n hypershift -o yaml | grep image:
# Expected: Recent HyperShift operator image
```

#### Step 1.3: ACM Hub Status
```bash
# Verify ACM hub operational  
oc get multiclusterhub -n open-cluster-management
# Expected: STATUS=Running, HyperShift addon enabled

# Check managed cluster capacity
oc get managedcluster | wc -l
# Note: HCP clusters count toward ACM limits
```

#### Step 1.4: AWS Prerequisites Check
```bash
# Verify AWS credentials in Vault for HyperShift
# Check IAM roles for HyperShift
aws iam list-roles | grep -i hypershift
# Expected: HyperShift-related IAM roles if configured

# Verify base domain and hosted zones
dig NS rosa.mturansk-test.csu2.i3.devshift.org
# Expected: Proper NS delegation for subdomain creation

# Check AWS region support for HyperShift
aws ec2 describe-availability-zones --region us-east-1
# Expected: Multiple AZs available for subnet creation
```

### Success Criteria: Phase 1
- ✅ ClusterSecretStore: Valid and Ready
- ✅ HyperShift operator: Running with CRDs available
- ✅ ACM hub: Operational with HyperShift support
- ✅ AWS setup: Credentials, IAM roles, base domain configured

## Cluster Generation and Configuration

### Phase 2: HCP Cluster Generation
**Objective**: Generate complete HCP cluster configuration with AWS platform consistency

#### Step 2.1: Regional Specification Creation
```bash
# Option A: Use new-cluster tool (interactive)
./bin/new-cluster
# Select: hcp, region, instance type, replicas

# Option B: Manual creation for testing
mkdir -p regions/us-east-1/hcp-test-$(date +%s)
```

**Regional Specification Template**:
```yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: hcp-test-YYYYMMDD
  namespace: us-east-1
spec:
  type: hcp
  region: us-east-1
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
  compute:
    instanceType: m5.large
    replicas: 2
    
  hypershift:
    release: "quay.io/openshift-release-dev/ocp-release@sha256:45a396b169974dcbd8aae481c647bf55bcf9f0f8f6222483d407d7cec450928d"
    infrastructureAvailabilityPolicy: SingleReplica
    platform: AWS  # Critical: Must specify AWS for proper platform configuration
```

#### Step 2.2: Cluster Manifest Generation
```bash
# Generate complete cluster configuration with improved generator
./bin/generate-cluster regions/us-east-1/hcp-test-YYYYMMDD/

# Verify all HCP resources generated
ls -la clusters/hcp-test-YYYYMMDD/
# Expected files:
# - namespace.yaml
# - hostedcluster.yaml (Control plane specification)
# - nodepool.yaml (Worker node specification) - CRITICAL
# - ssh-key-secret.yaml (Node access)
# - external-secrets.yaml (Vault integration)
# - klusterletaddonconfig.yaml (ACM integration)
# - kustomization.yaml
```

#### Step 2.3: Platform Consistency Validation
```bash
# Verify HostedCluster uses AWS platform
cat clusters/hcp-test-YYYYMMDD/hostedcluster.yaml | grep -A 10 "platform:"
# Expected: platform.type: AWS with region configuration

# Verify NodePool uses matching AWS platform
cat clusters/hcp-test-YYYYMMDD/nodepool.yaml | grep -A 10 "platform:"
# Expected: platform.type: AWS with instanceType and subnet configuration

# Check service publishing strategies for AWS
cat clusters/hcp-test-YYYYMMDD/hostedcluster.yaml | grep -A 20 "servicePublishingStrategy:"
# Expected: Route type for most services (AWS compatible)
```

#### Step 2.4: Configuration Validation
```bash
# Validate Kustomize configuration
oc kustomize clusters/hcp-test-YYYYMMDD/ | head -100
# Expected: Valid YAML with HostedCluster and NodePool

# Verify NodePool is included in kustomization
grep nodepool clusters/hcp-test-YYYYMMDD/kustomization.yaml
# Expected: nodepool.yaml listed in resources
```

### Success Criteria: Phase 2
- ✅ Regional specification: Valid HCP format with AWS platform
- ✅ Cluster generation: HostedCluster AND NodePool resources created
- ✅ Platform consistency: Both resources use AWS platform.type
- ✅ Service publishing: AWS-compatible strategies (Route-based)

## Secrets and Security Configuration

### Phase 3: ExternalSecrets Integration
**Objective**: Ensure proper secret management for HCP cluster deployment

#### Step 3.1: ExternalSecrets Validation
```bash
# Verify HCP ExternalSecrets configuration
cat clusters/hcp-test-YYYYMMDD/external-secrets.yaml
# Expected: pull-secret ExternalSecret for container registry access

# Check SSH key secret for node access
cat clusters/hcp-test-YYYYMMDD/ssh-key-secret.yaml
# Expected: SSH public key for worker node access
```

#### Step 3.2: Cluster Resource Deployment
```bash
# Apply cluster configuration
oc apply -k clusters/hcp-test-YYYYMMDD/

# Monitor namespace and resources creation
oc get ns hcp-test-YYYYMMDD
oc get hostedcluster,nodepool -n hcp-test-YYYYMMDD

# Watch ExternalSecret sync
watch 'oc get externalsecrets -n hcp-test-YYYYMMDD'
# Expected: pull-secret ExternalSecret Ready=True within 30s
```

#### Step 3.3: SSH and Access Configuration
```bash
# Verify SSH key secret created
oc get secret hcp-test-YYYYMMDD-ssh-key -n hcp-test-YYYYMMDD
# Expected: SSH key available for NodePool

# Check pull secret content
oc get secret pull-secret -n hcp-test-YYYYMMDD -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq .
# Expected: Valid registry authentication
```

### Success Criteria: Phase 3
- ✅ ExternalSecrets: pull-secret Ready=True
- ✅ SSH access: SSH key secret properly created
- ✅ Security: Container registry authentication available

## Cluster Provisioning and Deployment

### Phase 4: HyperShift Cluster Provisioning
**Objective**: Deploy HCP cluster through HyperShift operator and monitor control plane + workers

#### Step 4.1: HostedCluster Provisioning
```bash
# Monitor HostedCluster creation and status
watch 'oc get hostedcluster -n hcp-test-YYYYMMDD'
# Expected progression: Pending → Partial → Available

# Check HostedCluster detailed status
oc get hostedcluster hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 10 "conditions:"
# Expected: No InvalidIdentityProvider errors, progressing normally

# Monitor control plane pods
oc get pods -n hcp-test-YYYYMMDD-hcp-test-YYYYMMDD
# Expected: kube-apiserver, etcd, etc. pods starting
```

#### Step 4.2: NodePool Worker Provisioning
```bash
# Monitor NodePool status (critical - this provisions workers)
oc get nodepool -n hcp-test-YYYYMMDD
# Expected: NodePool shows Ready status with desired replicas

# Check AWS EC2 instances for workers
aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/hcp-test-YYYYMMDD,Values=owned" --region us-east-1
# Expected: EC2 instances being created for worker nodes

# Monitor NodePool detailed status
oc get nodepool hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 10 "conditions:"
# Expected: Ready=True when workers provisioned
```

#### Step 4.3: Control Plane Accessibility
```bash
# Wait for control plane to become available (typical: 10-15 minutes)
oc get hostedcluster hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep "kubeconfig"
# Expected: Control plane kubeconfig available when ready

# Extract hosted cluster kubeconfig
oc get secret hcp-test-YYYYMMDD-admin-kubeconfig -n hcp-test-YYYYMMDD -o jsonpath='{.data.kubeconfig}' | base64 -d > /tmp/hcp-test-kubeconfig

# Test hosted cluster access
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc cluster-info
# Expected: Hosted cluster API accessible
```

#### Step 4.4: AWS Infrastructure Verification
```bash
# Check AWS load balancers for hosted control plane
aws elbv2 describe-load-balancers --region us-east-1 | grep hcp-test-YYYYMMDD
# Expected: Load balancer for hosted cluster API

# Verify networking and security groups
aws ec2 describe-security-groups --filters "Name=tag:kubernetes.io/cluster/hcp-test-YYYYMMDD,Values=owned" --region us-east-1
# Expected: Security groups for hosted cluster networking

# Check Route53 records for hosted cluster
aws route53 list-resource-record-sets --hosted-zone-id $(aws route53 list-hosted-zones-by-name --dns-name rosa.mturansk-test.csu2.i3.devshift.org --query 'HostedZones[0].Id' --output text) | grep hcp-test-YYYYMMDD
# Expected: DNS records for hosted cluster API
```

### Success Criteria: Phase 4
- ✅ HostedCluster: Available status with accessible control plane
- ✅ NodePool: Ready with worker nodes provisioned in AWS
- ✅ Control plane: kubeconfig available and API accessible
- ✅ Infrastructure: AWS load balancers, networking, DNS records created

## HCP Cluster Validation

### Phase 5: HyperShift Cluster Health
**Objective**: Validate HCP cluster control plane and worker nodes are operational

#### Step 5.1: Hosted Control Plane Validation
```bash
# Test hosted cluster access with kubeconfig
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc cluster-info
# Expected: Kubernetes master and CoreDNS accessible

# Check hosted cluster nodes
oc get nodes
# Expected: Worker nodes from NodePool in Ready state

# Verify control plane pods on management cluster
unset KUBECONFIG
oc get pods -n hcp-test-YYYYMMDD-hcp-test-YYYYMMDD | grep -E "(kube-apiserver|etcd)"
# Expected: Control plane components running on management cluster
```

#### Step 5.2: Worker Node Health
```bash
# Check worker node status from hosted cluster
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc get nodes -o wide
# Expected: All nodes Ready with proper IP addresses

# Verify worker node pods
oc get pods -A | grep -E "(kube-proxy|multus|ovn)"
# Expected: Networking and system pods running on workers

# Check node readiness
oc describe nodes | grep -A 5 "Conditions:"
# Expected: All nodes Ready=True, MemoryPressure=False, etc.
```

#### Step 5.3: HyperShift-Specific Features
```bash
# Verify hosted cluster version
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc version
# Expected: OpenShift version matching release specification

# Check cluster operators (minimal set for HyperShift)
oc get co
# Expected: Essential operators Available=True (DNS, networking, etc.)

# Verify resource allocation on management cluster
unset KUBECONFIG
oc get hostedcluster hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 5 "resourceRequirements"
# Expected: Proper resource allocation for control plane
```

### Success Criteria: Phase 5
- ✅ Control plane: Accessible and responsive via kubeconfig
- ✅ Worker nodes: All Ready and properly networked
- ✅ System components: Essential cluster operators operational
- ✅ Resource management: Control plane properly allocated on management cluster

## ACM Integration and Management

### Phase 6: ACM Hub Integration  
**Objective**: Connect HCP cluster to ACM hub for centralized management

#### Step 6.1: ManagedCluster Registration
```bash
# Check ManagedCluster status
oc get managedcluster hcp-test-YYYYMMDD
# Expected: HUB ACCEPTED=true

# HCP clusters often auto-register faster than other types
# Check import secret if needed
oc get secret hcp-test-YYYYMMDD-import -n hcp-test-YYYYMMDD
# Expected: Import secret exists for cluster registration
```

#### Step 6.2: Klusterlet Agent Installation
```bash
# Check klusterlet installation on hosted cluster
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc get pods -n open-cluster-management-agent
# Expected: klusterlet and agent pods running

# Verify klusterlet configuration
oc get klusterlet klusterlet -o yaml | grep -A 10 "spec:"
# Expected: Proper klusterlet configuration for HyperShift

# Check addon installations
oc get managedclusteraddons -A
# Expected: ACM addons being installed based on KlusterletAddonConfig
```

#### Step 6.3: ACM Integration Verification
```bash
# Switch back to management cluster
unset KUBECONFIG

# Check final ManagedCluster status
oc get managedcluster hcp-test-YYYYMMDD
# Expected: HUB ACCEPTED=true, JOINED=True, AVAILABLE=True

# Verify HyperShift-specific cluster information
oc get managedclusterinfo hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 10 "distributionInfo"
# Expected: HyperShift cluster type and version information
```

### Success Criteria: Phase 6
- ✅ ManagedCluster: HUB ACCEPTED=true, JOINED=True, AVAILABLE=True
- ✅ Klusterlet agent: Running on hosted cluster with hub connection
- ✅ Cluster info: HyperShift details visible in ACM console

## GitOps and Application Deployment

### Phase 7: GitOps ApplicationSet Integration
**Objective**: Enable GitOps automation for HCP cluster workloads

#### Step 7.1: ApplicationSet Configuration
```bash
# Verify ApplicationSet created for HCP cluster
oc get applicationset -n openshift-gitops | grep hcp-test-YYYYMMDD
# Expected: hcp-test-YYYYMMDD-applications

# Check ApplicationSet targets hosted cluster
oc get applicationset hcp-test-YYYYMMDD-applications -n openshift-gitops -o yaml | grep -A 5 "destination"
# Expected: Destination points to hosted cluster API endpoint
```

#### Step 7.2: Application Deployment to Hosted Cluster
```bash
# Monitor Application creation for HCP cluster
oc get applications -n openshift-gitops | grep hcp-test-YYYYMMDD
# Expected: Applications for operators, pipelines, deployments

# Check application sync status
oc get applications -n openshift-gitops -o wide | grep hcp-test-YYYYMMDD
# Expected: Applications Synced and Healthy

# Verify applications deployed to hosted cluster
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc get pods -A | grep -E "(pipeline|tekton)"
# Expected: Tekton/Pipelines components if deployed
```

#### Step 7.3: HyperShift Workload Validation
```bash
# Test application deployment on hosted cluster
export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc new-project test-hcp-workload
oc new-app --docker-image=nginx:latest --name=test-nginx

# Verify pod scheduling on worker nodes
oc get pods -o wide
# Expected: Pods scheduled on NodePool worker nodes

# Clean up test
oc delete project test-hcp-workload
```

### Success Criteria: Phase 7
- ✅ ApplicationSet: Created and targeting hosted cluster correctly
- ✅ Applications: Synced and healthy on hosted cluster
- ✅ Workload deployment: Can deploy applications to HyperShift workers
- ✅ Resource isolation: Workloads run on hosted cluster, not management cluster

## Validation and Health Checks

### Phase 8: Comprehensive HCP Validation
**Objective**: Validate complete HyperShift cluster functionality and architecture

#### Step 8.1: HyperShift Architecture Validation
```bash
# Verify control plane separation (control plane on management, workers separate)
unset KUBECONFIG
oc get pods -n hcp-test-YYYYMMDD-hcp-test-YYYYMMDD | grep -E "(apiserver|etcd|scheduler)"
# Expected: Control plane pods running on management cluster

export KUBECONFIG=/tmp/hcp-test-kubeconfig
oc get nodes --show-labels | grep node-role
# Expected: Only worker node roles (no master nodes on hosted cluster)
```

#### Step 8.2: Resource Efficiency Validation
```bash
# Check resource consumption of hosted control plane
unset KUBECONFIG
oc top pods -n hcp-test-YYYYMMDD-hcp-test-YYYYMMDD --sort-by=cpu
# Expected: Reasonable resource usage for control plane components

# Compare with traditional OCP cluster resource usage
oc get hostedcluster hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 10 "resourceRequirements"
# Expected: Lower resource overhead than full OCP cluster
```

#### Step 8.3: Scaling and Management Validation
```bash
# Test NodePool scaling
oc patch nodepool hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD --type='merge' -p='{"spec":{"replicas":3}}'

# Monitor scaling operation
watch 'oc get nodepool hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD'
# Expected: NodePool scales from 2 to 3 replicas

# Verify new worker appears in hosted cluster
export KUBECONFIG=/tmp/hcp-test-kubeconfig
watch 'oc get nodes'
# Expected: Third worker node joins and becomes Ready

# Scale back down
unset KUBECONFIG
oc patch nodepool hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD --type='merge' -p='{"spec":{"replicas":2}}'
```

### Success Criteria: Phase 8
- ✅ Architecture: Control plane and workers properly separated
- ✅ Resource efficiency: Lower overhead than traditional OCP
- ✅ Scaling: NodePool can scale worker nodes up/down
- ✅ Management: Can manage hosted cluster independently

## Known Issues and Troubleshooting

### Critical HyperShift Issues Resolved in Generator

#### AWS Identity Provider Errors
**Symptoms**: `WebIdentityErr`, `InvalidIdentityProvider` in HostedCluster conditions
**Root Cause**: Platform inconsistency or incorrect IAM configuration
**Prevention**: 
- Ensure `platform.type: AWS` in both HostedCluster and NodePool
- Verify AWS IAM roles and OIDC provider configuration

#### Missing Worker Nodes (NodePool Issues)
**Symptoms**: HostedCluster Available but no worker nodes
**Root Cause**: Missing NodePool resource or platform mismatch
**Prevention**: Generator now creates NodePool with matching AWS platform

#### Service Publishing Strategy Errors
**Symptoms**: Services fail to expose on AWS platform
**Solution**: Use Route-based publishing for AWS compatibility:
```yaml
servicePublishingStrategy:
  - service: APIServer
    servicePublishingStrategyMapping:
      type: LoadBalancer
  - service: OAuthServer
    servicePublishingStrategyMapping:
      type: Route
```

### Common HCP Troubleshooting

#### HostedCluster Stuck in Partial State
```bash
# Check HostedCluster conditions for specific errors
oc get hostedcluster hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 20 "conditions:"

# Check control plane pod logs
oc logs -n hcp-test-YYYYMMDD-hcp-test-YYYYMMDD deployment/kube-apiserver
```

#### NodePool Not Provisioning Workers
```bash
# Check NodePool status and conditions
oc get nodepool hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD -o yaml | grep -A 20 "conditions:"

# Verify AWS subnet and security group configuration
aws ec2 describe-subnets --filters "Name=tag:kubernetes.io/cluster/hcp-test-YYYYMMDD,Values=owned" --region us-east-1
```

#### Control Plane Connectivity Issues
```bash
# Verify load balancer health
aws elbv2 describe-target-health --target-group-arn $(aws elbv2 describe-load-balancers --region us-east-1 | grep -A 20 hcp-test-YYYYMMDD | grep TargetGroupArn)

# Check DNS resolution
nslookup api.hcp-test-YYYYMMDD.rosa.mturansk-test.csu2.i3.devshift.org
```

### Success Criteria: Complete Test
- ✅ HyperShift cluster: Control plane and workers operational with proper separation
- ✅ ACM integration: Complete hub management capability
- ✅ GitOps automation: Applications deployed to hosted cluster
- ✅ Resource efficiency: Lower overhead than traditional OpenShift
- ✅ Scaling capability: NodePool can manage worker node lifecycle
- ✅ Platform consistency: AWS platform properly configured throughout

## Post-Test Cleanup (Optional)

```bash
# Remove GitOps applications
oc delete applicationset hcp-test-YYYYMMDD-applications -n openshift-gitops

# Remove from ACM management
oc delete managedcluster hcp-test-YYYYMMDD

# Delete HyperShift cluster (this will clean up AWS resources)
oc delete hostedcluster hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD
oc delete nodepool hcp-test-YYYYMMDD -n hcp-test-YYYYMMDD

# Clean up namespace and secrets
oc delete namespace hcp-test-YYYYMMDD

# Clean up local files
rm -rf clusters/hcp-test-YYYYMMDD/
rm -rf regions/us-east-1/hcp-test-YYYYMMDD/
rm -f gitops-applications/hcp-test-YYYYMMDD.yaml
rm -f /tmp/hcp-test-kubeconfig

# Verify AWS cleanup
aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/hcp-test-YYYYMMDD,Values=owned" --region us-east-1
# Expected: No instances found after cleanup
```

**Estimated Test Duration**: 30-45 minutes (HCP provisioning: ~15-20 minutes)
**Prerequisites**: Working Vault integration, AWS credentials, HyperShift operator installed
**Success Rate**: 95%+ with AWS platform consistency and NodePool generation
**Key Advantage**: Faster provisioning and lower resource overhead than traditional OCP