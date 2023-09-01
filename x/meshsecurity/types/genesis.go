package types

func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

func DefaultGenesisState(denom string) *GenesisState {
	return NewGenesisState(DefaultParams(denom))
}

func ValidateGenesis(gs *GenesisState) error {
	return gs.Params.ValidateBasic()
}
