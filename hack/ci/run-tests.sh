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

cd $(dirname $0)/../..
source hack/lib.sh

if [[ ! -z "${JOB_NAME:-}" ]] && [[ ! -z "${PROW_JOB_ID:-}" ]]; then
  start_docker_daemon_ci
fi

echodate "Setting up test environment..."
hack/env.sh || {
  echodate "ERROR: Failed to setup test environment"
  exit 1
}

echodate "Running tests..."

export KUBECONFIG=~/.kube/config
protokol --kubeconfig "$KUBECONFIG" --flat --output "$ARTIFACTS/logs/kubermatic" --namespace kubermatic > /dev/null 2>&1 &

go test -v ./tests/e2e/...
