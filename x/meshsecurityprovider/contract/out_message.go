package contract

type (
	SudoMsg struct {
		Jailing    *ValidatorSlash `json:"jailing,omitempty"`
		Tombstoned *ValidatorAddr  `json:tombstone,omitempty`
	}
	// ValidatorAddr alias for the Bech32 address string of sdk.ValAddress
	ValidatorAddr = string

	ValidatorSlash struct {
		Jailed     []ValidatorAddr `json:"jailed"`
		Tombstoned []ValidatorAddr `json:"tombstoned"`
	}
)
