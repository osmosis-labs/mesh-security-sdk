package keeper

import (
	"context"
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	evidencetypes "cosmossdk.io/x/evidence/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
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
		expStored []types.PipedValsetOperation
	}{
		"with existing validator": {
			addr:      myConsAddress,
			expPassed: []sdk.ConsAddress{myConsAddress},
			expStored: []types.PipedValsetOperation{types.ValidatorTombstoned},
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

	val := MinValidatorFixture(t)
	myConsAddress, err := val.GetConsAddr()
	require.NoError(t, err)
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
		op        func(context.Context, sdk.ConsAddress)
		expStored []types.PipedValsetOperation
		expJailed bool
	}{
		"jail": {
			consAddr:  myConsAddress,
			op:        decorator.Jail,
			expStored: []types.PipedValsetOperation{types.ValidatorJailed},
			expJailed: true,
		},
		"unjail": {
			consAddr:  myConsAddressJailed,
			op:        decorator.Unjail,
			expStored: []types.PipedValsetOperation{types.ValidatorUnjailed},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()

			// when
			spec.op(ctx, spec.consAddr)

			// then
			loadedVal, err := keepers.StakingKeeper.ValidatorByConsAddr(ctx, spec.consAddr)
			require.NoError(t, err)
			assert.Equal(t, spec.expJailed, loadedVal.IsJailed())
			// and stored for async propagation
			allStoredOps := FetchAllStoredOperations(t, ctx, keepers.MeshKeeper)
			assert.Equal(t, spec.expStored, allStoredOps[loadedVal.GetOperator()])
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

func (e *MockEvidenceSlashingKeeper) Tombstone(ctx context.Context, address sdk.ConsAddress) error {
	e.tombstoned = append(e.tombstoned, address)
	return nil
}
