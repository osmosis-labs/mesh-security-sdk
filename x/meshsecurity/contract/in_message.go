package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	CustomMsg struct {
		VirtualStake *VirtualStakeMsg `json:"virtual_stake,omitempty"`
	}
	VirtualStakeMsg struct {
		Bond   *BondMsg   `json:"bond,omitempty"`
		Unbond *UnbondMsg `json:"unbond,omitempty"`
	}
	BondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
	UnbondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
)
