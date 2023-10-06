package keeper

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
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

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeDelegate,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyDelegator, actor.String()),
	)}, nil, nil
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

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeUnbond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(sdk.AttributeKeySender, actor.String()),
	)}, nil, nil
}

// AuthSourceFn is helper for simple AuthSource types
type AuthSourceFn func(ctx sdk.Context, contractAddr sdk.AccAddress) bool

// IsAuthorized returns if the contract authorized to execute a virtual stake message
func (a AuthSourceFn) IsAuthorized(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
	return a(ctx, contractAddr)
}

// abstract keeper
type maxCapSource interface {
	HasMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) bool
}

// NewIntegrityHandler prevents any contract with max cap set to use staking
// or stargate messages. This ensures that staked "virtual" tokens are not bypassing
// the instant undelegate and burn mechanism provided by mesh-security.
//
// This handler should be chained before any other.
func NewIntegrityHandler(k maxCapSource) wasmkeeper.MessageHandlerFunc {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (
		events []sdk.Event,
		data [][]byte,
		err error,
	) {
		if msg.Stargate == nil && msg.Staking == nil ||
			!k.HasMaxCapLimit(ctx, contractAddr) {
			return nil, nil, wasmtypes.ErrUnknownMsg // pass down the chain
		}
		// reject
		return nil, nil, types.ErrUnsupported.Wrap("message type for contracts with max cap set")
	}
}
