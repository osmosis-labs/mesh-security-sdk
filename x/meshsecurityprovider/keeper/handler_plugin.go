package keeper

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

type CustomMsgHandler struct {
	k Keeper
}

// NewCustomMsgHandler constructor to set up CustomMsgHandler.
func NewCustomMsgHandler(k Keeper) *CustomMsgHandler {
	return &CustomMsgHandler{k: k}
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

	switch {
	case customMsg.ProviderMsg.Bond != nil:
		return h.handleBondMsg(ctx, contractAddr, customMsg.ProviderMsg.Bond)
	case customMsg.ProviderMsg.Unbond != nil:
		return h.handleUnbondMsg(ctx, contractAddr, customMsg.ProviderMsg.Unbond)
	case customMsg.ProviderMsg.Unstake != nil:
		return h.handleUnstakeMsg(ctx, contractAddr, customMsg.ProviderMsg.Unstake)
	}

	return nil, nil, wasmtypes.ErrUnknownMsg
}

func (h CustomMsgHandler) handleBondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *contract.BondMsg) ([]sdk.Event, [][]byte, error) {
	if actor.String() != h.k.VaultAddress(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

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
	if actor.String() != h.k.VaultAddress(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

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
		types.EventTypeUnbond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
	)}, nil, nil
}

func (h CustomMsgHandler) handleUnstakeMsg(ctx sdk.Context, actor sdk.AccAddress, unstakeMsg *contract.UnstakeMsg) ([]sdk.Event, [][]byte, error) {
	nativeContractAddr := h.k.NativeStakingAddress(ctx)
	var proxyRes types.ProxyByOwnerResponse

	resBytes, err := h.k.wasmKeeper.QuerySmart(ctx, 
		sdk.AccAddress(nativeContractAddr),
		[]byte(fmt.Sprintf(`{"proxy_by_owner": {"owner": "%s"}}`, actor.String())),
	)
	if err != nil {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}
	if err = json.Unmarshal(resBytes, &proxyRes); err != nil {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}
	if proxyRes.Proxy == "" {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(unstakeMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(unstakeMsg.Validator)
	if err != nil {
		return nil, nil, err
	}

	err = h.k.Unstake(ctx, actor, valAddr, coin)
	if err != nil {
		return nil, nil, err
	}

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeUnstake,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
	)}, nil, nil
}
