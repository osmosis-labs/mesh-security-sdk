package e2e

import (
	"path/filepath"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cometbft/cometbft/types"
	types2 "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting2 "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	consumerapp "github.com/osmosis-labs/mesh-security-sdk/demo/app/consumer"
	providerapp "github.com/osmosis-labs/mesh-security-sdk/demo/app/provider"
)

var (
	wasmContractPath    string
	wasmContractGZipped bool
)

func buildPathToWasm(fileName string) string {
	if wasmContractGZipped {
		fileName += ".gz"
	}
	return filepath.Join(wasmContractPath, fileName)
}

// NewIBCCoordinator initializes Coordinator with N meshd TestChain instances
func NewIBCCoordinator(t *testing.T, n int, opts ...[]wasmkeeper.Option) *ibctesting.Coordinator {
	return ibctesting.NewCoordinatorX(t, n,
		func(t *testing.T, valSet *types.ValidatorSet, genAccs []authtypes.GenesisAccount, chainID string, opts []wasmkeeper.Option, balances ...banktypes.Balance) ibctesting.ChainApp {
			if chainID == ibctesting.GetChainID(1) {
				return consumerapp.SetupWithGenesisValSet(t, valSet, genAccs, chainID, opts, balances...)
			} else {
				return providerapp.SetupWithGenesisValSet(t, valSet, genAccs, chainID, opts, balances...)
			}
		},
		opts...,
	)
}

func submitProviderGovProposal(t *testing.T, chain *TestChain, msgs ...sdk.Msg) uint64 {
	chainApp := chain.App.(*providerapp.MeshProviderApp)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())
	govMsg, err := govv1.NewMsgSubmitProposal(msgs, govParams.MinDeposit, chain.SenderAccount.GetAddress().String(), "", "my title", "my summary")
	require.NoError(t, err)
	rsp, err := chain.SendMsgs(govMsg)
	require.NoError(t, err)
	id := rsp.MsgResponses[0].GetCachedValue().(*govv1.MsgSubmitProposalResponse).ProposalId
	require.NotEmpty(t, id)
	return id
}

func voteAndPassProviderGovProposal(t *testing.T, chain *TestChain, proposalID uint64) {
	vote := govv1.NewMsgVote(chain.SenderAccount.GetAddress(), proposalID, govv1.OptionYes, "testing")
	_, err := chain.SendMsgs(vote)
	require.NoError(t, err)

	chainApp := chain.App.(*providerapp.MeshProviderApp)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())

	coord := chain.Coordinator
	coord.IncrementTimeBy(*govParams.VotingPeriod)
	coord.CommitBlock(chain.IBCTestChain())

	rsp, err := chainApp.GovKeeper.Proposal(sdk.WrapSDKContext(chain.GetContext()), &govv1.QueryProposalRequest{ProposalId: proposalID})
	require.NoError(t, err)
	require.Equal(t, rsp.Proposal.Status, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
}

func submitConsumerGovProposal(t *testing.T, chain *TestChain, msgs ...sdk.Msg) uint64 {
	chainApp := chain.App.(*consumerapp.MeshConsumerApp)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())
	govMsg, err := govv1.NewMsgSubmitProposal(msgs, govParams.MinDeposit, chain.SenderAccount.GetAddress().String(), "", "my title", "my summary")
	require.NoError(t, err)
	rsp, err := chain.SendMsgs(govMsg)
	require.NoError(t, err)
	id := rsp.MsgResponses[0].GetCachedValue().(*govv1.MsgSubmitProposalResponse).ProposalId
	require.NotEmpty(t, id)
	return id
}

func voteAndPassConsumerGovProposal(t *testing.T, chain *TestChain, proposalID uint64) {
	vote := govv1.NewMsgVote(chain.SenderAccount.GetAddress(), proposalID, govv1.OptionYes, "testing")
	_, err := chain.SendMsgs(vote)
	require.NoError(t, err)

	chainApp := chain.App.(*consumerapp.MeshConsumerApp)
	govParams := chainApp.GovKeeper.GetParams(chain.GetContext())

	coord := chain.Coordinator
	coord.IncrementTimeBy(*govParams.VotingPeriod)
	coord.CommitBlock(chain.IBCTestChain())

	rsp, err := chainApp.GovKeeper.Proposal(sdk.WrapSDKContext(chain.GetContext()), &govv1.QueryProposalRequest{ProposalId: proposalID})
	require.NoError(t, err)
	require.Equal(t, rsp.Proposal.Status, govv1.ProposalStatus_PROPOSAL_STATUS_PASSED)
}

