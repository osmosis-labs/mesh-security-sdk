package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

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
	clientKeeper  types.ClientKeeper
	channelKeeper types.ChannelKeeper
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

func (keeper Keeper) HandleRegistryConsumer(
	ctx sdk.Context,
	channelID string,
	contractAddress sdk.AccAddress,
) ([]sdk.Event, [][]byte, error) {
	// get portID
	portID := "wasm." + contractAddress.String()
	// get connection ID
	channel, found := keeper.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return nil, nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	connection, err := keeper.channelKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if err != nil {
		return nil, nil, err
	}

	clientState, found := keeper.clientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return nil, nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	tmCs, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return nil, nil, errorsmod.Wrapf(types.ErrUnsupportedCounterpartyClientTypes, "got (%s)", clientState.ClientType())
	}

	keeper.SetConsumerChainID(ctx, tmCs.ChainId, contractAddress, channelID)

	return []sdk.Event{sdk.NewEvent(
		types.EventTypeUnbond,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyContractAddress, string(contractAddress)),
		sdk.NewAttribute(types.AttributeConsumerChainID, tmCs.ChainId),
	)}, nil, nil
}

func (keeper Keeper) SetConsumerChainID(ctx sdk.Context, chainID string, contractAddress []byte, clientID string) {
	store := ctx.KVStore(keeper.storeKey)

	bz := append([]byte(chainID), []byte(clientID)...)
	key := append(types.ConsumerChainIDKey, bz...)

	store.Set(key, contractAddress)
}

func (keeper Keeper) IteratorExternalStakingContractAddr(ctx sdk.Context, chainID string, cb func(contractAddress sdk.AccAddress) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(types.ConsumerChainIDKey, []byte(chainID)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if cb(iterator.Value()) {
			return
		}
	}
}

func (keeper Keeper) GetExternalStakingContractAccAddr(ctx sdk.Context, chainID string, clientID string) sdk.AccAddress {
	store := ctx.KVStore(keeper.storeKey)

	bz := append([]byte(chainID), []byte(clientID)...)
	return store.Get(append(types.ConsumerChainIDKey, bz...))
}
