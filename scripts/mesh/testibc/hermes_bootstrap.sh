#!/bin/bash

set -ex

# initialize Hermes relayer configuration
touch $HOME/.hermes/config.toml

# setup Hermes relayer configuration
tee $HOME/.hermes/config.toml <<EOF
[global]
log_level = "info"

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = true
clear_interval = 10
clear_on_start = true
tx_confirmation = true
auto_register_counterparty_payee = true

[mode.packets.ics20_max_memo_size]
enabled = true
size = 32768

[mode.packets.ics20_max_receiver_size]
enabled = true
size = 2048

[rest]
enabled = false
host = "127.0.0.1"
port = 3000

[telemetry]
enabled = false
host = "127.0.0.1"
port = 3001

[telemetry.buckets.latency_submitted]
start = 500
end = 20000
buckets = 10

[telemetry.buckets.latency_confirmed]
start = 1000
end = 30000
buckets = 10

[[chains]]
type = "CosmosSdk"
id = "chain-1"
rpc_addr = "http://localhost:26657"
grpc_addr = "http://localhost:9090"
rpc_timeout = "10s"
trusted_node = true
account_prefix = "mesh"
key_name = "decentrio-main1"
key_store_type = "Test"
store_prefix = "ibc"
default_gas = 700000
max_gas = 9000000
gas_multiplier = 1.3
max_msg_num = 2
max_tx_size = 180000
max_grpc_decoding_size = 33554432
query_packets_chunk_size = 50
clock_drift = "5s"
max_block_time = "30s"
client_refresh_rate = "1/3"
ccv_consumer_chain = false
memo_prefix = "Relayed by Decentrio"
sequential_batch_tx = true
allow_ccq = false

[chains.event_source]
mode = "push"
interval = "100ms"
url = "ws://localhost:26657/websocket"
max_retries = 4

[chains.trust_threshold]
numerator = 2
denominator = 3

[chains.gas_price]
price = 0.025
denom = "stake"

[chains.packet_filter]
policy = "allow"
list = [
#[
#  "transfer",
 #  "channel-*",
#]
]

[chains.packet_filter.min_fees]

[chains.dynamic_gas_price]
enabled = true
multiplier = 1.1
max = 0.6

[chains.address_type]
derivation = "cosmos"

[chains.excluded_sequences]

[[chains]]
type = "CosmosSdk"
id = "chain-2"
rpc_addr = "http://localhost:26654"
grpc_addr = "http://localhost:9088"
rpc_timeout = "10s"
trusted_node = true
account_prefix = "mesh"
key_name = "decentrio-main2"
key_store_type = "Test"
store_prefix = "ibc"
default_gas = 700000
max_gas = 9000000
gas_multiplier = 1.5
max_msg_num = 2
max_tx_size = 180000
max_grpc_decoding_size = 33554432
query_packets_chunk_size = 50
clock_drift = "5s"
max_block_time = "30s"
client_refresh_rate = "1/3"
ccv_consumer_chain = false
memo_prefix = "Relayed by Decentrio"
sequential_batch_tx = true
allow_ccq = true

[chains.event_source]
mode = "push"
interval = "100ms"
url= "ws://localhost:26654/websocket"
max_retries = 4

[chains.trust_threshold]
numerator = 2
denominator = 3

[chains.gas_price]
price = 0.005
denom = "stake"

[chains.packet_filter]
policy = "allow"
list = [[
  "transfer",
  "channel-*",
]]

[chains.packet_filter.min_fees]

[chains.dynamic_gas_price]
enabled = false
multiplier = 1.1
max = 0.5

[chains.address_type]
derivation = "cosmos"

[chains.excluded_sequences]

[tracing_server]
enabled = false
port = 5555
EOF

# import keys
hermes version

hermes keys add --chain chain-1 --key-name "decentrio-main1" --mnemonic-file ./mesh/testibc/testdata/mnemonic1 --overwrite
hermes keys add --chain chain-2 --key-name "decentrio-main2" --mnemonic-file ./mesh/testibc/testdata/mnemonic2 --overwrite

# start Hermes relayer
screen -S hermes1 -t hermes1 -d -m  hermes start
sleep 7
# hermes create channel --a-chain chain-1 --b-chain chain-2 --a-port transfer --b-port transfer --yes --new-client-connection

hermes create connection  --a-chain chain-1 --b-chain chain-2
sleep 7
hermes create channel --a-chain chain-1 --a-connection connection-0  --a-port wasm.mesh1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3stmd2jl  --b-port wasm.mesh1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqsqwra5
