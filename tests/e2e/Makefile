#!/usr/bin/make -f

BUILD_DIR = ./build

all: test

test:
	go test -mod=readonly -race ./...

build:
	CGO_ENABLED=1 go test -c -mod=readonly -o $(BUILD_DIR)/mesh-security.e2e.test

build-linux:
	CGO_ENABLED=1  GOOS=linux go test -c -mod=readonly -o $(BUILD_DIR)/mesh-security.e2e.test

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./x/vendor*" -not -path "./contracts*" -not -path "./packages*" -not -path "./docs*"| xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./x/vendor*" -not -path "./contracts*" -not -path "./packages*" -not -path "./docs*"| xargs gofumpt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./x/vendor*" -not -path "./contracts*" -not -path "./packages*" -not -path "./docs*"| xargs goimports -w -local github.com/osmosis-labs/mesh-security-sdk


.PHONY: all build build-linux test format
