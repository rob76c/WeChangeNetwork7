#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

function init_namespace() {
  local namespaces=$(echo "$ORG0_NS $ORG1_NS $ORG2_NS" | xargs -n1 | sort -u)
  for ns in $namespaces; do
    push_fn "Creating namespace \"$ns\""
    kubectl create namespace $ns || true
    pop_fn
  done
}

function delete_namespace() {
  local namespaces=$(echo "$ORG0_NS $ORG1_NS $ORG2_NS" | xargs -n1 | sort -u)
  for ns in $namespaces; do
    push_fn "Deleting namespace \"$ns\""
    kubectl delete namespace $ns || true
    pop_fn
  done
}

function init_storage_volumes() {
  push_fn "Provisioning volume storage"

  # Both KIND and k3s use the Rancher local-path provider.  In KIND, this is installed
  # as the 'standard' storage class, and in Rancher as the 'local-path' storage class.
  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    export STORAGE_CLASS="standard"

  elif [ "${CLUSTER_RUNTIME}" == "k3s" ]; then
    export STORAGE_CLASS="local-path"

  else
    echo "Unknown CLUSTER_RUNTIME ${CLUSTER_RUNTIME}"
    exit 1
  fi

  cat kube/pvc-fabric-orderer.yaml | envsubst | kubectl -n $ORG0_NS create -f - || true
  cat kube/pvc-fabric-wechange.yaml | envsubst | kubectl -n $ORG1_NS create -f - || true
  cat kube/pvc-fabric-fednow.yaml | envsubst | kubectl -n $ORG2_NS create -f - || true

  pop_fn
}

function load_org_config() {
  push_fn "Creating fabric config maps"

  kubectl -n $ORG0_NS delete configmap orderer-config || true
  kubectl -n $ORG1_NS delete configmap wechange-config || true
  kubectl -n $ORG2_NS delete configmap fednow-config || true

  kubectl -n $ORG0_NS create configmap orderer-config --from-file=config/orderer
  kubectl -n $ORG1_NS create configmap wechange-config --from-file=config/wechange
  kubectl -n $ORG2_NS create configmap fednow-config --from-file=config/fednow

  pop_fn
}

function apply_k8s_builder_roles() {
  push_fn "Applying k8s chaincode builder roles"

  apply_template kube/fabric-builder-role.yaml $ORG1_NS
  apply_template kube/fabric-builder-rolebinding.yaml $ORG1_NS

  pop_fn
}

function apply_k8s_builders() {
  push_fn "Installing k8s chaincode builders"

  apply_template kube/wechange/wechange-install-k8s-builder.yaml $ORG1_NS
  apply_template kube/fednow/fednow-install-k8s-builder.yaml $ORG2_NS

  kubectl -n $ORG1_NS wait --for=condition=complete --timeout=60s job/wechange-install-k8s-builder
  kubectl -n $ORG2_NS wait --for=condition=complete --timeout=60s job/fednow-install-k8s-builder

  pop_fn
}