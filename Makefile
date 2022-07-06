.PHONY: help all bin controller node test image limage ubi push clean

ifdef DOCKER_HUB_REPOSITORY
DOCKER_HUB_REPOSITORY := $(DOCKER_HUB_REPOSITORY)
else
DOCKER_HUB_REPOSITORY := ghcr.io/seagate
endif

ifdef VERSION
VERSION := $(VERSION)
else
VERSION := v1.2.3
endif

VERSION_FLAG = -X github.com/Seagate/seagate-exos-x-csi/pkg/common.Version=$(VERSION)

ifndef BIN
	BIN = seagate-exos-x-csi
endif

IMAGE = $(DOCKER_HUB_REPOSITORY)/$(BIN):$(VERSION)

VAGRANT = vagrant

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
	@echo "make openshift  - Create OpenShift-certification candidate image ($(IMAGE))"
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

openshift:
	@echo ""
	@echo "[] openshift"
	sed < Dockerfile.redhat > Dockerfile.tmp \
		-e 's/^ARG version=.*/ARG version=$(VERSION)/' \
		-e 's/^ARG vcs_ref=.*/ARG vcs_ref=$(strip $(shell git rev-parse HEAD))/' \
		-e 's/^ARG build_date=.*/ARG build_date=$(strip $(shell date --utc -Iseconds))/'
	cmp Dockerfile.redhat Dockerfile.tmp && rm Dockerfile.tmp || mv Dockerfile.tmp Dockerfile.redhat
	docker build -f Dockerfile.redhat -t $(BIN) .
	docker inspect $(BIN):latest

# See https://connect.redhat.com/projects/${REDHAT_OSPID}/setup-preflight for these keys
REDHAT_OSPID := $(strip $(shell grep -v '\#' .redhat_ospid))
# find this registry key in the "Upload image manually" page of your Red Hat hosted container project.
REDHAT_REGISTRY_KEY := $(strip $(shell grep -v '\#' .redhat_registry_key))
# find the CTP API key
REDHAT_CTP_API_KEY := $(strip $(shell grep -v '\#' .redhat_ctp_api_key))
REDHAT_IMAGE_TAG := scan.connect.redhat.com/ospid-${REDHAT_OSPID}/$(BIN):${VERSION}
openshift-upload:
	docker login -u unused -p ${REDHAT_REGISTRY_KEY} scan.connect.redhat.com
	docker tag $(BIN):latest ${REDHAT_IMAGE_TAG}
	docker push ${REDHAT_IMAGE_TAG}

openshift-setup-preflight:
	test -d ~/openshift-preflight || (cd ~ && git clone https://github.com/redhat-openshift-ecosystem/openshift-preflight.git)
	cd ~/openshift-preflight
	$(VAGRANT) up
	$(VAGRANT) ssh -c "make -C preflight build"

openshift-preflight:
	cd ~/openshift-preflight; $(VAGRANT) ssh -c "\
		podman login -u unused -p ${REDHAT_REGISTRY_KEY} scan.connect.redhat.com; \
		cd preflight && \
		./preflight --loglevel debug -d \$${XDG_RUNTIME_DIR}/containers/auth.json \
			check container ${REDHAT_IMAGE_TAG} \
			--pyxis-api-token=${REDHAT_CTP_API_KEY} \
			--certification-project-id=${REDHAT_OSPID} \
			${PREFLIGHT_EXTRA_ARGS}"

openshift-submit:
	make openshift-preflight PREFLIGHT_EXTRA_ARGS=--submit

push:
	@echo ""
	@echo "[] push"
	docker push $(IMAGE)

clean:
	@echo ""
	@echo "[] clean"
	rm -vf $(BIN)-controller $(BIN)-node
