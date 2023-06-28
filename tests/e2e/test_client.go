package e2e

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	"cosmossdk.io/math"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type Query map[string]map[string]any

func Querier(t *testing.T, chain *ibctesting.TestChain) func(contract string, query Query) map[string]any {
	return func(contract string, query Query) map[string]any {
		qRsp := make(map[string]any)
		err := chain.SmartQuery(contract, query, &qRsp)
		require.NoError(t, err)
		return qRsp
	}
}

type TestProviderClient struct {
	t         *testing.T
	chain     *ibctesting.TestChain
	contracts ProviderContracts
}

func NewProviderClient(t *testing.T, chain *ibctesting.TestChain) *TestProviderClient {
	return &TestProviderClient{t: t, chain: chain}
}

type ProviderContracts struct {
	vault           sdk.AccAddress
	externalStaking sdk.AccAddress
}

func (p *TestProviderClient) BootstrapContracts(connId, portID string) ProviderContracts {
	t := p.t
	chain := p.chain
	vaultCodeID := chain.StoreCodeFile(buildPathToWasm("mesh_vault.wasm")).CodeID
	proxyCodeID := chain.StoreCodeFile(buildPathToWasm("mesh_native_staking_proxy.wasm")).CodeID
	nativeStakingCodeID := chain.StoreCodeFile(buildPathToWasm("mesh_native_staking.wasm")).CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d}`, sdk.DefaultBondDenom, proxyCodeID))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, sdk.DefaultBondDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	vaultContract := InstantiateContract(t, chain, vaultCodeID, initMsg)

	// external staking
	unbondingPeriod := 21 * 24 * 60 * 60 // 21 days - make configurable?
	extStakingCodeID := chain.StoreCodeFile(buildPathToWasm("external_staking.wasm")).CodeID
	rewardToken := "todo" // ics20 token
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q}`,
		connId, portID, sdk.DefaultBondDenom, vaultContract.String(), unbondingPeriod, rewardToken))
	externalStakingContract := InstantiateContract(t, chain, extStakingCodeID, initMsg)

	r := ProviderContracts{
		vault:           vaultContract,
		externalStaking: externalStakingContract,
	}
	p.contracts = r
	return r
}

func (p TestProviderClient) MustExecVault(payload string, funds ...sdk.Coin) *sdk.Result {
	return p.mustExec(p.contracts.vault, payload, funds)
}

func (p TestProviderClient) MustExecExtStaking(payload string, funds ...sdk.Coin) *sdk.Result {
	return p.mustExec(p.contracts.externalStaking, payload, funds)
}

func (p TestProviderClient) mustExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) *sdk.Result {
	rsp, err := p.chain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   p.chain.SenderAccount.GetAddress().String(),
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	})
	require.NoError(p.t, err)
	return rsp
}

func (p TestProviderClient) MustFailExecVault(payload string, funds ...sdk.Coin) error {
	return p.mustFailExec(p.contracts.vault, payload, funds)
}

// This will execute the contract and assert that it fails, returning the error if desired.
// (Unlike most functions, it will panic on failure, returning error is success case)
func (p TestProviderClient) mustFailExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) error {
	resp, err := p.chain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   p.chain.SenderAccount.GetAddress().String(),
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	})
	require.Error(p.t, err, "Response: %v", resp)
	// workaround sequence issue until wasmd testchain code is updated
	p.chain.NextBlock()
	require.NoError(p.t, p.chain.SenderAccount.SetSequence(p.chain.SenderAccount.GetSequence()+1))
	return err
}

func (p TestProviderClient) QueryExtStaking(q Query) map[string]any {
	return Querier(p.t, p.chain)(p.contracts.externalStaking.String(), q)
}

func (p TestProviderClient) QueryVault(q Query) map[string]any {
	return Querier(p.t, p.chain)(p.contracts.vault.String(), q)
}

