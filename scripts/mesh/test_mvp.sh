# !/bin/bash
killall meshconsumerd || true
killall meshproviderd || true
killall rly || true
# deploy chain 1
./scripts/mesh/testibc/chain1.sh
# deploy chain 2
./scripts/mesh/testibc/chain2.sh
sleep 7

./scripts/mesh/testibc/rly_band.sh

./scripts/mesh/testibc/instantiate.sh
# run relayer
./scripts/mesh/testibc/rly.sh

home1node1=$HOME/.meshd/chain1/node1
home2node1=$HOME/.meshd/chain2/node1
node1=tcp://127.0.0.1:26657
node2=tcp://127.0.0.1:26667

virtual_staking=$(meshconsumerd q wasm list-contract-by-code 2 --output json --node $node1| jq -r '.contracts[0]' )
converter=$(meshconsumerd q wasm list-contract-by-code 3 --output json --node $node1 | jq -r '.contracts[0]' )
vault=$(meshproviderd q wasm list-contract-by-code 1 --output json --node $node2 | jq -r '.contracts[0]' )
native_staking=$(meshproviderd q wasm list-contract-by-code 3 --output json --node $node2 | jq -r '.contracts[0]' )
ext_staking=$(meshproviderd q wasm list-contract-by-code 4 --output json --node $node2 | jq -r '.contracts[0]' )
test1_provider_addr=$(meshproviderd keys show test1 --keyring-backend test --home=$home2node1 --address)
meshconsumerd tx meshsecurity submit-proposal set-virtual-staking-max-cap $virtual_staking 100000000stake --title "a title" --summary "a summary" --from test1 --keyring-backend test --home=$home1node1 --node $node1 --chain-id chain-1 -y --deposit 10000000stake

sleep 7

meshconsumerd tx gov vote 1 yes --from val1 --keyring-backend test --node $node1 --home=$home1node1 --chain-id chain-1 -y

sleep 5

gov_addr=$(meshproviderd q auth --node $node2  module-account gov -o json | jq ".account.base_account.address")

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
meshproviderd tx gov submit-proposal ./scripts/mesh/update_params.json --from test1 --keyring-backend test --home=$home2node1 --node $node2 --chain-id chain-2 -y 

sleep 7

meshproviderd tx gov vote 1 yes --from val1 --keyring-backend test --home=$home2node1 --chain-id chain-2 --node $node2 -y

sleep 30

# stake from provider(chain2) '{"bond":{}}'=======bond============
meshproviderd tx wasm execute $vault '{"bond":{"amount":{"amount": "10000", "denom":"stake"}}}'  --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

sleep 7

account_query=$(cat <<EOF
{
    "account_details": {
        "account": "$test1_provider_addr"
    }
}
EOF
)

meshproviderd q wasm state smart $vault "$account_query" --node $node2
# ===========stake local=============================meshvaloper1w7f3xx7e75p4l7qdym5msqem9rd4dyc4u6ypev
# '{"stake_local":{"amount": {"denom":"stake", "amount":"190000000"}, "msg":"eyJ2YWxpZGF0b3IiOiAiY29zbW9zdmFsb3BlcjF5OHNzMzR6Y2RncTlnM3l6MDNyZjZrZ3J1ajZ6cGFneGh6ZXFmbSJ9"}}'
val1_provider_addr=$(meshproviderd q staking validators --output json --node $node2 | jq -r '.validators[0].operator_address')
stake_msg=$(cat <<EOF
{"validator": "$val1_provider_addr"}
EOF
)
encode_msg=$(echo "$stake_msg" | base64)

stake_local_msg=$(cat <<EOF
{
    "stake_local":{
        "amount": {
            "denom":"stake", 
            "amount":"9000"
        }, 
        "msg":"$(echo -n $encode_msg| tr -d '[:space:]')"
    }
}
EOF
)
meshproviderd tx wasm execute $vault "$stake_local_msg" --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

sleep 7

# =========extenal============================meshvaloper1f7twgcq4ypzg7y24wuywy06xmdet8pc4hsl6ty
# '{"stake_remote":{"contract":"cosmos1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqp82y57", "amount": {"denom":"stake", "amount":"100000000"}, "msg":"eyJ2YWxpZGF0b3IiOiAiY29zbW9zdmFsb3BlcjE0cjNhbmRuM3FyaDJ6NzN1eTNhNjZ2YzgyazU1ZHE1ZG04MmFxeSJ9"}}'
val1_consumer_addr=$(meshconsumerd q staking validators --output json --node $node1| jq -r '.validators[0].operator_address')
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
            "amount":"1000"
        }, 
        "msg":"$(echo -n $encode_msg| tr -d '[:space:]')"
    }
}
EOF
)

meshproviderd tx wasm execute $vault "$stake_remote_msg" --from test1 --home=$home2node1  --chain-id chain-2 --keyring-backend test --node $node2 --fees 1stake -y --gas 15406929

# Wait a while for relaying tx to consumer chain
sleep 20

# {"stake_remote":{"contract":"cosmos1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqp82y57", "amount": {"denom":"stake", "amount":"100000000"}, "msg":"eyJ2YWxpZGF0b3IiOiAiY29zbW9zdmFsb3BlcjF5czA0bnNhbTAyeHA1cGt3Y3A2eGozeTd6bDd0emtmZnN3ampteSJ9"}}
stake_query=$(cat <<EOF
{
    "stake": {
        "user": "$test1_provider_addr",
        "validator": "$val1_consumer_addr"
    }
}
EOF
)
meshproviderd q wasm state smart $ext_staking "$stake_query" --node $node2
