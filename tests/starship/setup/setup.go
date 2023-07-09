package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	_ "github.com/cosmology-tech/starship/clients/go/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type Query map[string]map[string]any

func Querier(chain *Client) func(contract string, query Query) map[string]any {
	return func(contract string, query Query) map[string]any {
		qRsp := make(map[string]any)
		err := Eventually(
			func() bool {
				qRsp = make(map[string]any)
				err := SmartQuery(chain, contract, query, &qRsp)
				if err != nil {
					if strings.Contains(err.Error(), "Value is already write locked") {
						fmt.Printf("smart query error (ignored): %s\n", err)
						return false
					}
					panic(fmt.Sprintf("error in query: %s\n. Stopping early...", err))
				}
				_, ok := qRsp["locked"]
				if ok {
					return false
				}
				return true
			},
			180*time.Second,
			time.Second,
			"locked contract response for too long: %v",
			qRsp,
		)
		if err != nil {
			panic(err)
		}
		return qRsp
	}
}

type ProviderClient struct {
	chain     *Client
	contracts *ProviderContracts
}

func NewProviderClient(chain *Client) *ProviderClient {
	return &ProviderClient{chain: chain}
}

type ProviderContracts struct {
	vault                 sdk.AccAddress
	nativeStakingContract sdk.AccAddress
	externalStaking       sdk.AccAddress
}

func (p *ProviderClient) BootstrapContracts(connId, portID string) (*ProviderContracts, error) {
	vaultCodeResp, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_vault.wasm"))
	if err != nil {
		return nil, err
	}
	vaultCodeID := vaultCodeResp.CodeID
	proxyCodeResp, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_native_staking_proxy.wasm"))
	if err != nil {
		return nil, err
	}
	proxyCodeID := proxyCodeResp.CodeID
	nativeStakingCodeResp, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_native_staking.wasm"))
	nativeStakingCodeID := nativeStakingCodeResp.CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d}`, p.chain.Denom, proxyCodeID))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, p.chain.Denom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	contracts, err := InstantiateContract(p.chain, vaultCodeID, initMsg)
	if err != nil {
		return nil, err
	}

	vaultContract := contracts[0]
	nativeStakingContract := contracts[1]

	// external staking
	unbondingPeriod := 100 // 100s - make configurable?
	extStaking, err := StoreCodeFile(p.chain, buildPathToWasm("external_staking.wasm"))
	if err != nil {
		return nil, err
	}
	extStakingCodeID := extStaking.CodeID
	rewardToken := "ujuno" // ics20 token
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q}`,
		connId, portID, p.chain.Denom, vaultContract.String(), unbondingPeriod, rewardToken))
	externalStakingContracts, err := InstantiateContract(p.chain, extStakingCodeID, initMsg)
	if err != nil {
		return nil, err
	}
	externalStakingContract := externalStakingContracts[0]

	r := &ProviderContracts{
		vault:                 vaultContract,
		externalStaking:       externalStakingContract,
		nativeStakingContract: nativeStakingContract,
	}
	fmt.Printf("Provider Contracts:\n  valut: %s\n  externalStaking: %s\n  nativeStaking: %s\n", r.vault.String(), r.externalStaking.String(), r.nativeStakingContract.String())
	p.contracts = r
	return r, nil
}

func (p ProviderClient) MustExecVault(payload string, funds ...sdk.Coin) (*sdk.TxResponse, error) {
	return p.mustExec(p.contracts.vault, payload, funds)
}

func (p ProviderClient) MustExecExtStaking(payload string, funds ...sdk.Coin) (*sdk.TxResponse, error) {
	return p.mustExec(p.contracts.externalStaking, payload, funds)
}

func (p ProviderClient) mustExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) (*sdk.TxResponse, error) {
	rsp, err := p.chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.chain.Address,
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	}, "")
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (p ProviderClient) MustFailExecVault(payload string, funds ...sdk.Coin) error {
	return p.mustFailExec(p.contracts.vault, payload, funds)
}

// This will execute the contract and assert that it fails, returning the error if desired.
// (Unlike most functions, it will panic on failure, returning error is success case)
func (p ProviderClient) mustFailExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) error {
	resp, err := p.chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.chain.Address,
		Contract: contract.String(),
		Msg:      []byte(payload),
		Funds:    funds,
	}, "")
	if err != nil {
		fmt.Printf("error wasm exec: error: %s, response: %v\n", err, resp)
	}

	return err
}

