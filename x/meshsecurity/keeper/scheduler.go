package keeper

import (
	"bytes"
	"math"

	"github.com/golang/protobuf/proto"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// ScheduleRegularRebalanceTask schedule a rebalance task for the given virtual staking contract using params defined epoch length
func (k Keeper) ScheduleRegularRebalanceTask(ctx sdk.Context, contract sdk.AccAddress) error {
	if !k.wasm.HasContractInfo(ctx, contract) {
		return types.ErrUnknown.Wrapf("contract: %s", contract.String())
	}
	epochLength := k.GetRebalanceEpochLength(ctx)
	nextExecBlock := uint64(ctx.BlockHeight()) + epochLength
	return k.ScheduleTask(ctx, types.SchedulerTaskRebalance, contract, nextExecBlock, true, nil)
}

// HasScheduledTask returns true if the contract has a task scheduled of the given type and repeat setting
func (k Keeper) HasScheduledTask(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, repeat bool) bool {
	var result bool
	err := k.iterateScheduledContractTasks(ctx, tp, contract, math.MaxUint, func(_ uint64, isRepeat bool) bool {
		result = repeat == isRepeat
		return result
	})
	return err == nil && result // we can ignore the unknown task type error and return false instead
}

// GetNextScheduledTaskHeight returns height for task to execute
func (k Keeper) GetNextScheduledTaskHeight(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress) (height uint64, found bool) {
	err := k.iterateScheduledContractTasks(ctx, tp, contract, math.MaxUint, func(atHeight uint64, _ bool) bool {
		height = atHeight
		found = true
		return true
	})
	found = err == nil && found // we can ignore the unknown task type error and return false instead
	return
}

func (k Keeper) getScheduledTaskAt(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, height uint64) (repeat, exists bool) {
	key, err := types.BuildSchedulerContractKey(tp, height, contract)
	if err != nil {
		return false, false
	}
	bz := ctx.KVStore(k.storeKey).Get(key)
	return isRepeat(bz), bz != nil
}

// ScheduleTask register a new task to be executed at given block height
func (k Keeper) ScheduleTask(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, execBlockHeight uint64, repeat bool, payload proto.Message) error {
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
			repeat = isRepeat(bz)
		}
	}
	store.Set(storeKey, []byte{toByte(repeat)})
	types.EmitSchedulerRegisteredEvent(ctx, contract, execBlockHeight, repeat)
	return nil
}

// IterateScheduledTasks iterate of all scheduled task executions for the given type up to given block height (included)
func (k Keeper) IterateScheduledTasks(ctx sdk.Context, tp types.SchedulerTaskType, maxHeight uint64, cb func(addr sdk.AccAddress, height uint64, repeat bool) bool) error {
	keyPrefix, err := types.BuildSchedulerTypeKeyPrefix(tp)
	if err != nil {
		return err
	}
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), keyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		key := iter.Key()
		scheduledHeight := sdk.BigEndianToUint64(key[0:8])
		if scheduledHeight > maxHeight || // abort for future heights
			cb(key[8:], scheduledHeight, isRepeat(iter.Value())) {
			return nil
		}
	}
	return nil
}

// DeleteAllScheduledTasks deletes all tasks of given type for the contract.
func (k Keeper) DeleteAllScheduledTasks(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress) error {
	var innerErr error
	err := k.iterateScheduledContractTasks(ctx, tp, contract, math.MaxUint, func(height uint64, _ bool) bool {
		if err := k.deleteScheduledTask(ctx, tp, contract, height); err != nil {
			innerErr = errorsmod.Wrapf(err, "remove task height: %d", height)
			return true
		}
		return false
	})
	if err != nil {
		return errorsmod.Wrap(err, "remove all scheduled tasks")
	}
	return innerErr
}

// deleteScheduledTask removes scheduled task from store
func (k Keeper) deleteScheduledTask(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, execBlockHeight uint64) error {
	storeKey, err := types.BuildSchedulerContractKey(tp, execBlockHeight, contract)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(storeKey)
	return nil
}

// Iterate through all scheduled tasks for given task type and contract
//
// The implementation not super efficient but as there should be only a small set of contracts and tasks lets not do a secondary index, now
func (k Keeper) iterateScheduledContractTasks(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, maxHeight uint64, cb func(height uint64, repeat bool) bool) error {
	return k.IterateScheduledTasks(ctx, tp, maxHeight, func(addr sdk.AccAddress, height uint64, repeat bool) bool {
		return addr.Equals(contract) && cb(height, repeat)
	})
}

// ExecResult are the results of a task execution
type ExecResult struct {
	Contract      sdk.AccAddress
	ExecErr       error
	RescheduleErr error
	DeleteTaskErr error
	GasUsed       sdk.Gas
	GasLimit      sdk.Gas
	NextRunHeight uint64
}

// callback interface to execute a scheduled task
type executor func(ctx sdk.Context, addr sdk.AccAddress) error

// ExecScheduledTasks execute scheduled task at current height
// The executor function is called within the scope of a new cached store. Any failure on execution
// reverts the state of this sub call. Rescheduling or other state changes due to the scheduler provisioning
// are not affected.
// The result type contains more details information of execution or provisioning errors.
// The given epoch length is used for re-scheduling the task, when set on the task and value >0
func (k Keeper) ExecScheduledTasks(pCtx sdk.Context, tp types.SchedulerTaskType, epochLength uint64, cb executor) ([]ExecResult, error) {
	var allResults []ExecResult
	currentHeight := uint64(pCtx.BlockHeight())
	// iterator is most gas cost-efficient currently
	err := k.IterateScheduledTasks(pCtx, tp, currentHeight, func(contract sdk.AccAddress, scheduledHeight uint64, repeat bool) bool {
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

		if repeat && epochLength != 0 {
			// re-schedule
			nextExecBlock := uint64(pCtx.BlockHeight()) + epochLength
			result.NextRunHeight = nextExecBlock
			if err := k.ScheduleTask(pCtx, tp, contract, nextExecBlock, repeat, nil); err != nil {
				result.RescheduleErr = err
			}
		}
		if err := k.deleteScheduledTask(pCtx, tp, contract, scheduledHeight); err != nil {
			result.DeleteTaskErr = err
		}
		allResults = append(allResults, result)
		return false
	})
	return allResults, err
}

// execute callback with panics recovered
func safeExec(cb func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.ErrPanic.Wrapf("execution: %v", r)
		}
	}()
	return cb()
}

// helper that returns true when given bytes are equal to the "repeat" flag byte representation
func isRepeat(bz []byte) bool {
	return bytes.Equal(bz, []byte{toByte(true)})
}

// converts boolean to byte representation
func toByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
