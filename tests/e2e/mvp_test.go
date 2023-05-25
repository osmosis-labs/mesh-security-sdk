package e2e

import (
	"testing"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	"github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestMVP(t *testing.T) {
	// scenario:
	// given a provider chain P and a consumer chain C
	// when
	// then
	coord := NewIBCCoordinator(t, 2)
	providerChain := coord.GetChain(ibctesting.GetChainID(1))
	consumerChain := coord.GetChain(ibctesting.GetChainID(2))
	consumerApp := consumerChain.App.(*app.MeshApp)

	path := wasmibctesting.NewPath(providerChain, consumerChain)
	coord.SetupConnections(path)
	randomContract := sdk.AccAddress(rand.Bytes(32)) // todo: use real contract address

	// when mesh security via gov proposal
	payloadMsg := &types.MsgSetVirtualStakingMaxCap{
		Authority: consumerApp.MeshSecKeeper.GetAuthority(),
		Contract:  randomContract.String(),
		MaxCap:    sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	}

	proposalID := submitGovProposal(t, consumerChain, payloadMsg)
	voteAndPassGovProposal(t, coord, consumerChain, proposalID)

	// then the max cap limit is persisted
	q := baseapp.QueryServiceTestHelper{GRPCQueryRouter: consumerApp.GRPCQueryRouter(), Ctx: consumerChain.GetContext()}
	var rsp types.QueryVirtualStakingMaxCapResponse
	err := q.Invoke(nil, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCap", &types.QueryVirtualStakingMaxCapRequest{Address: randomContract.String()}, &rsp)
	require.NoError(t, err)
	assert.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000), rsp.Limit)
}
