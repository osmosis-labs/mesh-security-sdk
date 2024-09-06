# Mesh security provider
Cosmos module implementation

# Integrate the mesh security provider module

## Prerequisites 
Projects that want to integrate the meshsecurityprovider module onto their Cosmos SDK chain must enable the following modules:
- [x/staking](https://github.com/cosmos/cosmos-sdk/tree/main/x/staking)
- [x/auth](https://github.com/cosmos/cosmos-sdk/tree/main/x/auth)
- [x/bank](https://github.com/cosmos/cosmos-sdk/tree/main/x/bank)
- [x/wasm](github.com/CosmWasm/wasmd/x/wasm)

## Configuring and Adding Module
1. Add the mesh security package to the go.mod and install it.
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
        meshsecprov "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider"
        meshsecprovkeeper "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/keeper"
        meshsecprovtypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
    ...
    )
    ```
3. In `app.go`: Register the AppModule for the mesh security provider module.
    ```
    ModuleBasics = module.NewBasicManager(
      ...
      meshsecprov.AppModuleBasic{},
      ...
    )
    ```
4. In `app.go`: Add mesh security provider keeper.
    ```
    type App struct {
      ...
      MeshSecProvKeeper *meshsecprovkeeper.Keeper
      ...
    }
    ```
5. In `app.go`: Add mesh security provider store key.
    ```
    keys := sdk.NewKVStoreKeys(
      ...
      meshsecprovtypes.StoreKey,
      ...
    )
    ```
6. In `app.go`: Instantiate mesh security provider keeper
    ```
	app.MeshSecProvKeeper = meshsecprovkeeper.NewKeeper(
		appCodec,
		keys[meshsecprovtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.BankKeeper,
		&app.WasmKeeper,
		app.StakingKeeper,
	)
    ```
7. In `app.go`: Add the mesh security provider module to the app manager instantiation.
    ```
    app.mm = module.NewManager(
        ...
        meshsecprov.NewAppModule(app.MeshSecProvKeeper)
        ...
    )
    ```
8. In `app.go`: Add the module as the final element to the following:
- SetOrderBeginBlockers
- SetOrderEndBlockers
- SetOrderInitGenesis
    ```
    // Add mesh security to begin blocker logic
    app.moduleManager.SetOrderBeginBlockers(
      ...
      meshsecprovtypes.ModuleName,
      ...
    )

    // Add mesh security to end blocker logic
    app.moduleManager.SetOrderEndBlockers(
      ...
      meshsecprovtypes.ModuleName,
      ...
    )

    // Add mesh security to init genesis logic
    app.moduleManager.SetOrderInitGenesis(
      ...
      meshsecprovtypes.ModuleName,
      ...
    )
    ```
9. In `app.go`: Add the mesh security wasm message handler decorator to the wasm module.
    ```
	meshMessageHandler := wasmkeeper.WithMessageHandlerDecorator(func(nested wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(
			meshsecprovkeeper.CustomMessageDecorator(app.MeshSecProvKeeper),
			nested,
		)
	})
	wasmOpts = append(wasmOpts, meshMessageHandler)
    ```