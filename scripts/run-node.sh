#!/bin/bash
rm -rf $HOME/.meshd/


meshd keys add val --keyring-backend test 
meshd keys add test1 --keyring-backend test 
meshd keys add test2 --keyring-backend test 
meshd keys add test3 --keyring-backend test 

# init chain
meshd init test-1 --chain-id testt

# Change parameter token denominations to stake
cat $HOME/.meshd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > $HOME/.meshd/config/tmp_genesis.json && mv $HOME/.meshd/config/tmp_genesis.json $HOME/.meshd/config/genesis.json
cat $HOME/.meshd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > $HOME/.meshd/config/tmp_genesis.json && mv $HOME/.meshd/config/tmp_genesis.json $HOME/.meshd/config/genesis.json
cat $HOME/.meshd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="stake"' > $HOME/.meshd/config/tmp_genesis.json && mv $HOME/.meshd/config/tmp_genesis.json $HOME/.meshd/config/genesis.json
cat $HOME/.meshd/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $HOME/.meshd/config/tmp_genesis.json && mv $HOME/.meshd/config/tmp_genesis.json $HOME/.meshd/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
meshd add-genesis-account val 1000000000000stake --keyring-backend test
meshd add-genesis-account test1 1000000000stake --keyring-backend test
meshd add-genesis-account test2 1000000000stake --keyring-backend test
meshd add-genesis-account test3 50000000stake --keyring-backend test

# Sign genesis transaction
meshd gentx val 1000000stake --keyring-backend test --chain-id testt

# Collect genesis tx
meshd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
meshd validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
meshd start 