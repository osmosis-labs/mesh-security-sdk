#!/bin/bash
set -xeu

rm -rf $HOME/.meshd/chain2
home1=$HOME/.meshd/chain2/node1
home2=$HOME/.meshd/chain2/node2
chainid=chain-2

# init chain
meshd init mesh-2 --chain-id $chainid --home=$home1
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
cat $home1/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $home1/config/tmp_genesis.json && mv $home1/config/tmp_genesis.json $home1/config/genesis.json

# Allocate genesis accounts (cosmos formatted addresses)
val1=$(meshd keys show val1 --keyring-backend test --home=$home1 -a)
test1=$(meshd keys show test1 --keyring-backend test --home=$home1 -a)
val2=$(meshd keys show val2 --keyring-backend test --home=$home2 -a)
meshd add-genesis-account $val1 1000000000000stake --keyring-backend test --home=$home1
meshd add-genesis-account $val2 1000000000000stake --keyring-backend test --home=$home1
meshd add-genesis-account $test1 1000000000stake --keyring-backend test --home=$home1
cp $home1/config/genesis.json $home2/config/genesis.json

# Sign genesis transactions
meshd gentx val1 900000000000stake --keyring-backend test --chain-id $chainid --home $home1
meshd gentx val2 100000000000stake --keyring-backend test --chain-id $chainid --home $home2
cp $home2/config/gentx/*.json $home1/config/gentx/

# Collect genesis tx
meshd collect-gentxs --home $home1

# Run this to ensure everything worked and that the genesis file is setup correctly
meshd validate-genesis --home $home1
cp $home1/config/genesis.json $home2/config/genesis.json

# change app.toml values
VALIDATOR1_APP_TOML=$home1/config/app.toml
VALIDATOR2_APP_TOML=$home2/config/app.toml

sed -i -E 's|tcp://localhost:1317|tcp://localhost:1327|g' $VALIDATOR1_APP_TOML
sed -i -E 's|localhost:9090|localhost:9190|g' $VALIDATOR1_APP_TOML
sed -i -E 's|localhost:9091|localhost:9191|g' $VALIDATOR1_APP_TOML
sed -i -E 's|tcp://0.0.0.0:10337|tcp://0.0.0.0:11337|g' $VALIDATOR1_APP_TOML
sed -i -E 's|tcp://localhost:1317|tcp://localhost:1326|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9090|localhost:9188|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9091|localhost:9189|g' $VALIDATOR2_APP_TOML
sed -i -E 's|tcp://0.0.0.0:10337|tcp://0.0.0.0:11347|g' $VALIDATOR2_APP_TOML

# change config.toml values
VALIDATOR1_CONFIG=$home1/config/config.toml
VALIDATOR2_CONFIG=$home2/config/config.toml
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26668|g' $VALIDATOR1_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26667|g' $VALIDATOR1_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26666|g' $VALIDATOR1_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26665|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26664|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26663|g' $VALIDATOR2_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG

# peers
NODE1=$(meshd tendermint show-node-id --home=$home1)
NODE2=$(meshd tendermint show-node-id --home=$home2)
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$NODE1@localhost:26666,$NODE2@localhost:26666\"|g" $home1/config/config.toml
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$NODE1@localhost:26666,$NODE2@localhost:26666\"|g" $home2/config/config.toml

# start
screen -S mesh2-node1 -t mesh2-node1 -d -m meshd start --home=$home1
screen -S mesh2-node2 -t mesh2-node2 -d -m meshd start --home=$home2