package app

import (
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/CosmWasm/wasmd/x/wasm"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var emptyWasmOpts []wasm.Option

func TestMeshdExport(t *testing.T) {
	db := dbm.NewMemDB()
	logger := log.NewCustomLogger(zerolog.New(os.Stdout).With().Timestamp().Logger())

	gapp, ctx := NewMeshAppWithCustomOptions(t, false, SetupOptions{
		Logger:   logger,
		DB:       db,
		AppOpts:  simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
		WasmOpts: emptyWasmOpts,
	})

	// finalize block so we have CheckTx state set
	_, err := gapp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   sdk.UnwrapSDKContext(ctx).BlockTime(),
	})
	require.NoError(t, err)

	_, err = gapp.Commit()
	require.NoError(t, err)

	// Making a new app object with the db, so that initchain hasn't been called

	newGapp := NewMeshApp(logger, db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()), emptyWasmOpts)
	_, err = newGapp.ExportAppStateAndValidators(false, []string{}, nil)
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// ensure that blocked addresses are properly set in bank keeper
func TestBlockedAddrs(t *testing.T) {
	gapp := Setup(t)

	for acc := range BlockedAddresses() {
		t.Run(acc, func(t *testing.T) {
			var addr sdk.AccAddress
			if modAddr, err := sdk.AccAddressFromBech32(acc); err == nil {
				addr = modAddr
			} else {
				addr = gapp.AccountKeeper.GetModuleAddress(acc)
			}
			require.True(t, gapp.BankKeeper.BlockedAddr(addr), "ensure that blocked addresses are properly set in bank keeper")
		})
	}
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}
