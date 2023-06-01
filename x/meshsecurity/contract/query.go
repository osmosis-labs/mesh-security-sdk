package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type CustomQuery struct {
	VirtualStake *VirtualStakeQuery `json:"virtual_stake,omitempty"`
}

type VirtualStakeQuery struct {
	BondStatus *BondStatusQuery `json:"max_cap,omitempty"`
}

type BondStatusQuery struct {
	Contract string `json:"contract"`
}

type BondStatusResponse struct {
	// MaxCap is the max cap limit
	MaxCap wasmvmtypes.Coin `json:"cap"`
	// Delegated is the used amount of the max cap
	Delegated wasmvmtypes.Coin `json:"delegated"`
}
