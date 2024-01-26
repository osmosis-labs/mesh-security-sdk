package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/runtime"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	codec2 "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type encodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

var moduleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	gov.NewAppModuleBasic([]govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		// TODO:
		// upgradeclient.LegacyProposalHandler,
		// upgradeclient.LegacyCancelProposalHandler,
	}),
)

func makeEncodingConfig(_ testing.TB) encodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	moduleBasics.RegisterLegacyAminoCodec(amino)
	moduleBasics.RegisterInterfaces(interfaceRegistry)
	// add mesh-security types
	types.RegisterInterfaces(interfaceRegistry)
	types.RegisterLegacyAminoCodec(amino)

	return encodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             amino,
	}
}

type TestKeepers struct {
	StakingKeeper  *stakingkeeper.Keeper
	SlashingKeeper slashingkeeper.Keeper
	BankKeeper     bankkeeper.Keeper
	StoreKey       *storetypes.KVStoreKey
	EncodingConfig encodingConfig
	MeshKeeper     *Keeper
	AccountKeeper  authkeeper.AccountKeeper
	WasmKeeper     *wasmkeeper.Keeper
	Faucet         *wasmkeeper.TestFaucet
}

func CreateDefaultTestInput(t testing.TB, opts ...Option) (sdk.Context, TestKeepers) {
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	ms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distributiontypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibcexported.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey, feegrant.StoreKey, authzkeeper.StoreKey,
		wasmtypes.StoreKey, types.StoreKey,
	)
	for _, v := range keys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
	}
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, types.MemStoreKey)
	for _, v := range memKeys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeMemory, db)
	}
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	for _, v := range tkeys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeTransient, db)
	}
	require.NoError(t, ms.LoadLatestVersion())

	encConfig := makeEncodingConfig(t)
	appCodec := encConfig.Marshaler

	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:     nil,
		distributiontypes.ModuleName:   nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		types.ModuleName:               {authtypes.Minter, authtypes.Burner},
	}
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]), // target store
		authtypes.ProtoBaseAccount,                          // prototype
		maccPerms,
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority,
	)
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddrs,
		authority,
		logger,
	)
	require.NoError(t, bankKeeper.SetParams(ctx, banktypes.DefaultParams()))

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority,
		authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)
	require.NoError(t, stakingKeeper.SetParams(ctx, stakingtypes.DefaultParams()))

	slashingKeeper := slashingkeeper.NewKeeper(
		appCodec,
		encConfig.Amino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	require.NoError(t, slashingKeeper.SetParams(ctx, slashingtypes.DefaultParams()))

	distKeeper := distributionkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distributiontypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(distributiontypes.ModuleName).String(),
	)
	require.NoError(t, distKeeper.Params.Set(ctx, distributiontypes.DefaultParams()))

	querier := baseapp.NewGRPCQueryRouter()
	querier.SetInterfaceRegistry(encConfig.InterfaceRegistry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encConfig.InterfaceRegistry)

	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedWasmKeeper := capabilityKeeper.ScopeToModule(types.ModuleName)

	paramsKeeper := paramskeeper.NewKeeper(
		appCodec,
		encConfig.Amino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		t.TempDir(),
		nil,
		authtypes.NewModuleAddress(upgradetypes.ModuleName).String(),
	)

	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		paramsKeeper.Subspace(ibcexported.ModuleName),
		stakingKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	cfg := sdk.GetConfig()
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())

	wasmKeeper := wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distributionkeeper.NewQuerier(distKeeper),
		ibcKeeper.ChannelKeeper, // ICS4Wrapper
		ibcKeeper.ChannelKeeper,
		ibcKeeper.PortKeeper,
		scopedWasmKeeper,
		wasmtesting.MockIBCTransferKeeper{},
		msgRouter,
		querier,
		t.TempDir(),
		wasmtypes.DefaultWasmConfig(),
		"iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3,virtual_staking",
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	require.NoError(t, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))

	msKeeper := NewKeeper(
		appCodec,
		keys[types.StoreKey],
		memKeys[types.MemStoreKey],
		bankKeeper,
		stakingKeeper,
		wasmKeeper,
		authority,
		opts...,
	)
	require.NoError(t, msKeeper.SetParams(ctx, types.DefaultParams(sdk.DefaultBondDenom)))

	faucet := wasmkeeper.NewTestFaucet(t, ctx, bankKeeper, minttypes.ModuleName, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000_000))
	return ctx, TestKeepers{
		AccountKeeper:  accountKeeper,
		StakingKeeper:  stakingKeeper,
		SlashingKeeper: slashingKeeper,
		BankKeeper:     bankKeeper,
		StoreKey:       keys[types.StoreKey],
		EncodingConfig: encConfig,
		MeshKeeper:     msKeeper,
		WasmKeeper:     &wasmKeeper,
		Faucet:         faucet,
	}
}

// FetchAllStoredOperations load all ops from temp db
func FetchAllStoredOperations(t *testing.T, ctx sdk.Context, msKeeper *Keeper) map[string][]types.PipedValsetOperation {
	index := make(map[string][]types.PipedValsetOperation, 1)
	err := msKeeper.iteratePipedValsetOperations(ctx, func(valAddr sdk.ValAddress, op types.PipedValsetOperation, slashInfo *types.SlashInfo) bool {
		ops, ok := index[valAddr.String()]
		if !ok {
			ops = []types.PipedValsetOperation{}
		}
		index[valAddr.String()] = append(ops, op)
		return false
	})
	require.NoError(t, err)
	return index
}

// for test code only
func must[t any](s t, err error) t {
	if err != nil {
		panic(err)
	}
	return s
}

// MinValidatorFixture creates minimal sdk validator object
func MinValidatorFixture(t *testing.T) stakingtypes.Validator {
	t.Helper()
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	pk, err := codec2.FromTmPubKeyInterface(pubKey)
	require.NoError(t, err)
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	return stakingtypes.Validator{
		ConsensusPubkey: pkAny,
		OperatorAddress: sdk.ValAddress(rand.Bytes(address.Len)).String(),
	}
}
