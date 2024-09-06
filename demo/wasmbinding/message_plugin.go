package contract

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consumermsg "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

type ConsumerKeeper interface {
	HandleBondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *consumermsg.BondMsg) ([]sdk.Event, [][]byte, error)
	HandleUnbondMsg(ctx sdk.Context, actor sdk.AccAddress, unbondMsg *consumermsg.UnbondMsg) ([]sdk.Event, [][]byte, error)
}

type CustomMessenger struct {
	consKeeper ConsumerKeeper
}

// DispatchMsg executes on the contractMsg.
func (h CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		var contractMsg CustomMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, errorsmod.Wrap(err, "mesh security msg")
		}

		if contractMsg.Bond != nil {
			return h.consKeeper.HandleBondMsg(ctx, contractAddr, contractMsg.Bond)
		}
		if contractMsg.Unbond != nil {
			return h.consKeeper.HandleUnbondMsg(ctx, contractAddr, contractMsg.Unbond)
		}
	}
	return nil, nil, wasmtypes.ErrUnknownMsg
}
