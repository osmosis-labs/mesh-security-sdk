package keeper

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainedCustomQuerier(t *testing.T) {
	myContractAddr := sdk.AccAddress(rand.Bytes(32))
	specs := map[string]struct {
		src                wasmvmtypes.QueryRequest
		mockMaxCapLimit    func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
		mockTotalDelegated func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
		expData            []byte
		expErr             bool
		expMocNextCalled   bool
	}{
		"all good": {
			src: wasmvmtypes.QueryRequest{
				Custom: []byte(fmt.Sprintf(`{"virtual_stake":{"bond_status":{"contract":%q}}}`, myContractAddr.String())),
			},
			mockMaxCapLimit: func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
				return sdk.NewCoin("ALX", math.NewInt(123))
			},
			mockTotalDelegated: func(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
				return sdk.NewCoin("ALX", math.NewInt(456))
			},
			expData: []byte(`{"cap":{"denom":"ALX","amount":"123"},"delegated":{"denom":"ALX","amount":"456"}}`),
		},
		"non custom query": {
			src: wasmvmtypes.QueryRequest{
				Bank: &wasmvmtypes.BankQuery{},
			},
			expMocNextCalled: true,
		},
		"custom non mesh query": {
			src: wasmvmtypes.QueryRequest{
				Custom: []byte(`{"foo":{}}`),
			},
			expMocNextCalled: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var nextCalled bool
			next := QueryHandlerFn(func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
				nextCalled = true
				return nil, nil
			})
			mock := &MockViewKeeper{GetMaxCapLimitFn: spec.mockMaxCapLimit, GetTotalDelegatedFn: spec.mockTotalDelegated}
			ctx := sdk.Context{}
			gotData, gotErr := ChainedCustomQuerier(mock, next).HandleQuery(ctx, myContractAddr, spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expData, gotData)
			assert.Equal(t, spec.expMocNextCalled, nextCalled)
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
