package keeper

import "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"

// option that is applied after keeper is setup with the VM. Used for decorators mainly.
type postOptsFn func(*Keeper)

func (f postOptsFn) apply(keeper *Keeper) {
	f(keeper)
}

// WithWasmKeeperDecorated can set a decorator to the wasm keeper
func WithWasmKeeperDecorated(cb func(types.WasmKeeper) types.WasmKeeper) Option {
	return postOptsFn(func(keeper *Keeper) {
		keeper.wasm = cb(keeper.wasm)
	})
}
