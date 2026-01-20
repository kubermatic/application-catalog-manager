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

cd $(dirname $0)/..
source hack/lib.sh

# Ensure output directory exists
mkdir -p deploy/crd

echodate "Generating CRD manifests"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  crd \
  object:headerFile="hack/boilerplate/ce/boilerplate.go.txt" \
  paths="./pkg/..." \
  output:crd:artifacts:config=./deploy/crd

echodate "CRD generation complete"
