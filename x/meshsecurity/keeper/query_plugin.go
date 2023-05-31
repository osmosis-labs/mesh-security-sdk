package keeper

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

// abstract query keeper
type viewKeeper interface {
	GetMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
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
	return QueryHandlerFn(func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
		if request.Custom == nil {
			return next.HandleQuery(ctx, caller, request)
		}
		var contractQuery contract.CustomQuery
		if err := json.Unmarshal(request.Custom, &contractQuery); err != nil {
			return nil, errorsmod.Wrap(err, "mesh-security query")
		}
		if contractQuery.VirtualStake == nil {
			return next.HandleQuery(ctx, caller, request)
		}
		maxCapLimit := k.GetMaxCapLimit(ctx, caller)
		res := contract.MaxCapResponse{
			MaxCap: wasmkeeper.ConvertSdkCoinToWasmCoin(maxCapLimit),
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
