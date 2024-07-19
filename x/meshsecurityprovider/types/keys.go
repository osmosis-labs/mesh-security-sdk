package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name.
	ModuleName = "meshsecurity-provider"
	// RouterKey is the message route
	RouterKey = ModuleName
	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
	// MemStoreKey defines the in-memory store key
	MemStoreKey = "memory:meshsecurity-provider"
)

var (
	ChainToChannelBytePrefix       = []byte{0x10}
	ChannelToChainBytePrefix       = []byte{0x2}
	InitChainHeightBytePrefix      = []byte{0x3}
	InitTimeoutTimestampBytePrefix = []byte{0x4}
	ChainToClientBytePrefix        = []byte{0x5}
	ConsumerCommissionRatePrefix   = []byte{0x6}
	ParamsKey                      = []byte{0x7}
	DepositorsKeyPrefix            = []byte{0x8}
	ContractWithNativeDenomPrefix  = []byte{0x9}
	IntermediaryKeyPrefix          = []byte{0x11}
)

func ChainToChannelKey(chainID string) []byte {
	return append(ChainToChannelBytePrefix, []byte(chainID)...)
}

func ChannelToChainKey(channelID string) []byte {
	return append(ChannelToChainBytePrefix, []byte(channelID)...)
}

func InitChainHeightKey(chainID string) []byte {
	return append(InitChainHeightBytePrefix, []byte(chainID)...)
}

func InitTimeoutTimestampKey(chainID string) []byte {
	return append(InitTimeoutTimestampBytePrefix, []byte(chainID)...)
}

func ChainToClientKey(chainID string) []byte {
	return append(ChainToClientBytePrefix, []byte(chainID)...)
}

func ConsumerCommissionRateKey(chainID string, providerAddr ProviderConsAddress) []byte {
	return append(append(ConsumerCommissionRatePrefix, []byte(chainID)...), providerAddr.Address...)
}

func DepositorsKey(del string) []byte {
	return append(DepositorsKeyPrefix, []byte(del)...)
}

func ContractWithNativeDenomKey(denom string) []byte {
	return append(ContractWithNativeDenomPrefix, []byte(denom)...)
}
func IntermediaryKey(denom string) []byte {
	return append(IntermediaryKeyPrefix, []byte(denom)...)
}

type ProviderConsAddress struct {
	Address sdk.ConsAddress
}

func NewProviderConsAddress(addr sdk.ConsAddress) ProviderConsAddress {
	return ProviderConsAddress{
		Address: addr,
	}
}
