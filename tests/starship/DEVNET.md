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
  valut: osmo1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqq7fzxcr
  ExternalStaking: osmo14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3lesvxz3qh
  nativeStaking: osmo1aaf9r6s7nxhysuegqrxv0wpm27ypyv4886medd3mrkrw6t4yfcnsgct3th
  proxycode-id: 6
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
  Staking: juno1j08452mqwadp8xu25kn9rleyl2gufgfjnv0sn8dvynynakkjukcqkfwuqa
  PriceFeed: juno1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrq722sry
  Converter: juno1mf6ptkssddfmxvhdx0ech0k03ktp6kf9yk59renau2gvht3nq2gq0zmt6e
```

## Chain Registry

Host: `ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com`

Port: [8080](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains)

Endpoints:
* Chains: [`/chains/{chain-id}`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains/provider)
* IBC: [`/ibc/{chain-1}/{chain-2}`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/ibc/provider/consumer)
* Mnemonics: [`/chains/{chain-id}/keys`](http://ad0908a89b64a448085c9bcbe933367c-597205196.ap-southeast-1.elb.amazonaws.com:8080/chains/provider/keys)
