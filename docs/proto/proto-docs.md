<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [osmosis/meshsecurityprovider/genesis.proto](#osmosis/meshsecurityprovider/genesis.proto)
    - [GenesisState](#osmosis.meshsecurityprovider.GenesisState)
    - [Params](#osmosis.meshsecurityprovider.Params)
  
- [osmosis/meshsecurityprovider/query.proto](#osmosis/meshsecurityprovider/query.proto)
    - [ParamsRequest](#osmosis.meshsecurityprovider.ParamsRequest)
    - [ParamsResponse](#osmosis.meshsecurityprovider.ParamsResponse)
  
    - [Query](#osmosis.meshsecurityprovider.Query)
  
- [osmosis/meshsecurityprovider/tx.proto](#osmosis/meshsecurityprovider/tx.proto)
    - [MsgBond](#osmosis.meshsecurityprovider.MsgBond)
    - [MsgBondResponse](#osmosis.meshsecurityprovider.MsgBondResponse)
    - [MsgTest](#osmosis.meshsecurityprovider.MsgTest)
    - [MsgTestResponse](#osmosis.meshsecurityprovider.MsgTestResponse)
    - [MsgUpdateParams](#osmosis.meshsecurityprovider.MsgUpdateParams)
    - [MsgUpdateParamsResponse](#osmosis.meshsecurityprovider.MsgUpdateParamsResponse)
  
    - [Msg](#osmosis.meshsecurityprovider.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="osmosis/meshsecurityprovider/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurityprovider/genesis.proto



<a name="osmosis.meshsecurityprovider.GenesisState"></a>

### GenesisState
GenesisState defines the meshsecurityprovider module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.meshsecurityprovider.Params) |  | params is the container of meshsecurityprovider parameters. |






<a name="osmosis.meshsecurityprovider.Params"></a>

### Params



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `vault_address` | [string](#string) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="osmosis/meshsecurityprovider/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurityprovider/query.proto



<a name="osmosis.meshsecurityprovider.ParamsRequest"></a>

### ParamsRequest
=============================== Params






<a name="osmosis.meshsecurityprovider.ParamsResponse"></a>

### ParamsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#osmosis.meshsecurityprovider.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.meshsecurityprovider.Query"></a>

### Query


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [ParamsRequest](#osmosis.meshsecurityprovider.ParamsRequest) | [ParamsResponse](#osmosis.meshsecurityprovider.ParamsResponse) |  | GET|/osmosis/meshsecurityprovider/Params|

 <!-- end services -->



<a name="osmosis/meshsecurityprovider/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## osmosis/meshsecurityprovider/tx.proto



<a name="osmosis.meshsecurityprovider.MsgBond"></a>

### MsgBond
MsgBond defines a message for bonding to vault contract.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `delegator_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="osmosis.meshsecurityprovider.MsgBondResponse"></a>

### MsgBondResponse
MsgBondResponse defines the Msg/Bond response type.






<a name="osmosis.meshsecurityprovider.MsgTest"></a>

### MsgTest
===================== MsgTest


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `sender` | [string](#string) |  |  |






<a name="osmosis.meshsecurityprovider.MsgTestResponse"></a>

### MsgTestResponse







<a name="osmosis.meshsecurityprovider.MsgUpdateParams"></a>

### MsgUpdateParams
MsgUpdateParams is the Msg/UpdateParams request type.

Since: cosmos-sdk 0.47


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `authority` | [string](#string) |  | authority is the address that controls the module (defaults to x/gov unless overwritten). |
| `params` | [Params](#osmosis.meshsecurityprovider.Params) |  | params defines the x/staking parameters to update.

NOTE: All parameters must be supplied. |






<a name="osmosis.meshsecurityprovider.MsgUpdateParamsResponse"></a>

### MsgUpdateParamsResponse
MsgUpdateParamsResponse defines the response structure for executing a
MsgUpdateParams message.

Since: cosmos-sdk 0.47





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="osmosis.meshsecurityprovider.Msg"></a>

### Msg


| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Test` | [MsgTest](#osmosis.meshsecurityprovider.MsgTest) | [MsgTestResponse](#osmosis.meshsecurityprovider.MsgTestResponse) |  | |
| `UpdateParams` | [MsgUpdateParams](#osmosis.meshsecurityprovider.MsgUpdateParams) | [MsgUpdateParamsResponse](#osmosis.meshsecurityprovider.MsgUpdateParamsResponse) | UpdateParams defines an operation for updating the module's parameters. Since: cosmos-sdk 0.47 | |
| `Bond` | [MsgBond](#osmosis.meshsecurityprovider.MsgBond) | [MsgBondResponse](#osmosis.meshsecurityprovider.MsgBondResponse) | Bond defines an operation for bonding to vault contract. | |

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

