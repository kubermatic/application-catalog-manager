#!/usr/bin/env bash

# Copyright 2025 The Application Catalog Manager contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

cd "$(dirname "$0")/.."
source "hack/lib.sh"

CLUSTER_NAME="${CLUSTER_NAME:-app-catalog-test}"
APPLICATIONS_DIR="${APPLICATIONS_DIR:-applications}"
SKIP_REGISTRY="${SKIP_REGISTRY:-false}"

cleanup() {
  echodate "Cleaning up resources..."
  rm kind-config.yaml || true
}

create_kind_config() {
  echodate "Creating KinD cluster configuration..."
  cat >kind-config.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: $CLUSTER_NAME
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
EOF
}

create_cluster() {
  if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
    echodate "Cluster '$CLUSTER_NAME' already exists. Skipping cluster creation..."
    return 0
  fi

  create_kind_config

  echodate "Creating KinD cluster '$CLUSTER_NAME'..."
  kind create cluster --config kind-config.yaml --wait 300s

  echodate "Waiting for cluster to be ready..."
  kubectl wait --for=condition=Ready nodes --all --timeout=300s
}

load_manager() {
  IMAGE_TAG=${IMAGE_TAG:-testing}
  IMAGE_TAG=$IMAGE_TAG make docker-image

  kubectl scale deployment --replicas 0 -n kubermatic app-manager-application-catalog || echodate "no existing deployment to scale down"
  kubectl scale deployment --replicas 0 -n kubermatic app-manager-application-catalog-webhook || echodate "no existing webhook deployment to scale down"
  kind load docker-image --name $CLUSTER_NAME quay.io/kubermatic/application-catalog-manager:$IMAGE_TAG

  helm upgrade --install app-manager ./deploy/charts/application-catalog -n kubermatic --create-namespace
  kubectl scale deployment --replicas 1 -n kubermatic app-manager-application-catalog
}

install_cert_manager() {
  if kubectl get deployment cert-manager -n cert-manager &>/dev/null; then
    echodate "cert-manager is already installed. Skipping..."
    return 0
  fi

  echodate "Installing cert-manager for webhook TLS..."
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.5/cert-manager.yaml

  echodate "Waiting for cert-manager to be ready..."
  kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager -n cert-manager
  kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-webhook -n cert-manager
  kubectl wait --for=condition=Available --timeout=300s deployment/cert-manager-cainjector -n cert-manager

  echodate "cert-manager installation complete"
}

install_crds() {
  echodate "Installing CRDs..."
  kubectl apply -f https://raw.githubusercontent.com/buraksekili/kubermatic/refs/heads/feat/application-catalog/pkg/crd/k8c.io/kubermatic.k8c.io_kubermaticconfigurations.yaml
  kubectl apply -f https://raw.githubusercontent.com/buraksekili/kubermatic/refs/heads/feat/application-catalog/pkg/crd/k8c.io/apps.kubermatic.k8c.io_applicationdefinitions.yaml

  # Install ApplicationCatalog CRD
  kubectl apply -f deploy/crd/applicationcatalog.k8c.io_applicationcatalogs.yaml

  echodate "CRDs installed"
}

main() {
  trap cleanup EXIT

  command -v kind >/dev/null 2>&1 || {
    echodate "ERROR: kind is required but not installed."
    exit 1
  }
  command -v kubectl >/dev/null 2>&1 || {
    echodate "ERROR: kubectl is required but not installed."
    exit 1
  }
  command -v helm >/dev/null 2>&1 || {
    echodate "ERROR: helm is required but not installed."
    exit 1
  }

  create_cluster || {
    echodate "ERROR: Failed to create cluster"
    exit 1
  }

  install_crds || {
    echodate "ERROR: Failed to install CRDs"
    exit 1
  }

  install_cert_manager || {
    echodate "ERROR: Failed to install cert-manager"
    exit 1
  }

  kubectl create namespace kubermatic || echo "kubermatic namespace exists"
  load_manager || {
    echodate "ERROR: Failed to load manager"
    exit 1
  }

  echodate "Setup completed successfully!"
}

main "$@"
