# EKS Cluster Provisioning Test Plan

**Date**: 2025-07-19  
**Version**: 2.0  
**Based on**: Successful eks-01-mturansk-test deployment + generator improvements

## Test Overview

Comprehensive test plan for provisioning AWS EKS clusters using CAPI (Cluster API) through the bootstrap system. This plan incorporates all lessons learned from the first successful EKS deployment and includes automated fixes for known issues.

## Prerequisites Verification

### Phase 1: Infrastructure Dependencies
**Objective**: Verify all required components are operational before EKS cluster creation

#### Step 1.1: Vault Integration Health Check
```bash
# Verify ClusterSecretStore connectivity
oc get clustersecretstore -A
# Expected: vault-cluster-store STATUS=Valid, READY=True

# Test secret access
oc get externalsecrets -A | grep -E "(Ready|SecretSynced)"
# Expected: All existing ExternalSecrets show Ready=True
```

#### Step 1.2: CAPI Provider Validation  
```bash
# Verify CAPI controllers are running
oc get pods -n multicluster-engine | grep -E "(capi|capa)"
# Expected: 
# - capi-controller-manager: Running
# - capa-controller-manager: Running (AWS provider)

# Check CAPI CRD availability
oc get crd | grep -E "(cluster\.x-k8s\.io|infrastructure\.cluster\.x-k8s\.io)"
# Expected: Core CAPI and AWS provider CRDs available
```

#### Step 1.3: ACM Hub Status
```bash
# Verify ACM hub can accept new managed clusters
oc get multiclusterhub -n open-cluster-management
# Expected: STATUS=Running, VERSION=2.13.3+

# Check available cluster capacity
oc get managedcluster | wc -l
# Note: Track current cluster count for quota planning
```

#### Step 1.4: AWS Prerequisites Check
```bash
# Verify AWS credentials are available in Vault
# Check Elastic IP quota (critical for EKS)
aws ec2 describe-account-attributes --attribute-names supported-platforms --region us-east-2
aws service-quotas get-service-quota --service-code ec2 --quota-code L-0263D0A3 --region us-east-2
# Expected: At least 5 available Elastic IPs (3 needed per cluster)
```

### Success Criteria: Phase 1
- ✅ ClusterSecretStore: Valid and Ready
- ✅ CAPI providers: Running and healthy  
- ✅ ACM hub: Operational and accepting clusters
- ✅ AWS quotas: Sufficient Elastic IPs available

## Cluster Generation and Configuration

### Phase 2: EKS Cluster Generation
**Objective**: Generate complete EKS cluster configuration using improved generators

#### Step 2.1: Regional Specification Creation
```bash
# Option A: Use new-cluster tool (interactive)
./bin/cluster-create
# Select: eks, region, instance type, replicas

# Option B: Manual creation for testing
mkdir -p regions/us-west-2/eks-test-$(date +%s)
```

**Regional Specification Template**:
```yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: eks-test-YYYYMMDD
  namespace: us-west-2
spec:
  type: eks
  region: us-west-2
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
  compute:
    instanceType: m5.large
    replicas: 3
    
  kubernetes:
    version: "1.28"
```

#### Step 2.2: Cluster Manifest Generation
```bash
# Generate complete cluster configuration
./bin/cluster-generate regions/us-west-2/eks-test-YYYYMMDD/

# Verify all required resources generated
ls -la clusters/eks-test-YYYYMMDD/
# Expected files:
# - namespace.yaml
# - cluster.yaml (CAPI Cluster)
# - awsmanagedcontrolplane.yaml
# - awsmanagedmachinepool.yaml
# - machinepool.yaml (CRITICAL: Links CAPI to workers)
# - managedcluster.yaml
# - klusterletaddonconfig.yaml  
# - external-secrets.yaml
# - klusterlet-crd.yaml (NEW: For ACM integration)
# - kustomization.yaml
```

#### Step 2.3: Configuration Validation
```bash
# Validate Kustomize configuration
oc kustomize clusters/eks-test-YYYYMMDD/ | head -50
# Expected: Valid YAML output with all resources

# Check for critical improvements
grep -A 5 "kind: MachinePool" clusters/eks-test-YYYYMMDD/machinepool.yaml
# Expected: MachinePool resource with semantic version (1.28.0)

grep "version:" clusters/eks-test-YYYYMMDD/awsmanagedcontrolplane.yaml
# Expected: EKS API format version (v1.28)
```

