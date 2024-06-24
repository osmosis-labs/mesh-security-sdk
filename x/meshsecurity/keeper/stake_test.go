package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestDelegateVirtualStake(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper

	myContractAddr := sdk.AccAddress(rand.Bytes(32))
	myDelegatorAddr := sdk.AccAddress(rand.Bytes(32))
	vAddrs := add3Validators(t, pCtx, keepers.StakingKeeper)
	myValAddr := vAddrs[0]
	totalBondTokenSupply := func(ctx sdk.Context) sdk.Coin {
		rsp, err := keepers.BankKeeper.SupplyOf(sdk.WrapSDKContext(ctx), &banktypes.QuerySupplyOfRequest{Denom: sdk.DefaultBondDenom})
		require.NoError(t, err)
		return rsp.Amount
	}
	startSupply := totalBondTokenSupply(pCtx)
	specs := map[string]struct {
		limit      sdk.Coin
		usedLimit  sdk.Coin
		delegation sdk.Coin
		valAddr    sdk.ValAddress
		expErr     bool
		expNewUsed sdk.Coin
	}{
		"all good - full limit": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)),
			valAddr:    myValAddr,
			expNewUsed: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10)),
		},
		"all good - used before": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 2),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt()),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt()),
			valAddr:    myValAddr,
			expNewUsed: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2)),
		},
		"exceed limit - nothing used before": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 1),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2)),
			valAddr:    myValAddr,
			expErr:     true,
		},
		"exceed limit - used before": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 2),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt()),
			delegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2)),
			valAddr:    myValAddr,
			expErr:     true,
		},
		"invalid amount": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			delegation: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(-1)},
			valAddr:    myValAddr,
			expErr:     true,
		},
		"nil amount": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			delegation: sdk.Coin{Denom: sdk.DefaultBondDenom},
			valAddr:    myValAddr,
			expErr:     true,
		},
		"non staking denom rejected": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			delegation: sdk.NewCoin("ALX", math.OneInt()),
			valAddr:    myValAddr,
			expErr:     true,
		},
		"unknown validator": {
			limit:      sdk.NewInt64Coin(sdk.DefaultBondDenom, 10),
			usedLimit:  sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
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
			k.setTotalDelegated(ctx, myContractAddr, spec.usedLimit)
			expSupplyDiff := spec.delegation

			// when
			gotShares, gotErr := k.Delegate(ctx, myContractAddr, myDelegatorAddr, spec.valAddr, spec.delegation)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				expSupplyDiff = sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt())
			} else {
				require.NoError(t, gotErr)
				require.False(t, gotShares.IsZero())
				// and usage updated
				assert.Equal(t, spec.expNewUsed, k.GetTotalDelegated(ctx, myContractAddr))
			}
			// and delegation was persisted
			_, ok := keepers.StakingKeeper.GetDelegation(ctx, myContractAddr, myValAddr)
			require.Equal(t, !spec.expErr, ok)

			currentSupply := totalBondTokenSupply(ctx)
			// total supply increased
			assert.Equal(t, expSupplyDiff.String(), currentSupply.Sub(startSupply).String())
			// and supply offset decreased (negative)
			assert.Equal(t, expSupplyDiff.Amount.Neg().String(), captBankKeeper.Offset[sdk.DefaultBondDenom].String())
		})
	}
}

func TestInstantUndelegateVirtualStake(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper

	myContractAddr := sdk.AccAddress(rand.Bytes(32))
	myDelegatorAddr := sdk.AccAddress(rand.Bytes(32))
	vAddrs := add3Validators(t, pCtx, keepers.StakingKeeper)
	myValAddr := vAddrs[0]
	totalBondTokenSupply := func(ctx sdk.Context) sdk.Coin {
		rsp, err := keepers.BankKeeper.SupplyOf(sdk.WrapSDKContext(ctx), &banktypes.QuerySupplyOfRequest{Denom: sdk.DefaultBondDenom})
		require.NoError(t, err)
		return rsp.Amount
	}
	initialDelegation := sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000_000)
	totalCapLimit := initialDelegation.AddAmount(math.OneInt())
	require.NoError(t, k.SetMaxCapLimit(pCtx, myContractAddr, totalCapLimit))
	_, err := k.Delegate(pCtx, myContractAddr, myDelegatorAddr, myValAddr, initialDelegation)
	require.NoError(t, err)

	startSupply := totalBondTokenSupply(pCtx)
	specs := map[string]struct {
		undelegation sdk.Coin
		valAddr      sdk.ValAddress
		expErr       bool
		expNoop      bool
		expNewUsed   sdk.Coin
	}{
		"partial undelegate": {
			undelegation: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100_000_000)),
			valAddr:      myValAddr,
			expNewUsed:   sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(900_000_000)),
		},
		"full undelegate": {
			undelegation: initialDelegation,
			valAddr:      myValAddr,
			expNewUsed:   sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
		},
		"non delegated validator": {
			undelegation: sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt()),
			valAddr:      vAddrs[1],
			expNewUsed:   sdk.NewCoin(sdk.DefaultBondDenom, initialDelegation.Amount),
			expNoop:      true,
		},
		"exceed total delegate": {
			undelegation: totalCapLimit.AddAmount(math.OneInt()),
			valAddr:      myValAddr,
			expErr:       true,
		},
		"exceed staked amount": {
			undelegation: initialDelegation.AddAmount(math.OneInt()),
			valAddr:      myValAddr,
			expErr:       true,
		},
		"zero amount undelegate": {
			undelegation: sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			valAddr:      myValAddr,
			expErr:       true,
		},
		"nil Amount": {
			undelegation: sdk.Coin{Denom: sdk.DefaultBondDenom},
			valAddr:      myValAddr,
			expErr:       true,
		},
		"negative Amount": {
			undelegation: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(-1)},
			valAddr:      myValAddr,
			expErr:       true,
		},
		"non staking denom rejected": {
			undelegation: sdk.NewCoin("ALX", math.OneInt()),
			valAddr:      myValAddr,
			expErr:       true,
		},
		"unknown validator": {
			undelegation: sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt()),
			valAddr:      rand.Bytes(20),
			expErr:       true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ := pCtx.CacheContext()
			captBankKeeper := NewCaptureOffsetBankKeeper(keepers.BankKeeper)
			k.bank = captBankKeeper

			// when
			gotErr := k.Undelegate(ctx, myContractAddr, myDelegatorAddr, spec.valAddr, spec.undelegation)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			// and usage updated
			assert.Equal(t, spec.expNewUsed.String(), k.GetTotalDelegated(ctx, myContractAddr).String())

			currentSupply := totalBondTokenSupply(ctx)
			expSupplyDiff := spec.undelegation.Amount.Neg()
			if spec.expNoop {
				expSupplyDiff = math.ZeroInt()
			}
			// total supply decreased
			assert.Equal(t, expSupplyDiff.String(), currentSupply.Amount.Sub(startSupply.Amount).String())
			// and supply offset increased (negative)
			assert.Equal(t, expSupplyDiff.Neg().String(), captBankKeeper.Offset[sdk.DefaultBondDenom].String())
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

func add3Validators(t *testing.T, pCtx sdk.Context, stakingKeeper *stakingkeeper.Keeper) []sdk.ValAddress {
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
