# Mesh Security SDK

Cosmos module for mesh-security consumer chains. Please see the official project [repo](https://github.com/osmosis-labs/mesh-security)
for specs and wasm contracts.

## Project Structure

* x - module code that is supposed to be imported by consumer chains
* demo - contains an example application and CLI that is using the mesh-security module
* tests/e2e - end-to-end tests with the demo app and contracts

# Integrate the mesh security consumer and provider modules

## Prerequisites 
Projects that want to integrate the meshsecurity-provider module onto their Cosmos SDK chain must enable the following modules:
- [x/staking](https://github.com/cosmos/cosmos-sdk/tree/main/x/staking)
- [x/auth](https://github.com/cosmos/cosmos-sdk/tree/main/x/auth)
- [x/bank](https://github.com/cosmos/cosmos-sdk/tree/main/x/bank)
- [x/wasm](github.com/CosmWasm/wasmd/x/wasm)

## Configuring and Adding Module
1. Install the mesh security package on the go.mod.
    ```
    require (
    ...
    github.com/osmosis-labs/mesh-security
    ...
    )
    ```
  
2. Add the following modules to `app.go`
    ```
    import (
    ... 
        "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity"
        meshseckeeper "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
        meshsectypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
        meshsecprov "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider"
        meshsecprovkeeper "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/keeper"
        meshsecprovtypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
    ...
    )
    ```
3. In `app.go`: Register the AppModule for the mesh security modules.
    ```
    ModuleBasics = module.NewBasicManager(
      ...
      meshsecurity.AppModuleBasic{},
      meshsecprov.AppModuleBasic{},
      ...
    )
    ```
4. In `app.go`: Add module account permissions:
    ```
    maccPerms = map[string][]string{
      ...
      meshsectypes.ModuleName: {authtypes.Minter, authtypes.Burner}
    }
    ```
5. In `app.go`: Add mesh security keepers.
    ```
    type App struct {
      ...
      MeshSecKeeper     *meshseckeeper.Keeper,
      MeshSecProvKeeper *meshsecprovkeeper.Keeper
      ...
    }
    ```
6. In `app.go`: Add mesh security store keys.
    ```
    keys := sdk.NewKVStoreKeys(
      ...
      meshsectypes.StoreKey,
      meshsecprovtypes.StoreKey,
      ...
    )
    ```
7. In `app.go`: Instantiate mesh security keepers
    ```
    app.MeshSecKeeper = meshseckeeper.NewKeeper(
		app.appCodec,
		keys[meshsectypes.StoreKey],
		memKeys[meshsectypes.MemStoreKey],
		app.BankKeeper,
		app.StakingKeeper,
		&app.WasmKeeper, // ensure this is a pointer as we instantiate the keeper a bit later
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
    app.MeshSecProvKeeper = meshsecprovkeeper.NewKeeper(
		appCodec,
		keys[meshsecprovtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		&app.WasmKeeper,
		app.StakingKeeper,
	)
    ```
8. In `app.go`: Add the mesh security modules to the app manager instantiation.
    ```
    app.mm = module.NewManager(
        ...
        meshsecurity.NewAppModule(appCodec, app.MeshSecKeeper),
        meshsecprov.NewAppModule(app.MeshSecProvKeeper)
        ...
    )
    ```
9. In `app.go`: Add the module as the final element to the following:
- SetOrderBeginBlockers
- SetOrderEndBlockers
- SetOrderInitGenesis
    ```
    // Add mesh security to begin blocker logic
    app.moduleManager.SetOrderBeginBlockers(
      ...
      meshsectypes.ModuleName,
      meshsecprovtypes.ModuleName,
      ...
    )

    // Add mesh security to end blocker logic
    app.moduleManager.SetOrderEndBlockers(
      ...
      meshsectypes.ModuleName,
      meshsecprovtypes.ModuleName,
      ...
    )

    // Add mesh security to init genesis logic
    app.moduleManager.SetOrderInitGenesis(
      ...
      meshsectypes.ModuleName,
      meshsecprovtypes.ModuleName,
      ...
    )
    ```
10. In `app.go`: Add the mesh security staking decorator to the slashing module.
    ```
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		// decorate the sdk keeper to capture all jail/ unjail events for MS
		meshseckeeper.NewStakingDecorator(app.StakingKeeper, app.MeshSecKeeper),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
    ```
11. In `app.go`: Add the mesh security hooks to the staking module.
    ```
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
            ...
			// register hook to capture valset updates
			app.MeshSecKeeper.Hooks()
		),
	)
    ```
12. In `app.go`: Add the mesh security hooks to the evidence module.
    ```
	evidenceKeeper := evidencekeeper.NewKeeper(
		...
		// decorate the SlashingKeeper to capture the tombstone event
		meshseckeeper.CaptureTombstoneDecorator(app.MeshSecKeeper, app.SlashingKeeper, app.StakingKeeper),
	)
    ```
13. In `app.go`: Add the mesh security wasm message handler decorator to the wasm module.
    ```
	meshMessageHandler := wasmkeeper.WithMessageHandlerDecorator(func(nested wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(
			// security layer for system integrity, should always be first in chain
			meshseckeeper.NewIntegrityHandler(app.MeshSecKeeper),
			nested,
			// append our custom message handler for mesh-security
			meshseckeeper.NewDefaultCustomMsgHandler(app.MeshSecKeeper),
			meshsecprovkeeper.CustomMessageDecorator(app.MeshSecProvKeeper),
		)
	})
	wasmOpts = append(wasmOpts, meshMessageHandler,
		// add support for the mesh-security queries
		wasmkeeper.WithQueryHandlerDecorator(meshseckeeper.NewQueryDecorator(app.MeshSecKeeper, app.SlashingKeeper)),
	)
    ```
