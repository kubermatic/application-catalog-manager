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

SHELL = /bin/bash -eu -o pipefail

export GOPATH?=$(shell go env GOPATH)
export CGO_ENABLED=0
export GOPROXY?=https://proxy.golang.org
export GO111MODULE=on
export GOFLAGS?=-mod=readonly -trimpath

# Build target OS/arch (only used for binary builds, not code generation)
BUILD_GOOS?=linux
BUILD_GOARCH?=amd64
export GIT_TAG ?= $(shell git tag --points-at HEAD)

GO_VERSION = 1.24.0

CMD = $(notdir $(wildcard ./cmd/*))
BUILD_DEST ?= _build

REGISTRY ?= quay.io
REGISTRY_NAMESPACE ?= kubermatic

IMAGE_TAG ?= \
		$(shell echo $$(git rev-parse HEAD && if [[ -n $$(git status --porcelain) ]]; then echo '-dirty'; fi)|tr -d ' ')
IMAGE_NAME ?= $(REGISTRY)/$(REGISTRY_NAMESPACE)/application-catalog-manager:$(IMAGE_TAG)

.PHONY: lint
lint:
	@golangci-lint --version
	golangci-lint run \
		--verbose \
		--print-resources-usage \
		./internal/... ./cmd/... ./pkg/...

.PHONY: verify
verify: lint
	./hack/verify-import-order.sh

.PHONY: verify-boilerplate
verify-boilerplate:
	./hack/verify-boilerplate.sh

.PHONY: all
all: build

.PHONY: build
build: $(CMD)

.PHONY: $(CMD)
$(CMD): %: $(BUILD_DEST)/%

$(BUILD_DEST)/%: cmd/%
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -v -o $@ ./cmd/$*

.PHONY: clean
clean:
	rm -rf $(BUILD_DEST)
	@echo "Cleaned $(BUILD_DEST)"

.PHONY: docker-image
docker-image:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(IMAGE_NAME) .

.PHONY: docker-image-publish
docker-image-publish: docker-image
	IMAGE_NAME="$(IMAGE_NAME)" ./hack/ci/publish-image.sh

.PHONY: shfmt
shfmt:
	shfmt -w -sr -i 2 hack/*

.PHONY: update-codegen
update-codegen:
	./hack/update-codegen.sh

.PHONY: verify-codegen
verify-codegen:
	./hack/verify-codegen.sh

.PHONY: update-crds
update-crds:
	./hack/update-crds.sh

.PHONY: verify-crds
verify-crds:
	./hack/verify-crds.sh

.PHONY: generate
generate: update-codegen update-crds
