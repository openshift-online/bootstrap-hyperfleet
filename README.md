# Bootstrap

This repository contains the scripting, configuration, GitOps content, and documentation necessary to bootstrap a region of a cloud service for Red Hat, based on OpenShift and the Red Hat products and supported community projects that we leverage in our reference architecture.  


# Install

Basic manual install

```mermaid
sequenceDiagram
   actor Admin
   actor Installer
   actor OCP
   actor Argo
   actor Git
   actor Tketon
   
   Admin->>Installer: Provision/Get Cluster
   Installer->>Admin: obtain KubeConfig
   
   Admin->>OCP: run ./bootstrap.sh
   OCP->>Admin: RHCP (ACM, Argo, config) is installed
      
   Argo->>Git: Get desired state
   Git->>Argo: Regional OCM YAML
   
   Argo->>OCP: Argo applies YAML
   OCP->>Argo: Deployment status
   
   Admin->>OCP: oc get pods 
   OCP->>Admin: Regional OCM is green 
    
```