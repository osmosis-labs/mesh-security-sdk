package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestQueryVirtualStakingMaxCapLimit(t *testing.T) {
	ctx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))
	myAmount := sdk.NewInt64Coin(keepers.StakingKeeper.BondDenom(ctx), 123)

	err := k.SetMaxCapLimit(ctx, myContract, myAmount)
	require.NoError(t, err)
	specs := map[string]struct {
		addr      string
		expAmount sdk.Coin
		expErr    bool
	}{
		"existing contract limit": {
			addr:      myContract.String(),
			expAmount: myAmount,
		},
		"non existing contract limit": {
			addr:      sdk.AccAddress(rand.Bytes(32)).String(),
			expAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
		},
		"invalid address": {
			addr:   "not-an-address",
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := NewQuerier(keepers.EncodingConfig.Marshaler, k).VirtualStakingMaxCapLimit(sdk.WrapSDKContext(ctx), &types.QueryVirtualStakingMaxCapLimitRequest{
				Address: spec.addr,
			})
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expAmount, gotRsp.Cap)
		})
	}
}

func TestQueryVirtualStakingMaxCapLimits(t *testing.T) {
	t.Skip("TODO")
	// setup
	ctx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper

	querier := NewQuerier(keepers.EncodingConfig.Marshaler, k)
	// when
	gotRsp, err := querier.VirtualStakingMaxCapLimits(sdk.WrapSDKContext(ctx), &types.QueryVirtualStakingMaxCapLimitsRequest{})
	// then
	require.NoError(t, err)
	assert.Zero(t, len(gotRsp.MaxCapInfos))

	// set max cap for a random contract
	myContract := sdk.AccAddress(rand.Bytes(32))
	myAmount := sdk.NewInt64Coin(keepers.StakingKeeper.BondDenom(ctx), 123)
	err = k.SetMaxCapLimit(ctx, myContract, myAmount)
	require.NoError(t, err)

	// when
	gotRsp, err = querier.VirtualStakingMaxCapLimits(sdk.WrapSDKContext(ctx), &types.QueryVirtualStakingMaxCapLimitsRequest{})
	// then
	require.NoError(t, err)
	assert.Equal(t, 1, len(gotRsp.MaxCapInfos))
	assert.Equal(t, myContract, gotRsp.MaxCapInfos[0].Contract)
	assert.Equal(t, myAmount, gotRsp.MaxCapInfos[0].Cap)

	// set max cap for another contract
	err = k.SetMaxCapLimit(ctx, sdk.AccAddress(rand.Bytes(32)), sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
	require.NoError(t, err)

	// when
	gotRsp, err = querier.VirtualStakingMaxCapLimits(sdk.WrapSDKContext(ctx), &types.QueryVirtualStakingMaxCapLimitsRequest{})
	// then
	require.NoError(t, err)
	assert.Equal(t, 2, len(gotRsp.MaxCapInfos))
}
