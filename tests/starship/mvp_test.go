package starship

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
		time.Second,
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
	assert.Equal(t, "80000000", qRsp["stake"], qRsp)
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
		time.Second,
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
	assert.Equal(t, "80000000", setup.ParseHighLow(qRsp["stake"]).Low, qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])
}

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

	// create clients for provider and consumer
	providerClient, consumerClient, err := setup.MeshSecurity(providerChain, consumerChain, configFile, wasmContractPath, wasmContractGZipped)
	require.NoError(t, err)

	var (
		consumerChain = consumerClient.Chain
		providerChain = providerClient.Chain
	)

	// wait for initial packets to be transferred via IBC over
	validators, err := stakingtypes.NewQueryClient(consumerChain.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myExtValidatorAddr := validators.Validators[0].OperatorAddress

	// stake tokens from the client address
	err = consumerChain.StakeTokens(myExtValidatorAddr, 5000000, consumerChain.Denom)
	require.NoError(t, err)

	// then the active set should be stored in the ext staking contract
	// and contain all active validator addresses
	qRsp := map[string]any{}

	require.Eventuallyf(t,
		func() bool {
			qRsp = providerClient.QueryExtStaking(setup.Query{"list_active_validators": {}})
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
		DelegatorAddr: consumerClient.Contracts.Staking,
	}
	delegations, err := stakingtypes.NewQueryClient(consumerChain.Client).DelegatorValidators(context.Background(), query)
	require.NoError(t, err)
	require.Empty(t, delegations.Validators)

	// then the max cap limit is persisted
	rsp := consumerClient.QueryMaxCap()
	require.Equal(t, sdk.NewInt64Coin(consumerChain.Denom, 1_000_000_000), rsp.Cap)

	// provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	fmt.Println("provider chain: deposit vault denom to provide some collateral to account")
	execMsg := `{"bond":{}}`
	vault, err := providerClient.MustExecVault(execMsg, sdk.NewInt64Coin(providerChain.Denom, 100_000_000))
	require.NoError(t, err)
	require.NotEmpty(t, vault)

	// then query contract state
	assert.Equal(t, 100_000_000, providerClient.QueryVaultFreeBalance())

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	fmt.Println("provider chain: stake locally")
	providerValidators, err := stakingtypes.NewQueryClient(providerChain.Client).Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Status: "BOND_STATUS_BONDED",
	})
	require.NoError(t, err)
	myLocalValidatorAddr := providerValidators.Validators[0].OperatorAddress
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerChain.Denom, 30_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	_, err = providerClient.MustExecVault(execLocalStakingMsg)
	require.NoError(t, err)

	assert.Equal(t, 70_000_000, providerClient.QueryVaultFreeBalance())

	// // Failure mode of cross-stake... trying to stake to an unknown validator
	fmt.Println("provider chain: failure care, trying to stake to unknown validator")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient.Contracts.ExternalStaking,
		providerChain.Denom, 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(`{"validator": "BAD-VALIDATOR"}`)))
	err = providerClient.MustFailExecVault(execMsg)
	require.Error(t, err)
	// // no change to free balance
	assert.Equal(t, 70_000_000, providerClient.QueryVaultFreeBalance())

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	fmt.Println("cross stake: additional liens on the same collateral")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient.Contracts.ExternalStaking,
		providerChain.Denom, 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr))))
	_, err = providerClient.MustExecVault(execMsg)
	require.NoError(t, err)

	// require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
	require.Equal(t, 20_000_000, providerClient.QueryVaultFreeBalance()) // = 70 (free)  + 30 (local) - 80 (remote staked)

	// then
	fmt.Println("provider chain: query ext staking")
	qRsp = providerClient.QueryExtStaking(setup.Query{
		"stake": {
			"user":      providerChain.Address,
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
	AssertTotalDelegated(t, consumerClient, math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	consumerClient.ExecNewEpoch()

	// then the total delegated amount is updated
	AssertTotalDelegated(t, consumerClient, math.NewInt(36_000_000)) // 80_000_000 /2 * (1 -0.1)

	// and the delegated amount is updated for the validator
	AssertShare(t, consumerClient, myExtValidatorAddr, math.LegacyNewDec(36_000_000)) // 36_000_000 / 1_000_000 # default sdk factor

	// provider chain
	// ==============
	// then
	qRsp = providerClient.QueryExtStaking(setup.Query{
		"stake": {
			"user":      providerClient.Chain.Address,
			"validator": myExtValidatorAddr,
		},
	})
	assert.Equal(t, "80000000", qRsp["stake"], qRsp)

	// Cross Stake - A user undelegates
	fmt.Println("provider chain: cross stake user undelegates")
	execMsg = fmt.Sprintf(`{"unstake":{"validator":"%s", "amount":{"denom":"%s", "amount":"30000000"}}}`, myExtValidatorAddr, providerChain.Denom)
	_, err = providerClient.MustExecExtStaking(execMsg)
	require.NoError(t, err)

	// then
	qRsp = providerClient.QueryExtStaking(setup.Query{
		"stake": {
			"user":      providerClient.Chain.Address,
			"validator": myExtValidatorAddr,
		},
	})
	require.Equal(t, "50000000", qRsp["stake"], qRsp)
	require.Len(t, qRsp["pending_unbonds"], 1)
	unbonds := qRsp["pending_unbonds"].([]any)[0].(map[string]any)
	assert.Equal(t, "30000000", unbonds["amount"], qRsp)

	// consumer chain
	// ====================

	consumerClient.ExecNewEpoch()

	// then the total delegated amount is updated
	AssertTotalDelegated(t, consumerClient, math.NewInt(22_500_000))                  // (80_000_000 - 30_000_000) /2 * (1 -0.1)
	AssertShare(t, consumerClient, myExtValidatorAddr, math.LegacyNewDec(22_500_000)) // 27_000_000 / 1_000_000 # default sdk factor

	// provider chain
	// ==============
	//
	// A user withdraws the undelegated amount

	require.Equal(t, 20_000_000, providerClient.QueryVaultFreeBalance())

	releaseData := unbonds["release_at"].(string)
	require.NotEmpty(t, releaseData)
	at, err := strconv.Atoi(releaseData)
	require.NoError(t, err)
	releasedAt := time.Unix(0, int64(at)).UTC()
	fmt.Printf("unbonding at: %v, time to: %v\n", releasedAt, releasedAt.Add(time.Minute).Sub(time.Now()))

	fmt.Printf("sleeping for: %v\n", releasedAt.Add(time.Minute).Sub(time.Now()))
	time.Sleep(releasedAt.Add(time.Minute).Sub(time.Now()))

	staking, err := providerClient.MustExecExtStaking(`{"withdraw_unbonded":{}}`)
	require.NoError(t, err)
	require.NotEmpty(t, staking)
	assert.Equal(t, 50_000_000, providerClient.QueryVaultFreeBalance())

	// provider chain
	// ==============
	//
	// A user unstakes some free amount from the vault
	balanceBefore, err := banktypes.NewQueryClient(providerClient.Chain.Client).Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: providerClient.Chain.Address,
		Denom:   providerClient.Chain.Denom,
	})
	require.NoError(t, err)
	execVault, err := providerClient.MustExecVault(fmt.Sprintf(`{"unbond":{"amount":{"denom":"%s", "amount": "30000000"}}}`, providerClient.Chain.Denom))
	require.NoError(t, err)
	require.NotEmpty(t, execVault)
	// then
	assert.Equal(t, 20_000_000, providerClient.QueryVaultFreeBalance())
	balanceAfter, err := banktypes.NewQueryClient(providerClient.Chain.Client).Balance(context.Background(), &banktypes.QueryBalanceRequest{
		Address: providerClient.Chain.Address,
		Denom:   providerClient.Chain.Denom,
	})
	require.NoError(t, err)
	assert.Less(t, math.NewInt(100_000), balanceAfter.Balance.Sub(*balanceBefore.Balance).Amount.Sub(sdk.NewInt(30_000_000)))
}
