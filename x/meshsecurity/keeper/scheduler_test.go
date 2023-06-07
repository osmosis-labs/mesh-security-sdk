package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			gotErr := k.ScheduleTask(ctx, types.SchedulerTaskRebalance, myContract, currentHeight, spec.repeat)
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
			repeat, exists := k.getScheduledTaskAt(ctx, types.SchedulerTaskRebalance, gotRes[0].NextRunHeight, myContract)
			assert.Equal(t, spec.expRescheduled, exists)
			assert.Equal(t, spec.repeat, repeat)
			// and old entry removed
			_, exists = k.getScheduledTaskAt(ctx, types.SchedulerTaskRebalance, currentHeight, myContract)
			assert.False(t, exists)
		})
	}
}

func TestScheduleTask(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))
	currentHeight := uint64(pCtx.BlockHeight())
	specs := map[string]struct {
		taskType types.SchedulerTaskType
		height   uint64
		repeat   bool
		expErr   bool
	}{
		"all good - single exec": {
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight + 1,
		},
		"all good - repeat": {
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight + 1,
			repeat:   true,
		},
		"scheduled for current height": {
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight,
		},
		"scheduled for past height": {
			taskType: types.SchedulerTaskRebalance,
			height:   currentHeight - 1,
			expErr:   true,
		},
		"undefined task type": {
			height: currentHeight,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			gotErr := k.ScheduleTask(ctx, spec.taskType, myContract, spec.height, spec.repeat)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			repeat, exists := k.getScheduledTaskAt(ctx, types.SchedulerTaskRebalance, spec.height, myContract)
			assert.True(t, exists)
			assert.Equal(t, spec.repeat, repeat)
		})
	}
}
