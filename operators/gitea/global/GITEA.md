# Gitea Internal Git Server - Installation and Usage Guide

## Overview

Gitea is deployed as an internal git server to solve the ArgoCD dependency issue where new cluster configurations must be available in a git repository before ArgoCD can sync them. This eliminates the manual step of pushing to GitHub before cluster provisioning.

## Problem Solved

**Before Gitea:** 
- Generate cluster configs locally
- Manually push to GitHub 
- ArgoCD fails if configs aren't in GitHub yet
- Manual intervention required for every new cluster

**After Gitea:**
- Generate cluster configs locally
- Automatically push to internal Gitea 
- ArgoCD syncs immediately from internal repository
- Zero manual intervention needed

## Architecture

### Components
- **PostgreSQL Database**: Persistent storage for Gitea data
- **Gitea Server**: Lightweight git server (200MB vs 4GB for GitLab)
- **ArgoCD Integration**: Repository secret for authenticated access
- **Enhanced bin/cluster-generate**: Automatic push capability

### Network Topology
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  bin/generate-  │───▶│  Internal Gitea  │───▶│     ArgoCD      │
│    cluster      │    │   Repository     │    │  Applications   │
│  --push-to-     │    │                  │    │                 │
│     gitea       │    │ gitea-system.svc │    │ openshift-gitops│
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Installation

### Prerequisites
- OpenShift cluster with admin access
- `oc` CLI tool configured
- Sufficient cluster resources (1 CPU, 1GB RAM minimum)

### Deployment Steps

1. **Deploy Complete Infrastructure**
   ```bash
   oc apply -k operators/gitea/global/
   ```

2. **Verify Deployment**
   ```bash
   # Check pods are running
   oc get pods -n gitea-system
   
   # Should show:
   # gitea-[hash]     1/1   Running
   # gitea-db-[hash]  1/1   Running
   ```

3. **Verify Database Connection**
   ```bash
   oc logs -n gitea-system deployment/gitea | grep "PING DATABASE"
   # Should show: PING DATABASE postgres
   ```

4. **Test API Access**
   ```bash
   oc exec -n gitea-system deployment/gitea -- curl -s http://localhost:3000/api/v1/version
   # Should return: {"version":"1.19.4"}
   ```

### Initial Setup (Automatic)

The installation automatically creates:
- **Admin User**: `bootstrap` (password: `bootstrap123`)
- **Repository**: `bootstrap/bootstrap` 
- **ArgoCD Integration**: Repository secret configured

## Usage

### Enhanced bin/cluster-generate

The enhanced script supports automatic Gitea integration:

**Basic Usage (Local Only):**
```bash
./bin/cluster-generate regions/us-east-2/my-cluster/
```

**Gitea Integration (Recommended):**
```bash
./bin/cluster-generate --push-to-gitea regions/us-east-2/my-cluster/
```

### Workflow Comparison

#### Traditional Workflow
```bash
# 1. Generate cluster
./bin/cluster-generate regions/us-east-2/my-cluster/

# 2. Manual git operations
git add .
git commit -m "Add my-cluster"
git push origin main

# 3. Wait for GitHub sync
# 4. ArgoCD can now sync
```

#### Gitea Workflow  
```bash
# 1. Generate and push automatically
./bin/cluster-generate --push-to-gitea regions/us-east-2/my-cluster/

# 2. ArgoCD syncs immediately (no manual steps)
```

### ArgoCD ApplicationSet Configuration

To use Gitea as a source repository, update your ApplicationSet:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-cluster-applications
spec:
  template:
    spec:
      source:
        repoURL: http://gitea.gitea-system.svc.cluster.local:3000/bootstrap/bootstrap.git
        targetRevision: main
```

## Configuration Details

### Database Configuration
- **Type**: PostgreSQL 16.2
- **Storage**: 5Gi persistent volume
- **Connection**: Internal service `gitea-db.gitea-system.svc.cluster.local:5432`

### Gitea Configuration
- **Version**: 1.19.4
- **Storage**: 10Gi persistent volume for repositories
- **Domain**: `gitea.gitea-system.svc.cluster.local`
- **Port**: 3000 (HTTP)

### Security Configuration
- **Authentication**: Local user accounts
- **Admin User**: `bootstrap` (first user, automatically admin)
- **Registration**: Disabled by default
- **Network**: Internal cluster access only (no external exposure)

### ArgoCD Integration
- **Repository URL**: `http://gitea.gitea-system.svc.cluster.local:3000/bootstrap/bootstrap.git`
- **Authentication**: Username/password via Kubernetes secret
- **Secret Name**: `gitea-bootstrap-repo` (in `openshift-gitops` namespace)

## Troubleshooting

### Common Issues

#### 1. Gitea Pod Not Starting
```bash
# Check pod status
oc get pods -n gitea-system

# Check logs
oc logs -n gitea-system deployment/gitea

# Common fix: Restart deployment
oc rollout restart deployment/gitea -n gitea-system
```

#### 2. Database Connection Issues
```bash
# Check PostgreSQL status
oc get pods -n gitea-system -l name=gitea-db

# Check database logs
oc logs -n gitea-system deployment/gitea-db

# Verify connection from Gitea
oc exec -n gitea-system deployment/gitea -- pg_isready -h gitea-db.gitea-system.svc.cluster.local -p 5432
```

