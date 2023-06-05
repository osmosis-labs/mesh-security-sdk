#!/bin/bash
set -o errexit -o nounset -o pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

echo "## Submit a mesh-security gov proposal for max cap: random contract"
RESP=$(meshd tx meshsecurity submit-proposal set-virtual-staking-max-cap \
  mesh1l94ptufswr6v7qntax4m7nvn3jgf6k4gn2rknq 100stake \
  --title "testing" --summary "Testing" --deposit "1000000000ustake" \
  --keyring-backend=test \
  --from validator --gas auto --gas-adjustment=1.5 -y  --chain-id=testing --node=http://localhost:26657 -b sync -o json)
echo $RESP
sleep 6
meshd q tx $(echo "$RESP"| jq -r '.txhash') -o json | jq

echo "## Query max cap for the contract"
meshd q meshsecurity max-cap-limit mesh1l94ptufswr6v7qntax4m7nvn3jgf6k4gn2rknq -o json | jq

echo "## Query max cap for each contract"
meshd q meshsecurity max-cap-limits -o json | jq

