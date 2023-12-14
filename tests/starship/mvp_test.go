package starship

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/tests/starship/setup"
)

func Test2WayContract(t *testing.T) {
	// create clients for provider and consumer
	providerClient1, consumerClient1, err := setup.MeshSecurity(providerChain, consumerChain, configFile, wasmContractPath, wasmContractGZipped)
	require.NoError(t, err)
	require.NotEmpty(t, providerClient1)
	require.NotEmpty(t, consumerClient1)

	qRsp := map[string]any{}
	// check list of validators on each chains
	require.Eventuallyf(t,
		func() bool {
			qRsp = providerClient1.QueryExtStaking(setup.Query{"list_active_validators": {}})
			v := qRsp["validators"].([]interface{})
			if len(v) > 0 {
				return true
			}
			return false
		},
		120*time.Second,
		5*time.Second,
		"list remote validators failed: %v",
		qRsp,
	)

	// provider chain 1
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	fmt.Println("provider chain: deposit vault denom to provide some collateral to account")
	execMsg := `{"bond":{}}`
	vault, err := providerClient1.MustExecVault(execMsg, sdk.NewInt64Coin(providerClient1.Chain.Denom, 100_000_000))
	require.NoError(t, err)
	require.NotEmpty(t, vault)

	// then query contract state
	assert.Equal(t, 100_000_000, providerClient1.QueryVaultFreeBalance())

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	fmt.Println("provider chain: stake locally")
	providerValidators1, err := stakingtypes.NewQueryClient(providerClient1.Chain.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myLocalValidatorAddr1 := providerValidators1.Validators[0].OperatorAddress
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient1.Chain.Denom, 30_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr1))))
	_, err = providerClient1.MustExecVault(execLocalStakingMsg)
	require.NoError(t, err)

	assert.Equal(t, 70_000_000, providerClient1.QueryVaultFreeBalance())

	// wait for initial packets to be transferred via IBC over
	validators, err := stakingtypes.NewQueryClient(consumerClient1.Chain.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myExtValidatorAddr1 := validators.Validators[0].OperatorAddress

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	fmt.Println("cross stake: additional liens on the same collateral")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient1.Contracts.ExternalStaking,
		providerClient1.Chain.Denom, 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr1))))
	_, err = providerClient1.MustExecVault(execMsg)
	require.NoError(t, err)

	// require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
	require.Equal(t, 20_000_000, providerClient1.QueryVaultFreeBalance()) // = 70 (free)  + 30 (local) - 80 (remote staked)

	// then
	fmt.Println("provider chain: query ext staking")
	qRsp = providerClient1.QueryExtStaking(setup.Query{
		"stake": {
			"user":      providerClient1.Chain.Address,
			"validator": myExtValidatorAddr1,
		},
	})
	assert.Equal(t, 80_000_000, setup.ParseHighLow(qRsp["stake"]).Low, qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])

	// create opposite clients
	providerClient2, consumerClient2, err := setup.MeshSecurity(consumerChain, providerChain, configFile, wasmContractPath, wasmContractGZipped)
	require.NoError(t, err)
	require.NotEmpty(t, providerClient2)
	require.NotEmpty(t, consumerClient2)

	require.Eventuallyf(t,
		func() bool {
			qRsp = providerClient2.QueryExtStaking(setup.Query{"list_active_validators": {}})
			v := qRsp["validators"].([]interface{})
			if len(v) > 0 {
				return true
			}
			return false
		},
		120*time.Second,
		5*time.Second,
		"list remote validators failed: %v",
		qRsp,
	)

	// provider chain 2
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	fmt.Println("provider chain: deposit vault denom to provide some collateral to account")
	execMsg = `{"bond":{}}`
	vault, err = providerClient2.MustExecVault(execMsg, sdk.NewInt64Coin(providerClient2.Chain.Denom, 100_000_000))
	require.NoError(t, err)
	require.NotEmpty(t, vault)

	// then query contract state
	assert.Equal(t, 100_000_000, providerClient2.QueryVaultFreeBalance())

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	fmt.Println("provider chain: stake locally")
	providerValidators2, err := stakingtypes.NewQueryClient(providerClient2.Chain.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myLocalValidatorAddr2 := providerValidators2.Validators[0].OperatorAddress
	execLocalStakingMsg = fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient2.Chain.Denom, 30_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr2))))
	_, err = providerClient2.MustExecVault(execLocalStakingMsg)
	require.NoError(t, err)

	assert.Equal(t, 70_000_000, providerClient2.QueryVaultFreeBalance())

	// wait for initial packets to be transferred via IBC over
	validators, err = stakingtypes.NewQueryClient(consumerClient2.Chain.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myExtValidatorAddr2 := validators.Validators[0].OperatorAddress

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	fmt.Println("cross stake: additional liens on the same collateral")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient2.Contracts.ExternalStaking,
		providerClient2.Chain.Denom, 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr2))))
	_, err = providerClient2.MustExecVault(execMsg)
	require.NoError(t, err)

	// require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
	require.Equal(t, 20_000_000, providerClient2.QueryVaultFreeBalance()) // = 70 (free)  + 30 (local) - 80 (remote staked)

	// then
	fmt.Println("provider chain: query ext staking")
	qRsp = providerClient2.QueryExtStaking(setup.Query{
		"stake": {
			"user":      providerClient2.Chain.Address,
			"validator": myExtValidatorAddr2,
		},
	})
	assert.Equal(t, 80_000_000, setup.ParseHighLow(qRsp["stake"]).Low, qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])
}
