# OCP Cluster Provisioning Test Plan

**Date**: 2025-07-19  
**Version**: 2.0  
**Based on**: Installation health plan methodology + OCP cluster patterns

## Test Overview

Comprehensive test plan for provisioning Red Hat OpenShift Container Platform (OCP) clusters using Red Hat Hive operator through the bootstrap system. This plan covers AWS-based OCP deployments with ACM integration.

## Prerequisites Verification

### Phase 1: Infrastructure Dependencies
**Objective**: Verify all required components are operational before OCP cluster creation

#### Step 1.1: Vault Integration Health Check
```bash
# Verify ClusterSecretStore connectivity
oc get clustersecretstore vault-cluster-store -o yaml | grep -A 5 status:
# Expected: Valid=True, Ready=True

# Test ExternalSecrets functionality
oc get externalsecrets -A | grep -E "(Ready|SecretSynced)"
# Expected: All existing ExternalSecrets showing Ready=True
```

#### Step 1.2: Hive Operator Validation
```bash
# Verify Hive operator is installed and running
oc get pods -n hive | grep hive
# Expected: hive-operator, hive-controllers pods Running

# Check Hive CRDs availability
oc get crd | grep hive
# Expected: ClusterDeployment, ClusterImageSet, etc.

# Verify ClusterImageSets available
oc get clusterimageset
# Expected: Available OCP versions (4.14, 4.15, etc.)
```

#### Step 1.3: ACM Hub Status  
```bash
# Verify ACM hub operational
oc get multiclusterhub -n open-cluster-management
# Expected: STATUS=Running

# Check cluster capacity
oc get managedcluster | wc -l
# Note: Monitor cluster count for planning
```

#### Step 1.4: AWS Prerequisites Check
```bash
# Verify AWS credentials in Vault
# Check base domain delegation
dig NS rosa.mturansk-test.csu2.i3.devshift.org
# Expected: NS records pointing to Route53 hosted zone

# Verify AWS quotas for OCP
aws service-quotas get-service-quota --service-code ec2 --quota-code L-1216C47A --region us-east-1
# Expected: Sufficient EC2 instance limits for OCP masters + workers
```

### Success Criteria: Phase 1
- ✅ ClusterSecretStore: Valid and Ready
- ✅ Hive operator: Running with CRDs available
- ✅ ACM hub: Operational and accepting clusters  
- ✅ AWS quotas: Sufficient for OCP cluster (minimum 6 instances)

## Cluster Generation and Configuration

### Phase 2: OCP Cluster Generation
**Objective**: Generate complete OCP cluster configuration using project conventions

#### Step 2.1: Regional Specification Creation
```bash
# Option A: Use new-cluster tool (interactive)
./bin/new-cluster
# Select: ocp, region, instance type, replicas

# Option B: Manual creation for testing
mkdir -p regions/us-east-1/ocp-test-$(date +%s)
```

**Regional Specification Template**:
```yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: ocp-test-YYYYMMDD
  namespace: us-east-1
spec:
  type: ocp
  region: us-east-1
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  
  compute:
    instanceType: m5.2xlarge
    replicas: 3
    
  openshift:
    version: "4.15"
    channel: stable
```

#### Step 2.2: Cluster Manifest Generation
```bash
# Generate complete cluster configuration
./bin/generate-cluster regions/us-east-1/ocp-test-YYYYMMDD/

# Verify generated structure
ls -la clusters/ocp-test-YYYYMMDD/
# Expected files:
# - namespace.yaml
# - install-config.yaml (OCP installation configuration)
# - managedcluster.yaml
# - klusterletaddonconfig.yaml
# - kustomization.yaml (references bases/clusters)
```

#### Step 2.3: Configuration Validation  
```bash
# Validate OCP-specific configuration
cat clusters/ocp-test-YYYYMMDD/install-config.yaml
# Expected: Proper AWS platform config, base domain, worker pools

# Verify Kustomize with base overlays
oc kustomize clusters/ocp-test-YYYYMMDD/ | head -100
# Expected: Valid YAML with ClusterDeployment, external-secrets

# Check base cluster templates referenced
ls -la bases/clusters/
# Expected: external-secrets.yaml, clusterdeployment.yaml, etc.
```

