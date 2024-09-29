#!/bin/bash
rly config init --home ./scripts/relayer

cp ./scripts/mesh/testibc/config_osmosis_testnet.yaml ./scripts/relayer/config/config.yaml

rly keys add consumer key1 --home ./scripts/relayer

rly_wallet_1=$(rly keys show consumer key1 --home ./scripts/relayer)

rly keys add osmo testnet --home ./scripts/relayer
rly_wallet_2=$(rly keys show osmo testnet --home ./scripts/relayer)

meshconsumerd tx bank send test1 --keyring-backend test $rly_wallet_1 10000000stake --node http://localhost:26657 --fees 200000stake -y --home $HOME/.meshd/chain1/node1 --chain-id chain-1

osmo_balance=$(osmosisd q bank balances $rly_wallet_2 -o json --node https://rpc.testnet.osmosis.zone | jq -r '.balances[0].amount')

if [ "$(echo "$osmo_balance < 40000000" | bc)" -eq 1 ]; then
  echo "please request osmo token for ${rly_wallet_2} in https://faucet.testnet.osmosis.zone/"
  echo "waiting to fund accounts. Press to continue..."
  read -r answer
fi

sleep 3

rly paths new chain-1 osmo-test-5 demo-osmo --home ./scripts/relayer

sleep 5

rly tx clients demo-osmo --home ./scripts/relayer --override

sleep 10

rly tx connection demo-osmo --home ./scripts/relayer

sleep 10

home1=$HOME/.meshd/chain1/node1/
chainid1=chain-1
node1=tcp://127.0.0.1:26657
test1=$(meshconsumerd keys show test1  --keyring-backend test -a --home=$home1)
val1=$(meshconsumerd keys show val1  --keyring-backend test -a --home=$home1)
meshconsumerd tx wasm store ./tests/testdata/mesh_osmosis_price_feed.wasm.gz --node $node1 --from $val1 --home=$home1  --chain-id $chainid1 --keyring-backend test --fees 1stake -y --gas 10059023
sleep 7

connection_id=$(yq -r '.paths.demo-osmo.dst.connection-id' ./scripts/relayer/config/config.yaml)
echo "connection_id: $connection_id"
init_osmosis_price_feed=$(cat <<EOF
{
    "trading_pair": {
        "pool_id": 3,
        "base_asset": "ibc/8E2FEFCBD754FA3C97411F0126B9EC76191BAA1B3959CB73CECF396A4037BBF0",
        "quote_asset": "uosmo"
    },
    "epoch_in_secs": 30,
    "price_info_ttl_in_secs": 60
}
EOF
)
meshconsumerd tx wasm instantiate 1 "$init_osmosis_price_feed" --node $node1 --label contract-pricefeed  --admin $val1 --from $val1 --home=$home1  --chain-id $chainid1 --keyring-backend test --fees 1stake -y --gas 3059023 
sleep 7

price_feed=$(meshconsumerd q wasm list-contract-by-code 1 --node $node1 --output json | jq -r '.contracts[-1]' )
echo "price feed contract: $price_feed"


rly tx channel demo-osmo --src-port wasm.$price_feed --dst-port icqhost --order unordered --version icq-1 --home ./scripts/relayer --override

sleep 5

screen -S relayer-osmo -t relayer-osmo -d -m rly start demo-osmo --home ./scripts/relayer
sleep 5