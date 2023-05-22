package keeper

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec
	bank     types.XBankKeeper
	staking  types.XStakingKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper constructor with vanilla sdk keepers
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bank types.SDKBankKeeper,
	staking types.SDKStakingKeeper,
	authority string,
) *Keeper {
	return NewKeeperX(cdc, storeKey, NewBankKeeperAdapter(bank), NewStakingKeeperAdapter(staking, bank), authority)
}

// NewKeeperX constructor with extended Osmosis SDK keepers
func NewKeeperX(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bank types.XBankKeeper,
	staking types.XStakingKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		bank:      bank,
		staking:   staking,
		authority: authority,
	}
}

// GetParams returns the total set of wasm parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	// todo: impl
	//	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets all wasm parameters.
func (k Keeper) SetParams(ctx sdk.Context, ps types.Params) error {
	if err := ps.ValidateBasic(); err != nil {
		return err
	}
	// todo: impl
	//store := ctx.KVStore(k.storeKey)
	//bz, err := k.cdc.Marshal(&ps)
	//if err != nil {
	//	return err
	//}
	//store.Set(types.ParamsKey, bz)
	return nil
}

func (k Keeper) HasMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BuildMaxCapLimitKey(actor))
}

// getMaxCapLimit the cap limit is set per consumer contract. Different providers can have different limits
func (k Keeper) getMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) math.Int {
	return k.mustLoadInt(ctx, k.storeKey, types.BuildMaxCapLimitKey(actor))
}

// getTotalDelegatedAmount returns the total amount delegated by the given consumer contract
func (k Keeper) getTotalDelegatedAmount(ctx sdk.Context, actor sdk.AccAddress) math.Int {
	return k.mustLoadInt(ctx, k.storeKey, types.BuildTotalDelegatedAmountKey(actor))
}

// helper to deserialize a math.Int from store.
// Panics when Unmarshal fails
func (k Keeper) mustLoadInt(ctx sdk.Context, storeKey storetypes.StoreKey, key []byte) math.Int {
	store := ctx.KVStore(storeKey)
	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroInt()
	}
	var r math.Int
	if err := r.Unmarshal(bz); err != nil {
		panic(err)
	}
	return r
}

func (k Keeper) setTotalDelegatedAmount(ctx sdk.Context, actor sdk.AccAddress, newAmount math.Int) {
	store := ctx.KVStore(k.storeKey)
	bz, err := newAmount.Marshal()
	if err != nil { // always nil
		panic(err)
	}
	store.Set(types.BuildTotalDelegatedAmountKey(actor), bz)
}
