package keeper

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

// AuthSource abstract type that provides contract authorization.
// This is an extension point for custom implementations.
type AuthSource interface {
	// IsAuthorized returns if the contract authorized to execute a virtual stake message
	IsAuthorized(ctx sdk.Context, contractAddr sdk.AccAddress) bool
}

// abstract keeper
type msKeeper interface {
	Bond(ctx sdk.Context, actor sdk.AccAddress, delegator sdk.AccAddress, coin sdk.Coin) error
	Unbond(ctx sdk.Context, actor sdk.AccAddress, delegator sdk.AccAddress, coin sdk.Coin) error
}

type CustomMsgHandler struct {
	k    msKeeper
	auth AuthSource
}

// NewDefaultCustomMsgHandler constructor to set up the CustomMsgHandler with vault contract authorization
func NewDefaultCustomMsgHandler(k *Keeper) *CustomMsgHandler {
	return &CustomMsgHandler{k: k, auth: isVaultContract(k)}
}

func isVaultContract(k *Keeper) AuthSourceFn {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
		return k.VaultAddress(ctx) == contractAddr.String()
	}
}

// NewCustomMsgHandler constructor to set up CustomMsgHandler with an individual auth source.
// This is an extension point for non default contract authorization logic.
func NewCustomMsgHandler(k msKeeper, auth AuthSource) *CustomMsgHandler {
	return &CustomMsgHandler{k: k, auth: auth}
}

// DispatchMsg handle contract message of type Custom in the mesh-security namespace
func (h CustomMsgHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom == nil {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	var customMsg contract.CustomMsg
	if err := json.Unmarshal(msg.Custom, &customMsg); err != nil {
		return nil, nil, sdkerrors.ErrJSONUnmarshal.Wrap("custom message")
	}
	if customMsg.ProviderMsg == nil {
		// not our message type
		return nil, nil, wasmtypes.ErrUnknownMsg
	}

	if !h.auth.IsAuthorized(ctx, contractAddr) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	switch {
	case customMsg.ProviderMsg.Bond != nil:
		return h.handleBondMsg(ctx, contractAddr, customMsg.ProviderMsg.Bond)
	case customMsg.ProviderMsg.Unbond != nil:
		return h.handleUnbondMsg(ctx, contractAddr, customMsg.ProviderMsg.Unbond)
	}

	return nil, nil, wasmtypes.ErrUnknownMsg
}

func (h CustomMsgHandler) handleBondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *contract.BondMsg) ([]sdk.Event, [][]byte, error) {
	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(bondMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	delAddr, err := sdk.AccAddressFromBech32(bondMsg.Delegator)
	if err != nil {
		return nil, nil, err
	}

	err = h.k.Bond(ctx, actor, delAddr, coin)
	if err != nil {
		return nil, nil, err
	}

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeBond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
	)}, nil, nil
}

func (h CustomMsgHandler) handleUnbondMsg(ctx sdk.Context, actor sdk.AccAddress, unbondMsg *contract.UnbondMsg) ([]sdk.Event, [][]byte, error) {
	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(unbondMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	delAddr, err := sdk.AccAddressFromBech32(unbondMsg.Delegator)
	if err != nil {
		return nil, nil, err
	}

	err = h.k.Unbond(ctx, actor, delAddr, coin)
	if err != nil {
		return nil, nil, err
	}

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeBond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
	)}, nil, nil
}

// AuthSourceFn is helper for simple AuthSource types
type AuthSourceFn func(ctx sdk.Context, contractAddr sdk.AccAddress) bool

// IsAuthorized returns if the contract authorized to execute a stake message
func (a AuthSourceFn) IsAuthorized(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
	return a(ctx, contractAddr)
}
