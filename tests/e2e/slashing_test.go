package e2e

import (
	"encoding/base64"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlashingScenario1(t *testing.T) {
	// Slashing scenario 1:
	// https://github.com/osmosis-labs/mesh-security/blob/main/docs/ibc/Slashing.md#scenario-1-slashed-delegator-has-free-collateral-on-the-vault
	//
	// - We use millions instead of unit tokens.
	x := setupExampleChains(t)
	consumerCli, _, providerCli := setupMeshSecurity(t, x)

	// Provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	execMsg := fmt.Sprintf(`{"bond":{"amount":{"denom":"%s", "amount":"200000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVault(execMsg)

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

	// Check collateral
	require.Equal(t, 200_000_000, providerCli.QueryVaultBalance())
	// Check max lien
	require.Equal(t, 190_000_000, providerCli.QueryMaxLien())
	// Check slashable amount
	require.Equal(t, 68_000_000, providerCli.QuerySlashableAmount())
	// Check free collateral
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

	ctx := x.ConsumerChain.GetContext()
	validator1, found := x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, found)
	require.False(t, validator1.IsJailed())
	// Off by 1_000_000, because of validator self bond on setup
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(46_000_000))
	validator2, found := x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator2)
	require.True(t, found)
	require.False(t, validator2.IsJailed())
	// Off by 1_000_000, because of validator self bond on setup
	require.Equal(t, validator2.GetTokens(), sdk.NewInt(23_500_000))

	// Validator 1 on the Consumer chain is jailed
	myExtValidator1ConsAddr := sdk.ConsAddress(x.ConsumerChain.Vals.Validators[1].PubKey.Address())
	jailValidator(t, myExtValidator1ConsAddr, x.Coordinator, x.ConsumerChain, x.ConsumerApp)

	x.ConsumerChain.NextBlock()

	// Assert that the validator's stake has been slashed
	// and that the validator has been jailed
	validator1, _ = x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, validator1.IsJailed())
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(41_400_000)) // 10% slash

	// Relay IBC packets to the Provider chain
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// Next block on the Provider chain
	x.ProviderChain.NextBlock()

	// Check new collateral
	require.Equal(t, 190_000_001, providerCli.QueryVaultBalance())
	// Check new max lien
	require.Equal(t, 190_000_000, providerCli.QueryMaxLien())
	// Check new slashable amount
	require.Equal(t, 66_000_001, providerCli.QuerySlashableAmount())
	// Check new free collateral
	require.Equal(t, 1, providerCli.QueryVaultFreeBalance()) // 190 - max(33, 190) = 190 - 190 = 0
}

