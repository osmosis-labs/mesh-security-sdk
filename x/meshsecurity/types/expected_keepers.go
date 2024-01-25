package types

import (
	context "context"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SDKBankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type XBankKeeper interface {
	SDKBankKeeper
	AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount math.Int)
}

// SDKStakingKeeper expected staking keeper.
type SDKStakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
	GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	ValidateUnbondAmount(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt math.Int) (shares math.LegacyDec, err error)
	Delegate(ctx context.Context, delAddr sdk.AccAddress, bondAmt math.Int, tokenSrc stakingtypes.BondStatus, validator stakingtypes.Validator, subtractAccount bool) (newShares math.LegacyDec, err error)
	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, err error)
	UnbondingTime(ctx context.Context) (time.Duration, error)
	GetParams(ctx context.Context) (stakingtypes.Params, error)
	Unbond(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, shares math.LegacyDec) (amount math.Int, err error)
	IterateBondedValidatorsByPower(ctx context.Context, fn func(int64, stakingtypes.ValidatorI) bool) error
	TotalBondedTokens(ctx context.Context) (math.Int, error)
	IterateDelegations(ctx context.Context, delegator sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool) error
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)
}

type XStakingKeeper interface {
	SDKStakingKeeper
	InstantUndelegate(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount math.LegacyDec) (sdk.Coins, error)
}

// CommunityPoolKeeper expected distribution keeper.
type CommunityPoolKeeper interface {
	WithdrawDelegationRewards(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// AccountKeeper interface contains functions for getting accounts and the module address
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) authtypes.ModuleAccountI
}

// WasmKeeper abstract wasm keeper
type WasmKeeper interface {
	Sudo(ctx context.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
}
