# !/bin/bash

home1=$HOME/.meshd/chain1/
chainid1=chain-1
home2=$HOME/.meshd/chain2/
chainid2=chain-2

test1=$(meshd keys show test1  --keyring-backend test -a --home=$home1)
val1=$(meshd keys show val1  --keyring-backend test -a --home=$home1)
test2=$(meshd keys show test1  --keyring-backend test -a --home=$home2)
val2=$(meshd keys show val1  --keyring-backend test -a --home=$home2)

# # #=======bpatcontract consumer
meshd tx wasm store ./tests/testdata/mesh_simple_price_feed.wasm.gz --from $val1 --home=$HOME/.meshd/chain1  --chain-id chain-1 --keyring-backend test --fees 1stake -y --gas 10059023
sleep 7
meshd tx wasm store ./tests/testdata/mesh_virtual_staking.wasm.gz --from $val1 --home=$HOME/.meshd/chain1  --chain-id chain-1 --keyring-backend test --fees 1stake -y --gas 10059023
sleep 7
meshd tx wasm store ./tests/testdata/mesh_converter.wasm.gz --from $val1 --home=$HOME/.meshd/chain1  --chain-id chain-1 --keyring-backend test --fees 1stake -y --gas 10059023
sleep 7
meshd tx wasm instantiate 1 '{"native_per_foreign": "0.5"}' --label contract-pricefeed  --admin $val1 --from $val1 --home=$HOME/.meshd/chain1  --chain-id chain-1 --keyring-backend test --fees 1stake -y --gas 3059023 
sleep 7

price_feed=$(meshd q wasm list-contract-by-code 1 --output json | jq -r '.contracts[0]' )
echo "price feed contract: $price_feed"

init_converter=$(cat <<EOF
{
    "price_feed": "$price_feed", 
    "discount": "0.1", 
    "remote_denom": "stake",
    "virtual_staking_code_id": 2, 
    "max_retrieve":100
}
EOF
)
meshd tx wasm instantiate 3 "$init_converter" --label contract-converter  --admin $val1 --from $val1 --home=$HOME/.meshd/chain1  --chain-id chain-1 --keyring-backend test --fees 1stake -y --gas 3059023 
sleep 7

virtual_staking=$(meshd q wasm list-contract-by-code 2 --output json | jq -r '.contracts[0]' )
echo "virtual staking contract: $virtual_staking"

converter=$(meshd q wasm list-contract-by-code 3 --output json | jq -r '.contracts[0]' )
echo "converter contract: $converter"

# 1:mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf
# 2:mesh1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3syz4y6d
# 3:mesh1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3stmd2jl

# # #========== bpatcontract provider
meshd tx wasm store ./tests/testdata/mesh_vault.wasm.gz --from $val2 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 10059023
sleep 7
meshd tx wasm store ./tests/testdata/mesh_native_staking_proxy.wasm.gz --from $val2 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 10059023
sleep 7
meshd tx wasm store ./tests/testdata/mesh_native_staking.wasm.gz --from $val2 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 10059023
sleep 7
meshd tx wasm store ./tests/testdata/mesh_external_staking.wasm.gz --from $val2 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 15406929
sleep 7

msg=$(cat <<EOF
{"denom": "stake", "owner": "$val1", "proxy_code_id": 2,"slash_ratio_dsign": "0.20","slash_ratio_offline": "0.10"}
EOF
)
encode_msg=$(echo "$msg" | base64)
init_vault=$(cat <<EOF
{
    "denom": "stake", 
    "local_staking": {
        "code_id": 3, 
        "msg": "$encode_msg"
    }
}
EOF
)
meshd tx wasm instantiate 1 "$init_vault" --label contract-vault --admin $val2 --from $val2 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 3059023 

sleep 7
vault=$(meshd q wasm list-contract-by-code 1 --output json --node tcp://127.0.0.1:26654 | jq -r '.contracts[0]' )
echo "vault contract: $vault"

init_ext_staking=$(cat <<EOF
{
    "remote_contact": {
        "connection_id":"connection-0", 
        "port_id":"wasm.$converter"
    }, 
    "denom": "stake", 
    "vault": "$vault", 
    "unbonding_period": 1814400, 
    "rewards_denom": "stake", 
    "slash_ratio": { 
        "double_sign": "0.20", 
        "offline": "0.10" 
    }  
}
EOF
)
meshd tx wasm instantiate 4 "$init_ext_staking" --label contract-externalstaking --admin $val2 --from $val2 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 3059023 
sleep 7


ext_staking=$(meshd q wasm list-contract-by-code 4 --output json --node tcp://127.0.0.1:26654 | jq -r '.contracts[0]' )
echo "ext_staking contract: $ext_staking"

echo "Finish instantiate steps"
#1 mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf
#3 mesh1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3stmd2jl
#4 mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5

# 1:mesh1ltd0maxmte3xf4zshta9j5djrq9cl692ctsp9u5q0p9wss0f5lmstd9usy|mesh1dt3lk455ed360pna38fkhqn0p8y44qndsr77qu73ghyaz2zv4whq5m7u7c
# 3:mesh1up07dctjqud4fns75cnpejr4frmjtddzsmwgcktlyxd4zekhwecqmzzytn|mesh1k8re7jwz6rnnwrktnejdwkwnncte7ek7gt29gvnl3sdrg9mtnqksaq0y7h
