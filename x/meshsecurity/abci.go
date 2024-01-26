package meshsecurity

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// TaskExecutionResponseHandler is an extension point for custom implementations
type TaskExecutionResponseHandler interface {
	Handle(ctx sdk.Context, e keeper.ExecResult)
}

// EndBlocker is called after every block
func EndBlocker(ctx sdk.Context, k *keeper.Keeper, h TaskExecutionResponseHandler) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	do := rspHandler(ctx, h)
	epochLength := k.GetRebalanceEpochLength(ctx)
	do(k.ExecScheduledTasks(ctx, types.SchedulerTaskValsetUpdate, epochLength, func(ctx sdk.Context, contract sdk.AccAddress) error {
		report, err := k.ValsetUpdateReport(ctx)
		if err != nil {
			return err
		}
		// If there was a slashing event, multiply the total slash amount by the delegator shares ratio for `contract`
		if report.Slashed != nil {
			for i, slash := range report.Slashed {
				valAddr, err := sdk.ValAddressFromBech32(slash.ValidatorAddr)
				if err != nil {
					return fmt.Errorf("invalid validator address %s", slash.ValidatorAddr)
				}
				totalSlashAmount, ok := sdkmath.NewIntFromString(slash.SlashAmount)
				if !ok {
					return fmt.Errorf("invalid slash amount %s", slash.SlashAmount)
				}
				// Get total validator shares
				validator, err := k.Staking.GetValidator(ctx, valAddr)
				if err != nil {
					return fmt.Errorf("validator %s not found", slash.ValidatorAddr)
				}
				validatorShares := validator.GetDelegatorShares()

				delegatorSlashAmount := sdkmath.LegacyZeroDec()
				if !validatorShares.IsZero() {
					// Query the `contract` delegation
					delegation, err := k.Staking.GetDelegation(ctx, contract, valAddr)
					delegatorShares := sdkmath.LegacyZeroDec()
					if err != nil {
						delegatorShares = delegation.GetShares()
					}
					delegatorSlashAmount = delegatorShares.Quo(validatorShares).MulInt(totalSlashAmount)
				}

				// Pass it to the contract
				// FIXME? Remove entries with zero slash amounts from the Slashed array
				// TODO: Convert to Coin
				report.Slashed[i].SlashAmount = delegatorSlashAmount.RoundInt().String()
			}
		}
		return k.SendValsetUpdate(ctx, contract, report)
	}))
	k.ClearPipedValsetOperations(ctx)
	do(k.ExecScheduledTasks(ctx, types.SchedulerTaskHandleEpoch, epochLength, func(ctx sdk.Context, contract sdk.AccAddress) error {
		return k.SendHandleEpoch(ctx, contract)
	}))
}

func rspHandler(ctx sdk.Context, h TaskExecutionResponseHandler) func(results []keeper.ExecResult, err error) {
	return func(results []keeper.ExecResult, err error) {
		if err != nil {
			panic(fmt.Sprintf("task scheduler: %s", err)) // todo: log or fail?
		}
		for _, r := range results {
			h.Handle(ctx, r)
		}
	}
}

var _ TaskExecutionResponseHandler = TaskExecutionResponseHandlerFn(nil)

// TaskExecutionResponseHandlerFn helper type that implements TaskExecutionResponseHandler
type TaskExecutionResponseHandlerFn func(ctx sdk.Context, e keeper.ExecResult)

func (h TaskExecutionResponseHandlerFn) Handle(ctx sdk.Context, e keeper.ExecResult) {
	h(ctx, e)
}

// DefaultExecutionResponseHandler default implementation that panics on reschedule errors but otherwise logs only
// TODO: revisit, is this a good default?
func DefaultExecutionResponseHandler() TaskExecutionResponseHandlerFn {
	return func(ctx sdk.Context, r keeper.ExecResult) {
		logger := keeper.ModuleLogger(ctx).
			With("contract", r.Contract.String())
		switch {
		case r.ExecErr != nil:
			logger.Error("failed to execute scheduled task", "cause", r.ExecErr)
		case r.RescheduleErr != nil:
			panic(fmt.Sprintf("failed to reschedule task for contract %q: %s", r.Contract.String(), r.RescheduleErr))
		case r.DeleteTaskErr != nil:
			logger.Error("failed to delete scheduled task after completion", "cause", r.ExecErr)
		default:
			logger.Info("scheduled task executed successfully", "gas_used", r.GasUsed,
				"gas_limit", r.GasLimit)
		}
	}
}

// PanicOnErrorExecutionResponseHandler is an alternative TaskExecutionResponseHandler implementation that always panics on errors
func PanicOnErrorExecutionResponseHandler() TaskExecutionResponseHandlerFn {
	return func(ctx sdk.Context, r keeper.ExecResult) {
		switch {
		case r.ExecErr != nil:
			panic(fmt.Sprintf("failed to execute scheduled task for contract %q: %s", r.Contract.String(), r.ExecErr))
		case r.RescheduleErr != nil:
			panic(fmt.Sprintf("failed to reschedule task for contract %q: %s", r.Contract.String(), r.RescheduleErr))
		case r.DeleteTaskErr != nil:
			panic(fmt.Sprintf("failed to delete scheduled task after completion for contract %q: %s", r.Contract.String(), r.ExecErr))
		default:
			keeper.ModuleLogger(ctx).
				With("contract", r.Contract.String()).
				Info("scheduled task executed successfully", "gas_used", r.GasUsed,
					"gas_limit", r.GasLimit)
		}
	}
}
