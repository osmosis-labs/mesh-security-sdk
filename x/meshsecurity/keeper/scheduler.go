package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func (k Keeper) GetRebalanceGasLimit(ctx sdk.Context) sdk.Gas {
	return 0 // todo: impl
}

func (k Keeper) GetRebalanceEpochLength(ctx sdk.Context) int64 {
	return 0 // todo: impl
}

func (k Keeper) ScheduleRebalance(ctx sdk.Context, tp types.SchedulerType, contract sdk.AccAddress, execBlockHeight uint64, repeat bool) {
	if execBlockHeight < uint64(ctx.BlockHeight()) { // sanity check
		panic(fmt.Sprintf("can not schedule for past block: %d", execBlockHeight))
	}
	store := ctx.KVStore(k.storeKey)
	storeKey := types.BuildSchedulerContractKey(tp, execBlockHeight, contract)

	if !repeat { // ensure that we do not overwrite a repeating scheduled event
		if bz := store.Get(storeKey); bz != nil {
			repeat = isRepeatFlag(bz)
		}
	}
	store.Set(storeKey, []byte{toByte(repeat)})
	types.EmitSchedulerRegisteredEvent(ctx, contract, execBlockHeight, repeat)
}

// callback interface to execute the rebalance action
type rebalancer func(ctx sdk.Context, addr sdk.AccAddress) error

func (k Keeper) IterateScheduledExecutions(ctx sdk.Context, tp types.SchedulerType, height uint64, cb func(addr sdk.AccAddress, repeat bool) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.BuildSchedulerKeyPrefix(tp, height))
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		repeat := isRepeatFlag(bz)
		// cb returns true to stop early
		if cb(iter.Key(), repeat) {
			return
		}
	}
}

func (k Keeper) ExecScheduled(pCtx sdk.Context, tp types.SchedulerType, cb rebalancer) {
	// iterator is most gas cost-efficient currently
	k.IterateScheduledExecutions(pCtx, tp, uint64(pCtx.BlockHeight()), func(contract sdk.AccAddress, repeat bool) bool {
		gasLimit := k.GetRebalanceGasLimit(pCtx)

		cachedCtx, done := pCtx.CacheContext()
		gasMeter := sdk.NewGasMeter(gasLimit)
		cachedCtx = cachedCtx.WithGasMeter(gasMeter)
		err := safeExec(func() error { return cb(cachedCtx, contract) })
		if err == nil {
			done()
		}

		types.EmitSchedulerExecutionEvent(pCtx, contract, err)
		pCtx.Logger().Info("Rebalance scheduler executed", "gas_used", gasMeter.GasConsumed(), "gas_limit", gasLimit)
		if repeat {
			// re-schedule
			epochLength := k.GetRebalanceEpochLength(pCtx)
			nextExecBlock := uint64(pCtx.BlockHeight() + epochLength)
			k.ScheduleRebalance(pCtx, tp, contract, nextExecBlock, true)
		}
		return false
	})
}

func safeExec(cb func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.ErrPanic.Wrapf("execution: %v", r)
		}
	}()
	return cb()
}

func isRepeatFlag(bz []byte) bool {
	return bytes.Equal(bz, []byte{toByte(true)})
}

func toByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
