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

func (app *MeshConsumerApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

func (app *MeshConsumerApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *MeshConsumerApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app *MeshConsumerApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *MeshConsumerApp) GetStakingKeeper() *stakingkeeper.Keeper {
	return app.StakingKeeper
}

func (app *MeshConsumerApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *MeshConsumerApp) GetWasmKeeper() wasmkeeper.Keeper {
	return app.WasmKeeper
}
