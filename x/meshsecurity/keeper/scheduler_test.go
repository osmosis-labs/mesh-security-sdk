package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestExecuteScheduledTask(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))

	k.wasm = MockWasmKeeper{HasContractInfoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
		return contractAddress.Equals(myContract)
	}}

	var execCount int
	incrExec := func(t *testing.T) executor {
		return func(ctx sdk.Context, addr sdk.AccAddress) error {
			require.Equal(t, myContract, addr)
			execCount++
			return nil
		}
	}
	currentHeight := uint64(pCtx.BlockHeight())

	specs := map[string]struct {
		repeat         bool
		exec           func(t *testing.T) executor
		expErr         bool
		expRescheduled bool
		expResult      []ExecResult
	}{
		"exec and reschedule": {
			repeat:         true,
			exec:           incrExec,
			expRescheduled: true,
			expResult: []ExecResult{{
				Contract:      myContract,
				GasLimit:      500000,
				NextRunHeight: 1234667,
			}},
		},
		"exec and not reschedule": {
			repeat: false,
			exec:   incrExec,
			expResult: []ExecResult{{
				Contract: myContract,
				GasLimit: 500000,
			}},
		},
		"exec fails": {
			repeat: true,
			exec: func(t *testing.T) executor {
				return func(ctx sdk.Context, addr sdk.AccAddress) error {
					_ = incrExec(t)(ctx, addr)
					return types.ErrUnknown.Wrap("testing")
				}
			},
			expRescheduled: true,
			expResult: []ExecResult{{
				Contract:      myContract,
				GasLimit:      500000,
				ExecErr:       types.ErrUnknown,
				NextRunHeight: 1234667,
			}},
		},
		"exec panics": {
			repeat: true,
			exec: func(t *testing.T) executor {
				return func(ctx sdk.Context, addr sdk.AccAddress) error {
					_ = incrExec(t)(ctx, addr)
					panic("testing")
				}
			},
			expRescheduled: true,
			expResult: []ExecResult{{
				Contract:      myContract,
				GasLimit:      500000,
				ExecErr:       sdkerrors.ErrPanic,
				NextRunHeight: 1234667,
			}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			execCount = 0
			var gotErr error
			if spec.repeat {
				gotErr = k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContract, currentHeight)
			} else {
				gotErr = k.ScheduleOneShotTask(ctx, types.SchedulerTaskRebalance, myContract, currentHeight)
			}
			require.NoError(t, gotErr)
			// when
			gotRes, gotErr := k.ExecScheduledTasks(ctx, types.SchedulerTaskRebalance, 100, spec.exec(t))
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Len(t, gotRes, len(spec.expResult))
			for i, exp := range spec.expResult {
				got := gotRes[i]
				require.ErrorIs(t, exp.ExecErr, got.ExecErr)
				got.ExecErr = exp.ExecErr // now that error types are checked, clear details
				assert.Equal(t, exp, got)
			}
			// and callback executed
			assert.Equal(t, 1, execCount)
			// and new execution scheduled
			repeat, exists := k.getScheduledTaskAt(ctx, types.SchedulerTaskRebalance, myContract, gotRes[0].NextRunHeight)
			assert.Equal(t, spec.expRescheduled, exists)
			assert.Equal(t, spec.repeat, repeat)
			// and old entry removed
			_, exists = k.getScheduledTaskAt(ctx, types.SchedulerTaskRebalance, myContract, currentHeight)
			assert.False(t, exists)
		})
	}
}

