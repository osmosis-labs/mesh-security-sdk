package keeper

import (
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
)

func TestCustomMeshSecDispatchMsg(t *testing.T) {
	specs := map[string]struct {
		src    wasmvmtypes.CosmosMsg
		expErr bool
	}{
		"handle bond msg":             {},
		"handle unbond msg":           {},
		"non custom msg":              {},
		"non json msg":                {},
		"non mesh-sec msg":            {},
		"unknown mesh-sec custom msg": {},
		"unauthorized contract":       {},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			_ = spec
		})
	}
}
