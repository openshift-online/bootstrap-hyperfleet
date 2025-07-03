#!/bin/bash

echo "Temporary workaround: Creating cluster namespaces so we can hack create necessary secrets"#

oc create namespace cluster-10 2>/dev/null
oc apply -f aws-creds.yaml -n cluster-10
oc apply -f pull-secret.yaml -n cluster-10

./wait.kube.sh secret aws-creds cluster-10 {.kind} Secret
./wait.kube.sh secret pull-secret cluster-10 {.kind} Secret

oc create namespace cluster-20 2>/dev/null
oc apply -f aws-creds.yaml -n cluster-20
oc apply -f pull-secret.yaml -n cluster-20

./wait.kube.sh secret aws-creds cluster-20 '{.kind}' Secret
./wait.kube.sh secret pull-secret cluster-20 '{.kind}' Secret


