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

# This script pushes all applications stored under /applications directory
# to the upstream Kubermatic OCI registry.

cd $(dirname $0)/../..
source hack/lib.sh

# Define variables needed.
QUAY_IO_USERNAME="${QUAY_IO_USERNAME:-}"
QUAY_IO_PASSWORD="${QUAY_IO_PASSWORD:-}"
FILES_DIR="${FILES_DIR:-applications/}"
QUAY_ADDR="quay.io/kubermatic/applications"
TAG="${TAG:-$(git rev-parse HEAD)}"

if [ -z "$QUAY_IO_USERNAME" ]; then
  echodate "QUAY_IO_USERNAME is not set. Please provide a valid username."
  exit 1
fi

if [ -z "$QUAY_IO_PASSWORD" ]; then
  echodate "QUAY_IO_PASSWORD is not set. Please provide a valid password."
  exit 1
fi

if [ -z "$TAG" ]; then
  echodate "TAG is not set. Please provide a valid tag."
  exit 1
fi

if ! command -v oras &> /dev/null; then
  install_oras
fi

if command -v oras &> /dev/null; then
  ORAS_BIN=$(command -v oras)
elif [ -f ./oras ]; then
  ORAS_BIN="./oras"
else
  echodate "ORAS CLI not found after installation"
  exit 1
fi

echodate "Using oras CLI: ${ORAS_BIN}"

echodate "Pushing applications from ${FILES_DIR} to OCI Registry at ${QUAY_ADDR}:${TAG}"

${ORAS_BIN} login -u ${QUAY_IO_USERNAME} -p ${QUAY_IO_PASSWORD} quay.io

if [ $? -ne 0 ]; then
  echodate "Failed to login to Quay.io. Please check your credentials."
  exit 1
fi

${ORAS_BIN} push ${QUAY_ADDR}:${TAG},latest $FILES_DIR --artifact-type application/vnd.kubermatic.application-catalog.v1

if [ $? -ne 0 ]; then
  echodate "Failed to push applications to Quay.io. Please check your configuration."
  exit 1
fi