### Success Criteria: Phase 2
- ✅ Regional specification: Valid format with all required fields
- ✅ Cluster generation: All critical resources created including MachinePool
- ✅ Version formatting: Semantic versioning for CAPI, EKS format for control plane
- ✅ Kustomization: Validates without errors

## Secrets and Security Configuration

### Phase 3: ExternalSecrets Integration
**Objective**: Ensure proper secret management for EKS cluster

#### Step 3.1: ExternalSecrets Validation
```bash
# Verify ExternalSecrets configuration exists
cat clusters/eks-test-YYYYMMDD/external-secrets.yaml
# Expected: aws-credentials and pull-secret ExternalSecrets

# Check Vault secret paths
# aws-credentials: secret/data/aws/credentials  
# pull-secret: secret/data/clusters/pull-secret
```

#### Step 3.2: Cluster Resource Deployment
```bash
# Apply cluster configuration
oc apply -k clusters/eks-test-YYYYMMDD/

# Monitor ExternalSecret sync
watch 'oc get externalsecrets -n eks-test-YYYYMMDD'
# Expected: Both ExternalSecrets transition to Ready=True within 30s
```

#### Step 3.3: Secret Verification
```bash
# Verify secrets created successfully
oc get secrets -n eks-test-YYYYMMDD | grep -E "(aws-credentials|pull-secret)"
# Expected: Both secrets present with data

# Check secret content (non-sensitive verification)
oc get secret aws-credentials -n eks-test-YYYYMMDD -o jsonpath='{.data}' | jq 'keys'
# Expected: ["credentials"] key present
```

### Success Criteria: Phase 3
- ✅ ExternalSecrets: Both aws-credentials and pull-secret Ready=True
- ✅ Secret content: Properly formatted and accessible
- ✅ Vault integration: No authentication errors

## Cluster Provisioning and Deployment

### Phase 4: CAPI Cluster Provisioning  
**Objective**: Deploy EKS cluster through CAPI and monitor provisioning

#### Step 4.1: Initial Provisioning Status
```bash
# Monitor CAPI cluster creation
watch 'oc get cluster -n eks-test-YYYYMMDD'
# Expected progression: Pending → Provisioning → Provisioned

# Check AWS resources being created
oc get awsmanagedcontrolplane -n eks-test-YYYYMMDD -o yaml | grep -A 10 status
# Monitor: EKS cluster creation in AWS
```

#### Step 4.2: Control Plane Readiness
```bash
# Wait for EKS control plane to become ACTIVE
# Typical time: 10-15 minutes
aws eks describe-cluster --name eks-test-YYYYMMDD --region us-west-2 --query 'cluster.status'
# Expected: "ACTIVE"

# Verify control plane endpoint
aws eks describe-cluster --name eks-test-YYYYMMDD --region us-west-2 --query 'cluster.endpoint'
# Expected: HTTPS URL with unique identifier
```

#### Step 4.3: Worker Node Provisioning
```bash
# Monitor MachinePool and AWSManagedMachinePool
oc get machinepool -n eks-test-YYYYMMDD
oc get awsmanagedmachinepool -n eks-test-YYYYMMDD

# Check AWS EKS node group
aws eks describe-nodegroup --cluster-name eks-test-YYYYMMDD --nodegroup-name eks-test-YYYYMMDD --region us-west-2 --query 'nodegroup.status'
# Expected: "ACTIVE"

# Verify worker nodes are Ready
export KUBECONFIG=/tmp/eks-test-kubeconfig
aws eks update-kubeconfig --name eks-test-YYYYMMDD --region us-west-2
kubectl get nodes
# Expected: 3 nodes in Ready state
```

### Step 4.4: Network Infrastructure Verification
```bash
# Check VPC and networking components
aws ec2 describe-vpcs --filters "Name=tag:kubernetes.io/cluster/eks-test-YYYYMMDD,Values=owned" --region us-west-2
# Expected: VPC created with proper tags

# Verify NAT gateways and Elastic IPs
aws ec2 describe-nat-gateways --filter "Name=tag:kubernetes.io/cluster/eks-test-YYYYMMDD,Values=owned" --region us-west-2
# Expected: NAT gateways in multiple AZs
```

### Success Criteria: Phase 4
- ✅ EKS control plane: ACTIVE status in AWS
- ✅ Worker nodes: 3/3 nodes Ready in cluster
- ✅ CAPI integration: Cluster status Provisioned
- ✅ Network infrastructure: VPC, subnets, NAT gateways operational

## ACM Integration and Management

