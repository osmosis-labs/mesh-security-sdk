package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestCaptureTombstone(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)

	val := validatorFixture(t)
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
			appStoredOps := fetchAllStoredOperations(t, ctx, keepers.MeshKeeper)
			assert.Equal(t, spec.expStored, appStoredOps[val.OperatorAddress])
		})
	}
}

func TestCaptureStakingEvents(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)

	val := validatorFixture(t)
	myConsAddress, err := val.GetConsAddr()
	require.NoError(t, err)
	keepers.StakingKeeper.SetValidatorByConsAddr(pCtx, val)
	keepers.StakingKeeper.SetValidator(pCtx, val)

	valJailed := validatorFixture(t)
	valJailed.Jailed = true
	myConsAddressJailed, err := valJailed.GetConsAddr()
	require.NoError(t, err)
	keepers.StakingKeeper.SetValidatorByConsAddr(pCtx, valJailed)
	keepers.StakingKeeper.SetValidator(pCtx, valJailed)

	decorator := NewStakingDecorator(keepers.StakingKeeper, keepers.MeshKeeper)
	specs := map[string]struct {
		consAddr  sdk.ConsAddress
		op        func(sdk.Context, sdk.ConsAddress)
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
			loadedVal := keepers.StakingKeeper.ValidatorByConsAddr(ctx, spec.consAddr)
			assert.Equal(t, spec.expJailed, loadedVal.IsJailed())
			// and stored for async propagation
			allStoredOps := fetchAllStoredOperations(t, ctx, keepers.MeshKeeper)
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

// creates minimal validator object
func validatorFixture(t *testing.T) stakingtypes.Validator {
	t.Helper()
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	pk, err := cryptocodec.FromTmPubKeyInterface(pubKey)
	require.NoError(t, err)
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	return stakingtypes.Validator{
		ConsensusPubkey: pkAny,
		OperatorAddress: sdk.ValAddress(rand.Bytes(address.Len)).String(),
	}
}