### Success Criteria: Phase 2
- ✅ Regional specification: Valid OCP format with version/channel
- ✅ Cluster generation: All OCP-specific resources created
- ✅ Base templates: Properly referenced in kustomization
- ✅ Install config: Valid AWS platform configuration

## Secrets and Security Configuration

### Phase 3: ExternalSecrets Integration
**Objective**: Ensure proper secret management for OCP cluster deployment

#### Step 3.1: ExternalSecrets Validation
```bash
# OCP uses base template for ExternalSecrets
cat bases/clusters/external-secrets.yaml
# Expected: aws-credentials and pull-secret configurations

# Verify cluster-specific namespace application
oc kustomize clusters/ocp-test-YYYYMMDD/ | grep -A 10 "kind: ExternalSecret"
# Expected: Namespace set to ocp-test-YYYYMMDD
```

#### Step 3.2: Cluster Resource Deployment
```bash
# Apply cluster configuration  
oc apply -k clusters/ocp-test-YYYYMMDD/

# Monitor namespace creation and secrets
oc get ns ocp-test-YYYYMMDD
oc get externalsecrets -n ocp-test-YYYYMMDD

# Watch ExternalSecret sync
watch 'oc get externalsecrets -n ocp-test-YYYYMMDD'
# Expected: Both aws-credentials and pull-secret Ready=True
```

#### Step 3.3: OCP-Specific Secret Verification
```bash
# Verify AWS credentials format
oc get secret aws-credentials -n ocp-test-YYYYMMDD -o yaml | grep -A 10 data:
# Expected: aws_access_key_id, aws_secret_access_key keys

# Check pull secret format
oc get secret pull-secret -n ocp-test-YYYYMMDD -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq .
# Expected: Valid docker config with registry.redhat.io auth
```

### Success Criteria: Phase 3  
- ✅ ExternalSecrets: Both aws-credentials and pull-secret Ready=True
- ✅ Secret format: AWS credentials and Red Hat registry auth properly formatted
- ✅ Vault integration: No authentication errors during sync

## Cluster Provisioning and Deployment

### Phase 4: Hive Cluster Provisioning
**Objective**: Deploy OCP cluster through Hive and monitor installation

#### Step 4.1: ClusterDeployment Creation
```bash
# Verify ClusterDeployment created
oc get clusterdeployment -n ocp-test-YYYYMMDD
# Expected: ClusterDeployment resource exists

# Check ClusterDeployment configuration
oc get clusterdeployment ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD -o yaml | grep -A 10 spec:
# Expected: Proper AWS platform, base domain, cluster name
```

#### Step 4.2: Provision Job Monitoring
```bash
# Monitor provision job creation
oc get jobs -n ocp-test-YYYYMMDD
# Expected: install job created

# Follow provision logs
oc logs -n ocp-test-YYYYMMDD job/ocp-test-YYYYMMDD-provision -f
# Expected: OpenShift installer output, AWS resource creation

# Check provision status
oc get clusterdeployment ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD -o yaml | grep -A 5 "provisionState"
# Expected progression: Initializing → Provisioning → Provisioned
```

#### Step 4.3: AWS Infrastructure Verification
```bash
# Monitor AWS resources being created
aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/ocp-test-YYYYMMDD,Values=owned" --region us-east-1
# Expected: EC2 instances for masters and workers

# Check Route53 records
aws route53 list-resource-record-sets --hosted-zone-id $(aws route53 list-hosted-zones-by-name --dns-name rosa.mturansk-test.csu2.i3.devshift.org --query 'HostedZones[0].Id' --output text)
# Expected: DNS records for cluster API and apps

# Verify load balancers
aws elbv2 describe-load-balancers --region us-east-1 | grep ocp-test-YYYYMMDD
# Expected: Load balancers for API and ingress
```