func TestSlashingScenario2(t *testing.T) {
	// Slashing scenario 2:
	// https://github.com/osmosis-labs/mesh-security/blob/main/docs/ibc/Slashing.md#scenario-2-slashed-delegator-has-no-free-collateral-on-the-vault
	//
	// - We use millions instead of unit tokens.
	x := setupExampleChains(t)
	consumerCli, _, providerCli := setupMeshSecurity(t, x)

	// Provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	execMsg := fmt.Sprintf(`{"bond":{"amount":{"denom":"%s", "amount":"200000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVault(execMsg)

	// Stake Locally - A user triggers a local staking action to a chosen validator.
	myLocalValidatorAddr := sdk.ValAddress(x.ProviderChain.Vals.Validators[0].Address).String()
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		x.ProviderDenom, 200_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	providerCli.MustExecVault(execLocalStakingMsg)

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	myExtValidator1 := sdk.ValAddress(x.ConsumerChain.Vals.Validators[1].Address)
	myExtValidator1Addr := myExtValidator1.String()
	err := providerCli.ExecStakeRemote(myExtValidator1Addr, sdk.NewInt64Coin(x.ProviderDenom, 200_000_000))
	require.NoError(t, err)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// Check collateral
	require.Equal(t, 200_000_000, providerCli.QueryVaultBalance())
	// Check max lien
	require.Equal(t, 200_000_000, providerCli.QueryMaxLien())
	// Check slashable amount
	require.Equal(t, 80_000_000, providerCli.QuerySlashableAmount())
	// Check free collateral
	require.Equal(t, 0, providerCli.QueryVaultFreeBalance()) // 200 - max(40, 200) = 200 - 200 = 0

	// Consumer chain
	// ====================
	//
	// then delegated amount is not updated before the epoch
	consumerCli.assertTotalDelegated(math.ZeroInt()) // ensure nothing cross staked yet

	// when an epoch ends, the delegation rebalance is triggered
	consumerCli.ExecNewEpoch()

	// then the total delegated amount is updated
	consumerCli.assertTotalDelegated(math.NewInt(90_000_000)) // 200_000_000 / 2 * (1 - 0.1)

	// and the delegated amount is updated for the validators
	consumerCli.assertShare(myExtValidator1, math.LegacyMustNewDecFromStr("90")) // 200_000_000 / 2 * (1 - 0.1) / 1_000_000 # default sdk factor

	ctx := x.ConsumerChain.GetContext()
	validator1, found := x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, found)
	require.False(t, validator1.IsJailed())
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(91_000_000))

	// Validator 1 on the Consumer chain is jailed
	myExtValidator1ConsAddr := sdk.ConsAddress(x.ConsumerChain.Vals.Validators[1].PubKey.Address())
	jailValidator(t, myExtValidator1ConsAddr, x.Coordinator, x.ConsumerChain, x.ConsumerApp)

	x.ConsumerChain.NextBlock()

	// Assert that the validator's stake has been slashed
	// and that the validator has been jailed
	validator1, _ = x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, validator1.IsJailed())
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(81_900_000)) // 10% slash

	// Relay IBC packets to the Provider chain
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// Next block on the Provider chain
	x.ProviderChain.NextBlock()

	// Check new collateral
	require.Equal(t, 180_000_001, providerCli.QueryVaultBalance())
	// Check new max lien
	require.Equal(t, 180_000_001, providerCli.QueryMaxLien())
	// Check new slashable amount
	require.Equal(t, 72_000_002, providerCli.QuerySlashableAmount())
	// Check new free collateral
	require.Equal(t, 0, providerCli.QueryVaultFreeBalance()) // 190 - max(36, 190) = 190 - 190 = 0
}

