# bin/bootstrap Requirements

## Purpose

Bootstrap the HyperShell hosting platform on an OpenShift cluster by deploying all operators, GitOps applications, and HyperShell environments (int/stage/prod) with a live progress dashboard.

## Functional Requirements

### Preflight Validation
- **Cluster Authentication**: Verify `oc cluster-info` succeeds
- **Permission Check**: Verify `oc auth can-i create subscriptions.operators.coreos.com --all-namespaces`
- **Exit Condition**: Exit code 1 with clear messaging on failure

### Deployment Sequence
1. **GitOps Operator**: Apply `clusters/global/operators/openshift-gitops` directly via `oc apply -k`
2. **CRD Gate**: Wait for `applications.argoproj.io` CRD to be Established (5min timeout)
3. **Full Apply**: Apply `clusters/global/` which deploys all ArgoCD Applications via sync-waves
4. **Monitor**: Enter live dashboard polling loop until all components report Ready

### Operator Components (Phase 1)

All operators deploy in parallel via ArgoCD sync-waves:

| Wave | Component | Readiness Check |
|------|-----------|----------------|
| -1 | OpenShift GitOps | CSV phase = Succeeded |
| 1 | OpenShift Pipelines | CSV phase = Succeeded |
| 1 | User Workload Monitoring | prometheus-operator Deployment ready |
| 2 | CloudNativePG | CSV phase = Succeeded |
| 2 | Keycloak Operator (RHBK) | CSV phase = Succeeded |
| 2 | Grafana Operator | CSV phase = Succeeded |
| 2 | External Secrets Operator | eso-external-secrets Deployment ready |
| 2 | Vault | vault-helm StatefulSet ready |
| 3 | Agent Sandbox Controller | agent-sandbox-controller Deployment ready |

### HyperShell Environments (Phase 2)

| Wave | Environment | Readiness Check |
|------|-------------|----------------|
| 10 | hypershell-int | api-server + postgres + controller + keycloak all ready |
| 11 | hypershell-stage | api-server + postgres + controller + keycloak all ready |
| 12 | hypershell-prod | api-server + postgres + controller + keycloak all ready |

### Live Progress Dashboard

The script renders an in-place terminal table that updates every poll cycle:

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

Status states: `Pending` (not created), `Installing` (in progress), `Ready` (healthy), `Failed` (error), `Waiting` (blocked on dependency), `Syncing` (ArgoCD reconciling).

### Idempotency

- Safe to Ctrl+C at any time; ArgoCD continues reconciling independently
- Safe to re-run; `oc apply -k` is idempotent, monitor resumes where it left off
- `--monitor` flag skips the apply phase and only observes current state

### Error Handling

- **Authentication**: Clear message with `oc login` guidance
- **Permissions**: Explicit cluster-admin requirement message
- **CRD Timeout**: 5-minute deadline for ArgoCD CRD with exit code 1
- **Component Failure**: Detect CSV Failed phase, print diagnostic commands
- **Overall Timeout**: Configurable via `BOOTSTRAP_TIMEOUT` (default 3600s)

## Usage

```bash
./bin/bootstrap              # Apply manifests + monitor until ready
./bin/bootstrap --monitor    # Monitor only (skip apply)
./bin/bootstrap --help       # Show usage

# Environment variables
BOOTSTRAP_POLL_INTERVAL=10   # Seconds between polls
BOOTSTRAP_TIMEOUT=3600       # Max wait in seconds
NO_COLOR=1                   # Disable ANSI color output
```

## Dependencies

### Kustomize Paths
- `clusters/global/operators/openshift-gitops` - GitOps operator (applied first, gates everything)
- `clusters/global/` - All operators, pipelines, gitops applications, HyperShell environments

### Required Permissions
- **cluster-admin** role for operator deployment and CRD creation

### External Tools
- `oc` - OpenShift CLI (authenticated to target cluster)
- `kustomize` - Used via `oc apply -k`

## Design Principles

- **Observe, don't orchestrate**: The script applies declarative state, then observes convergence. ArgoCD does the actual orchestration via sync-waves.
- **Idempotent by design**: Every operation is safe to repeat. No mutable state outside the cluster.
- **Terminal-native UX**: ANSI dashboard with NO_COLOR support, cursor management, and graceful cleanup on exit.
- **Self-contained**: No external script dependencies (replaces monitor-status, wait-kube).
