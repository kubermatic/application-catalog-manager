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

APPLICATIONS_DIR="${1:-applications}"
ERRORS=0

echodate "Verifying applications in ${APPLICATIONS_DIR}/"

if [[ ! -d "${APPLICATIONS_DIR}" ]]; then
  echodate "Error: Applications directory '${APPLICATIONS_DIR}' does not exist"
  exit 1
fi

validate_application_yaml() {
  local app_dir="$1"
  local app_file="${app_dir}/application.yaml"

  if [[ ! -f "${app_file}" ]]; then
    echodate "Missing application.yaml"
    return 1
  fi

  local kind=$(yq eval '.kind' "${app_file}" 2> /dev/null || echo "null")
  if [[ "${kind}" != "ApplicationDefinition" ]]; then
    echodate "application.yaml: kind should be 'ApplicationDefinition', found '${kind}'"
    return 1
  fi

  local api_version=$(yq eval '.apiVersion' "${app_file}" 2> /dev/null || echo "null")
  if [[ "${api_version}" == "null" ]]; then
    echodate "application.yaml: missing apiVersion"
    return 1
  fi

  local name=$(yq eval '.metadata.name' "${app_file}" 2> /dev/null || echo "null")
  if [[ "${name}" == "null" ]]; then
    echodate "application.yaml: missing metadata.name"
    return 1
  fi

  return 0
}

validate_metadata_yaml() {
  local app_dir="$1"
  local metadata_file="${app_dir}/metadata.yaml"

  if [[ ! -f "${metadata_file}" ]]; then
    echodate "Missing metadata.yaml"
    return 1
  fi

  local tier=$(yq eval '.tier' "${metadata_file}" 2> /dev/null || echo "null")
  if [[ "${tier}" == "null" || "${tier}" == "" ]]; then
    echodate "metadata.yaml: missing or empty 'tier' field"
    return 1
  fi

  return 0
}

if ! command -v yq &> /dev/null; then
  echodate "Error: yq is required but not installed"
  exit 1
fi

for app_dir in "${APPLICATIONS_DIR}"/*/; do
  if [[ -d "${app_dir}" ]]; then
    app_name=$(basename "${app_dir}")
    echodate "Checking ${app_name}:"

    if ! validate_application_yaml "${app_dir}"; then
      ((ERRORS++))
    fi

    if ! validate_metadata_yaml "${app_dir}"; then
      ((ERRORS++))
    fi

    if [[ $ERRORS -eq 0 ]]; then
      echodate "  Validations passed for ${app_dir}"
    fi
  fi
done

if [[ $ERRORS -eq 0 ]]; then
  echodate "All applications are valid!"
  exit 0
else
  echodate "X Found ${ERRORS} validation errors"
  exit 1
fi
