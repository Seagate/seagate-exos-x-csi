ifndef DOCKER_HUB_REPOSITORY
	DOCKER_HUB_REPOSITORY = ghcr.io/seagate
endif

ifndef VERSION
	VERSION = latest
endif

ifndef BIN
	BIN = seagate-exos-x-testapp
endif

IMAGE = $(DOCKER_HUB_REPOSITORY)/$(BIN):$(VERSION)

help:
	@echo ""
	@echo "Build Targets:"
	@echo "-----------------------------------------------------------------------------------"
	@echo "make clean      - remove $(BIN)"
	@echo "make image      - create a repo docker image ($(IMAGE))"
	@echo "make push       - push the docker image to '$(DOCKER_HUB_REPOSITORY)'"
	@echo "make all        - clean, image, push"
	@echo ""

all:		clean image push
.PHONY: all

image:
	docker build -f Dockerfile.testapp -t $(IMAGE) --build-arg version="$(VERSION)" .
.PHONY: image

push:
	docker push $(IMAGE)
.PHONY: push

clean:
	rm -vf $(BIN)
.PHONY: clean
