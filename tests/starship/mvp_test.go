package starship

import (
	"context"
	"cosmossdk.io/math"
	"encoding/base64"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/mesh-security-sdk/tests/starship/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// create clients for provider and consumer
	providerClient, consumerClient, err := setup.MeshSecurity(providerChain, consumerChain, configFile, wasmContractPath, wasmContractGZipped)
	require.NoError(t, err)

	var (
		consumerChain = consumerClient.Chain
		providerChain = providerClient.Chain
	)

	// wait for initial packets to be transfered via IBC over
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
			qRsp = providerClient.QueryExtStaking(setup.Query{"list_remote_validators": {}})
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
		DelegatorAddr: consumerClient.Contracts.Staking.String(),
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
		providerClient.Contracts.ExternalStaking.String(),
		providerChain.Denom, 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(`{"validator": "BAD-VALIDATOR"}`)))
	err = providerClient.MustFailExecVault(execMsg)
	require.Error(t, err)
	// // no change to free balance
	assert.Equal(t, 70_000_000, providerClient.QueryVaultFreeBalance())

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	fmt.Println("cross stake: additional liens on the same collateral")
	execMsg = fmt.Sprintf(`{"stake_remote":{"contract":"%s", "amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		providerClient.Contracts.ExternalStaking.String(),
		providerChain.Denom, 80_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myExtValidatorAddr))))
	_, err = providerClient.MustExecVault(execMsg)
	require.NoError(t, err)

	//require.NoError(t, coord.RelayAndAckPendingPackets(ibcPath))
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

	//providerCli.MustExecExtStaking(`{"withdraw_unbonded":{}}`)
	//assert.Equal(t, 50_000_000, providerCli.QueryVaultFreeBalance())
	//
	//// provider chain
	//// ==============
	////
	//// A user unstakes some free amount from the vault
	//balanceBefore, err := banktypes.NewQueryClient(providerClient.Client).Balance(context.Background(), &banktypes.QueryBalanceRequest{
	//	Address: providerClient.Address,
	//	Denom:   providerChain.Denom,
	//})
	//require.NoError(t, err)
	//providerCli.MustExecVault(`{"unbond":{"amount":{"denom":"stake", "amount": "30000000"}}}`)
	//// then
	//assert.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance())
	//balanceAfter, err := banktypes.NewQueryClient(providerClient.Client).Balance(context.Background(), &banktypes.QueryBalanceRequest{
	//	Address: providerClient.Address,
	//	Denom:   providerChain.Denom,
	//})
	//require.NoError(t, err)
	//assert.Equal(t, math.NewInt(30_000_000), balanceAfter.Balance.Sub(*balanceBefore.Balance).Amount)
}
