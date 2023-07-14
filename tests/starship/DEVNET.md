## Mesh Security Devnet

> NOTE: Host and smart contract address are subject to change on redeployment of the system

### Provider Chain

Chain-id: `provider`
Denom: `uosmo`

Host: `ac455e48cce834e60a7ebdae648dc73d-382195732.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://ac455e48cce834e60a7ebdae648dc73d-382195732.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://ac455e48cce834e60a7ebdae648dc73d-382195732.ap-southeast-1.elb.amazonaws.com:1317)
* Faucet: [8000](http://ac455e48cce834e60a7ebdae648dc73d-382195732.ap-southeast-1.elb.amazonaws.com:8000)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"mesh1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://ac455e48cce834e60a7ebdae648dc73d-382195732.ap-southeast-1.elb.amazonaws.com:8000/credit
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

Host: `a2fec3c7e87014bfeaa3ae755158ce9e-885015271.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://a2fec3c7e87014bfeaa3ae755158ce9e-885015271.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://a2fec3c7e87014bfeaa3ae755158ce9e-885015271.ap-southeast-1.elb.amazonaws.com:1317/status)
* Faucet: [8000](http://a2fec3c7e87014bfeaa3ae755158ce9e-885015271.ap-southeast-1.elb.amazonaws.com:8000/status)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"mesh1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://a2fec3c7e87014bfeaa3ae755158ce9e-885015271.ap-southeast-1.elb.amazonaws.com:8000/credit
```

```
Consumer Contracts:
  Staking: mesh1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3syz4y6d
  PriceFeed: mesh14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sysl6kf
  Converter: mesh1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3stmd2jl
```

## Chain Registry

Host: `a5d3c50ca6a6245d3ab6bd54d64a559f-1780147229.ap-southeast-1.elb.amazonaws.com`

Port: [8080](http://a5d3c50ca6a6245d3ab6bd54d64a559f-1780147229.ap-southeast-1.elb.amazonaws.com:8080/chains)

Endpoints:
* Chains: [`/chains/{chain-id}`](http://a5d3c50ca6a6245d3ab6bd54d64a559f-1780147229.ap-southeast-1.elb.amazonaws.com:8080/chains/provider)
* IBC: [`/ibc/{chain-1}/{chain-2}`](http://a5d3c50ca6a6245d3ab6bd54d64a559f-1780147229.ap-southeast-1.elb.amazonaws.com:8080/ibc/provider/consumer)
* Mnemonics: [`/chains/{chain-id}/keys`](http://a5d3c50ca6a6245d3ab6bd54d64a559f-1780147229.ap-southeast-1.elb.amazonaws.com:8080/chains/provider/keys)
