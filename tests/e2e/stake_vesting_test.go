package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestStakeVestingTokenScenario1(t *testing.T) {
	// Slashing scenario:
	// Permanent lock vesting account should be able to bond in vault contract
	// Unbond the amount, balance should keep being locked
	x := setupExampleChains(t)
	_, _, providerCli := setupMeshSecurity(t, x)

	// Generate vesting address
	vestingPrivKey := secp256k1.GenPrivKey()
	vestingBacc := authtypes.NewBaseAccount(vestingPrivKey.PubKey().Address().Bytes(), vestingPrivKey.PubKey(), uint64(19), 0)

	vestingBalance := sdk.NewCoin(x.ProviderDenom, sdk.NewInt(100000000))
	providerCli.MustCreatePermanentLockedAccount(vestingBacc.GetAddress().String(), vestingBalance)

	balance := x.ProviderChain.Balance(vestingBacc.GetAddress(), x.ProviderDenom)
	assert.Equal(t, balance, vestingBalance)

	// Exec contract by vesting account
	execMsg := fmt.Sprintf(`{"bond":{"amount":{"denom":"%s", "amount":"100000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVaultWithSigner(vestingPrivKey, vestingBacc, execMsg)

	balance = x.ProviderChain.Balance(providerCli.Contracts.Vault, x.ProviderDenom)
	assert.Equal(t, balance, vestingBalance)

	assert.Equal(t, 100_000_000, providerCli.QuerySpecificAddressVaultBalance(vestingBacc.GetAddress().String()))

	// Try to exec msg bond again, should fail
	err := providerCli.ExecVaultWithSigner(vestingPrivKey, vestingBacc, execMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispatch: submessages: failed to delegate")

	// Make sure vault balance doesn't change
	assert.Equal(t, 100_000_000, providerCli.QuerySpecificAddressVaultBalance(vestingBacc.GetAddress().String()))

	// Unbond vault
	execMsg = fmt.Sprintf(`{"unbond":{"amount":{"denom":"%s", "amount": "30000000"}}}`, x.ProviderDenom)
	providerCli.MustExecVaultWithSigner(vestingPrivKey, vestingBacc, execMsg)

	vestingBalance = sdk.NewCoin(x.ProviderDenom, sdk.NewInt(30000000))

	balance = x.ProviderChain.Balance(vestingBacc.GetAddress(), x.ProviderDenom)
	assert.Equal(t, balance, vestingBalance)

	// Vesting account is still locked
	err = providerCli.BankSendWithSigner(vestingPrivKey, vestingBacc, x.ProviderChain.SenderAccount.GetAddress().String(), vestingBalance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")
}
