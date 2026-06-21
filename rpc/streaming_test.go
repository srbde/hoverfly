package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/srbde/hoverfly/state"
)

func rpcRequest(t *testing.T, handler *RPCHandler, body string) jsonRPCResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var response jsonRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode RPC response: %v", err)
	}
	if response.Error != nil {
		t.Fatalf("unexpected RPC error: %v", response.Error)
	}
	return response
}

func TestCondenserStreamingUsesLiveStateBackedBlocks(t *testing.T) {
	s, err := state.NewState("", false)
	if err != nil {
		t.Fatalf("failed to create state: %v", err)
	}
	defer s.Close()

	const headBlock = uint32(100000122)
	props, err := s.GetDynamicProperties()
	if err != nil {
		t.Fatalf("failed to get dynamic properties: %v", err)
	}
	props.HeadBlockNumber = headBlock
	props.HeadBlockID = state.BlockID(headBlock)
	props.LastIrreversibleBlockNum = headBlock - 10
	if err := s.SaveDynamicProperties(props); err != nil {
		t.Fatalf("failed to set dynamic properties: %v", err)
	}

	handler := NewRPCHandler(s, false, false)

	propertiesResponse := rpcRequest(t, handler, `{"jsonrpc":"2.0","method":"condenser_api.get_dynamic_global_properties","params":[],"id":1}`)
	properties, ok := propertiesResponse.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %T", propertiesResponse.Result)
	}
	if got := uint32(properties["head_block_number"].(float64)); got != headBlock {
		t.Fatalf("expected live head block %d, got %d", headBlock, got)
	}

	broadcastResponse := rpcRequest(t, handler, `{
		"jsonrpc":"2.0",
		"method":"condenser_api.broadcast_transaction",
		"params":[{
			"ref_block_num":36312,
			"ref_block_prefix":3608149636,
			"expiration":"2026-06-21T16:18:22",
			"operations":[{
				"type":"custom_json_operation",
				"value":{
					"required_auths":[],
					"required_posting_auths":["bob"],
					"id":"cambium_register",
					"json":"{\"l2_public_key\":\"9c09ebd94e3b2d5828ea806382866df6d1558dcf280c479a64b622b157817abd\"}"
				}
			}],
			"extensions":[],
			"signatures":[]
		}],
		"id":2
	}`)
	broadcastResult, ok := broadcastResponse.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected broadcast result object, got %T", broadcastResponse.Result)
	}
	txID, ok := broadcastResult["id"].(string)
	if !ok || txID == "" {
		t.Fatalf("expected transaction ID, got %v", broadcastResult["id"])
	}

	const transactionBlock = headBlock + 1
	props.HeadBlockNumber = transactionBlock
	props.HeadBlockID = state.BlockID(transactionBlock)
	props.LastIrreversibleBlockNum = transactionBlock - 10
	if err := s.SaveDynamicProperties(props); err != nil {
		t.Fatalf("failed to advance dynamic properties: %v", err)
	}

	blockResponse := rpcRequest(t, handler, fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"condenser_api.get_block","params":[%d],"id":3}`,
		transactionBlock,
	))
	block, ok := blockResponse.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected block object, got %T", blockResponse.Result)
	}

	blockID, ok := block["block_id"].(string)
	if !ok || len(blockID) != 40 || blockID[:8] != fmt.Sprintf("%08x", transactionBlock) {
		t.Fatalf("block ID %q does not encode height %d", blockID, transactionBlock)
	}

	transactionIDs, ok := block["transaction_ids"].([]any)
	if !ok || len(transactionIDs) != 1 || transactionIDs[0] != txID {
		t.Fatalf("expected transaction ID %s in block, got %v", txID, block["transaction_ids"])
	}

	transactions, ok := block["transactions"].([]any)
	if !ok || len(transactions) != 1 {
		t.Fatalf("expected one transaction in block, got %v", block["transactions"])
	}
	transaction := transactions[0].(map[string]any)
	operations := transaction["operations"].([]any)
	operation := operations[0].(map[string]any)
	if operation["type"] != "custom_json_operation" {
		t.Fatalf("expected custom_json operation, got %v", operation)
	}
}
