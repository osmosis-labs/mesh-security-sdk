package e2e

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
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

	// instantiate virtual staking contract
	codeID := consumerChain.StoreCodeFile("testdata/mesh_virtual_staking.wasm.gz").CodeID
	initMsg := []byte(fmt.Sprintf(`{"denom": %q}`, sdk.DefaultBondDenom))
	ibctesting.TestCoin = sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt())
	stakingContract := InstantiateContract(t, consumerChain, codeID, initMsg)

	// ensure nothing staked
	valAddr := sdk.ValAddress(consumerChain.Vals.Validators[0].Address)
	_, found := consumerApp.StakingKeeper.GetDelegation(consumerChain.GetContext(), stakingContract, valAddr)
	require.False(t, found)

	// add authority to mint/burn virtual tokens gov proposal
	payloadMsg := &types.MsgSetVirtualStakingMaxCap{
		Authority: consumerApp.MeshSecKeeper.GetAuthority(),
		Contract:  stakingContract.String(),
		MaxCap:    sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	}

	proposalID := submitGovProposal(t, consumerChain, payloadMsg)
	voteAndPassGovProposal(t, coord, consumerChain, proposalID)

	// then the max cap limit is persisted
	q := baseapp.QueryServiceTestHelper{GRPCQueryRouter: consumerApp.GRPCQueryRouter(), Ctx: consumerChain.GetContext()}
	var rsp types.QueryVirtualStakingMaxCapResponse
	err := q.Invoke(nil, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCap", &types.QueryVirtualStakingMaxCapRequest{Address: stakingContract.String()}, &rsp)
	require.NoError(t, err)
	assert.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000), rsp.Limit)

	// when staking contract is instructed to bond tokens
	doExec := func(payload string) {
		_, err = consumerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
			Sender:   consumerChain.SenderAccount.GetAddress().String(),
			Contract: stakingContract.String(),
			Msg:      []byte(payload),
		})
		require.NoError(t, err)
	}
	doExec(fmt.Sprintf(`{"bond":{"validator":"%s", "amount":{"denom":"%s", "amount":"10000000"}}}`, valAddr.String(), sdk.DefaultBondDenom))

	// then delegated amount is not updated before the epoch

	assertTotalDelegated := func(expTotalDelegated math.Int) {
		usedAmount := consumerApp.MeshSecKeeper.GetTotalDelegated(consumerChain.GetContext(), stakingContract)
		assert.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, expTotalDelegated), usedAmount)
	}
	assertTotalDelegated(math.ZeroInt())

	// when an epoch ends, the delegation rebalance is triggered
	doRebalance := func() {
		rebalanceMsg := []byte(`{"rebalance":{}}`)
		_, err = consumerApp.WasmKeeper.Sudo(consumerChain.GetContext(), stakingContract, rebalanceMsg)
		require.NoError(t, err)
	}
	doRebalance()

	// then the total delegated amount is updated
	assertTotalDelegated(math.NewInt(10_000_000))

	// and the delegated amount is updated for the validator
	assertShare := func(exp int64) {
		del, found := consumerApp.StakingKeeper.GetDelegation(consumerChain.GetContext(), stakingContract, valAddr)
		require.True(t, found)
		assert.Equal(t, math.LegacyNewDec(exp), del.Shares)
	}
	assertShare(10)

	// when undelegated
	doExec(fmt.Sprintf(`{"unbond":{"validator":"%s", "amount":{"denom":"%s", "amount":"1000000"}}}`, valAddr.String(), sdk.DefaultBondDenom))
	// when an epoch ends, the delegation rebalance is triggered
	doRebalance()
	// then undelegated and burned
	assertTotalDelegated(math.NewInt(9_000_000))
	assertShare(9)
}
