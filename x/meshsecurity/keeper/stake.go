package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	types "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// Delegate mints new "virtual" bonding tokens and delegates them to the given validator.
// The amount minted is added to the SupplyOffset, when supported.
// Authorization of the caller should be handled before entering this method.
func (k Keeper) Delegate(pCtx sdk.Context, actor sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Coin) (sdk.Dec, error) {
	if amt.Amount.IsZero() || amt.Amount.IsNegative() {
		return sdk.ZeroDec(), errors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure staking constraints
	bondDenom := k.staking.BondDenom(pCtx)
	if amt.Denom != bondDenom {
		return sdk.ZeroDec(), errors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", amt.Denom, bondDenom)
	}
	validator, found := k.staking.GetValidator(pCtx, valAddr)
	if !found {
		return sdk.ZeroDec(), stakingtypes.ErrNoValidatorFound
	}

	// Ensure MS constraints:
	newTotalDelegatedAmount := k.getTotalDelegatedAmount(pCtx, actor).Add(amt.Amount)
	max := k.GetMaxCapLimit(pCtx, actor).Amount
	if newTotalDelegatedAmount.GT(max) {
		return sdk.ZeroDec(), types.ErrMaxCapExceeded.Wrapf("%s exceeds %s", newTotalDelegatedAmount, max)
	}

	cacheCtx, done := pCtx.CacheContext() // work in a cached store as osmosis (safety net?)

	// mint tokens as virtual coins that do not count to the total supply
	coins := sdk.NewCoins(amt)
	err := k.bank.MintCoins(cacheCtx, types.ModuleName, coins)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	k.bank.AddSupplyOffset(cacheCtx, bondDenom, amt.Amount.Neg())
	err = k.bank.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, actor, coins)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	// delegate virtual coins to the validator
	newShares, err := k.staking.Delegate(
		cacheCtx,
		actor,
		amt.Amount,
		stakingtypes.Unbonded,
		validator,
		true,
	)

	// and update our records
	k.setTotalDelegatedAmount(cacheCtx, actor, newTotalDelegatedAmount)
	done()

	// TODO: emit events?
	// TODO: add to telemetry?
	return newShares, err
}

func (k Keeper) Undelegate(pCtx sdk.Context, actor sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Coin) error {
	panic("not implemented, yet")
}
