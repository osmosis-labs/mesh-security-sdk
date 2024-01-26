package keeper

import (
	"context"

	"cosmossdk.io/math"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"

	outmessage "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

// ScheduleBonded store a validator update to bonded status for the valset update report
func (k Keeper) ScheduleBonded(ctx context.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorBonded, addr, nil)
}

// ScheduleUnbonded store a validator update to unbonded status for the valset update report
func (k Keeper) ScheduleUnbonded(ctx context.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorUnbonded, addr, nil)
}

// ScheduleSlashed store a validator slash event / data for the valset update report
func (k Keeper) ScheduleSlashed(ctx context.Context, addr sdk.ValAddress, power int64, height int64, totalSlashAmount math.Int, slashRatio math.LegacyDec) error {
	var slashInfo = &types.SlashInfo{
		Power:            power,
		InfractionHeight: height,
		TotalSlashAmount: totalSlashAmount.String(),
		SlashFraction:    slashRatio.String(),
	}
	return k.sendAsync(ctx, types.ValidatorSlashed, addr, slashInfo)
}

// ScheduleJailed store a validator update to jailed status for the valset update report
func (k Keeper) ScheduleJailed(ctx context.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorJailed, addr, nil)
}

// ScheduleTombstoned store a validator update to tombstoned status for the valset update report
func (k Keeper) ScheduleTombstoned(ctx context.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorTombstoned, addr, nil)
}

// ScheduleUnjailed store a validator update to unjailed status for the valset update report
func (k Keeper) ScheduleUnjailed(ctx context.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorUnjailed, addr, nil)
}

// ScheduleModified store a validator metadata update for the valset update report
func (k Keeper) ScheduleModified(ctx context.Context, addr sdk.ValAddress) error {
	return k.sendAsync(ctx, types.ValidatorModified, addr, nil)
}

// instead of sync calls to the contracts for the different kind of valset changes in a block, we store them in the mem db
// and async send to all registered contracts in the end blocker
func (k Keeper) sendAsync(ctx context.Context, op types.PipedValsetOperation, valAddr sdk.ValAddress, slashInfo *types.SlashInfo) error {
	goCtx := sdk.UnwrapSDKContext(ctx)

	ModuleLogger(ctx).Debug("storing for async update", "operation", int(op), "val", valAddr.String())
	goCtx.KVStore(k.memKey).Set(types.BuildPipedValsetOpKey(op, valAddr, slashInfo), []byte{})
	// and schedule an update callback for all registered contracts
	var innerErr error
	k.IterateMaxCapLimit(goCtx, func(contractAddr sdk.AccAddress, m math.Int) bool {
		if m.GT(math.ZeroInt()) {
			innerErr = k.ScheduleOneShotTask(goCtx, types.SchedulerTaskValsetUpdate, contractAddr, uint64(goCtx.BlockHeight()))
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
func (k Keeper) ValsetUpdateReport(ctx context.Context) (contract.ValsetUpdate, error) {
	goCtx := sdk.UnwrapSDKContext(ctx)
	var innerErr error
	appendValidator := func(set *[]wasmvmtypes.Validator, valAddr sdk.ValAddress) bool {
		val, err := k.Staking.GetValidator(ctx, valAddr)
		if err != nil {
			innerErr = types.ErrUnknown.Wrapf("validator %s", valAddr)
			return true
		}
		*set = append(*set, ConvertSdkValidatorToWasm(val))
		return false
	}
	slashValidator := func(set *[]outmessage.ValidatorSlash, valAddr sdk.ValAddress, power int64, infractionHeight int64,
		infractionTime int64, slashAmount string, slashRatio string) bool {
		valSlash := outmessage.ValidatorSlash{
			ValidatorAddr:    valAddr.String(),
			Power:            power,
			InfractionHeight: infractionHeight,
			InfractionTime:   infractionTime,
			Height:           goCtx.BlockHeight(),
			Time:             goCtx.BlockTime().Unix(),
			SlashAmount:      slashAmount,
			SlashRatio:       slashRatio,
		}
		*set = append(*set, valSlash)
		return false
	}
	r := contract.ValsetUpdate{ // init with empty slices for contract that does not handle null or omitted fields
		Additions:  make([]contract.Validator, 0),
		Removals:   make([]contract.ValidatorAddr, 0),
		Updated:    make([]contract.Validator, 0),
		Jailed:     make([]contract.ValidatorAddr, 0),
		Unjailed:   make([]contract.ValidatorAddr, 0),
		Tombstoned: make([]contract.ValidatorAddr, 0),
		Slashed:    make([]contract.ValidatorSlash, 0),
	}
	err := k.iteratePipedValsetOperations(ctx, func(valAddr sdk.ValAddress, op types.PipedValsetOperation, slashInfo *types.SlashInfo) bool {
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
		case types.ValidatorSlashed:
			// TODO: Add / send the infraction time
			return slashValidator(&r.Slashed, valAddr, slashInfo.Power, slashInfo.InfractionHeight, 0,
				slashInfo.TotalSlashAmount, slashInfo.SlashFraction)
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
func (k Keeper) ClearPipedValsetOperations(ctx context.Context) {
	goCtx := sdk.UnwrapSDKContext(ctx)
	var keys [][]byte
	pStore := prefix.NewStore(goCtx.KVStore(k.memKey), types.PipedValsetPrefix)
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
func (k Keeper) iteratePipedValsetOperations(ctx context.Context, cb func(valAddress sdk.ValAddress, op types.PipedValsetOperation, slashInfo *types.SlashInfo) bool) error {
	goCtx := sdk.UnwrapSDKContext(ctx)
	pStore := prefix.NewStore(goCtx.KVStore(k.memKey), types.PipedValsetPrefix)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		addrLen := key[0]
		addr, op := key[1:addrLen+1], key[addrLen+1]
		var slashInfo *types.SlashInfo = nil
		if types.PipedValsetOperation(op) == types.ValidatorSlashed {
			if len(key) <= 1+int(addrLen)+1+8+8+1 {
				return types.ErrInvalid.Wrapf("invalid slash key length %d", len(key))
			}
			totalSlashAmountLen := key[addrLen+2+8+8]
			slashInfo = &types.SlashInfo{
				Power:            int64(sdk.BigEndianToUint64(key[addrLen+2 : addrLen+2+8])),
				InfractionHeight: int64(sdk.BigEndianToUint64(key[addrLen+2+8 : addrLen+2+8+8])),
				TotalSlashAmount: string(key[addrLen+2+8+8+1 : addrLen+2+8+8+1+totalSlashAmountLen]),
				SlashFraction:    string(key[addrLen+2+8+8+1+totalSlashAmountLen:]),
			}
		}
		if cb(addr, types.PipedValsetOperation(op), slashInfo) {
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
