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
	mock, capturedTombstones := NewMockEvidenceSlashingKeeper()
	cap := CaptureTombstoneDecorator(keepers.MeshKeeper, mock, keepers.StakingKeeper)
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
			cap.Tombstone(ctx, spec.addr)

			// then
			assert.Equal(t, spec.expPassed, *capturedTombstones)
			// and stored for async propagation
			operations := fetchAllStoredOperations(t, keepers.MeshKeeper, ctx, val)
			assert.Equal(t, spec.expStored, operations)
		})
	}
}

func fetchAllStoredOperations(t *testing.T, msKeeper *Keeper, ctx sdk.Context, val stakingtypes.Validator) []types.PipedValsetOperation {
	index := make(map[string][]types.PipedValsetOperation, 1)

	err := msKeeper.iteratePipedValsetOperations(ctx, func(valAddr sdk.ValAddress, op types.PipedValsetOperation) bool {
		ops, ok := index[valAddr.String()]
		if !ok {
			ops = []types.PipedValsetOperation{}
		}
		index[valAddr.String()] = append(ops, op)
		return false
	})
	require.NoError(t, err)
	operations := index[val.OperatorAddress]
	return operations
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
