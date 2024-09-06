# Mesh-security
Cosmos module implementation

# Integrate the mesh security module

## Prerequisites 
Projects that want to integrate the meshsecurity module onto their Cosmos SDK chain must enable the following modules:
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
        "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity"
        meshseckeeper "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
        meshsectypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
    ...
    )
    ```
3. In `app.go`: Register the AppModule for the mesh security module.
    ```
    ModuleBasics = module.NewBasicManager(
      ...
      meshsecurity.AppModuleBasic{},
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
5. In `app.go`: Add mesh security keeper.
    ```
    type App struct {
      ...
      MeshSecKeeper *meshseckeeper.Keeper
      ...
    }
    ```
6. In `app.go`: Add mesh security store key.
    ```
    keys := sdk.NewKVStoreKeys(
      ...
      meshsectypes.StoreKey,
      ...
    )
    ```
7. In `app.go`: Instantiate mesh security keeper
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
    ```
8. In `app.go`: Add the mesh security module to the app manager instantiation.
    ```
    app.mm = module.NewManager(
        ...
        meshsecurity.NewAppModule(appCodec, app.MeshSecKeeper),
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
      ...
    )

    // Add mesh security to end blocker logic
    app.moduleManager.SetOrderEndBlockers(
      ...
      meshsectypes.ModuleName,
      ...
    )

    // Add mesh security to init genesis logic
    app.moduleManager.SetOrderInitGenesis(
      ...
      meshsectypes.ModuleName,
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
            meshseckeeper.NewIntegrityHandler(app.MeshSecKeeper),
            nested,
            meshseckeeper.NewDefaultCustomMsgHandler(app.MeshSecKeeper),
        )
    })
    wasmOpts = append(wasmOpts, meshMessageHandler,
        // add support for the mesh-security queries
        wasmkeeper.WithQueryHandlerDecorator(meshseckeeper.NewQueryDecorator(app.MeshSecKeeper, app.SlashingKeeper)),
    )
    ```