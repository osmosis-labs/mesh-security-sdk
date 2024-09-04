#!/bin/bash

cp ./scripts/mesh/testibc/config.yaml $HOME/.relayer/config/

rly keys add consumer key1
rly_wallet_1=$(rly keys show consumer key1)

rly keys add provider key2
rly_wallet_2=$(rly keys show provider key2)

meshd tx bank send test1 --keyring-backend test $rly_wallet_1 10000000stake --node http://localhost:26657 --fees 200000stake -y --home $HOME/.meshd/chain1 --chain-id chain-1
meshd tx bank send test1 --keyring-backend test $rly_wallet_2 10000000stake --node http://localhost:26654 --fees 200000stake -y --home $HOME/.meshd/chain2 --chain-id chain-2

sleep 3

rly paths new chain-1 chain-2 demo

sleep 5

rly tx clients demo

sleep 10

rly tx connection demo

sleep 10

converter=$(meshd q wasm list-contract-by-code 3 --output json | jq -r '.contracts[0]' )
ext_staking=$(meshd q wasm list-contract-by-code 4 --output json --node tcp://127.0.0.1:26654 | jq -r '.contracts[0]' )

rly tx channel demo --src-port wasm.$converter --dst-port wasm.$ext_staking --order unordered --version '{"protocol":"mesh-security","version":"0.11.0"}'

sleep 15

screen -S rly -t rly -d -m  rly start demo

sleep 10