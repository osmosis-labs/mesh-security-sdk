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

// AuthSource abstract type that provides contract authorization.
// This is an extension point for custom implementations.
type AuthSource interface {
	// IsAuthorized returns if the contract authorized to execute a virtual stake message
	IsAuthorized(ctx sdk.Context, contractAddr sdk.AccAddress) bool
}

// abstract keeper
type msKeeper interface {
	Delegate(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) (sdk.Dec, error)
	Undelegate(ctx sdk.Context, actor sdk.AccAddress, addr sdk.ValAddress, coin sdk.Coin) error
}

type CustomMsgHandler struct {
	k    msKeeper
	auth AuthSource
}

// NewDefaultCustomMsgHandler constructor to set up the CustomMsgHandler with default max cap authorization
func NewDefaultCustomMsgHandler(k *Keeper) *CustomMsgHandler {
	return &CustomMsgHandler{k: k, auth: defaultMaxCapAuthorizator(k)}
}

// NewCustomMsgHandler constructor to set up CustomMsgHandler with an individual auth source.
// This is an extension point for non default contract authorization logic.
func NewCustomMsgHandler(k msKeeper, auth AuthSource) *CustomMsgHandler {
	return &CustomMsgHandler{k: k, auth: auth}
}

// default authorization logic that ensures any max cap limit was set. It does not take the amount into account
// as contracts with a limit 0 tokens may need to instant undelegate or run other operations.
// Safety mechanisms for these operations need to be placed on the implementation side.
func defaultMaxCapAuthorizator(k *Keeper) AuthSourceFn {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
		return k.HasMaxCapLimit(ctx, contractAddr)
	}
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
	if customMsg.VirtualStake == nil {
		// not our message type
		return nil, nil, wasmtypes.ErrUnknownMsg
	}

	if !h.auth.IsAuthorized(ctx, contractAddr) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	switch {
	case customMsg.VirtualStake.Bond != nil:
		return h.handleBondMsg(ctx, contractAddr, customMsg.VirtualStake.Bond)
	case customMsg.VirtualStake.Unbond != nil:
		return h.handleUnbondMsg(ctx, contractAddr, customMsg.VirtualStake.Unbond)
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
	_, err = h.k.Delegate(ctx, actor, valAddr, coin)
	if err != nil {
		return nil, nil, err
	}
	// todo: events here?
	// todo: response data format?
	return []sdk.Event{}, nil, nil
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
	return []sdk.Event{}, nil, nil
}

// AuthSourceFn is helper for simple AuthSource types
type AuthSourceFn func(ctx sdk.Context, contractAddr sdk.AccAddress) bool

// IsAuthorized returns if the contract authorized to execute a virtual stake message
func (a AuthSourceFn) IsAuthorized(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
	return a(ctx, contractAddr)
}
