package keeper

import (
	"testing"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/rand"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegateVirtualStake(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper

	myContractAddr := sdk.AccAddress(rand.Bytes(32))
	vAddrs := addValidators(t, pCtx, keepers.StakingKeeper)
	myValAddr := vAddrs[0]
	totalBondTokenSupply := func(ctx sdk.Context) sdk.Coin {
		rsp, err := keepers.BankKeeper.SupplyOf(sdk.WrapSDKContext(ctx), &banktypes.QuerySupplyOfRequest{Denom: sdk.DefaultBondDenom})
		require.NoError(t, err)
		return rsp.Amount
	}
	startSupply := totalBondTokenSupply(pCtx)
	specs := map[string]struct {
		limit      sdk.Coin
		delegation sdk.Coin
		valAddr    sdk.ValAddress
		expErr     bool
	}{
		"all good": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)),
			valAddr:    myValAddr,
		},
		"exceed limit": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2)),
			valAddr:    myValAddr,
			expErr:     true,
		},
		"invalid amount": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			delegation: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(-1)},
			valAddr:    myValAddr,
			expErr:     true,
		},
		"non staking denom rejected": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			delegation: sdk.NewCoin("ALX", math.OneInt()),
			valAddr:    myValAddr,
			expErr:     true,
		},
		"unknown validator": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)),
			valAddr:    rand.Bytes(20),
			expErr:     true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			captBankKeeper := NewCaptureOffsetBankKeeper(keepers.BankKeeper)
			k.bank = captBankKeeper
			require.NoError(t, k.SetMaxCapLimit(ctx, myContractAddr, spec.limit))

			expSupplyInc := spec.delegation

			// when
			gotShares, gotErr := k.Delegate(ctx, myContractAddr, spec.valAddr, spec.delegation)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				expSupplyInc = sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt())
			} else {
				require.NoError(t, gotErr)
				require.False(t, gotShares.IsZero())
			}
			// and delegation was persisted
			_, ok := keepers.StakingKeeper.GetDelegation(ctx, myContractAddr, myValAddr)
			require.Equal(t, !spec.expErr, ok)

			// and supply increased
			currentSupply := totalBondTokenSupply(ctx)
			assert.Equal(t, expSupplyInc.String(), currentSupply.Sub(startSupply).String())
			assert.Equal(t, expSupplyInc.Amount.Neg().String(), captBankKeeper.Offset[sdk.DefaultBondDenom].String())
		})
	}
}

var _ types.XBankKeeper = &CaptureOffsetBankKeeper{}

type CaptureOffsetBankKeeper struct {
	types.SDKBankKeeper
	Offset map[string]math.Int
}

func NewCaptureOffsetBankKeeper(b types.SDKBankKeeper) *CaptureOffsetBankKeeper {
	k := &CaptureOffsetBankKeeper{SDKBankKeeper: b, Offset: make(map[string]math.Int, 1)}
	k.Offset[sdk.DefaultBondDenom] = math.ZeroInt()
	return k
}

func (c *CaptureOffsetBankKeeper) AddSupplyOffset(_ sdk.Context, denom string, offsetAmount math.Int) {
	old, ok := c.Offset[denom]
	if !ok {
		old = math.ZeroInt()
	}
	c.Offset[denom] = old.Add(offsetAmount)
}

func addValidators(t *testing.T, pCtx sdk.Context, stakingKeeper *stakingkeeper.Keeper) []sdk.ValAddress {
	accNum := 3
	valAddrs := simtestutil.ConvertAddrsToValAddrs(simtestutil.CreateIncrementalAccounts(accNum))
	PKs := simtestutil.CreateTestPubKeys(accNum)

	// construct the validators
	amts := []math.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	validators := make([]stakingtypes.Validator, accNum)
	for i, amt := range amts {
		validators[i] = stakingtestutil.NewValidator(t, valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		validators[i] = stakingkeeper.TestingUpdateValidator(stakingKeeper, pCtx, validators[i], true)
	}
	return valAddrs
}
