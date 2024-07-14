package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "meshsecurity"
	RouterKey  = ModuleName
)

var (
	ChainToChannelBytePrefix       = []byte{0x1}
	ChannelToChainBytePrefix       = []byte{0x2}
	InitChainHeightBytePrefix      = []byte{0x3}
	InitTimeoutTimestampBytePrefix = []byte{0x4}
	ChainToClientBytePrefix        = []byte{0x5}
	ConsumerCommissionRatePrefix   = []byte{0x6}
	ParamsKey                      = []byte{0x7}
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

type ProviderConsAddress struct {
	Address sdk.ConsAddress
}

func NewProviderConsAddress(addr sdk.ConsAddress) ProviderConsAddress {
	return ProviderConsAddress{
		Address: addr,
	}
}
