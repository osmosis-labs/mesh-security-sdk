package meshsecurity

import (
	"time"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called after every block
func EndBlocker(pCtx sdk.Context, k *keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	k.ExecScheduled(pCtx, types.SchedulerTypeRebalance, func(ctx sdk.Context, addr sdk.AccAddress) error {
		return k.Rebalance(ctx, addr)
	})
}
