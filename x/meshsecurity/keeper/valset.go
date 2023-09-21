package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) ScheduleTombstoned(ctx sdk.Context, operator sdk.ValAddress) error {
	return nil
}

func (k Keeper) ScheduleUnboding(ctx sdk.Context, addr sdk.ValAddress) error {
	return nil
}
