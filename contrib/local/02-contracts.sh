#!/bin/bash
set -o errexit -o nounset -o pipefail -x

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
CHAIN_ID="testing"

echo "-----------------------"
echo "# Setup Consumer contracts"
echo "## Add mesh_simple_price_feed contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_simple_price_feed.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

echo "## Add mesh_virtual_staking contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_virtual_staking.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

echo "## Add mesh_converter contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_converter.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

echo "-----------------------"
echo "# Setup Provider contracts"
echo "## Add mesh_vault contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_vault.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

echo "## Add mesh_native_staking_proxy contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_native_staking_proxy.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

echo "## Add mesh_native_staking contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_native_staking.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

echo "## Add mesh_external_staking contract"
RESP=$(meshd tx wasm store "$DIR/../../tests/testdata/mesh_external_staking.wasm.gz" \
  --from validator --gas 15000000 -y --chain-id=$CHAIN_ID --node=http://localhost:26657 -b sync -o json --keyring-backend=test)
sleep 6
RESP=$(meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json)
CODE_ID=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_id").value')
CODE_HASH=$(echo "$RESP" | jq -r '.events[]| select(.type=="store_code").attributes[]| select(.key=="code_checksum").value')
echo "* Code id: $CODE_ID"
echo "* Code checksum: $CODE_HASH"

