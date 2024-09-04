package keeper

import (
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGetSetConsumerChainID(t *testing.T) {
	ctx, keepers := CreateDefaultTestInput(t)
	keeper := keepers.MeshProviderKeeper

	chainId1 := "consumer-1"
	chainId2 := "consumer-2"

	client1 := "client-1"
	client2 := "client-2"

	contractChain1 := sdk.AccAddress(rand.Bytes(32))
	contractChain2 := sdk.AccAddress(rand.Bytes(32))
	keeper.SetConsumerChainID(ctx, chainId1, contractChain1, client1)
	keeper.SetConsumerChainID(ctx, chainId2, contractChain2, client2)

	require.Equal(t, keeper.GetExternalStakingContractAccAddr(ctx, chainId1, client1), contractChain1)
	require.Equal(t, keeper.GetExternalStakingContractAccAddr(ctx, chainId2, client2), contractChain2)
}

func TestIterateProxyContractAddr(t *testing.T) {
	ctx, keepers := CreateDefaultTestInput(t)
	keeper := keepers.MeshProviderKeeper

	chainId := "consumer"

	contractsAddr := make([]sdk.AccAddress, 5)
	for i := 0; i < 5; i++ {
		clientId := "client-" + strconv.Itoa(i)
		contractsAddr[i] = sdk.AccAddress(rand.Bytes(32))
		keeper.SetConsumerChainID(ctx, chainId, contractsAddr[i], clientId)
	}

	expContractsAddr := []sdk.AccAddress{}
	keeper.IteratorExternalStakingContractAddr(ctx, chainId, func(contractAccAddr sdk.AccAddress) (stop bool) {
		expContractsAddr = append(expContractsAddr, contractAccAddr)
		return false
	})

	require.Equal(t, expContractsAddr, contractsAddr)
}
