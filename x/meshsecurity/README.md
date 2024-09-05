# Meshi-security
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
    github.com/osmosis-labs/mesh security v<VERSION>
    ...
    )
    ```
  **Note:** The version of the mesh security module will depend on which version of the Cosmos SDK your chain is using. If in doubt about which version to use, please consult the documentation: https://github.com/osmosis-labs/mesh security
  
2. Add the following modules to `app.go`
    ```
    import (
    ... 
        feeabsmodule "github.com/notional-labs/mesh security/v2/x/feeabs"
        feeabskeeper "github.com/notional-labs/mesh security/v2/x/feeabs/keeper"
        feeabstypes "github.com/notional-labs/mesh security/v2/x/feeabs/types"
    ...
    )
    ```
3. In `app.go`: Register the AppModule for the fee middleware module.
    ```
    ModuleBasics = module.NewBasicManager(
      ...
      feeabsmodule.AppModuleBasic{},
      ...
    )
    ```
4. In `app.go`: Add module account permissions for the fee abstractions.
    ```
    maccPerms = map[string][]string{
      ...
      feeabsmodule.ModuleName:            nil,
    }
    // module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		feeabstypes.ModuleName: true,
	}
    ```
5. In `app.go`: Add fee abstraction keeper.
    ```
    type App struct {
      ...
      FeeabsKeeper feeabskeeper.Keeper
      ...
    }
    ```
6. In `app.go`: Add fee abstraction store key.
    ```
    keys := sdk.NewKVStoreKeys(
      ...
      feeabstypes.StoreKey,
      ...
    )
    ```
7. In `app.go`: Instantiate Fee abstraction keeper
    ```
    app.FeeabsKeeper = feeabskeeper.NewKeeper(
      appCodec,
      keys[feeabstypes.StoreKey],
      keys[feeabstypes.MemStoreKey],
      app.GetSubspace(feeabstypes.ModuleName),
      app.StakingKeeper,
      app.AccountKeeper,
      app.BankKeeper,
      app.TransferKeeper,
      app.IBCKeeper.ChannelKeeper,
      &app.IBCKeeper.PortKeeper,
      scopedFeeabsKeeper,
    )
    ```
8. In `app.go`: Add the IBC router.
    ```
    feeabsIBCModule := feeabsmodule.NewIBCModule(appCodec, app.FeeabsKeeper)
    
    ibcRouter := porttypes.NewRouter()
    ibcRouter.
    ...
    AddRoute(feeabstypes.ModuleName, feeabsIBCModule)
    ...
    ```
9. In `app.go`: Add the mesh security module to the app manager and simulation manager instantiations.
    ```
    app.mm = module.NewManager(
        ...
        feeabsModule := feeabsmodule.NewAppModule(appCodec, app.FeeabsKeeper),
        ...
    )
    ```
    ```
    app.sm = module.NewSimulationManager(
        ...
        transferModule,
        feeabsModule := feeabsmodule.NewAppModule(appCodec, app.FeeabsKeeper),
        ...
    )
    ```
10. In `app.go`: Add the module as the final element to the following:
- SetOrderBeginBlockers
- SetOrderEndBlockers
- SetOrderInitGenesis
    ```
    // Add fee abstraction to begin blocker logic
    app.moduleManager.SetOrderBeginBlockers(
      ...
      feeabstypes.ModuleName,
      ...
    )

    // Add fee abstraction to end blocker logic
    app.moduleManager.SetOrderEndBlockers(
      ...
      feeabstypes.ModuleName,
      ...
    )

    // Add fee abstraction to init genesis logic
    app.moduleManager.SetOrderInitGenesis(
      ...
      feeabstypes.ModuleName,
      ...
    )
    ```
11. In `app.go`: Allow module account address.
    ```
    func (app *FeeAbs) ModuleAccountAddrs() map[string]bool {
	    blockedAddrs := make(map[string]bool)

	    accs := make([]string, 0, len(maccPerms))
	    for k := range maccPerms {
		    accs = append(accs, k)
	    }
	    sort.Strings(accs)

	    for _, acc := range accs {
		    blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	    }

	    return blockedAddrs
    }
    ```
12. In `app.go`: Add to Param keeper.
    ```
    func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey sdk.StoreKey) paramskeeper.Keeper {
	    paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)
        ...
	    paramsKeeper.Subspace(feeabstypes.ModuleName)
        ...
	    return paramsKeeper
    }
    ```
13. Modified Fee Antehandler

    To allow for this, we use modified versions of `MempoolFeeDecorator` and `DeductFeeDecorate`. In these ante handlers, IBC tokens are swapped to the native token before the next fee handler logic is executed.

    If a blockchain uses the Fee Abstraction module, it is necessary to replace the `MempoolFeeDecorator` and `DeductFeeDecorate` with the `FeeAbstrationMempoolFeeDecorator` and `FeeAbstractionDeductFeeDecorate`, respectively. These can be found in `app/ante.go`, and should be implemented as below:
    
    Example:
    ```
    anteDecorators := []sdk.AnteDecorator{
      ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
      ante.NewRejectExtensionOptionsDecorator(),
      feeabsante.NewFeeAbstrationMempoolFeeDecorator(options.FeeAbskeeper),
      ante.NewValidateBasicDecorator(),
      ante.NewTxTimeoutHeightDecorator(),
      ante.NewValidateMemoDecorator(options.AccountKeeper),
      ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
      feeabsante.NewFeeAbstractionDeductFeeDecorate(options.AccountKeeper, options.BankKeeper, options.FeeAbskeeper, options.FeegrantKeeper),
      // SetPubKeyDecorator must be called before all signature verification decorators
      ante.NewSetPubKeyDecorator(options.AccountKeeper),
      ante.NewValidateSigCountDecorator(options.AccountKeeper),
      ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
      ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
      ante.NewIncrementSequenceDecorator(options.AccountKeeper),
      ibcante.NewAnteDecorator(options.IBCKeeper),
     }
    ```
    