package e2e

import (
	"cosmossdk.io/math"
	"encoding/base64"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSlashing(t *testing.T) {
	// Slashing scenario 1:
	// https://github.com/osmosis-labs/mesh-security/blob/main/docs/ibc/Slashing.md#scenario-1-slashed-delegator-has-free-collateral-on-the-vault
	//
	// - We use a different slash ratio here (50%?).
	// - We use millions instead of unit tokens.
	x := setupExampleChains(t)
	//consumerCli, consumerContracts, providerCli := setupMeshSecurity(t, x)
	consumerCli, _, providerCli := setupMeshSecurity(t, x)

	// Provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	execMsg := `{"bond":{}}`
	providerCli.MustExecVault(execMsg, sdk.NewInt64Coin(x.ProviderDenom, 200_000_000))

	// Stake Locally - A user triggers a local staking action to a chosen validator.
	myLocalValidatorAddr := sdk.ValAddress(x.ProviderChain.Vals.Validators[0].Address).String()
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		x.ProviderDenom, 190_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	providerCli.MustExecVault(execLocalStakingMsg)

	assert.Equal(t, 10_000_000, providerCli.QueryVaultFreeBalance())

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	myExtValidator1 := sdk.ValAddress(x.ConsumerChain.Vals.Validators[1].Address)
	myExtValidator1Addr := myExtValidator1.String()
	err := providerCli.ExecStakeRemote(myExtValidator1Addr, sdk.NewInt64Coin(x.ProviderDenom, 100_000_000))
	require.NoError(t, err)
	myExtValidator2 := sdk.ValAddress(x.ConsumerChain.Vals.Validators[2].Address)
	myExtValidator2Addr := myExtValidator2.String()
	err = providerCli.ExecStakeRemote(myExtValidator2Addr, sdk.NewInt64Coin(x.ProviderDenom, 50_000_000))
	require.NoError(t, err)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// TODO: Check max lien (190)
	// TODO: Check slashable amount (34)
	require.Equal(t, 10_000_000, providerCli.QueryVaultFreeBalance()) // 200 - max(34, 190) = 200 - 190 = 10

	// Consumer chain
	// ====================
	//
	// then delegated amount is not updated before the epoch
	consumerCli.assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(67_500_000)) // 150_000_000 / 2 * (1 - 0.1)

	// and the delegated amount is updated for the validators
	consumerCli.assertShare(myExtValidator1, math.LegacyMustNewDecFromStr("45"))   // 100_000_000 / 2 * (1 - 0.1) / 1_000_000 # default sdk factor
	consumerCli.assertShare(myExtValidator2, math.LegacyMustNewDecFromStr("22.5")) // 50_000_000 / 2 * (1 - 0.1) / 1_000_000 # default sdk factor

	// Validator 1 on the Consumer chain is jailed
	myExtValidator1ConsAddr := sdk.ConsAddress(x.ConsumerChain.Vals.Validators[1].PubKey.Address())
	jailValidator(t, myExtValidator1ConsAddr, x.Coordinator, x.ConsumerChain, x.ConsumerApp)

	// Then delegated amount is not updated before the epoch
	consumerCli.assertTotalDelegated(math.NewInt(67_500_000))

	// When an epoch ends, the slashing is triggered
	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(63_000_000)) // (150_000_000 - 10_000_000) / 2 * (1 - 0.1)

	// and the delegated amount is updated for the slashed validator
	consumerCli.assertShare(myExtValidator1, math.LegacyMustNewDecFromStr("40.5")) // 90_000_000 / 2 * (1 - 0.1) / 1_000_000 # default sdk factor
	consumerCli.assertShare(myExtValidator2, math.LegacyMustNewDecFromStr("22.5")) // 50_000_000 / 2 * (1 - 0.1) / 1_000_000 # default sdk factor

	// TODO: Check new collateral (190)
	// TODO: Check new max lien (190)
	// TODO: Check new slashable amount (33)
	// New free collateral
	require.Equal(t, 0, providerCli.QueryVaultFreeBalance()) // 190 - max(33, 190) = 190 - 190 = 0
}
