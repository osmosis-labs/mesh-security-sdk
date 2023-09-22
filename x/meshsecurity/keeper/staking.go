package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
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

func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.ScheduleBonded(ctx, valAddr)
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	// removed from the active set
	return h.k.ScheduleUnbonded(ctx, valAddr)
}

func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	// before the BeforeValidatorModified is called
	// slashed
	// check ValidatorSigningInfo if jailed or tombstoned
	return h.k.ScheduleSlashed(ctx, valAddr, fraction)
}

func (h Hooks) AfterUnbondingInitiated(ctx sdk.Context, id uint64) error {
	// ignore as we hook into AfterValidatorBeginUnbonding already
	return nil
}

func (h Hooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return h.k.ScheduleModified(ctx, valAddr)
}

// AfterValidatorCreated noop
func (h Hooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}

// AfterValidatorRemoved noop
func (h Hooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationCreated noop
func (h Hooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified noop
func (h Hooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved noop
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// AfterDelegationModified noop
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

// StakingDecorator decorate vanilla staking keeper to capture the unjail event
type StakingDecorator struct {
	slashingtypes.StakingKeeper
	k *Keeper
}

// NewStakingDecorator constructor
func NewStakingDecorator(stakingKeeper slashingtypes.StakingKeeper, k *Keeper) *StakingDecorator {
	return &StakingDecorator{StakingKeeper: stakingKeeper, k: k}
}

func (s StakingDecorator) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	val := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if val == nil {
		ModuleLogger(ctx).
			Error("can not propagate unjail: validator not found",
				"validator", consAddr.String())
		s.StakingKeeper.Unjail(ctx, consAddr)
		return
	}
	if err := s.k.ScheduleUnjailed(ctx, val.GetOperator()); err != nil {
		ModuleLogger(ctx).
			Error("can not propagate unjail: schedule event",
				"cause", err,
				"validator", consAddr.String())
	}
	s.StakingKeeper.Unjail(ctx, consAddr)
}