func (p ProviderClient) QueryExtStaking(q Query) map[string]any {
	return Querier(p.chain)(p.contracts.externalStaking.String(), q)
}

func (p ProviderClient) QueryVault(q Query) map[string]any {
	return Querier(p.chain)(p.contracts.vault.String(), q)
}

func (p ProviderClient) QueryVaultFreeBalance() int {
	qRsp := map[string]any{}
	err := Eventually(
		func() bool {
			qRsp = p.QueryVault(Query{
				"account": {"account": p.chain.Address},
			})
			_, ok := qRsp["locked"]
			if ok {
				return false
			}
			return true
		},
		60*time.Second,
		time.Second,
		"valut token locked for too long: %v",
		qRsp,
	)
	if err != nil {
		panic(err)
	}
	acct := qRsp["account"].(map[string]any)
	r, err := strconv.Atoi(acct["free"].(string))
	if err != nil {
		panic(err)
	}
	return r
}

type ConsumerClient struct {
	chain     *Client
	contracts *ConsumerContract
}

func NewConsumerClient(chain *Client) *ConsumerClient {
	return &ConsumerClient{chain: chain}
}

type ConsumerContract struct {
	staking   sdk.AccAddress
	priceFeed sdk.AccAddress
	converter sdk.AccAddress
}

func (p *ConsumerClient) BootstrapContracts() (*ConsumerContract, error) {
	// what does this do????
	// modify end-blocker to fail fast in tests
	//msModule := p.app.ModuleManager.Modules[types.ModuleName].(*meshsecurity.AppModule)
	//msModule.SetAsyncTaskRspHandler(meshsecurity.PanicOnErrorExecutionResponseHandler())

	code, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_simple_price_feed.wasm"))
	if err != nil {
		return nil, err
	}
	codeID := code.CodeID

	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContracts, err := InstantiateContract(p.chain, codeID, initMsg)
	if err != nil {
		return nil, err
	}
	priceFeedContract := priceFeedContracts[0]

	// virtual staking is setup by the consumer
	virtStakeCode, err := StoreCodeFile(p.chain, buildPathToWasm("mesh_virtual_staking.wasm"))
	if err != nil {
		return nil, err
	}
	virtStakeCodeID := virtStakeCode.CodeID

	// instantiate converter
	code, err = StoreCodeFile(p.chain, buildPathToWasm("mesh_converter.wasm"))
	if err != nil {
		return nil, err
	}
	codeID = code.CodeID

	discount := "0.1"      // todo: configure price
	remoteToken := "uosmo" // todo: figure out if this is correct
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract.String(), discount, remoteToken, virtStakeCodeID))
	// bug in lens that returns second contract instantiated
	contracts, err := InstantiateContract(p.chain, codeID, initMsg)
	if err != nil {
		return nil, err
	}
	converterContract, virtualStakingContract := contracts[0], contracts[1]

	r := &ConsumerContract{
		staking:   virtualStakingContract,
		priceFeed: priceFeedContract,
		converter: converterContract,
	}
	fmt.Printf("Consumer Contracts:\n  staking: %s\n  priceFeed: %s\n  converter: %s\n", r.staking.String(), r.priceFeed.String(), r.converter.String())
	p.contracts = r
	return r, nil
}

func (p *ConsumerClient) ExecNewEpoch() {
	// wait for epoch length
	time.Sleep(150 * time.Second)
}

// MustExecGovProposal submit and vote yes on proposal
func (p *ConsumerClient) MustExecGovProposal(msg *types.MsgSetVirtualStakingMaxCap) {
	proposalID, err := submitGovProposal(p.chain, msg)
	if err != nil {
		panic(err)
	}
	err = voteAndPassGovProposal(p.chain, proposalID)
	if err != nil {
		panic(err)
	}
}

func (p *ConsumerClient) QueryMaxCap() types.QueryVirtualStakingMaxCapLimitResponse {
	q := &types.QueryVirtualStakingMaxCapLimitRequest{Address: p.contracts.staking.String()}
	rsp, err := types.NewQueryClient(p.chain.Client).VirtualStakingMaxCapLimit(context.Background(), q)
	if err != nil {
		panic(err)
	}
	return *rsp
}
