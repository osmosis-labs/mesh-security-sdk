package wasmbinding

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/mesh-security-sdk/wasmbinding/bindings"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type ConsumerKeeper interface {
	HandleBondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *bindings.BondMsg) ([]sdk.Event, [][]byte, error)
	HandleUnbondMsg(ctx sdk.Context, actor sdk.AccAddress, unbondMsg *bindings.UnbondMsg) ([]sdk.Event, [][]byte, error)
}

type ProviderKeeper interface {
	HandleDepositMsg(ctx sdk.Context, actor sdk.AccAddress, depositMsg *bindings.DepositMsg) ([]sdk.Event, [][]byte, error)
	HandleWithdrawMsg(ctx sdk.Context, actor sdk.AccAddress, withdrawMsg *bindings.WithdrawMsg) ([]sdk.Event, [][]byte, error)
	HandleUnstakeMsg(ctx sdk.Context, actor sdk.AccAddress, unstakeMsg *bindings.UnstakeMsg) ([]sdk.Event, [][]byte, error)
}

// CustomMessageDecorator returns decorator for custom CosmWasm bindings messages
func CustomMessageDecorator(consKeeper ConsumerKeeper, provKeeper ProviderKeeper) *CustomMessenger {
	return &CustomMessenger{
		consKeeper: consKeeper,
		provKeeper: provKeeper,
	}
}

type CustomMessenger struct {
	consKeeper ConsumerKeeper
	provKeeper ProviderKeeper
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

// DispatchMsg executes on the contractMsg.
func (h CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom != nil {
		var contractMsg bindings.CustomMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			return nil, nil, errorsmod.Wrap(err, "mesh security msg")
		}

		if contractMsg.Bond != nil {
			return h.consKeeper.HandleBondMsg(ctx, contractAddr, contractMsg.Bond)
		}
		if contractMsg.Unbond != nil {
			return h.consKeeper.HandleUnbondMsg(ctx, contractAddr, contractMsg.Unbond)
		}

		if contractMsg.Deposit != nil {
			return h.provKeeper.HandleDepositMsg(ctx, contractAddr, contractMsg.Deposit)
		}
		if contractMsg.Withdraw != nil {
			return h.provKeeper.HandleWithdrawMsg(ctx, contractAddr, contractMsg.Withdraw)
		}
		if contractMsg.Unstake != nil {
			return h.provKeeper.HandleUnstakeMsg(ctx, contractAddr, contractMsg.Unstake)
		}
	}
	return nil, nil, wasmtypes.ErrUnknownMsg
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
