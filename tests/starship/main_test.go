package starship

import (
	"context"
	"cosmossdk.io/math"
	"flag"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/osmosis-labs/mesh-security-sdk/tests/starship/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var (
	wasmContractPath    string
	wasmContractGZipped bool
	configFile          string
	providerChain       string
	consumerChain       string
)

func AssertTotalDelegated(t *testing.T, p *setup.ConsumerClient, expTotalDelegated math.Int) {
	delegations, err := stakingtypes.NewQueryClient(p.Chain.Client).DelegatorDelegations(context.Background(), &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: p.Contracts.Staking.String(),
		Pagination:    nil,
	})
	assert.NoError(t, err)
	if expTotalDelegated == sdk.ZeroInt() {
		assert.Nil(t, delegations.DelegationResponses)
		return
	}
	actualDelegated := sdk.NewCoin(p.Chain.Denom, sdk.ZeroInt())
	for _, delegation := range delegations.DelegationResponses {
		actualDelegated = actualDelegated.Add(delegation.Balance)
	}
	assert.Equal(t, sdk.NewCoin(p.Chain.Denom, expTotalDelegated), actualDelegated)
}

func AssertShare(t *testing.T, p *setup.ConsumerClient, val string, exp math.LegacyDec) {
	fmt.Printf("consumer chain: staking contract %v\n", p.Contracts.Staking.String())
	delegations, err := stakingtypes.NewQueryClient(p.Chain.Client).DelegatorDelegations(context.Background(), &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: p.Contracts.Staking.String(),
		Pagination:    nil,
	})
	require.NoError(t, err)
	for _, delegation := range delegations.DelegationResponses {
		if delegation.Delegation.ValidatorAddress == val {
			assert.Equal(t, exp, delegation.Delegation.Shares)
		}
	}
}

func TestMain(m *testing.M) {
	flag.StringVar(&wasmContractPath, "contracts-path", "testdata", "Set path to dir with gzipped wasm contracts")
	flag.BoolVar(&wasmContractGZipped, "gzipped", true, "Use `.gz` file ending when set")
	flag.StringVar(&configFile, "config", "configs/devnet.yaml", "starship config file for the infra")
	flag.StringVar(&providerChain, "provider-chain", "mesh-osmosis-1", "provider chain name, from config file")
	flag.StringVar(&consumerChain, "consumer-chain", "mesh-juno-1", "consumer chain name, from config file")
	flag.Parse()

	os.Exit(m.Run())
}
