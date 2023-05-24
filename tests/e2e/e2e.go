package e2e

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	"github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
)

// NewIBCCoordinator initializes Coordinator with N meshd TestChain instances
func NewIBCCoordinator(t *testing.T, n int) *ibctesting.Coordinator {
	return ibctesting.NewCoordinatorX(t, n, func(t *testing.T, valSet *types.ValidatorSet, genAccs []authtypes.GenesisAccount, chainID string, opts []wasm.Option, balances ...banktypes.Balance) ibctesting.ChainApp {
		return app.SetupWithGenesisValSet(t, valSet, genAccs, chainID, opts, balances...)
	})
}

func submitGovProposal(t *testing.T, chain *ibctesting.TestChain, msgs ...sdk.Msg) uint64 {
	chainApp := chain.App.(*app.MeshApp)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())
	govMsg, err := govv1.NewMsgSubmitProposal(msgs, govParams.MinDeposit, chain.SenderAccount.GetAddress().String(), "", "my title", "my summary")
	require.NoError(t, err)
	rsp, err := chain.SendMsgs(govMsg)
	require.NoError(t, err)
	id := rsp.MsgResponses[0].GetCachedValue().(*govv1.MsgSubmitProposalResponse).ProposalId
	require.NotEmpty(t, id)
	return id
}

func voteAndPassGovProposal(t *testing.T, coord *ibctesting.Coordinator, chain *ibctesting.TestChain, proposalID uint64) {
	vote := govv1.NewMsgVote(chain.SenderAccount.GetAddress(), proposalID, govv1.OptionYes, "testing")
	_, err := chain.SendMsgs(vote)
	require.NoError(t, err)

	chainApp := chain.App.(*app.MeshApp)
	rsp, err := chainApp.GovKeeper.Proposal(sdk.WrapSDKContext(chain.GetContext()), &govv1.QueryProposalRequest{ProposalId: proposalID})
	require.NoError(t, err)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())

	coord.IncrementTimeBy(*govParams.VotingPeriod)
	coord.CommitBlock(chain)

	rsp, err = chainApp.GovKeeper.Proposal(sdk.WrapSDKContext(chain.GetContext()), &govv1.QueryProposalRequest{ProposalId: proposalID})
	require.NoError(t, err)
	require.Equal(t, rsp.Proposal.Status, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
}
