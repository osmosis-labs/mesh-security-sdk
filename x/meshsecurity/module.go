package meshsecurity

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	meshseckeeper "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the mesh-security module.
type AppModuleBasic struct{}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
}

// Name returns the meshsecurity module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the mesh-security
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis performs genesis state validation for the mesh-security module.
func (b AppModuleBasic) ValidateGenesis(marshaler codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

// GetTxCmd returns the root tx command for the mesh-security module.
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command for the mesh-security module.
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// RegisterInterfaces implements InterfaceModule
func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements an application module interface
type AppModule struct {
	AppModuleBasic
	cdc codec.Codec
}

func NewAppModule(cdc codec.Codec, m *meshseckeeper.Keeper) *AppModule {
	return &AppModule{cdc: cdc}
}

func (a AppModule) GenerateGenesisState(input *module.SimulationState) {
}

func (a AppModule) RegisterStoreDecoder(registry sdk.StoreDecoderRegistry) {
}

func (a AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
