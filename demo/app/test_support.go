package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
)

func (app *MeshApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

func (app *MeshApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *MeshApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app *MeshApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *MeshApp) GetStakingKeeper() *stakingkeeper.Keeper {
	return app.StakingKeeper
}

func (app *MeshApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *MeshApp) GetWasmKeeper() wasmkeeper.Keeper {
	return app.WasmKeeper
}