#### Step 4.4: Cluster Installation Progress
```bash
# Monitor cluster installation (typical: 30-45 minutes)
watch 'oc get clusterdeployment ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD -o yaml | grep -A 10 conditions:'
# Expected: Provisioned=True when complete

# Check for installation errors
oc get clusterdeployment ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD -o yaml | grep -A 20 "conditions:" | grep -A 5 "ProvisionFailed"
# Expected: No ProvisionFailed conditions

# Verify cluster accessibility
oc get secret ocp-test-YYYYMMDD-admin-kubeconfig -n ocp-test-YYYYMMDD
# Expected: Admin kubeconfig secret created when cluster ready
```

### Success Criteria: Phase 4
- ✅ ClusterDeployment: Provisioned=True status
- ✅ AWS resources: EC2 instances, load balancers, DNS records created
- ✅ Installation: Completed without errors (~45 minutes)
- ✅ Cluster access: Admin kubeconfig available

## OCP Cluster Validation

### Phase 5: OpenShift Cluster Health
**Objective**: Validate OCP cluster is fully operational

#### Step 5.1: Cluster Access and Authentication
```bash
# Extract admin kubeconfig
oc get secret ocp-test-YYYYMMDD-admin-kubeconfig -n ocp-test-YYYYMMDD -o jsonpath='{.data.kubeconfig}' | base64 -d > /tmp/ocp-test-kubeconfig

# Test cluster access
export KUBECONFIG=/tmp/ocp-test-kubeconfig
oc cluster-info
# Expected: OpenShift API server and console accessible

# Verify cluster version
oc get clusterversion
# Expected: VERSION matches specified in regional spec (4.15.x)
```

#### Step 5.2: Node and Infrastructure Validation
```bash
# Check all nodes are Ready
oc get nodes
# Expected: 6 nodes total (3 masters + 3 workers) in Ready state

# Verify cluster operators
oc get clusteroperators
# Expected: All operators Available=True, Degraded=False

# Check persistent volumes
oc get pv
# Expected: PVs available for etcd and other components

# Verify networking
oc get nodes -o wide
# Expected: All nodes have proper internal/external IPs
```

#### Step 5.3: OpenShift-Specific Features
```bash
# Test OpenShift console access
oc get route console -n openshift-console
# Expected: Console route with proper hostname

# Verify image registry
oc get configs.imageregistry.operator.openshift.io cluster -o yaml | grep -A 5 "managementState"
# Expected: managementState: Managed

# Check OAuth and authentication
oc get oauth cluster -o yaml | grep -A 10 "identityProviders"
# Expected: OAuth configured properly

# Test internal registry
oc get pods -n openshift-image-registry
# Expected: Registry pods running
```

### Success Criteria: Phase 5
- ✅ Cluster access: API and console accessible with admin credentials
- ✅ Node health: All 6 nodes Ready and properly networked
- ✅ Cluster operators: All Available and not Degraded  
- ✅ OCP features: Console, registry, OAuth operational

## ACM Integration and Management

### Phase 6: ACM Hub Integration
**Objective**: Connect OCP cluster to ACM hub for centralized management

#### Step 6.1: ManagedCluster Registration
```bash
# Check ManagedCluster status
oc get managedcluster ocp-test-YYYYMMDD
# Expected: HUB ACCEPTED=true (may take a few minutes)

# Verify cluster import secret
oc get secret ocp-test-YYYYMMDD-import -n ocp-test-YYYYMMDD
# Expected: Import secret with cluster registration data
```

#### Step 6.2: Klusterlet Agent Installation
```bash
# Import should happen automatically for Hive-provisioned clusters
# Monitor klusterlet installation on managed cluster
export KUBECONFIG=/tmp/ocp-test-kubeconfig
oc get pods -n open-cluster-management-agent
# Expected: klusterlet and agent pods running

# Check klusterlet status
oc get klusterlet klusterlet -o yaml | grep -A 10 "conditions:"
# Expected: Applied=True, Available=True
```