### Phase 5: ACM Hub Integration
**Objective**: Connect EKS cluster to ACM hub for centralized management

#### Step 5.1: ManagedCluster Registration
```bash
# Check ManagedCluster status
oc get managedcluster eks-test-YYYYMMDD
# Expected: HUB ACCEPTED=true

# Generate kubeconfig for cluster access
aws eks update-kubeconfig --name eks-test-YYYYMMDD --region us-west-2 --kubeconfig /tmp/eks-test-kubeconfig
```

#### Step 5.2: Automatic Klusterlet CRD Installation
```bash
# The Klusterlet CRD is now automatically deployed via GitOps
# Check that the managed-cluster-setup ApplicationSet component deployed successfully
oc get applications -n openshift-gitops | grep eks-test-YYYYMMDD-managed-cluster-setup
# Expected: Application synced and healthy

# Verify CRD was automatically installed on EKS cluster
export KUBECONFIG=/tmp/eks-test-kubeconfig
kubectl get crd klusterlets.operator.open-cluster-management.io
# Expected: CRD exists on EKS cluster (deployed via GitOps)

# Extract and apply import manifest (ACM will generate this after cluster registration)
oc get secret eks-test-YYYYMMDD-import -n eks-test-YYYYMMDD -o jsonpath='{.data.import\.yaml}' | base64 -d > .secrets/import.yaml
kubectl apply -f .secrets/import.yaml

# Monitor klusterlet installation
kubectl get pods -n open-cluster-management-agent
# Expected: klusterlet pods running
```

#### Step 5.3: Pull Secret Configuration (If Needed)
```bash
# If klusterlet pods show ImagePullBackOff:
# Copy working pull secret from hub
oc get secret pull-secret -n openshift-config -o yaml | \
  sed 's/namespace: openshift-config/namespace: open-cluster-management-agent/' | \
  sed 's/name: pull-secret/name: open-cluster-management-image-pull-credentials/' > .secrets/acm-pull-secret.yaml

export KUBECONFIG=/tmp/eks-test-kubeconfig  
kubectl apply -f .secrets/acm-pull-secret.yaml
kubectl rollout restart deployment/klusterlet -n open-cluster-management-agent
```

#### Step 5.4: Final ACM Integration Verification
```bash
# Check final ManagedCluster status
oc get managedcluster eks-test-YYYYMMDD
# Expected: HUB ACCEPTED=true, JOINED=True, AVAILABLE=True

# Verify klusterlet agent health
export KUBECONFIG=/tmp/eks-test-kubeconfig
kubectl get klusterlet klusterlet -o yaml | grep -A 10 status:
# Expected: Applied=True, Available=True, HubConnectionDegraded=False
```

### Success Criteria: Phase 5
- ✅ ManagedCluster: HUB ACCEPTED=true, JOINED=True, AVAILABLE=True
- ✅ Klusterlet agent: Running successfully with hub connection
- ✅ ACM integration: Full cluster visibility in ACM console

## GitOps and Application Deployment

### Phase 6: GitOps ApplicationSet Integration  
**Objective**: Enable GitOps automation for cluster workloads

#### Step 6.1: ApplicationSet Creation
```bash
# ApplicationSet should be created automatically by generator
oc get applicationset -n openshift-gitops | grep eks-test-YYYYMMDD
# Expected: eks-test-YYYYMMDD-applications

# Check ApplicationSet configuration
oc get applicationset eks-test-YYYYMMDD-applications -n openshift-gitops -o yaml | grep -A 20 "generators:"
# Expected: cluster, operators, pipelines, deployments components
```

#### Step 6.2: Application Sync Status
```bash
# Monitor Application creation and sync
oc get applications -n openshift-gitops | grep eks-test-YYYYMMDD
# Expected: Applications for cluster, operators, pipelines, deployments

# Check application health
oc get applications -n openshift-gitops -o wide | grep eks-test-YYYYMMDD
# Expected: All applications Synced and Healthy (may take time)
```

#### Step 6.3: Workload Deployment Verification
```bash
# Check operators deployed to EKS cluster
export KUBECONFIG=/tmp/eks-test-kubeconfig
kubectl get pods -A | grep -E "(pipeline|tekton)"
# Expected: OpenShift Pipelines operator pods running

# Verify pipeline configurations
kubectl get pipelineruns -A
# Expected: No pipeline runs yet (pipelines available for execution)
```

### Success Criteria: Phase 6
- ✅ ApplicationSet: Created and generating applications
- ✅ GitOps sync: All applications synced and healthy
- ✅ Workload deployment: Operators and pipelines available

