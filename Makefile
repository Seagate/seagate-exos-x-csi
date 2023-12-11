.PHONY: help all bin controller node test openshift push clean

VENDOR := seagate
GITHUB_ORG := Seagate
# Project name, without vendor
NAME := exos-x-csi
# Project name, including vendor
PROJECT := $(VENDOR)-$(NAME)
GITHUB_URL := github.com/$(GITHUB_ORG)/$(PROJECT)
NAMESPACE := $(VENDOR)

ifdef DOCKER_HUB_REPOSITORY
DOCKER_HUB_REPOSITORY := $(DOCKER_HUB_REPOSITORY)
else
DOCKER_HUB_REPOSITORY := ghcr.io/seagate
endif

# Note: the version number takes the form "v1.2.3" when used as a repository tag, but
# appears as "1.2.3" in other contexts such as the Helm chart.
ifdef VERSION
VERSION := $(VERSION)
else
VERSION := v1.8.3
endif
HELM_VERSION := $(subst v,,$(VERSION))
VERSION_FLAG = -X $(GITHUB_URL)/pkg/common.Version=$(VERSION)

ifndef BIN
	BIN = $(PROJECT)
endif

# $HELM_KEY must be the name of a secret key in the invoker's default keyring if package is to be signed
HELM_KEY := css-host-software
ifneq (,$(HELM_KEY))
  HELM_KEYRING := ~/.gnupg/secring.gpg
  HELM_SIGN := --sign --key $(HELM_KEY) --keyring $(HELM_KEYRING)
endif 
HELM_PACKAGE := $(BIN)-$(HELM_VERSION).tgz
HELM_IMAGE_REPO := $(DOCKER_HUB_REPOSITORY)/$(BIN)
IMAGE = $(DOCKER_HUB_REPOSITORY)/$(BIN):$(VERSION)

help:
	@echo ""
	@echo "Build Targets:"
	@echo "-----------------------------------------------------------------------------------"
	@echo "make all          - clean, build openshift docker image, push to registry"
	@echo "make bin          - create controller and node driver binaries"
	@echo "make clean        - remove '$(BIN)-controller' and '$(BIN)-node'"
	@echo "make controller   - create controller driver image ($(BIN)-controller)"
	@echo "make helm-package - create signed helm package using HELM_VERSION, HELM_KEY environment variables"
	@echo "make node         - create node driver image ($(BIN)-node)"
	@echo "make openshift    - Create OpenShift-certification candidate image ($(IMAGE))"
	@echo "make push         - push the docker image to '$(DOCKER_HUB_REPOSITORY)'"
	@echo "make test         - build test/sanity"
	@echo ""

all: clean openshift push

bin: controller node


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

push:
	@echo ""
	@echo "[] push"
	docker push $(IMAGE)

pull:
	@echo ""
	@echo "[] pull"
	docker pull $(IMAGE)

clean:
	@echo ""
	@echo "[] clean"
	rm -vf $(BIN)-controller $(BIN)-node *.zip *.tgz *.prov helm/$(BIN)-$(HELM_VERSION)*

######################## Openshift certification stuff ########################

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

# Makefile.secrets should include the following lines:
#
#   PYXIS_API_TOKEN=<pyxis-api-token>
#   REGISTRY_KEY=<redhat-registry-key> (see )
#
# For more info, see the CSI Driver "Re-certifying" OneNote page, or
#   https://connect.redhat.com/account/api-keys?extIdCarryOver=true&sc_cid=701f2000001OH7JAAW
#   https://connect.redhat.com/projects/610494ea40182fa9651cdab0/setup-preflight
#
# Make sure this file does not get checked in to git!  Note that for automation purposes you can
# just pass these variables in the environment instead of a file.
-include Makefile.secrets

PREFLIGHT=../openshift-preflight/preflight
PREFLIGHT_REGISTRY=localhost:5000
PREFLIGHT_IMAGE=$(PREFLIGHT_REGISTRY)/$(BIN):$(VERSION)
REDHAT_PROJECT_ID=610494ea40182fa9651cdab0
REDHAT_IMAGE_BASE=quay.io/redhat-isv-containers/$(REDHAT_PROJECT_ID)
REDHAT_IMAGE=$(REDHAT_IMAGE_BASE):$(VERSION)
REDHAT_IMAGE_LATEST=$(REDHAT_IMAGE_BASE):latest
PREFLIGHT_OPTIONS=
PREFLIGHT_SUBMIT=
PFLT_DOCKERCONFIG=.preflight-auth.json

