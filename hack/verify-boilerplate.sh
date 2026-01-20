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

echodate "Checking CE licenses..."
boilerplate -verbose \
  -boilerplates hack/boilerplate/ce \
  -exclude 'applications/*' \
  -exclude 'deploy/charts/*' \
  -exclude 'deploy/crd/*.y*ml' \
  -exclude 'charts/*'

echodate "Checking Kubermatic EE licenses..."
boilerplate -verbose \
  -boilerplates hack/boilerplate/ee \
  -exclude 'internal/*' \
  -exclude 'pkg/*' \
  -exclude '.prow/*' \
  -exclude '.github/*' \
  -exclude 'hack/*' \
  -exclude 'cmd/*' \
  -exclude 'Dockerfile' \
  -exclude 'deploy/charts/*' \
  -exclude 'charts/*' \
  -exclude 'tests/*' \
  -exclude 'deploy/crd/*.y*ml' \
  -exclude 'deploy/samples/*.y*ml' \
  -exclude '.*.y*ml'
