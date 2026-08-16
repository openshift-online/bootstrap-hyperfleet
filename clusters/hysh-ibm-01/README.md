# hysh-ibm-01 — IBM ROKS cluster profile

The GitOps profile for the IBM Cloud ROKS cluster `hysh-ibm-01`. It is the ROKS
counterpart to [`clusters/global`](../global) (the ROSA/OCP hub): same operators
and HyperShell stack, but installed the way ROKS requires.

## Why ROKS is different

`hysh-ibm-01` is HyperShift-hosted, which imposes two hard constraints the hub
does not have:

1. **OperatorHub is unusable** — the marketplace catalog pods `ImagePullBackOff`,
   so OLM `Subscription`s never resolve.
2. **Node-level image mirroring is admission-denied** — `IDMS`/`ImageContentSourcePolicy`
   are rejected by the HostedCluster's `mirror` ValidatingAdmissionPolicy, and
   worker nodes can pull **only** IBM registries + the in-cluster registry
   (`image-registry.openshift-image-registry.svc:5000`).

So this profile installs every operator from **raw manifests** (no OLM) and
repoints **every** image to the internal registry mirror via Kustomize
`images:` — no cluster-level mirroring required. The hub keeps using OLM
Subscriptions; the shared `bases/` are reused unchanged wherever the CRs are
image-agnostic.

## Layout

```
clusters/hysh-ibm-01/
  operators/     raw operator profile (keycloak + cnpg + agent-sandbox + UWM),
                 images repointed to the mirror   -> kustomize build target
  gitops/        standalone app-of-apps root      -> oc apply -k target
  README.md      this file
```

`operators/` reuses `bases/operators/*/overlays/*` (the RAW overlays) and layers
the mirror `images:` block + two self-reference env patches on top. `gitops/`
holds the ArgoCD Applications, all targeting `destination: name: in-cluster`.

## How to deploy (hub-of-one)

ROKS cannot be an in-cluster OLM/ApplicationSet target of the hub, so it runs its
own GitOps:

```shell
# 1. Log into the ROKS cluster.
oc login <hysh-ibm-01 api>

# 2. Mirror the operator images into the internal registry FIRST (nodes cannot
#    reach the source registries). Push into `openshift` (globally pullable):
#      quay.io/keycloak/keycloak-operator:26.6.4    -> openshift/keycloak-operator:26.6.4
#      quay.io/keycloak/keycloak:26.6.4             -> openshift/keycloak:26.6.4
#      ghcr.io/cloudnative-pg/cloudnative-pg:1.26.1 -> openshift/cloudnative-pg:1.26.1
#    (agent-sandbox-controller is mirrored + repointed by its own overlay.)

# 3. Install openshift-gitops on the ROKS cluster (raw / manual — the hub's
#    openshift-gitops OLM overlay does not converge here).

# 4. Apply the ROKS app-of-apps. Its own ArgoCD reconciles the rest by wave.
oc apply -k clusters/hysh-ibm-01/gitops
```

Sync waves: **2** operators → **5** shared Keycloak (before any api-server; its
JWKS fetch is eager and fatal) → **10** HyperShell env(s) *(deferred, below)*.

## Validate locally

```shell
kustomize build clusters/hysh-ibm-01/operators   # raw operators, mirrored images
kustomize build clusters/hysh-ibm-01/gitops      # the two ArgoCD Applications
```

The operators build must contain **no** `quay.io`/`ghcr.io` refs — every image
(and the `RELATED_IMAGE_KEYCLOAK` / `OPERATOR_IMAGE_NAME` self-references the
operators launch their operand pods from) is repointed to the mirror.

## Deferred — needs a decision before wave 10

The HyperShell app envs (`hypershell-int` etc.) are **not** wired in yet. Two
things must be resolved first:

1. **App-image mirror overlay.** `bases/hypershell/overlays/int` points the
   api-server / control-plane / web-console images at
   `quay.io/redhat-services-prod/...:latest`, which ROKS nodes cannot pull. A
   ROKS hypershell overlay must mirror those three (like the operators profile
   does) before an env can sync.
2. **Hostname reconciliation.** `bases/hypershell/keycloak/keycloak.yaml` and the
   `int` overlay use the ingress domain
   `hypershell-cluster-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud`,
   but the live gateway host observed on this cluster is
   `…hysh-ibm-01-4c28435107377e996c6eb39230b7bcf5-0000.us-east.containers.appdomain.cloud`
   — same cluster hash, different cluster-name prefix. One is wrong for routing;
   confirm the real ROKS ingress subdomain and reconcile both files (or add a
   ROKS overlay that patches the hostname).

Optionally add `regions/us-east/hysh-ibm-01/region.yaml` (`kind: RegionalCluster`)
for desired-state metadata — deferred because no `roks`/`hcp` `spec.type` is
established in-repo yet (providers documented: eks, hcp, ocp).
