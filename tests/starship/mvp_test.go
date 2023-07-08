package e2e

import (
	"context"
	"cosmossdk.io/math"
	"encoding/base64"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	starship "github.com/cosmology-tech/starship/clients/go/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMVP(t *testing.T) {
	// scenario:
	// given a provider chain P and a consumer chain C
	// when	 a user on P deposits an amount as collateral in the vault contract
	// then	 it can be used as "lien" to stake local and to "cross stake" on chain C
	// when  an amount is "cross staked" to a validator on chain C
	// and	 the ibc package is relayed
	// then  the amount is converted into an amount in the chain C bonding token
	// and   scheduled to be staked as synthetic token on the validator
	// when  the next epoch is executed on chain C
	// then  the synthetic tokens are minted and staked
	//
	// when  the user on chain P starts an undelegate
	//
	// ...

	// read config file from yaml
	yamlFile, err := os.ReadFile(configFile)
	require.NoError(t, err)
	config := &starship.Config{}
	err = yaml.Unmarshal(yamlFile, config)
	require.NoError(t, err)

	// create chain clients
	chainClients, err := starship.NewChainClients(zap.L(), config)
	require.NoError(t, err)

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
	require.NoError(t, err)
	providerClient, err := NewClient("provider-client", zap.L(), providerChain, mm)
	require.NoError(t, err)

	// setup contracts on both chains
	consumerCli := NewConsumerClient(t, consumerClient)
	consumerContracts := consumerCli.BootstrapContracts()
	converterPortID := wasmkeeper.PortIDForContract(consumerContracts.converter)
	providerCli := NewProviderClient(t, providerClient)

	ibcInfo, err := consumerChain.GetIBCInfo("provider")
	require.NoError(t, err)

	connectionID := ibcInfo.Chain_1.ConnectionId
	providerContracts := providerCli.BootstrapContracts(connectionID, converterPortID)

	require.NotEmpty(t, providerContracts)

	// create channel between 2 chains for the given port and channel
	cmdRunner, err := starship.NewCmdRunner(zap.L(), config)
	require.NoError(t, err)

	consumerPortID := wasmkeeper.PortIDForContract(providerContracts.externalStaking)

	cmd := fmt.Sprintf("hermes create channel --a-chain %s --a-connection %s --a-port %s --b-port %s --yes", "consumer", connectionID, converterPortID, consumerPortID)
	err = cmdRunner.RunExec("provider-consumer", cmd)
	require.NoError(t, err)

	//
	//// setup ibc control path: consumer -> provider (direction matters)
	//ibcPath.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
	//	PortID: wasmkeeper.PortIDForContract(providerContracts.externalStaking),
	//	Order:  channeltypes.UNORDERED,
	//}
	//ibcPath.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
	//	PortID: converterPortID,
	//	Order:  channeltypes.UNORDERED,
	//}
	//coord.CreateChannels(ibcPath)
	//
	// when ibc package is relayed
	//require.NotEmpty(t, consumerChain.PendingSendPackets)
	//require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))

	// wait for initial packets to be transfered via IBC over
	validators, err := stakingtypes.NewQueryClient(consumerClient.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myExtValidatorAddr := validators.Validators[0].OperatorAddress

	// stake tokens from the client address
	err = consumerClient.StakeTokens(myExtValidatorAddr, 5000000, "ustake")
	require.NoError(t, err)

	// then the active set should be stored in the ext staking contract
	// and contain all active validator addresses
	qRsp := map[string]any{}

	require.Eventuallyf(t,
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

	require.NotEmpty(t, qRsp)
	require.Len(t, qRsp["validators"], len(validators.Validators), qRsp)
	for _, v := range validators.Validators {
		require.NotEmpty(t, v)
		require.Contains(t, qRsp["validators"], v.OperatorAddress)
	}
	//
	//// ----------------------------
	// ensure nothing staked by the virtual staking contract yet

	fmt.Println("ensure nothing staked by virtual contract")
	query := &stakingtypes.QueryDelegatorValidatorsRequest{
		DelegatorAddr: consumerContracts.staking.String(),
	}
	delegations, err := stakingtypes.NewQueryClient(consumerClient.Client).DelegatorValidators(context.Background(), query)
	require.NoError(t, err)
	require.Empty(t, delegations.Validators)

	// add authority to mint/burn virtual tokens gov proposal
	fmt.Println("add auth to mint/burn virtual tokens")
	govProposal := &types.MsgSetVirtualStakingMaxCap{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Contract:  consumerContracts.staking.String(),
		MaxCap:    sdk.NewInt64Coin("ustake", 1_000_000_000),
	}
	fmt.Printf("create a gov proposal: %v\n", govProposal)
	consumerCli.MustExecGovProposal(govProposal)

	// then the max cap limit is persisted
	rsp := consumerCli.QueryMaxCap()
	require.Equal(t, sdk.NewInt64Coin("ustake", 1_000_000_000), rsp.Cap)

	// provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	fmt.Println("provider chain: deposit vault denom to provide some collateral to account")
	execMsg := `{"bond":{}}`
	providerCli.MustExecVault(execMsg, sdk.NewInt64Coin("ustake", 100_000_000))

	// then query contract state
	assert.Equal(t, 100_000_000, providerCli.QueryVaultFreeBalance())

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	fmt.Println("provider chain: stake locally")
	providerValidators, err := stakingtypes.NewQueryClient(providerClient.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myLocalValidatorAddr := providerValidators.Validators[0].OperatorAddress
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		"ustake", 30_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	providerCli.MustExecVault(execLocalStakingMsg)

	assert.Equal(t, 70_000_000, providerCli.QueryVaultFreeBalance())

	// // Failure mode of cross-stake... trying to stake to an unknown validator
	fmt.Println("provider chain: failure care, trying to stake to unknown validator")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerContracts.externalStaking.String(),
		"ustake", 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(`{"validator": "BAD-VALIDATOR"}`)))
	_ = providerCli.MustFailExecVault(execMsg)
	// // no change to free balance
	assert.Equal(t, 70_000_000, providerCli.QueryVaultFreeBalance())

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	fmt.Println("cross stake: additional liens on the same collateral")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerContracts.externalStaking.String(),
		"ustake", 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr))))
	providerCli.MustExecVault(execMsg)

	//require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
	require.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance()) // = 70 (free)  + 30 (local) - 80 (remote staked)

	// then
	fmt.Println("provider chain: query ext staking")
	qRsp = providerCli.QueryExtStaking(Query{
		"stake": {
			"user":      providerClient.Address,
			"validator": myExtValidatorAddr,
		},
	})
	assert.Equal(t, "80000000", qRsp["stake"], qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])

	//consumer chain
	// ====================
	//
	fmt.Println("consumer chain: deleted amount is not updated before epoch")
	// then delegated amount is not updated before the epoch
	consumerCli.assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(36_000_000)) // 80_000_000 /2 * (1 -0.1)

	// and the delegated amount is updated for the validator
	consumerCli.assertShare(myExtValidatorAddr, math.LegacyNewDec(36_000_000)) // 36_000_000 / 1_000_000 # default sdk factor

	// provider chain
	// ==============
	// then
	qRsp = providerCli.QueryExtStaking(Query{
		"stake": {
			"user":      providerClient.Address,
			"validator": myExtValidatorAddr,
		},
	})
	assert.Equal(t, "80000000", qRsp["stake"], qRsp)

	// Cross Stake - A user undelegates
	fmt.Println("provider chain: cross stake user undelegates")
	execMsg = fmt.Sprintf(`{"unstake":{"validator":"%s", "amount":{"denom":"%s", "amount":"30000000"}}}`, myExtValidatorAddr, "ustake")
	providerCli.MustExecExtStaking(execMsg)

	// then
	qRsp = providerCli.QueryExtStaking(Query{
		"stake": {
			"user":      providerClient.Address,
			"validator": myExtValidatorAddr,
		},
	})
	require.Equal(t, "50000000", qRsp["stake"], qRsp)
	require.Len(t, qRsp["pending_unbonds"], 1)
	unbonds := qRsp["pending_unbonds"].([]any)[0].(map[string]any)
	assert.Equal(t, "30000000", unbonds["amount"], qRsp)

	// consumer chain
	// ====================

	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(22_500_000))                  // (80_000_000 - 30_000_000) /2 * (1 -0.1)
	consumerCli.assertShare(myExtValidatorAddr, math.LegacyNewDec(22_500_000)) // 27_000_000 / 1_000_000 # default sdk factor

	// provider chain
	// ==============
	//
	// A user withdraws the undelegated amount

	require.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance())

	releaseData := unbonds["release_at"].(string)
	require.NotEmpty(t, releaseData)
	at, err := strconv.Atoi(releaseData)
	require.NoError(t, err)
	releasedAt := time.Unix(0, int64(at)).UTC()
	fmt.Printf("unbonding at: %v, time to: %v\n", releasedAt, releasedAt.Add(time.Minute).Sub(time.Now()))

	//providerCli.MustExecExtStaking(`{"withdraw_unbonded":{}}`)
	//assert.Equal(t, 50_000_000, providerCli.QueryVaultFreeBalance())
	//
	//// provider chain
	//// ==============
	////
	//// A user unstakes some free amount from the vault
	//balanceBefore, err := banktypes.NewQueryClient(providerClient.Client).Balance(context.Background(), &banktypes.QueryBalanceRequest{
	//	Address: providerClient.Address,
	//	Denom:   "ustake",
	//})
	//require.NoError(t, err)
	//providerCli.MustExecVault(`{"unbond":{"amount":{"denom":"stake", "amount": "30000000"}}}`)
	//// then
	//assert.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance())
	//balanceAfter, err := banktypes.NewQueryClient(providerClient.Client).Balance(context.Background(), &banktypes.QueryBalanceRequest{
	//	Address: providerClient.Address,
	//	Denom:   "ustake",
	//})
	//require.NoError(t, err)
	//assert.Equal(t, math.NewInt(30_000_000), balanceAfter.Balance.Sub(*balanceBefore.Balance).Amount)
}
