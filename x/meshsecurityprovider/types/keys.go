package types

const (
	ModuleName = "mesh-provider"

	StoreKey = ModuleName

	RouterKey = ModuleName
)

var (
	// Key defines the store key for mesh security provider module.
	Key       = []byte{0x01}
	ParamsKey = []byte{0x02}
)
