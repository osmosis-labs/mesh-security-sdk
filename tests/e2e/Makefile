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
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./tests/system/vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gci write --skip-generated -s standard -s default -s "prefix(cosmossdk.io)" -s "prefix(github.com/cosmos/cosmos-sdk)" -s "prefix(github.com/osmosis-labs/mesh-security-sdk)" --custom-order


.PHONY: all build build-linux test format
