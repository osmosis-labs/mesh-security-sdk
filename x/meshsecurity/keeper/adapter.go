package keeper

import (
	"context"

	"cosmossdk.io/math"

	evidencetypes "cosmossdk.io/x/evidence/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var _ types.XBankKeeper = &BankKeeperAdapter{}

// BankKeeperAdapter adapter to vanilla SDK bank keeper
type BankKeeperAdapter struct {
	types.SDKBankKeeper
}

// NewBankKeeperAdapter constructor
func NewBankKeeperAdapter(k types.SDKBankKeeper) *BankKeeperAdapter {
	return &BankKeeperAdapter{SDKBankKeeper: k}
}

// AddSupplyOffset noop
func (b BankKeeperAdapter) AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount math.Int) {
}

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
func (s StakingKeeperAdapter) InstantUndelegate(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount math.LegacyDec) (sdk.Coins, error) {
	validator, err := s.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := s.Unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return nil, err
	}

	bondDenom, err := s.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

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

var _ evidencetypes.SlashingKeeper = SlashingKeeperDecorator{}

// SlashingKeeperDecorator to capture tombstone events
type SlashingKeeperDecorator struct {
	evidencetypes.SlashingKeeper
	stakingKeeper types.SDKStakingKeeper
	k             *Keeper
}

// CaptureTombstoneDecorator constructor
func CaptureTombstoneDecorator(k *Keeper, slashingKeeper evidencetypes.SlashingKeeper, stakingKeeper types.SDKStakingKeeper) *SlashingKeeperDecorator {
	return &SlashingKeeperDecorator{SlashingKeeper: slashingKeeper, stakingKeeper: stakingKeeper, k: k}
}

// Tombstone is executed in the end-blocker by the evidence module
func (e SlashingKeeperDecorator) Tombstone(ctx context.Context, address sdk.ConsAddress) error {
	v, err := e.stakingKeeper.GetValidatorByConsAddr(ctx, address)
	if err != nil {
		ModuleLogger(ctx).
			Error("can not propagate tompstone: validator not found", "validator", address.String())
	} else {
		vAddr, err2 := sdk.ValAddressFromBech32(v.GetOperator())
		if err2 != nil {
			ModuleLogger(ctx).
				Error("can not propagate tompstone: validator address is invalid", "validator", address.String())
			return err2
		}

		if err := e.k.ScheduleTombstoned(ctx, vAddr); err != nil {
			ModuleLogger(ctx).
				Error("can not propagate tompstone: scheduler",
					"cause", err,
					"validator", address.String())
		}
	}
	return e.SlashingKeeper.Tombstone(ctx, address)
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
func (s StakingDecorator) Slash(ctx context.Context, consAddr sdk.ConsAddress, power int64, height int64, slashRatio math.LegacyDec) math.Int {
	val, err := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		ModuleLogger(ctx).
			Error("can not propagate slash: validator not found", "validator", consAddr.String())
		return math.ZeroInt()
	}

	vAddr, err2 := sdk.ValAddressFromBech32(val.GetOperator())
	if err2 != nil {
		ModuleLogger(ctx).
			Error("can not propagate slash: validator address is invalid", "validator", consAddr.String())
		return math.ZeroInt()
	}

	totalSlashAmount, err := s.StakingKeeper.Slash(ctx, consAddr, power, height, slashRatio)
	if err := s.k.ScheduleSlashed(ctx, vAddr, power, height, totalSlashAmount, slashRatio); err != nil {
		ModuleLogger(ctx).
			Error("can not propagate slash: schedule event",
				"cause", err,
				"validator", consAddr.String())
	}
	return totalSlashAmount
}

// SlashWithInfractionReason implementation doesn't require the infraction (types.Infraction) to work but is required by Interchain Security.
func (s StakingDecorator) SlashWithInfractionReason(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor math.LegacyDec, _ stakingtypes.Infraction) math.Int {
	return s.Slash(ctx, consAddr, infractionHeight, power, slashFactor)
}

// Jail captures the jail event and calls the decorated staking keeper jail method
func (s StakingDecorator) Jail(ctx context.Context, consAddr sdk.ConsAddress) {

	val, err := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		ModuleLogger(ctx).
			Error("can not propagate jail: validator not found", "validator", consAddr.String())
	} else {
		vAddr, err2 := sdk.ValAddressFromBech32(val.GetOperator())
		if err2 != nil {
			ModuleLogger(ctx).
				Error("can not propagate jail: validator address is invalid", "validator", consAddr.String())
			return
		}

		if err := s.k.ScheduleJailed(ctx, vAddr); err != nil {
			ModuleLogger(ctx).
				Error("can not propagate jail: schedule event",
					"cause", err,
					"validator", consAddr.String())
		}
	}
	s.StakingKeeper.Jail(ctx, consAddr)
}

// Unjail captures the unjail event and calls the decorated staking keeper unjail method
func (s StakingDecorator) Unjail(ctx context.Context, consAddr sdk.ConsAddress) {
	val, err := s.StakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		ModuleLogger(ctx).
			Error("can not propagate unjail: validator not found", "validator", consAddr.String())
	} else {
		vAddr, err2 := sdk.ValAddressFromBech32(val.GetOperator())
		if err2 != nil {
			ModuleLogger(ctx).
				Error("can not propagate unjail: validator address is invalid", "validator", consAddr.String())
			return
		}

		if err := s.k.ScheduleUnjailed(ctx, vAddr); err != nil {
			ModuleLogger(ctx).
				Error("can not propagate unjail: schedule event",
					"cause", err,
					"validator", consAddr.String())
		}
	}
	s.StakingKeeper.Unjail(ctx, consAddr)
}