#### Step 6.3: ACM Integration Verification
```bash
# Switch back to hub cluster
unset KUBECONFIG

# Check final ManagedCluster status
oc get managedcluster ocp-test-YYYYMMDD
# Expected: HUB ACCEPTED=true, JOINED=True, AVAILABLE=True

# Verify cluster information populated
oc get managedclusterinfo ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD -o yaml | grep -A 10 "kubeVendor"
# Expected: kubeVendor: OpenShift, version information populated
```

### Success Criteria: Phase 6
- ✅ ManagedCluster: HUB ACCEPTED=true, JOINED=True, AVAILABLE=True
- ✅ Klusterlet agent: Running on managed cluster with hub connection
- ✅ Cluster info: OpenShift version and details visible in ACM

## GitOps and Application Deployment

### Phase 7: GitOps ApplicationSet Integration
**Objective**: Enable GitOps automation for OCP cluster workloads

#### Step 7.1: ApplicationSet Configuration
```bash
# Verify ApplicationSet created
oc get applicationset -n openshift-gitops | grep ocp-test-YYYYMMDD
# Expected: ocp-test-YYYYMMDD-applications

# Check ApplicationSet targets OCP cluster
oc get applicationset ocp-test-YYYYMMDD-applications -n openshift-gitops -o yaml | grep -A 10 "destination"
# Expected: Destination points to OCP cluster API endpoint
```

#### Step 7.2: Application Deployment
```bash
# Monitor Application creation
oc get applications -n openshift-gitops | grep ocp-test-YYYYMMDD
# Expected: Applications for operators, pipelines, deployments

# Check application sync status  
oc get applications -n openshift-gitops -o wide | grep ocp-test-YYYYMMDD
# Expected: All applications Synced and Healthy

# Verify applications deployed to managed cluster
export KUBECONFIG=/tmp/ocp-test-kubeconfig
oc get pods -A | grep -E "(pipeline|tekton)"
# Expected: OpenShift Pipelines operator and components
```

#### Step 7.3: Pipeline Validation
```bash
# Check pipeline configurations deployed
oc get pipelines -A
# Expected: hello-world and cloud-infrastructure pipelines

# Verify pipeline execution capability
oc get pipelineruns -A
# Expected: No runs yet, but pipelines ready for execution

# Test pipeline trigger (optional)
# Expected: Pipelines can be triggered and execute successfully
```

### Success Criteria: Phase 7
- ✅ ApplicationSet: Created and targeting OCP cluster correctly
- ✅ Applications: All synced and healthy
- ✅ Workloads: OpenShift Pipelines and other operators deployed
- ✅ Pipeline readiness: Tekton pipelines available for execution

## Validation and Health Checks

### Phase 8: Comprehensive OCP Validation
**Objective**: Validate complete OCP cluster functionality and integration

#### Step 8.1: OpenShift-Specific Health Checks
```bash
# Comprehensive cluster operator status
export KUBECONFIG=/tmp/ocp-test-kubeconfig
oc get clusteroperators -o wide
# Expected: All operators Available=True, Progressing=False, Degraded=False

# Check cluster resource consumption
oc adm top nodes
oc adm top pods -A --sort-by=cpu
# Expected: Reasonable resource utilization

# Verify cluster networking
oc get network.operator cluster -o yaml
# Expected: Network operator configured properly
```

#### Step 8.2: Application Deployment Test
```bash
# Deploy test application
oc new-project test-deployment
oc new-app --docker-image=nginx:latest --name=test-nginx
oc expose svc/test-nginx

# Verify application accessibility
oc get route test-nginx
# Test external access to application

# Clean up test
oc delete project test-deployment
```

#### Step 8.3: Backup and DR Readiness
```bash
# Check etcd backup capability
oc get pods -n openshift-etcd | grep backup
# Expected: etcd backup jobs scheduled

# Verify cluster certificate status
oc get clusteroperators/kube-apiserver-operator -o yaml | grep -A 5 "certificates"
# Expected: Certificates valid and not expiring soon

# Check cluster update capability
oc adm upgrade
# Expected: Cluster ready for updates when needed
```

