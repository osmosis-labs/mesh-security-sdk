package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestHasMaxCapLimit(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContractAddr := sdk.AccAddress(rand.Bytes(32))

	specs := map[string]struct {
		setup     func(ctx sdk.Context)
		expResult bool
	}{
		"limit set": {
			setup: func(ctx sdk.Context) {
				err := k.SetMaxCapLimit(ctx, myContractAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
				require.NoError(t, err)
			},
			expResult: true,
		},
		"limit with empty amount set": {
			setup: func(ctx sdk.Context) {
				err := k.SetMaxCapLimit(ctx, myContractAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
				require.NoError(t, err)
			},
			expResult: true,
		},
		"limit not set": {
			setup:     func(ctx sdk.Context) {}, // noop
			expResult: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			spec.setup(ctx)
			got := k.HasMaxCapLimit(ctx, myContractAddr)
			assert.Equal(t, spec.expResult, got)
		})
	}
}

func TestSetMaxCapLimit(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	var (
		myContractAddr = sdk.AccAddress(rand.Bytes(32))
		oneStakeCoin   = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)
		zeroStakeCoin  = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	)

	specs := map[string]struct {
		setup  func(ctx sdk.Context) sdk.Coin
		expErr bool
	}{
		"all good": {
			setup: func(_ sdk.Context) sdk.Coin {
				return oneStakeCoin
			},
		},
		"zero amount allowed": {
			setup: func(_ sdk.Context) sdk.Coin {
				return zeroStakeCoin
			},
		},
		"overwrite existing value": {
			setup: func(ctx sdk.Context) sdk.Coin {
				err := k.SetMaxCapLimit(ctx, myContractAddr, oneStakeCoin)
				require.NoError(t, err)
				return oneStakeCoin.AddAmount(math.NewInt(1))
			},
		},
		"within total contracts max limit": {
			setup: func(ctx sdk.Context) sdk.Coin {
				p := k.GetParams(ctx)
				p.TotalContractsMaxCap = oneStakeCoin
				require.NoError(t, k.SetParams(ctx, p))
				return oneStakeCoin
			},
		},
		"non staking denom rejected": {
			setup: func(_ sdk.Context) sdk.Coin {
				return sdk.NewInt64Coin("NON", 1)
			},
			expErr: true,
		},
		"total contracts max exceeded - with other contract": {
			setup: func(ctx sdk.Context) sdk.Coin {
				p := k.GetParams(ctx)
				p.TotalContractsMaxCap = oneStakeCoin
				require.NoError(t, k.SetParams(ctx, p))
				require.NoError(t, k.SetMaxCapLimit(ctx, rand.Bytes(32), oneStakeCoin))
				return oneStakeCoin
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			limitAmount := spec.setup(ctx)

			// when
			gotErr := k.SetMaxCapLimit(ctx, myContractAddr, limitAmount)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				limitAmount = zeroStakeCoin
			} else {
				require.NoError(t, gotErr)
			}
			assert.Equal(t, limitAmount, k.GetMaxCapLimit(ctx, myContractAddr))
		})
	}
}
