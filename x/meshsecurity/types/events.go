package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypeSchedulerExec       = "scheduler_execution"
	EventTypeSchedulerRegistered = "scheduler_registered"
	EventTypeMaxCapLimitUpdated  = "max_cap_limit_updated"
	EventTypeUnbond              = "instant_unbond"
	EventTypeDelegate            = "instant_delegate"
)

const (
	AttributeKeyContractAddr         = "virtual_staking_contract"
	AttributeKeySchedulerNextExec    = "next_exececution_block"
	AttributeKeySchedulerExecSuccess = "execution_success"
	AttributeKeySchedulerRepeat      = "repeat"
	AttributeKeySchedulerExecError   = "error"
	AttributeKeyValidator            = "validator"
	AttributeKeyDelegator            = "delegator"
)

// EmitSchedulerExecutionEvent emits an event signalling a successful or failed scheduler execution and including the error
// details if any.
func EmitSchedulerExecutionEvent(ctx sdk.Context, contractAddr sdk.AccAddress, err error) {
	success := err == nil
	attributes := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		sdk.NewAttribute(AttributeKeyContractAddr, contractAddr.String()),
		sdk.NewAttribute(AttributeKeySchedulerExecSuccess, fmt.Sprintf("%t", success)),
	}

	if err != nil {
		attributes = append(attributes, sdk.NewAttribute(AttributeKeySchedulerExecError, err.Error()))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeSchedulerExec,
			attributes...,
		),
	)
}

// EmitSchedulerRegisteredEvent emits an event signalling a new scheduler execution is registered
func EmitSchedulerRegisteredEvent(ctx sdk.Context, contractAddr sdk.AccAddress, nextExecBlock uint64, repeat bool) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeSchedulerRegistered,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(AttributeKeyContractAddr, contractAddr.String()),
			sdk.NewAttribute(AttributeKeySchedulerNextExec, fmt.Sprintf("%d", nextExecBlock)),
			sdk.NewAttribute(AttributeKeySchedulerRepeat, fmt.Sprintf("%t", repeat)),
		),
	)
}

// EmitMaxCapLimitUpdatedEvent emits an event signalling that max cap limit is updated
func EmitMaxCapLimitUpdatedEvent(ctx sdk.Context, contractAddr sdk.AccAddress, amount sdk.Coin) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeMaxCapLimitUpdated,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(AttributeKeyContractAddr, contractAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
		),
	)
}
