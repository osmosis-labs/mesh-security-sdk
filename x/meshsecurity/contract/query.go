package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type CustomQuery struct {
	VirtualStake *VirtualStakeQuery `json:"virtual_stake,omitempty"`
}

type VirtualStakeQuery struct {
	MaxCap *struct{} `json:"max_cap,omitempty"`
}

type MaxCapResponse struct {
	MaxCap wasmvmtypes.Coin `json:"cap"`
}
