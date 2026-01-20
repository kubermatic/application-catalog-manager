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

echodate "Verifying generated code is up-to-date"

# Create a temporary copy of the pkg directory
tmpdir=$(mktemp -d)
trap "rm -rf $tmpdir" EXIT

cp -r pkg/ "$tmpdir/pkg"

# Run code generation
./hack/update-codegen.sh

# Compare the results
if ! diff -Naupr "$tmpdir/pkg" pkg/; then
  echodate "Generated code is out of date. Please run 'make update-codegen' and commit the changes."
  exit 1
fi

echodate "Generated code is up-to-date"
