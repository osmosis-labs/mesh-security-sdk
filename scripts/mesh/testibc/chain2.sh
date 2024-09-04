#!/bin/bash
set -xeu

rm -rf $HOME/.meshd/chain2
home=$HOME/.meshd/chain2
chainid=chain-2


# meshd keys add val --keyring-backend test --home=$home
# meshd keys add val --keyring-backend test --home=$home
mnm_val=$(cat ./scripts/mesh/testibc/mnemonic2)

echo "$mnm_val"| meshd keys add val1 --keyring-backend test --home=$home --recover

meshd keys add test1 --keyring-backend test --home=$home

# init chain
meshd init mesh-2 --chain-id $chainid --home=$home

# Change parameter token denominations to stake
cat $HOME/.meshd/chain2/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > $HOME/.meshd/chain2/config/tmp_genesis.json && mv $HOME/.meshd/chain2/config/tmp_genesis.json $HOME/.meshd/chain2/config/genesis.json
cat $HOME/.meshd/chain2/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > $HOME/.meshd/chain2/config/tmp_genesis.json && mv $HOME/.meshd/chain2/config/tmp_genesis.json $HOME/.meshd/chain2/config/genesis.json
cat $HOME/.meshd/chain2/config/genesis.json | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="stake"' > $HOME/.meshd/chain2/config/tmp_genesis.json && mv $HOME/.meshd/chain2/config/tmp_genesis.json $HOME/.meshd/chain2/config/genesis.json
cat $HOME/.meshd/chain2/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="30s"' > $HOME/.meshd/chain2/config/tmp_genesis.json && mv $HOME/.meshd/chain2/config/tmp_genesis.json $HOME/.meshd/chain2/config/genesis.json
cat $HOME/.meshd/chain2/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $HOME/.meshd/chain2/config/tmp_genesis.json && mv $HOME/.meshd/chain2/config/tmp_genesis.json $HOME/.meshd/chain2/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
meshd add-genesis-account val1 1000000000000stake --keyring-backend test --home=$home 
meshd add-genesis-account test1 1000000000stake --keyring-backend test --home=$home

# Sign genesis transaction
meshd gentx val1 1000000stake --keyring-backend test --chain-id $chainid --home=$home

# Collect genesis tx
meshd collect-gentxs --home=$home

# Run this to ensure everything worked and that the genesis file is setup correctly
meshd validate-genesis --home=$home

# validator2
VALIDATOR2_CONFIG=$HOME/.meshd/chain2/config/config.toml
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26630"|g' $VALIDATOR2_CONFIG

VALIDATOR2_APP_TOML=$HOME/.meshd/chain2/config/app.toml
sed -i -E 's|tcp://localhost:1317|tcp://localhost:1316|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9090|localhost:9088|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9091|localhost:9089|g' $VALIDATOR2_APP_TOML
sed -i -E 's|tcp://0.0.0.0:10337|tcp://0.0.0.0:10347|g' $VALIDATOR2_APP_TOML

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
screen -S mesh2 -t mesh2 -d -m meshd start --home=$home

# sleep 7

# test2=$(meshd keys show test1  --keyring-backend test -a --home=$home)
# val2=$(meshd keys show val  --keyring-backend test -a --home=$home)

# meshd tx bank send $val2 $test2 100000stake --home=$home --chain-id $chainid --keyring-backend test --fees 10stake -y --node tcp://127.0.0.1:26654