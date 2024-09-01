package types

const (
	EventTypeUnbond                     = "vault_unbond"
	EventTypeBond                       = "vault_bond"
	EventTypeUnstake                    = "native_unstake"
	EventTypeSubmitConsumerDoubleVoting = "submit_consumer_double_voting"
	EventTypeSubmitConsumerMisbehaviour = "submit_consumer_misbehaviour"
)

const (
	AttributeValueCategory      = ModuleName
	AttributeKeyDelegator       = "delegator"
	AttributeKeyValidator       = "validator"
	AttributeKeyContractAddress = "proxry_staking_contract_address"
	AttributeConsumerChainID    = "consunmer_chain_id"
)
