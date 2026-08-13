# HyperShell Platform GitOps

GitOps configuration for deploying and operating HyperShell on OpenShift. Manages three progressive environments (int, stage, prod) with ArgoCD, providing gateway-based sandbox infrastructure for Red Hat engineers.

## Environments

| Environment | Namespace | Branch/Tag | Replicas | Purpose |
|-------------|-----------|------------|----------|---------|
| Integration | `hypershell-int` | `main` (`:latest`) | 1 | Auto-deploys every merge to main |
| Staging | `hypershell-stage` | sha256 digest | 1 | Promoted from int after validation |
| Production | `hypershell-prod` | sha256 digest | 2 | Promoted from stage |

### Current Endpoints (IBM ROKS us-east)

**Integration**

| Service | URL |
|---------|-----|
| HyperShell API | `https://int-hypershell-api-hypershell-int.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| Keycloak SSO | `https://int-keycloak-hypershell-int.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| gRPC (cluster-internal) | `int-hypershell-api-server.hypershell-int.svc:9000` |

**Staging**

| Service | URL |
|---------|-----|
| HyperShell API | `https://stage-hypershell-api-hypershell-stage.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| Keycloak SSO | `https://stage-keycloak-hypershell-stage.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |

**Production**

| Service | URL |
|---------|-----|
| HyperShell API | `https://prod-hypershell-api-hypershell-prod.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| Keycloak SSO | `https://prod-keycloak-hypershell-prod.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |

**Shared Services**

| Service | URL |
|---------|-----|
| ArgoCD | `https://openshift-gitops-server-openshift-gitops.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| Prometheus | `https://prometheus-k8s-openshift-monitoring.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| Thanos Querier | `https://thanos-querier-openshift-monitoring.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |
| Alertmanager | `https://alertmanager-main-openshift-monitoring.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud` |

## Quick Start

```bash
oc login https://api.your-cluster.example.com:6443
./bin/bootstrap
```

The bootstrap script applies all manifests and monitors convergence:

```
HyperShell Bootstrap                          https://api.cluster:6443
----------------------------------------------------------------------

Phase 1: Operators                                        [9/9 Ready]
  OpenShift GitOps .......................................  Ready
  OpenShift Pipelines ....................................  Ready
  CloudNativePG ..........................................  Ready
  Keycloak Operator ......................................  Ready
  Grafana Operator .......................................  Ready
  Vault ..................................................  Ready
  External Secrets Operator ..............................  Ready
  Agent Sandbox Controller ...............................  Ready
  User Workload Monitoring ...............................  Ready

Phase 2: HyperShell Environments                          [3/3 Ready]
  hypershell-int .........................................  Ready
  hypershell-stage .......................................  Ready
  hypershell-prod ........................................  Ready

Elapsed: 12m 34s                                  Ctrl+C safe (idempotent)
```

Options:

```bash
./bin/bootstrap              # Apply manifests + monitor
./bin/bootstrap --monitor    # Monitor only (skip apply)
./bin/bootstrap --help       # Show usage
```

## User Access

All Red Hat engineers from the OCM production roster (938 accounts) are pre-seeded into Keycloak with temporary passwords.

| Field | Value |
|-------|-------|
| Username | your kerberos ID (e.g. `mturansk`) |
| Password | `changeme` (temporary) |
| Group | `hypershell-users` |
| Role | `openshell-user` |

### First Login

Change your temporary password before using the CLI. Open the Keycloak account console:

```
https://int-keycloak-hypershell-int.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud/realms/hypershell/account
```

Sign in with your kerberos ID and `changeme`, then set a new password.

### Get a Token

```bash
curl -s -X POST \
  "https://int-keycloak-hypershell-int.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud/realms/hypershell/protocol/openid-connect/token" \
  -d "client_id=openshell-cli" \
  -d "grant_type=password" \
  -d "username=YOUR_KERBEROS_ID" \
  -d "password=YOUR_PASSWORD" | jq .access_token
```

### Admin Accounts

| Username | Password | Role |
|----------|----------|------|
| `admin` | `admin` | `openshell-admin` |
| `developer` | `developer` | `openshell-user` |

### Managing Users

Users live in `bases/hypershell/base/hypershell-realm.json`. Edit the `users` array, commit, and restart Keycloak:

```bash
oc rollout restart deployment -l app=keycloak -n hypershell-int
```

## Progressive Delivery

