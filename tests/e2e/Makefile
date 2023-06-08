#!/usr/bin/make -f

BUILD_DIR = ./build

all: test

test:
	go test -mod=readonly -race ./...

build:
	CGO_ENABLED=0 go test -c -mod=readonly -o $(BUILD_DIR)/mesh-security.e2e.test

build-linux:
	CGO_ENABLED=0  GOOS=linux go test -c -mod=readonly -o $(BUILD_DIR)/mesh-security.e2e.test

.PHONY: all build build-linux test
