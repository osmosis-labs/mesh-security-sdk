package keeper

import (
	"errors"
	"fmt"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

func TestCustomMeshSecDispatchMsg(t *testing.T) {
	var (
		noAuthZ    = AuthSourceFn(func(_ sdk.Context, _ sdk.AccAddress) bool { return false })
		allAuthZ   = AuthSourceFn(func(_ sdk.Context, _ sdk.AccAddress) bool { return true })
		panicAuthZ = AuthSourceFn(func(_ sdk.Context, _ sdk.AccAddress) bool { panic("not supposed to be called") })
	)
	var (
		myContractAddr  = sdk.AccAddress(rand.Bytes(wasmtypes.ContractAddrLen))
		myValidatorAddr = sdk.ValAddress(rand.Bytes(address.Len))
		myAmount        = sdk.NewInt64Coin("ALX", 1234)
		myErr           = errors.New("testing")
	)
	validBondMsg := []byte(fmt.Sprintf(
		`{"virtual_stake":{"bond":{"amount":{"denom":"ALX", "amount":"1234"},"validator":%q}}}`,
		myValidatorAddr.String()))
	validUnbondMsg := []byte(fmt.Sprintf(
		`{"virtual_stake":{"unbond":{"amount":{"denom":"ALX", "amount":"1234"},"validator":%q}}}`,
		myValidatorAddr.String()))

	specs := map[string]struct {
		src     wasmvmtypes.CosmosMsg
		auth    AuthSource
		setup   func(*testing.T) (msKeeper, func())
		expErr  error
		expData [][]byte
	}{
		"handle bond msg - success": {
			src:  wasmvmtypes.CosmosMsg{Custom: validBondMsg},
			auth: allAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				fn, asserts := captureCall(t, myContractAddr, myValidatorAddr, myAmount)
				return &msKeeperMock{DelegateFn: fn}, asserts
			},
		},
		"handle bond failed": {
			src:  wasmvmtypes.CosmosMsg{Custom: validBondMsg},
			auth: allAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				m := msKeeperMock{DelegateFn: func(_ sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error {
					return myErr
				}}
				return &m, t.FailNow
			},
			expErr: myErr,
		},
		"handle unbond msg - success": {
			src:  wasmvmtypes.CosmosMsg{Custom: validUnbondMsg},
			auth: allAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				fn, asserts := captureCall(t, myContractAddr, myValidatorAddr, myAmount)
				return &msKeeperMock{UndelegateFn: fn}, asserts
			},
		},
		"handle unbond failed": {
			src:  wasmvmtypes.CosmosMsg{Custom: validUnbondMsg},
			auth: allAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				m := msKeeperMock{UndelegateFn: func(_ sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error {
					return myErr
				}}
				return &m, t.FailNow
			},
			expErr: myErr,
		},
		"non custom msg- skip": {
			src:  wasmvmtypes.CosmosMsg{},
			auth: panicAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				return msKeeperMock{}, t.FailNow
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"non json msg": {
			src:  wasmvmtypes.CosmosMsg{Custom: []byte("not-json")},
			auth: panicAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				return msKeeperMock{}, t.FailNow
			},
			expErr: sdkerrors.ErrJSONUnmarshal,
		},
		"non mesh-sec msg - skip": {
			src:  wasmvmtypes.CosmosMsg{Custom: []byte("{}")},
			auth: panicAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				return msKeeperMock{}, t.FailNow
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"unauthorized contract": {
			src:  wasmvmtypes.CosmosMsg{Custom: validBondMsg},
			auth: noAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				return msKeeperMock{}, t.FailNow
			},
			expErr: sdkerrors.ErrUnauthorized,
		},
		"unknown mesh-sec custom msg": {
			src:  wasmvmtypes.CosmosMsg{Custom: []byte(`{"virtual_stake":{"unknown_msg":{}}}`)},
			auth: allAuthZ,
			setup: func(t *testing.T) (msKeeper, func()) {
				return msKeeperMock{}, t.FailNow
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeperMock, verify := spec.setup(t)
			h := NewCustomMsgHandler(keeperMock, spec.auth)
			var ctx sdk.Context
			_, gotData, gotErr := h.DispatchMsg(ctx, myContractAddr, "", spec.src)
			if spec.expErr != nil {
				assert.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expData, gotData)
			verify()
		})
	}
}

func captureCall(t *testing.T, myContractAddr sdk.AccAddress, myValidatorAddr sdk.ValAddress, expCoin sdk.Coin) (func(_ sdk.Context, actor sdk.AccAddress, val sdk.ValAddress, coin sdk.Coin) error, func()) {
	var (
		captureVal     sdk.ValAddress
		capturedAmount sdk.Coin
	)
	fn := func(_ sdk.Context, actor sdk.AccAddress, val sdk.ValAddress, coin sdk.Coin) error {
		require.Equal(t, myContractAddr, actor)
		captureVal = val
		capturedAmount = coin
		return nil
	}
	asserts := func() {
		assert.Equal(t, myValidatorAddr, captureVal)
		assert.Equal(t, expCoin, capturedAmount)
	}
	return fn, asserts
}

var _ msKeeper = msKeeperMock{}

type msKeeperMock struct {
	DelegateFn   func(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error
	UndelegateFn func(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error
}

func (m msKeeperMock) Delegate(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error {
	if m.DelegateFn == nil {
		panic("not expected to be called")
	}
	return m.DelegateFn(ctx, actor, addr, coin)
}

func (m msKeeperMock) Undelegate(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error {
	if m.UndelegateFn == nil {
		panic("not expected to be called")
	}
	return m.UndelegateFn(ctx, actor, addr, coin)
}
