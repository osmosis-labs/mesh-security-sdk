package types

// DefaultParams returns default mesh-security parameters
func DefaultParams(denom string) Params {
	return Params{ // todo: revisit and set proper defaults
		InfractionTime: 0,
		TimeoutPeriod:  60,
	}
}

// ValidateBasic performs basic validation on mesh-security parameters.
func (p Params) ValidateBasic() error {
	if p.InfractionTime <= 0 {
		return ErrInvalid.Wrap("empty epoch length")
	}
	if p.TimeoutPeriod <= 0 {
		return ErrInvalid.Wrap("empty max gas end-blocker setting")
	}
	return nil
}
