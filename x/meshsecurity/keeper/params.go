package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// SetParams sets the module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams gets the module's parameters.
func (k Keeper) GetParams(clientCtx sdk.Context) (params types.Params) {
	store := clientCtx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) GetMaxSudoGas(ctx sdk.Context) sdk.Gas {
	return sdk.Gas(k.GetParams(ctx).MaxGasEndBlocker)
}

func (k Keeper) GetRebalanceEpochLength(ctx sdk.Context) uint64 {
	return uint64(k.GetParams(ctx).EpochLength)
}

func (k Keeper) GetTotalContractsMaxCap(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).TotalContractsMaxCap
}
