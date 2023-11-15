package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	CustomQuery struct {
		VirtualStake *VirtualStakeQuery `json:"virtual_stake,omitempty"`
	}
	VirtualStakeQuery struct {
		BondStatus *BondStatusQuery `json:"bond_status,omitempty"`
		SlashRatio *struct{}        `json:"slash_ratio,omitempty"`
	}
	BondStatusQuery struct {
		Contract string `json:"contract"`
	}
	BondStatusResponse struct {
		// MaxCap is the max cap limit
		MaxCap wasmvmtypes.Coin `json:"cap"`
		// Delegated is the used amount of the max cap
		Delegated wasmvmtypes.Coin `json:"delegated"`
	}
	SlashRatioResponse struct {
		SlashFractionDowntime   string `json:"slash_fraction_downtime"`
		SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
	}
)
