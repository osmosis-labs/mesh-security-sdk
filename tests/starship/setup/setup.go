package setup

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
			300*time.Second,
			5*time.Second,
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
	Chain               *Client
	Contracts           *ProviderContracts
	wasmContractPath    string
	wasmContractGZipped bool
}

func NewProviderClient(chain *Client, wasmContractPath string, wasmContractGZipped bool) *ProviderClient {
	return &ProviderClient{Chain: chain, wasmContractPath: wasmContractPath, wasmContractGZipped: wasmContractGZipped}
}

type ProviderContracts struct {
	Vault                 sdk.AccAddress
	NativeStakingContract sdk.AccAddress
	ExternalStaking       sdk.AccAddress
}

func (p *ProviderClient) BootstrapContracts(connId, portID, rewardDenom string) (*ProviderContracts, error) {
	var (
		unbondingPeriod  = 100 // 21 days - make configurable?
		maxLocalSlashing = "0.10"
		maxExtSlashing   = "0.05"
		localTokenDenom  = p.Chain.Denom
	)
	vaultCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_vault.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	vaultCodeID := vaultCodeResp.CodeID
	proxyCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_native_staking_proxy.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	proxyCodeID := proxyCodeResp.CodeID
	nativeStakingCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_native_staking.wasm", p.wasmContractGZipped))
	nativeStakingCodeID := nativeStakingCodeResp.CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d, "max_slashing": %q}`, localTokenDenom, proxyCodeID, maxLocalSlashing))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, localTokenDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	contracts, err := InstantiateContract(p.Chain, vaultCodeID, "provider-valut-contract", initMsg)
	if err != nil {
		return nil, err
	}

	vaultContract := contracts[vaultCodeID]
	nativeStakingContract := contracts[nativeStakingCodeID]

	// external Staking
	extStaking, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "external_staking.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	extStakingCodeID := extStaking.CodeID
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q, "max_slashing": %q }`,
		connId, portID, localTokenDenom, vaultContract.String(), unbondingPeriod, rewardDenom, maxExtSlashing))
	externalStakingContracts, err := InstantiateContract(p.Chain, extStakingCodeID, "provider-external-staking-contract", initMsg)
	if err != nil {
		return nil, err
	}
	externalStakingContract := externalStakingContracts[extStakingCodeID]

	r := &ProviderContracts{
		Vault:                 vaultContract,
		ExternalStaking:       externalStakingContract,
		NativeStakingContract: nativeStakingContract,
	}
	fmt.Printf("Provider Contracts:\n  valut: %s\n  ExternalStaking: %s\n  nativeStaking: %s\n  proxycode-id: %d\n",
		r.Vault.String(),
		r.ExternalStaking.String(),
		r.NativeStakingContract.String(),
		proxyCodeID)
	p.Contracts = r
	return r, nil
}

func (p ProviderClient) MustExecVault(payload string, funds ...sdk.Coin) (*sdk.TxResponse, error) {
	return p.mustExec(p.Contracts.Vault, payload, funds)
}

func (p ProviderClient) MustExecExtStaking(payload string, funds ...sdk.Coin) (*sdk.TxResponse, error) {
	return p.mustExec(p.Contracts.ExternalStaking, payload, funds)
}

func (p ProviderClient) mustExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) (*sdk.TxResponse, error) {
	rsp, err := p.Chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.Chain.Address,
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
	return p.mustFailExec(p.Contracts.Vault, payload, funds)
}

// This will execute the contract and assert that it fails, returning the error if desired.
// (Unlike most functions, it will panic on failure, returning error is success case)
func (p ProviderClient) mustFailExec(contract sdk.AccAddress, payload string, funds []sdk.Coin) error {
	resp, err := p.Chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.Chain.Address,
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
	return Querier(p.Chain)(p.Contracts.ExternalStaking.String(), q)
}

func (p ProviderClient) QueryVault(q Query) map[string]any {
	return Querier(p.Chain)(p.Contracts.Vault.String(), q)
}

func (p ProviderClient) QueryVaultFreeBalance() int {
	qRsp := map[string]any{}
	err := Eventually(
		func() bool {
			qRsp = p.QueryVault(Query{
				"account": {"account": p.Chain.Address},
			})
			_, ok := qRsp["locked"]
			if ok {
				return false
			}
			return true
		},
		300*time.Second,
		5*time.Second,
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
	Chain               *Client
	Contracts           *ConsumerContract
	wasmContractPath    string
	wasmContractGZipped bool
}

func NewConsumerClient(chain *Client, wasmContractPath string, wasmContractGZipped bool) *ConsumerClient {
	return &ConsumerClient{Chain: chain, wasmContractPath: wasmContractPath, wasmContractGZipped: wasmContractGZipped}
}

type ConsumerContract struct {
	Staking   sdk.AccAddress
	PriceFeed sdk.AccAddress
	Converter sdk.AccAddress
}

func (p *ConsumerClient) BootstrapContracts(remoteDenom string) (*ConsumerContract, error) {
	// what does this do????
	// modify end-blocker to fail fast in tests
	//msModule := p.app.ModuleManager.Modules[types.ModuleName].(*meshsecurity.AppModule)
	//msModule.SetAsyncTaskRspHandler(meshsecurity.PanicOnErrorExecutionResponseHandler())

	code, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_simple_price_feed.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	codeID := code.CodeID

	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContracts, err := InstantiateContract(p.Chain, codeID, "consumer-price-feeder-contract", initMsg)
	if err != nil {
		return nil, err
	}
	priceFeedContract := priceFeedContracts[codeID]

	// virtual Staking is setup by the consumer
	virtStakeCode, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_virtual_staking.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	virtStakeCodeID := virtStakeCode.CodeID

	// instantiate Converter
	code, err = StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_converter.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	codeID = code.CodeID

	discount := "0.1" // todo: configure price
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract.String(), discount, remoteDenom, virtStakeCodeID))
	// bug in lens that returns second contract instantiated
	contracts, err := InstantiateContract(p.Chain, codeID, "consumer-converter-contract", initMsg)
	if err != nil {
		return nil, err
	}
	converterContract, virtualStakingContract := contracts[codeID], contracts[virtStakeCodeID]

	r := &ConsumerContract{
		Staking:   virtualStakingContract,
		PriceFeed: priceFeedContract,
		Converter: converterContract,
	}
	fmt.Printf("Consumer Contracts:\n  Staking: %s\n  PriceFeed: %s\n  Converter: %s\n", r.Staking.String(), r.PriceFeed.String(), r.Converter.String())
	p.Contracts = r
	return r, nil
}

func (p *ConsumerClient) ExecNewEpoch() {
	// wait for epoch length
	fmt.Printf("sleeping for 150secs.....")
	time.Sleep(150 * time.Second)
}

// MustExecGovProposal submit and vote yes on proposal
func (p *ConsumerClient) MustExecGovProposal(msg *types.MsgSetVirtualStakingMaxCap) {
	proposalID, err := submitGovProposal(p.Chain, msg)
	if err != nil {
		panic(err)
	}
	err = voteAndPassGovProposal(p.Chain, proposalID)
	if err != nil {
		panic(err)
	}
}

func (p *ConsumerClient) QueryMaxCap() types.QueryVirtualStakingMaxCapLimitResponse {
	q := &types.QueryVirtualStakingMaxCapLimitRequest{Address: p.Contracts.Staking.String()}
	rsp, err := types.NewQueryClient(p.Chain.Client).VirtualStakingMaxCapLimit(context.Background(), q)
	if err != nil {
		panic(err)
	}
	return *rsp
}
