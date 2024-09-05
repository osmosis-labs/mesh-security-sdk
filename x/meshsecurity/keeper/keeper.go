package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// Option is an extension point to instantiate keeper with non default values
type Option interface {
	apply(*Keeper)
}

type Keeper struct {
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey
	cdc      codec.Codec
	bank     types.XBankKeeper
	Staking  types.XStakingKeeper
	wasm     types.WasmKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper constructor with vanilla sdk keepers
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	memoryStoreKey storetypes.StoreKey,
	bank types.SDKBankKeeper,
	staking types.SDKStakingKeeper,
	wasm types.WasmKeeper,
	authority string,
	opts ...Option,
) *Keeper {
	return NewKeeperX(cdc, storeKey, memoryStoreKey, NewBankKeeperAdapter(bank), NewStakingKeeperAdapter(staking, bank), wasm, authority, opts...)
}

// NewKeeperX constructor with extended Osmosis SDK keepers
func NewKeeperX(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	memoryStoreKey storetypes.StoreKey,
	bank types.XBankKeeper,
	staking types.XStakingKeeper,
	wasm types.WasmKeeper,
	authority string,
	opts ...Option,
) *Keeper {
	k := &Keeper{
		storeKey:  storeKey,
		memKey:    memoryStoreKey,
		cdc:       cdc,
		bank:      bank,
		Staking:   staking,
		wasm:      wasm,
		authority: authority,
	}
	for _, o := range opts {
		o.apply(k)
	}

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// HasMaxCapLimit returns true when any max cap limit was set. The amount is not taken into account for the result.
// A 0 value would be true as well.
func (k Keeper) HasMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BuildMaxCapLimitKey(actor))
}

// GetMaxCapLimit the cap limit is set per consumer contract. Different providers can have different limits
// Returns zero amount when no limit is stored.
func (k Keeper) GetMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
	return sdk.NewCoin(k.Staking.BondDenom(ctx), k.mustLoadInt(ctx, k.storeKey, types.BuildMaxCapLimitKey(actor)))
}

// SetMaxCapLimit stores the max cap limit for the given contract address.
// Any existing limit for this contract will be overwritten
func (k Keeper) SetMaxCapLimit(ctx sdk.Context, contract sdk.AccAddress, newAmount sdk.Coin) error {
	if k.Staking.BondDenom(ctx) != newAmount.Denom {
		return sdkerrors.ErrInvalidCoins
	}
	// ensure that the total max cap amount for all contracts is not exceeded
	total := math.ZeroInt()
	k.IterateMaxCapLimit(ctx, func(addr sdk.AccAddress, m math.Int) bool {
		if !addr.Equals(contract) {
			total = total.Add(m)
		}
		return false
	})
	totalMaxCap := k.GetTotalContractsMaxCap(ctx)
	if total.Add(newAmount.Amount).GT(totalMaxCap.Amount) {
		return types.ErrInvalid.Wrapf("amount exceeds total available max cap (used %s of %s)", total, totalMaxCap)
	}
	// persist
	store := ctx.KVStore(k.storeKey)
	bz, err := newAmount.Amount.Marshal()
	if err != nil { // always nil
		return errorsmod.Wrap(err, "marshal amount")
	}
	store.Set(types.BuildMaxCapLimitKey(contract), bz)

	types.EmitMaxCapLimitUpdatedEvent(ctx, contract, newAmount)
	return nil
}

// GetTotalDelegated returns the total amount delegated by the given consumer contract.
// This amount can be 0 is never negative.
func (k Keeper) GetTotalDelegated(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
	v := k.mustLoadInt(ctx, k.storeKey, types.BuildTotalDelegatedAmountKey(actor))
	if v.IsNegative() {
		v = math.ZeroInt()
	}
	return sdk.NewCoin(k.Staking.BondDenom(ctx), v)
}

// internal setter. must only be used with bonding token denom or panics
func (k Keeper) setTotalDelegated(ctx sdk.Context, actor sdk.AccAddress, newAmount sdk.Coin) {
	if k.Staking.BondDenom(ctx) != newAmount.Denom {
		panic(sdkerrors.ErrInvalidCoins.Wrapf("not a staking denom: %s", newAmount.Denom))
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := newAmount.Amount.Marshal()
	if err != nil { // always nil
		panic(err)
	}
	store.Set(types.BuildTotalDelegatedAmountKey(actor), bz)
}

// GetDelegation returns contract delegation for a specified delegator bond with validator.
func (k Keeper) GetDelegation(ctx sdk.Context, actor, delAddr sdk.AccAddress, valAddr sdk.ValAddress) types.Delegation {
	store := ctx.KVStore(k.storeKey)
	key := types.BuildDelegationKey(actor, delAddr, valAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.Delegation{
			DelegatorAddress: delAddr.String(),
			ValidatorAddress: valAddr.String(),
			Amount:           math.ZeroInt(),
		}
	}
	var del types.Delegation
	if err := del.Unmarshal(bz); err != nil {
		panic(err)
	}
	return del
}

// GetAllDelegations returns all delegations for a specific contract
func (k Keeper) GetAllDelegations(ctx sdk.Context, actor sdk.AccAddress, maxRetrieve uint16) (delegations []types.Delegation) {
	delegations = make([]types.Delegation, maxRetrieve)
	store := ctx.KVStore(k.storeKey)
	contractPrefixKey := types.BuildDelegationsKey(actor)

	iterator := sdk.KVStorePrefixIterator(store, contractPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		var del types.Delegation
		if err := del.Unmarshal(iterator.Value()); err != nil {
			panic(err)
		}

		delegations[i] = del
		i++
	}

	return delegations[:i] // trim if the array length < maxRetrieve
}

// setDelegation store the delegation of a given delegator bond with validator
func (k Keeper) setDelegation(ctx sdk.Context, actor, delAddr sdk.AccAddress, valAddr sdk.ValAddress, changeAmount math.Int) {
	store := ctx.KVStore(k.storeKey)

	newDelegation := k.GetDelegation(ctx, actor, delAddr, valAddr)
	newDelegation.Amount = newDelegation.Amount.Add(changeAmount)
	if newDelegation.Amount.IsZero() {
		store.Delete(types.BuildDelegationKey(actor, delAddr, valAddr))
		return
	}

	bz, err := newDelegation.Marshal()
	if err != nil { // always nil
		panic(err)
	}
	store.Set(types.BuildDelegationKey(actor, delAddr, valAddr), bz)
}

// helper to deserialize a math.Int from store. Returns zero when key does not exist.
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

// IterateMaxCapLimit iterate over contract addresses with max cap limit set
// Callback can return true to stop early
func (k Keeper) IterateMaxCapLimit(ctx sdk.Context, cb func(sdk.AccAddress, math.Int) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.MaxCapLimitKeyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var r math.Int
		if err := r.Unmarshal(iter.Value()); err != nil {
			panic(err)
		}
		// cb returns true to stop early
		if cb(iter.Key(), r) {
			return
		}
	}
}

// ModuleLogger returns logger with module attribute
func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
