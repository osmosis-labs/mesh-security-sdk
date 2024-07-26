package keeper

import (
	"bytes"
	stdrand "math/rand"
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestSendAsync(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	var (
		myValAddr                   = sdk.ValAddress(rand.Bytes(address.Len))
		myOtherValAddr              = sdk.ValAddress(rand.Bytes(address.Len))
		myVStakingContractAddr      = sdk.AccAddress(rand.Bytes(address.Len))
		myOtherVStakingContractAddr = sdk.AccAddress(rand.Bytes(address.Len))
	)
	specs := map[string]struct {
		setup  func(t *testing.T, ctx sdk.Context)
		assert func(t *testing.T, ctx sdk.Context, ops map[string][]types.PipedValsetOperation)
	}{
		"no duplicates": {
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.sendAsync(ctx, types.ValidatorModified, myValAddr, nil))
			},
			assert: func(t *testing.T, ctx sdk.Context, ops map[string][]types.PipedValsetOperation) {
				assert.Len(t, ops, 1)
				exp := []types.PipedValsetOperation{types.ValidatorModified}
				assert.Equal(t, exp, ops[myValAddr.String()])
			},
		},
		"separated by validator": {
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.sendAsync(ctx, types.ValidatorModified, myOtherValAddr, nil))
			},
			assert: func(t *testing.T, ctx sdk.Context, ops map[string][]types.PipedValsetOperation) {
				assert.Len(t, ops, 2)
				exp := []types.PipedValsetOperation{types.ValidatorModified}
				assert.Equal(t, exp, ops[myValAddr.String()])
				assert.Equal(t, exp, ops[myOtherValAddr.String()])
			},
		},
		"separated by type": {
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.sendAsync(ctx, types.ValidatorBonded, myValAddr, nil))
			},
			assert: func(t *testing.T, ctx sdk.Context, ops map[string][]types.PipedValsetOperation) {
				assert.Len(t, ops, 1)
				exp := []types.PipedValsetOperation{types.ValidatorBonded, types.ValidatorModified}
				assert.Equal(t, exp, ops[myValAddr.String()])
			},
		},
		"with scheduler per virtual staking": {
			setup: func(t *testing.T, ctx sdk.Context) {
				err := k.SetMaxCapLimit(ctx, myVStakingContractAddr, sdk.NewCoin("stake", sdkmath.NewInt(100_000_000)))
				require.NoError(t, err)
				err = k.SetMaxCapLimit(ctx, myOtherVStakingContractAddr, sdk.NewCoin("stake", sdkmath.NewInt(100_000_000)))
				require.NoError(t, err)
			},
			assert: func(t *testing.T, ctx sdk.Context, ops map[string][]types.PipedValsetOperation) {
				isScheduled := k.HasScheduledTask(ctx, types.SchedulerTaskValsetUpdate, myVStakingContractAddr, false)
				assert.True(t, isScheduled)
				isScheduled = k.HasScheduledTask(ctx, types.SchedulerTaskValsetUpdate, myOtherVStakingContractAddr, false)
				assert.True(t, isScheduled)
			},
		},
		"with scheduler for enabled contracts only": {
			setup: func(t *testing.T, ctx sdk.Context) {
				err := k.SetMaxCapLimit(ctx, myVStakingContractAddr, sdk.NewCoin("stake", sdkmath.NewInt(100_000_000)))
				require.NoError(t, err)
			},
			assert: func(t *testing.T, ctx sdk.Context, ops map[string][]types.PipedValsetOperation) {
				isScheduled := k.HasScheduledTask(ctx, types.SchedulerTaskValsetUpdate, myVStakingContractAddr, false)
				assert.True(t, isScheduled)
				isScheduled = k.HasScheduledTask(ctx, types.SchedulerTaskValsetUpdate, myOtherVStakingContractAddr, false)
				assert.False(t, isScheduled)
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			spec.setup(t, ctx)
			// when
			gotErr := k.sendAsync(ctx, types.ValidatorModified, myValAddr, nil)
			// then
			require.NoError(t, gotErr)
			allStoredOps := FetchAllStoredOperations(t, ctx, k)
			spec.assert(t, ctx, allStoredOps)
		})
	}
}

