package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	CustomMsg struct {
		ProviderMsg *ProviderMsg `json:"virtual_stake,omitempty"`
	}
	ProviderMsg struct {
		Bond   *BondMsg   `json:"bond,omitempty"`
		Unbond *UnbondMsg `json:"unbond,omitempty"`
		Unstake *UnstakeMsg `json:"unstake,omitempty"`
	}

	BondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Delegator string           `json:"delegator"`
	}
	UnbondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Delegator string           `json:"delegator"`
	}
	UnstakeMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
)
