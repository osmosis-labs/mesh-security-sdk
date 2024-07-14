package types

import (
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/osmosis-labs/mesh-security-sdk/x/types"
)

const (
	// ModuleName defines the module name.
	ModuleName = "meshsecurity"

	// ConsumerPortID is the default port id the consumer module binds to
	ConsumerPortID = "consumer"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "memory:meshsecurity"

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

	PipedValsetPrefix      = []byte{0x5}
	ProviderChannelByteKey = []byte{0x6}
)

// BuildMaxCapLimitKey build max cap limit store key
func BuildMaxCapLimitKey(contractAddr sdk.AccAddress) []byte {
	return append(MaxCapLimitKeyPrefix, contractAddr.Bytes()...)
}

// BuildTotalDelegatedAmountKey build delegated amount store key for given contract
func BuildTotalDelegatedAmountKey(contractAddr sdk.AccAddress) []byte {
	return append(TotalDelegatedAmountKeyPrefix, contractAddr.Bytes()...)
}

// BuildSchedulerTypeKeyPrefix internal scheduler store key
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

// BuildPipedValsetOpKey build store key for the temporary valset operation store
func BuildPipedValsetOpKey(op types.PipedValsetOperation, val sdk.ValAddress, slashInfo *types.SlashInfo) []byte {
	if op == types.PipedValsetOperation_UNSPECIFIED {
		panic("empty operation")
	}
	pn, an := len(PipedValsetPrefix), len(val)
	sn := 0
	if op == types.PipedValsetOperation_VALIDATOR_SLASHED {
		if slashInfo == nil {
			panic("slash info is nil")
		}
		sn = 8 + 8 + 1 + len(slashInfo.TotalSlashAmount) + len(slashInfo.SlashFraction) // 8 for height, 8 for power, +1 for total amount length
	}
	r := make([]byte, pn+an+sn+1+1) // +1 for address prefix, +1 for op
	copy(r, PipedValsetPrefix)
	copy(r[pn:], address.MustLengthPrefix(val))
	r[pn+an+1] = byte(op)
	if op == types.PipedValsetOperation_VALIDATOR_SLASHED {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(slashInfo.InfractionHeight))
		copy(r[pn+an+1+1:], b)
		binary.BigEndian.PutUint64(b, uint64(slashInfo.Power))
		copy(r[pn+an+1+1+8:], b)
		tn := len(slashInfo.TotalSlashAmount)
		r[pn+an+1+1+8+8] = byte(tn)
		copy(r[pn+an+1+1+8+8+1:], slashInfo.TotalSlashAmount)
		copy(r[pn+an+1+1+8+8+1+tn:], slashInfo.SlashFraction)
	}
	return r
}

// ProviderChannelKey returns the key for storing channelID of the provider chain
func ProviderChannelKey() []byte {
	return ProviderChannelByteKey
}