func TestBuildValsetUpdateReport(t *testing.T) {
	var (
		val1 = sdk.ValAddress(bytes.Repeat([]byte{1}, address.Len))
		val2 = sdk.ValAddress(bytes.Repeat([]byte{2}, address.Len))
		val3 = sdk.ValAddress(bytes.Repeat([]byte{3}, address.Len))
		val4 = sdk.ValAddress(bytes.Repeat([]byte{4}, address.Len))
	)
	ctx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	vals := make(map[string]stakingtypes.Validator)
	for i, v := range []sdk.ValAddress{val1, val2, val3, val4} {
		val := MinValidatorFixture(t)
		val.OperatorAddress = v.String()
		val.Commission.CommissionRates.Rate = sdkmath.LegacyNewDec(int64(i + 1))
		keepers.StakingKeeper.SetValidator(ctx, val)
		vals[v.String()] = val
	}

	type tuple struct {
		op      types.PipedValsetOperation
		valAddr sdk.ValAddress
	}
	allOps := []tuple{
		{op: types.ValidatorJailed, valAddr: val1},
		{op: types.ValidatorUnjailed, valAddr: val2},
		{op: types.ValidatorModified, valAddr: val2},
		{op: types.ValidatorUnbonded, valAddr: val3},
		{op: types.ValidatorJailed, valAddr: val3},
		{op: types.ValidatorTombstoned, valAddr: val3},
		{op: types.ValidatorModified, valAddr: val3},
		{op: types.ValidatorBonded, valAddr: val4},
		{op: types.ValidatorModified, valAddr: val4},
	}
	stdrand.Shuffle(len(allOps), func(i, j int) {
		allOps[i], allOps[j] = allOps[j], allOps[i]
	})

	for _, v := range allOps {
		require.NoError(t, k.sendAsync(ctx, v.op, v.valAddr, nil))
	}
	// when
	got, err := k.ValsetUpdateReport(ctx)
	// then
	require.NoError(t, err)
	exp := contract.HandleValsetUpdate{
		Additions: []contract.Validator{
			{
				Address:       val4.String(),
				Commission:    "4.000000000000000000",
				MaxCommission: "0.000000000000000000",
				MaxChangeRate: "0.000000000000000000",
			},
		},
		Removals: []contract.ValidatorAddr{val3.String()},
		Updated: []contract.Validator{
			{
				Address:       val2.String(),
				Commission:    "2.000000000000000000",
				MaxCommission: "0.000000000000000000",
				MaxChangeRate: "0.000000000000000000",
			},
			{
				Address:       val3.String(),
				Commission:    "3.000000000000000000",
				MaxCommission: "0.000000000000000000",
				MaxChangeRate: "0.000000000000000000",
			},
			{
				Address:       val4.String(),
				Commission:    "4.000000000000000000",
				MaxCommission: "0.000000000000000000",
				MaxChangeRate: "0.000000000000000000",
			},
		},
		Jailed:     []contract.ValidatorAddr{val1.String(), val3.String()},
		Unjailed:   []contract.ValidatorAddr{val2.String()},
		Slashed:    []contract.ValidatorSlash{},
		Tombstoned: []contract.ValidatorAddr{val3.String()},
	}
	assert.Equal(t, exp, got)
}

func TestValsetUpdateReportErrors(t *testing.T) {
	nonValAddr := sdk.ValAddress(bytes.Repeat([]byte{1}, address.Len))
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper

	specs := map[string]struct {
		setup  func(t *testing.T, ctx sdk.Context)
		expErr error
	}{
		"unknown val address": {
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.sendAsync(ctx, types.ValidatorBonded, nonValAddr, nil))
			},
			expErr: types.ErrUnknown,
		},
		"unsupported val operation": {
			setup: func(t *testing.T, ctx sdk.Context) {
				require.NoError(t, k.sendAsync(ctx, 0xff, nonValAddr, nil))
			},
			expErr: types.ErrInvalid,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			spec.setup(t, ctx)
			// when
			_, gotErr := k.ValsetUpdateReport(ctx)
			require.Error(t, gotErr)
			assert.ErrorIs(t, spec.expErr, gotErr)
		})
	}
}

func TestClearPipedValsetOperations(t *testing.T) {
	ctx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	err := k.sendAsync(ctx, types.ValidatorModified, rand.Bytes(address.Len), nil)
	require.NoError(t, err)
	err = k.sendAsync(ctx, types.ValidatorUnjailed, rand.Bytes(address.Len), nil)
	require.NoError(t, err)
	// when
	k.ClearPipedValsetOperations(ctx)
	// then
	assert.Empty(t, FetchAllStoredOperations(t, ctx, k))
}
