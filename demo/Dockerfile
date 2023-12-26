FROM golang:1.21-bookworm AS go-builder

RUN apt-get update && apt-get install -y ca-certificates wget make gcc build-essential git

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
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build-vendored \
  && file /code/build/meshd \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/build/meshd | grep "statically linked")

# --------------------------------------------------------
FROM debian:bookworm

COPY --from=go-builder /code/build/meshd /usr/bin/meshd

# Install dependencies used for Starship
RUN apk add --no-cache curl make bash jq sed

WORKDIR /opt

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657

CMD ["/usr/bin/meshd", "version"]
