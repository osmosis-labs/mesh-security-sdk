package types

// Parameter store keys.
var (
	KeyParamField = []byte("TODO: CHANGE ME")
)

func NewParams(vaultAddress string) Params {
	return Params{
		VaultAddress: vaultAddress,
	}
}

// DefaultParams are the default meshsecurityprovider module parameters.
func DefaultParams() Params {
	return Params{}
}

// Validate validates params.
func (p Params) Validate() error {
	return nil
}
