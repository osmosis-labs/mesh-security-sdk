## Mesh Security Devnet

> NOTE: Host and smart contract address are subject to change on redeployment of the system

### Provider Chain

Chain-id: `provider`
Denom: `uosmo`
Prefix: `osmo`

Host: `af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:1317)
* Faucet: [8000](http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:8000)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"osmo1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://af776f7a0c6b74a0592cd0f960db567a-587001225.ap-southeast-1.elb.amazonaws.com:8000/credit
```

```
Provider Contracts:
  valut: osmo1plr28ztj64a47a32lw7tdae8vluzm2lm7nqk364r4ws50rgwyzgsvucjw7
  ExternalStaking: osmo1zqtwuecz2k9g5xs6q4vsahnvj7rkax8gwmanygppeudvmzv6txqqkga9a5
  nativeStaking: osmo19h0d6k4mtxw5qjr0aretjy9kwyem0hxclf88ka2uwjn47e90mqrqg68p6j
  proxycode-id: 10
```

### Consumer Chain

Chain-id: `consumer`
Denom: `ujuno`
Prefix: `juno`

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
  Staking: juno1rl8su3hadqqq2v86lscpuklsh2mh84cxqvjdew4jt9yd07dzekyqr2vlrj
  PriceFeed: juno10qt8wg0n7z740ssvf3urmvgtjhxpyp74hxqvqt7z226gykuus7eqs24840
  Converter: juno1gurgpv8savnfw66lckwzn4zk7fp394lpe667dhu7aw48u40lj6jsez6g8g
```

## Chain Registry

Host: `ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com`

Port: [8080](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains)

Endpoints:
* Chains: [`/chains/{chain-id}`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains/provider)
* IBC: [`/ibc/{chain-1}/{chain-2}`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/ibc/provider/consumer)
* Mnemonics: [`/chains/{chain-id}/keys`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains/provider/keys)
