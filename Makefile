# Import development related environment variables from dev.env
ifneq ("$(wildcard dev.env)","")
    include dev.env
endif

DOCKER_REGISTRY ?= docker.io/rocm
IMAGE_NAME ?= k8s-network-node-labeller
IMAGE_TAG_BASE ?= $(DOCKER_REGISTRY)/$(IMAGE_NAME)
IMAGE_TAG ?= dev
IMG ?= $(IMAGE_TAG_BASE):$(IMAGE_TAG)

BINARY_NAME ?= network-node-labeller
BUILDDIR=$(CURDIR)/build
GO_BUILD_OPTS ?=

.PHONY: all
all: build test

.PHONY: create-dirs
create-dirs:
	$(info Creating directories $(BUILDDIR)...)
	mkdir -p $(BUILDDIR)
	$(info Done!)

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} .; \

.PHONY: docker-push
docker-push:
	docker push ${IMG}

.PHONY: build
build: create-dirs
	$(info Building $(BINARY_NAME)...)
	cd $(CURDIR)/cmd && $(GO_BUILD_OPTS) go build -o $(BUILDDIR)/$(BINARY_NAME) -v ./main.go
	$(info Done!)

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	$(info Running tests...)
	go test ./... -v
	$(info Done!)
