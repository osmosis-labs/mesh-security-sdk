package contract

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type (
	CustomQuery struct {
		VirtualStake *VirtualStakeQuery `json:"virtual_stake,omitempty"`
	}
	VirtualStakeQuery struct {
		BondStatus     *BondStatusQuery     `json:"bond_status,omitempty"`
		AllDelegations *AllDelegationsQuery `json:"all_delegations,omitempty"`
		SlashRatio     *struct{}            `json:"slash_ratio,omitempty"`
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
	AllDelegationsQuery struct {
		Contract    string `json:"contract"`
		MaxRetrieve uint16 `json:"max_retrieve"`
	}
	AllDelegationsResponse struct {
		Delegations []Delegation `json:"delegations"`
	}
	Delegation struct {
		Delegator string `json:"delegator"`
		Validator string `json:"validator"`
		Amount    string `json:"amount"`
	}
	SlashRatioResponse struct {
		SlashFractionDowntime   string `json:"slash_fraction_downtime"`
		SlashFractionDoubleSign string `json:"slash_fraction_double_sign"`
	}
)

func ConvertDelegationsToWasm(delegations []types.Delegation) (newDelegations []Delegation) {
	for _, del := range delegations {
		delegation := Delegation{
            Delegator: del.DelegatorAddress,
            Validator: del.ValidatorAddress,
            Amount:    del.Amount.String(),
        }
        newDelegations = append(newDelegations, delegation)
	}
	return
}