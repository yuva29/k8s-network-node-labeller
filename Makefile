# Import development related environment variables from dev.env
ifneq ("$(wildcard dev.env)","")
    include dev.env
endif

DOCKER_REGISTRY ?= docker.io/rocm
IMAGE_NAME ?= k8s-network-node-labeller
IMAGE_TAG_BASE ?= $(DOCKER_REGISTRY)/$(IMAGE_NAME)
IMAGE_TAG ?= dev
IMG ?= $(IMAGE_TAG_BASE):$(IMAGE_TAG)

IMG_FILE_NAME ?= $(IMAGE_NAME)-$(IMAGE_TAG)
BINARY_NAME ?= network-node-labeller
BUILDDIR=$(CURDIR)/build
IMG_DIR=$(BUILDDIR)/docker
DIRS=$(BUILDDIR) $(IMG_DIR)

GO_BUILD_OPTS ?=

.PHONY: all
all: build test

.PHONY: create-dirs
create-dirs:
	$(info Creating directories $(DIRS)...)
	mkdir -p $(DIRS)
	$(info Done!)

.PHONY: docker-build
docker-build:
ifeq ($(HOURLY_TAG_LABEL),)
	docker build -t ${IMG} .
else
	docker build --label HOURLY_TAG_LABEL=$(HOURLY_TAG_LABEL) -t ${IMG} .
endif

.PHONY: docker-save
docker-save: create-dirs
	$(info Saving docker image to $(IMG_FILE_NAME).tar.gz...)
	docker save $(IMG) | gzip > $(IMG_DIR)/$(IMG_FILE_NAME).tar.gz

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

copyrights:
	GOFLAGS=-mod=mod go run tools/build/copyright/main.go && ${MAKE} fmt && ./tools/build/check-local-files.sh