func InstantiateContract(t *testing.T, chain *TestChain, codeID uint64, initMsg []byte, funds ...sdk.Coin) sdk.AccAddress {
	instantiateMsg := &wasmtypes.MsgInstantiateContract{
		Sender: chain.SenderAccount.GetAddress().String(),
		Admin:  chain.SenderAccount.GetAddress().String(),
		CodeID: codeID,
		Label:  "ibc-test",
		Msg:    initMsg,
		Funds:  funds,
	}

	r, err := chain.SendMsgs(instantiateMsg)
	require.NoError(t, err)
	require.Len(t, r.MsgResponses, 1)
	require.NotEmpty(t, r.MsgResponses[0].GetCachedValue())
	pExecResp := r.MsgResponses[0].GetCachedValue().(*wasmtypes.MsgInstantiateContractResponse)
	a, err := sdk.AccAddressFromBech32(pExecResp.Address)
	require.NoError(t, err)
	return a
}

type example struct {
	Coordinator      *ibctesting.Coordinator
	ConsumerChain    *TestChain
	ProviderChain    *TestChain
	ConsumerApp      *consumerapp.MeshConsumerApp
	ProviderApp      *providerapp.MeshProviderApp
	IbcPath          *ibctesting.Path
	ProviderDenom    string
	ConsumerDenom    string
	MyProvChainActor string
	MaxRetrieve      uint16
}

func setupExampleChains(t *testing.T) example {
	coord := NewIBCCoordinator(t, 2)
	consChain := coord.GetChain(ibctesting2.GetChainID(1))
	provChain := coord.GetChain(ibctesting2.GetChainID(2))
	return example{
		Coordinator:      coord,
		ConsumerChain:    NewTestChain(t, consChain),
		ProviderChain:    NewTestChain(t, provChain),
		ConsumerApp:      consChain.App.(*consumerapp.MeshConsumerApp),
		ProviderApp:      provChain.App.(*providerapp.MeshProviderApp),
		IbcPath:          ibctesting.NewPath(consChain, provChain),
		ProviderDenom:    sdk.DefaultBondDenom,
		ConsumerDenom:    sdk.DefaultBondDenom,
		MyProvChainActor: provChain.SenderAccount.GetAddress().String(),
		MaxRetrieve:      50,
	}
}

func setupMeshSecurity(t *testing.T, x example) (*TestConsumerClient, ConsumerContract, *TestProviderClient) {
	x.Coordinator.SetupConnections(x.IbcPath)

	// setup contracts on both chains
	consumerCli := NewConsumerClient(t, x.ConsumerChain)
	consumerContracts := consumerCli.BootstrapContracts(x)
	converterPortID := wasmkeeper.PortIDForContract(consumerContracts.converter)
	// add some fees so that we can distribute something
	x.ConsumerChain.DefaultMsgFees = sdk.NewCoins(sdk.NewCoin(x.ConsumerDenom, math.NewInt(1_000_000)))

	providerCli := NewProviderClient(t, x.ProviderChain)
	providerContracts := providerCli.BootstrapContracts(x.ProviderApp, x.IbcPath.EndpointA.ConnectionID, converterPortID)

	// setup ibc control path: consumer -> provider (direction matters)
	x.IbcPath.EndpointB.ChannelConfig = &ibctesting2.ChannelConfig{
		PortID: wasmkeeper.PortIDForContract(providerContracts.ExternalStaking),
		Order:  types2.UNORDERED,
	}
	x.IbcPath.EndpointA.ChannelConfig = &ibctesting2.ChannelConfig{
		PortID: converterPortID,
		Order:  types2.UNORDERED,
	}
	x.Coordinator.CreateChannels(x.IbcPath)

	// when ibc package is relayed
	require.NotEmpty(t, x.ConsumerChain.PendingSendPackets)
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	consumerCli.MustEnableVirtualStaking(sdk.NewInt64Coin(x.ConsumerDenom, 1_000_000_000))
	return consumerCli, consumerContracts, providerCli
}
