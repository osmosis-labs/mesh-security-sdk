# Mesh security deployment
This document describes about how to deploy and setup mesh security for cosmos chains

## Consumer chain
1. Integrate `meshsecurity` module by this [guide](./x/meshsecurity/README.md).
2. Deploy price feed contract:
- Store code file `mesh_simple_price_feed.wasm`
- Instantiate contract with parameters: `{"native_per_foreign": $token_ratio}`
3. Deploy converter contract:
- Store code file `mesh_virtual_staking.wasm`
- Store code file `mesh_converter.wasm`
- Instantiate contract with parameters:
`{"price_feed": $price_feed_code_id, "discount": $discount, "remote_denom": $denom,"virtual_staking_code_id": $virtual_staking_code_id, "max_retrieve": $max_retrieve}`
4. Set virtual staking maximum capacity:
- Create proposal with message `SetVirtualStakingMaxCap`:
```go=
type MsgSetVirtualStakingMaxCap struct {
	Authority string 
	Contract string 
	MaxCap types.Coin 
}
```
## Provider chain
1. Integrate `meshsecurityprovider` module by this [guide](./x/meshsecurityprovider/README.md).
2. Store contract codes:
- `mesh_vault.wasm`
- `mesh_native_staking_proxy.wasm`
- `mesh_native_staking.wasm`
- `mesh_external_staking.wasm`
3. Deploy vault contract:
- Std encoding instantiate native staking contract:
`{"denom": $denom, "proxy_code_id": $native_staking_proxy_code_id, "slash_ratio_dsign": $ratio_dsign, "slash_ratio_offline": $ratio }
- Instantiate vault contract with parameters:
`{"denom": $denom, "local_staking": {"code_id": $native_staking_code_id, "msg": $native_staking_init_msg}}`
4. Deploy external staking contract:
- Instantiate contract with parameters:
`{"remote_contact": {"connection_id":$connection_id, "port_id":$port_id}, "denom": $denom, "vault": $vault_address, "unbonding_period": $unbonding_period, "rewards_denom": $reward_denom, "slash_ratio": { "double_sign": $ratio_dsign, "offline": $ratio }  }`
5. Set vault address and native staking address params:
- Create proposal with message: `UpdateParams`:
```go=
type MsgUpdateParams struct {
	Authority string 
	Params Params `protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}
```
## Setup IBC

1. Create config for both chains as usual.
2. Create path, clients, connections
3. Create channel with port `wasm.$converter` for provider chain, port `wasm.$ext_staking` for consumer chain, version is `'{"protocol":"mesh-security","version":"0.11.0"}'`

This is a template scripts for starting go-relayer:
```bash=
rly paths new consumer-chain consumer-chain demo

sleep 5

rly tx clients demo

sleep 10

rly tx connection demo

sleep 10

converter=$(meshd q wasm list-contract-by-code $converter_code_id --output json --node $consumer_node| jq -r '.contracts[0]' )
ext_staking=$(meshd q wasm list-contract-by-code $ext_staking_code_id --output json --node $provider_node | jq -r '.contracts[0]' )

rly tx channel demo --src-port wasm.$converter --dst-port wasm.$ext_staking --order unordered --version '{"protocol":"mesh-security","version":"0.11.0"}'

sleep 5

screen -S relayer -t relayer -d -m  rly start demo
sleep 5
```
