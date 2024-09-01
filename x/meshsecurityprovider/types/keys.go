package types

const (
	ModuleName = "test_module"

	StoreKey = ModuleName

	RouterKey = ModuleName
)

var (
	// Key defines the store key for test_module.
	Key                = []byte{0x01}
	ParamsKey          = []byte{0x02}
	ConsumerChainIDKey = []byte{0x03}
)
