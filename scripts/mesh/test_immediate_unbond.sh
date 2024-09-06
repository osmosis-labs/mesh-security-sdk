# !/bin/bash
killall meshd || true
killall rly || true
# deploy chain 1
./scripts/mesh/testibc/chain1.sh
# deploy chain 2
./scripts/mesh/testibc/chain2.sh
sleep 7
./scripts/mesh/testibc/instantiate.sh
# run relayer
./scripts/mesh/testibc/rly.sh

home1node1=$HOME/.meshd/chain1/node1
home2node1=$HOME/.meshd/chain2/node1
node2=tcp://127.0.0.1:26667

virtual_staking=$(meshd q wasm list-contract-by-code 2 --output json | jq -r '.contracts[0]' )
converter=$(meshd q wasm list-contract-by-code 3 --output json | jq -r '.contracts[0]' )
vault=$(meshd q wasm list-contract-by-code 1 --output json --node $node2 | jq -r '.contracts[0]' )
native_staking=$(meshd q wasm list-contract-by-code 3 --output json --node $node2 | jq -r '.contracts[0]' )
ext_staking=$(meshd q wasm list-contract-by-code 4 --output json --node $node2 | jq -r '.contracts[0]' )
test1_provider_addr=$(meshd keys show test1 --keyring-backend test --home=$home2node1 --address)

# Set virtual staking max cap
meshd tx meshsecurity submit-proposal set-virtual-staking-max-cap $virtual_staking 100000000stake --title "a title" --summary "a summary" --from test1 --keyring-backend test --home=$home1node1 --chain-id chain-1 -y --deposit 10000000stake

sleep 7

meshd tx gov vote 1 yes --from val1 --keyring-backend test --home=$home1node1 --chain-id chain-1 -y

sleep 5

# Update mesh security provider module's params
gov_addr=$(meshd q auth --node $node2  module-account gov -o json | jq ".account.base_account.address")

echo "gov addr: $gov_addr"
proposal=$(cat <<EOF
{
 "messages": [
  {
   "@type": "/osmosis.meshsecurityprovider.MsgUpdateParams",
   "authority": $gov_addr,
   "params": {
    "vault_address": "$vault",
    "native_staking_address": "$native_staking"
   }
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100000000stake",
 "title": "Update params",
 "summary": "Update params"
}
EOF
)
echo $proposal
echo $proposal > ./scripts/mesh/update_params.json
meshd tx gov submit-proposal ./scripts/mesh/update_params.json --from test1 --keyring-backend test --home=$home2node1 --node $node2 --chain-id chain-2 -y 

sleep 7

meshd tx gov vote 1 yes --from val1 --keyring-backend test --home=$home2node1 --chain-id chain-2 --node $node2 -y

sleep 30

# stake from provider(chain2) '{"bond":{}}'=======bond============
meshd tx wasm execute $vault '{"bond":{"amount":{"amount": "20000000", "denom":"stake"}}}'  --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

sleep 7

account_query=$(cat <<EOF
{
    "account_details": {
        "account": "$test1_provider_addr"
    }
}
EOF
)

meshd q wasm state smart $vault "$account_query" --node $node2

# Compare tokens and select the operator address with the larger tokens value
validators=$(meshd q staking validators --output json --node $node2)
tokens_0=$(echo "$validators" | jq -r '.validators[0].tokens')
tokens_1=$(echo "$json_data" | jq -r '.validators[1].tokens')

if (( tokens_0 > tokens_1 )); then
  val2_provider_addr=$(echo "$validators" | jq -r '.validators[0].operator_address')
else
  val2_provider_addr=$(echo "$validators" | jq -r '.validators[1].operator_address')
fi

echo "validator 2: $val2_provider_addr"

# Stake 10_000_000 stake to val 1 provider chain
stake_msg=$(cat <<EOF
{"validator": "$val2_provider_addr"}
EOF
)
encode_msg=$(echo "$stake_msg" | base64)

stake_local_msg=$(cat <<EOF
{
    "stake_local":{
        "amount": {
            "denom":"stake", 
            "amount":"10000000"
        }, 
        "msg":"$encode_msg"
    }
}
EOF
)


meshd tx wasm execute $vault "$stake_local_msg" --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

# Wait a while for relaying tx to consumer chain
sleep 20

account_query=$(cat <<EOF
{
    "account_details": {
        "account": "$test1_provider_addr"
    }
}
EOF
)
meshd q wasm state smart $vault "$account_query" --node $node2

# Stop running validator 2 to make it jail
pid=$(ps waux | grep -i screen | grep -i mesh2-node2 | grep -v grep | awk '{print $2}' | xargs -I{} pgrep -P {} | xargs -I{} pgrep -P {} | xargs -I{} ps -w -p {} | grep meshd | awk '{print $1}')
kill -9 $pid

# wait for 2 minutes to jail validator
sleep 120

meshd q staking validator $val2_provider_addr --node $node2

val2_status=$(meshd q staking validator $val2_provider_addr  --node $node2 -o json| jq '.jailed')
echo "jailed: $val2_status"

# Get proxy contract
proxy_by_owner_query=$(cat <<EOF
{
    "proxy_by_owner": {"owner": "$test1_provider_addr"}
}
EOF
)
native_staking_proxy=$(meshd q wasm state smart $native_staking "$proxy_by_owner_query" --node $node2 -o json | jq -r '.data.proxy')

echo "proxy address: $native_staking_proxy"

# Unstake
unstake_msg=$(cat <<EOF
{
    "unstake":{
        "validator":"$val2_provider_addr",
        "amount": {
            "denom":"stake", 
            "amount":"10000000"
        }
    }
}
EOF
)

meshd tx wasm execute $native_staking_proxy "$unstake_msg" --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

sleep 7

meshd tx wasm execute $native_staking_proxy '{"release_unbonded": {}}' --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

sleep 7 
account_query=$(cat <<EOF
{
    "account": {
        "account": "$test1_provider_addr"
    }
}
EOF
)

meshd q wasm state smart $vault "$account_query" --node $node2

account_query=$(cat <<EOF
{
    "account_details": {
        "account": "$test1_provider_addr"
    }
}
EOF
)

meshd q wasm state smart $vault "$account_query" --node $node2