func TestSlashingScenario3(t *testing.T) {
	// Slashing scenario 3:
	// https://github.com/osmosis-labs/mesh-security/blob/main/docs/ibc/Slashing.md#scenario-3-slashed-delegator-has-some-free-collateral-on-the-vault
	//
	// - We use millions instead of unit tokens.
	x := setupExampleChains(t)
	consumerCli, _, providerCli := setupMeshSecurity(t, x)

	// Provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	execMsg := fmt.Sprintf(`{"bond":{"amount":{"denom":"%s", "amount":"200000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVault(execMsg)

	// Stake Locally - A user triggers a local staking action to a chosen validator.
	myLocalValidatorAddr := sdk.ValAddress(x.ProviderChain.Vals.Validators[0].Address).String()
	execLocalStakingMsg := fmt.Sprintf(`{"stake_local":{"amount": {"denom":%q, "amount":"%d"}, "msg":%q}}`,
		x.ProviderDenom, 190_000_000,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"validator": "%s"}`, myLocalValidatorAddr))))
	providerCli.MustExecVault(execLocalStakingMsg)

	// Cross Stake - A user pulls out additional liens on the same collateral "cross staking" it on different chains.
	myExtValidator1 := sdk.ValAddress(x.ConsumerChain.Vals.Validators[1].Address)
	myExtValidator1Addr := myExtValidator1.String()
	err := providerCli.ExecStakeRemote(myExtValidator1Addr, sdk.NewInt64Coin(x.ProviderDenom, 150_000_000))
	require.NoError(t, err)

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// Check collateral
	require.Equal(t, 200_000_000, providerCli.QueryVaultBalance())
	// Check max lien
	require.Equal(t, 190_000_000, providerCli.QueryMaxLien())
	// Check slashable amount
	require.Equal(t, 68_000_000, providerCli.QuerySlashableAmount())
	// Check free collateral
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
	consumerCli.assertShare(myExtValidator1, math.LegacyMustNewDecFromStr("67.5")) // 150_000_000 / 2 * (1 - 0.1) / 1_000_000 # default sdk factor

	ctx := x.ConsumerChain.GetContext()
	validator1, found := x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, found)
	require.False(t, validator1.IsJailed())
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(68_500_000))

	// Validator 1 on the Consumer chain is jailed
	myExtValidator1ConsAddr := sdk.ConsAddress(x.ConsumerChain.Vals.Validators[1].PubKey.Address())
	jailValidator(t, myExtValidator1ConsAddr, x.Coordinator, x.ConsumerChain, x.ConsumerApp)

	x.ConsumerChain.NextBlock()

	// Assert that the validator's stake has been slashed
	// and that the validator has been jailed
	validator1, _ = x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, validator1.IsJailed())
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(61_700_000)) // 10% slash (plus 50_000 rounding)

	// Relay IBC packets to the Provider chain
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// Next block on the Provider chain
	x.ProviderChain.NextBlock()

	// Check new collateral
	require.Equal(t, 185_109_490, providerCli.QueryVaultBalance())
	// Check new max lien
	require.Equal(t, 185_109_490, providerCli.QueryMaxLien())
	// Check new slashable amount
	require.Equal(t, 64_043_796, providerCli.QuerySlashableAmount())
	// Check new free collateral
	require.Equal(t, 0, providerCli.QueryVaultFreeBalance()) // 185 - max(32, 185) = 185 - 185 = 0
}

func TestValidatorTombstone(t *testing.T) {
	x := setupExampleChains(t)
	consumerCli, _, providerCli := setupMeshSecurity(t, x)

	// Provider chain
	// ==============
	// Deposit - A user deposits the vault denom to provide some collateral to their account
	execMsg := fmt.Sprintf(`{"bond":{"amount":{"denom":"%s", "amount":"200000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVault(execMsg)

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

	// Check collateral
	require.Equal(t, 200_000_000, providerCli.QueryVaultBalance())
	// Check max lien
	require.Equal(t, 190_000_000, providerCli.QueryMaxLien())
	// Check slashable amount
	require.Equal(t, 68_000_000, providerCli.QuerySlashableAmount())
	// Check free collateral
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

	ctx := x.ConsumerChain.GetContext()
	validator1, found := x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, found)
	require.False(t, validator1.IsJailed())
	// Off by 1_000_000, because of validator self bond on setup
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(46_000_000))
	validator2, found := x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator2)
	require.True(t, found)
	require.False(t, validator2.IsJailed())
	// Off by 1_000_000, because of validator self bond on setup
	require.Equal(t, validator2.GetTokens(), sdk.NewInt(23_500_000))

	// Validator 1 on the Consumer chain is tombstoned
	myExtValidator1ConsAddr := sdk.ConsAddress(x.ConsumerChain.Vals.Validators[1].PubKey.Address())
	tombstoneValidator(t, myExtValidator1ConsAddr, myExtValidator1, x.ConsumerChain, x.ConsumerApp)

	x.ConsumerChain.NextBlock()

	// Assert that the validator's stake has been slashed
	// and that the validator has been jailed
	validator1, _ = x.ConsumerApp.StakingKeeper.GetValidator(ctx, myExtValidator1)
	require.True(t, validator1.IsJailed())
	require.Equal(t, validator1.GetTokens(), sdk.NewInt(36_000_000)) // 20% slash
	validator1SigningInfo, _ := x.ConsumerApp.SlashingKeeper.GetValidatorSigningInfo(ctx, myExtValidator1ConsAddr)
	require.True(t, validator1SigningInfo.Tombstoned)

	// Relay IBC packets to the Provider chain
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// Next block on the Provider chain
	x.ProviderChain.NextBlock()

	// Check new collateral
	require.Equal(t, 178_260_869, providerCli.QueryVaultBalance())
	// Check new max lien
	require.Equal(t, 178_260_869, providerCli.QueryMaxLien())
	// Check new slashable amount
	require.Equal(t, 61_304_348, providerCli.QuerySlashableAmount())
	// Check new free collateral
	require.Equal(t, 0, providerCli.QueryVaultFreeBalance())

	consumerCli.ExecNewEpoch()
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
	delegation, _ := x.ConsumerApp.StakingKeeper.GetDelegation(ctx, consumerCli.contracts.staking, myExtValidator1)
	// Nearly unbond all token
	require.Equal(t, delegation.Shares, sdk.MustNewDecFromStr("0.000000388888888889"))
}
