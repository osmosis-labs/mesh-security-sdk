package e2e

import (
	// "encoding/base64"
	// "fmt"
	// "strconv"
	"testing"
	// "time"

	// "github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/cosmos/cosmos-sdk/types/address"
	// distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	// distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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
	myExtValidator := sdk.ValAddress(x.ConsumerChain.Vals.Validators[1].Address)
	// myExtValidatorAddr := myExtValidator.String()
	_, found := x.ConsumerApp.StakingKeeper.GetDelegation(x.ConsumerChain.GetContext(), consumerContracts.staking, myExtValidator)
	require.False(t, found)

	// the max cap limit is persisted
	rsp := consumerCli.QueryMaxCap()
	assert.Equal(t, sdk.NewInt64Coin(x.ConsumerDenom, 1_000_000_000), rsp.Cap)
}
