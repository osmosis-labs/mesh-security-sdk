package keeper

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

type Keeper struct {
	storeKey  storetypes.StoreKey
	cdc       codec.BinaryCodec
	authority string

	bankKeeper    types.BankKeeper
	wasmKeeper    types.WasmKeeper
	stakingKeeper types.StakingKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey,
	authority string, bankKeeper types.BankKeeper, wasmKeeper types.WasmKeeper,
	stakingKeeper types.StakingKeeper,
) *Keeper {
	return &Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		authority:  authority,
		bankKeeper: bankKeeper,
		wasmKeeper: wasmKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the x/staking module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams gets the module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// InitGenesis initializes the meshsecurity provider module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the meshsecurity provider module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

func (k Keeper) Bond(ctx sdk.Context, actor sdk.AccAddress, delegator sdk.AccAddress, coin sdk.Coin) error {
	return k.bankKeeper.DelegateCoins(ctx, delegator, actor, sdk.NewCoins(coin))
}

func (k Keeper) Unbond(ctx sdk.Context, actor sdk.AccAddress, delegator sdk.AccAddress, coin sdk.Coin) error {
	return k.bankKeeper.UndelegateCoins(ctx, actor, delegator, sdk.NewCoins(coin))
}

func (k Keeper) Unstake(ctx sdk.Context, actor sdk.AccAddress, validator sdk.ValAddress, coin sdk.Coin) error {
	if coin.Amount.IsNil() || coin.Amount.IsZero() || coin.Amount.IsNegative() {
		return errors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure staking constraints
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if coin.Denom != bondDenom {
		return errors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", coin.Denom, bondDenom)
	}

	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, actor, validator, coin.Amount)
	if err == stakingtypes.ErrNoDelegation {
		return nil
	} else if err != nil {
		return err
	}

	err = k.InstantUndelegate(ctx, actor, validator, shares)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) error {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return stakingtypes.ErrNoDelegatorForAddress
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, delAddr, valAddr, sharesAmount)
	if err != nil {
		return err
	}

	bondDenom := k.stakingKeeper.BondDenom(ctx)

	amt := sdk.NewCoin(bondDenom, returnAmount)
	res := sdk.NewCoins(amt)

	moduleName := stakingtypes.NotBondedPoolName
	if validator.IsBonded() {
		moduleName = stakingtypes.BondedPoolName
	}
	err = k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, moduleName, delAddr, res)
	if err != nil {
		return err
	}
	return nil
}