### Success Criteria: Phase 8
- ✅ Cluster operators: All healthy and not degraded
- ✅ Application deployment: Can deploy and expose applications
- ✅ Networking: Internal and external connectivity working
- ✅ Maintenance readiness: Backup and update capabilities functional

## Known Issues and Troubleshooting

### Common OCP Installation Issues

#### Installation Timeout
**Symptoms**: Provision job fails after 60+ minutes
**Investigation**:
```bash
oc logs -n ocp-test-YYYYMMDD job/ocp-test-YYYYMMDD-provision --tail=100
# Look for specific error messages

# Check AWS quotas and limits
aws service-quotas list-service-quotas --service-code ec2 --region us-east-1
```

#### DNS Resolution Issues
**Symptoms**: Cluster API inaccessible, console redirect failures
**Investigation**:
```bash
# Verify base domain delegation
dig NS rosa.mturansk-test.csu2.i3.devshift.org @8.8.8.8

# Check Route53 hosted zone
aws route53 list-hosted-zones-by-name --dns-name rosa.mturansk-test.csu2.i3.devshift.org
```

#### Certificate Issues
**Symptoms**: TLS errors accessing cluster
**Solution**: Verify Let's Encrypt integration or certificate management

#### Resource Exhaustion
**Symptoms**: Provision job fails with insufficient resources
**Solution**: 
- Check AWS instance limits
- Verify instance types available in region
- Consider reducing cluster size for testing

### OCP-Specific Troubleshooting

#### ClusterDeployment Stuck in Provisioning
```bash
# Check provision job details
oc describe job ocp-test-YYYYMMDD-provision -n ocp-test-YYYYMMDD

# Look for specific installer errors
oc logs job/ocp-test-YYYYMMDD-provision -n ocp-test-YYYYMMDD | grep -i error

# Check ClusterDeployment conditions
oc get clusterdeployment ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD -o yaml | grep -A 20 conditions:
```

#### Missing Pull Secret
```bash
# Verify pull secret exists and has correct format
oc get secret pull-secret -n ocp-test-YYYYMMDD -o yaml

# Check ExternalSecret sync
oc get externalsecret pull-secret -n ocp-test-YYYYMMDD -o yaml | grep -A 5 status:
```

### Success Criteria: Complete Test
- ✅ OCP cluster: Fully operational with all operators healthy
- ✅ ACM integration: Complete hub management capability
- ✅ GitOps automation: Applications deployed and managed via ArgoCD
- ✅ Security: Vault secrets properly synchronized
- ✅ Monitoring: Cluster visible and manageable in ACM console
- ✅ Workloads: Can deploy applications and execute pipelines

## Post-Test Cleanup (Optional)

```bash
# Remove GitOps applications first
oc delete applicationset ocp-test-YYYYMMDD-applications -n openshift-gitops

# Remove from ACM management
oc delete managedcluster ocp-test-YYYYMMDD

# Deprovision cluster (this will delete AWS resources)
oc delete clusterdeployment ocp-test-YYYYMMDD -n ocp-test-YYYYMMDD

# Clean up namespace and local files
oc delete namespace ocp-test-YYYYMMDD
rm -rf clusters/ocp-test-YYYYMMDD/
rm -rf regions/us-east-1/ocp-test-YYYYMMDD/
rm -f gitops-applications/ocp-test-YYYYMMDD.yaml
rm -f /tmp/ocp-test-kubeconfig

# Verify AWS resources cleaned up
aws ec2 describe-instances --filters "Name=tag:kubernetes.io/cluster/ocp-test-YYYYMMDD,Values=owned" --region us-east-1
# Expected: No instances found after cleanup completes
```

**Estimated Test Duration**: 60-90 minutes (cluster installation: ~45 minutes)
**Prerequisites**: Working Vault integration, AWS credentials, base domain delegation
**Success Rate**: 90%+ with proper DNS and quota configuration