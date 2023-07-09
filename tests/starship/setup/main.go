package main

import (
	"context"
	"flag"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	starship "github.com/cosmology-tech/starship/clients/go/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

func MeshSecurity() error {
	// read config file from yaml
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	config := &starship.Config{}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return err
	}

	// create chain clients
	chainClients, err := starship.NewChainClients(zap.L(), config)
	if err != nil {
		return err
	}

	var (
		consumerChain, _ = chainClients.GetChainClient("consumer")
		providerChain, _ = chainClients.GetChainClient("provider")
	)

	// create Client
	mm := []module.AppModuleBasic{}
	for _, am := range app.ModuleBasics {
		mm = append(mm, am)
	}
	consumerClient, err := NewClient("consume-client", zap.L(), consumerChain, mm)
	if err != nil {
		return err
	}
	providerClient, err := NewClient("provider-client", zap.L(), providerChain, mm)
	if err != nil {
		return err
	}

	// setup contracts on both chains
	consumerCli := NewConsumerClient(consumerClient)
	consumerContracts, err := consumerCli.BootstrapContracts()
	if err != nil {
		return err
	}
	converterPortID := wasmkeeper.PortIDForContract(consumerContracts.converter)
	providerCli := NewProviderClient(providerClient)

	ibcInfo, err := consumerChain.GetIBCInfo("provider")
	if err != nil {
		return err
	}

	connectionID := ibcInfo.Chain_1.ConnectionId
	providerContracts, err := providerCli.BootstrapContracts(connectionID, converterPortID)
	if err != nil {
		return err
	}

	// create channel between 2 chains for the given port and channel
	cmdRunner, err := starship.NewCmdRunner(zap.L(), config)
	if err != nil {
		return err
	}

	consumerPortID := wasmkeeper.PortIDForContract(providerContracts.externalStaking)

	cmd := fmt.Sprintf("hermes create channel --a-chain %s --a-connection %s --a-port %s --b-port %s --yes", "consumer", connectionID, converterPortID, consumerPortID)
	err = cmdRunner.RunExec("provider-consumer", cmd)
	if err != nil {
		return err
	}

	// wait for initial packets to be transfered via IBC over
	validators, err := stakingtypes.NewQueryClient(consumerClient.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	if err != nil {
		return err
	}
	myExtValidatorAddr := validators.Validators[0].OperatorAddress

	// stake tokens from the client address
	err = consumerClient.StakeTokens(myExtValidatorAddr, 5000000, consumerClient.Denom)
	if err != nil {
		return err
	}

	// then the active set should be stored in the ext staking contract
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
		120*time.Second,
		time.Second,
		"list remote validators failed: %v",
		qRsp,
	)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.StringVar(&wasmContractPath, "contracts-path", "../testdata", "Set path to dir with gzipped wasm contracts")
	flag.BoolVar(&wasmContractGZipped, "gzipped", true, "Use `.gz` file ending when set")
	flag.StringVar(&configFile, "config", "../configs/starship.yaml", "starship config file for the infra")
	flag.Parse()

	err := MeshSecurity()
	if err != nil {
		panic(err)
	}
}
