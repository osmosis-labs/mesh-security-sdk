package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	types "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// Delegate mints new "virtual" bonding tokens and delegates them to the given validator.
// The amount minted is removed from the SupplyOffset (so that it will become negative), when supported.
// Authorization of the actor should be handled before entering this method.
func (k Keeper) Delegate(pCtx sdk.Context, actor sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Coin) (sdk.Dec, error) {
	if amt.Amount.IsNil() || amt.Amount.IsZero() || amt.Amount.IsNegative() {
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
	newTotalDelegatedAmount := k.GetTotalDelegated(pCtx, actor).Add(amt)
	max := k.GetMaxCapLimit(pCtx, actor)
	if max.IsLT(newTotalDelegatedAmount) {
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
	k.setTotalDelegated(cacheCtx, actor, newTotalDelegatedAmount)
	done()
	return newShares, err
}

// Undelegate executes an instant undelegate and burns the released virtual staking tokens.
// The amount burned is added to the (negative) SupplyOffset, when supported.
// Authorization of the actor should be handled before entering this method.
func (k Keeper) Undelegate(pCtx sdk.Context, actor sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Coin) error {
	if amt.Amount.IsNil() || amt.Amount.IsZero() || amt.Amount.IsNegative() {
		return errors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure staking constraints
	bondDenom := k.staking.BondDenom(pCtx)
	if amt.Denom != bondDenom {
		return errors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", amt.Denom, bondDenom)
	}

	cacheCtx, done := pCtx.CacheContext() // work in a cached store (safety net?)
	totalDelegatedAmount := k.GetTotalDelegated(cacheCtx, actor)
	if totalDelegatedAmount.IsLT(amt) {
		return errors.ErrInvalidRequest.Wrap("amount exceeds total delegated")
	}
	shares, err := k.staking.ValidateUnbondAmount(cacheCtx, actor, valAddr, amt.Amount)
	if err == stakingtypes.ErrNoDelegation {
		return nil
	} else if err != nil {
		return err
	}

	undelegatedCoins, err := k.staking.InstantUndelegate(cacheCtx, actor, valAddr, shares)
	if err != nil {
		return err
	}
	err = k.bank.SendCoinsFromAccountToModule(cacheCtx, actor, types.ModuleName, undelegatedCoins)
	if err != nil {
		return err
	}

	err = k.bank.BurnCoins(cacheCtx, types.ModuleName, undelegatedCoins)
	if err != nil {
		return err
	}

	unbondedAmount := sdk.NewCoin(bondDenom, undelegatedCoins.AmountOf(bondDenom))
	k.bank.AddSupplyOffset(cacheCtx, bondDenom, unbondedAmount.Amount)
	newDelegatedAmt := totalDelegatedAmount.Sub(unbondedAmount)
	if newDelegatedAmt.IsNegative() {
		newDelegatedAmt = sdk.NewCoin(bondDenom, math.ZeroInt())
	}
	k.setTotalDelegated(cacheCtx, actor, newDelegatedAmt)

	done()
	return nil
}
