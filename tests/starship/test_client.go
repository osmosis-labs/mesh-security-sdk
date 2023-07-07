package e2e

import (
	"context"
	"encoding/base64"
	_ "encoding/base64"
	"fmt"
	"strconv"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	_ "github.com/cosmology-tech/starship/clients/go/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type Query map[string]map[string]any

func Querier(t *testing.T, chain *Client) func(contract string, query Query) map[string]any {
	return func(contract string, query Query) map[string]any {
		qRsp := make(map[string]any)
		err := SmartQuery(chain, contract, query, &qRsp)
		require.NoError(t, err)
		return qRsp
	}
}

type TestProviderClient struct {
	t         *testing.T
	chain     *Client
	contracts ProviderContracts
}

func NewProviderClient(t *testing.T, chain *Client) *TestProviderClient {
	return &TestProviderClient{t: t, chain: chain}
}

type ProviderContracts struct {
	vault                 sdk.AccAddress
	nativeStakingContract sdk.AccAddress
	externalStaking       sdk.AccAddress
}

func (p *TestProviderClient) BootstrapContracts(connId, portID string) ProviderContracts {
	t := p.t
	vaultCodeResp, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_vault.wasm"))
	require.NoError(t, err)
	vaultCodeID := vaultCodeResp.CodeID
	proxyCodeResp, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_native_staking_proxy.wasm"))
	require.NoError(t, err)
	proxyCodeID := proxyCodeResp.CodeID
	nativeStakingCodeResp, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_native_staking.wasm"))
	nativeStakingCodeID := nativeStakingCodeResp.CodeID

	// test non empty codeids
	require.NotEmpty(t, vaultCodeID)
	require.NotEmpty(t, proxyCodeID)
	require.NotEmpty(t, nativeStakingCodeID)

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d}`, sdk.DefaultBondDenom, proxyCodeID))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, sdk.DefaultBondDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	contracts := InstantiateContract(t, p.chain, vaultCodeID, initMsg)

	fmt.Printf("contracts: %v\n", contracts)

	vaultContract := contracts[0]
	nativeStakingContract := contracts[1]

	// external staking
	unbondingPeriod := 21 * 24 * 60 * 60 // 21 days - make configurable?
	extStaking, err := StoreCodeFile(p.chain, buildPathToWasm("external_staking.wasm"))
	require.NoError(t, err)
	extStakingCodeID := extStaking.CodeID
	rewardToken := "ustake" // ics20 token
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q}`,
		connId, portID, sdk.DefaultBondDenom, vaultContract.String(), unbondingPeriod, rewardToken))
	externalStakingContract := InstantiateContract(t, p.chain, extStakingCodeID, initMsg)[0]

	r := ProviderContracts{
		vault:                 vaultContract,
		externalStaking:       externalStakingContract,
		nativeStakingContract: nativeStakingContract,
	}
	p.contracts = r
	return r
}

func (p TestProviderClient) MustExecVault(payload string, funds ...sdk.Coin) *sdk.TxResponse {
	return p.mustExec(p.contracts.vault, payload, funds)
}

func (p TestProviderClient) MustExecExtStaking(payload string, funds ...sdk.Coin) *sdk.TxResponse {
	return p.mustExec(p.contracts.externalStaking, payload, funds)
}

func (p TestProviderClient) mustExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) *sdk.TxResponse {
	rsp, err := p.chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.chain.Address,
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	}, "")
	require.NoError(p.t, err)
	return rsp
}

func (p TestProviderClient) MustFailExecVault(payload string, funds ...sdk.Coin) error {
	return p.mustFailExec(p.contracts.vault, payload, funds)
}

// This will execute the contract and assert that it fails, returning the error if desired.
// (Unlike most functions, it will panic on failure, returning error is success case)
func (p TestProviderClient) mustFailExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) error {
	resp, err := p.chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.chain.Address,
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	}, "")
	require.Error(p.t, err, "Response: %v", resp)

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
		"account": {"account": p.chain.Address},
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
	chain     *Client
	contracts ConsumerContract
}

