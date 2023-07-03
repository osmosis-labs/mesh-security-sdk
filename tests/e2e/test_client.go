package e2e

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
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

// Query is a query type used in tests only
type Query map[string]map[string]any

// QueryResponse is a response type used in tests only
type QueryResponse map[string]any

// To can be used to navigate through the map structure
func (q QueryResponse) To(path ...string) QueryResponse {
	r, ok := q[path[0]]
	if !ok {
		panic(fmt.Sprintf("key %q does not exist", path[0]))
	}
	var x QueryResponse = r.(map[string]any)
	if len(path) == 1 {
		return x
	}
	return x.To(path[1:]...)
}

func (q QueryResponse) Array(key string) []QueryResponse {
	val, ok := q[key]
	if !ok {
		panic(fmt.Sprintf("key %q does not exist", key))
	}
	sl := val.([]any)
	result := make([]QueryResponse, len(sl))
	for i, v := range sl {
		result[i] = v.(map[string]any)
	}
	return result
}

func Querier(t *testing.T, chain *ibctesting.TestChain) func(contract string, query Query) QueryResponse {
	return func(contract string, query Query) QueryResponse {
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
	var (
		unbondingPeriod  = 21 * 24 * 60 * 60 // 21 days - make configurable?
		maxLocalSlashing = "0.10"
		maxExtSlashing   = "0.05"
		rewardTokenDenom = sdk.DefaultBondDenom
		localTokenDenom  = sdk.DefaultBondDenom
	)
	vaultCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_vault.wasm")).CodeID
	proxyCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_native_staking_proxy.wasm")).CodeID
	nativeStakingCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_native_staking.wasm")).CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d, "max_slashing": %q }`, localTokenDenom, proxyCodeID, maxLocalSlashing))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, localTokenDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	vaultContract := InstantiateContract(p.t, p.chain, vaultCodeID, initMsg)

	// external staking
	extStakingCodeID := p.chain.StoreCodeFile(buildPathToWasm("external_staking.wasm")).CodeID
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q, "max_slashing": %q }`,
		connId, portID, localTokenDenom, vaultContract.String(), unbondingPeriod, rewardTokenDenom, maxExtSlashing))
	externalStakingContract := InstantiateContract(p.t, p.chain, extStakingCodeID, initMsg)

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
	rsp, err := p.Exec(contract, payload, funds...)
	require.NoError(p.t, err)
	return rsp
}

func (p TestProviderClient) Exec(contract sdk.AccAddress, payload string, funds ...sdk.Coin) (*sdk.Result, error) {
	rsp, err := p.chain.SendMsgs(&wasmtypes.MsgExecuteContract{
		Sender:   p.chain.SenderAccount.GetAddress().String(),
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	})
	return rsp, err
}

func (p TestProviderClient) MustFailExecVault(payload string, funds ...sdk.Coin) error {
	rsp, err := p.Exec(p.contracts.vault, payload, funds...)
	require.Error(p.t, err, "Response: %v", rsp)
	return err
}

func (p TestProviderClient) MustExecStakeRemote(val string, amt sdk.Coin) {
	require.NoError(p.t, p.ExecStakeRemote(val, amt))
}

func (p TestProviderClient) ExecStakeRemote(val string, amt sdk.Coin) error {
	payload := fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%s"}, "msg":%q}}`,
		p.contracts.externalStaking.String(),
		amt.Denom, amt.Amount.String(),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, val))))
	_, err := p.Exec(p.contracts.vault, payload)
	return err
}

func (p TestProviderClient) QueryExtStakingAmount(user, validator string) int {
	qRsp := p.QueryExtStaking(Query{
		"stake": {
			"user":      user,
			"validator": validator,
		},
	})
	require.Contains(p.t, qRsp, "stake")
	r, err := strconv.Atoi(qRsp["stake"].(string))
	require.NoError(p.t, err)
	return r
}

func (p TestProviderClient) QueryExtStaking(q Query) QueryResponse {
	return Querier(p.t, p.chain)(p.contracts.externalStaking.String(), q)
}

func (p TestProviderClient) QueryVault(q Query) QueryResponse {
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

	var ( // todo: configure
		tokenRatio  = "0.5"
		discount    = "0.1"
		remoteDenom = sdk.DefaultBondDenom
	)
	codeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_simple_price_feed.wasm")).CodeID
	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, tokenRatio))
	priceFeedContract := InstantiateContract(p.t, p.chain, codeID, initMsg)
	// virtual staking is setup by the consumer
	virtStakeCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_virtual_staking.wasm")).CodeID
	// instantiate converter
	codeID = p.chain.StoreCodeFile(buildPathToWasm("mesh_converter.wasm")).CodeID
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract.String(), discount, remoteDenom, virtStakeCodeID))
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
	execHeight, ok := p.app.MeshSecKeeper.GetNextScheduledTaskHeight(p.chain.GetContext(), types.SchedulerTaskRebalance, p.contracts.staking)
	require.True(p.t, ok)
	if ch := uint64(p.chain.GetContext().BlockHeight()); ch < execHeight {
		p.chain.Coordinator.CommitNBlocks(p.chain, execHeight-ch)
	}
	rsp := p.chain.NextBlock()
	// ensure capture events do not contain a contract error
	for _, e := range rsp.Events {
		if !strings.HasPrefix(e.Type, "wasm") {
			continue
		}
		for _, a := range e.Attributes {
			if strings.HasSuffix(a.String(), "error") {
				p.t.Fatalf("received error event: %s in %#v", a.Value, rsp.Events)
			}
		}
	}
}

// add authority to mint/burn virtual tokens gov proposal
func (p *TestConsumerClient) MustEnableVirtualStaking(maxCap sdk.Coin) {
	govProposal := &types.MsgSetVirtualStakingMaxCap{
		Authority: p.app.MeshSecKeeper.GetAuthority(),
		Contract:  p.contracts.staking.String(),
		MaxCap:    maxCap,
	}
	p.MustExecGovProposal(govProposal)
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
