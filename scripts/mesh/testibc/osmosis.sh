set -xeu

killall osmosisd || true
sleep 3
rm -rf $HOME/.osmosisd
home=$HOME/.osmosisd
chainid=osmo

osmosisd init localnet --chain-id $chainid --home $home

# Create accounts
osmosisd keys add val1 --keyring-backend test --home $home
osmosisd keys add test1 --keyring-backend test --home $home

cat $home/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

val1=$(osmosisd keys show val1 --keyring-backend test --home=$home -a)
test1=$(osmosisd keys show test1 --keyring-backend test --home=$home -a)
osmosisd add-genesis-account $val1 1000000000000uosmo,10000000000000000000stake --keyring-backend test --home=$home
osmosisd add-genesis-account $test1 1000000000uosmo --keyring-backend test --home=$home

cat $home/config/genesis.json | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["gov"]["params"]["voting_period"]="40s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["gov"]["params"]["expedited_voting_period"]="30s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["interchainquery"]["params"]["allow_queries"]=["/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow"]' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update staking genesis
cat $home/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="240s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update crisis variable to uosmo
cat $home/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json


# update epochs genesis
cat $home/config/genesis.json | jq '.app_state["epochs"]["epochs"][1]["duration"]="60s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update poolincentives genesis
cat $home/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update incentives genesis
cat $home/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update mint genesis
cat $home/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json
cat $home/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="day"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update gamm genesis
cat $home/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

# update cl genesis
cat $home/config/genesis.json | jq '.app_state["concentratedliquidity"]["params"]["is_permissionless_pool_creation_enabled"]=true' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

cat $home/config/genesis.json | jq '.app_state["txfees"]["basedenom"] = "uosmo"' > $home/config/tmp_genesis.json && mv $home/config/tmp_genesis.json $home/config/genesis.json

osmosisd gentx val1 500000000000uosmo --keyring-backend test --chain-id $chainid --home=$home

osmosisd collect-gentxs --home=$home

VALIDATOR_APP_TOML=$home/config/app.toml

sed -i -E 's|tcp://localhost:1317|tcp://localhost:1337|g' $VALIDATOR_APP_TOML
sed -i -E 's|localhost:9090|localhost:9290|g' $VALIDATOR_APP_TOML
sed -i -E 's|localhost:9091|localhost:9291|g' $VALIDATOR_APP_TOML
sed -i -E 's|tcp://0.0.0.0:10337|tcp://0.0.0.0:12337|g' $VALIDATOR_APP_TOML

VALIDATOR_CONFIG=$home/config/config.toml
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26678|g' $VALIDATOR_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26677|g' $VALIDATOR_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26676|g' $VALIDATOR_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR_CONFIG

screen -S osmosis -t osmosis -d -m osmosisd start --home=$home

sleep 10
osmosisd tx gamm create-pool --pool-file ./scripts/mesh/testibc/pool.json --from val1 --keyring-backend test --node http://localhost:26677 --chain-id osmo --fees 10000uosmo -y --gas auto --gas-adjustment 1.5
sleep 7
osmosisd q gamm num-pools --node http://localhost:26677
