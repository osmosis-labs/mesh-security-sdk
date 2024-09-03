# !/bin/bash

home2=$HOME/.meshd/chain2/

val2=$(meshd keys show val1  --keyring-backend test -a --home=$home2)


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
