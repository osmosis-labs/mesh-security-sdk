package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VaultAddress - Address of vault contract
func (k Keeper) VaultAddress(ctx sdk.Context) string {
	return k.GetParams(ctx).VaultAddress
}