package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// ModuleName defines the module name.
	ModuleName = "meshsecurity"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName
)

var (
	// ParamsKey is the prefix for the module parameters
	ParamsKey                     = []byte{0x1}
	MaxCapLimitKeyPrefix          = []byte{0x2}
	TotalDelegatedAmountKeyPrefix = []byte{0x3}
)

// BuildMaxCapLimitKey build max cap limit store key
func BuildMaxCapLimitKey(contractAddr sdk.AccAddress) []byte {
	return append(TotalDelegatedAmountKeyPrefix, contractAddr.Bytes()...)
}
