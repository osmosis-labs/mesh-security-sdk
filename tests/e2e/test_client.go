package e2e

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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

type TestConsumerClient struct {
	t                 *testing.T
	chain             ibctesting.TestChain
	consumerContracts ConsumerContract
}

func NewConsumerClient(t *testing.T, chain ibctesting.TestChain) *TestConsumerClient {
	return &TestConsumerClient{t: t, chain: chain}
}

type TestProviderClient struct {
	t                 *testing.T
	chain             *ibctesting.TestChain
	providerContracts ProviderContracts
}

type ProviderContracts struct {
	vault           types.AccAddress
	externalStaking types.AccAddress
}

func (p *TestProviderClient) bootstrapProviderContracts(t *testing.T, connId, portID string) ProviderContracts {
	chain := p.chain
	vaultCodeID := chain.StoreCodeFile(buildPathToWasm("mesh_vault.wasm")).CodeID
	proxyCodeID := chain.StoreCodeFile(buildPathToWasm("mesh_native_staking_proxy.wasm")).CodeID
	nativeStakingCodeID := chain.StoreCodeFile(buildPathToWasm("mesh_native_staking.wasm")).CodeID

	nativeInitMsg := []byte(fmt.Sprintf(`{"denom": %q, "proxy_code_id": %d}`, types.DefaultBondDenom, proxyCodeID))
	initMsg := []byte(fmt.Sprintf(`{"denom": %q, "local_staking": {"code_id": %d, "msg": %q}}`, types.DefaultBondDenom, nativeStakingCodeID, base64.StdEncoding.EncodeToString(nativeInitMsg)))
	vaultContract := InstantiateContract(t, chain, vaultCodeID, initMsg)

	// external staking
	unbondingPeriod := 21 * 24 * 60 * 60 // 21 days - make configurable?
	extStakingCodeID := chain.StoreCodeFile(buildPathToWasm("external_staking.wasm")).CodeID
	rewardToken := "todo" // ics20 token
	initMsg = []byte(fmt.Sprintf(
		`{"remote_contact": {"connection_id":%q, "port_id":%q}, "denom": %q, "vault": %q, "unbonding_period": %d, "rewards_denom": %q}`,
		connId, portID, types.DefaultBondDenom, vaultContract.String(), unbondingPeriod, rewardToken))
	externalStakingContract := InstantiateContract(t, chain, extStakingCodeID, initMsg)

	return ProviderContracts{
		vault:           vaultContract,
		externalStaking: externalStakingContract,
	}
}

type ConsumerContract struct {
	staking   types.AccAddress
	priceFeed types.AccAddress
	converter types.AccAddress
}

func (p *TestConsumerClient) bootstrapContracts(t *testing.T, consumerChain *ibctesting.TestChain) ConsumerContract {
	codeID := consumerChain.StoreCodeFile(buildPathToWasm("mesh_simple_price_feed.wasm")).CodeID
	initMsg := []byte(fmt.Sprintf(`{"native_per_foreign": "%s"}`, "0.5")) // todo: configure price
	priceFeedContract := InstantiateContract(t, consumerChain, codeID, initMsg)
	// virtual staking is setup by the consumer
	virtStakeCodeID := consumerChain.StoreCodeFile(buildPathToWasm("mesh_virtual_staking.wasm")).CodeID
	// instantiate converter
	codeID = consumerChain.StoreCodeFile(buildPathToWasm("mesh_converter.wasm")).CodeID
	discount := "0.1" // todo: configure price
	initMsg = []byte(fmt.Sprintf(`{"price_feed": %q, "discount": %q, "remote_denom": %q,"virtual_staking_code_id": %d}`,
		priceFeedContract.String(), discount, types.DefaultBondDenom, virtStakeCodeID))
	converterContract := InstantiateContract(t, consumerChain, codeID, initMsg)

	staking := Querier(t, consumerChain)(converterContract.String(), Query{"config": {}})["virtual_staking"]
	return ConsumerContract{
		staking:   types.MustAccAddressFromBech32(staking.(string)),
		priceFeed: priceFeedContract,
		converter: converterContract,
	}
}
