package keeper

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

type (
	// abstract query keeper
	viewKeeper interface {
		GetMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
		GetTotalDelegated(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin
	}
	stakingKeeper interface {
		BondDenom(ctx sdk.Context) string
		Validator(sdk.Context, sdk.ValAddress) stakingtypes.ValidatorI
		Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) stakingtypes.DelegationI
	}
	slashingKeeper interface {
		SlashFractionDoubleSign(ctx sdk.Context) (res sdk.Dec)
		SlashFractionDowntime(ctx sdk.Context) (res sdk.Dec)
	}
)

// NewQueryDecorator constructor to build a chained custom querier.
// The mesh-security custom query handler is placed at the first position
// and delegates to the next in chain for any queries that do not match
// the mesh-security custom query namespace.
//
// To be used with `wasmkeeper.WithQueryHandlerDecorator(meshseckeeper.NewQueryDecorator(app.MeshSecKeeper)))`
func NewQueryDecorator(k viewKeeper, stk stakingKeeper, slk slashingKeeper) func(wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
	return func(next wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
		return ChainedCustomQuerier(k, stk, slk, next)
	}
}

// ChainedCustomQuerier implements the mesh-security custom query handler.
// The given WasmVMQueryHandler is receiving all unhandled queries and must therefore
// not be nil.
//
// This CustomQuerier is designed as an extension point. See the NewQueryDecorator impl how to
// set this up for wasmd.
func ChainedCustomQuerier(k viewKeeper, stk stakingKeeper, slk slashingKeeper, next wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
	if k == nil {
		panic("ms keeper must not be nil")
	}
	if stk == nil {
		panic("staking Keeper must not be nil")
	}
	if slk == nil {
		panic("slashing Keeper must not be nil")
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
		query := contractQuery.VirtualStake
		if query == nil {
			return next.HandleQuery(ctx, caller, request)
		}

		var res any
		switch {
		case query.BondStatus != nil:
			contractAddr, err := sdk.AccAddressFromBech32(query.BondStatus.Contract)
			if err != nil {
				return nil, sdkerrors.ErrInvalidAddress.Wrap(query.BondStatus.Contract)
			}
			res = contract.BondStatusResponse{
				MaxCap:    wasmkeeper.ConvertSdkCoinToWasmCoin(k.GetMaxCapLimit(ctx, contractAddr)),
				Delegated: wasmkeeper.ConvertSdkCoinToWasmCoin(k.GetTotalDelegated(ctx, contractAddr)),
			}
		case query.SlashRatio != nil:
			res = contract.SlashRatioResponse{
				SlashFractionDowntime:   slk.SlashFractionDowntime(ctx).String(),
				SlashFractionDoubleSign: slk.SlashFractionDoubleSign(ctx).String(),
			}
		case query.TotalDelegation != nil:
			contractAddr, err := sdk.AccAddressFromBech32(query.TotalDelegation.Contract)
			if err != nil {
				return nil, sdkerrors.ErrInvalidAddress.Wrap(query.TotalDelegation.Contract)
			}
			valAddr, err := sdk.ValAddressFromBech32(query.TotalDelegation.Validator)
			if err != nil {
				return nil, sdkerrors.ErrInvalidAddress.Wrap(query.TotalDelegation.Validator)
			}

			totalShares := stk.Delegation(ctx, contractAddr, valAddr).GetShares()
			amount := stk.Validator(ctx, valAddr).TokensFromShares(totalShares).TruncateInt()
			totalDelegation := sdk.NewCoin(stk.BondDenom(ctx), amount)
			res = contract.TotalDelegationResponse{
				Delegation: wasmkeeper.ConvertSdkCoinToWasmCoin(totalDelegation),
			}
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown virtual_stake query variant"}
		}
		return json.Marshal(res)
	})
}

var _ wasmkeeper.WasmVMQueryHandler = QueryHandlerFn(nil)

// QueryHandlerFn helper type that implements wasmkeeper.WasmVMQueryHandler
type QueryHandlerFn func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error)

// HandleQuery handles contract query
func (q QueryHandlerFn) HandleQuery(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
	return q(ctx, caller, request)
}
