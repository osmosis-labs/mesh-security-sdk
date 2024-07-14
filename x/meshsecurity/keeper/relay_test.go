package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/stretchr/testify/require"

	// "github.com/cosmos/cosmos-sdk/types/address"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestDDD(t *testing.T) {
	ctx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myValidatorAddr := sdk.ValAddress(rand.Bytes(20))
	// myOtherValAddr              = sdk.ValAddress(rand.Bytes(address.Len))
	// myVStakingContractAddr      = sdk.AccAddress(rand.Bytes(address.Len))
	// myOtherVStakingContractAddr = sdk.AccAddress(rand.Bytes(address.Len))

	k.SetProviderChannel(ctx, "consumerCCVChannelID")
	k.SetParams(ctx, types.DefaultParams(sdk.DefaultBondDenom))

	k.ScheduleSlashed(ctx, myValidatorAddr, 46, 56, math.NewInt(66), math.LegacyMustNewDecFromStr("0.1"))
	k.ScheduleSlashed(ctx, myValidatorAddr, 45, 55, math.NewInt(66), math.LegacyMustNewDecFromStr("0.1"))

	k.SendPackets(ctx)
}
