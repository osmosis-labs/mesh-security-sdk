## Mesh Security Devnet

> NOTE: Host and smart contract address are subject to change on redeployment of the system

### Provider Chain

Chain-id: `provider`
Denom: `uosmo`

Host: `a8dbf4ded5fde419b9c068489cef72ff-704314997.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://a8dbf4ded5fde419b9c068489cef72ff-704314997.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://a8dbf4ded5fde419b9c068489cef72ff-704314997.ap-southeast-1.elb.amazonaws.com:1317)
* Faucet: [8000](http://a8dbf4ded5fde419b9c068489cef72ff-704314997.ap-southeast-1.elb.amazonaws.com:8000)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"mesh1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://a8dbf4ded5fde419b9c068489cef72ff-704314997.ap-southeast-1.elb.amazonaws.com:8000/credit
```

```
Provider Contracts:
  valut: mesh1xhcxq4fvxth2hn3msmkpftkfpw73um7s4et3lh4r8cfmumk3qsmsw3arzh
  ExternalStaking: mesh1m74wv3xew5dsy2thf3jp0xadg8pdrk4h8ym70z0ehfwxl8a547asz8p7f7
  nativeStaking: mesh1uvt40rsp68wtas0y75w34qdn5h0g5eyefy5gmvzftdnupyv7q7vq5jz5rv
```

### Consumer Chain

Chain-id: `consumer`
Denom: `ujuno`
Prefix: `mesh`

Host: `a98000cb656f34d12a138f0189a8ca38-543023852.ap-southeast-1.elb.amazonaws.com`

Ports:
* RPC: [26657](http://a98000cb656f34d12a138f0189a8ca38-543023852.ap-southeast-1.elb.amazonaws.com:26657/status)
* Rest: [1317](http://a98000cb656f34d12a138f0189a8ca38-543023852.ap-southeast-1.elb.amazonaws.com:1317/status)
* Faucet: [8000](http://a98000cb656f34d12a138f0189a8ca38-543023852.ap-southeast-1.elb.amazonaws.com:8000/status)

Get tokens from faucet
```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"denom":"ustake","address":"mesh1yre6ac7qfgyfgvh58ph0rgw627rhw766y430qq"}' \
  http://a98000cb656f34d12a138f0189a8ca38-543023852.ap-southeast-1.elb.amazonaws.com:8000/credit
```

```
Consumer Contracts:
  Staking: mesh1fuyxwxlsgjkfjmxfthq8427dm2am3ya3cwcdr8gls29l7jadtazsvvhd03
  PriceFeed: mesh13we0myxwzlpx8l5ark8elw5gj5d59dl6cjkzmt80c5q5cv5rt54qu2r0a5
  Converter: mesh18yn206ypuxay79gjqv6msvd9t2y49w4fz8q7fyenx5aggj0ua37q6xkydx
```

## Chain Registry

Host: `a87656a56ccec4d06aa33db894dff957-1524822698.ap-southeast-1.elb.amazonaws.com`

Port: [8080](http://a87656a56ccec4d06aa33db894dff957-1524822698.ap-southeast-1.elb.amazonaws.com:8080/chains)

Endpoints:
* Chains: [`/chains/{chain-id}`](http://a87656a56ccec4d06aa33db894dff957-1524822698.ap-southeast-1.elb.amazonaws.com:8080/chains/provider)
* IBC: [`/ibc/{chain-1}/{chain-2}`](http://a87656a56ccec4d06aa33db894dff957-1524822698.ap-southeast-1.elb.amazonaws.com:8080/ibc/provider/consumer)
* Mnemonics: [`/chains/{chain-id}/keys`](http://a87656a56ccec4d06aa33db894dff957-1524822698.ap-southeast-1.elb.amazonaws.com:8080/chains/provider/keys)
