package meshsecurity

import (
	"fmt"
	"time"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called after every block
func EndBlocker(ctx sdk.Context, k *keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	results, err := k.ExecScheduledTasks(ctx, types.SchedulerTaskRebalance, func(ctx sdk.Context, addr sdk.AccAddress) error {
		return k.Rebalance(ctx, addr)
	})
	if err != nil {
		panic(err) // todo: log or fail?
	}
	for _, r := range results {
		logger := keeper.ModuleLogger(ctx).
			With("contract", r.Contract.String())
		switch {
		case r.ExecErr != nil:
			logger.Error("failed to execute scheduled task")
		case r.RescheduleErr != nil: // todo: log or fail?
			panic(fmt.Sprintf("failed to reschedule task for contract %s", r.Contract.String()))
		case r.DeleteTaskErr != nil:
			logger.Error("failed to delete scheduled task after completion")
		default:
			logger.Info("scheduled task executed successfully", "gas_used", r.GasUsed, "gas_limit", r.GasLimit)
		}
	}
}
