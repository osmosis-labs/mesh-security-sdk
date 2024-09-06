#!/bin/bash
set -xeu

rm -rf $HOME/.meshd/chain1
home1=$HOME/.meshd/chain1/node1
home2=$HOME/.meshd/chain1/node2
chainid=chain-1

# init chain
meshd init mesh-1 --chain-id $chainid --home=$home1
meshd init mesh-2 --chain-id $chainid --home=$home2

# keys add
meshd keys add val1 --keyring-backend test --home=$home1
meshd keys add val2 --keyring-backend test --home=$home2
meshd keys add test1 --keyring-backend test --home=$home1

# Change parameter token denominations to stake
cat $home1/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="stake"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="30s"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["slashing"]["params"]["downtime_jail_duration"]="60s"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="10"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["meshsecurity"]["params"]["epoch_length"]=5' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json
cat $home1/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
val1=$(meshd keys show val1 --keyring-backend test --home=$home1 -a)
test1=$(meshd keys show test1 --keyring-backend test --home=$home1 -a)
val2=$(meshd keys show val2 --keyring-backend test --home=$home2 -a)
meshd add-genesis-account $val1 1000000000000stake --keyring-backend test --home=$home1
meshd add-genesis-account $val2 1000000000000stake --keyring-backend test --home=$home1
meshd add-genesis-account $test1 1000000000stake --keyring-backend test --home=$home1
cp $home1/config/genesis.json $home2/config/genesis.json

# Sign genesis transaction
meshd gentx val1 900000000000stake --keyring-backend test --chain-id $chainid --home=$home1
meshd gentx val2 100000000000stake --keyring-backend test --chain-id $chainid --home=$home2
cp $home2/config/gentx/*.json $home1/config/gentx/

# Collect genesis tx
meshd collect-gentxs --home=$home1
cp $home1/config/genesis.json $home2/config/genesis.json

# Run this to ensure everything worked and that the genesis file is setup correctly
meshd validate-genesis --home=$home1

# change app.toml values
VALIDATOR1_APP_TOML=$home1/config/app.toml
VALIDATOR2_APP_TOML=$home2/config/app.toml

sed -i -E 's|tcp://localhost:1317|tcp://localhost:1316|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9090|localhost:9088|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9091|localhost:9089|g' $VALIDATOR2_APP_TOML
sed -i -E 's|tcp://0.0.0.0:10337|tcp://0.0.0.0:10347|g' $VALIDATOR2_APP_TOML

# change config.toml values
VALIDATOR1_CONFIG=$home1/config/config.toml
VALIDATOR2_CONFIG=$home2/config/config.toml
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR1_CONFIG
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26630"|g' $VALIDATOR2_CONFIG

node1=$(meshd tendermint show-node-id --home=$home1)
node2=$(meshd tendermint show-node-id --home=$home2)
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$node1@localhost:26656,$node2@localhost:26656\"|g" $home1/config/config.toml
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$node1@localhost:26656,$node2@localhost:26656\"|g" $home2/config/config.toml

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
screen -S mesh1-node1 -t mesh1-node1 -d -m meshd start --home=$home1
screen -S mesh1-node2 -t mesh1-node2 -d -m meshd start --home=$home2