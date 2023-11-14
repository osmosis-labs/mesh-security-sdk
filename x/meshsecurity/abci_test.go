package meshsecurity

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type MockWasmKeeper struct{}

func TestEndBlocker(t *testing.T) {
	caturedCalls := make([]capturedSudo, 0, 2)
	pCtx, keepers := keeper.CreateDefaultTestInput(t, keeper.WithWasmKeeperDecorated(func(original types.WasmKeeper) types.WasmKeeper {
		return captureSudos(&caturedCalls)
	}))
	k := keepers.MeshKeeper
	myContractAddr := sdk.AccAddress(rand.Bytes(32))
	specs := map[string]struct {
		setup  func(t *testing.T, ctx sdk.Context)
		assert func(t *testing.T, ctx sdk.Context)
	}{
		"rebalance": {
			setup: func(t *testing.T, ctx sdk.Context) {
				err := k.ScheduleRepeatingTask(pCtx, types.SchedulerTaskRebalance, myContractAddr, uint64(pCtx.BlockHeight()))
				require.NoError(t, err)
			},
			assert: func(t *testing.T, ctx sdk.Context) {
				require.Len(t, caturedCalls, 1)
				assert.Equal(t, myContractAddr, caturedCalls[0].contractAddress)
				assert.JSONEq(t, `{"rebalance":{}}`, string(caturedCalls[0].msg))
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			clear(caturedCalls)
			ctx, _ := pCtx.CacheContext()
			spec.setup(t, ctx)
			// when
			EndBlocker(pCtx, k, DefaultExecutionResponseHandler())
			spec.assert(t, ctx)
		})
	}
}

type capturedSudo = struct {
	contractAddress sdk.AccAddress
	msg             []byte
}

func captureSudos(captured *[]capturedSudo) *keeper.MockWasmKeeper {
	return &keeper.MockWasmKeeper{
		SudoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
			*captured = append(*captured, capturedSudo{contractAddress: contractAddress, msg: msg})
			return nil, nil
		},
		HasContractInfoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
			return true
		},
	}
}
