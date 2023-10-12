package e2e

import (
	"testing"
	"time"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/demo/app"
)

func TestValsetUpdate(t *testing.T) {
	// scenario:
	// given a provider chain P and a consumer chain C with staked tokens

	x := setupExampleChains(t)

	consumerCli, _, providerCli := setupMeshSecurity(t, x)
	_ = consumerCli
	_ = providerCli

	operatorKeys := secp256k1.GenPrivKey()
	x.ConsumerChain.Fund(sdk.AccAddress(operatorKeys.PubKey().Address()), sdkmath.NewInt(1_000_000_000))

	myVal := CreateNewValidator(t, operatorKeys, x.ConsumerChain)
	// then
	require.Len(t, x.ConsumerChain.Vals.Validators, 5)
	// and ibc packet pending
	require.Len(t, x.ConsumerChain.PendingSendPackets, 1)
	added := gjson.Get(string(x.ConsumerChain.PendingSendPackets[0].Data), "add_validators.#.valoper").Array()
	require.Len(t, added, 1)
	require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), added[0].String())

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	t.Log("remove from active set")
	// remove from active set
	val1 := x.ConsumerApp.StakingKeeper.Validator(x.ConsumerChain.GetContext(), sdk.ValAddress(x.ConsumerChain.Vals.Validators[3].Address))
	require.NotNil(t, val1)
	m := stakingtypes.NewMsgUndelegate(x.ConsumerChain.SenderAccount.GetAddress(), val1.GetOperator(), sdk.NewCoin(x.ConsumerDenom, val1.GetBondedTokens()))
	// when
	_, err := x.ConsumerChain.SendMsgs(m)
	require.NoError(t, err)
	x.ConsumerChain.NextBlock()
	// then
	// TODO: removals are currently ignored in the contract https://github.com/osmosis-labs/mesh-security/blob/b0c0bd483623d12703229772049063c06bb6e537/contracts/consumer/virtual-staking/src/contract.rs#L233
	//require.Len(t, x.ConsumerChain.PendingSendPackets, 1)
	//addr = gjson.Get(string(x.ConsumerChain.PendingSendPackets[0].Data), "removals").Array()[0].String()
	//require.Equal(t, val1.GetOperator().String(), addr)
	//
	//require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	t.Log("Update commission")
	// commission updated
	x.Coordinator.IncrementTimeBy(24 * time.Hour) // bump time to allow updates
	description := stakingtypes.NewDescription("my updated val", "", "", "", "")
	newRate := sdkmath.LegacyNewDec(1)
	updateMsg := stakingtypes.NewMsgEditValidator(sdk.ValAddress(operatorKeys.PubKey().Address()), description, &newRate, nil)
	// when
	_, err = x.ConsumerChain.SendNonDefaultSenderMsgs(operatorKeys, updateMsg)
	require.NoError(t, err)
	// then
	// TODO: updates are currently ignored in the contract: https://github.com/osmosis-labs/mesh-security/blob/b0c0bd483623d12703229772049063c06bb6e537/contracts/consumer/virtual-staking/src/contract.rs#L209
	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	t.Log("jail validator")
	jailValidator(t, sdk.ConsAddress(myVal.PrivKey.PubKey().Address()), x.Coordinator, x.ConsumerChain, x.ConsumerApp)
	// then
	require.Len(t, x.ConsumerChain.PendingSendPackets, 1)
	jailed := gjson.Get(string(x.ConsumerChain.PendingSendPackets[0].Data), "jail_validators.#.valoper").Array()
	require.Len(t, jailed, 1, string(x.ConsumerChain.PendingSendPackets[0].Data))
	require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), jailed[0].String())

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	t.Log("unjail validator")
	// bump time to expire jail time
	unjailValidator(t, sdk.ConsAddress(myVal.PrivKey.PubKey().Address()), operatorKeys, x.Coordinator, x.ConsumerChain, x.ConsumerApp)
	// then
	require.Len(t, x.ConsumerChain.PendingSendPackets, 1)
	unjailed := gjson.Get(string(x.ConsumerChain.PendingSendPackets[0].Data), "add_validators.#.valoper").Array()
	require.Len(t, unjailed, 1, string(x.ConsumerChain.PendingSendPackets[0].Data))
	require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), unjailed[0].String())

	require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

	t.Log("tombstone validator")

	e := &types.Equivocation{
		Height:           1,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(myVal.PrivKey.PubKey().Address().Bytes()).String(),
	}
	// when
	x.ConsumerApp.EvidenceKeeper.HandleEquivocationEvidence(x.ConsumerChain.GetContext(), e)
	// and
	require.NoError(t, err)
	x.ConsumerChain.NextBlock()
	// then
	// todo: expected ibc packet but none sent
	//require.Len(t, x.ConsumerChain.PendingSendPackets, 1)
	//tombstoned := gjson.Get(string(x.ConsumerChain.PendingSendPackets[0].Data), "tombstoned.#.valoper").Array()
	//require.Len(t, tombstoned, 1, string(x.ConsumerChain.PendingSendPackets[0].Data))
	//require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), tombstoned[0].String())
	//
	//require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
}

func jailValidator(t *testing.T, consAddr sdk.ConsAddress, coordinator *wasmibctesting.Coordinator, chain *wasmibctesting.TestChain, app *app.MeshApp) {
	ctx := chain.GetContext()
	signInfo, found := app.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
	require.True(t, found)
	// bump height to be > block window
	coordinator.CommitNBlocks(chain, 100)
	ctx = chain.GetContext()
	signInfo.MissedBlocksCounter = app.SlashingKeeper.MinSignedPerWindow(ctx)
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, signInfo)
	app.SlashingKeeper.HandleValidatorSignature(ctx, cryptotypes.Address(consAddr), 100, false)
	// when updates trigger
	chain.NextBlock()
}

func unjailValidator(t *testing.T, consAddr sdk.ConsAddress, operatorKeys *secp256k1.PrivKey, coordinator *wasmibctesting.Coordinator, chain *wasmibctesting.TestChain, app *app.MeshApp) {
	// move clock
	aa, ok := app.SlashingKeeper.GetValidatorSigningInfo(chain.GetContext(), consAddr)
	require.True(t, ok)
	coordinator.IncrementTimeBy(aa.JailedUntil.Sub(chain.GetContext().BlockTime()) + time.Minute)
	// when
	unjaiMsg := slashingtypes.NewMsgUnjail(sdk.ValAddress(operatorKeys.PubKey().Address()))
	_, err := chain.SendNonDefaultSenderMsgs(operatorKeys, unjaiMsg)
	require.NoError(t, err)
	// and
	chain.NextBlock()
}

func CreateNewValidator(t *testing.T, operatorKeys *secp256k1.PrivKey, chain *wasmibctesting.TestChain) mock.PV {
	privVal := mock.NewPV()
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction))
	description := stakingtypes.NewDescription("my new val", "", "", "", "")
	commissionRates := stakingtypes.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyNewDec(1), sdkmath.LegacyNewDec(1))
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(operatorKeys.PubKey().Address()), privVal.PrivKey.PubKey(), bondCoin, description, commissionRates, sdkmath.OneInt(),
	)
	require.NoError(t, err)
	_, err = chain.SendNonDefaultSenderMsgs(operatorKeys, createValidatorMsg)
	require.NoError(t, err)
	chain.NextBlock()
	// add to signers
	chain.Signers[privVal.PrivKey.PubKey().Address().String()] = privVal
	return privVal
}
