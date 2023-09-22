package contract

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

type (
	SudoMsg struct {
		Rebalance    *struct{}     `json:"rebalance,omitempty"`
		ValsetUpdate *ValsetUpdate `json:"valset_update,omitempty"`
	}

	// Validator alias to wasmVM type
	Validator = wasmvmtypes.Validator
	// ValidatorAddr alias for the Bech32 address string of sdk.ValAddress
	ValidatorAddr = string

	// ValsetUpdate updates to the active validator set
	ValsetUpdate struct {
		Additions  []Validator     `json:"additions,omitempty"`
		Removals   []ValidatorAddr `json:"removals,omitempty"`
		Updated    []Validator     `json:"updated,omitempty"`
		Jailed     []ValidatorAddr `json:"slashed,omitempty"`
		Unjailed   []ValidatorAddr `json:"unjailed,omitempty"`
		Tombstoned []ValidatorAddr `json:"tombstoned,omitempty"`
	}
)
