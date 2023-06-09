package e2e

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var (
	wasmContractPath    string
	wasmContractGZipped bool
)

func TestMain(m *testing.M) {
	flag.StringVar(&wasmContractPath, "contracts-path", "testdata", "Set path to dir with gzipped wasm contracts")
	flag.BoolVar(&wasmContractGZipped, "gzipped", true, "Use `.gz` file ending when set")
	flag.Parse()

	os.Exit(m.Run())
}

func buildPathToWasm(fileName string) string {
	if wasmContractGZipped {
		fileName += ".gz"
	}
	return filepath.Join(wasmContractPath, fileName)
}
