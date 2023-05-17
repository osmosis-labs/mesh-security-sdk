package keeper

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type msKeeper interface {
	IsAuthorized(ctx sdk.Context, contractAddr sdk.AccAddress) (bool, error)
	Delegate(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error
	Undelegate(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error
}

type CustomMsgHandler struct {
	k msKeeper
}

func (h CustomMsgHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom == nil {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	var customMsg contract.CustomMsg
	if err := json.Unmarshal(msg.Custom, &customMsg); err != nil {
		return nil, nil, sdkerrors.ErrJSONUnmarshal.Wrap("custom message")
	}
	if customMsg.VirtualStake == nil {
		// not our message type
		return nil, nil, wasmtypes.ErrUnknownMsg
	}

	ok, err := h.k.IsAuthorized(ctx, contractAddr)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	switch {
	case customMsg.VirtualStake.Bond != nil:
		events, i, err := h.handleBondMsg(ctx, contractAddr, customMsg.VirtualStake.Bond)
		if err != nil {
			return events, i, err
		}
	case customMsg.VirtualStake.Unbond != nil:
		events, i, err := h.handleUnbondMsg(ctx, contractAddr, customMsg.VirtualStake.Unbond)
		if err != nil {
			return events, i, err
		}
	}
	return nil, nil, wasmtypes.ErrUnknownMsg
}

func (h CustomMsgHandler) handleBondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *contract.BondMsg) ([]sdk.Event, [][]byte, error) {
	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(bondMsg.Amount)
	if err != nil {
		return nil, nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(bondMsg.Validator)
	if err != nil {
		return nil, nil, err
	}
	err = h.k.Delegate(ctx, actor, valAddr, coin)
	if err != nil {
		return nil, nil, err
	}
	// todo: events here?
	// todo: response data format?
	return []sdk.Event{}, [][]byte{}, nil
}

func (h CustomMsgHandler) handleUnbondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *contract.UnbondMsg) ([]sdk.Event, [][]byte, error) {
	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(bondMsg.Amount)
	if err != nil {
		return nil, nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(bondMsg.Validator)
	if err != nil {
		return nil, nil, err
	}
	err = h.k.Undelegate(ctx, actor, valAddr, coin)
	if err != nil {
		return nil, nil, err
	}
	// todo: events here?
	// todo: response data format?
	return []sdk.Event{}, [][]byte{}, nil
}
