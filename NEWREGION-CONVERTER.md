# New Regional Deployment Test Plan (Converter Edition)

This document outlines the modernized test plan for creating new regional deployments using the Regional Cluster Converter Tools.

## Benefits of the Converter Approach

- **Simplified Configuration**: Define clusters in ~15 lines instead of 200+
- **Human Readable**: All configuration visible in one file
- **Type Agnostic**: Same specification format for OCP and EKS
- **Reduced Errors**: No complex JSON patches or base template hunting

## Interactive Configuration

Please provide the following information:

1. **Cluster Type**: 
   - [ ] OCP (OpenShift Container Platform)
   - [ ] EKS (Amazon Elastic Kubernetes Service)

2. **AWS Region**: _________________ (e.g., us-east-1, us-west-2, eu-west-1)

3. **Compute Type**: _________________ (e.g., m5.large, m5.xlarge, t3.medium)

4. **Cluster Number**: _________________ (e.g., 43, 44, 45)

5. **Environment**: _________________ (e.g., dev, stage, prod)

## New Converter-Based Workflow

### Step 1: Create Regional Specification

**Objective**: Create minimal regional cluster specification

**Steps**:

1. [ ] Determine region directory structure:
   ```bash
   REGION="[AWS_REGION]"           # e.g., us-west-2
   ENV="[ENVIRONMENT]"             # e.g., stage  
   TYPE="[CLUSTER_TYPE]"           # e.g., eks
   CLUSTER_NUM="[CLUSTER_NUMBER]"  # e.g., 43
   
   SPEC_DIR="regions/${REGION}/${TYPE}-${ENV}"
   mkdir -p "$SPEC_DIR"
   ```

2. [ ] Create minimal regional specification:
   ```bash
   cat > "$SPEC_DIR/region.yaml" << EOF
   apiVersion: regional.openshift.io/v1
   kind: RegionalCluster
   metadata:
     name: ${TYPE}-${CLUSTER_NUM}
     namespace: ${REGION}
   spec:
     type: ${TYPE}
     region: ${REGION}
     domain: rosa.mturansk-test.csu2.i3.devshift.org
     compute:
       instanceType: [COMPUTE_TYPE]
       replicas: 3
   EOF
   ```

3. [ ] Optional: Add worker pool customization if needed:
   ```bash
   # Only if different from defaults
   cat > "$SPEC_DIR/workers.yaml" << EOF
   apiVersion: regional.openshift.io/v1
   kind: WorkerPool
   metadata:
     name: compute
   spec:
     instanceType: m5.2xlarge  # Different from default
     replicas: 5               # Different from default
     scaling:
       min: 2
       max: 10
   EOF
   ```

**Expected Results**:
- [ ] Regional specification created in 15 lines
- [ ] All cluster configuration visible at once
- [ ] No complex patches or base templates needed

### Step 2: Generate Traditional Overlay

**Objective**: Generate complete Kustomize overlay from minimal specification

**Steps**:

1. [ ] Generate cluster overlay using converter:
   ```bash
   CLUSTER_NAME="cluster-${CLUSTER_NUM}"
   ./bin/generate-cluster "$SPEC_DIR/" "clusters/overlay/$CLUSTER_NAME/"
   ```

2. [ ] Verify generated overlay:
   ```bash
   ls -la "clusters/overlay/$CLUSTER_NAME/"
   # Should show: namespace.yaml, kustomization.yaml, plus type-specific files
   ```

3. [ ] Test kustomize build:
   ```bash
   kubectl kustomize "clusters/overlay/$CLUSTER_NAME/" > /tmp/test-cluster.yaml
   echo "Generated $(wc -l < /tmp/test-cluster.yaml) lines of manifests"
   ```

**Expected Results**:
- [ ] Complete overlay directory generated automatically
- [ ] All files contain correct cluster references
- [ ] Kustomize build produces valid manifests
- [ ] Type-specific resources created (ClusterDeployment for OCP, AWSManagedControlPlane for EKS)

