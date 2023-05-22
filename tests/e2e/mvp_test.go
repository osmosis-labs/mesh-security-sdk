package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

func TestMVP(t *testing.T) {
	// scenario:
	// given a provider chain A and a consumer chain B
	// when
	// then
	coord := NewIBCCoordinator(t, 2)
	providerChain := coord.GetChain(ibctesting.GetChainID(1))
	consumerChain := coord.GetChain(ibctesting.GetChainID(2))

	path := wasmibctesting.NewPath(providerChain, consumerChain)
	coord.SetupConnections(path)

	consumerApp := consumerChain.App.(*app.MeshApp)
	require.NotNil(t, consumerApp.MeshSecKeeper)
}
