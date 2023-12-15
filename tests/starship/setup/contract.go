package setup

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func buildPathToWasm(wasmContractPath string, fileName string, wasmContractGZipped bool) string {
	if wasmContractGZipped {
		fileName += ".gz"
	}
	return filepath.Join(wasmContractPath, fileName)
}

func submitGovProposal(chain *Client, msgs ...sdk.Msg) (uint64, error) {
	// fetch gov params from the local
	initialDeposit := sdk.NewCoins(sdk.NewCoin(chain.Denom, math.NewInt(10000000)))
	govMsg, err := govv1.NewMsgSubmitProposal(msgs, initialDeposit, chain.Address, "", "my title", "my summary")
	if err != nil {
		return 0, err
	}
	rsp, err := chain.Client.SendMsg(context.Background(), govMsg, "")
	if err != nil {
		return 0, err
	}
	// Fetch the proposal id
	proposalID := 0
	for _, event := range rsp.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if attr.Key == "proposal_id" {
					value, err := strconv.Atoi(attr.Value)
					if err != nil {
						return 0, err
					}
					proposalID = value
				}
			}
		}
	}

	fmt.Printf("submitted gov proposalID: %v\n", proposalID)
	return uint64(proposalID), nil
}

func voteAndPassGovProposal(chain *Client, proposalID uint64) error {
	vote := &govv1.MsgVote{
		ProposalId: proposalID,
		Voter:      chain.Address,
		Option:     govv1.OptionYes,
		Metadata:   "testing",
	}
	res, err := chain.Client.SendMsg(context.Background(), vote, "")
	if err != nil {
		return err
	}

	fmt.Printf("submitted vote for proposalid: %v, from %v, tx: %s\n", proposalID, chain.Address, res.TxHash)

	queryProposal := &govv1.QueryProposalRequest{
		ProposalId: proposalID,
	}

	var proposal *govv1.QueryProposalResponse
	err = Eventually(
		func() bool {
			proposal, err = govv1.NewQueryClient(chain.Client).Proposal(context.Background(), queryProposal)
			if err != nil {
				return false
			}
			if proposal.Proposal.Status >= govv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
				return true
			}
			return false
		},
		300*time.Second,
		5*time.Second,
		"waited for too long, still proposal did not pass",
	)
	if err != nil {
		return err
	}
	fmt.Print("proposal successfully passed...")
	if proposal.Proposal.Status == govv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
		return nil
	}
	return fmt.Errorf("proposal failed: id: %d, status: %d\n", proposal.Proposal.Id, proposal.Proposal.Status)
}

func InstantiateContract(chain *Client, codeID uint64, label string, initMsg []byte, funds ...sdk.Coin) (map[uint64]string, error) {
	fmt.Printf("instantiating contract with id: %v, label: %v, chain: %v\n", codeID, label, chain.ChainID)
	instantiateMsg := &wasmtypes.MsgInstantiateContract{
		Sender: chain.Address,
		Admin:  chain.Address,
		CodeID: codeID,
		Label:  label,
		Msg:    initMsg,
		Funds:  funds,
	}

	r, err := chain.Client.SendMsg(context.Background(), instantiateMsg, "")
	if err != nil {
		return nil, err
	}
	fmt.Printf("response for instantiate contract: %d\n", r.Code)

	// map of codeid and contract address
	addrs := map[uint64]string{}

	for _, event := range r.Logs[0].Events {
		if event.Type == "instantiate" {
			var eventCodeID uint64
			var addr string
			for _, attr := range event.Attributes {
				if attr.Key == "_contract_address" {
					addr = attr.Value
				}
				if attr.Key == "code_id" {
					eventCodeID, err = strconv.ParseUint(attr.Value, 10, 64)
					if err != nil {
						return nil, err
					}
				}
			}
			addrs[eventCodeID] = addr
		}
	}

	return addrs, nil
}

func StoreCodeFile(chain *Client, filename string) (wasmtypes.MsgStoreCodeResponse, error) {
	fmt.Printf("storecode file: %s chain: %s \n", filename, chain.ChainID)
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

	fmt.Printf("response for storeCodeID: %v, response: %v\n", resp.CodeID, resp)

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

	fmt.Printf("smart query response: %s\n", res.String())

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

func Eventually(condition func() bool, waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{}) error {
	ch := make(chan bool, 1)

	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			return fmt.Errorf("condition never satisfied: %s", msgAndArgs...)
		case <-tick:
			tick = nil
			go func() { ch <- condition() }()
		case v := <-ch:
			if v {
				return nil
			}
			tick = ticker.C
		}
	}
}
