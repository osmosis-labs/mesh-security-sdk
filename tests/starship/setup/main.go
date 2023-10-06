package setup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	starship "github.com/cosmology-tech/starship/clients/go/client"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func MeshSecurity(provider, consumer, configFile, wasmContractPath string, wasmContractGZipped bool) (*ProviderClient, *ConsumerClient, error) {
	// read config file from yaml
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, nil, err
	}
	config := &starship.Config{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, nil, err
	}

	// create Chain clients
	chainClients, err := starship.NewChainClients(zap.L(), config)
	if err != nil {
		return nil, nil, err
	}

	var (
		consumerChain, _ = chainClients.GetChainClient(consumer)
		providerChain, _ = chainClients.GetChainClient(provider)
	)

	// create lens Client for the provider and consumer chains
	mm := []module.AppModuleBasic{}
	for _, am := range app.ModuleBasics {
		mm = append(mm, am)
	}
	consumerClient, err := NewClient(fmt.Sprintf("consume-client-%s", consumer), zap.L(), consumerChain, mm)
	if err != nil {
		return nil, nil, err
	}
	providerClient, err := NewClient(fmt.Sprintf("provider-client-%s", provider), zap.L(), providerChain, mm)
	if err != nil {
		return nil, nil, err
	}

	// transfer ibc tokens from consumer to provider
	err = IBCTransferTokens(consumerClient, providerClient, providerClient.Address, 123000000)
	if err != nil {
		return nil, nil, err
	}
	time.Sleep(5 * time.Second)

	// get ibc denom from account balance
	coins, err := GetBalance(providerClient, providerClient.Address)
	if err != nil {
		return nil, nil, err
	}
	ibcDenom := ""
	for _, denom := range coins.Denoms() {
		if strings.HasPrefix(denom, "ibc/") {
			ibcDenom = denom
		}
	}
	if ibcDenom == "" {
		return nil, nil, fmt.Errorf("ibc denom not found in balance: %v", coins)
	}

	// setup Contracts on both chains
	consumerCli := NewConsumerClient(consumerClient, wasmContractPath, wasmContractGZipped)
	consumerContracts, err := consumerCli.BootstrapContracts(providerClient.Denom)
	if err != nil {
		return nil, nil, err
	}
	converterPortID := portIDForContract(consumerContracts.Converter)
	providerCli := NewProviderClient(providerClient, wasmContractPath, wasmContractGZipped)

	ibcInfo, err := consumerChain.GetIBCInfo(provider)
	if err != nil {
		return nil, nil, err
	}

	connectionID := ibcInfo.Chain_1.ConnectionId

	// set reward denom to the other chain base denom
	rewardDenom := consumerCli.Chain.Denom
	providerContracts, err := providerCli.BootstrapContracts(connectionID, converterPortID, rewardDenom)
	if err != nil {
		return nil, nil, err
	}

	// create channel between 2 chains for the given port and channel
	cmdRunner, err := NewCmdRunner(zap.L(), config)
	if err != nil {
		return nil, nil, err
	}

	providerPortID := portIDForContract(providerContracts.ExternalStaking)

	cmd := fmt.Sprintf("hermes create channel --a-chain %s --a-connection %s --a-port %s --b-port %s --yes", consumer, connectionID, converterPortID, providerPortID)
	output, err := cmdRunner.RunExec(config.Relayers[0].Name, cmd)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("hermes output: ", output)

	// wait for initial packets to be transferred via IBC over
	validators, err := stakingtypes.NewQueryClient(consumerClient.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	if err != nil {
		return nil, nil, err
	}
	myExtValidatorAddr := validators.Validators[0].OperatorAddress

	// stake tokens from the client address
	err = consumerClient.StakeTokens(myExtValidatorAddr, 5000000, consumerClient.Denom)
	if err != nil {
		return nil, nil, err
	}

	// then the active set should be stored in the ext Staking contract
	// and contain all active validator addresses
	qRsp := map[string]any{}
	err = Eventually(
		func() bool {
			qRsp = providerCli.QueryExtStaking(Query{"list_remote_validators": {}})
			v := qRsp["validators"].([]interface{})
			if len(v) > 0 {
				return true
			}
			return false
		},
		300*time.Second,
		5*time.Second,
		"list remote validators failed: %v",
		qRsp,
	)
	if err != nil {
		return nil, nil, err
	}

	// add authority to mint/burn virtual tokens gov proposal
	fmt.Println("add auth to mint/burn virtual tokens")
	// bech32Addr, err := bech32.ConvertAndEncode(prefix, addr)
	authAddr := authtypes.NewModuleAddress(govtypes.ModuleName)
	registry, err := consumerChain.GetChainRegistry()
	if err != nil {
		return nil, nil, err
	}
	authAddrStr, err := bech32.ConvertAndEncode(*registry.Bech32Prefix, authAddr)
	govProposal := &types.MsgSetVirtualStakingMaxCap{
		Authority: authAddrStr,
		Contract:  consumerContracts.Staking,
		MaxCap:    sdk.NewInt64Coin(consumerClient.Denom, 1_000_000_000),
	}
	fmt.Printf("create a gov proposal: %v\n", govProposal)
	consumerCli.MustExecGovProposal(govProposal)

	return providerCli, consumerCli, nil
}
