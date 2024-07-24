package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	SudoMsg struct {
		VaultSudoMsg *StakeMsg `json:"vault,omitempty"`
	}

	StakeMsg struct {
		StakeRemote `json:"stake_remote"`
		StakeLocal  `json:"stake_local"`
	}

	StakeRemote struct {
		Contract string           `json:"contract"`
		Amount   wasmvmtypes.Coin `json:"amount"`
	}
	StakeLocal struct {
		Amount wasmvmtypes.Coin `json:"amount"`
	}
)

type (
	CustomMsg struct {
		VaultCustomMsg *VaultCustomMsg `json:"vault,omitempty"`
	}
	VaultCustomMsg struct {
		Bond   *BondMsg   `json:"bond,omitempty"`
		Unbond *UnbondMsg `json:"unbond,omitempty"`
	}
	BondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
	UnbondMsg struct {
		Amount wasmvmtypes.Coin `json:"amount"`
	}
)
