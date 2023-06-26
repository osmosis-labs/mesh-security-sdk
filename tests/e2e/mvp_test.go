package e2e

import (
	"encoding/base64"
	"fmt"
	"testing"

	"cosmossdk.io/math"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestMVP(t *testing.T) {
	// scenario:
	// given a provider chain P and a consumer chain C
	// when	 a user on P deposits an amount as collateral in the vault contract
	// then	 it can be used as "lien" to stake local and to "cross stake" on chain C
	// when  an amount is "cross staked" to a validator on chain C
	//
	// not fully implemented:
	//
	// and	 the ibc package is relayed
	// then  the amount is converted into an amount in the chain C bonding token
	// and   scheduled to be staked as synthetic token on the validator
	// when  the next epoch is executed on chain C
	// then  the synthetic tokens are minted and staked
	// when  the user on chain P starts an undelegate
	// ...

	var (
		coord         = NewIBCCoordinator(t, 2)
		consumerChain = coord.GetChain(ibctesting.GetChainID(2))
		providerChain = coord.GetChain(ibctesting.GetChainID(1))
		consumerApp   = consumerChain.App.(*app.MeshApp)
		ibcPath       = wasmibctesting.NewPath(consumerChain, providerChain)
	)
	msModule := consumerApp.ModuleManager.Modules[types.ModuleName].(*meshsecurity.AppModule)
	msModule.SetAsyncTaskRspHandler(meshsecurity.PanicOnErrorExecutionResponseHandler()) // fail fast in test

	coord.SetupConnections(ibcPath)

	// setup contracts on both chains
	consClient := NewConsumerClient(t, consumerChain)
	consClient.BootstrapContracts()
	consumerContracts := consClient.consumerContracts
	converterPortID := wasmkeeper.PortIDForContract(consumerContracts.converter)
	providerContracts := bootstrapProviderContracts(t, providerChain, ibcPath.EndpointA.ConnectionID, converterPortID)

	// setup ibc control path: consumer -> provider (direction matters)
	ibcPath.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
		PortID: wasmkeeper.PortIDForContract(providerContracts.externalStaking),
		Order:  channeltypes.UNORDERED,
	}
	ibcPath.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
		PortID: converterPortID,
		Order:  channeltypes.UNORDERED,
	}
	coord.CreateChannels(ibcPath)

	// when ibc package is relayed
	require.NotEmpty(t, consumerChain.PendingSendPackets)
	coord.RelayAndAckPendingPackets(ibcPath)

	// then the active set should be stored in the ext staking contract
	queryProvContract := Querier(t, providerChain)
	// and contain all active validator addresses
	qRsp := queryProvContract(providerContracts.externalStaking.String(), Query{"list_remote_validators": {}})
	require.Len(t, qRsp["validators"], 4, qRsp)
	for _, v := range consumerChain.Vals.Validators {
		require.Contains(t, qRsp["validators"], sdk.ValAddress(v.Address).String())
	}

	// ----------------------------
	// ensure nothing staked by the virtual staking contract yet
	myExtValidator := sdk.ValAddress(consumerChain.Vals.Validators[0].Address)
	myExtValidatorAddr := myExtValidator.String()
	_, found := consumerApp.StakingKeeper.GetDelegation(consumerChain.GetContext(), consumerContracts.staking, myExtValidator)
	require.False(t, found)

	// add authority to mint/burn virtual tokens gov proposal
	payloadMsg := &types.MsgSetVirtualStakingMaxCap{
		Authority: consumerApp.MeshSecKeeper.GetAuthority(),
		Contract:  consumerContracts.staking.String(),
		MaxCap:    sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	}

	proposalID := submitGovProposal(t, consumerChain, payloadMsg)
	voteAndPassGovProposal(t, coord, consumerChain, proposalID)

	// then the max cap limit is persisted
	q := baseapp.QueryServiceTestHelper{GRPCQueryRouter: consumerApp.GRPCQueryRouter(), Ctx: consumerChain.GetContext()}
	var rsp types.QueryVirtualStakingMaxCapLimitResponse
	err := q.Invoke(nil, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimit", &types.QueryVirtualStakingMaxCapLimitRequest{Address: consumerContracts.staking.String()}, &rsp)
	require.NoError(t, err)
	assert.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000), rsp.Cap)

	// provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	_, err = providerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   providerChain.SenderAccount.GetAddress().String(),
		Contract: providerContracts.vault.String(),
		Msg:      []byte(`{"bond":{}}`),
		Funds:    sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)),
	})
	require.NoError(t, err)
	// then query contract state
	qRsp = queryProvContract(providerContracts.vault.String(), Query{
		"account": {"account": providerChain.SenderAccount.GetAddress().String()},
	})
	assert.Equal(t, "100000000", qRsp["free"], qRsp)

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	myLocalValidatorAddr := sdk.ValAddress(providerChain.Vals.Validators[0].Address).String()
	stakeMsg := fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr)
	_, err = providerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   providerChain.SenderAccount.GetAddress().String(),
		Contract: providerContracts.vault.String(),
		Msg: []byte(fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
			sdk.DefaultBondDenom, 50_000_000,
			base64.StdEncoding.EncodeToString([]byte(stakeMsg)))),
	})
	require.NoError(t, err)

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	stakeMsg = fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr)
	_, err = providerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   providerChain.SenderAccount.GetAddress().String(),
		Contract: providerContracts.vault.String(),
		Msg: []byte(fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
			providerContracts.externalStaking.String(),
			sdk.DefaultBondDenom, 40_000_000,
			base64.StdEncoding.EncodeToString([]byte(stakeMsg)))),
	})
	require.NoError(t, err)
	require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))

	// then
	qRsp = queryProvContract(providerContracts.externalStaking.String(), Query{
		"stake": {
			"user":      providerChain.SenderAccount.GetAddress().String(),
			"validator": myExtValidatorAddr,
		},
	})
	assert.Equal(t, "40000000", qRsp["stake"], qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])

	// consumer chain
	// ====================

	// then delegated amount is not updated before the epoch

	assertTotalDelegated := func(expTotalDelegated math.Int) {
		usedAmount := consumerApp.MeshSecKeeper.GetTotalDelegated(consumerChain.GetContext(), consumerContracts.staking)
		assert.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, expTotalDelegated), usedAmount)
	}
	assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	doRebalance := func() {
		epochLength := consumerApp.MeshSecKeeper.GetRebalanceEpochLength(consumerChain.GetContext())
		coord.CommitNBlocks(consumerChain, epochLength)
	}
	doRebalance() // execute epoch

	// then the total delegated amount is updated
	assertTotalDelegated(math.NewInt(18_000_000)) // 40_000_000 /2 * (1 -0.1)

	// and the delegated amount is updated for the validator
	assertShare := func(exp int64) {
		del, found := consumerApp.StakingKeeper.GetDelegation(consumerChain.GetContext(), consumerContracts.staking, myExtValidator)
		require.True(t, found)
		assert.Equal(t, math.LegacyNewDec(exp), del.Shares)
	}
	assertShare(18) // 18_000_000 / 1_000_000 # default sdk factor

	// provider chain
	// ==============
	//
	// Cross Stake - A user undelegates
	_, err = providerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   providerChain.SenderAccount.GetAddress().String(),
		Contract: providerContracts.externalStaking.String(),
		Msg:      []byte(fmt.Sprintf(`{"unstake":{"validator":"%s", "amount":{"denom":"%s", "amount":"20000000"}}}`, myExtValidator.String(), sdk.DefaultBondDenom)),
	})
	require.NoError(t, err)
	require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))

	// then
	qRsp = queryProvContract(providerContracts.externalStaking.String(), Query{
		"stake": {
			"user":      providerChain.SenderAccount.GetAddress().String(),
			"validator": myExtValidatorAddr,
		},
	})
	assert.Equal(t, "20000000", qRsp["stake"], qRsp)
	require.Len(t, qRsp["pending_unbonds"], 1)
	unbonds := qRsp["pending_unbonds"].([]any)[0].(map[string]any)
	assert.Equal(t, "20000000", unbonds["amount"], qRsp)

	// consumer chain
	// ====================

	doRebalance() // execute epoch

	// then the total delegated amount is updated
	assertTotalDelegated(math.NewInt(9000000)) // (40_000_000 - 20_000_000) /2 * (1 -0.1)
	assertShare(9)                             // 20_000_000 / 1_000_000 # default sdk factor
}
