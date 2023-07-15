## Mesh Security Devnet

> NOTE: Host and smart contract address are subject to change on redeployment of the system

### Osmosis Chain

Chain-id: `mesh-osmosis-1`
Denom: `uosmo`
Prefix: `osmo`

Host: `localhost`

Ports:
* RPC: [26657](http://localhost:26653/status)
* Rest: [1317](http://localhost:1313)
* Faucet: [8000](http://localhost:8003)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"uosmo","address":"osmo1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://localhost:8003/credit
```

#### Contracts
```
Provider Contracts:
  valut: osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9
  ExternalStaking: osmo1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftq56jurc
  nativeStaking: osmo1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3s0p34vn
  proxycode-id: 2
 Consumer Contracts:
  Staking: osmo1pvrwmjuusn9wh34j7y520g8gumuy9xtl3gvprlljfdpwju3x7ucsxrqwu2
  PriceFeed: osmo1436kxs0w2es6xlqpp9rd35e3d0cjnw4sv8j3a7483sgks29jqwgsrdzxtj
  Converter: osmo13ehuhysn5mqjeaheeuew2gjs785f6k7jm8vfsqg3jhtpkwppcmzqg496z0
```

### Juno Chain

Chain-id: `mesh-juno-1`
Denom: `ujuno`
Prefix: `juno`

Host: `localhost`

Ports:
* RPC: [26657](http://localhost:26657/status)
* Rest: [1317](http://localhost:1317/status)
* Faucet: [8000](http://localhost:8007/status)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ujuno","address":"juno1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://localhost:8007/credit
```

#### Contracts
```
Consumer Contracts:
  Staking: juno1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3skqksyr
  PriceFeed: juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8
  Converter: juno1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3seew7v3
Provider Contracts:
  valut: juno1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrq722sry
  ExternalStaking: juno1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8sdvm67v
  nativeStaking: juno1mf6ptkssddfmxvhdx0ech0k03ktp6kf9yk59renau2gvht3nq2gq0zmt6e
  proxycode-id: 5
```

## Chain Registry

Host: `localhost`

Port: [8080](http://localhost:8081/chains)

Endpoints:
* Chains: [`/chains/{chain-id}`](http://localhost:8081/chains/provider)
* IBC: [`/ibc/{chain-1}/{chain-2}`](http://localhost:8081/ibc/provider/consumer)
* Mnemonics: [`/chains/{chain-id}/keys`](http://localhost:8081/chains/provider/keys)