func NewConsumerClient(t *testing.T, chain *Client) *TestConsumerClient {
	return &TestConsumerClient{t: t, chain: chain}
}

type ConsumerContract struct {
	staking   sdk.AccAddress
	priceFeed sdk.AccAddress
	converter sdk.AccAddress
}

func (p *TestConsumerClient) BootstrapContracts() ConsumerContract {
	// what does this do????
	// modify end-blocker to fail fast in tests
	//msModule := p.app.ModuleManager.Modules[types.ModuleName].(*meshsecurity.AppModule)
	//msModule.SetAsyncTaskRspHandler(meshsecurity.PanicOnErrorExecutionResponseHandler())

	code, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_simple_price_feed.wasm"))
	require.NoError(p.t, err)
	codeID := code.CodeID
	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContract := InstantiateContract(p.t, p.chain, codeID, initMsg)[0]
	// virtual staking is setup by the consumer
	virtStakeCode, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_virtual_staking.wasm"))
	require.NoError(p.t, err)
	virtStakeCodeID := virtStakeCode.CodeID
	// instantiate converter
	code, err = StoreCodeFile(p.chain, buildPathToWasm("mesh_converter.wasm"))
	require.NoError(p.t, err)
	codeID = code.CodeID

	discount := "0.1" // todo: configure price
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract.String(), discount, sdk.DefaultBondDenom, virtStakeCodeID))
	// bug in lens that returns second contract instantiated
	contracts := InstantiateContract(p.t, p.chain, codeID, initMsg)
	converterContract := contracts[0]
	virtualStakingContract := contracts[1]

	r := ConsumerContract{
		staking:   virtualStakingContract,
		priceFeed: priceFeedContract,
		converter: converterContract,
	}
	fmt.Printf("r is: %v\n", r)
	p.contracts = r
	return r
}

//func (p *TestConsumerClient) ExecNewEpoch() {
//	epochLength := p.app.MeshSecKeeper.GetRebalanceEpochLength(p.chain.GetContext())
//	p.chain.Coordinator.CommitNBlocks(p.chain, epochLength)
//}

//// MustExecGovProposal submit and vote yes on proposal
//func (p *TestConsumerClient) MustExecGovProposal(msg *types.MsgSetVirtualStakingMaxCap) {
//	proposalID := submitGovProposal(p.t, p.chain, msg)
//	voteAndPassGovProposal(p.t, p.chain, proposalID)
//}

//func (p *TestConsumerClient) QueryMaxCap() types.QueryVirtualStakingMaxCapLimitResponse {
//	q := baseapp.QueryServiceTestHelper{GRPCQueryRouter: p.app.GRPCQueryRouter(), Ctx: p.chain.GetContext()}
//	var rsp types.QueryVirtualStakingMaxCapLimitResponse
//	err := q.Invoke(nil, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimit", &types.QueryVirtualStakingMaxCapLimitRequest{Address: p.contracts.staking.String()}, &rsp)
//	require.NoError(p.t, err)
//	return rsp
//}

//func (p *TestConsumerClient) assertTotalDelegated(expTotalDelegated math.Int) {
//	usedAmount := p.app.MeshSecKeeper.GetTotalDelegated(p.chain.GetContext(), p.contracts.staking)
//	assert.Equal(p.t, sdk.NewCoin(sdk.DefaultBondDenom, expTotalDelegated), usedAmount)
//}
//
//func (p *TestConsumerClient) assertShare(val sdk.ValAddress, exp math.LegacyDec) {
//	del, found := p.app.StakingKeeper.GetDelegation(p.chain.GetContext(), p.contracts.staking, val)
//	require.True(p.t, found)
//	assert.Equal(p.t, exp, del.Shares)
//}
