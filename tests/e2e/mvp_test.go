package e2e

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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
	x := setupExampleChains(t)
	consumerCli, consumerContracts, providerCli := setupMeshSecurity(t, x)

	// then the active set should be stored in the ext staking contract
	// and contain all active validator addresses
	qRsp := providerCli.QueryExtStaking(Query{"list_active_validators": {}})
	require.Len(t, qRsp["validators"], 4, qRsp)
	for _, v := range x.ConsumerChain.Vals.Validators {
		require.Contains(t, qRsp["validators"], sdk.ValAddress(v.Address).String())
	}

	// ----------------------------
	// ensure nothing staked by the virtual staking contract yet
	myExtValidator := sdk.ValAddress(x.ConsumerChain.Vals.Validators[1].Address)
	myExtValidatorAddr := myExtValidator.String()
	_, found := x.ConsumerApp.StakingKeeper.GetDelegation(x.ConsumerChain.GetContext(), consumerContracts.staking, myExtValidator)
	require.False(t, found)

	// the max cap limit is persisted
	rsp := consumerCli.QueryMaxCap()
	assert.Equal(t, sdk.NewInt64Coin(x.ConsumerDenom, 1_000_000_000), rsp.Cap)

	// provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	execMsg := fmt.Sprintf(`{"bond":{"amount":{"denom":"%s", "amount":"100000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVault(execMsg)

	// then query contract state
	assert.Equal(t, 100_000_000, providerCli.QueryVaultFreeBalance())

	// Stake Locally - A user triggers a local staking action to a chosen validator. They then can manage their delegation and vote via the local staking contract.
	myLocalValidatorAddr := sdk.ValAddress(x.ProviderChain.Vals.Validators[0].Address).String()
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		x.ProviderDenom, 30_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	providerCli.MustExecVault(execLocalStakingMsg)

	assert.Equal(t, 70_000_000, providerCli.QueryVaultFreeBalance())

	// Failure mode of cross-stake... trying to stake to an unknown validator
	err := providerCli.ExecStakeRemote("BAD-VALIDATOR", sdk.NewInt64Coin(x.ProviderDenom, 80_000_000))
	require.Error(t, err)
	// no change to free balance
	assert.Equal(t, 70_000_000, providerCli.QueryVaultFreeBalance())

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	err = providerCli.ExecStakeRemote(myExtValidatorAddr, sdk.NewInt64Coin(x.ProviderDenom, 80_000_000))
	require.NoError(t, err)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
	require.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance()) // = 70 (free)  + 30 (local) - 80 (remote staked)

	// then
	qRsp = providerCli.QueryExtStaking(Query{
		"stake": {
			"user":      x.MyProvChainActor,
			"validator": myExtValidatorAddr,
		},
	})
	assert.Equal(t, 80_000_000, ParseHighLow(t, qRsp["stake"]).High, qRsp)
	assert.Empty(t, qRsp["pending_unbonds"])

	// consumer chain
	// ====================
	//
	// then delegated amount is not updated before the epoch
	consumerCli.assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(36_000_000)) // 80_000_000 /2 * (1 -0.1)

	// and the delegated amount is updated for the validator
	consumerCli.assertShare(myExtValidator, math.LegacyNewDec(36)) // 36_000_000 / 1_000_000 # default sdk factor

	// provider chain
	// ==============
	//
	// Cross Stake - A user undelegates
	execMsg = fmt.Sprintf(`{"unstake":{"validator":"%s", "amount":{"denom":"%s", "amount":"30000000"}}}`, myExtValidator.String(), x.ProviderDenom)
	providerCli.MustExecExtStaking(execMsg)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// then
	qRsp = providerCli.QueryExtStaking(Query{
		"stake": {
			"user":      x.MyProvChainActor,
			"validator": myExtValidatorAddr,
		},
	})
	require.Equal(t, 50_000_000, providerCli.QueryExtStakingAmount(x.MyProvChainActor, myExtValidatorAddr))
	require.Len(t, qRsp["pending_unbonds"], 1)
	unbonds := qRsp.Array("pending_unbonds")[0]
	assert.Equal(t, "30000000", unbonds["amount"], qRsp)

	// consumer chain
	// ====================

	consumerCli.ExecNewEpoch() // x
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(22_500_000))                  // (80_000_000 - 30_000_000) /2 * (1 -0.1)
	consumerCli.assertShare(myExtValidator, math.LegacyNewDecWithPrec(225, 1)) // 27_000_000 / 1_000_000 # default sdk factor

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
	// update system time with healthy ibc clients
	for x.Coordinator.CurrentTime.Before(releasedAt.Add(time.Minute)) {
		x.Coordinator.CurrentTime = x.Coordinator.CurrentTime.Add(96 * time.Hour)
		x.Coordinator.UpdateTime()
		require.NoError(t, x.IbcPath.EndpointA.UpdateClient())
		require.NoError(t, x.IbcPath.EndpointB.UpdateClient())
	}

	providerCli.MustExecExtStaking(`{"withdraw_unbonded":{}}`)
	assert.Equal(t, 50_000_000, providerCli.QueryVaultFreeBalance())

	// provider chain
	// ==============
	//
	// A user unstakes some free amount from the vault
	balanceBefore := x.ProviderChain.Balance(x.ProviderChain.SenderAccount.GetAddress(), "stake")
	providerCli.MustExecVault(fmt.Sprintf(`{"unbond":{"amount":{"denom":"%s", "amount": "30000000"}}}`, x.ProviderDenom))
	// then
	assert.Equal(t, 20_000_000, providerCli.QueryVaultFreeBalance())
	balanceAfter := x.ProviderChain.Balance(x.ProviderChain.SenderAccount.GetAddress(), x.ProviderDenom)
	assert.Equal(t, math.NewInt(30_000_000), balanceAfter.Sub(balanceBefore).Amount)

	// ----------------------
	// Claim rewards
	// ----------------------
	// provider chain
	// ==============

	// assert rewards exist
	rewardsResp, err := distrkeeper.NewQuerier(consumerCli.app.DistrKeeper).DelegationRewards(sdk.WrapSDKContext(x.ConsumerChain.GetContext()),
		&distrtypes.QueryDelegationRewardsRequest{ValidatorAddress: myExtValidator.String(), DelegatorAddress: consumerCli.contracts.staking.String()})
	require.NoError(t, err)
	require.False(t, rewardsResp.Rewards.AmountOf(x.ConsumerDenom).IsZero())

	// when claimed in epoch
	consumerCli.ExecNewEpoch()
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// then should be tracked on provider side
	qRsp = providerCli.QueryExtStaking(Query{"all_pending_rewards": {"user": x.MyProvChainActor}})
	myRewardsAmount := qRsp.Array("rewards")[0].To("rewards", "rewards")["amount"]
	v, err := strconv.Atoi(myRewardsAmount.(string))
	require.NoError(t, err)
	require.NotEmpty(t, v)

	// and when claimed
	targetAddr := sdk.AccAddress(rand.Bytes(address.Len))
	gotBalance := x.ConsumerChain.Balance(targetAddr, x.ConsumerDenom).Amount
	require.Equal(t, math.ZeroInt(), gotBalance)
	providerCli.MustExecExtStaking(fmt.Sprintf(`{"withdraw_rewards":{"validator": %q, "remote_recipient": %q}}`, myExtValidatorAddr, targetAddr.String()))
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// then should be on destination account
	gotBalance = x.ConsumerChain.Balance(targetAddr, x.ConsumerDenom).Amount
	require.Equal(t, myRewardsAmount, gotBalance.String())
	t.Log("got Rewards: " + gotBalance.String())
	for _, v := range x.ConsumerChain.Vals.Validators {
		del, found := consumerCli.app.StakingKeeper.GetDelegation(consumerCli.chain.GetContext(), consumerCli.contracts.staking, sdk.ValAddress(v.Address))
		if !found {
			val, ok := consumerCli.app.StakingKeeper.GetValidator(consumerCli.chain.GetContext(), sdk.ValAddress(v.Address))
			require.True(t, ok)
			t.Logf("skipping: %#v\n", val)
			continue
		}
		require.True(t, found)
		t.Logf("::%s", del.Shares)
	}
}
