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
meshd tx wasm execute $vault '{"bond":{"amount":{"amount": "100000000", "denom":"stake"}}}'  --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

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

# Stake 50_000_000 stake to val 1 consumer chain
val1_consumer_addr=$(meshd q staking validators --output json | jq -r '.validators[0].operator_address')
stake_msg=$(cat <<EOF
{"validator": "$val1_consumer_addr"}
EOF
)
encode_msg=$(echo "$stake_msg" | base64)

stake_remote_msg=$(cat <<EOF
{
    "stake_remote":{
        "contract":"$ext_staking", 
        "amount": {
            "denom":"stake", 
            "amount":"50000000"
        }, 
        "msg":"$encode_msg"
    }
}
EOF
)


meshd tx wasm execute $vault "$stake_remote_msg" --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

# Wait a while for relaying tx to consumer chain
sleep 20

# Stake 20_000_000 stake to val 1 consumer chain
val2_consumer_addr=$(meshd q staking validators --output json | jq -r '.validators[1].operator_address')
stake_msg=$(cat <<EOF
{"validator": "$val2_consumer_addr"}
EOF
)
encode_msg=$(echo "$stake_msg" | base64)

stake_remote_msg=$(cat <<EOF
{
    "stake_remote":{
        "contract":"$ext_staking", 
        "amount": {
            "denom":"stake", 
            "amount":"20000000"
        }, 
        "msg":"$encode_msg"
    }
}
EOF
)


meshd tx wasm execute $vault "$stake_remote_msg" --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

# Wait a while for relaying tx to consumer chain
sleep 20

# Check if the delegate is increased in consumer chain
meshd q meshsecurity max-cap-limits
# Create a proposal to set max cap to zero
meshd tx meshsecurity submit-proposal set-virtual-staking-max-cap $virtual_staking 0stake --title "a title" --summary "a summary" --from test1 --keyring-backend test --home=$home1node1 --chain-id chain-1 -y --deposit 10000000stake

sleep 7

meshd tx gov vote 2 yes --from val1 --keyring-backend test --home=$home1node1 --chain-id chain-1 -y

sleep 30

meshd q meshsecurity max-cap-limits
# Sleep for a while to wait for relaying ibc packet
sleep 30

meshd tx wasm execute $ext_staking '{"withdraw_unbonded":{}}' --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

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

