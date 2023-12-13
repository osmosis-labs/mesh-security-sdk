#!/usr/bin/make -f

BUILD_DIR = ./build

all: test

test:
	go test -count=1 -mod=readonly -v -race ./...

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
FILE = configs/local.yaml

HELM_REPO = starship
HELM_CHART = devnet
HELM_VERSION = v0.1.46-rc1

.PHONY: check
setup-deps:
	bash $(CURDIR)/scripts/dev-setup.sh

install:
	bash $(CURDIR)/scripts/install.sh --config $(FILE) --version $(HELM_VERSION) --name $(NAME)

debug:
	bash $(CURDIR)/scripts/install.sh --config $(FILE) --version $(HELM_VERSION) --name $(NAME) --dry-run

delete:
	-helm delete $(NAME)

stop: stop-forward delete

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
