# Integrate mesh security using band oracle price source.

## Prerequisites

- Setup mesh-security consumer and provider chains as [integration_consumer](./integration_consumer.md), [integration_provider](./integration_provider.md)
- Choose a price source for 2 chains in [band laozi testnet](https://laozi-testnet6.cosmoscan.io/data-sources), [band mainnet](https://www.cosmoscan.io/data-sources)

You can request 4 symbols on [oracle script id: 360](https://laozi-testnet6.cosmoscan.io/oracle-script/360)

- BTC, ETH, USDT, BAND

## Band Price Feeder

Band oracle price feeder is design base on cw-band. More details could be found in: <https://github.com/bandprotocol/cw-band/tree/main/docs>

The contract needs some information to instantiate, which is:

```
trading_pair: {
    base_asset: String,
    quote_asset: String,
} -> Consumer chain denom is base_asset, provider chain denom is quote_asset 
client_id: String -> Arbitary string for your request
oracle_script_id: Uint64, -> The data source's id you choose 
ask_count: Uint64, -> The number of validator you want to ask (Recommend: 4 on testnet)
min_count: Uint64, -> The minimum number of validator need to answer to aggregate result (Recommend: 3 on testnet)
fee_limit: Vec<Coin>, -> Data source fee that you willing to pay (Recommend: 250000uband, which is [{"denom": "uband", "amount":"250000"}])
prepare_gas: Uint64, -> Gas for running prepare phrase (Recommend: 100000)
execute_gas: Uint64, -> Gas for running execute phrase (Recommend: 500000)
minimum_sources: u8, -> The minimum available sources to determine price is aggregated from at least minimum sources (for data integrity) 1 should be ok for testing
price_info_ttl_in_secs: u64, -> The price only can live in an amount of time, after that, they should be update (Recommend: 60) 
```

## Configure relayer

1. Create config for consumer chain and band chain as usual.
2. Create path, clients, connections
3. Deploy band price feeder on consumer chain
4. Setup channel using port wasm on consumer chain, port oracle on band chain and version bandchain-1
5. Using contract address as price source in converter contract
Deploy contract:

```bash
meshconsumerd tx wasm store ./tests/testdata/mesh_band_price_feed.wasm.gz --node $node --from $wallet --home=$home  --chain-id $chainid --keyring-backend test --gas 300000 -y

band_code_id=$(meshconsumerd q wasm  list-code --node $node --output json | jq -r '.code_infos[-1].code_id')
client_id=$(yq e '.paths.{$PATH_NAME}.src.client-id' $relayer_home/config/config.yaml)
oracle_script_id=445

init_band_price_feed=$(cat <<EOF
{
    "trading_pair": {
        "base_asset": "OSMO",
        "quote_asset": "ATOM"
    },
    "client_id": "$client_id",
    "oracle_script_id": "$oracle_script_id",
    "ask_count": "1",
    "min_count": "1",
    "fee_limit": [{"denom": "uband", "amount":"10000"}],
    "prepare_gas": "40000",
    "execute_gas": "300000",
    "minimum_sources": 2,
    "price_info_ttl_in_secs": 60
}
EOF
)
meshconsumerd tx wasm instantiate $band_code_id "$init_band_price_feed" --node $node --label contract-pricefeed  --admin $wallet --from $wallet --home=$home  --chain-id $chainid --keyring-backend test -y --gas 300000
sleep 7

price_feed=$(meshconsumerd q wasm list-contract-by-code $band_code_id --node $node --output json | jq -r '.contracts[0]' )
echo "price feed contract: $price_feed"

```

Create channel

```bash
rly tx channel {$PATH_NAME} --src-port wasm.$price_feed --dst-port oracle --order unordered --version bandchain-1 --home $relayer_home --override
```