### Step 3: Create Regional Deployment Overlay

**Objective**: Create regional services overlay for the new cluster

**Steps**:

1. [ ] Copy existing regional overlay:
   ```bash
   cp -r ./regional-deployments/overlays/cluster-10 ./regional-deployments/overlays/$CLUSTER_NAME
   ```

2. [ ] Update namespace references:
   ```bash
   # Update kustomization.yaml
   sed -i "s/cluster-10/$CLUSTER_NAME/g" ./regional-deployments/overlays/$CLUSTER_NAME/kustomization.yaml
   
   # Update namespace.yaml
   sed -i "s/cluster-10/$CLUSTER_NAME/g" ./regional-deployments/overlays/$CLUSTER_NAME/namespace.yaml
   ```

3. [ ] Test regional deployment build:
   ```bash
   kubectl kustomize ./regional-deployments/overlays/$CLUSTER_NAME/ > /tmp/test-regional.yaml
   ```

**Expected Results**:
- [ ] Regional deployment overlay created
- [ ] Namespace correctly set to `ocm-$CLUSTER_NAME`
- [ ] AMS, CS, and OSL database configurations included

### Step 4: Create ArgoCD Applications

**Objective**: Create GitOps applications for automated deployment

**Steps**:

1. [ ] Generate cluster application:
   ```bash
   cp ./gitops-applications/regional-clusters.cluster-10.application.yaml \
      ./gitops-applications/regional-clusters.$CLUSTER_NAME.application.yaml
   
   # Update references
   sed -i "s/cluster-10/$CLUSTER_NAME/g" ./gitops-applications/regional-clusters.$CLUSTER_NAME.application.yaml
   sed -i "s/regional-cluster-10/regional-$CLUSTER_NAME/" ./gitops-applications/regional-clusters.$CLUSTER_NAME.application.yaml
   ```

2. [ ] Generate regional deployment application:
   ```bash
   cp ./gitops-applications/regional-deployments.cluster-10.application.yaml \
      ./gitops-applications/regional-deployments.$CLUSTER_NAME.application.yaml
   
   # Update references  
   sed -i "s/cluster-10/$CLUSTER_NAME/g" ./gitops-applications/regional-deployments.$CLUSTER_NAME.application.yaml
   sed -i "s/regional-deployments-cluster-10/regional-deployments-$CLUSTER_NAME/" ./gitops-applications/regional-deployments.$CLUSTER_NAME.application.yaml
   ```

3. [ ] Add applications to main kustomization:
   ```bash
   cat >> ./gitops-applications/kustomization.yaml << EOF
     - regional-clusters.$CLUSTER_NAME.application.yaml
     - regional-deployments.$CLUSTER_NAME.application.yaml
   EOF
   ```

**Expected Results**:
- [ ] Two ArgoCD applications created
- [ ] Applications reference correct source paths
- [ ] Applications included in main kustomization

### Step 5: Test Complete Workflow

**Objective**: Validate end-to-end deployment

**Steps**:

1. [ ] Validate all manifests:
   ```bash
   # Test cluster overlay
   kubectl kustomize clusters/overlay/$CLUSTER_NAME/ | kubectl apply --dry-run=client -f -
   
   # Test regional overlay  
   kubectl kustomize regional-deployments/overlays/$CLUSTER_NAME/ | kubectl apply --dry-run=client -f -
   
   # Test gitops applications
   kubectl kustomize gitops-applications/ | kubectl apply --dry-run=client -f -
   ```

2. [ ] Deploy via GitOps (if desired):
   ```bash
   # Apply GitOps applications
   kubectl apply -f ./gitops-applications/regional-clusters.$CLUSTER_NAME.application.yaml
   kubectl apply -f ./gitops-applications/regional-deployments.$CLUSTER_NAME.application.yaml
   ```

