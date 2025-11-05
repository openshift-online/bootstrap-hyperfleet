# Vault + External Secrets Setup

## Prerequisites
- Running hub cluster, cluster-admin access

## Installation Steps

### 1. Deploy Vault and ESO
```bash
oc apply -k clusters/global/operators/vault/ && oc wait --for=condition=Ready pod -l app.kubernetes.io/instance=vault -n vault
```

### 2. Initialize Vault
```bash
# For development - get unseal key from logs
oc logs vault-0 -n vault | grep "Unseal Key"
# Output: Unseal Key: /4k5SX8RvkEZYYXIDqqDKbPyV1cClnZpjMrzZ9lurYQ=

# Initialize and unseal (dev environment auto-unseals)
oc exec vault-0 -n vault -- vault operator init -key-shares=1 -key-threshold=1
oc exec vault-0 -n vault -- vault operator unseal /4k5SX8RvkEZYYXIDqqDKbPyV1cClnZpjMrzZ9lurYQ=
oc exec vault-0 -n vault -- vault auth -method=token root
```

### 3. Setup Authentication
```bash
# Create ServiceAccount and RBAC
oc apply -f clusters/global/operators/vault/vault-auth-serviceaccount.yaml

# Enable Kubernetes auth
oc exec vault-0 -n vault -- vault auth enable kubernetes
oc exec vault-0 -n vault -- vault write auth/kubernetes/config \
    kubernetes_host="$(oc config view --minify -o jsonpath='{.clusters[0].cluster.server}')" \
    kubernetes_ca_cert="$(oc get secret vault-auth-token -n vault -o jsonpath='{.data.ca\.crt}' | base64 -d)" \
    disable_iss_validation=true
```

### 4. Configure Policies
```bash
# Create cluster secrets policy
echo 'path "secret/data/aws-credentials" { capabilities = ["read"] }
  path "secret/data/pull-secret" { capabilities = ["read"] }' | oc exec vault-0 -n vault -i -- vault policy write cluster-secrets -

# Create Kubernetes role
oc exec vault-0 -n vault -- vault write auth/kubernetes/role/cluster-role \
    bound_service_account_names=vault-auth \
    bound_service_account_namespaces=vault \
    policies=cluster-secrets \
    ttl=1h
```

### 5. Deploy ClusterSecretStore
```bash
oc apply -f clusters/global/operators/vault/cluster-secret-store.yaml
oc get clustersecretstore vault-cluster-store  # Should show STATUS=Valid
```

## Secret Management

### Store Secrets in Vault
```bash
# AWS credentials
oc exec vault-0 -n vault -- vault kv put secret/aws-credentials \
    aws_access_key_id="$(cat .secrets/aws.secret.id )" \
    aws_secret_access_key="$(cat .secrets/aws.secret.key )"

# Pull secret
oc exec vault-0 -n vault -- vault kv put secret/pull-secret \
    .dockerconfigjson="$(cat .secrets/pull-secret.txt)"
```

### Deploy Secrets to Clusters
```bash
# Per cluster
sed 's/CLUSTER_NAMESPACE/ocp-02/g' bases/deployments/ocm/external-secrets-template.yaml | oc apply -f -

# Multiple clusters
for cluster in ocp-02 ocp-03 eks-02; do
    sed "s/CLUSTER_NAMESPACE/$cluster/g" bases/deployments/ocm/external-secrets-template.yaml | oc apply -f -
done
```

## Status Checks
```bash
oc exec vault-0 -n vault -- vault status
oc get clustersecretstore vault-cluster-store
oc get externalsecret -A
oc get secret aws-credentials -n ocp-02
```

## Troubleshooting
```bash
# ESO logs
oc logs -n external-secrets deployment/external-secrets

# Vault connectivity test
oc exec -n external-secrets deployment/external-secrets -- \
    curl -s http://vault.vault.svc.cluster.local:8200/v1/sys/health

# Test authentication
oc exec vault-0 -n vault -- vault write auth/kubernetes/login \
    role=cluster-role \
    jwt="$(oc serviceaccounts get-token vault-auth -n vault)"
```