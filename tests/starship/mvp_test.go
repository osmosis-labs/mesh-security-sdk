package e2e

import (
	"context"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	starship "github.com/cosmology-tech/starship/clients/go/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
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

	time.Sleep(5 * time.Second)

	validators, err := stakingtypes.NewQueryClient(consumerClient.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)

	// then the active set should be stored in the ext staking contract
	// and contain all active validator addresses
	qRsp := providerCli.QueryExtStaking(Query{"list_remote_validators": {}})
	require.Len(t, qRsp["validators"], len(validators.Validators), qRsp)
	for _, v := range validators.Validators {
		require.Contains(t, qRsp["validators"], v.OperatorAddress)
	}
	//
	//// ----------------------------
	// ensure nothing staked by the virtual staking contract yet
	myExtValidatorAddr := validators.Validators[0].OperatorAddress

	query := &stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: string(consumerContracts.staking),
		ValidatorAddr: myExtValidatorAddr,
	}
	delegations, err := stakingtypes.NewQueryClient(consumerClient.Client).DelegatorValidator(context.Background(), query)
	require.NoError(t, err)
	require.Empty(t, delegations)

	//// add authority to mint/burn virtual tokens gov proposal
	//govProposal := &types.MsgSetVirtualStakingMaxCap{
	//	Authority: consumerApp.MeshSecKeeper.GetAuthority(),
	//	Contract:  consumerContracts.staking.String(),
	//	MaxCap:    sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000),
	//}
	//consumerCli.MustExecGovProposal(govProposal)
	//
	//// then the max cap limit is persisted
	//rsp := consumerCli.QueryMaxCap()
	//assert.Equal(t, sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000), rsp.Cap)
	//
	//// provider chain
	//// ==============
	//// Deposit - A user deposits the vault denom to provide some collateral to their account
	//execMsg := `{"bond":{}}`
	//providerCli.MustExecVault(execMsg, sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000))
	//
	//// then query contract state
	//assert.Equal(t, 100_000_000, providerCli.QueryVaultFreeBalance())
	//
	//// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	//myLocalValidatorAddr := sdk.ValAddress(providerChain.Vals.Validators[0].Address).String()
	//execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
	//	sdk.DefaultBondDenom, 30_000_000,
	//	base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	//providerCli.MustExecVault(execLocalStakingMsg)
	//
	//assert.Equal(t, 70_000_000, providerCli.QueryVaultFreeBalance())
	//
	//// // Failure mode of cross-stake... trying to stake to an unknown validator
	//execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
	//	providerContracts.externalStaking.String(),
	//	sdk.DefaultBondDenom, 80_000_000,
	//	base64.StdEncoding.EncodeToString([]byte(`{"validator": "BAD-VALIDATOR"}`)))
	//_ = providerCli.MustFailExecVault(execMsg)
	//// // no change to free balance
	//assert.Equal(t, 70_000_000, providerCli.QueryVaultFreeBalance())
	//
	//// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	//execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
	//	providerContracts.externalStaking.String(),
	//	sdk.DefaultBondDenom, 80_000_000,
	//	base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr))))
	//providerCli.MustExecVault(execMsg)
	//
	//require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
	//require.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance()) // = 70 (free)  + 30 (local) - 80 (remote staked)
	//
	//// then
	//qRsp = providerCli.QueryExtStaking(Query{
	//	"stake": {
	//		"user":      providerChain.SenderAccount.GetAddress().String(),
	//		"validator": myExtValidatorAddr,
	//	},
	//})
	//assert.Equal(t, "80000000", qRsp["stake"], qRsp)
	//assert.Empty(t, qRsp["pending_unbonds"])
	//
	//// consumer chain
	//// ====================
	////
	//// then delegated amount is not updated before the epoch
	//consumerCli.assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet
	//
	//// when an epoch ends, the delegation rebalance is triggered
	//consumerCli.ExecNewEpoch()
	//
	//// then the total delegated amount is updated
	//consumerCli.assertTotalDelegated(math.NewInt(36_000_000)) // 80_000_000 /2 * (1 -0.1)
	//
	//// and the delegated amount is updated for the validator
	//consumerCli.assertShare(myExtValidator, math.LegacyNewDec(36)) // 36_000_000 / 1_000_000 # default sdk factor
	//
	//// provider chain
	//// ==============
	////
	//// Cross Stake - A user undelegates
	//execMsg = fmt.Sprintf(`{"unstake":{"validator":"%s", "amount":{"denom":"%s", "amount":"30000000"}}}`, myExtValidator.String(), sdk.DefaultBondDenom)
	//providerCli.MustExecExtStaking(execMsg)
	//
	//require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
	//
	//// then
	//qRsp = providerCli.QueryExtStaking(Query{
	//	"stake": {
	//		"user":      providerChain.SenderAccount.GetAddress().String(),
	//		"validator": myExtValidatorAddr,
	//	},
	//})
	//require.Equal(t, "50000000", qRsp["stake"], qRsp)
	//require.Len(t, qRsp["pending_unbonds"], 1)
	//unbonds := qRsp["pending_unbonds"].([]any)[0].(map[string]any)
	//assert.Equal(t, "30000000", unbonds["amount"], qRsp)
	//
	//// consumer chain
	//// ====================
	//
	//consumerCli.ExecNewEpoch()
	//
	//// then the total delegated amount is updated
	//consumerCli.assertTotalDelegated(math.NewInt(22_500_000))                  // (80_000_000 - 30_000_000) /2 * (1 -0.1)
	//consumerCli.assertShare(myExtValidator, math.LegacyNewDecWithPrec(225, 1)) // 27_000_000 / 1_000_000 # default sdk factor
	//
	//// provider chain
	//// ==============
	////
	//// A user withdraws the undelegated amount
	//
	//require.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance())
	//
	//releaseData := unbonds["release_at"].(string)
	//require.NotEmpty(t, releaseData)
	//at, err := strconv.Atoi(releaseData)
	//require.NoError(t, err)
	//releasedAt := time.Unix(0, int64(at)).UTC()
	//// update system time
	//coord.CurrentTime = releasedAt.Add(time.Minute)
	//coord.UpdateTime()
	//coord.CommitBlock(providerChain, consumerChain)
	//
	//providerCli.MustExecExtStaking(`{"withdraw_unbonded":{}}`)
	//assert.Equal(t, 50_000_000, providerCli.QueryVaultFreeBalance())
	//
	//// provider chain
	//// ==============
	////
	//// A user unstakes some free amount from the vault
	//balanceBefore := providerChain.Balance(providerChain.SenderAccount.GetAddress(), "stake")
	//providerCli.MustExecVault(`{"unbond":{"amount":{"denom":"stake", "amount": "30000000"}}}`)
	//// then
	//assert.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance())
	//balanceAfter := providerChain.Balance(providerChain.SenderAccount.GetAddress(), "stake")
	//assert.Equal(t, math.NewInt(30_000_000), balanceAfter.Sub(balanceBefore).Amount)
}
