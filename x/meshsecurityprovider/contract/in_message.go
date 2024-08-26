package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	CustomMsg struct {
		Provider *ProviderMsg `json:"provider,omitempty"`
	}
	ProviderMsg struct {
		Bond   *BondMsg   `json:"bond,omitempty"`
		Unbond *UnbondMsg `json:"unbond,omitempty"`
	}
	BondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Delegator string           `json:"delegator"`
	}
	UnbondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Delegator string           `json:"delegator"`
	}
)