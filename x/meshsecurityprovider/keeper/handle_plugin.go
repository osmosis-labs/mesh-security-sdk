package keeper

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
)

type CustomMessenger struct {
	k *Keeper
}

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func CustomMessageDecorator(provKeeper *Keeper) *CustomMessenger {
	return &CustomMessenger{
		k: provKeeper,
	}
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg executes on the contractMsg.
func (h CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom == nil {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	var customMsg contract.CustomMsg
	if err := json.Unmarshal(msg.Custom, &customMsg); err != nil {
		return nil, nil, sdkerrors.ErrJSONUnmarshal.Wrap("custom message")
	}
	if customMsg.Provider == nil {
		// not our message type
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	switch {
	case customMsg.Provider.Bond != nil:
		return h.k.HandleBondMsg(ctx, contractAddr, customMsg.Provider.Bond)
	case customMsg.Provider.Unbond != nil:
		return h.k.HandleUnbondMsg(ctx, contractAddr, customMsg.Provider.Unbond)
	case customMsg.Register != nil:
		return h.k.HandleRegistryConsumer(ctx, customMsg.Register.ChannelID, contractAddr)
	}
	return nil, nil, wasmtypes.ErrUnknownMsg
}
