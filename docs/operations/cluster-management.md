# Day-to-Day Cluster Management

**Audience**: Operators  
**Complexity**: Intermediate  
**Estimated Time**: Ongoing reference  
**Prerequisites**: Working hub cluster, deployed managed clusters

## Daily Operations

### Morning Health Check
```bash
# Complete environment overview
./bin/health-check

# Quick cluster status
oc get managedcluster

# ArgoCD application status  
oc get applications -n openshift-gitops

# Check for any failed sync
oc get applications -n openshift-gitops | grep -E "(OutOfSync|Unknown|Degraded)"
```

### Weekly Maintenance
```bash
# Update STATUS.md with current state
./bin/health-check

# Check cluster resource usage
oc adm top nodes --use-protocol-buffers

# Review ArgoCD sync history
oc get applications -n openshift-gitops -o custom-columns=NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status,AGE:.metadata.creationTimestamp
```

## Managing Existing Clusters

### Scaling Clusters

#### EKS Cluster Scaling:
```bash
# Update machine pool replicas
oc patch awsmanagedmachinepool cluster-name -n cluster-name --type='merge' -p='{"spec":{"scaling":{"maxSize":10,"minSize":3}}}'

# Check scaling status
oc get awsmanagedmachinepool cluster-name -n cluster-name -o yaml | grep -A 5 scaling
```

#### OCP Cluster Scaling:
```bash
# Update machine pool replicas  
oc patch machinepool cluster-name-worker -n cluster-name --type='merge' -p='{"spec":{"replicas":5}}'

# Check scaling status
oc get machinepool cluster-name-worker -n cluster-name
```

### Cluster Upgrades

#### EKS Cluster Upgrades:
```bash
# Update Kubernetes version
oc patch awsmanagedcontrolplane cluster-name -n cluster-name --type='merge' -p='{"spec":{"version":"1.28"}}'

# Monitor upgrade progress
oc describe awsmanagedcontrolplane cluster-name -n cluster-name
```

#### OCP Cluster Upgrades:
```bash
# Update OpenShift version in ClusterDeployment
oc patch clusterdeployment cluster-name -n cluster-name --type='merge' -p='{"spec":{"provisioning":{"imageSetRef":{"name":"openshift-v4.14.0-img"}}}}'

# Monitor upgrade progress
oc describe clusterdeployment cluster-name -n cluster-name
```

### Managing Cluster Configuration

#### Update Regional Specifications:
```bash
# Edit regional specification
vim regions/us-east-1/cluster-name/region.yaml

# Regenerate cluster configuration
./bin/generate-cluster regions/us-east-1/cluster-name/

# Apply changes via GitOps
git add . && git commit -m "Update cluster-name configuration" && git push
```

#### Add New Services to Existing Clusters:
```bash
# Add new deployment overlay
mkdir -p deployments/new-service/cluster-name/

# Create kustomization and resources
cat > deployments/new-service/cluster-name/kustomization.yaml << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: ocm-cluster-name
resources:
- ../../../bases/new-service/
EOF

# Update ApplicationSet to include new service
vim gitops-applications/cluster-name.yaml
```

## Application Management

### Troubleshooting Sync Issues

#### Application Out of Sync:
```bash
# Check application status
oc describe application cluster-name-cluster -n openshift-gitops

# Force sync
oc patch application cluster-name-cluster -n openshift-gitops --type='merge' -p='{"operation":{"sync":{"syncStrategy":{"hook":{"force":true}}}}}'

# Check sync result
oc get application cluster-name-cluster -n openshift-gitops -o yaml | grep -A 10 status
```

#### Resource Drift Detection:
```bash
# Compare desired vs actual state
oc get application cluster-name-deployments-ocm -n openshift-gitops -o yaml | grep -A 20 comparedTo

# View diff details
argocd app diff cluster-name-deployments-ocm --server openshift-gitops-server-openshift-gitops.apps.hub-cluster.com
```

### Managing ApplicationSets

#### Update ApplicationSet Template:
```bash
# Edit ApplicationSet for cluster
vim gitops-applications/cluster-name.yaml

# Validate changes
oc --dry-run=client apply -f gitops-applications/cluster-name.yaml

# Apply changes  
oc apply -f gitops-applications/cluster-name.yaml
```

#### Suspend/Resume ApplicationSet:
```bash
# Suspend ApplicationSet (stops creating new apps)
oc patch applicationset cluster-name-applications -n openshift-gitops --type='merge' -p='{"spec":{"syncPolicy":{"suspend":true}}}'

# Resume ApplicationSet
oc patch applicationset cluster-name-applications -n openshift-gitops --type='merge' -p='{"spec":{"syncPolicy":{"suspend":false}}}'
```

## Secret Management

### Vault Integration Operations

#### Rotate Secrets in Vault:
```bash
# Update AWS credentials in Vault
oc exec vault-0 -n vault -- vault kv put secret/aws-credentials \
  aws_access_key_id="new-access-key" \
  aws_secret_access_key="new-secret-key"

# Update pull secret in Vault
oc exec vault-0 -n vault -- vault kv put secret/pull-secret \
  .dockerconfigjson="$(cat new-pull-secret.json)"

# Secrets automatically refresh within 1 hour via ESO
```

#### Manual Secret Refresh:
```bash
# Force ExternalSecret refresh
oc annotate externalsecret aws-credentials -n cluster-name force-sync="$(date +%s)"

# Check refresh status
oc describe externalsecret aws-credentials -n cluster-name
```

