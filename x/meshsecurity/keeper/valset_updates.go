package keeper

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// ScheduleBonded store a validator update to bonded status for the valset update report
func (k Keeper) ScheduleBonded(ctx sdk.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorBonded, addr)
}

// ScheduleUnbonded store a validator update to unbonded status for the valset update report
func (k Keeper) ScheduleUnbonded(ctx sdk.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorUnbonded, addr)
}

// ScheduleJailed store a validator update to jailed status for the valset update report
func (k Keeper) ScheduleJailed(ctx sdk.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorJailed, addr)
}

// ScheduleTombstoned store a validator update to tombstoned status for the valset update report
func (k Keeper) ScheduleTombstoned(ctx sdk.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorTombstoned, addr)
}

// ScheduleUnjailed store a validator update to unjailed status for the valset update report
func (k Keeper) ScheduleUnjailed(ctx sdk.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorUnjailed, addr)
}

// ScheduleModified store a validator metadata update for the valset update report
func (k Keeper) ScheduleModified(ctx sdk.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorModified, addr)
}

// instead of sync calls to the contracts for the different kind of valset changes in a block, we store them in the mem db
// and async send to all registered contracts in the end blocker
func (k Keeper) sendAsync(ctx sdk.Context, op types.PipedValsetOperation, valAddr sdk.ValAddress) error {
	ModuleLogger(ctx).Debug("storing for async update", "operation", int(op), "val", valAddr.String())
	ctx.KVStore(k.memKey).Set(types.BuildPipedValsetOpKey(op, valAddr), []byte{})
	// and schedule an update callback for all registered contracts
	var innerErr error
	k.IterateMaxCapLimit(ctx, func(contractAddr sdk.AccAddress, m math.Int) bool {
		if m.GT(math.ZeroInt()) {
			innerErr = k.ScheduleOneShotTask(ctx, types.SchedulerTaskValsetUpdate, contractAddr, uint64(ctx.BlockHeight()))
			if innerErr != nil {
				return true
			}
		}
		return false
	})
	return innerErr
}

// ValsetUpdateReport aggregate all stored changes of the current block. Should be called by an end-blocker.
// The events reported are categorized by type and not time. Conflicting events as Bonded/ Unbonded
// are not supposed to happen within the same block
func (k Keeper) ValsetUpdateReport(ctx sdk.Context) (contract.ValsetUpdate, error) {
	var innerErr error
	appendValidator := func(set *[]wasmvmtypes.Validator, valAddr sdk.ValAddress) bool {
		val, ok := k.staking.GetValidator(ctx, valAddr)
		if !ok {
			innerErr = types.ErrUnknown.Wrapf("validator %s", valAddr)
			return true
		}
		*set = append(*set, ConvertSdkValidatorToWasm(val))
		return false
	}
	r := contract.ValsetUpdate{ // init with empty slices for contract that does not handle null or omitted fields
		Additions:  make([]contract.Validator, 0),
		Removals:   make([]contract.ValidatorAddr, 0),
		Updated:    make([]contract.Validator, 0),
		Jailed:     make([]contract.ValidatorAddr, 0),
		Unjailed:   make([]contract.ValidatorAddr, 0),
		Tombstoned: make([]contract.ValidatorAddr, 0),
	}
	err := k.iteratePipedValsetOperations(ctx, func(valAddr sdk.ValAddress, op types.PipedValsetOperation) bool {
		switch op {
		case types.ValidatorBonded:
			return appendValidator(&r.Additions, valAddr)
		case types.ValidatorUnbonded:
			r.Removals = append(r.Removals, valAddr.String())
		case types.ValidatorJailed:
			r.Jailed = append(r.Jailed, valAddr.String())
		case types.ValidatorTombstoned:
			r.Tombstoned = append(r.Tombstoned, valAddr.String())
		case types.ValidatorUnjailed:
			r.Unjailed = append(r.Unjailed, valAddr.String())
		case types.ValidatorModified:
			return appendValidator(&r.Updated, valAddr)
		default:
			innerErr = types.ErrInvalid.Wrapf("undefined operation type %X", op)
			return true
		}
		return false
	})
	if err != nil {
		return r, err
	}
	return r, innerErr
}

// ClearPipedValsetOperations delete all entries from the temporary store that contains the valset updates.
func (k Keeper) ClearPipedValsetOperations(ctx sdk.Context) {
	var keys [][]byte
	pStore := prefix.NewStore(ctx.KVStore(k.memKey), types.PipedValsetPrefix)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}
	_ = iter.Close()
	for _, k := range keys {
		pStore.Delete(k)
	}
}

// iterate through all stored valset updates. Due to the storage key, there are no contract duplicates within an operation type.
func (k Keeper) iteratePipedValsetOperations(ctx sdk.Context, cb func(valAddress sdk.ValAddress, op types.PipedValsetOperation) bool) error {
	pStore := prefix.NewStore(ctx.KVStore(k.memKey), types.PipedValsetPrefix)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		addrLen := key[0]
		addr, op := key[1:addrLen+1], key[addrLen+1]
		if cb(addr, types.PipedValsetOperation(op)) {
			break
		}
	}
	return iter.Close()
}

// ConvertSdkValidatorToWasm helper method
func ConvertSdkValidatorToWasm(v stakingtypes.Validator) wasmvmtypes.Validator {
	return wasmvmtypes.Validator{
		Address:       v.OperatorAddress,
		Commission:    v.Commission.Rate.String(),
		MaxCommission: v.Commission.MaxRate.String(),
		MaxChangeRate: v.Commission.MaxChangeRate.String(),
	}
}
