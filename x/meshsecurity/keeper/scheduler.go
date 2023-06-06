package keeper

import (
	"bytes"

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

func (k Keeper) deleteScheduledTask(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, execBlockHeight uint64) error {
	storeKey, err := types.BuildSchedulerContractKey(tp, execBlockHeight, contract)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(storeKey)
	return nil
}

func (k Keeper) ScheduleTask(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, execBlockHeight uint64, repeat bool) error {
	if execBlockHeight < uint64(ctx.BlockHeight()) { // sanity check
		return types.ErrInvalid.Wrapf("can not schedule for past block: %d", execBlockHeight)
	}
	storeKey, err := types.BuildSchedulerContractKey(tp, execBlockHeight, contract)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	if !repeat { // ensure that we do not overwrite a repeating scheduled event
		if bz := store.Get(storeKey); bz != nil {
			repeat = isRepeatFlag(bz)
		}
	}
	store.Set(storeKey, []byte{toByte(repeat)})
	types.EmitSchedulerRegisteredEvent(ctx, contract, execBlockHeight, repeat)
	return nil
}

// callback interface to execute the rebalance action
type rebalancer func(ctx sdk.Context, addr sdk.AccAddress) error

func (k Keeper) IterateScheduledTasks(ctx sdk.Context, tp types.SchedulerTaskType, height uint64, cb func(addr sdk.AccAddress, repeat bool) bool) error {
	keyPrefix, err := types.BuildSchedulerKeyPrefix(tp, height)
	if err != nil {
		return err
	}
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), keyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		repeat := isRepeatFlag(bz)
		// cb returns true to stop early
		if cb(iter.Key(), repeat) {
			return nil
		}
	}
	return nil
}

type ExecResult struct {
	Contract      sdk.AccAddress
	ExecErr       error
	RescheduleErr error
	DeleteTaskErr error
	GasUsed       sdk.Gas
	GasLimit      sdk.Gas
}

// ExecScheduledTasks execute scheduled task at current height
func (k Keeper) ExecScheduledTasks(pCtx sdk.Context, tp types.SchedulerTaskType, cb rebalancer) ([]ExecResult, error) {
	var allResults []ExecResult
	currentHeight := uint64(pCtx.BlockHeight())
	// iterator is most gas cost-efficient currently
	err := k.IterateScheduledTasks(pCtx, tp, currentHeight, func(contract sdk.AccAddress, repeat bool) bool {
		gasLimit := k.GetRebalanceGasLimit(pCtx)
		cachedCtx, done := pCtx.CacheContext()
		gasMeter := sdk.NewGasMeter(gasLimit)
		cachedCtx = cachedCtx.WithGasMeter(gasMeter)
		result := ExecResult{Contract: contract, GasLimit: gasLimit}
		err := safeExec(func() error { return cb(cachedCtx, contract) })
		if err != nil {
			result.ExecErr = err
		} else {
			done()
			ModuleLogger(pCtx).
				Info("Scheduler executed successfully", "gas_used", gasMeter.GasConsumed(),
					"gas_limit", gasLimit, "contract", contract.String(), "task_type", tp)
		}
		result.GasUsed = gasMeter.GasConsumed()
		types.EmitSchedulerExecutionEvent(pCtx, contract, err)

		if repeat {
			// re-schedule
			epochLength := k.GetRebalanceEpochLength(pCtx)
			nextExecBlock := uint64(pCtx.BlockHeight() + epochLength)
			if err := k.ScheduleTask(pCtx, tp, contract, nextExecBlock, repeat); err != nil {
				result.RescheduleErr = err
			}
		}
		if err := k.deleteScheduledTask(pCtx, tp, contract, currentHeight); err != nil {
			result.DeleteTaskErr = err
		}
		allResults = append(allResults, result)
		return false
	})
	return allResults, err
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
