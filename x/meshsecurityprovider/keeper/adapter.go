package keeper

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	meshsecuritykeeper "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var _ types.XStakingKeeper = &StakingKeeperAdapter{}

// StakingKeeperAdapter is an adapter to enhance the vanilla sdk staking keeper with additional functionality
// required for MS. The methods match Osmosis SDK fork.
type StakingKeeperAdapter struct {
	types.SDKStakingKeeper
	bank types.SDKBankKeeper
}

// NewStakingKeeperAdapter constructor
func NewStakingKeeperAdapter(k types.SDKStakingKeeper, b types.SDKBankKeeper) *StakingKeeperAdapter {
	return &StakingKeeperAdapter{SDKStakingKeeper: k, bank: b}
}

// InstantUndelegate allows another module account to undelegate while bypassing unbonding time.
// This function is a combination of Undelegate and CompleteUnbonding,
// but skips the creation and deletion of UnbondingDelegationEntry
//
// The code is copied from the Osmosis SDK fork https://github.com/osmosis-labs/cosmos-sdk/blob/v0.45.0x-osmo-v9.3/x/staking/keeper/delegation.go#L757
func (s StakingKeeperAdapter) InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (sdk.Coins, error) {
	validator, found := s.GetValidator(ctx, valAddr)
	if !found {
		return nil, stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := s.Unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return nil, err
	}

	bondDenom := s.BondDenom(ctx)

	amt := sdk.NewCoin(bondDenom, returnAmount)
	res := sdk.NewCoins(amt)

	moduleName := stakingtypes.NotBondedPoolName
	if validator.IsBonded() {
		moduleName = stakingtypes.BondedPoolName
	}
	err = s.bank.UndelegateCoinsFromModuleToAccount(ctx, moduleName, delAddr, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// StakingDecorator decorate vanilla staking keeper to capture the jail and unjail events
type StakingDecorator struct {
	slashingtypes.StakingKeeper
	k *Keeper
}

// NewStakingDecorator constructor
func NewStakingDecorator(stakingKeeper slashingtypes.StakingKeeper, k *Keeper) *StakingDecorator {
	return &StakingDecorator{StakingKeeper: stakingKeeper, k: k}
}

// Slash captures the slash event and calls the decorated staking keeper slash method
func (s StakingDecorator) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, power int64, height int64, slashRatio sdk.Dec) math.Int {
	if s.k.meshConsumer == nil {
		return s.StakingKeeper.Slash(ctx, consAddr, power, height, slashRatio)
	}

	val := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	totalSlashAmount := s.StakingKeeper.Slash(ctx, consAddr, power, height, slashRatio)
	if val == nil {
		meshsecuritykeeper.ModuleLogger(ctx).
			Error("can not propagate slash: validator not found", "validator", consAddr.String())
	} else if err := s.k.meshConsumer.ScheduleSlashed(ctx, val.GetOperator(), power, height, totalSlashAmount, slashRatio); err != nil {
		meshsecuritykeeper.ModuleLogger(ctx).
			Error("can not propagate slash: schedule event",
				"cause", err,
				"validator", consAddr.String())
	}
	return totalSlashAmount
}

// SlashWithInfractionReason implementation doesn't require the infraction (types.Infraction) to work but is required by Interchain Security.
func (s StakingDecorator) SlashWithInfractionReason(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec, infraction stakingtypes.Infraction) math.Int {
	// foward it to native-staking contract
	params := s.k.GetParams(ctx)
	nativeStakingAddr := sdk.MustAccAddressFromBech32(params.NativeStakingAddress)

	if infraction == stakingtypes.Infraction_INFRACTION_DOUBLE_SIGN {
		s.k.SendJailHandlingMsg(ctx, nativeStakingAddr, nil, []string{consAddr.String()})
	}

	if infraction == stakingtypes.Infraction_INFRACTION_DOWNTIME {
		s.k.SendJailHandlingMsg(ctx, nativeStakingAddr, []string{consAddr.String()}, nil)
	}
	return s.Slash(ctx, consAddr, infractionHeight, power, slashFactor)
}

// Jail captures the jail event and calls the decorated staking keeper jail method
func (s StakingDecorator) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	if s.k.meshConsumer == nil {
		s.StakingKeeper.Jail(ctx, consAddr)
	}

	val := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if val == nil {
		meshsecuritykeeper.ModuleLogger(ctx).
			Error("can not propagate jail: validator not found", "validator", consAddr.String())
	} else if err := s.k.meshConsumer.ScheduleJailed(ctx, val.GetOperator()); err != nil {
		meshsecuritykeeper.ModuleLogger(ctx).
			Error("can not propagate jail: schedule event",
				"cause", err,
				"validator", consAddr.String())
	}
	s.StakingKeeper.Jail(ctx, consAddr)
}

// Unjail captures the unjail event and calls the decorated staking keeper unjail method
func (s StakingDecorator) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	if s.k.meshConsumer == nil {
		s.StakingKeeper.Unjail(ctx, consAddr)
	}

	val := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if val == nil {
		meshsecuritykeeper.ModuleLogger(ctx).
			Error("can not propagate unjail: validator not found", "validator", consAddr.String())
	} else if err := s.k.meshConsumer.ScheduleUnjailed(ctx, val.GetOperator()); err != nil {
		meshsecuritykeeper.ModuleLogger(ctx).
			Error("can not propagate unjail: schedule event",
				"cause", err,
				"validator", consAddr.String())
	}
	s.StakingKeeper.Unjail(ctx, consAddr)
}
