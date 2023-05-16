package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type CustomMsg struct {
	VirtualStake *VirtualStakeMsg `json:"virtual_stake,omitempty"`
}

type VirtualStakeMsg struct {
	Bond   *BondMsg   `json:"bond,omitempty"`
	Unbond *UnbondMsg `json:"unbond,omitempty"`
}

type BondMsg struct {
	Amount    wasmvmtypes.Coin `json:"amount"`
	Validator string           `json:"validator"`
}

type UnbondMsg struct {
	Amount    wasmvmtypes.Coin `json:"amount"`
	Validator string           `json:"validator"`
}

type CustomQuery struct {
	VirtualStake *VirtualStakeQuery `json:"virtual_stake,omitempty"`
}

type VirtualStakeQuery struct {
	MaxCap *struct{} `json:"max_cap,omitempty"`
}
