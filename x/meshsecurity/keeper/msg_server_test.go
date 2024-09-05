package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestSetVirtualStakingMaxCap(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))
	denom := keepers.StakingKeeper.BondDenom(pCtx)
	myAmount := sdk.NewInt64Coin(denom, 123)

	k.wasm = MockWasmKeeper{HasContractInfoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
		return contractAddress.Equals(myContract)
	}}
	m := NewMsgServer(k)

	specs := map[string]struct {
		src         types.MsgSetVirtualStakingMaxCap
		setup       func(ctx sdk.Context)
		expErr      bool
		expLimit    sdk.Coin
		expSchedule func(t *testing.T, ctx sdk.Context)
	}{
		"limit stored with scheduler for existing contract": {
			setup: func(ctx sdk.Context) {},
			src: types.MsgSetVirtualStakingMaxCap{
				Authority: k.GetAuthority(),
				Contract:  myContract.String(),
				MaxCap:    myAmount,
			},
			expLimit: myAmount,
			expSchedule: func(t *testing.T, ctx sdk.Context) {
				assert.True(t, k.HasScheduledTask(ctx, types.SchedulerTaskHandleEpoch, myContract, true))
			},
		},
		"fails for non existing contract": {
			setup: func(ctx sdk.Context) {},
			src: types.MsgSetVirtualStakingMaxCap{
				Authority: k.GetAuthority(),
				Contract:  sdk.AccAddress(rand.Bytes(32)).String(),
				MaxCap:    myAmount,
			},
			expErr: true,
		},
		"unauthorized rejected": {
			setup: func(ctx sdk.Context) {},
			src: types.MsgSetVirtualStakingMaxCap{
				Authority: myContract.String(),
				Contract:  myContract.String(),
				MaxCap:    myAmount,
			},
			expErr: true,
		},
		"invalid data rejected": {
			setup:  func(ctx sdk.Context) {},
			src:    types.MsgSetVirtualStakingMaxCap{},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			spec.setup(ctx)

			// when
			gotRsp, gotErr := m.SetVirtualStakingMaxCap(sdk.WrapSDKContext(ctx), &spec.src)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.NotNil(t, gotRsp)
			require.True(t, k.HasMaxCapLimit(ctx, myContract))
			assert.Equal(t, spec.expLimit, k.GetMaxCapLimit(ctx, myContract))
			// and scheduled
			spec.expSchedule(t, ctx)
		})
	}
}
