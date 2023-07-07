package e2e

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.StringVar(&wasmContractPath, "contracts-path", "testdata", "Set path to dir with gzipped wasm contracts")
	flag.BoolVar(&wasmContractGZipped, "gzipped", true, "Use `.gz` file ending when set")
	flag.StringVar(&configFile, "config", "configs/starship.yaml", "starship config file for the infra")
	flag.Parse()

	os.Exit(m.Run())
}
