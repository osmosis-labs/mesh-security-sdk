package e2e

import (
	"encoding/base64"
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

	coord := NewIBCCoordinator(t, 2)
	providerChain := coord.GetChain(ibctesting.GetChainID(1))
	providerContracts := bootstrapProviderContracts(t, providerChain)
	consumerChain := coord.GetChain(ibctesting.GetChainID(2))
	consumerApp := consumerChain.App.(*app.MeshApp)

	path := wasmibctesting.NewPath(providerChain, consumerChain)
	coord.SetupConnections(path)

	consumerContracts := bootstrapConsumerContracts(t, consumerChain)
	// ensure nothing staked
	valAddr := sdk.ValAddress(consumerChain.Vals.Validators[0].Address)
	_, found := consumerApp.StakingKeeper.GetDelegation(consumerChain.GetContext(), consumerContracts.staking, valAddr)
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
	queryProvContract := func(contract string, query Query) map[string]any {
		qRsp := make(map[string]any)
		err = providerChain.SmartQuery(contract, query, &qRsp)
		require.NoError(t, err)
		return qRsp
	}

	qRsp := queryProvContract(providerContracts.vault.String(), Query{
		"account": {"account": providerChain.SenderAccount.GetAddress().String()},
	})
	assert.Equal(t, "100000000", qRsp["free"], qRsp)

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	myValidatorAddr := sdk.ValAddress(providerChain.Vals.Validators[0].Address).String()
	stakeMsg := fmt.Sprintf(`{"validator": "%s"}`, myValidatorAddr)
	_, err = providerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   providerChain.SenderAccount.GetAddress().String(),
		Contract: providerContracts.vault.String(),
		Msg: []byte(fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
			sdk.DefaultBondDenom, 50_000_000,
			base64.StdEncoding.EncodeToString([]byte(stakeMsg)))),
	})
	require.NoError(t, err)

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	_, err = providerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   providerChain.SenderAccount.GetAddress().String(),
		Contract: providerContracts.vault.String(),
		Msg: []byte(fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
			providerContracts.externalStaking.String(),
			sdk.DefaultBondDenom, 40_000_000,
			base64.StdEncoding.EncodeToString([]byte(stakeMsg)))),
	})
	require.NoError(t, err)

	// then
	qRsp = queryProvContract(providerContracts.externalStaking.String(), Query{
		"stake": {
			"user":      providerChain.SenderAccount.GetAddress().String(),
			"validator": myValidatorAddr,
		},
	})
	assert.Equal(t, "40000000", qRsp["stake"], qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])

	// consumer chain tests
	// ====================
	// until fully ibc enabled

	// when staking contract is instructed to bond tokens
	doExecStaking := func(payload string) {
		_, err = consumerChain.SendMsgs(&wasmtypes.MsgExecuteContract{
			Sender:   consumerChain.SenderAccount.GetAddress().String(),
			Contract: consumerContracts.staking.String(),
			Msg:      []byte(payload),
		})
		require.NoError(t, err)
	}
	doExecStaking(fmt.Sprintf(`{"bond":{"validator":"%s", "amount":{"denom":"%s", "amount":"10000000"}}}`, valAddr.String(), sdk.DefaultBondDenom))

	// then delegated amount is not updated before the epoch

	assertTotalDelegated := func(expTotalDelegated math.Int) {
		usedAmount := consumerApp.MeshSecKeeper.GetTotalDelegated(consumerChain.GetContext(), consumerContracts.staking)
		assert.Equal(t, sdk.NewCoin(sdk.DefaultBondDenom, expTotalDelegated), usedAmount)
	}
	assertTotalDelegated(math.ZeroInt())

	// when an epoch ends, the delegation rebalance is triggered
	doRebalance := func() {
		rebalanceMsg := []byte(`{"rebalance":{}}`)
		_, err = consumerApp.WasmKeeper.Sudo(consumerChain.GetContext(), consumerContracts.staking, rebalanceMsg)
		require.NoError(t, err)
	}
	doRebalance()

	// then the total delegated amount is updated
	assertTotalDelegated(math.NewInt(10_000_000))

	// and the delegated amount is updated for the validator
	assertShare := func(exp int64) {
		del, found := consumerApp.StakingKeeper.GetDelegation(consumerChain.GetContext(), consumerContracts.staking, valAddr)
		require.True(t, found)
		assert.Equal(t, math.LegacyNewDec(exp), del.Shares)
	}
	assertShare(10)

	// when undelegated
	doExecStaking(fmt.Sprintf(`{"unbond":{"validator":"%s", "amount":{"denom":"%s", "amount":"1000000"}}}`, valAddr.String(), sdk.DefaultBondDenom))
	// when an epoch ends, the delegation rebalance is triggered
	doRebalance()
	// then undelegated and burned
	assertTotalDelegated(math.NewInt(9_000_000))
	assertShare(9)
}

type Query map[string]map[string]any

type ProviderContracts struct {
	vault           sdk.AccAddress
	externalStaking sdk.AccAddress
}

func bootstrapProviderContracts(t *testing.T, chain *wasmibctesting.TestChain) ProviderContracts {
	vaultCodeID := chain.StoreCodeFile("testdata/mesh_vault.wasm.gz").CodeID
	proxyCodeID := chain.StoreCodeFile("testdata/mesh_native_staking_proxy.wasm.gz").CodeID
	nativeStakingCodeID := chain.StoreCodeFile("testdata/mesh_native_staking.wasm.gz").CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d}`, sdk.DefaultBondDenom, proxyCodeID))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, sdk.DefaultBondDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	vaultContract := InstantiateContract(t, chain, vaultCodeID, initMsg)

	// external staking
	unbondingPeriod := 21 * 24 * 60 * 60 // 21 days - make configurable?
	extStakingCodeID := chain.StoreCodeFile("testdata/external_staking.wasm.gz").CodeID
	rewardToken := "todo" // ics20 token
	initMsg = []byte(fmt.Sprintf(`{"denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q}`, sdk.DefaultBondDenom, vaultContract.String(), unbondingPeriod, rewardToken))
	externalStakingContract := InstantiateContract(t, chain, extStakingCodeID, initMsg)

	return ProviderContracts{
		vault:           vaultContract,
		externalStaking: externalStakingContract,
	}
}

type ConsumerContract struct {
	staking   sdk.AccAddress
	priceFeed sdk.AccAddress
	converter sdk.AccAddress
}

func bootstrapConsumerContracts(t *testing.T, consumerChain *wasmibctesting.TestChain) ConsumerContract {
	codeID := consumerChain.StoreCodeFile("testdata/mesh_simple_price_feed.wasm.gz").CodeID
	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContract := InstantiateContract(t, consumerChain, codeID, initMsg)
	// instantiate virtual staking contract
	codeID = consumerChain.StoreCodeFile("testdata/mesh_virtual_staking.wasm.gz").CodeID
	initMsg = []byte(fmt.Sprintf(`{"denom": %q}`, sdk.DefaultBondDenom))
	stakingContract := InstantiateContract(t, consumerChain, codeID, initMsg)
	// instantiate converter
	codeID = consumerChain.StoreCodeFile("testdata/mesh_converter.wasm.gz").CodeID
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "virtual_staking": %q, "discount": "0.1"}`, priceFeedContract.String(), stakingContract.String())) // todo: configure price
	converterContract := InstantiateContract(t, consumerChain, codeID, initMsg)

	return ConsumerContract{
		staking:   stakingContract,
		priceFeed: priceFeedContract,
		converter: converterContract,
	}
}
