package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
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
	SchedulerKeyPrefix            = []byte{0x4}
)

// BuildMaxCapLimitKey build max cap limit store key
func BuildMaxCapLimitKey(contractAddr sdk.AccAddress) []byte {
	return append(MaxCapLimitKeyPrefix, contractAddr.Bytes()...)
}

// BuildTotalDelegatedAmountKey build delegated amount store key for given contract
func BuildTotalDelegatedAmountKey(contractAddr sdk.AccAddress) []byte {
	return append(TotalDelegatedAmountKeyPrefix, contractAddr.Bytes()...)
}

func BuildSchedulerTypeKeyPrefix(tp SchedulerTaskType) ([]byte, error) {
	if tp == SchedulerTaskUndefined {
		return nil, ErrInvalid.Wrapf("scheduler type: %x", tp)
	}
	return append(SchedulerKeyPrefix, byte(tp)), nil
}

// BuildSchedulerHeightKeyPrefix build store key prefix
func BuildSchedulerHeightKeyPrefix(tp SchedulerTaskType, blockHeight uint64) ([]byte, error) {
	prefix, err := BuildSchedulerTypeKeyPrefix(tp)
	if err != nil {
		return nil, err
	}
	return append(prefix, sdk.Uint64ToBigEndian(blockHeight)...), nil
}

// BuildSchedulerContractKey build store key
func BuildSchedulerContractKey(tp SchedulerTaskType, blockHeight uint64, contractAddr sdk.AccAddress) ([]byte, error) {
	prefix, err := BuildSchedulerHeightKeyPrefix(tp, blockHeight)
	if err != nil {
		return nil, err
	}
	return append(prefix, contractAddr.Bytes()...), nil
}
