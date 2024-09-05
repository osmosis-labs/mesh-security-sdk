package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
	providertypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
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

func Querier(t *testing.T, chain *TestChain) func(contract string, query Query) QueryResponse {
	return func(contract string, query Query) QueryResponse {
		qRsp := make(map[string]any)
		err := chain.SmartQuery(contract, query, &qRsp)
		require.NoError(t, err)
		return qRsp
	}
}

type TestProviderClient struct {
	t         *testing.T
	chain     *TestChain
	Contracts ProviderContracts
}

type TestChain struct {
	*ibctesting.TestChain
	t *testing.T
}

func NewTestChain(t *testing.T, chain *ibctesting.TestChain) *TestChain {
	return &TestChain{
		t:         t,
		TestChain: chain,
	}
}

func (tc *TestChain) IBCTestChain() *ibctesting.TestChain {
	return tc.TestChain
}

func NewProviderClient(t *testing.T, chain *TestChain) *TestProviderClient {
	return &TestProviderClient{t: t, chain: chain}
}

func (tc *TestChain) SendMsgsWithSigner(privKey cryptotypes.PrivKey, signer *authtypes.BaseAccount, msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	tc.Coordinator.UpdateTimeForChain(tc.TestChain)
	_, r, gotErr := app.SignAndDeliverWithoutCommit(
		tc.t,
		tc.TxConfig,
		tc.App.GetBaseApp(),
		msgs,
		tc.DefaultMsgFees,
		tc.ChainID,
		[]uint64{signer.GetAccountNumber()},
		[]uint64{signer.GetSequence()},
		privKey,
	)

	// NextBlock calls app.Commit()
	tc.NextBlock()

	// increment sequence for successful and failed transaction execution
	require.NoError(tc.t, signer.SetSequence(signer.GetSequence()+1))
	tc.Coordinator.IncrementTime()

	if gotErr != nil {
		return nil, gotErr
	}

	tc.CaptureIBCEvents(r.Events)

	return r, nil
}

type ProviderContracts struct {
	Vault           sdk.AccAddress
	NativeStaking   sdk.AccAddress
	ExternalStaking sdk.AccAddress
}

