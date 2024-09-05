package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	CustomMsg struct {
		VirtualStake *VirtualStakeMsg `json:"virtual_stake,omitempty"`
	}
	VirtualStakeMsg struct {
		Bond                    *BondMsg                    `json:"bond,omitempty"`
		Unbond                  *UnbondMsg                  `json:"unbond,omitempty"`
		UpdateDelegation        *UpdateDelegationMsg        `json:"update_delegation,omitempty"`
		DeleteAllScheduledTasks *DeleteAllScheduledTasksMsg `json:"delete_all_scheduled_tasks,omitempty"`
	}
	BondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
	UnbondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
	UpdateDelegationMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		IsDeduct  bool             `json:"is_deduct"`
		Delegator string           `json:"delegator"`
		Validator string           `json:"validator"`
	}

	DeleteAllScheduledTasksMsg struct{}
)
