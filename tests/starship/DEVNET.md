## Mesh Security Devnet

> NOTE: Host and smart contract address are subject to change on redeployment of the system

### Provider Chain

Chain-id: `provider`
Denom: `uosmo`

Host: `af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:1317)
* Faucet: [8000](http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:8000)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"mesh1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:8000/credit
```

```
Provider Contracts:
  valut: mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf
  ExternalStaking: mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5
  nativeStaking: mesh1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3stmd2jl
```

### Consumer Chain

Chain-id: `consumer`
Denom: `ujuno`
Prefix: `mesh`

Host: `ae972d435ccad4ff3875cc05f31be3cb-209753913.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://ae972d435ccad4ff3875cc05f31be3cb-209753913.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://ae972d435ccad4ff3875cc05f31be3cb-209753913.ap-southeast-1.elb.amazonaws.com:1317/status)
* Faucet: [8000](http://ae972d435ccad4ff3875cc05f31be3cb-209753913.ap-southeast-1.elb.amazonaws.com:8000/status)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"mesh1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://ae972d435ccad4ff3875cc05f31be3cb-209753913.ap-southeast-1.elb.amazonaws.com:8000/credit
```

```
Consumer Contracts:
  Staking: mesh1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3syz4y6d
  PriceFeed: mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf
  Converter: mesh1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3stmd2jl
```

## Chain Registry

Host: `ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com`

Port: [8080](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains)

Endpoints:
* Chains: [`/chains/{chain-id}`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains/provider)
* IBC: [`/ibc/{chain-1}/{chain-2}`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/ibc/provider/consumer)
* Mnemonics: [`/chains/{chain-id}/keys`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains/provider/keys)
