FROM golang:1.21-bookworm AS go-builder

RUN apt-get update && \
    apt-get install -y ca-certificates wget make git file libc6-dev clang gcc build-essential

WORKDIR /code

# Download dependencies and CosmWasm libwasmvm if found.
ADD demo/go.mod demo/go.sum ./

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
ENV CGO_LDFLAGS -Wl,--no-as-needed
RUN cd demo/ && LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/demo/build/meshd \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/demo/build/meshd | grep "statically linked")

# --------------------------------------------------------
FROM debian:bookworm

COPY --from=go-builder /code/demo/build/meshd /usr/bin/meshd

# Install dependencies used for Starship
RUN apk add --no-cache curl make bash jq sed

WORKDIR /opt

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657

CMD ["/usr/bin/meshd", "version"]
