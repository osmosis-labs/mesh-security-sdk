package consumer

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

func (app *MeshProviderApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

func (app *MeshProviderApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *MeshProviderApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app *MeshProviderApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *MeshProviderApp) GetStakingKeeper() *stakingkeeper.Keeper {
	return app.StakingKeeper
}

func (app *MeshProviderApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *MeshProviderApp) GetWasmKeeper() wasmkeeper.Keeper {
	return app.WasmKeeper
}