func (p *TestProviderClient) BootstrapContracts(provApp *app.MeshApp, connId, portID string) ProviderContracts {
	var (
		unbondingPeriod           = 21 * 24 * 60 * 60 // 21 days - make configurable?
		localSlashRatioDoubleSign = "0.20"
		localSlashRatioOffline    = "0.10"
		extSlashRatioDoubleSign   = "0.20"
		extSlashRatioOffline      = "0.10"
		rewardTokenDenom          = sdk.DefaultBondDenom
		localTokenDenom           = sdk.DefaultBondDenom
	)
	vaultCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_vault.wasm")).CodeID
	proxyCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_native_staking_proxy.wasm")).CodeID
	nativeStakingCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_native_staking.wasm")).CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d, "slash_ratio_dsign": %q, "slash_ratio_offline": %q }`, localTokenDenom, proxyCodeID, localSlashRatioDoubleSign, localSlashRatioOffline))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, localTokenDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	vaultContract := InstantiateContract(p.t, p.chain, vaultCodeID, initMsg)
	ctx := p.chain.GetContext()
	params := provApp.MeshSecProvKeeper.GetParams(ctx)
	params.VaultAddress = vaultContract.String()
	provApp.MeshSecProvKeeper.SetParams(ctx, params)

	// external staking
	extStakingCodeID := p.chain.StoreCodeFile(buildPathToWasm("mesh_external_staking.wasm")).CodeID
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q, "slash_ratio": { "double_sign": %q, "offline": %q }  }`,
		connId, portID, localTokenDenom, vaultContract.String(), unbondingPeriod, rewardTokenDenom, extSlashRatioDoubleSign, extSlashRatioOffline))
	externalStakingContract := InstantiateContract(p.t, p.chain, extStakingCodeID, initMsg)

	r := ProviderContracts{
		Vault:           vaultContract,
		ExternalStaking: externalStakingContract,
	}
	p.Contracts = r

	// local staking
	vaultConfig := p.QueryVault(Query{
		"config": {},
	})
	require.Contains(p.t, vaultConfig, "local_staking")
	nativeStaking, err := sdk.AccAddressFromBech32(vaultConfig["local_staking"].(string))
	require.NoError(p.t, err)
	r.NativeStaking = nativeStaking
	p.Contracts = r

	p.MustExecParamsChangeProposal(provApp, vaultContract.String(), nativeStaking.String())

	return r
}

func (p TestProviderClient) MustCreatePermanentLockedAccount(acc string, coins ...sdk.Coin) *sdk.Result {
	rsp, err := p.chain.SendMsgs(&vestingtypes.MsgCreatePermanentLockedAccount{
		FromAddress: p.chain.SenderAccount.GetAddress().String(),
		ToAddress:   acc,
		Amount:      coins,
	})
	require.NoError(p.t, err)
	return rsp
}

func (p TestProviderClient) BankSendWithSigner(privKey cryptotypes.PrivKey, signer *authtypes.BaseAccount, to string, coins ...sdk.Coin) error {
	_, err := p.chain.SendMsgsWithSigner(
		privKey,
		signer,
		&banktypes.MsgSend{
			FromAddress: signer.GetAddress().String(),
			ToAddress:   to,
			Amount:      coins,
		},
	)
	return err
}

func (p TestProviderClient) MustExecVault(payload string, funds ...sdk.Coin) *sdk.Result {
	return p.mustExec(p.Contracts.Vault, payload, funds)
}

func (p TestProviderClient) MustExecVaultWithSigner(privKey cryptotypes.PrivKey, signer *authtypes.BaseAccount, payload string, funds ...sdk.Coin) *sdk.Result {
	rsp, err := p.ExecWithSigner(privKey, signer, p.Contracts.Vault, payload, funds...)
	require.NoError(p.t, err)
	return rsp
}

func (p TestProviderClient) ExecVaultWithSigner(privKey cryptotypes.PrivKey, signer *authtypes.BaseAccount, payload string, funds ...sdk.Coin) error {
	_, err := p.ExecWithSigner(privKey, signer, p.Contracts.Vault, payload, funds...)
	return err
}

func (p TestProviderClient) MustExecExtStaking(payload string, funds ...sdk.Coin) *sdk.Result {
	return p.mustExec(p.Contracts.ExternalStaking, payload, funds)
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

func (p TestProviderClient) ExecWithSigner(privKey cryptotypes.PrivKey, signer *authtypes.BaseAccount, contract sdk.AccAddress, payload string, funds ...sdk.Coin) (*sdk.Result, error) {
	rsp, err := p.chain.SendMsgsWithSigner(
		privKey,
		signer,
		&wasmtypes.MsgExecuteContract{
			Sender:   signer.GetAddress().String(),
			Contract: contract.String(),
			Msg:      []byte(payload),
			Funds:    funds,
		},
	)
	return rsp, err
}

// MustExecGovProposal submit and vote yes on proposal
func (p TestProviderClient) MustExecParamsChangeProposal(provApp *app.MeshApp, vault, nativeStaking string) {
	msg := &providertypes.MsgUpdateParams{
		Authority: provApp.MeshSecKeeper.GetAuthority(),
		Params: providertypes.Params{
			VaultAddress:         vault,
			NativeStakingAddress: nativeStaking,
		},
	}
	proposalID := submitGovProposal(p.t, p.chain, msg)
	voteAndPassGovProposal(p.t, p.chain, proposalID)
}

func (p TestProviderClient) MustFailExecVault(payload string, funds ...sdk.Coin) error {
	rsp, err := p.Exec(p.Contracts.Vault, payload, funds...)
	require.Error(p.t, err, "Response: %v", rsp)
	return err
}

func (p TestProviderClient) MustExecStakeRemote(val string, amt sdk.Coin) {
	require.NoError(p.t, p.ExecStakeRemote(val, amt))
}

func (p TestProviderClient) ExecStakeRemote(val string, amt sdk.Coin) error {
	payload := fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%s"}, "msg":%q}}`,
		p.Contracts.ExternalStaking.String(),
		amt.Denom, amt.Amount.String(),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, val))))
	_, err := p.Exec(p.Contracts.Vault, payload)
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
	return ParseHighLow(p.t, qRsp["stake"]).Low
}

func (p TestProviderClient) QueryNativeStakingProxyByOwner(user string) sdk.AccAddress {
	qRsp := p.QueryNativeStaking(Query{
		"proxy_by_owner": {
			"owner": user,
		},
	})
	require.Contains(p.t, qRsp, "proxy")
	a, err := sdk.AccAddressFromBech32(qRsp["proxy"].(string))
	require.NoError(p.t, err)

	return a
}
func (p TestProviderClient) QueryExtStaking(q Query) QueryResponse {
	return Querier(p.t, p.chain)(p.Contracts.ExternalStaking.String(), q)
}

