.PHONY: help all bin controller node test image limage ubi push clean

ifdef DOCKER_HUB_REPOSITORY
DOCKER_HUB_REPOSITORY := $(DOCKER_HUB_REPOSITORY)
else
DOCKER_HUB_REPOSITORY := ghcr.io/seagate
endif

ifdef VERSION
VERSION := $(VERSION)
else
VERSION := v1.0.13
endif

VERSION_FLAG = -X github.com/Seagate/seagate-exos-x-csi/pkg/common.Version=$(VERSION)

ifndef BIN
	BIN = seagate-exos-x-csi
endif

IMAGE = $(DOCKER_HUB_REPOSITORY)/$(BIN):$(VERSION)

help:
	@echo ""
	@echo "Build Targets:"
	@echo "-----------------------------------------------------------------------------------"
	@echo "make clean      - remove '$(BIN)-controller' and '$(BIN)-node'"
	@echo "make all        - clean, create driver images, create ubi docker image, push to registry"
	@echo "make bin        - create controller and node driver images"
	@echo "make controller - create controller driver image ($(BIN)-controller)"
	@echo "make node       - create node driver image ($(BIN)-node)"
	@echo "make test       - build test/sanity"
	@echo "make image      - create a repo docker image ($(IMAGE))"
	@echo "make limage     - create a local docker image ($(IMAGE))"
	@echo "make ubi        - create a local docker image using Redhat UBI ($(IMAGE))"
	@echo "make push       - push the docker image to '$(DOCKER_HUB_REPOSITORY)'"
	@echo ""

all: clean bin ubi push

bin: controller node

controller:
	@echo ""
	@echo "[] controller"
	go build -v -ldflags "$(VERSION_FLAG)" -o $(BIN)-controller ./cmd/controller

node:
	@echo ""
	@echo "[] node"
	go build -v -ldflags "$(VERSION_FLAG)" -o $(BIN)-node ./cmd/node

test:
	@echo ""
	@echo "[] test"
	./test/sanity

image:
	@echo ""
	@echo "[] image"
	docker build -t $(IMAGE) --build-arg version="$(VERSION)" --build-arg vcs_ref="$(shell git rev-parse HEAD)" --build-arg build_date="$(shell date --rfc-3339=seconds)" .

limage:
	@echo ""
	@echo "[] limage"
	docker build -f Dockerfile.local -t $(IMAGE) --build-arg version="$(VERSION)" --build-arg vcs_ref="$(shell git rev-parse HEAD)" --build-arg build_date="$(shell date --rfc-3339=seconds)" .

ubi:
	@echo ""
	@echo "[] ubi"
	docker build -f Dockerfile.ubi -t $(IMAGE) --build-arg version="$(VERSION)" --build-arg vcs_ref="$(shell git rev-parse HEAD)" --build-arg build_date="$(shell date --rfc-3339=seconds)" .

push:
	@echo ""
	@echo "[] push"
	docker push $(IMAGE)

clean:
	@echo ""
	@echo "[] clean"
	rm -vf $(BIN)-controller $(BIN)-node
