#!/bin/bash
rly config init --home ./scripts/relayer

cp ./scripts/mesh/testibc/config_band.yaml ./scripts/relayer/config/config.yaml

rly keys add consumer key1 --home ./scripts/relayer

rly_wallet_1=$(rly keys show consumer key1 --home ./scripts/relayer)

# rly keys add band testnet --home ./scripts/relayer
rly_wallet_2=$(rly keys show band testnet --home ./scripts/relayer)

meshconsumerd tx bank send test1 --keyring-backend test $rly_wallet_1 10000000stake --node http://localhost:26657 --fees 200000stake -y --home $HOME/.meshd/chain1/node1 --chain-id chain-1

node ./scripts/mesh/testibc/faucet.js $rly_wallet_2

sleep 3

rly paths new chain-1 band-laozi-testnet6 demo-band --home ./scripts/relayer

sleep 5

rly tx clients demo-band --home ./scripts/relayer --override

sleep 10

rly tx connection demo-band --home ./scripts/relayer

sleep 10

home1=$HOME/.meshd/chain1/node1/
chainid1=chain-1
node1=tcp://127.0.0.1:26657
test1=$(meshconsumerd keys show test1  --keyring-backend test -a --home=$home1)
val1=$(meshconsumerd keys show val1  --keyring-backend test -a --home=$home1)
meshconsumerd tx wasm store ./tests/testdata/mesh_band_price_feed.wasm.gz --node $node1 --from $val1 --home=$home1  --chain-id $chainid1 --keyring-backend test --fees 1stake -y --gas 10059023
sleep 7

client_id=$(yq e '.paths.demo-band.src.client-id' ./scripts/relayer/config/config.yaml)
oracle_script_id=445

init_band_price_feed=$(cat <<EOF
{
    "trading_pair": {
        "base_asset": "OSMO",
        "quote_asset": "ATOM"
    },
    "client_id": "$client_id",
    "oracle_script_id": "$oracle_script_id",
    "ask_count": "1",
    "min_count": "1",
    "fee_limit": [{"denom": "uband", "amount":"10000"}],
    "prepare_gas": "40000",
    "execute_gas": "300000",
    "minimum_sources": 2,
    "epoch_in_secs": 30,
    "price_info_ttl_in_secs": 60
}
EOF
)
meshconsumerd tx wasm instantiate 1 "$init_band_price_feed" --node $node1 --label contract-pricefeed  --admin $val1 --from $val1 --home=$home1  --chain-id $chainid1 --keyring-backend test --fees 1stake -y --gas 3059023 
sleep 7

price_feed=$(meshconsumerd q wasm list-contract-by-code 1 --node $node1 --output json | jq -r '.contracts[0]' )
echo "price feed contract: $price_feed"


rly tx channel demo-band --src-port wasm.$price_feed --dst-port oracle --order unordered --version bandchain-1 --home ./scripts/relayer --override

sleep 7

screen -S relayer-band -t relayer-band -d -m  rly start demo-band --home ./scripts/relayer
sleep 5