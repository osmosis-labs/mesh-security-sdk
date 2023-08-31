package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestInitGenesis(t *testing.T) {
	specs := map[string]struct {
		state  types.GenesisState
		expErr bool
	}{
		"custom param, should pass": {
			state: types.GenesisState{
				Params: types.Params{
					TotalContractsMaxCap: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(15_000_000_000)),
					EpochLength:          2_000,
					MaxGasEndBlocker:     600_000,
				},
			},
			expErr: false,
		},
		"custom small value param, should pass": {
			state: types.GenesisState{
				Params: types.Params{
					TotalContractsMaxCap: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
					EpochLength:          20,
					MaxGasEndBlocker:     10_000,
				},
			},
			expErr: false,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			pCtx, keepers := CreateDefaultTestInput(t)
			k := keepers.MeshKeeper

			k.InitGenesis(pCtx, spec.state)

			p := k.GetParams(pCtx)
			assert.Equal(t, spec.state.Params.MaxGasEndBlocker, p.MaxGasEndBlocker)
			assert.Equal(t, spec.state.Params.EpochLength, p.EpochLength)
			assert.Equal(t, spec.state.Params.TotalContractsMaxCap, p.TotalContractsMaxCap)
		})
	}
}

func TestExportGenesis(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	params := types.DefaultParams(sdk.DefaultBondDenom)

	err := k.SetParams(pCtx, params)
	require.NoError(t, err)

	exported := k.ExportGenesis(pCtx)
	assert.Equal(t, params.MaxGasEndBlocker, exported.Params.MaxGasEndBlocker)
	assert.Equal(t, params.EpochLength, exported.Params.EpochLength)
	assert.Equal(t, params.TotalContractsMaxCap, exported.Params.TotalContractsMaxCap)
}