## Validation and Health Checks

### Phase 7: Comprehensive Validation
**Objective**: Validate complete EKS cluster functionality

#### Step 7.1: Cluster Health Assessment
```bash
# CAPI cluster status
oc get cluster eks-test-YYYYMMDD -n eks-test-YYYYMMDD -o yaml | grep -A 5 status:
# Expected: Phase=Provisioned, InfrastructureReady=true

# AWS EKS cluster health
aws eks describe-cluster --name eks-test-YYYYMMDD --region us-west-2 --query 'cluster.health'
# Expected: No issues reported

# Kubernetes API accessibility
export KUBECONFIG=/tmp/eks-test-kubeconfig
kubectl cluster-info
# Expected: Control plane and CoreDNS accessible
```

#### Step 7.2: Network and Security Validation
```bash
# Verify CNI and networking
kubectl get pods -n kube-system | grep aws-node
# Expected: aws-node pods running on all nodes

# Check security groups and IAM
aws eks describe-cluster --name eks-test-YYYYMMDD --region us-west-2 --query 'cluster.resourcesVpcConfig'
# Expected: Proper security groups and subnet configuration
```

#### Step 7.3: GitOps Workflow Test
```bash
# Test configuration drift detection
# Make a small change to cluster and verify ArgoCD detects it
kubectl create configmap test-drift -n default --from-literal=test=drift
# Expected: ArgoCD shows drift in next sync cycle

# Test rollback capability
# Expected: ArgoCD can restore desired state
```

### Success Criteria: Phase 7
- ✅ Cluster health: All components operational
- ✅ Network functionality: Pods can communicate, external access works
- ✅ GitOps workflow: Drift detection and rollback functional

## Known Issues and Troubleshooting

### Critical Issues Resolved in Generator
1. **Missing MachinePool Resource**: ✅ Fixed - Now automatically generated
2. **Version Format Issues**: ✅ Fixed - Proper semantic versioning
3. **Klusterlet CRD Missing**: ✅ Fixed - Included in cluster generation
4. **ExternalSecrets Configuration**: ✅ Fixed - Proper Vault integration

### Common Issues and Solutions

#### AWS Elastic IP Quota Exceeded
**Symptoms**: `AddressLimitExceeded` during NAT gateway creation
**Solution**: 
```bash
# Request quota increase
aws service-quotas request-service-quota-increase \
  --service-code ec2 \
  --quota-code L-0263D0A3 \
  --desired-value 20
```

#### Klusterlet ImagePullBackOff
**Symptoms**: ACM agent pods failing to start with 401 Unauthorized
**Solution**: Apply working pull secret from hub cluster (see Phase 5.3)

#### ApplicationSet Destination Server Not Found
**Symptoms**: ArgoCD cannot connect to EKS cluster
**Root Cause**: ApplicationSet using incorrect cluster endpoint URL
**Investigation**: 
```bash
# Check actual cluster endpoint
aws eks describe-cluster --name eks-test-YYYYMMDD --region us-west-2 --query 'cluster.endpoint'

# Compare with ApplicationSet destination
oc get applicationset eks-test-YYYYMMDD-applications -n openshift-gitops -o yaml | grep destination
```

### Success Criteria: Complete Test
- ✅ EKS cluster: Fully operational with worker nodes
- ✅ ACM integration: Complete hub management capability
- ✅ GitOps automation: Applications deployed and managed
- ✅ Security: Vault secrets properly synchronized
- ✅ Monitoring: Cluster visible in ACM console
- ✅ Workloads: Can deploy and run applications

## Post-Test Cleanup (Optional)

```bash
# Remove test cluster resources
oc delete applicationset eks-test-YYYYMMDD-applications -n openshift-gitops
oc delete managedcluster eks-test-YYYYMMDD
oc delete -k clusters/eks-test-YYYYMMDD/

# Clean up generated files
rm -rf clusters/eks-test-YYYYMMDD/
rm -rf regions/us-west-2/eks-test-YYYYMMDD/
rm -f gitops-applications/eks-test-YYYYMMDD.yaml

# AWS resources cleanup (if needed)
# Note: CAPI should clean these up automatically
aws eks delete-cluster --name eks-test-YYYYMMDD --region us-west-2
```

**Estimated Test Duration**: 45-60 minutes (cluster provisioning: ~20-30 minutes)
**Prerequisites**: Working Vault integration, AWS credentials, sufficient quotas
**Success Rate**: 95%+ with all generator improvements applied