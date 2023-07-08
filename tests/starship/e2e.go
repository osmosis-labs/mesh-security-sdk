package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"
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

func submitGovProposal(t *testing.T, chain *Client, msgs ...sdk.Msg) uint64 {
	// fetch gov params from the local
	initialDeposit := sdk.NewCoins(sdk.NewCoin("ustake", sdk.NewInt(10000000)))
	govMsg, err := govv1.NewMsgSubmitProposal(msgs, initialDeposit, chain.Address, "", "my title", "my summary")
	require.NoError(t, err)
	rsp, err := chain.Client.SendMsg(context.Background(), govMsg, "")
	require.NoError(t, err)
	// Fetch the proposal id
	proposalID := 0
	for _, event := range rsp.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if attr.Key == "proposal_id" {
					value, err := strconv.Atoi(attr.Value)
					require.NoError(t, err)
					proposalID = value
				}
			}
		}
	}
	require.NotEmpty(t, proposalID)
	fmt.Printf("submitted gov proposalID: %v\n", proposalID)
	return uint64(proposalID)
}

func voteAndPassGovProposal(t *testing.T, chain *Client, proposalID uint64) {
	vote := &govv1.MsgVote{
		ProposalId: proposalID,
		Voter:      chain.Address,
		Option:     govv1.OptionYes,
		Metadata:   "testing",
	}
	res, err := chain.Client.SendMsg(context.Background(), vote, "")
	require.NoError(t, err)
	require.NotEmpty(t, res.TxHash)

	fmt.Printf("submitted vote for proposalid: %v from %v\n", proposalID, chain.Address)

	queryProposal := &govv1.QueryProposalRequest{
		ProposalId: proposalID,
	}

	var proposal *govv1.QueryProposalResponse
	require.Eventuallyf(t,
		func() bool {
			proposal, err = govv1.NewQueryClient(chain.Client).Proposal(context.Background(), queryProposal)
			if err != nil {
				return false
			}
			if proposal.Proposal.Status >= govv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
				return true
			}
			return false
		},
		300*time.Second,
		time.Second,
		"waited for too long, still proposal did not pass",
	)
	fmt.Print("proposal sucessfully passed...")
	require.NotNil(t, proposal)
	require.Equal(t, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Proposal.Status)
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
