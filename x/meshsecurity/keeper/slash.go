package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) HandleBeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, slashRatio sdk.Dec) error {
	if valAddr == nil {
		ModuleLogger(ctx).
			Error("can not propagate slash: validator not found", "validator", valAddr.String())
	}
	if err := k.ScheduleSlashed(ctx, valAddr, slashRatio); err != nil {
		ModuleLogger(ctx).
			Error("can not propagate slash: schedule event",
				"cause", err,
				"validator", valAddr.String())
	}
	if err := k.ScheduleJailed(ctx, valAddr); err != nil {
		ModuleLogger(ctx).
			Error("can not propagate jail: schedule event",
				"cause", err,
				"validator", valAddr.String())
	}

	return nil
}
