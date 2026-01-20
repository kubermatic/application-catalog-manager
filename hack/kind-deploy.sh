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

# Simple helper script to build, load, and deploy to a kind cluster.
#
# Usage:
#   ./hack/kind-deploy.sh                    # Uses defaults
#   ./hack/kind-deploy.sh -c my-cluster      # Specify kind cluster name
#   ./hack/kind-deploy.sh -t mytag           # Specify image tag
#   ./hack/kind-deploy.sh -n my-namespace    # Specify namespace
#   ./hack/kind-deploy.sh -s                 # Skip build (just redeploy)
#   ./hack/kind-deploy.sh -e                 # Run e2e tests after deploy

set -euo pipefail

cd "$(dirname "$0")/.."

# Defaults
KIND_CLUSTER="${KIND_CLUSTER:-app-catalog-test}"
IMAGE_TAG="${IMAGE_TAG:-testing}"
NAMESPACE="${NAMESPACE:-kubermatic}"
SKIP_BUILD="${SKIP_BUILD:-false}"
RUN_E2E="${RUN_E2E:-false}"

# Parse arguments
while getopts "c:t:n:seh" opt; do
  case $opt in
    c) KIND_CLUSTER="$OPTARG" ;;
    t) IMAGE_TAG="$OPTARG" ;;
    n) NAMESPACE="$OPTARG" ;;
    s) SKIP_BUILD="true" ;;
    e) RUN_E2E="true" ;;
    h)
      echo "Usage: $0 [-c cluster] [-t tag] [-n namespace] [-s] [-e]"
      echo "  -c  Kind cluster name (default: app-catalog-test)"
      echo "  -t  Image tag (default: testing)"
      echo "  -n  Namespace (default: kubermatic)"
      echo "  -s  Skip build (just redeploy)"
      echo "  -e  Run e2e tests after deploy"
      exit 0
      ;;
    *) exit 1 ;;
  esac
done

IMAGE_NAME="quay.io/kubermatic/application-catalog-manager:${IMAGE_TAG}"

echo "==> Configuration:"
echo "    Kind cluster: ${KIND_CLUSTER}"
echo "    Image: ${IMAGE_NAME}"
echo "    Namespace: ${NAMESPACE}"
echo ""

# Build Docker image
if [[ "${SKIP_BUILD}" != "true" ]]; then
  echo "==> Building Docker image..."
  IMAGE_TAG="${IMAGE_TAG}" make docker-image
else
  echo "==> Skipping build (using existing image)"
fi

# Load image to kind
echo "==> Loading image to kind cluster..."
kind load docker-image "${IMAGE_NAME}" --name "${KIND_CLUSTER}"

# Deploy with Helm
echo "==> Deploying with Helm..."
helm upgrade --install app-manager ./deploy/charts/application-catalog \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --set image.tag="${IMAGE_TAG}" \
  --set webhook.debug=true \
  --wait

# Restart deployments to ensure new image is used
echo "==> Restarting deployments..."
kubectl rollout restart deployment -n "${NAMESPACE}" -l app.kubernetes.io/name=application-catalog

# Wait for rollout
echo "==> Waiting for rollout..."
kubectl rollout status deployment -n "${NAMESPACE}" -l app.kubernetes.io/name=application-catalog --timeout=120s

echo ""
echo "==> Deployment complete!"
kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=application-catalog

# Run e2e tests if requested
if [[ "${RUN_E2E}" == "true" ]]; then
  echo ""
  echo "==> Running e2e tests..."
  go test -v ./tests/e2e/... -count=1
fi
