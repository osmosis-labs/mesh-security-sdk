#!/bin/bash
set -o errexit -o nounset -o pipefail
command -v shellcheck > /dev/null && shellcheck "$0"

if [ $# -ne 1 ]; then
  echo "Usage: ./download_releases.sh RELEASE_TAG"
  exit 1
fi
tag="$1"

for contract in mesh_external_staking mesh_converter mesh_native_staking mesh_native_staking_proxy mesh_osmosis_price_provider \
mesh_remote_price_feed mesh_simple_price_feed mesh_vault mesh_virtual_staking; do
  url="https://github.com/osmosis-labs/mesh-security/releases/download/$tag/${contract}.wasm"
  echo "Downloading $url ..."
  wget -O "${contract}.wasm" "$url"
  gzip -fk ${contract}.wasm
  rm -f ${contract}.wasm
done

rm -f version.txt
echo "$tag" >version.txt
