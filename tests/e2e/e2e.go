package e2e

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	"github.com/cometbft/cometbft/types"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
	types3 "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
)

// NewIBCCoordinator initializes Coordinator with N meshd TestChain instances
func NewIBCCoordinator(t *testing.T, n int) *ibctesting.Coordinator {
	return ibctesting.NewCoordinatorX(t, n, func(t *testing.T, valSet *types.ValidatorSet, genAccs []types2.GenesisAccount, chainID string, opts []wasm.Option, balances ...types3.Balance) ibctesting.ChainApp {
		return app.SetupWithGenesisValSet(t, valSet, genAccs, chainID, opts, balances...)
	})
}
