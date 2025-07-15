# Bootstrap config

GitOps applications and scripts to bootstrap managed OpenShift in a box

# Install Base

Install ARGO CD

```
$ oc create namespace argocd

$ oc apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

```

# Create ArgoCD Applications (i.e, start gitops)

```

$ oc create namespace openshift-gitops
$ oc apply -k gitops-applications

```

This is the entry point for the MOSBox.

creates argo application custom resources that point back to this repo for other yamls bundles.

> Ideally: Final step. Everything is auto-applied from here.

## Directory structure

### gitops-applications

#### kustomization.yaml

Init the gitops platform -- lists the other applications yamls to apply

points to application yamls in this same repo.

### applications

a single directory per application deployed

#### kustomization.yaml

Init the application.

##

#
