package contract

import (
	consumermsg "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

type (
	CustomMsg struct {
		Bond   *consumermsg.BondMsg   `json:"bond,omitempty"`
		Unbond *consumermsg.UnbondMsg `json:"unbond,omitempty"`
	}
)
