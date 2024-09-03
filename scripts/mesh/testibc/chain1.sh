#!/bin/bash

rm -rf $HOME/.meshd/chain1
home=$HOME/.meshd/chain1/
chainid=chain-1


# meshd keys add val --keyring-backend test --home=$home
mnm_val=$(cat ./scripts/mesh/testibc/mnemonic1)

echo "$mnm_val"| meshd keys add val1 --keyring-backend test --home=$home --recover

meshd keys add test1 --keyring-backend test --home=$home

# init chain
meshd init mesh-1 --chain-id chain-1 --home=$home

# Change parameter token denominations to stake
cat $HOME/.meshd/chain1/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > $HOME/.meshd/chain1/config/tmp_genesis.json && mv $HOME/.meshd/chain1/config/tmp_genesis.json $HOME/.meshd/chain1/config/genesis.json
cat $HOME/.meshd/chain1/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > $HOME/.meshd/chain1/config/tmp_genesis.json && mv $HOME/.meshd/chain1/config/tmp_genesis.json $HOME/.meshd/chain1/config/genesis.json
cat $HOME/.meshd/chain1/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="stake"' > $HOME/.meshd/chain1/config/tmp_genesis.json && mv $HOME/.meshd/chain1/config/tmp_genesis.json $HOME/.meshd/chain1/config/genesis.json
cat $HOME/.meshd/chain1/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="30s"' > $HOME/.meshd/chain1/config/tmp_genesis.json && mv $HOME/.meshd/chain1/config/tmp_genesis.json $HOME/.meshd/chain1/config/genesis.json
cat $HOME/.meshd/chain1/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $HOME/.meshd/chain1/config/tmp_genesis.json && mv $HOME/.meshd/chain1/config/tmp_genesis.json $HOME/.meshd/chain1/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
meshd add-genesis-account val1 1000000000000stake --keyring-backend test --home=$home 
meshd add-genesis-account test1 1000000000stake --keyring-backend test --home=$home

# Sign genesis transaction
meshd gentx val1 1000000stake --keyring-backend test --chain-id $chainid --home=$home

# Collect genesis tx
meshd collect-gentxs --home=$home

# Run this to ensure everything worked and that the genesis file is setup correctly
meshd validate-genesis --home=$home

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
screen -S mesh1 -t mesh1 -d -m meshd start --home=$home

# mesh14p7ermueemlmxmegzrq2ku4tukz3wzgremaywj
# night room resemble sting basic damp senior primary person slab trick picnic embark work hill base combine double ride relief journey marble salute math