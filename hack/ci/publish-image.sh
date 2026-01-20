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

IMAGE_NAME="${IMAGE_NAME:-}"
GIT_TAG="${GIT_TAG:-}"

if [ -z "${IMAGE_NAME}" ]; then
  echodate "Error: IMAGE_NAME environment variable is not set."
  exit 1
fi

echodate "Publishing image: ${IMAGE_NAME}"
docker push "${IMAGE_NAME}"

if [ -n "${GIT_TAG}" ]; then
  BASE_IMAGE_NAME="${IMAGE_NAME%:*}"
  echodate "Git tag detected, pushing images based on GIT_TAG."
  echodate "Image to push: ${BASE_IMAGE_NAME}:${GIT_TAG}"
  echodate "Image to push: ${BASE_IMAGE_NAME}:latest"
  docker tag "${IMAGE_NAME}" "${BASE_IMAGE_NAME}:${GIT_TAG}"
  docker tag "${IMAGE_NAME}" "${BASE_IMAGE_NAME}:latest"

  docker push "${BASE_IMAGE_NAME}:${GIT_TAG}"
  docker push "${BASE_IMAGE_NAME}:latest"
else
  echodate "No GIT_TAG detected, skipping pushing images based on GIT_TAG."
fi

echodate "Script finished successfully."
