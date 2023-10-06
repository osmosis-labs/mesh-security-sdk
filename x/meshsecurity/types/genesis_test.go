package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateGenesis(t *testing.T) {
	specs := map[string]struct {
		state  GenesisState
		expErr bool
	}{
		"default params": {
			state:  *DefaultGenesisState(sdk.DefaultBondDenom),
			expErr: false,
		},
		"custom param, should pass": {
			state: GenesisState{
				Params: Params{
					TotalContractsMaxCap: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(15_000_000_000)),
					EpochLength:          2_000,
					MaxGasEndBlocker:     600_000,
				},
			},
			expErr: false,
		},
		"custom small value param, should pass": {
			state: GenesisState{
				Params: Params{
					TotalContractsMaxCap: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000)),
					EpochLength:          20,
					MaxGasEndBlocker:     10_000,
				},
			},
			expErr: false,
		},
		"invalid epoch length, should fail": {
			state: GenesisState{
				Params: Params{
					TotalContractsMaxCap: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(15_000_000_000)),
					EpochLength:          0,
					MaxGasEndBlocker:     600_000,
				},
			},
			expErr: true,
		},
		"invalid max gas length, should fail": {
			state: GenesisState{
				Params: Params{
					TotalContractsMaxCap: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(15_000_000_000)),
					EpochLength:          10,
					MaxGasEndBlocker:     0,
				},
			},
			expErr: true,
		},
		"invalid max cap coin denom, should fail": {
			state: GenesisState{
				Params: Params{
					TotalContractsMaxCap: sdk.Coin{Denom: "invalid denom test", Amount: math.Int{}},
					EpochLength:          10,
					MaxGasEndBlocker:     0,
				},
			},
			expErr: true,
		},
		"invalid max cap coin amount, should fail": {
			state: GenesisState{
				Params: Params{
					TotalContractsMaxCap: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(-100)},
					EpochLength:          10,
					MaxGasEndBlocker:     0,
				},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			err := ValidateGenesis(&spec.state)
			if spec.expErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
