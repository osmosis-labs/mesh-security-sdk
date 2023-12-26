FROM golang:1.21-bullseye AS go-builder

RUN apt-get update && apt-get install -y ca-certificates wget git clang gcc g++ zlib1g-dev libmpc-dev libmpfr-dev libgmp-dev build-essential cmake make bash jq sed

WORKDIR /code

# Download dependencies and CosmWasm libwasmvm if found.
ADD go.mod go.sum ./

#ADD https://github.com/CosmWasm/wasmvm/releases/download/v$wasmvm/libwasmvm_muslc.$arch.a /lib/libwasmvm_muslc.$arch.a
## Download
RUN set -eux; \
    WASM_VERSION=v$(go list -m github.com/CosmWasm/wasmvm | cut -d" " -f2 | cut -d"v" -f2); \
    echo $WASM_VERSION; \
    wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.$(uname -m).a

# Copy over code
COPY . /code

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
# then log output of file /code/bin/meshd
# then ensure static linking
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build-vendored
  && file /code/build/meshd
