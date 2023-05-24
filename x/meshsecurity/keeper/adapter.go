package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	panic("not implemented, yet")
}
