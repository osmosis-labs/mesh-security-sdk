package e2e

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

func StoreCodeFile(chain *Client, filename string) (wasmtypes.MsgStoreCodeResponse, error) {
	var resp wasmtypes.MsgStoreCodeResponse
	wasmCode, err := os.ReadFile(filename)
	if err != nil {
		return resp, err
	}

	if !strings.HasSuffix(filename, "wasm") {
		return StoreCode(chain, wasmCode)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err = gz.Write(wasmCode); err != nil {
		return resp, err
	}
	if err := gz.Close(); err != nil {
		return resp, err
	}
	wasmCode = buf.Bytes()

	return StoreCode(chain, wasmCode)
}

// todo
func StoreCode(chain *Client, byteCode []byte) (wasmtypes.MsgStoreCodeResponse, error) {
	var resp wasmtypes.MsgStoreCodeResponse

	storeMsg := &wasmtypes.MsgStoreCode{
		Sender:       chain.Address,
		WASMByteCode: byteCode,
	}
	r, err := chain.Client.SendMsg(context.Background(), storeMsg, "")
	if err != nil {
		return resp, err
	}

	for _, event := range r.Logs[0].Events {
		if event.Type == "store_code" {
			for _, attr := range event.Attributes {
				if attr.Key == "code_checksum" {
					resp.Checksum = []byte(attr.Value)
				}
				if attr.Key == "code_id" {
					codeID, err := strconv.Atoi(attr.Value)
					if err != nil {
						return resp, err
					}
					resp.CodeID = uint64(codeID)
				}
			}
		}
	}

	fmt.Printf("response for storeCode: %v\n", resp)

	return resp, nil
}

// SmartQuery This will serialize the query message and submit it to the contract.
// The response is parsed into the provided interface.
// Usage: SmartQuery(addr, QueryMsg{Foo: 1}, &response)
func SmartQuery(chain *Client, contractAddr string, queryMsg interface{}, response interface{}) error {
	msg, err := json.Marshal(queryMsg)
	if err != nil {
		return err
	}

	req := wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: msg,
	}
	reqBin, err := proto.Marshal(&req)
	if err != nil {
		return err
	}

	res, err := chain.Client.QueryABCI(context.Background(), abci.RequestQuery{
		Path: "/cosmwasm.wasm.v1.Query/SmartContractState",
		Data: reqBin,
	})
	if err != nil {
		return err
	}

	fmt.Printf("smart query response: %s", res.String())

	if res.Code != 0 {
		return fmt.Errorf("query failed: (%d) %s", res.Code, res.Log)
	}

	// unpack protobuf
	var resp wasmtypes.QuerySmartContractStateResponse
	err = proto.Unmarshal(res.Value, &resp)
	if err != nil {
		return err
	}
	// unpack json content
	return json.Unmarshal(resp.Data, response)
}
