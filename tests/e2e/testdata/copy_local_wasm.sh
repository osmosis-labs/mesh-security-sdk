#!/bin/bash
set -o errexit -o nounset -o pipefail
command -v shellcheck > /dev/null && shellcheck "$0"

echo "DEV-only: copy from local built instead of downloading"

for contract in external_staking mesh_converter mesh_native_staking mesh_native_staking_proxy mesh_simple_price_feed \
mesh_vault mesh_virtual_staking ; do
cp -f  ../../../../mesh-security/artifacts/${contract}.wasm .
gzip -fk ${contract}.wasm
rm -f ${contract}.wasm
done

cd ../../../../mesh-security
tag=$(git rev-parse HEAD)
cd -
rm -f version.txt
echo "$tag" >version.txt