### Traditional Secret Management:
```bash
# Update secret files
vim secrets/aws-credentials.yaml
vim secrets/pull-secret.yaml

# Apply to specific cluster namespace
oc apply -f secrets/aws-credentials.yaml -n cluster-name
oc apply -f secrets/pull-secret.yaml -n cluster-name
```

## Monitoring and Alerting

### Custom Health Checks
```bash
# Check specific cluster health
oc get managedcluster cluster-name -o yaml | grep -A 10 conditions

# Verify cluster connectivity
oc get klusterletaddonconfig cluster-name -n cluster-name -o yaml | grep -A 5 applicationManager

# Check resource consumption
oc adm top nodes --use-protocol-buffers --selector=kubernetes.io/hostname=cluster-name-node
```

### Setting Up Alerts
```bash
# Monitor application sync failures
oc get applications -n openshift-gitops -o json | jq '.items[] | select(.status.sync.status != "Synced") | .metadata.name'

# Monitor cluster provisioning time
oc get clusterdeployment -A -o custom-columns=NAME:.metadata.name,CREATED:.metadata.creationTimestamp,STATUS:.status.conditions[-1].type

# Monitor resource conflicts
oc get applications -n openshift-gitops -o json | jq '.items[] | select(.status.health.status == "Degraded") | {name: .metadata.name, message: .status.health.message}'
```

## Performance Optimization

### ArgoCD Performance
```bash
# Check ArgoCD resource usage
oc adm top pods -n openshift-gitops

# Optimize sync frequency
oc patch argocd openshift-gitops -n openshift-gitops --type='merge' -p='{"spec":{"controller":{"appResyncPeriod":"300s"}}}'

# Enable resource caching
oc patch argocd openshift-gitops -n openshift-gitops --type='merge' -p='{"spec":{"controller":{"enableGRPCWeb":true}}}'
```

### Cluster Resource Management
```bash
# Monitor cluster resource allocation
oc describe nodes | grep -E "(Allocated|Capacity)"

# Check pod resource requests vs limits
oc get pods -A -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace,CPU-REQ:.spec.containers[*].resources.requests.cpu,MEM-REQ:.spec.containers[*].resources.requests.memory
```

## Backup and Recovery

### Configuration Backup
```bash
# Backup regional specifications
tar -czf regions-backup-$(date +%Y%m%d).tar.gz regions/

# Backup generated overlays
tar -czf overlays-backup-$(date +%Y%m%d).tar.gz clusters/ operators/ pipelines/ deployments/ gitops-applications/

# Backup ArgoCD applications
oc get applications -n openshift-gitops -o yaml > argocd-apps-backup-$(date +%Y%m%d).yaml
```

### Disaster Recovery
```bash
# Restore from backup
tar -xzf overlays-backup-YYYYMMDD.tar.gz

# Restore ArgoCD applications
oc apply -f argocd-apps-backup-YYYYMMDD.yaml

# Regenerate cluster configurations
./bin/regenerate-all-clusters
```

## Compliance and Governance

### ACM Policy Management
```bash
# Check policy compliance across clusters
oc get policy -A

# View policy violations
oc describe policy cluster-policy -n open-cluster-management

# Update policy templates
vim operators/advanced-cluster-management/global/gitops-integration/policies/
```

### Audit and Reporting
```bash
# Generate cluster inventory report
oc get managedcluster -o custom-columns=NAME:.metadata.name,STATUS:.status.conditions[-1].type,VERSION:.status.version.kubernetes,CREATED:.metadata.creationTimestamp

# Application deployment report
oc get applications -n openshift-gitops -o custom-columns=NAME:.metadata.name,CLUSTER:.spec.destination.name,SYNC:.status.sync.status,HEALTH:.status.health.status

# Resource usage report
oc adm top nodes --use-protocol-buffers > cluster-resource-usage-$(date +%Y%m%d).txt
```

## Automation Scripts

### Bulk Operations
```bash
# Update all EKS clusters to new instance type
for cluster in $(oc get awsmanagedmachinepool -A -o custom-columns=NAMESPACE:.metadata.namespace --no-headers); do
  oc patch awsmanagedmachinepool $(oc get awsmanagedmachinepool -n $cluster -o name) -n $cluster --type='merge' -p='{"spec":{"instanceType":"m5.xlarge"}}'
done

# Force sync all applications
for app in $(oc get applications -n openshift-gitops -o name); do
  oc patch $app -n openshift-gitops --type='merge' -p='{"operation":{"sync":{"syncStrategy":{"hook":{"force":true}}}}}'
done
```

### Scheduled Maintenance
```bash
# Create maintenance script
cat > maintenance.sh << 'EOF'
#!/bin/bash
# Daily maintenance tasks
./bin/health-check
oc get applications -n openshift-gitops | grep -E "(OutOfSync|Unknown|Degraded)" | mail -s "ArgoCD Sync Issues" admin@company.com
EOF

# Schedule with cron
echo "0 8 * * * /path/to/maintenance.sh" | crontab -
```

## Related Documentation

- **[Monitoring Guide](monitoring.md)** - Status checking and troubleshooting
- **[Troubleshooting Guide](troubleshooting.md)** - Common issues and solutions  
- **[Architecture Overview](../architecture/gitops-flow.md)** - Technical architecture details
- **[Reference Commands](../reference/commands.md)** - Quick command reference