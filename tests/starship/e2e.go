package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
)

var (
	wasmContractPath    string
	wasmContractGZipped bool
	configFile          string
)

func buildPathToWasm(fileName string) string {
	if wasmContractGZipped {
		fileName += ".gz"
	}
	return filepath.Join(wasmContractPath, fileName)
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

func voteAndPassGovProposal(t *testing.T, chain *ibctesting.TestChain, proposalID uint64) {
	vote := govv1.NewMsgVote(chain.SenderAccount.GetAddress(), proposalID, govv1.OptionYes, "testing")
	_, err := chain.SendMsgs(vote)
	require.NoError(t, err)

	chainApp := chain.App.(*app.MeshApp)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())

	coord := chain.Coordinator
	coord.IncrementTimeBy(*govParams.VotingPeriod)
	coord.CommitBlock(chain)

	rsp, err := chainApp.GovKeeper.Proposal(sdk.WrapSDKContext(chain.GetContext()), &govv1.QueryProposalRequest{ProposalId: proposalID})
	require.NoError(t, err)
	require.Equal(t, rsp.Proposal.Status, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
}

func InstantiateContract(t *testing.T, chain *Client, codeID uint64, initMsg []byte, funds ...sdk.Coin) []sdk.AccAddress {
	instantiateMsg := &wasmtypes.MsgInstantiateContract{
		Sender: chain.Address,
		Admin:  chain.Address,
		CodeID: codeID,
		Label:  "ibc-test",
		Msg:    initMsg,
		Funds:  funds,
	}

	r, err := chain.Client.SendMsg(context.Background(), instantiateMsg, "")
	require.NoError(t, err)
	require.NotEmpty(t, r)
	fmt.Printf("response for instantiate contract: %s", r)

	var addrs []sdk.AccAddress

	for _, event := range r.Logs[0].Events {
		if event.Type == "instantiate" {
			for _, attr := range event.Attributes {
				if attr.Key == "_contract_address" {
					addr := sdk.MustAccAddressFromBech32(attr.Value)
					addrs = append(addrs, addr)
				}
			}
		}
	}

	return addrs
}
