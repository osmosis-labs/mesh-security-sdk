#!/bin/bash
rly config init --home ./scripts/relayer

cp ./scripts/mesh/testibc/config_osmosis_local.yaml ./scripts/relayer/config/config.yaml

rly keys add consumer key1 --home ./scripts/relayer

rly_wallet_1=$(rly keys show consumer key1 --home ./scripts/relayer)

rly keys add osmo testnet --home ./scripts/relayer
rly_wallet_2=$(rly keys show osmo testnet --home ./scripts/relayer)

meshconsumerd tx bank send test1 --keyring-backend test $rly_wallet_1 10000000stake --node http://localhost:26657 --fees 200000stake -y --home $HOME/.meshd/chain1/node1 --chain-id chain-1
osmosisd tx bank send test1 --keyring-backend test $rly_wallet_2 10000000uosmo --node http://localhost:26677 --fees 100000uosmo -y --home $HOME/.osmosisd --chain-id osmo

sleep 7

rly paths new chain-1 osmo demo-osmo --home ./scripts/relayer

sleep 5

rly tx clients demo-osmo --home ./scripts/relayer

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
        "pool_id": 1,
        "base_asset": "stake",
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


rly tx channel demo-osmo --src-port wasm.$price_feed --dst-port icqhost --order unordered --version icq-1 --home ./scripts/relayer

sleep 5

screen -S relayer-osmo -t relayer-osmo -d -m rly start demo-osmo --home ./scripts/relayer
sleep 5