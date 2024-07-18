package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	cptypes "github.com/osmosis-labs/mesh-security-sdk/x/types"
)

func TestCaptureTombstone(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)

	val := MinValidatorFixture(t)
	myConsAddress, err := val.GetConsAddr()
	require.NoError(t, err)
	keepers.StakingKeeper.SetValidatorByConsAddr(pCtx, val)
	keepers.StakingKeeper.SetValidator(pCtx, val)
	skMock, capturedTombstones := NewMockEvidenceSlashingKeeper()
	decorator := CaptureTombstoneDecorator(keepers.MeshKeeper, skMock, keepers.StakingKeeper)
	otherConsAddress := rand.Bytes(address.Len)
	specs := map[string]struct {
		addr      sdk.ConsAddress
		expPassed []sdk.ConsAddress
		expStored []cptypes.PipedValsetOperation
	}{
		"with existing validator": {
			addr:      myConsAddress,
			expPassed: []sdk.ConsAddress{myConsAddress},
			expStored: []cptypes.PipedValsetOperation{cptypes.PipedValsetOperation_VALIDATOR_TOMBSTONED},
		},
		"unknown consensus address": {
			addr:      otherConsAddress,
			expPassed: []sdk.ConsAddress{otherConsAddress},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			*capturedTombstones = make([]sdk.ConsAddress, 0, 1)
			// when
			decorator.Tombstone(ctx, spec.addr)

			// then
			assert.Equal(t, spec.expPassed, *capturedTombstones)
			// and stored for async propagation
			appStoredOps := FetchAllStoredOperations(t, ctx, keepers.MeshKeeper)
			assert.Equal(t, spec.expStored, appStoredOps[val.OperatorAddress])
		})
	}
}

func TestCaptureStakingEvents(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	denom := keepers.StakingKeeper.BondDenom(pCtx)
	coin := sdk.NewCoin(denom, sdk.NewIntFromUint64(1000000))

	val := MinValidatorFixture(t)
	val.Status = types.Bonded
	val.Tokens = sdk.NewIntFromUint64(1000000)
	myConsAddress, err := val.GetConsAddr()
	require.NoError(t, err)
	acc := sdk.AccAddress(rand.Bytes(32))
	keepers.BankKeeper.MintCoins(pCtx, minttypes.ModuleName, sdk.NewCoins([]sdk.Coin{coin}...))
	keepers.BankKeeper.SendCoinsFromModuleToAccount(pCtx, minttypes.ModuleName, acc, sdk.NewCoins([]sdk.Coin{coin}...))
	keepers.BankKeeper.DelegateCoinsFromAccountToModule(pCtx, acc, types.BondedPoolName, sdk.NewCoins([]sdk.Coin{coin}...))
	keepers.StakingKeeper.SetValidatorByConsAddr(pCtx, val)
	keepers.StakingKeeper.SetValidator(pCtx, val)

	valJailed := MinValidatorFixture(t)
	valJailed.Jailed = true
	myConsAddressJailed, err := valJailed.GetConsAddr()
	require.NoError(t, err)
	keepers.StakingKeeper.SetValidatorByConsAddr(pCtx, valJailed)
	keepers.StakingKeeper.SetValidator(pCtx, valJailed)

	decorator := NewStakingDecorator(keepers.StakingKeeper, keepers.MeshKeeper)
	specs := map[string]struct {
		consAddr  sdk.ConsAddress
		op        func(sdk.Context, sdk.ConsAddress)
		expStored []cptypes.PipedValsetOperation
		expJailed bool
	}{
		"slash and jail": {
			consAddr: myConsAddress,
			op: func(ctx sdk.Context, ca sdk.ConsAddress) {
				decorator.Slash(ctx, ca, ctx.BlockHeight(), 1, sdk.MustNewDecFromStr("0.1"))
				decorator.Jail(ctx, ca)
			},
			expStored: []cptypes.PipedValsetOperation{
				cptypes.PipedValsetOperation_VALIDATOR_JAILED,
				cptypes.PipedValsetOperation_VALIDATOR_MODIFIED,
				cptypes.PipedValsetOperation_VALIDATOR_SLASHED,
			},
			expJailed: true,
		},
		"unjail": {
			consAddr:  myConsAddressJailed,
			op:        decorator.Unjail,
			expStored: []cptypes.PipedValsetOperation{cptypes.PipedValsetOperation_VALIDATOR_UNJAILED},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()

			// when
			spec.op(ctx, spec.consAddr)

			// then
			loadedVal, found := keepers.StakingKeeper.GetValidatorByConsAddr(ctx, spec.consAddr)
			assert.True(t, found)
			assert.Equal(t, spec.expJailed, loadedVal.IsJailed())
			// and stored for async propagation
			allStoredOps := FetchAllStoredOperations(t, ctx, keepers.MeshKeeper)
			assert.Equal(t, spec.expStored, allStoredOps[loadedVal.GetOperator().String()])
		})
	}
}

type MockEvidenceSlashingKeeper struct {
	evidencetypes.SlashingKeeper
	tombstoned []sdk.ConsAddress
}

func NewMockEvidenceSlashingKeeper() (*MockEvidenceSlashingKeeper, *[]sdk.ConsAddress) {
	r := MockEvidenceSlashingKeeper{tombstoned: make([]sdk.ConsAddress, 0)}
	return &r, &r.tombstoned
}

func (e *MockEvidenceSlashingKeeper) Tombstone(ctx sdk.Context, address sdk.ConsAddress) {
	e.tombstoned = append(e.tombstoned, address)
}
