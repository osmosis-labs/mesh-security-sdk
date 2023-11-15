package types

// NewGenesisState constructor
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState default genesis state
func DefaultGenesisState(denom string) *GenesisState {
	return NewGenesisState(DefaultParams(denom))
}

// ValidateGenesis does basic validation on genesis state
func ValidateGenesis(gs *GenesisState) error {
	return gs.Params.ValidateBasic()
}
