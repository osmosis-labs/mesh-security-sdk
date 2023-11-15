package meshsecurity

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestEndBlocker(t *testing.T) {
	var (
		capturedCalls []capturedSudo
		contractErr   error
		logRecords    bytes.Buffer
	)
	pCtx, keepers := keeper.CreateDefaultTestInput(t, keeper.WithWasmKeeperDecorated(func(original types.WasmKeeper) types.WasmKeeper {
		return captureSudos(&capturedCalls, &contractErr)
	}))
	val1 := keeper.MinValidatorFixture(t)
	keepers.StakingKeeper.SetValidator(pCtx, val1)
	k := keepers.MeshKeeper
	var (
		myError             = errors.New("my test error")
		myContractAddr      = sdk.AccAddress(bytes.Repeat([]byte{1}, 32))
		myOtherContractAddr = sdk.AccAddress(bytes.Repeat([]byte{2}, 32))
	)

	specs := map[string]struct {
		setup  func(t *testing.T, ctx sdk.Context)
		assert func(t *testing.T, ctx sdk.Context)
	}{
		"rebalance - multiple contracts": {
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t,
					k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContractAddr, uint64(ctx.BlockHeight())))
				require.NoError(t,
					k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myOtherContractAddr, uint64(ctx.BlockHeight())))
			},
			assert: func(t *testing.T, ctx sdk.Context) {
				require.Len(t, capturedCalls, 2)
				assert.Equal(t, myContractAddr, capturedCalls[0].contractAddress)
				assert.JSONEq(t, `{"rebalance":{}}`, string(capturedCalls[0].msg))
				assert.Equal(t, myOtherContractAddr, capturedCalls[1].contractAddress)
				assert.JSONEq(t, `{"rebalance":{}}`, string(capturedCalls[1].msg))
				assert.NotContains(t, logRecords.String(), "failed")
			},
		},
		"rebalance - contract errored": {
			setup: func(t *testing.T, ctx sdk.Context) {
				contractErr = myError
				require.NoError(t,
					k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myContractAddr, uint64(ctx.BlockHeight())))
				require.NoError(t,
					k.ScheduleRepeatingTask(ctx, types.SchedulerTaskRebalance, myOtherContractAddr, uint64(ctx.BlockHeight())))
			},
			assert: func(t *testing.T, ctx sdk.Context) {
				require.Len(t, capturedCalls, 2)
				assert.Contains(t, logRecords.String(), "failed to execute scheduled task")
			},
		},
		"valset update - multiple contracts": {
			setup: func(t *testing.T, ctx sdk.Context) {
				anyLimit := sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1_000_000_000))
				require.NoError(t, k.SetMaxCapLimit(ctx, myContractAddr, anyLimit))
				require.NoError(t, k.SetMaxCapLimit(ctx, myOtherContractAddr, anyLimit))
				require.NoError(t, k.Hooks().AfterValidatorBonded(ctx, nil, val1.GetOperator()))
			},
			assert: func(t *testing.T, ctx sdk.Context) {
				require.Len(t, capturedCalls, 2)
				assert.Equal(t, myContractAddr, capturedCalls[0].contractAddress)
				exp := fmt.Sprintf(`{"valset_update":{"additions":[{"address":"%s","commission":"0.000000000000000000","max_commission":"0.000000000000000000","max_change_rate":"0.000000000000000000"}],"removals":[],"updated":[],"jailed":[],"unjailed":[],"tombstoned":[]}}`, val1.GetOperator())
				assert.JSONEq(t, exp, string(capturedCalls[0].msg))

				assert.Equal(t, myOtherContractAddr, capturedCalls[1].contractAddress)
				assert.JSONEq(t, exp, string(capturedCalls[1].msg))
				assert.NotContains(t, logRecords.String(), "failed")
			},
		},
		"valset update - contract errored": {
			setup: func(t *testing.T, ctx sdk.Context) {
				anyLimit := sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1_000_000_000))
				require.NoError(t, k.SetMaxCapLimit(ctx, myContractAddr, anyLimit))
				require.NoError(t, k.SetMaxCapLimit(ctx, myOtherContractAddr, anyLimit))
				require.NoError(t, k.Hooks().AfterValidatorBonded(ctx, nil, val1.GetOperator()))
				contractErr = myError
			},
			assert: func(t *testing.T, ctx sdk.Context) {
				require.Len(t, capturedCalls, 2)
				assert.Contains(t, logRecords.String(), "failed to execute scheduled task")
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedCalls, contractErr = nil, nil
			logRecords.Reset()
			ctx, _ := pCtx.CacheContext()
			spec.setup(t, ctx)
			// when
			EndBlocker(ctx.WithLogger(log.NewTMLogger(log.NewSyncWriter(&logRecords))), k, DefaultExecutionResponseHandler())
			spec.assert(t, ctx)
		})
	}
}

type capturedSudo = struct {
	contractAddress sdk.AccAddress
	msg             []byte
}

func captureSudos(captured *[]capturedSudo, e *error) *keeper.MockWasmKeeper {
	return &keeper.MockWasmKeeper{
		SudoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
			*captured = append(*captured, capturedSudo{contractAddress: contractAddress, msg: msg})
			return nil, *e
		},
		HasContractInfoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
			return true
		},
	}
}