```
main branch commit
       |
       v
  hypershell-int (:latest, auto-deploy)
       |
       | promote sha digest
       v
  hypershell-stage (pinned digest)
       |
       | promote sha digest
       v
  hypershell-prod (pinned digest, 2 replicas)
```

To promote from int to stage, update `bases/hypershell/overlays/stage/kustomization.yaml`:

```yaml
images:
  - name: quay.io/redhat-services-prod/hcm-eng-prod-tenant/hypershell-main/hypershell-api-server-main
    digest: sha256:<promoted-digest>
  - name: quay.io/redhat-services-prod/hcm-eng-prod-tenant/hypershell-main/hypershell-control-plane-main
    digest: sha256:<promoted-digest>
```

Commit and push. ArgoCD syncs automatically.

## Monitoring

### Prometheus (User Workload Monitoring)

OpenShift User Workload Monitoring is enabled with 15-day retention and 40Gi storage. HyperShell application metrics are collected automatically via the platform Prometheus stack.

Access Prometheus directly (requires cluster auth):

```
https://prometheus-k8s-openshift-monitoring.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud
```

Query user-workload metrics via Thanos:

```
https://thanos-querier-openshift-monitoring.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud
```

### Grafana

The Grafana Operator (v5) is installed via OLM. A standalone Grafana 10.2 deployment with pre-configured dashboards is available in `clusters/global/operators/grafana/`. It connects to the in-cluster Prometheus and includes a cluster provisioning dashboard.

### Quick Health Checks

```bash
oc get pods -n hypershell-int
oc get pods -n hypershell-stage
oc get pods -n hypershell-prod

oc get applications.argoproj.io -n openshift-gitops

oc get csv -A | rg -v Succeeded
```

## Architecture

### Per-Environment Components

Each HyperShell environment deploys:

- **API Server** -- REST (8000) + gRPC (9000) with JWT/OIDC auth
- **Controller** -- reconciles Fleet, ManagedCluster, GatewayRelease, Gateway resources
- **PostgreSQL** -- CNPG-managed StatefulSet with persistent storage
- **Keycloak** -- OIDC provider with pre-configured `hypershell` realm

### Operators (ArgoCD sync-wave order)

| Wave | Operator | Purpose |
|------|----------|---------|
| -1 | OpenShift GitOps | ArgoCD -- manages everything else |
| 1 | OpenShift Pipelines | Tekton CI/CD |
| 1 | User Workload Monitoring | Prometheus for application metrics |
| 2 | CloudNativePG | PostgreSQL operator |
| 2 | Keycloak (RHBK) | OIDC/SSO operator |
| 2 | Grafana Operator | Dashboard management |
| 2 | External Secrets | Vault-backed secrets |
| 2 | Vault (Helm) | Secret storage |
| 3 | Vault Init | Vault configuration and Kubernetes auth |
| 3 | Agent Sandbox Controller | OpenShell sandbox runtime |

### Directory Structure

```
clusters/global/
  operators/           Operator subscriptions and configs
  gitops/              ArgoCD Applications
    global/            Operator apps (sync-wave -1 to 3)
    hypershell/        Environment apps (sync-wave 10-12)
  pipelines/           Tekton pipelines

bases/
  hypershell/
    base/              Shared manifests (api-server, controller, postgres, keycloak)
    overlays/
      int/             :latest tags, 1 replica
      stage/           sha256 digests, 1 replica
      prod/            sha256 digests, 2 replicas
  operators/           Reusable operator bases with channel overlays
    vault/             Vault init resources (helm repo, auth, secret store, init job)
    cloudnative-pg/    CNPG operator subscription
    keycloak-operator/ RHBK operator subscription
    grafana-operator/  Grafana operator subscription
    openshift-gitops/  GitOps operator subscription
    openshift-pipelines/ Pipelines operator subscription
    agent-sandbox-controller/ Sandbox controller deployment
```

### Vault

Vault provides centralized secret management via External Secrets Operator. Each managed cluster in the fleet gets its own Vault instance.

Setup details: [`clusters/global/operators/vault/VAULT-SETUP.md`](clusters/global/operators/vault/VAULT-SETUP.md)

```bash
oc exec vault-helm-0 -n vault -- vault status
oc get clustersecretstore vault-cluster-store
oc get externalsecret -A
```

### Kustomize repoURL Management

All ArgoCD Applications reference this repository via a centralized Kustomize replacement. The canonical URL is set once in `clusters/global/gitops/repo-url-patch.yaml` and propagated to all Application/ApplicationSet resources at build time. To fork this repo, update only that file.
