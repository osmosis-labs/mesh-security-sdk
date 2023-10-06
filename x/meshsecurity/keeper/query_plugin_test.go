package keeper

import (
	"fmt"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestChainedCustomQuerier(t *testing.T) {
	myContractAddr := sdk.AccAddress(rand.Bytes(32))
	pCtx, keepers := CreateDefaultTestInput(t)

	specs := map[string]struct {
		src           wasmvmtypes.QueryRequest
		viewKeeper    viewKeeper
		expData       []byte
		expErr        bool
		expNextCalled bool
	}{
		"bond status query": {
			src: wasmvmtypes.QueryRequest{
				Custom: []byte(fmt.Sprintf(`{"virtual_stake":{"bond_status":{"contract":%q}}}`, myContractAddr.String())),
			},
			viewKeeper: &MockViewKeeper{
				GetMaxCapLimitFn: func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
					return sdk.NewCoin("ALX", math.NewInt(123))
				},
				GetTotalDelegatedFn: func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
					return sdk.NewCoin("ALX", math.NewInt(456))
				},
			},
			expData: []byte(`{"cap":{"denom":"ALX","amount":"123"},"delegated":{"denom":"ALX","amount":"456"}}`),
		},
		"slash ratio query": {
			src: wasmvmtypes.QueryRequest{
				Custom: []byte(`{"virtual_stake":{"slash_ratio":{}}}`),
			},
			viewKeeper: keepers.MeshKeeper,
			expData:    []byte(`{"slash_fraction_downtime":"0.010000000000000000","slash_fraction_double_sign":"0.050000000000000000"}`),
		},
		"non custom query": {
			src: wasmvmtypes.QueryRequest{
				Bank: &wasmvmtypes.BankQuery{},
			},
			viewKeeper:    keepers.MeshKeeper,
			expNextCalled: true,
		},
		"custom non mesh query": {
			src: wasmvmtypes.QueryRequest{
				Custom: []byte(`{"foo":{}}`),
			},
			viewKeeper:    keepers.MeshKeeper,
			expNextCalled: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var nextCalled bool
			next := QueryHandlerFn(func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
				nextCalled = true
				return nil, nil
			})

			ctx, _ := pCtx.CacheContext()
			gotData, gotErr := ChainedCustomQuerier(spec.viewKeeper, keepers.SlashingKeeper, next).HandleQuery(ctx, myContractAddr, spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expData, gotData, string(gotData))
			assert.Equal(t, spec.expNextCalled, nextCalled)
		})
	}
}

var _ viewKeeper = &MockViewKeeper{}

type MockViewKeeper struct {
	GetMaxCapLimitFn    func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
	GetTotalDelegatedFn func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
}

func (m MockViewKeeper) GetMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
	if m.GetMaxCapLimitFn == nil {
		panic("not expected to be called")
	}
	return m.GetMaxCapLimitFn(ctx, actor)
}

func (m MockViewKeeper) GetTotalDelegated(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
	if m.GetTotalDelegatedFn == nil {
		panic("not expected to be called")
	}
	return m.GetTotalDelegatedFn(ctx, actor)
}