func (p TestProviderClient) QueryVaultFreeBalance() int {
	qRsp := p.QueryVault(Query{
		"account": {"account": p.chain.SenderAccount.GetAddress().String()},
	})
	require.NotEmpty(p.t, qRsp["account"], qRsp)
	acct := qRsp["account"].(map[string]any)
	require.NotEmpty(p.t, acct["free"], qRsp)
	r, err := strconv.Atoi(acct["free"].(string))
	require.NoError(p.t, err, qRsp)
	return r
}

type TestConsumerClient struct {
	t         *testing.T
	chain     *ibctesting.TestChain
	contracts ConsumerContract
	app       *app.MeshApp
}

func NewConsumerClient(t *testing.T, chain *ibctesting.TestChain) *TestConsumerClient {
	return &TestConsumerClient{t: t, chain: chain, app: chain.App.(*app.MeshApp)}
}

type ConsumerContract struct {
	staking   sdk.AccAddress
	priceFeed sdk.AccAddress
	converter sdk.AccAddress
}

func (p *TestConsumerClient) BootstrapContracts() ConsumerContract {
	// modify end-blocker to fail fast in tests
	msModule := p.app.ModuleManager.Modules[types.ModuleName].(*meshsecurity.AppModule)
	msModule.SetAsyncTaskRspHandler(meshsecurity.PanicOnErrorExecutionResponseHandler())

	codeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_simple_price_feed.wasm")).CodeID
	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContract := InstantiateContract(p.t, p.chain, codeID, initMsg)
	// virtual staking is setup by the consumer
	virtStakeCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_virtual_staking.wasm")).CodeID
	// instantiate converter
	codeID = p.chain.StoreCodeFile(buildPathToWasm("mesh_converter.wasm")).CodeID
	discount := "0.1" // todo: configure price
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract.String(), discount, sdk.DefaultBondDenom, virtStakeCodeID))
	converterContract := InstantiateContract(p.t, p.chain, codeID, initMsg)

	staking := Querier(p.t, p.chain)(converterContract.String(), Query{"config": {}})["virtual_staking"]
	r := ConsumerContract{
		staking:   sdk.MustAccAddressFromBech32(staking.(string)),
		priceFeed: priceFeedContract,
		converter: converterContract,
	}
	p.contracts = r
	return r
}

func (p *TestConsumerClient) ExecNewEpoch() {
	epochLength := p.app.MeshSecKeeper.GetRebalanceEpochLength(p.chain.GetContext())
	p.chain.Coordinator.CommitNBlocks(p.chain, epochLength)
}

// MustExecGovProposal submit and vote yes on proposal
func (p *TestConsumerClient) MustExecGovProposal(msg *types.MsgSetVirtualStakingMaxCap) {
	proposalID := submitGovProposal(p.t, p.chain, msg)
	voteAndPassGovProposal(p.t, p.chain, proposalID)
}

func (p *TestConsumerClient) QueryMaxCap() types.QueryVirtualStakingMaxCapLimitResponse {
	q := baseapp.QueryServiceTestHelper{GRPCQueryRouter: p.app.GRPCQueryRouter(), Ctx: p.chain.GetContext()}
	var rsp types.QueryVirtualStakingMaxCapLimitResponse
	err := q.Invoke(nil, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimit", &types.QueryVirtualStakingMaxCapLimitRequest{Address: p.contracts.staking.String()}, &rsp)
	require.NoError(p.t, err)
	return rsp
}

func (p *TestConsumerClient) assertTotalDelegated(expTotalDelegated math.Int) {
	usedAmount := p.app.MeshSecKeeper.GetTotalDelegated(p.chain.GetContext(), p.contracts.staking)
	assert.Equal(p.t, sdk.NewCoin(sdk.DefaultBondDenom, expTotalDelegated), usedAmount)
}

func (p *TestConsumerClient) assertShare(val sdk.ValAddress, exp math.LegacyDec) {
	del, found := p.app.StakingKeeper.GetDelegation(p.chain.GetContext(), p.contracts.staking, val)
	require.True(p.t, found)
	assert.Equal(p.t, exp, del.Shares)
}
