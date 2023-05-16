package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
func (b BankKeeperAdapter) AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount sdk.Int) {
}

var _ types.XStakingKeeper = &StakingKeeperAdapter{}

// StakingKeeperAdapter is an adapter to enhance the vanilla sdk staking keeper with additional functionality
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
// This is copied from the Osmosis sdk fork
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
