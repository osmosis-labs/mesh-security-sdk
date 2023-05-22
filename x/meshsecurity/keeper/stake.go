package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	types "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func (k Keeper) Delegate(pCtx sdk.Context, actor sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Coin) error {
	if amt.Amount.IsZero() || amt.Amount.IsNegative() {
		return errors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure MS constraints:
	newTotalDelegatedAmount := k.getTotalDelegatedAmount(pCtx, actor).Add(amt.Amount)
	if newTotalDelegatedAmount.GT(k.getMaxCapLimit(pCtx, actor)) {
		return types.ErrMaxCapExceeded
	}

	// Ensure staking constraints
	bondDenom := k.staking.BondDenom(pCtx)
	if amt.Denom != bondDenom {
		return errors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", amt.Denom, bondDenom)
	}
	validator, found := k.staking.GetValidator(pCtx, valAddr)
	if !found {
		return stakingtypes.ErrNoValidatorFound
	}

	cacheCtx, done := pCtx.CacheContext() // work in a cached store as osmosis (safety net?)

	// mint tokens as virtual coins that do not count to the total supply
	coins := sdk.NewCoins(amt)
	err := k.bank.MintCoins(cacheCtx, types.ModuleName, coins)
	if err != nil {
		return err
	}
	k.bank.AddSupplyOffset(cacheCtx, bondDenom, amt.Amount.Neg())
	err = k.bank.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, actor, coins)
	if err != nil {
		return err
	}
	// delegate virtual coins to the validator
	_, err = k.staking.Delegate(
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
	return err
}

func (k Keeper) Undelegate(pCtx sdk.Context, actor sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Coin) error {
	if amt.Amount.IsZero() || amt.Amount.IsNegative() {
		return errors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure staking constraints
	bondDenom := k.staking.BondDenom(pCtx)
	if amt.Denom != bondDenom {
		return errors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", amt.Denom, bondDenom)
	}

	cacheCtx, done := pCtx.CacheContext() // work in a cached store (safety net?)
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

	unbondedAmount := undelegatedCoins.AmountOf(bondDenom)
	k.bank.AddSupplyOffset(cacheCtx, bondDenom, unbondedAmount)
	newDelegatedAmt := k.getTotalDelegatedAmount(cacheCtx, actor).Sub(unbondedAmount)
	k.setTotalDelegatedAmount(cacheCtx, actor, newDelegatedAmt)

	done()
	return nil
}
