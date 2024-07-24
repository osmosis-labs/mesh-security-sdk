package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/wasmbinding/bindings"
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
		storeKey:      storeKey,
		cdc:           cdc,
		authority:     authority,
		bankKeeper:    bankKeeper,
		wasmKeeper:    wasmKeeper,
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

func (k Keeper) HandleDepositMsg(ctx sdk.Context, actor sdk.AccAddress, depositMsg *bindings.DepositMsg) ([]sdk.Event, [][]byte, error) {
	if actor.String() != k.VaultAddress(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(depositMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	delAddr, err := sdk.AccAddressFromBech32(depositMsg.Delegator)
	if err != nil {
		return nil, nil, err
	}

	err = k.bankKeeper.DelegateCoins(ctx, delAddr, actor, sdk.NewCoins(coin))
	if err != nil {
		return nil, nil, err
	}

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeBond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
	)}, nil, nil
}

func (k Keeper) HandleWithdrawMsg(ctx sdk.Context, actor sdk.AccAddress, withdrawMsg *bindings.WithdrawMsg) ([]sdk.Event, [][]byte, error) {
	if actor.String() != k.VaultAddress(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(withdrawMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	delAddr, err := sdk.AccAddressFromBech32(withdrawMsg.Delegator)
	if err != nil {
		return nil, nil, err
	}

	err = k.bankKeeper.UndelegateCoins(ctx, actor, delAddr, sdk.NewCoins(coin))
	if err != nil {
		return nil, nil, err
	}

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeUnbond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
	)}, nil, nil
}

func (k Keeper) HandleUnstakeMsg(ctx sdk.Context, actor sdk.AccAddress, unstakeMsg *bindings.UnstakeMsg) ([]sdk.Event, [][]byte, error) {
	nativeContractAddr := k.NativeStakingAddress(ctx)
	var proxyRes types.ProxyByOwnerResponse

	resBytes, err := k.wasmKeeper.QuerySmart(ctx,
		sdk.AccAddress(nativeContractAddr),
		[]byte(fmt.Sprintf(`{"proxy_by_owner": {"owner": "%s"}}`, actor.String())),
	)
	if err != nil {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}
	if err = json.Unmarshal(resBytes, &proxyRes); err != nil {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}
	if proxyRes.Proxy == "" {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(unstakeMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(unstakeMsg.Validator)
	if err != nil {
		return nil, nil, err
	}

	err = k.unstake(ctx, actor, valAddr, coin)
	if err != nil {
		return nil, nil, err
	}

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeUnstake,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
		sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
	)}, nil, nil
}

func (k Keeper) unstake(ctx sdk.Context, actor sdk.AccAddress, validator sdk.ValAddress, coin sdk.Coin) error {
	if coin.Amount.IsNil() || coin.Amount.IsZero() || coin.Amount.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure staking constraints
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if coin.Denom != bondDenom {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", coin.Denom, bondDenom)
	}

	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, actor, validator, coin.Amount)
	if err == stakingtypes.ErrNoDelegation {
		return nil
	} else if err != nil {
		return err
	}

	validatorInfo, found := k.stakingKeeper.GetValidator(ctx, validator)
	if !found {
		return sdkerrors.ErrNotFound.Wrapf("can not found validator with address: %s", validator.String())
	}
	if validatorInfo.IsBonded() {
		_, err = k.stakingKeeper.Undelegate(ctx, actor, validator, shares)
	} else {
		_, err = k.InstantUndelegate(ctx, actor, validatorInfo, shares)
	}

	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, validator stakingtypes.Validator, sharesAmount sdk.Dec) (sdk.Coin, error) {
	returnAmount, err := k.stakingKeeper.Unbond(ctx, delAddr, sdk.ValAddress(validator.OperatorAddress), sharesAmount)
	if err != nil {
		return sdk.Coin{}, err
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
		return sdk.Coin{}, err
	}
	return amt, nil
}
