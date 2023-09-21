package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingHooks struct{}

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

// Hooks return the mesh-security hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// events
// new validator
// add to active set
// remove from active set
// delete validator
// update (commission)
// slash

// validator commission updated: BeforeValidatorModified

func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	// new in list
	// query validator
	return nil
}

func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	// removed as validator
	return nil
}

func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	// added to the active set
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	// removed from the active set
	return h.k.ScheduleUnboding(ctx, valAddr)
}

func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	// before the BeforeValidatorModified is called
	// slashed
	// check ValidatorSigningInfo if jailed or tombstoned
	return nil
}

func (h Hooks) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	// ignore as we hook into AfterValidatorBeginUnbonding already
	return nil
}

// BeforeDelegationCreated Deprecated unused
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified Deprecated unused
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	// called before commission is modified
	// or slash
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
