package main

import (
	"flag"
	"github.com/osmosis-labs/mesh-security-sdk/tests/starship/setup"
)

var (
	WasmContractPath    string
	WasmContractGZipped bool
	ConfigFile          string
)

func main() {
	flag.StringVar(&WasmContractPath, "Contracts-path", "../testdata", "Set path to dir with gzipped wasm Contracts")
	flag.BoolVar(&WasmContractGZipped, "gzipped", true, "Use `.gz` file ending when set")
	flag.StringVar(&ConfigFile, "config", "../configs/starship.yaml", "starship config file for the infra")
	flag.Parse()

	_, _, err := setup.MeshSecurity(ConfigFile, WasmContractPath, WasmContractGZipped)
	if err != nil {
		panic(err)
	}
}
