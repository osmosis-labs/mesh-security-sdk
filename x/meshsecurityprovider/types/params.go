package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// DefaultParams returns default mesh-security parameters
func DefaultParams(denom string) Params {
	return Params{ // todo: revisit and set proper defaults
		TimeoutPeriod:        60,
		VaultContractAddress: "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s4hmalr",
	}
}

// ValidateBasic performs basic validation on mesh-security parameters.
func (p Params) ValidateBasic() error {
	if p.TimeoutPeriod <= 0 {
		return ErrInvalid.Wrap("empty max gas end-blocker setting")
	}
	if p.VaultContractAddress == "" {
		return ErrInvalid.Wrap("vault contract address cannot be empty")
	}
	return nil
}

func (p Params) GetVaultContractAddress() sdk.AccAddress {
	vaultContractAddress, err := sdk.AccAddressFromBech32(p.VaultContractAddress)
	if err != nil {
		panic("params error: VaultContractAddress")
	}
	return vaultContractAddress
}
