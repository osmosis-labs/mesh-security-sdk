package e2e

import (
	"flag"
	"os"
	"testing"
)

var wasmContractPath string

func TestMain(m *testing.M) {
	flag.StringVar(&wasmContractPath, "contracts-path", "testdata", "Set path to dir with gzipped wasm contracts")
	flag.Parse()

	os.Exit(m.Run())
}