func (p TestProviderClient) QueryVault(q Query) QueryResponse {
	return Querier(p.t, p.chain)(p.Contracts.Vault.String(), q)
}

func (p TestProviderClient) QueryNativeStaking(q Query) QueryResponse {
	return Querier(p.t, p.chain)(p.Contracts.NativeStaking.String(), q)
}

type HighLowType struct {
	High, Low int
}

// ParseHighLow convert json source type into custom type
func ParseHighLow(t *testing.T, a any) HighLowType {
	m, ok := a.(map[string]any)
	require.True(t, ok, "%T", a)
	require.Contains(t, m, "h")
	require.Contains(t, m, "l")
	h, err := strconv.Atoi(m["h"].(string))
	require.NoError(t, err)
	l, err := strconv.Atoi(m["l"].(string))
	require.NoError(t, err)
	return HighLowType{High: h, Low: l}
}

func (p TestProviderClient) QueryVaultFreeBalance() int {
	qRsp := p.QueryVault(Query{
		"account": {"account": p.chain.SenderAccount.GetAddress().String()},
	})
	require.NotEmpty(p.t, qRsp["free"], qRsp)
	return ParseHighLow(p.t, qRsp["free"]).Low
}

func (p TestProviderClient) QueryVaultBalance() int {
	qRsp := p.QueryVault(Query{
		"account_details": {"account": p.chain.SenderAccount.GetAddress().String()},
	})
	require.NotEmpty(p.t, qRsp["bonded"], qRsp)
	b, err := strconv.Atoi(qRsp["bonded"].(string))
	require.NoError(p.t, err)
	return b
}

func (p TestProviderClient) QuerySpecificAddressVaultBalance(address string) int {
	qRsp := p.QueryVault(Query{
		"account_details": {"account": address},
	})
	require.NotEmpty(p.t, qRsp["bonded"], qRsp)
	b, err := strconv.Atoi(qRsp["bonded"].(string))
	require.NoError(p.t, err)
	return b
}

func (p TestProviderClient) QueryMaxLien() int {
	qRsp := p.QueryVault(Query{
		"account_details": {"account": p.chain.SenderAccount.GetAddress().String()},
	})
	require.NotEmpty(p.t, qRsp["max_lien"], qRsp)
	return ParseHighLow(p.t, qRsp["max_lien"]).Low
}

func (p TestProviderClient) QuerySlashableAmount() int {
	qRsp := p.QueryVault(Query{
		"account_details": {"account": p.chain.SenderAccount.GetAddress().String()},
	})
	require.NotEmpty(p.t, qRsp["total_slashable"], qRsp)
	return ParseHighLow(p.t, qRsp["total_slashable"]).Low
}

type TestConsumerClient struct {
	t         *testing.T
	chain     *TestChain
	contracts ConsumerContract
	app       *app.MeshApp
}

func NewConsumerClient(t *testing.T, chain *TestChain) *TestConsumerClient {
	return &TestConsumerClient{t: t, chain: chain, app: chain.App.(*app.MeshApp)}
}

type ConsumerContract struct {
	staking   sdk.AccAddress
	priceFeed sdk.AccAddress
	converter sdk.AccAddress
}

func (p *TestConsumerClient) BootstrapContracts(x example) ConsumerContract {
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
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d, "max_retrieve": %d, "tombstoned_unbond_enable": true}`,
		priceFeedContract.String(), discount, remoteDenom, virtStakeCodeID, x.MaxRetrieve))
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
	execHeight, ok := p.app.MeshSecKeeper.GetNextScheduledTaskHeight(p.chain.GetContext(), types.SchedulerTaskHandleEpoch, p.contracts.staking)
	require.True(p.t, ok)
	if ch := uint64(p.chain.GetContext().BlockHeight()); ch < execHeight {
		p.chain.Coordinator.CommitNBlocks(p.chain.IBCTestChain(), execHeight-ch)
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

func (p *TestConsumerClient) ExecSetMaxCap(cap sdk.Coin) {
	msgServer := keeper.NewMsgServer(p.app.MeshSecKeeper)
	msgServer.SetVirtualStakingMaxCap(p.chain.GetContext(), &types.MsgSetVirtualStakingMaxCap{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Contract:  p.contracts.staking.String(),
		MaxCap:    cap,
	})
}

// MustEnableVirtualStaking add authority to mint/burn virtual tokens gov proposal
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
	err := q.Invoke(context.TODO(), "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimit", &types.QueryVirtualStakingMaxCapLimitRequest{Address: p.contracts.staking.String()}, &rsp)
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