#### 3. ArgoCD Cannot Connect to Repository
```bash
# Check repository secret
oc get secret gitea-bootstrap-repo -n openshift-gitops -o yaml

# Test connectivity from ArgoCD
oc exec -n openshift-gitops deployment/argocd-server -- curl -s http://gitea.gitea-system.svc.cluster.local:3000/api/v1/version
```

#### 4. Push to Gitea Fails
```bash
# Check recent push job logs
oc get jobs -n gitea-system | grep gitea-push

# Check job logs
oc logs job/gitea-push-[cluster-name] -n gitea-system

# Common fix: Verify credentials
oc exec -n gitea-system deployment/gitea -- curl -u "bootstrap:bootstrap123" http://localhost:3000/api/v1/user
```

### Recovery Procedures

#### Reset Gitea Installation
```bash
# Delete all Gitea resources
oc delete -k operators/gitea/global/

# Wait for cleanup
sleep 30

# Redeploy
oc apply -k operators/gitea/global/
```

#### Reset Repository
```bash
# Delete and recreate repository
oc exec -n gitea-system deployment/gitea -- curl -X DELETE -u "bootstrap:bootstrap123" http://localhost:3000/api/v1/repos/bootstrap/bootstrap

# Recreate repository
oc exec -n gitea-system deployment/gitea -- curl -X POST -u "bootstrap:bootstrap123" http://localhost:3000/api/v1/user/repos -H "Content-Type: application/json" -d '{"name": "bootstrap", "description": "Bootstrap cluster configurations", "private": false, "auto_init": true}'
```

## Monitoring and Maintenance

### Health Checks
```bash
# Overall health
oc get pods -n gitea-system

# Gitea API health
oc exec -n gitea-system deployment/gitea -- curl -s http://localhost:3000/api/healthz

# Database health  
oc exec -n gitea-system deployment/gitea-db -- pg_isready

# Repository access
oc exec -n gitea-system deployment/gitea -- curl -s -u "bootstrap:bootstrap123" http://localhost:3000/api/v1/repos/bootstrap/bootstrap
```

### Resource Usage
```bash
# Check resource consumption
oc top pods -n gitea-system

# Check storage usage
oc get pvc -n gitea-system
```

### Backup Considerations
- **Database**: PostgreSQL data in `gitea-db` PVC
- **Repositories**: Git data in `gitea-data` PVC  
- **Configuration**: All config in this repository (`operators/gitea/global/`)

## Integration with Getting Started Guide

This Gitea installation integrates with the bootstrap project's getting started workflow:

1. **[First Cluster Guide](../../docs/getting-started/first-cluster.md)**: Use `--push-to-gitea` flag for immediate ArgoCD sync
2. **[Quickstart Guide](../../docs/getting-started/quickstart.md)**: Eliminates manual git push steps
3. **[Concepts Guide](../../docs/getting-started/concepts.md)**: Gitea serves as internal GitOps repository

For complete cluster provisioning workflow, see: [Getting Started Documentation](../../docs/getting-started/README.md)

## Advanced Configuration

### Custom Domain Configuration
To use a different domain, update the `app-ini-configmap.yaml`:
```yaml
data:
  app.ini: |
    [server]
    DOMAIN = your-custom-domain.local
    ROOT_URL = http://your-custom-domain.local:3000/
```

### External Access (Optional)
To expose Gitea externally, the Route is already configured:
```bash
# Get external URL
oc get route gitea -n gitea-system -o jsonpath='{.spec.host}'
```

### Repository Mirroring
To sync with GitHub for backup:
```bash
# Configure push mirror in Gitea UI or via API
# This allows automatic sync to GitHub while maintaining internal workflow
```

## Security Considerations

- **Internal Only**: Default configuration provides internal-cluster access only
- **Authentication**: Uses local user accounts (can be integrated with LDAP/OAuth)
- **Network Policies**: Consider implementing network policies for additional isolation
- **Secrets Management**: Repository credentials managed via Kubernetes secrets
- **TLS**: Currently HTTP-only for internal use (can be upgraded to HTTPS)

## Performance Tuning

### Resource Limits
Current configuration uses minimal resources. For heavy usage, consider:

```yaml
# In deployment.yaml
resources:
  requests:
    memory: "500Mi"
    cpu: "200m"
  limits:
    memory: "1Gi" 
    cpu: "500m"
```

### Database Tuning
For large repositories or many users:
```yaml
# In postgres.yaml environment variables
- name: POSTGRES_SHARED_BUFFERS
  value: "256MB"
- name: POSTGRES_MAX_CONNECTIONS
  value: "200"
```

## Comparison with Alternatives

| Feature | Gitea | GitLab | GitHub |
|---------|-------|--------|--------|
| Resource Usage | 200MB | 4GB+ | External |
| Setup Time | 30 minutes | 2-3 days | N/A |
| Internal Access | ✅ | ✅ | ❌ |
| Enterprise Support | Community | ✅ | ✅ |
| CI/CD Integration | Basic | Advanced | Advanced |
| Git Functionality | ✅ | ✅ | ✅ |

**Conclusion**: Gitea provides the optimal balance of functionality, resource efficiency, and ease of management for solving the ArgoCD dependency issue.