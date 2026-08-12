# HyperShell Platform

GitOps infrastructure for deploying and operating the HyperShell gateway platform on OpenShift. Manages operators, environments, and progressive delivery through ArgoCD.

## Quick Start

```bash
oc login https://api.your-cluster.example.com:6443
./bin/bootstrap
```

The bootstrap script applies all manifests, then monitors convergence with a live dashboard:

```
HyperShell Bootstrap                          https://api.cluster:6443
──────────────────────────────────────────────────────────────────────────

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

## HyperShell Environments

Three environments provide progressive delivery. Int auto-deploys from `main`, stage and prod pin to promoted sha digests.

### Integration (hypershell-int)

| Service | URL |
|---------|-----|
| API Server | `https://int-hypershell-api-hypershell-int.<ingress-domain>` |
| Keycloak SSO | `https://int-keycloak-hypershell-int.<ingress-domain>` |
| gRPC (internal) | `int-hypershell-api-server.hypershell-int.svc:9000` |

- **Namespace:** `hypershell-int`
- **Image tag:** `:latest` (auto-deploys main branch)
- **Replicas:** 1

### Staging (hypershell-stage)

| Service | URL |
|---------|-----|
| API Server | `https://stage-hypershell-api-hypershell-stage.<ingress-domain>` |
| Keycloak SSO | `https://stage-keycloak-hypershell-stage.<ingress-domain>` |
| gRPC (internal) | `stage-hypershell-api-server.hypershell-stage.svc:9000` |

- **Namespace:** `hypershell-stage`
- **Image tag:** pinned to sha256 digest (promoted from int)
- **Replicas:** 1

### Production (hypershell-prod)

| Service | URL |
|---------|-----|
| API Server | `https://prod-hypershell-api-hypershell-prod.<ingress-domain>` |
| Keycloak SSO | `https://prod-keycloak-hypershell-prod.<ingress-domain>` |
| gRPC (internal) | `prod-hypershell-api-server.hypershell-prod.svc:9000` |

- **Namespace:** `hypershell-prod`
- **Image tag:** pinned to sha256 digest (promoted from stage)
- **Replicas:** 2 (api-server, keycloak)

### Current Cluster Endpoints

On the IBM ROKS cluster (`hypershell-cluster`, us-east):

| Environment | API | Keycloak |
|-------------|-----|----------|
| int | https://int-hypershell-api-hypershell-int.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud | https://int-keycloak-hypershell-int.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud |
| stage | https://stage-hypershell-api-hypershell-stage.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud | https://stage-keycloak-hypershell-stage.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud |
| prod | https://prod-hypershell-api-hypershell-prod.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud | https://prod-keycloak-hypershell-prod.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud |

**ArgoCD:** https://openshift-gitops-server-openshift-gitops.hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud

## Progressive Delivery

```
main branch commit
       |
       v
  hypershell-int (:latest, auto-deploy)
       |
       | tests pass, promote sha
       v
  hypershell-stage (sha256 digest)
       |
       | validation, promote sha
       v
  hypershell-prod (sha256 digest, 2 replicas)
```

To promote a sha from int to stage, update `bases/hypershell/overlays/stage/kustomization.yaml`:

```yaml
images:
  - name: quay.io/redhat-services-prod/hcm-eng-prod-tenant/hypershell-main/hypershell-server
    digest: sha256:<promoted-digest>
  - name: quay.io/redhat-services-prod/hcm-eng-prod-tenant/hypershell-main/hypershell-controller
    digest: sha256:<promoted-digest>
```

Commit and push. ArgoCD syncs the change automatically.

## Architecture

### Components

Each HyperShell environment deploys:
- **API Server** — HTTP (8000) and gRPC (9000) with JWT auth and OIDC
- **Controller** — reconciles gateway provisioning
- **PostgreSQL** — StatefulSet with persistent storage
- **Keycloak** — OIDC provider with pre-configured realm

### Operators (deployed via ArgoCD sync-waves)

| Wave | Operator | Purpose |
|------|----------|---------|
| -1 | OpenShift GitOps | ArgoCD — manages everything else |
| 1 | OpenShift Pipelines | Tekton CI/CD |
| 1 | User Workload Monitoring | Prometheus for app metrics |
| 2 | CloudNativePG | PostgreSQL operator |
| 2 | Keycloak (RHBK) | OIDC/SSO operator |
| 2 | Grafana | Dashboards |
| 2 | External Secrets | Vault-backed secrets |
| 2 | Vault | Secret storage |
| 3 | Agent Sandbox Controller | OpenShell sandbox runtime |

### Directory Structure

```
clusters/global/
  operators/           Operator subscriptions and configs
  gitops/              ArgoCD Applications
    global/            Operator apps (sync-wave 1-3)
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
```

## Usage

```bash
./bin/bootstrap              # Apply manifests + monitor
./bin/bootstrap --monitor    # Monitor only (skip apply)
./bin/bootstrap --help       # Show usage

# Environment variables
BOOTSTRAP_POLL_INTERVAL=10   # Seconds between polls
BOOTSTRAP_TIMEOUT=3600       # Max wait in seconds
NO_COLOR=1                   # Disable ANSI color output
```

## Monitoring

```bash
# Check environment health
oc get pods -n hypershell-int
oc get pods -n hypershell-stage
oc get pods -n hypershell-prod

# ArgoCD application status
oc get applications.argoproj.io -n openshift-gitops

# Operator status
oc get csv -A | grep -v Succeeded
```
