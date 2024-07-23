package bindings

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type (
	CustomMsg struct {
		Bond     *BondMsg     `json:"bond,omitempty"`
		Unbond   *UnbondMsg   `json:"unbond,omitempty"`
		Deposit  *DepositMsg  `json:"deposit,,omitempty"`
		Withdraw *WithdrawMsg `json:"withdraw,,omitempty"`
		Unstake  *UnstakeMsg  `json:"unstake,,omitempty"`
	}

	BondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
	UnbondMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
	DepositMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Delegator string           `json:"delegator"`
	}
	WithdrawMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Delegator string           `json:"delegator"`
	}
	UnstakeMsg struct {
		Amount    wasmvmtypes.Coin `json:"amount"`
		Validator string           `json:"validator"`
	}
)

type (
	SudoMsg struct {
		HandleEpoch  *struct{}     `json:"handle_epoch,omitempty"`
		ValsetUpdate *ValsetUpdate `json:"handle_valset_update,omitempty"`
	}

	// Validator alias to wasmVM type
	Validator = wasmvmtypes.Validator
	// ValidatorAddr alias for the Bech32 address string of sdk.ValAddress
	ValidatorAddr = string

	ValidatorSlash struct {
		ValidatorAddr    string `json:"address"`
		Height           int64  `json:"height"`
		Time             int64  `json:"time"`
		InfractionHeight int64  `json:"infraction_height"`
		InfractionTime   int64  `json:"infraction_time"`
		Power            int64  `json:"power"`
		SlashAmount      string `json:"slash_amount"`
		SlashRatio       string `json:"slash_ratio"`
	}

	// ValsetUpdate updates to the active validator set
	ValsetUpdate struct {
		Additions  []Validator      `json:"additions"`
		Removals   []ValidatorAddr  `json:"removals"`
		Updated    []Validator      `json:"updated"`
		Jailed     []ValidatorAddr  `json:"jailed"`
		Unjailed   []ValidatorAddr  `json:"unjailed"`
		Tombstoned []ValidatorAddr  `json:"tombstoned"`
		Slashed    []ValidatorSlash `json:"slashed"`
	}
)
