package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	"github.com/stretchr/testify/require"

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
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
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
	BankKeeper     bankkeeper.Keeper
	StoreKey       *storetypes.KVStoreKey
	EncodingConfig encodingConfig
	MeshKeeper     *Keeper
}

func CreateDefaultTestInput(t testing.TB) (sdk.Context, TestKeepers) {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, types.StoreKey,
	)
	for _, v := range keys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
	}
	require.NoError(t, ms.LoadLatestVersion())

	encConfig := makeEncodingConfig(t)
	appCodec := encConfig.Marshaler

	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:     nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		types.ModuleName:               {authtypes.Minter, authtypes.Burner},
	}
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],   // target store
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
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
		keys[banktypes.StoreKey],
		accountKeeper,
		blockedAddrs,
		authority,
	)
	require.NoError(t, bankKeeper.SetParams(ctx, banktypes.DefaultParams()))

	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		accountKeeper,
		bankKeeper,
		authority,
	)
	require.NoError(t, stakingKeeper.SetParams(ctx, stakingtypes.DefaultParams()))

	msKeeper := NewKeeper(
		appCodec,
		keys[types.StoreKey],
		bankKeeper,
		stakingKeeper,
		authority,
	)
	return ctx, TestKeepers{
		StakingKeeper:  stakingKeeper,
		BankKeeper:     bankKeeper,
		StoreKey:       keys[types.StoreKey],
		EncodingConfig: encConfig,
		MeshKeeper:     msKeeper,
	}
}