func TestScheduleTask(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))
	myOtherContractWithScheduledTask := sdk.AccAddress(rand.Bytes(32))
	currentHeight := uint64(pCtx.BlockHeight())

	// set a scheduler to overwrite
	err := k.ScheduleRepeatingTask(pCtx, types.SchedulerTaskRebalance, myOtherContractWithScheduledTask, currentHeight+1)
	require.NoError(t, err)

	specs := map[string]struct {
		taskType  types.SchedulerTaskType
		contract  sdk.AccAddress
		height    uint64
		repeat    bool
		expRepeat bool
		expErr    bool
	}{
		"all good - single exec": {
			contract: myContract,
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight + 1,
		},
		"all good - repeat": {
			contract:  myContract,
			taskType:  types.SchedulerTaskRebalance,
			height:    currentHeight + 1,
			repeat:    true,
			expRepeat: true,
		},
		"overwrite existing scheduler - preserve repeat": {
			contract:  myOtherContractWithScheduledTask,
			taskType:  types.SchedulerTaskRebalance,
			height:    currentHeight + 1,
			repeat:    false,
			expRepeat: true, // it was set up with repeat
		},
		"scheduled for current height": {
			contract: myContract,
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight,
		},
		"scheduled for past height": {
			contract: myContract,
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight - 1,
			expErr:   true,
		},
		"undefined task type": {
			contract: myContract,
			height:   currentHeight,
			expErr:   true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			var gotErr error
			if spec.repeat {
				gotErr = k.ScheduleRepeatingTask(ctx, spec.taskType, myContract, spec.height)
			} else {
				gotErr = k.ScheduleOneShotTask(ctx, spec.taskType, myContract, spec.height)
			}
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			repeat, exists := k.getScheduledTaskAt(ctx, spec.taskType, spec.contract, spec.height)
			assert.True(t, exists)
			assert.Equal(t, spec.expRepeat, repeat)
		})
	}
}

func TestDeleteAllScheduledTasks(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))
	myOtherContractWithScheduledTask := sdk.AccAddress(rand.Bytes(32))
	currentHeight := uint64(pCtx.BlockHeight())

	// set a scheduler to overwrite
	err := k.ScheduleRepeatingTask(pCtx, types.SchedulerTaskRebalance, myOtherContractWithScheduledTask, currentHeight+1)
	require.NoError(t, err)
	specs := map[string]struct {
		setup   func(t *testing.T, ctx sdk.Context)
		tp      types.SchedulerTaskType
		expErr  bool
		asserts func(t *testing.T, ctx sdk.Context)
	}{
		"current height": {
			tp: types.SchedulerTaskRebalance,
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContract, uint64(ctx.BlockHeight())))
				require.NoError(t, k.ScheduleOneShotTask(ctx, types.SchedulerTaskRebalance, myContract, uint64(ctx.BlockHeight())))
			},
		},
		"future height": {
			tp: types.SchedulerTaskRebalance,
			setup: func(t *testing.T, ctx sdk.Context) {
				height := uint64(ctx.BlockHeight()) + 1
				require.NoError(t, k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContract, height))
				require.NoError(t, k.ScheduleOneShotTask(ctx, types.SchedulerTaskRebalance, myContract, height))
			},
		},
		"multiple heights": {
			tp: types.SchedulerTaskRebalance,
			setup: func(t *testing.T, ctx sdk.Context) {
				height := uint64(ctx.BlockHeight())
				require.NoError(t, k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContract, height))
				require.NoError(t, k.ScheduleOneShotTask(ctx, types.SchedulerTaskRebalance, myContract, height))
				height++
				require.NoError(t, k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContract, height))
				require.NoError(t, k.ScheduleOneShotTask(ctx, types.SchedulerTaskRebalance, myContract, height))
			},
		},
		"different type not removed": {
			tp: types.SchedulerTaskRebalance,
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.ScheduleRepeatingTask(ctx, types.SchedulerTaskType(0xff), myContract, uint64(ctx.BlockHeight())))
			},
			asserts: func(t *testing.T, ctx sdk.Context) {
				assert.True(t, k.HasScheduledTask(ctx, types.SchedulerTaskType(0xff), myContract, true))
			},
		},
		"unknown type": {
			setup:  func(t *testing.T, ctx sdk.Context) {},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			spec.setup(t, ctx)
			// when
			gotErr := k.DeleteAllScheduledTasks(ctx, spec.tp, myContract)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.False(t, k.HasScheduledTask(ctx, spec.tp, myContract, true))
			assert.False(t, k.HasScheduledTask(ctx, spec.tp, myContract, false))
			// and other contract data not touched
			assert.True(t, k.HasScheduledTask(ctx, spec.tp, myOtherContractWithScheduledTask, true))
			if spec.asserts != nil {
				spec.asserts(t, ctx)
			}
		})
	}
}

func (k Keeper) getScheduledTaskAt(ctx sdk.Context, tp types.SchedulerTaskType, contract sdk.AccAddress, height uint64) (repeat, exists bool) {
	key, err := types.BuildSchedulerContractKey(tp, height, contract)
	if err != nil {
		return false, false
	}
	bz := ctx.KVStore(k.storeKey).Get(key)
	return isRepeat(bz), bz != nil
}
