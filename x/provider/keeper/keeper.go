package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	ctypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
	"github.com/osmosis-labs/mesh-security-sdk/x/provider/types"
	cptypes "github.com/osmosis-labs/mesh-security-sdk/x/types"
)

// Option is an extension point to instantiate keeper with non default values
type Option interface {
	apply(*Keeper)
}

type Keeper struct {
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey
	cdc      codec.Codec
	bank     types.BankKeeper
	Staking  types.StakingKeeper
	wasm     types.WasmKeeper

	scopedKeeper     types.ScopedKeeper
	channelKeeper    types.ChannelKeeper
	connectionKeeper types.ConnectionKeeper
	clientKeeper     types.ClientKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper constructor with vanilla sdk keepers
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	memoryStoreKey storetypes.StoreKey,
	bank types.BankKeeper,
	staking types.StakingKeeper,
	wasm types.WasmKeeper,
	scopedKeeper types.ScopedKeeper,
	channelKeeper types.ChannelKeeper,
	authority string,
	opts ...Option,
) *Keeper {
	k := &Keeper{
		storeKey:      storeKey,
		memKey:        memoryStoreKey,
		cdc:           cdc,
		bank:          bank,
		Staking:       staking,
		wasm:          wasm,
		scopedKeeper:  scopedKeeper,
		channelKeeper: channelKeeper,
		authority:     authority,
	}
	for _, o := range opts {
		o.apply(k)
	}

	return k
}

// ModuleLogger returns logger with module attribute
func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetConsumerChain is called by OnChanOpenConfirm.
func (k Keeper) SetConsumerChain(ctx sdk.Context, channelID string) error {
	channel, ok := k.channelKeeper.GetChannel(ctx, cptypes.ProviderPortID, channelID)
	if !ok {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "channel not found for channel ID: %s", channelID)
	}
	if len(channel.ConnectionHops) != 1 {
		return errorsmod.Wrap(channeltypes.ErrTooManyConnectionHops, "must have direct connection to consumer chain")
	}
	connectionID := channel.ConnectionHops[0]
	clientID, tmClient, err := k.getUnderlyingClient(ctx, connectionID)
	if err != nil {
		return err
	}
	// Verify that there isn't already a CCV channel for the consumer chain
	chainID := tmClient.ChainId
	if prevChannelID, ok := k.GetChainToChannel(ctx, chainID); ok {
		return errorsmod.Wrapf(cptypes.ErrDuplicateChannel, "CCV channel with ID: %s already created for consumer chain %s", prevChannelID, chainID)
	}

	// the CCV channel is established:
	// - set channel mappings
	k.SetChainToChannel(ctx, chainID, channelID)
	k.SetChannelToChain(ctx, channelID, chainID)
	// - set current block height for the consumer chain initialization
	k.SetInitChainHeight(ctx, chainID, uint64(ctx.BlockHeight()))
	// - remove init timeout timestamp
	k.DeleteInitTimeoutTimestamp(ctx, chainID)

	// emit event on successful addition
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			cptypes.EventTypeChannelEstablished,
			sdk.NewAttribute(sdk.AttributeKeyModule, ctypes.ModuleName),
			sdk.NewAttribute(cptypes.AttributeChainID, chainID),
			sdk.NewAttribute(conntypes.AttributeKeyClientID, clientID),
			sdk.NewAttribute(channeltypes.AttributeKeyChannelID, channelID),
			sdk.NewAttribute(conntypes.AttributeKeyConnectionID, connectionID),
		),
	)
	return nil
}

// Retrieves the underlying client state corresponding to a connection ID.
func (k Keeper) getUnderlyingClient(ctx sdk.Context, connectionID string) (
	clientID string, tmClient *ibctmtypes.ClientState, err error,
) {
	conn, ok := k.connectionKeeper.GetConnection(ctx, connectionID)
	if !ok {
		return "", nil, errorsmod.Wrapf(conntypes.ErrConnectionNotFound,
			"connection not found for connection ID: %s", connectionID)
	}
	clientID = conn.ClientId
	clientState, ok := k.clientKeeper.GetClientState(ctx, clientID)
	if !ok {
		return "", nil, errorsmod.Wrapf(clienttypes.ErrClientNotFound,
			"client not found for client ID: %s", conn.ClientId)
	}
	tmClient, ok = clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", nil, errorsmod.Wrapf(clienttypes.ErrInvalidClientType,
			"invalid client type. expected %s, got %s", ibchost.Tendermint, clientState.ClientType())
	}
	return clientID, tmClient, nil
}

func (k Keeper) GetChainToChannel(ctx sdk.Context, chainID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ChainToChannelKey(chainID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetChainToChannel(ctx sdk.Context, chainID, channelID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ChainToChannelKey(chainID), []byte(channelID))
}

func (k Keeper) SetChannelToChain(ctx sdk.Context, channelID, chainID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ChannelToChainKey(channelID), []byte(chainID))
}

