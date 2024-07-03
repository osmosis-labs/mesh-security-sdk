package keeper

import (
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

type Keeper struct {
	storeKey  storetypes.StoreKey
	cdc       codec.BinaryCodec
	authority string

	bankKeeper types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey,
	authority string, bankKeeper types.BankKeeper,
	) *Keeper {
	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
		bankKeeper: bankKeeper,
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
