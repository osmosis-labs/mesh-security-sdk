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

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
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

func (k Keeper) HandleBondMsg(ctx sdk.Context, actor sdk.AccAddress, bondMsg *contract.BondMsg) ([]sdk.Event, [][]byte, error) {
	if actor.String() != k.VaultAddress(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(bondMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	delAddr, err := sdk.AccAddressFromBech32(bondMsg.Delegator)
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

func (k Keeper) HandleUnbondMsg(ctx sdk.Context, actor sdk.AccAddress, unbondMsg *contract.UnbondMsg) ([]sdk.Event, [][]byte, error) {
	if actor.String() != k.VaultAddress(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("contract has no permission for mesh security operations")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(unbondMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	delAddr, err := sdk.AccAddressFromBech32(unbondMsg.Delegator)
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

func (k Keeper) HandleUnstakeMsg(ctx sdk.Context, actor sdk.AccAddress, unstakeMsg *contract.UnstakeMsg) ([]sdk.Event, [][]byte, error) {
	nativeContract := k.NativeStakingAddress(ctx)
	nativeContractAddr, err := sdk.AccAddressFromBech32(nativeContract)
	if err != nil {
		return nil, nil, sdkerrors.ErrInvalidAddress.Wrapf("native staking contract not able to get")
	}
	var proxyRes types.ProxyByOwnerResponse

	resBytes, err := k.wasmKeeper.QuerySmart(
		ctx,
		nativeContractAddr,
		[]byte(fmt.Sprintf(`{"proxy_by_owner": {"owner": "%s"}}`, unstakeMsg.Delegator)),
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

	proxyContract, err := sdk.AccAddressFromBech32(proxyRes.Proxy)
	if err != nil {
		return nil, nil, sdkerrors.ErrInvalidAddress.Wrapf("native staking proxy contract not able to get")
	}

	coin, err := wasmkeeper.ConvertWasmCoinToSdkCoin(unstakeMsg.Amount)
	if err != nil {
		return nil, nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(unstakeMsg.Validator)
	if err != nil {
		return nil, nil, err
	}

	err = k.unstake(ctx, proxyContract, valAddr, coin)
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

func (k Keeper) unstake(ctx sdk.Context, proxyContract sdk.AccAddress, validator sdk.ValAddress, coin sdk.Coin) error {
	if coin.Amount.IsNil() || coin.Amount.IsZero() || coin.Amount.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("amount")
	}

	// Ensure staking constraints
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if coin.Denom != bondDenom {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", coin.Denom, bondDenom)
	}

	validatorInfo, found := k.stakingKeeper.GetValidator(ctx, validator)
	if !found {
		return sdkerrors.ErrNotFound.Wrapf("can not found validator with address: %s", validator.String())
	}

	shares, err := k.stakingKeeper.ValidateUnbondAmount(ctx, proxyContract, validator, coin.Amount)
	if err == stakingtypes.ErrNoDelegation {
		return nil
	} else if err != nil {
		return err
	}

	if validatorInfo.IsBonded() {
		_, err = k.stakingKeeper.Undelegate(ctx, proxyContract, validator, shares)
		if err != nil {
			return err
		}

		return nil
	}

	return k.InstantUndelegate(ctx, proxyContract, validator, shares)
}

func (k Keeper) InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, validator sdk.ValAddress, sharesAmount sdk.Dec) error {
	returnAmount, err := k.stakingKeeper.Unbond(ctx, delAddr, validator, sharesAmount)
	if err != nil {
		return err
	}

	bondDenom := k.stakingKeeper.BondDenom(ctx)

	amt := sdk.NewCoin(bondDenom, returnAmount)
	res := sdk.NewCoins(amt)

	return k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delAddr, res)
}
