package setup

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

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
					} else if strings.Contains(err.Error(), "post failed:") {
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
	Vault                 string
	NativeStakingContract string
	ExternalStaking       string
}

func (p *ProviderClient) StoreContracts() ([]uint64, error) {
	vaultCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_vault.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	vaultCodeID := vaultCodeResp.CodeID
	time.Sleep(6 * time.Second)

	proxyCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_native_staking_proxy.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	proxyCodeID := proxyCodeResp.CodeID
	time.Sleep(6 * time.Second)

	nativeStakingCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_native_staking.wasm", p.wasmContractGZipped))
	nativeStakingCodeID := nativeStakingCodeResp.CodeID

	// external Staking
	extStaking, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_external_staking.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	time.Sleep(6 * time.Second)

	return []uint64{extStaking.CodeID, nativeStakingCodeID, proxyCodeID, vaultCodeID}, nil
}

func (p *ProviderClient) BootstrapContracts(connId, portID, rewardDenom string) (*ProviderContracts, error) {
	var (
		unbondingPeriod           = 100 // 21 days - make configurable?
		localSlashRatioDoubleSign = "0.20"
		localSlashRatioOffline    = "0.10"
		extSlashRatioDoubleSign   = "0.20"
		extSlashRatioOffline      = "0.10"
		localTokenDenom           = p.Chain.Denom
	)
	vaultCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_vault.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	vaultCodeID := vaultCodeResp.CodeID
	proxyCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_native_staking_proxy.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	proxyCodeID := proxyCodeResp.CodeID
	nativeStakingCodeResp, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_native_staking.wasm", p.wasmContractGZipped))
	nativeStakingCodeID := nativeStakingCodeResp.CodeID

	time.Sleep(time.Second)

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d, "slash_ratio_dsign": %q, "slash_ratio_offline": %q }`, localTokenDenom, proxyCodeID, localSlashRatioDoubleSign, localSlashRatioOffline))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, localTokenDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	contracts, err := InstantiateContract(p.Chain, vaultCodeID, "provider-valut-contract", initMsg)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	vaultContract := contracts[vaultCodeID]
	nativeStakingContract := contracts[nativeStakingCodeID]

	// external Staking
	extStaking, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_external_staking.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second)

	extStakingCodeID := extStaking.CodeID
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q, "slash_ratio": { "double_sign": %q, "offline": %q } }`,
		connId, portID, localTokenDenom, vaultContract, unbondingPeriod, rewardDenom, extSlashRatioDoubleSign, extSlashRatioOffline))
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
		r.Vault,
		r.ExternalStaking,
		r.NativeStakingContract,
		proxyCodeID)
	p.Contracts = r
	time.Sleep(time.Second)
	return r, nil
}

func (p ProviderClient) MustExecVault(payload string, funds ...sdk.Coin) (*sdk.TxResponse, error) {
	return p.mustExec(p.Contracts.Vault, payload, funds)
}

func (p ProviderClient) MustExecExtStaking(payload string, funds ...sdk.Coin) (*sdk.TxResponse, error) {
	return p.mustExec(p.Contracts.ExternalStaking, payload, funds)
}

func (p ProviderClient) mustExec(contract string, payload string, funds []sdk.Coin) (*sdk.TxResponse, error) {
	rsp, err := p.Chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.Chain.Address,
		Contract: contract,
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
func (p ProviderClient) mustFailExec(contract string, payload string, funds []sdk.Coin) error {
	resp, err := p.Chain.Client.SendMsg(context.Background(), &wasmtypes.MsgExecuteContract{
		Sender:   p.Chain.Address,
		Contract: contract,
		Msg:      []byte(payload),
		Funds:    funds,
	}, "")
	if err != nil {
		fmt.Printf("error wasm exec: error: %s, response: %v\n", err, resp)
	}

	return err
}

func (p ProviderClient) QueryExtStaking(q Query) map[string]any {
	return Querier(p.Chain)(p.Contracts.ExternalStaking, q)
}

func (p ProviderClient) QueryVault(q Query) map[string]any {
	return Querier(p.Chain)(p.Contracts.Vault, q)
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
		"vault token locked for too long: %v",
		qRsp,
	)
	if err != nil {
		panic(err)
	}
	return ParseHighLow(qRsp["free"]).Low
}

type HighLowType struct {
	High, Low int
}

func ParseHighLow(a any) HighLowType {
	m, ok := a.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("unsupported type %T", a))
	}
	h, err := strconv.Atoi(m["h"].(string))
	if err != nil {
		panic(fmt.Sprintf("high: %s", err))
	}
	l, err := strconv.Atoi(m["l"].(string))
	if err != nil {
		panic(fmt.Sprintf("low: %s", err))
	}
	return HighLowType{High: h, Low: l}
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
	Staking   string
	PriceFeed string
	Converter string
}

func (p *ConsumerClient) StoreContracts() ([]uint64, error) {
	code, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_simple_price_feed.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	codeID := code.CodeID
	time.Sleep(6 * time.Second)

	// virtual Staking is setup by the consumer
	virtStakeCode, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_virtual_staking.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	virtStakeCodeID := virtStakeCode.CodeID
	time.Sleep(6 * time.Second)

	// instantiate Converter
	converterCode, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_converter.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	converterCodeID := converterCode.CodeID
	time.Sleep(6 * time.Second)

	return []uint64{codeID, virtStakeCodeID, converterCodeID}, nil
}

func (p *ConsumerClient) BootstrapContracts(remoteDenom string) (*ConsumerContract, error) {
	// what does this do????
	// modify end-blocker to fail fast in tests
	// msModule := p.app.ModuleManager.Modules[types.ModuleName].(*meshsecurity.AppModule)
	// msModule.SetAsyncTaskRspHandler(meshsecurity.PanicOnErrorExecutionResponseHandler())

	code, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_simple_price_feed.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	codeID := code.CodeID
	time.Sleep(time.Second)

	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContracts, err := InstantiateContract(p.Chain, codeID, "consumer-price-feeder-contract", initMsg)
	if err != nil {
		return nil, err
	}
	priceFeedContract := priceFeedContracts[codeID]
	time.Sleep(time.Second)

	// virtual Staking is setup by the consumer
	virtStakeCode, err := StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_virtual_staking.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	virtStakeCodeID := virtStakeCode.CodeID
	time.Sleep(time.Second)

	// instantiate Converter
	code, err = StoreCodeFile(p.Chain, buildPathToWasm(p.wasmContractPath, "mesh_converter.wasm", p.wasmContractGZipped))
	if err != nil {
		return nil, err
	}
	codeID = code.CodeID
	time.Sleep(time.Second)

	discount := "0.1" // todo: configure price
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract, discount, remoteDenom, virtStakeCodeID))
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
	fmt.Printf("Consumer Contracts:\n  Staking: %s\n  PriceFeed: %s\n  Converter: %s\n", r.Staking, r.PriceFeed, r.Converter)
	p.Contracts = r
	time.Sleep(time.Second)
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
	q := &types.QueryVirtualStakingMaxCapLimitRequest{Address: p.Contracts.Staking}
	rsp, err := types.NewQueryClient(p.Chain.Client).VirtualStakingMaxCapLimit(context.Background(), q)
	if err != nil {
		panic(err)
	}
	return *rsp
}
