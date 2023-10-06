package types

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateMsgSetVirtualStakingMaxCap(t *testing.T) {
	var (
		validAddr      = sdk.AccAddress(rand.Bytes(20)).String()
		validContrAddr = sdk.AccAddress(rand.Bytes(32)).String()
		validCoin      = sdk.NewInt64Coin("ALX", 1)
	)
	specs := map[string]struct {
		src    MsgSetVirtualStakingMaxCap
		expErr bool
	}{
		"all valid": {
			src: MsgSetVirtualStakingMaxCap{
				Authority: validAddr,
				Contract:  validContrAddr,
				MaxCap:    validCoin,
			},
		},
		"empty amount": {
			src: MsgSetVirtualStakingMaxCap{
				Authority: validAddr,
				Contract:  validContrAddr,
				MaxCap:    sdk.NewInt64Coin("ALX", 0),
			},
		},
		"empty authority": {
			src: MsgSetVirtualStakingMaxCap{
				Contract: validContrAddr,
				MaxCap:   validCoin,
			},
			expErr: true,
		},
		"invalid authority addr": {
			src: MsgSetVirtualStakingMaxCap{
				Authority: "invalid-addr",
				Contract:  validContrAddr,
				MaxCap:    validCoin,
			},
			expErr: true,
		},
		"invalid contract addr": {
			src: MsgSetVirtualStakingMaxCap{
				Authority: validAddr,
				Contract:  "invalid-addr",
				MaxCap:    validCoin,
			},
			expErr: true,
		},
		"empty cap": {
			src: MsgSetVirtualStakingMaxCap{
				Authority: validAddr,
				Contract:  validContrAddr,
			},
			expErr: true,
		},
		"invalid cap coin": {
			src: MsgSetVirtualStakingMaxCap{
				Authority: validAddr,
				Contract:  validContrAddr,
				MaxCap:    sdk.Coin{Amount: math.NewInt(1)},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
		})
	}
}
