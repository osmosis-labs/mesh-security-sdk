package main

import (
	"flag"
	"github.com/osmosis-labs/mesh-security-sdk/tests/starship/setup"
)

var (
	WasmContractPath    string
	WasmContractGZipped bool
	ConfigFile          string
	ProviderChain       string
	ConsumerChain       string
)

func main() {
	flag.StringVar(&WasmContractPath, "Contracts-path", "../testdata", "Set path to dir with gzipped wasm Contracts")
	flag.BoolVar(&WasmContractGZipped, "gzipped", true, "Use `.gz` file ending when set")
	flag.StringVar(&ConfigFile, "config", "../configs/starship.yaml", "starship config file for the infra")
	flag.StringVar(&ProviderChain, "provider-chain", "mesh-osmosis-1", "provider chain name, from config file")
	flag.StringVar(&ConsumerChain, "consumer-chain", "mesh-juno-1", "consumer chain name, from config file")
	flag.Parse()

	_, _, err := setup.MeshSecurity(ProviderChain, ConsumerChain, ConfigFile, WasmContractPath, WasmContractGZipped)
	if err != nil {
		panic(err)
	}
}
