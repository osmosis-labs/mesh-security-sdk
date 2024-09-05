package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestZeroMaxCapScenario1(t *testing.T) {
	// scenario:
	// given a provider chain P and a consumer chain C
	// some amount has been "cross stake" on chain C
	// a proposal is created to change max cap to zero
	// all delegations will be unstake in one epoch

	x := setupExampleChains(t)
	consumerCli, consumerContracts, providerCli := setupMeshSecurity(t, x)

	// the active set should be stored in the ext staking contract
	// and contain all active validator addresses
	qRsp := providerCli.QueryExtStaking(Query{"list_active_validators": {}})
	require.Len(t, qRsp["validators"], 4, qRsp)
	for _, v := range x.ConsumerChain.Vals.Validators {
		require.Contains(t, qRsp["validators"], sdk.ValAddress(v.Address).String())
	}

	// ----------------------------
	// ensure nothing staked by the virtual staking contract yet
	extValidator1 := sdk.ValAddress(x.ConsumerChain.Vals.Validators[1].Address)
	extValidator1Addr := extValidator1.String()

	extValidator2 := sdk.ValAddress(x.ConsumerChain.Vals.Validators[2].Address)
	extValidator2Addr := extValidator2.String()

	_, found := x.ConsumerApp.StakingKeeper.GetDelegation(x.ConsumerChain.GetContext(), consumerContracts.staking, extValidator1)
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

	// Cross Stake
	err := providerCli.ExecStakeRemote(extValidator1Addr, sdk.NewInt64Coin(x.ProviderDenom, 50_000_000))
	require.NoError(t, err)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
	require.Equal(t, 50_000_000, providerCli.QueryVaultFreeBalance())

	err = providerCli.ExecStakeRemote(extValidator2Addr, sdk.NewInt64Coin(x.ProviderDenom, 20_000_000))
	require.NoError(t, err)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
	require.Equal(t, 30_000_000, providerCli.QueryVaultFreeBalance())

	// consumer chain
	// ====================
	//
	// then delegated amount is not updated before the epoch
	consumerCli.assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(31_500_000)) // 70_000_000 /2 * (1 -0.1)

	// and the delegated amount is updated for the validator
	consumerCli.assertShare(extValidator1, math.LegacyMustNewDecFromStr("22.5")) // 50_000_000 /2 * (1 -0.1) / 1_000_000 # default sdk factor
	consumerCli.assertShare(extValidator2, math.LegacyNewDec(9))                 // 20_000_000 /2 * (1 -0.1) / 1_000_000 # default sdk factor

	// Zero max cap
	consumerCli.ExecSetMaxCap(sdk.NewInt64Coin(x.ConsumerDenom, 0))

	// the max cap limit is persisted
	rsp = consumerCli.QueryMaxCap()
	assert.Equal(t, sdk.NewInt64Coin(x.ConsumerDenom, 0), rsp.Cap)

	// when an epoch ends, the unstaking msgs is triggered
	consumerCli.ExecNewEpoch()

	// 2 internal unstake msg, 1 distribute batch msg
	require.Len(t, x.IbcPath.EndpointA.Chain.PendingSendPackets, 3)
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	consumerCli.assertTotalDelegated(math.ZeroInt())

	x.ProviderChain.NextBlock()
	providerCli.MustExecExtStaking(`{"withdraw_unbonded":{}}`)
	// When calculate inverted price, 50_000_000 will become 49_999_999, 20_000_000 will be 19_999_999
	assert.Equal(t, 99_999_998, providerCli.QueryVaultFreeBalance())
}
