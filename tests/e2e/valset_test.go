package e2e

import (
	"testing"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
)

func TestValsetUpdate(t *testing.T) {
	// scenario:
	// given a provider chain P and a consumer chain C with staked tokens

	x := setupExampleChains(t)

	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	valKey := ed25519.GenPrivKey()

	_, _, _ = setupMeshSecurity(t, x)
	x.ConsumerChain.Fund(addr1, sdkmath.NewInt(1_000_000_000))

	// scenario: new validator added to the active set
	bondCoin := sdk.NewCoin(x.ConsumerDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))
	description := stakingtypes.NewDescription("my new val", "", "", "", "")
	commissionRates := stakingtypes.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec())

	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr1), valKey.PubKey(), bondCoin, description, commissionRates, sdkmath.OneInt(),
	)
	require.NoError(t, err)
	_, err = x.ConsumerChain.SendNonDefaultSenderMsgs(priv1, createValidatorMsg)
	require.NoError(t, err)

	// when
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	// then

	// val removed
	// commission updated
	// unrelated update
	// slash
	// jail
}

type example struct {
	Coordinator      *wasmibctesting.Coordinator
	ConsumerChain    *wasmibctesting.TestChain
	ProviderChain    *wasmibctesting.TestChain
	ConsumerApp      *app.MeshApp
	IbcPath          *wasmibctesting.Path
	ProviderDenom    string
	ConsumerDenom    string
	MyProvChainActor string
}

func setupExampleChains(t *testing.T) example {
	coord := NewIBCCoordinator(t, 2)
	provChain := coord.GetChain(ibctesting.GetChainID(1))
	consChain := coord.GetChain(ibctesting.GetChainID(2))
	return example{
		Coordinator:      coord,
		ConsumerChain:    consChain,
		ProviderChain:    provChain,
		ConsumerApp:      consChain.App.(*app.MeshApp),
		IbcPath:          wasmibctesting.NewPath(consChain, provChain),
		ProviderDenom:    sdk.DefaultBondDenom,
		ConsumerDenom:    sdk.DefaultBondDenom,
		MyProvChainActor: provChain.SenderAccount.GetAddress().String(),
	}
}

func setupMeshSecurity(t *testing.T, x example) (*TestConsumerClient, ConsumerContract, *TestProviderClient) {
	x.Coordinator.SetupConnections(x.IbcPath)

	// setup contracts on both chains
	consumerCli := NewConsumerClient(t, x.ConsumerChain)
	consumerContracts := consumerCli.BootstrapContracts()
	converterPortID := wasmkeeper.PortIDForContract(consumerContracts.converter)
	// add some fees so that we can distribute something
	x.ConsumerChain.DefaultMsgFees = sdk.NewCoins(sdk.NewCoin(x.ConsumerDenom, sdkmath.NewInt(1_000_000)))

	providerCli := NewProviderClient(t, x.ProviderChain)
	providerContracts := providerCli.BootstrapContracts(x.IbcPath.EndpointA.ConnectionID, converterPortID)

	// setup ibc control path: consumer -> provider (direction matters)
	x.IbcPath.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
		PortID: wasmkeeper.PortIDForContract(providerContracts.externalStaking),
		Order:  channeltypes.UNORDERED,
	}
	x.IbcPath.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
		PortID: converterPortID,
		Order:  channeltypes.UNORDERED,
	}
	x.Coordinator.CreateChannels(x.IbcPath)

	// when ibc package is relayed
	require.NotEmpty(t, x.ConsumerChain.PendingSendPackets)
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	consumerCli.MustEnableVirtualStaking(sdk.NewInt64Coin(x.ConsumerDenom, 1_000_000_000))
	return consumerCli, consumerContracts, providerCli
}