func (k Keeper) SetInitChainHeight(ctx sdk.Context, chainID string, height uint64) {
	store := ctx.KVStore(k.storeKey)
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, height)

	store.Set(types.InitChainHeightKey(chainID), heightBytes)
}

func (k Keeper) GetInitChainHeight(ctx sdk.Context, chainID string) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.InitChainHeightKey(chainID))
	if bz == nil {
		return 0, false
	}

	return binary.BigEndian.Uint64(bz), true
}

func (k Keeper) DeleteInitTimeoutTimestamp(ctx sdk.Context, chainID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.InitTimeoutTimestampKey(chainID))
}

func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) VerifyConsumerChain(ctx sdk.Context, channelID string, connectionHops []string) error {
	if len(connectionHops) != 1 {
		return errorsmod.Wrap(channeltypes.ErrTooManyConnectionHops, "must have direct connection to provider chain")
	}
	connectionID := connectionHops[0]
	clientID, tmClient, err := k.getUnderlyingClient(ctx, connectionID)
	if err != nil {
		return err
	}
	ccvClientId, found := k.GetConsumerClientId(ctx, tmClient.ChainId)
	if !found {
		return errorsmod.Wrapf(cptypes.ErrClientNotFound, "cannot find client for consumer chain %s", tmClient.ChainId)
	}
	if ccvClientId != clientID {
		return errorsmod.Wrapf(types.ErrInvalidConsumerClient, "CCV channel must be built on top of CCV client. expected %s, got %s", ccvClientId, clientID)
	}

	// Verify that there isn't already a CCV channel for the consumer chain
	if prevChannel, ok := k.GetChainToChannel(ctx, tmClient.ChainId); ok {
		return errorsmod.Wrapf(cptypes.ErrDuplicateChannel, "CCV channel with ID: %s already created for consumer chain %s", prevChannel, tmClient.ChainId)
	}
	return nil
}

func (k Keeper) GetConsumerClientId(ctx sdk.Context, chainID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	clientIdBytes := store.Get(types.ChainToClientKey(chainID))
	if clientIdBytes == nil {
		return "", false
	}
	return string(clientIdBytes), true
}

func (k Keeper) GetChannelToChain(ctx sdk.Context, channelID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ChannelToChainKey(channelID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetConsumerCommissionRate(
	ctx sdk.Context,
	chainID string,
	providerAddr types.ProviderConsAddress,
	commissionRate sdk.Dec,
) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := commissionRate.Marshal()
	if err != nil {
		err = fmt.Errorf("consumer commission rate marshalling failed: %s", err)
		ModuleLogger(ctx).Error(err.Error())
		return err
	}

	store.Set(types.ConsumerCommissionRateKey(chainID, providerAddr), bz)
	return nil
}

func (k Keeper) SetDepositors(ctx sdk.Context, del types.Depositors) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := types.ModuleCdc.Marshal(&del)
	if err != nil {
		err = fmt.Errorf("external staker marshalling failed: %s", err)
		ModuleLogger(ctx).Error(err.Error())
		return err
	}
	store.Set(types.DepositorsKey(del.Address), bz)

	return nil
}

func (k Keeper) GetDepositors(ctx sdk.Context, del string) (types.Depositors, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DepositorsKey(del))
	if bz == nil {
		return types.Depositors{}, false
	}
	var r types.Depositors
	if err := r.Unmarshal(bz); err != nil {
		panic(err)
	}
	return r, true
}

func (k Keeper) iterateDepositors(ctx sdk.Context, cb func(depositors types.Depositors) bool) error {
	pStore := prefix.NewStore(ctx.KVStore(k.memKey), types.DepositorsKeyPrefix)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		var r types.Depositors
		if err := r.Unmarshal(iter.Value()); err != nil {
			panic(err)
		}
		if cb(r) {
			break
		}
	}
	return iter.Close()
}

func (k Keeper) SetIntermediary(ctx sdk.Context, inter types.Intermediary) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := types.ModuleCdc.Marshal(&inter)
	if err != nil {
		err = fmt.Errorf("external staker marshalling failed: %s", err)
		ModuleLogger(ctx).Error(err.Error())
		return err
	}
	store.Set(types.IntermediaryKey(inter.Token.Denom), bz)

	return nil
}

func (k Keeper) GetIntermediary(ctx sdk.Context, denom string) (types.Intermediary, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.IntermediaryKey(denom))
	if bz == nil {
		return types.Intermediary{}, false
	}
	var r types.Intermediary
	if err := r.Unmarshal(bz); err != nil {
		panic(err)
	}
	return r, true
}

func (k Keeper) GetContractWithNativeDenom(ctx sdk.Context, denom string) sdk.AccAddress {
	var contractAddr sdk.AccAddress

	store := ctx.KVStore(k.storeKey)
	contractAddr = store.Get(types.ContractWithNativeDenomKey(denom))
	return contractAddr
}

func (k Keeper) SetContractWithNativeDenom(ctx sdk.Context, denom string, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ContractWithNativeDenomKey(denom), contractAddr)
}
