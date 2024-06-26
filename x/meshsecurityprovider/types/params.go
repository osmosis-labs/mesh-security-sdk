package types

import (
	"sigs.k8s.io/yaml"
)

// Parameter store keys.
var (
	KeyParamField = []byte("TODO: CHANGE ME")
)

func NewParams(vaultAddress string) Params {
	return Params{
		VaultAddress: vaultAddress,
	}
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// DefaultParams are the default meshsecurityprovider module parameters.
func DefaultParams() Params {
	return Params{}
}

// Validate validates params.
func (p Params) Validate() error {
	return nil
}
