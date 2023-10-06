package keeper

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

// abstract query keeper
type viewKeeper interface {
	GetMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
	GetTotalDelegated(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
}

// NewQueryDecorator constructor to build a chained custom querier.
// The mesh-security custom query handler is placed at the first position
// and delegates to the next in chain for any queries that do not match
// the mesh-security custom query namespace.
//
// To be used with `wasmkeeper.WithQueryHandlerDecorator(meshseckeeper.NewQueryDecorator(app.MeshSecKeeper)))`
func NewQueryDecorator(k viewKeeper) func(wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
	return func(next wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
		return ChainedCustomQuerier(k, next)
	}
}

// ChainedCustomQuerier implements the mesh-security custom query handler.
// The given WasmVMQueryHandler is receiving all unhandled queries and must therefore
// not be nil.
//
// This CustomQuerier is designed as an extension point. See the NewQueryDecorator impl how to
// set this up for wasmd.
func ChainedCustomQuerier(k viewKeeper, next wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
	if k == nil {
		panic("keeper must not be nil")
	}
	if next == nil {
		panic("next handler must not be nil")
	}
	return QueryHandlerFn(func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
		if request.Custom == nil {
			return next.HandleQuery(ctx, caller, request)
		}
		var contractQuery contract.CustomQuery
		if err := json.Unmarshal(request.Custom, &contractQuery); err != nil {
			return nil, errorsmod.Wrap(err, "mesh-security query")
		}
		if contractQuery.VirtualStake == nil || contractQuery.VirtualStake.BondStatus == nil {
			return next.HandleQuery(ctx, caller, request)
		}
		contractAddr, err := sdk.AccAddressFromBech32(contractQuery.VirtualStake.BondStatus.Contract)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrap(contractQuery.VirtualStake.BondStatus.Contract)
		}
		res := contract.BondStatusResponse{
			MaxCap:    wasmkeeper.ConvertSdkCoinToWasmCoin(k.GetMaxCapLimit(ctx, contractAddr)),
			Delegated: wasmkeeper.ConvertSdkCoinToWasmCoin(k.GetTotalDelegated(ctx, contractAddr)),
		}
		bz, err := json.Marshal(res)
		if err != nil {
			return nil, errorsmod.Wrap(err, "mesh-security max cap query response")
		}
		return bz, nil
	})
}

var _ wasmkeeper.WasmVMQueryHandler = QueryHandlerFn(nil)

// QueryHandlerFn helper type that implements wasmkeeper.WasmVMQueryHandler
type QueryHandlerFn func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error)

// HandleQuery handles contract query
func (q QueryHandlerFn) HandleQuery(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
	return q(ctx, caller, request)
}