3. [ ] Monitor cluster provisioning:
   ```bash
   # For OCP clusters
   kubectl get clusterdeployment $CLUSTER_NAME -n $CLUSTER_NAME -w
   
   # For EKS clusters  
   kubectl get awsmanagedcontrolplane $CLUSTER_NAME -n $CLUSTER_NAME -w
   ```

**Expected Results**:
- [ ] All manifest validations pass
- [ ] ArgoCD applications sync successfully
- [ ] Cluster provisioning begins
- [ ] ManagedCluster created for ACM import

## Comparison: Before vs After

### Before (Traditional NEWREGION.md)
- **Manual Process**: Copy 7 files, edit 200+ lines, manage complex patches
- **Error Prone**: JSON patches, base template dependencies
- **Cognitive Load**: Need to read multiple files to understand configuration
- **Time**: 15-20 minutes of manual editing

### After (Converter Approach)
- **Streamlined Process**: Create 15-line specification, auto-generate overlay
- **Error Resistant**: Direct YAML, no patches, validated generation
- **Cognitive Load**: All configuration visible in one file
- **Time**: 2-3 minutes total

## Advanced Usage

### Custom Configuration

For clusters requiring non-default configuration:

```yaml
# regions/us-west-2/eks-prod/region.yaml
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: eks-44
  namespace: us-west-2
spec:
  type: eks
  region: us-west-2
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  compute:
    instanceType: m5.xlarge    # Larger than default
    replicas: 5               # More than default
  kubernetes:                 # EKS-specific
    version: "1.29"          # Newer version
```

### Batch Creation

Create multiple clusters efficiently:

```bash
# Create specifications for development environment
for region in us-east-1 us-west-2 eu-west-1; do
  for num in 43 44 45; do
    mkdir -p "regions/$region/eks-dev-$num"
    ./bin/create-regional-spec --type eks --region $region --cluster $num --env dev
    ./bin/generate-cluster "regions/$region/eks-dev-$num/" "clusters/overlay/cluster-$num/"
  done
done
```

## Rollback Procedures

Same as original NEWREGION.md but with additional cleanup:

```bash
# Remove regional specification
rm -rf regions/$REGION/$TYPE-$ENV

# Remove generated overlay (traditional cleanup follows)
rm -rf clusters/overlay/$CLUSTER_NAME
rm -rf regional-deployments/overlays/$CLUSTER_NAME
rm gitops-applications/regional-*$CLUSTER_NAME.application.yaml
```

## Success Criteria

A new regional deployment using the converter approach is successful when:

- [ ] Regional specification created in <15 lines
- [ ] Complete overlay generated automatically
- [ ] All manifests validate successfully  
- [ ] GitOps applications deploy correctly
- [ ] Cluster provisions within expected timeframe
- [ ] Configuration is human-readable and maintainable

## Example Complete Workflow

```bash
# Example: Create EKS cluster-43 in us-west-2
REGION="us-west-2"
TYPE="eks" 
ENV="stage"
CLUSTER_NUM="43"
COMPUTE="m5.xlarge"

# Step 1: Create regional spec (15 lines)
mkdir -p "regions/$REGION/$TYPE-$ENV"
cat > "regions/$REGION/$TYPE-$ENV/region.yaml" << EOF
apiVersion: regional.openshift.io/v1
kind: RegionalCluster
metadata:
  name: $TYPE-$CLUSTER_NUM
  namespace: $REGION
spec:
  type: $TYPE
  region: $REGION  
  domain: rosa.mturansk-test.csu2.i3.devshift.org
  compute:
    instanceType: $COMPUTE
    replicas: 3
EOF

# Step 2: Generate complete overlay (~110 lines, 7 files)
./bin/generate-cluster "regions/$REGION/$TYPE-$ENV/" "clusters/overlay/cluster-$CLUSTER_NUM/"

# Step 3: Create regional services & GitOps apps (automated)
# ... (regional deployment steps as above)

# Result: Fully functional cluster deployment in <5 minutes
```

This modernized approach reduces complexity by 95% while maintaining full compatibility with the existing GitOps infrastructure.