package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns default mesh-security parameters
func DefaultParams(denom string) Params {
	return Params{ // todo: revisit and set proper defaults
		TotalContractsMaxCap: sdk.NewCoin(denom, math.NewInt(10_000_000_000)),
		EpochLength:          1_000,
		MaxGasEndBlocker:     500_000,
	}
}

// ValidateBasic performs basic validation on mesh-security parameters.
func (p Params) ValidateBasic() error {
	if err := p.TotalContractsMaxCap.Validate(); err != nil {
		return errorsmod.Wrap(err, "total contracts max cap")
	}
	if p.EpochLength == 0 {
		return ErrInvalid.Wrap("empty epoch length")
	}
	if p.MaxGasEndBlocker == 0 {
		return ErrInvalid.Wrap("empty max gas end-blocker setting")
	}
	return nil
}
