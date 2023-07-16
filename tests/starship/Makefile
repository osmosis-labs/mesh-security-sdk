#!/usr/bin/make -f

BUILD_DIR = ./build

all: test

test:
	go test -mod=readonly -v -race -run Test2WayContract ./...

build-setup:
	cd setup && go build cmd/... -o build/ ./setup

run-setup:
	cd setup && go run ./...

build:
	CGO_ENABLED=1 go test -c -mod=readonly -o $(BUILD_DIR)/mesh-security.e2e.test

build-linux:
	CGO_ENABLED=1  GOOS=linux go test -c -mod=readonly -o $(BUILD_DIR)/mesh-security.e2e.test

.PHONY: all build build-linux test

###############################################################################
###                          Starship Helm Charts                           ###
###############################################################################
NAME = mesh-security
FILE = configs/starship.yaml

HELM_REPO = starship
HELM_CHART = devnet
HELM_VERSION = v0.1.39-rc5

.PHONY: check
setup-deps:
	bash $(CURDIR)/scripts/dev-setup.sh

setup-helm:
	helm repo add $(HELM_REPO) https://cosmology-tech.github.io/starship/
	helm repo update
	helm search repo $(HELM_REPO)/$(HELM_CHART) --version $(HELM_VERSION)

install:
	helm install -f $(FILE) $(NAME) $(HELM_REPO)/$(HELM_CHART) --version $(HELM_VERSION)

debug:
	helm install -f $(FILE) $(NAME) $(HELM_REPO)/$(HELM_CHART) --version $(HELM_VERSION) --dry-run --debug

delete:
	-helm delete $(NAME)

###############################################################################
###                                 Port forward                            ###
###############################################################################

.PHONY: port-forward
port-forward:
	bash $(CURDIR)/scripts/port-forward.sh --config=$(FILE)

.PHONY: stop-forward
stop-forward:
	-pkill -f "port-forward"

###############################################################################
###                          Local Kind Setup                               ###
###############################################################################
KIND_CLUSTER=starship

.PHONY: setup-kind
setup-kind:
	kind create cluster --name $(KIND_CLUSTER)

.PHONY: clean-kind
clean-kind:
	kind delete cluster --name $(KIND_CLUSTER)