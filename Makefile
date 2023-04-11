.PHONY: help all bin controller node test image limage ubi openshift push clean

ifdef DOCKER_HUB_REPOSITORY
DOCKER_HUB_REPOSITORY := $(DOCKER_HUB_REPOSITORY)
else
DOCKER_HUB_REPOSITORY := ghcr.io/seagate
endif

ifdef VERSION
VERSION := $(VERSION)
else
VERSION := v1.5.7
endif

VERSION_FLAG = -X github.com/Seagate/seagate-exos-x-csi/pkg/common.Version=$(VERSION)

ifndef BIN
	BIN = seagate-exos-x-csi
endif

HELM_VERSION := 1.0.1
HELM_KEY := css-host-software
HELM_IMAGE_REPO := $(DOCKER_HUB_REPOSITORY)/$(BIN)
# $HELM_KEY should be the name of a secret key in the invoker's default keyring
ifneq (,$(HELM_KEY))
  HELM_KEYRING := ~/.gnupg/secring.gpg
  HELM_SIGN := --sign --key $(HELM_KEY) --keyring $(HELM_KEYRING)
endif 
HELM_PACKAGE := $(BIN)-$(HELM_VERSION).tgz

IMAGE = $(DOCKER_HUB_REPOSITORY)/$(BIN):$(VERSION)

help:
	@echo ""
	@echo "Build Targets:"
	@echo "-----------------------------------------------------------------------------------"
	@echo "make all          - clean, create driver images, create ubi docker image, push to registry"
	@echo "make bin          - create controller and node driver images"
	@echo "make clean        - remove '$(BIN)-controller' and '$(BIN)-node'"
	@echo "make controller   - create controller driver image ($(BIN)-controller)"
	@echo "make helm-package - create signed helm package using HELM_VERSION, HELM_KEY environment variables"
	@echo "make image        - create a repo docker image ($(IMAGE))"
	@echo "make limage       - create a local docker image ($(IMAGE))"
	@echo "make node         - create node driver image ($(BIN)-node)"
	@echo "make openshift    - Create OpenShift-certification candidate image ($(IMAGE))"
	@echo "make push         - push the docker image to '$(DOCKER_HUB_REPOSITORY)'"
	@echo "make test         - build test/sanity"
	@echo "make ubi          - create a local docker image using Redhat UBI ($(IMAGE))"
	@echo ""

all: clean bin openshift push
openshift-all: clean openshift push

bin: protoc controller node

protoc: 
	@echo ""
	@echo "[] protocol buffers"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/node_service/node_servicepb/node_rpc.proto

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

openshift:
	@echo ""
	@echo "[] openshift"
	sed < Dockerfile.redhat > Dockerfile.tmp \
		-e 's/^ARG version=.*/ARG version=$(VERSION)/' \
		-e 's/^ARG vcs_ref=.*/ARG vcs_ref=$(strip $(shell git rev-parse HEAD))/' \
		-e 's/^ARG build_date=.*/ARG build_date=$(strip $(shell date --utc -Iseconds))/'
	cmp Dockerfile.redhat Dockerfile.tmp && rm Dockerfile.tmp || mv Dockerfile.tmp Dockerfile.redhat
	docker build -f Dockerfile.redhat -t $(IMAGE) .
	docker inspect $(IMAGE)

PREFLIGHT=../openshift-preflight/preflight
PREFLIGHT_REGISTRY=localhost:5000
PREFLIGHT_IMAGE=$(PREFLIGHT_REGISTRY)/$(BIN):$(VERSION)
# PREFLIGHT_OPTIONS would typically include "--certification-project-id=xxx --pyxis-api-token=xxx"
PREFLIGHT_OPTIONS:=$(strip $(shell test ! -f .preflight_options || cat .preflight_options))
PREFLIGHT_SUBMIT=

preflight:
	-docker run -d -p 5000:5000 --name registry registry:2 # make sure local registry is running
	docker tag $(IMAGE) $(PREFLIGHT_IMAGE)
	docker push $(PREFLIGHT_IMAGE)
	$(PREFLIGHT) check container $(PREFLIGHT_SUBMIT) $(PREFLIGHT_OPTIONS) $(PREFLIGHT_IMAGE)

preflight-submit: .preflight_options
	$(MAKE) preflight PREFLIGHT_SUBMIT=--submit

build-preflight:
	(cd ..; git clone https://github.com/redhat-openshift-ecosystem/openshift-preflight.git)
	cd ../openshift-preflight && make build

push:
	@echo ""
	@echo "[] push"
	docker push $(IMAGE)

clean:
	@echo ""
	@echo "[] clean"
	rm -vf $(BIN)-controller $(BIN)-node *.zip *.tgz *.prov helm/$(BIN)-$(HELM_VERSION)*


# Create a helm package that can be installed from a remote HTTPS URL with, e.g.
# helm install seagate-csi https://<server>/<path>/seagate-exos-x-csi-1.0.0.tgz
helm-package: $(HELM_PACKAGE)

# To create a package without signing it, specify "make helm-package HELM_KEY="
# Note that helm doesn't support GPG v2.1 kbx files; if signing fails, try:
# gpg --export-secret-keys > ~/.gnupg/secring.gpg
$(HELM_PACKAGE):
	cd helm; helm package $(HELM_SIGN) \
		--set image.tag=$(VERSION) --set image.repository=$(HELM_IMAGE_REPO) \
		$$PWD/csi-charts
	cp -p helm/$@* .
ifdef HELM_KEYRING
	helm verify --keyring $(HELM_KEYRING) $@
	zip -r $(subst .tgz,-signed-helm-package.zip,$@) $@ $@.prov
endif
