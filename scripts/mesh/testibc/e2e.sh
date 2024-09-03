# !/bin/bash
# # # screen -r mesh1
# # # screen -r mesh2
# killall hermes || true
killall meshd || true
killall hermes || true
# deploy chain 1
./mesh/testibc/chain1.sh
# deploy chain 2
./mesh/testibc/chain2.sh
sleep 7
./mesh/testibc/instantiate.sh
# meshd tx interchain-accounts controller register connection-0 --from val --keyring-backend test --home=$HOME/.meshd/chain1 --chain-id chain-1 --yes
# meshd tx interchain-accounts controller register connection-0 --from val --keyring-backend test --home=$HOME/.meshd/chain2 --chain-id chain-2 --node tcp://127.0.0.1:26654 --yes
# run relayer
./mesh/testibc/hermes_bootstrap.sh


meshd tx meshsecurity submit-proposal set-virtual-staking-max-cap mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5 1000000stake --title "a title" --summary "a summary" --authority mesh1f7twgcq4ypzg7y24wuywy06xmdet8pc4lkaukm --keyring-backend test --home=$HOME/.meshd/chain1 --chain-id chain-1
# stake from provider(chain2) '{"bond":{}}'=======bond============
meshd tx wasm execute mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf '{"bond":{}}' --amount 10000stake --from test1 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 5406929

sleep 7
# ===========stake local=============================meshvaloper1w7f3xx7e75p4l7qdym5msqem9rd4dyc4u6ypev
# '{"stake_local":{"amount": {"denom":"stake", "amount":"190000000"}, "msg":"eyJ2YWxpZGF0b3IiOiAiY29zbW9zdmFsb3BlcjF5OHNzMzR6Y2RncTlnM3l6MDNyZjZrZ3J1ajZ6cGFneGh6ZXFmbSJ9"}}'
# msg se thay doi do validator thay doi base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, "meshvaloper1w7f3xx7e75p4l7qdym5msqem9rd4dyc4u6ypev"))))
meshd tx wasm execute mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf '{"stake_local":{"amount": {"denom":"stake", "amount":"9000"}, "msg":"eyJ2YWxpZGF0b3IiOiAibWVzaHZhbG9wZXIxdzdmM3h4N2U3NXA0bDdxZHltNW1zcWVtOXJkNGR5YzR1NnlwZXYifQ=="}}' --from test1 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 5406929

sleep 7

# =========extenal============================meshvaloper1f7twgcq4ypzg7y24wuywy06xmdet8pc4hsl6ty
# '{"stake_remote":{"contract":"cosmos1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqp82y57", "amount": {"denom":"stake", "amount":"100000000"}, "msg":"eyJ2YWxpZGF0b3IiOiAiY29zbW9zdmFsb3BlcjE0cjNhbmRuM3FyaDJ6NzN1eTNhNjZ2YzgyazU1ZHE1ZG04MmFxeSJ9"}}'
meshd tx wasm execute mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf '{"stake_remote":{"contract":"mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5", "amount": {"denom":"stake", "amount":"1000"}, "msg":"eyJ2YWxpZGF0b3IiOiAibWVzaHZhbG9wZXIxZjd0d2djcTR5cHpnN3kyNHd1eXd5MDZ4bWRldDhwYzRoc2w2dHkifQ=="}}' --from val --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 5406929

# {"stake_remote":{"contract":"cosmos1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqp82y57", "amount": {"denom":"stake", "amount":"100000000"}, "msg":"eyJ2YWxpZGF0b3IiOiAiY29zbW9zdmFsb3BlcjF5czA0bnNhbTAyeHA1cGt3Y3A2eGozeTd6bDd0emtmZnN3ampteSJ9"}}

# meshd tx wasm execute mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5 '{"stake_remote":{"contract":"mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5", "amount": {"denom":"stake", "amount":"1000"}, "msg":"eyJ2YWxpZGF0b3IiOiAibWVzaHZhbG9wZXIxZjd0d2djcTR5cHpnN3kyNHd1eXd5MDZ4bWRldDhwYzRoc2w2dHkifQ=="}}' --amount 1000stake  --from test1 --home=$HOME/.meshd/chain2  --chain-id chain-2 --keyring-backend test --node tcp://127.0.0.1:26654 --fees 1stake -y --gas 5406929


# echo '{"validator": "meshvaloper1w7f3xx7e75p4l7qdym5msqem9rd4dyc4u6ypev"}' | base64
