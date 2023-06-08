package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var _ types.WasmKeeper = &MockWasmKeeper{}

type MockWasmKeeper struct {
	SudoFn            func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	HasContractInfoFn func(ctx sdk.Context, contractAddress sdk.AccAddress) bool
}

func (m MockWasmKeeper) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	if m.SudoFn == nil {
		panic("not expected to be called")
	}
	return m.SudoFn(ctx, contractAddress, msg)
}

func (m MockWasmKeeper) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	if m.HasContractInfoFn == nil {
		panic("not expected to be called")
	}
	return m.HasContractInfoFn(ctx, contractAddress)
}
