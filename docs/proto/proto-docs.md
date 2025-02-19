<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [osmosis/meshsecurity/v1beta1/meshsecurity.proto](#osmosis/meshsecurity/v1beta1/meshsecurity.proto)
    - [Delegation](#osmosis.meshsecurity.v1beta1.Delegation)
    - [Params](#osmosis.meshsecurity.v1beta1.Params)
    - [VirtualStakingMaxCapInfo](#osmosis.meshsecurity.v1beta1.VirtualStakingMaxCapInfo)
  
- [osmosis/meshsecurity/v1beta1/genesis.proto](#osmosis/meshsecurity/v1beta1/genesis.proto)
    - [GenesisState](#osmosis.meshsecurity.v1beta1.GenesisState)
  
- [osmosis/meshsecurity/v1beta1/query.proto](#osmosis/meshsecurity/v1beta1/query.proto)
    - [QueryParamsRequest](#osmosis.meshsecurity.v1beta1.QueryParamsRequest)
    - [QueryParamsResponse](#osmosis.meshsecurity.v1beta1.QueryParamsResponse)
    - [QueryVirtualStakingMaxCapLimitRequest](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitRequest)
    - [QueryVirtualStakingMaxCapLimitResponse](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitResponse)
    - [QueryVirtualStakingMaxCapLimitsRequest](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsRequest)
    - [QueryVirtualStakingMaxCapLimitsResponse](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsResponse)
  
    - [Query](#osmosis.meshsecurity.v1beta1.Query)
  
- [osmosis/meshsecurity/v1beta1/scheduler.proto](#osmosis/meshsecurity/v1beta1/scheduler.proto)
    - [ValidatorAddress](#osmosis.meshsecurity.v1beta1.ValidatorAddress)
  
- [osmosis/meshsecurity/v1beta1/tx.proto](#osmosis/meshsecurity/v1beta1/tx.proto)
    - [MsgSetPriceFeedContract](#osmosis.meshsecurity.v1beta1.MsgSetPriceFeedContract)
    - [MsgSetPriceFeedContractResponse](#osmosis.meshsecurity.v1beta1.MsgSetPriceFeedContractResponse)
    - [MsgSetVirtualStakingMaxCap](#osmosis.meshsecurity.v1beta1.MsgSetVirtualStakingMaxCap)
    - [MsgSetVirtualStakingMaxCapResponse](#osmosis.meshsecurity.v1beta1.MsgSetVirtualStakingMaxCapResponse)
  
    - [Msg](#osmosis.meshsecurity.v1beta1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="osmosis/meshsecurity/v1beta1/meshsecurity.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurity/v1beta1/meshsecurity.proto



<a name="osmosis.meshsecurity.v1beta1.Delegation"></a>

### Delegation
Delegation represents the bond with tokens held by an account.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegator_address` | [string](#string) |  | delegator_address is the bech32-encoded address of the delegator. |
| `validator_address` | [string](#string) |  | validator_address is the bech32-encoded address of the validator. |
| `amount` | [string](#string) |  | amount define the delegation amount. |






<a name="osmosis.meshsecurity.v1beta1.Params"></a>

### Params
Params defines the parameters for the x/meshsecurity module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `total_contracts_max_cap` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | TotalContractsMaxCap is the maximum that the sum of all contract max caps must not exceed |
| `epoch_length` | [uint32](#uint32) |  | Epoch length is the number of blocks that defines an epoch |
| `max_gas_end_blocker` | [uint32](#uint32) |  | MaxGasEndBlocker defines the maximum gas that can be spent in a contract sudo callback |






<a name="osmosis.meshsecurity.v1beta1.VirtualStakingMaxCapInfo"></a>

### VirtualStakingMaxCapInfo
VirtualStakingMaxCapInfo stores info about
virtual staking max cap


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | Contract is the address of the contract |
| `delegated` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Delegated is the total amount currently delegated |
| `cap` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Cap is the current max cap limit |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/meshsecurity/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurity/v1beta1/genesis.proto



<a name="osmosis.meshsecurity.v1beta1.GenesisState"></a>

### GenesisState
GenesisState defines meshsecurity module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.meshsecurity.v1beta1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/meshsecurity/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurity/v1beta1/query.proto



<a name="osmosis.meshsecurity.v1beta1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the
Query/Params RPC method






<a name="osmosis.meshsecurity.v1beta1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the
Query/Params RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.meshsecurity.v1beta1.Params) |  |  |






<a name="osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitRequest"></a>

### QueryVirtualStakingMaxCapLimitRequest
QueryVirtualStakingMaxCapLimitRequest is the request type for the
Query/VirtualStakingMaxCapLimit RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the address of the contract to query |






<a name="osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitResponse"></a>

### QueryVirtualStakingMaxCapLimitResponse
QueryVirtualStakingMaxCapLimitResponse is the response type for the
Query/VirtualStakingMaxCapLimit RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegated` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `cap` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsRequest"></a>

### QueryVirtualStakingMaxCapLimitsRequest
QueryVirtualStakingMaxCapLimitsRequest is the request type for the
Query/VirtualStakingMaxCapLimits RPC method






<a name="osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsResponse"></a>

### QueryVirtualStakingMaxCapLimitsResponse
QueryVirtualStakingMaxCapLimitsResponse is the response type for the
Query/VirtualStakingMaxCapLimits RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `max_cap_infos` | [VirtualStakingMaxCapInfo](#osmosis.meshsecurity.v1beta1.VirtualStakingMaxCapInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.meshsecurity.v1beta1.Query"></a>

### Query
Query provides defines the gRPC querier service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `VirtualStakingMaxCapLimit` | [QueryVirtualStakingMaxCapLimitRequest](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitRequest) | [QueryVirtualStakingMaxCapLimitResponse](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitResponse) | VirtualStakingMaxCapLimit gets max cap limit for the given contract | GET|/osmosis/meshsecurity/v1beta1/max_cap_limit/{address}|
| `VirtualStakingMaxCapLimits` | [QueryVirtualStakingMaxCapLimitsRequest](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsRequest) | [QueryVirtualStakingMaxCapLimitsResponse](#osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsResponse) | VirtualStakingMaxCapLimits gets max cap limits | GET|/osmosis/meshsecurity/v1beta1/max_cap_limits|
| `Params` | [QueryParamsRequest](#osmosis.meshsecurity.v1beta1.QueryParamsRequest) | [QueryParamsResponse](#osmosis.meshsecurity.v1beta1.QueryParamsResponse) | Params queries the parameters of x/meshsecurity module. | GET|/osmosis/meshsecurity/v1beta1/params|

 <!-- end services -->



<a name="osmosis/meshsecurity/v1beta1/scheduler.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurity/v1beta1/scheduler.proto



<a name="osmosis.meshsecurity.v1beta1.ValidatorAddress"></a>

### ValidatorAddress
ValidatorAddress payload data to be used with the scheduler


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the ValAddress bech32 string |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/meshsecurity/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurity/v1beta1/tx.proto



<a name="osmosis.meshsecurity.v1beta1.MsgSetPriceFeedContract"></a>

### MsgSetPriceFeedContract
MsgSetPriceFeedContract sets the price feed contract to the chain 
to trigger handle epoch task


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address that controls the module (defaults to x/gov unless overwritten). |
| `contract` | [string](#string) |  | Contract is the address of the price feed smart contract. |






<a name="osmosis.meshsecurity.v1beta1.MsgSetPriceFeedContractResponse"></a>

### MsgSetPriceFeedContractResponse
MsgSetPriceFeedContractResponse returns result data.






<a name="osmosis.meshsecurity.v1beta1.MsgSetVirtualStakingMaxCap"></a>

### MsgSetVirtualStakingMaxCap
MsgSetVirtualStakingMaxCap creates or updates a maximum cap limit for virtual
staking coins to the given contract.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | Authority is the address that controls the module (defaults to x/gov unless overwritten). |
| `contract` | [string](#string) |  | Contract is the address of the smart contract that is given permission do virtual staking which includes minting and burning staking tokens. |
| `max_cap` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | MaxCap is the limit up this the virtual tokens can be minted. |






<a name="osmosis.meshsecurity.v1beta1.MsgSetVirtualStakingMaxCapResponse"></a>

### MsgSetVirtualStakingMaxCapResponse
MsgSetVirtualStakingMaxCap returns result data.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.meshsecurity.v1beta1.Msg"></a>

### Msg
Msg defines the wasm Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `SetVirtualStakingMaxCap` | [MsgSetVirtualStakingMaxCap](#osmosis.meshsecurity.v1beta1.MsgSetVirtualStakingMaxCap) | [MsgSetVirtualStakingMaxCapResponse](#osmosis.meshsecurity.v1beta1.MsgSetVirtualStakingMaxCapResponse) | SetVirtualStakingMaxCap creates or updates a maximum cap limit for virtual staking coins | |
| `SetPriceFeedContract` | [MsgSetPriceFeedContract](#osmosis.meshsecurity.v1beta1.MsgSetPriceFeedContract) | [MsgSetPriceFeedContractResponse](#osmosis.meshsecurity.v1beta1.MsgSetPriceFeedContractResponse) | SetPriceFeedContract sets the price feed contract to the chain to trigger handle epoch task | |

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

