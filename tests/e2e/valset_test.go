package e2e

import (
	"testing"
	"time"

	wasmibctesting "github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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

func TestValsetTransitions(t *testing.T) {
	operatorKeys := secp256k1.GenPrivKey()
	setupVal := func(t *testing.T, x example) mock.PV {
		myVal := CreateNewValidator(t, operatorKeys, x.ConsumerChain, 2)
		require.Len(t, x.ConsumerChain.Vals.Validators, 5)
		return myVal
	}
	specs := map[string]struct {
		setup         func(t *testing.T, x example) mock.PV
		doTransition  func(t *testing.T, val mock.PV, x example)
		assertPackets func(t *testing.T, packets []channeltypes.Packet)
	}{
		"new validator to active": {
			setup: func(t *testing.T, x example) mock.PV {
				return mock.PV{}
			},
			doTransition: func(t *testing.T, _ mock.PV, x example) {
				CreateNewValidator(t, operatorKeys, x.ConsumerChain, 2)
				require.Len(t, x.ConsumerChain.Vals.Validators, 5)
			},
			assertPackets: func(t *testing.T, packets []channeltypes.Packet) {
				require.Len(t, packets, 1)
				added := gjson.Get(string(packets[0].Data), "valset_update.additions.#.valoper").Array()
				require.Len(t, added, 1)
				require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), added[0].String())
			},
		},
		"active to tombstone": {
			setup: setupVal,
			doTransition: func(t *testing.T, val mock.PV, x example) {
				e := &types.Equivocation{
					Height:           1,
					Power:            100,
					Time:             time.Now().UTC(),
					ConsensusAddress: sdk.ConsAddress(val.PrivKey.PubKey().Address().Bytes()).String(),
				}
				// when
				x.ConsumerApp.EvidenceKeeper.HandleEquivocationEvidence(x.ConsumerChain.GetContext(), e)
			},
			assertPackets: func(t *testing.T, packets []channeltypes.Packet) {
				require.Len(t, packets, 1)
				tombstoned := gjson.Get(string(packets[0].Data), "valset_update.tombstoned").Array()
				require.Len(t, tombstoned, 1, string(packets[0].Data))
				require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), tombstoned[0].String())
			},
		},
		"active to jailed": {
			setup: setupVal,
			doTransition: func(t *testing.T, val mock.PV, x example) {
				jailValidator(t, sdk.ConsAddress(val.PrivKey.PubKey().Address()), x.Coordinator, x.ConsumerChain, x.ConsumerApp)
			},
			assertPackets: func(t *testing.T, packets []channeltypes.Packet) {
				require.Len(t, packets, 1)
				jailed := gjson.Get(string(packets[0].Data), "valset_update.jailed").Array()
				require.Len(t, jailed, 1, string(packets[0].Data))
				require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), jailed[0].String())
			},
		},
		"jailed to active": {
			setup: func(t *testing.T, x example) mock.PV {
				val := CreateNewValidator(t, operatorKeys, x.ConsumerChain, 200)
				jailValidator(t, sdk.ConsAddress(val.PrivKey.PubKey().Address()), x.Coordinator, x.ConsumerChain, x.ConsumerApp)
				x.ConsumerChain.NextBlock()
				require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
				return val
			},
			doTransition: func(t *testing.T, val mock.PV, x example) {
				unjailValidator(t, sdk.ConsAddress(val.PrivKey.PubKey().Address()), operatorKeys, x.Coordinator, x.ConsumerChain, x.ConsumerApp)
			},
			assertPackets: func(t *testing.T, packets []channeltypes.Packet) {
				for _, v := range packets {
					t.Log("\n\npacket: " + string(v.Data))
				}
				require.Len(t, packets, 1)
				unjailed := gjson.Get(string(packets[0].Data), "valset_update.additions.#.valoper").Array()
				require.Len(t, unjailed, 1, string(packets[0].Data))
				require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), unjailed[0].String())
			},
		},
		"jailed to remove": {
			setup: func(t *testing.T, x example) mock.PV {
				val := CreateNewValidator(t, operatorKeys, x.ConsumerChain, 200)
				t.Log("jail validator")
				jailValidator(t, sdk.ConsAddress(val.PrivKey.PubKey().Address()), x.Coordinator, x.ConsumerChain, x.ConsumerApp)

				t.Log("Add new validator")
				otherOperator := secp256k1.GenPrivKey()
				x.ConsumerChain.Fund(sdk.AccAddress(otherOperator.PubKey().Address()), sdkmath.NewInt(1_000_000_000))
				CreateNewValidator(t, otherOperator, x.ConsumerChain, 200) // add a new val to fill the slot
				x.ConsumerChain.NextBlock()
				require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

				// undelegate
				t.Log("undelegating")
				undelegate(t, operatorKeys, sdk.NewInt(999_999), x)
				return val
			},
			doTransition: func(t *testing.T, val mock.PV, x example) {
				t.Log("unjail")
				unjailValidator(t, sdk.ConsAddress(val.PrivKey.PubKey().Address()), operatorKeys, x.Coordinator, x.ConsumerChain, x.ConsumerApp)
			},
			assertPackets: func(t *testing.T, packets []channeltypes.Packet) {
				require.Len(t, packets, 1)
				unjailed := gjson.Get(string(packets[0].Data), "valset_update.unjailed").Array()
				require.Len(t, unjailed, 1, string(packets[0].Data))
				require.Equal(t, sdk.ValAddress(operatorKeys.PubKey().Address()).String(), unjailed[0].String())
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			x := setupExampleChains(t)
			setupMeshSecurity(t, x)
			// set max validators
			ctx, sk := x.ConsumerChain.GetContext(), x.ConsumerApp.StakingKeeper
			params := sk.GetParams(ctx)
			params.MaxValidators = 5
			require.NoError(t, sk.SetParams(ctx, params))

			x.ConsumerChain.Fund(sdk.AccAddress(operatorKeys.PubKey().Address()), sdkmath.NewInt(1_000_000_000))
			myVal := spec.setup(t, x)
			require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))

			// when
			spec.doTransition(t, myVal, x)
			x.ConsumerChain.NextBlock()
			// then
			spec.assertPackets(t, x.ConsumerChain.PendingSendPackets)
			require.NoError(t, x.Coordinator.RelayAndAckPendingPackets(x.IbcPath))
		})
	}
}

func undelegate(t *testing.T, operatorKeys *secp256k1.PrivKey, amount sdkmath.Int, x example) {
	ctx, sk := x.ConsumerChain.GetContext(), x.ConsumerApp.StakingKeeper
	val1 := sk.Validator(ctx, sdk.ValAddress(operatorKeys.PubKey().Address()))
	require.NotNil(t, val1)
	msg := stakingtypes.NewMsgUndelegate(sdk.AccAddress(operatorKeys.PubKey().Address()), val1.GetOperator(), sdk.NewCoin(x.ConsumerDenom, amount))
	_, err := x.ConsumerChain.SendNonDefaultSenderMsgs(operatorKeys, msg)
	require.NoError(t, err)
	x.ConsumerChain.NextBlock()
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
	power := app.StakingKeeper.GetLastValidatorPower(ctx, sdk.ValAddress(consAddr))
	app.SlashingKeeper.HandleValidatorSignature(ctx, cryptotypes.Address(consAddr), power, false)
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

func CreateNewValidator(t *testing.T, operatorKeys *secp256k1.PrivKey, chain *wasmibctesting.TestChain, power int64) mock.PV {
	privVal := mock.NewPV()
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction))
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
