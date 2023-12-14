NODE_DIR := ~/.meshd
CODE_DIR := $(CURDIR)/../..

.PHONY: setup
setup: cleanup
	bash setup_node.sh

.PHONY: run-docker
start-chain:
	cd $(CODE_DIR) && docker run \
		--platform linux/amd64 -it --name meshd --rm --entrypoint='/bin/bash' \
		-v .:/opt \
		-v $(NODE_DIR):/root/.meshd \
		-p 26657:26657 \
		ghcr.io/osmosis-labs/meshd:latest

.PHONY: cleanup
cleanup:
	rm -rf $(NODE_DIR)