preflight: $(PFLT_DOCKERCONFIG)
	-docker run -d -p 5000:5000 --name registry registry:2 # make sure local registry is running
	docker tag $(IMAGE) $(PREFLIGHT_IMAGE)
	docker push $(PREFLIGHT_IMAGE)
	PFLT_DOCKERCONFIG=$(PFLT_DOCKERCONFIG) $(PREFLIGHT) check container $(PREFLIGHT_SUBMIT) $(PREFLIGHT_OPTIONS) $(PREFLIGHT_IMAGE)
		
preflight-submit: .preflight-auth.json
	$(MAKE) preflight PREFLIGHT_SUBMIT="--submit" \
		PREFLIGHT_OPTIONS="--pyxis-api-token=$(PYXIS_API_TOKEN) --certification-project-id=$(REDHAT_PROJECT_ID)" \
		PREFLIGHT_IMAGE=$(REDHAT_IMAGE) PFLT_DOCKERCONFIG=$(PFLT_DOCKERCONFIG)

tag-latest:
	podman tag $(REDHAT_IMAGE) $(REDHAT_IMAGE_LATEST)
	podman push $(REDHAT_IMAGE_LATEST)

preflight-push:
	podman login -u redhat-isv-containers+$(REDHAT_PROJECT_ID)-robot -p $(REGISTRY_KEY) quay.io
	podman tag $(IMAGE) $(REDHAT_IMAGE)
	podman push $(REDHAT_IMAGE)

.preflight-auth.json:
	podman login -u redhat-isv-containers+610494ea40182fa9651cdab0-robot -p $(REGISTRY_KEY) --authfile "$@" quay.io

build-preflight:
	(cd ..; git clone https://github.com/redhat-openshift-ecosystem/openshift-preflight.git)
	cd ../openshift-preflight && make build

######################## Helm package creation ########################


# Create a helm package that can be installed from a remote HTTPS URL with, e.g.
# helm install exos-x-csi https://<server>/<path>/seagate-exos-x-csi-1.0.0.tgz
helm-package: $(HELM_PACKAGE)

# Update version numbers in the Helm chart.  If yq is not installed, try "go install github.com/mikefarah/yq/v4@latest"
update-chart:	$(MAKEFILE)
	yq -i '.image.tag="$(VERSION)" | .image.repository="$(HELM_IMAGE_REPO)"' helm/csi-charts/values.yaml

# Make a helm package. If yq is installed, the chart will be updated to reflect version $(VERSION)
# To create a package without signing it, specify "make helm-package HELM_KEY="
# Note that helm doesn't support GPG v2.1 kbx files; if signing fails, try:
# gpg --export-secret-keys > ~/.gnupg/secring.gpg
$(HELM_PACKAGE):
	echo HELM_PACKAGE:=$@
	( which yq >/dev/null && $(MAKE) update-chart ) || true
	cd helm; helm package --app-version "$(HELM_VERSION)" --version "$(HELM_VERSION)" $(HELM_SIGN) $$PWD/csi-charts
	cp -p helm/$@* .

# Verify a signed package create a zip file containing the package and its provenance file
signed-helm-package: $(HELM_PACKAGE)
	helm verify --keyring $(HELM_KEYRING) $<
	zip -r $(subst .tgz,-signed-helm-package.zip,$<) $< $<.prov

# This will allow the package to be installed directly from Github, with the command:
# helm install -n $(NAMESPACE) exos-x-csi https://$(GITHUB_URL)/releases/download/$(VERSION)/$(PROJECT)-$(HELM_VERSION).tgz
helm-upload: $(HELM_PACKAGE)
	gh release upload $(VERSION) '$^#Helm Package' -R $(GITHUB_ORG)/$(PROJECT)
	@echo Install package with:
	@echo ' ' helm install -n $(NAMESPACE) $(NAME) https://$(GITHUB_URL)/releases/download/$(VERSION)/$(PROJECT)-$(HELM_VERSION).tgz

