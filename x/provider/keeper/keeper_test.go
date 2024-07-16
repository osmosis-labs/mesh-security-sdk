package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/mesh-security-sdk/x/provider/types"
	"github.com/stretchr/testify/require"
)

func TestStoreDelegator(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeperProvider

	d1 := types.Depositors{
		Address: "delegate",
		Tokens:  sdk.NewCoins([]sdk.Coin{sdk.NewCoin("osmo", sdk.NewInt(123))}...),
	}
	err := k.SetDepositors(pCtx, d1)
	require.NoError(t, err)

	d2, f := k.GetDepositors(pCtx, d1.Address)
	require.True(t, f)
	require.Equal(t, d1, d2)

	d2.Tokens = d2.Tokens.Add(sdk.NewCoin("juno", sdk.NewInt(10000)))
	d2.Tokens = d2.Tokens.Add(sdk.NewCoin("juno", sdk.NewInt(10000)))
	err = k.SetDepositors(pCtx, d2)
	require.NoError(t, err)

	d3, f := k.GetDepositors(pCtx, d1.Address)
	require.True(t, f)
	require.Equal(t, sdk.NewInt(20000), d3.Tokens.AmountOf("juno"))
}

func TestStoreIntermediary(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeperProvider

	coin := sdk.NewCoin("osmo", sdk.NewInt(123))
	e1 := types.Intermediary{
		ConsumerValidator: "validator",
		ChainId:           "osmo-1",
		ContractAddress:   "address",
		Jailed:            false,
		Tombstoned:        false,
		Status:            types.Bonded,
		Token:             &coin,
	}
	err := k.SetIntermediary(pCtx, e1)
	require.NoError(t, err)

	e2, f := k.GetIntermediary(pCtx, e1.Token.Denom)
	require.True(t, f)
	require.Equal(t, e1, e2)
}
