package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/thecrazygm/hoverfly/crypto"
	"github.com/thecrazygm/hoverfly/state"
)

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      any             `json:"id"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
	ID      any    `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCHandler struct {
	state *state.State
	debug bool
}

func NewRPCHandler(s *state.State, debug bool) *RPCHandler {
	return &RPCHandler{state: s, debug: debug}
}

func (h *RPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeError(w, nil, -32600, "Invalid Request Method (POST required)")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, nil, -32700, "Parse error reading request body")
		return
	}

	var req jsonRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, nil, -32700, "Parse error invalid JSON-RPC payload")
		return
	}

	log.Debugf("RPC Call: %s | Params: %s", req.Method, string(req.Params))

	result, rpcErr := h.route(req.Method, req.Params)
	if rpcErr != nil {
		log.Warnf("RPC Error: %d - %s", rpcErr.Code, rpcErr.Message)
		writeError(w, req.ID, rpcErr.Code, rpcErr.Message)
		return
	}

	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
	respBytes, _ := json.Marshal(resp)
	w.Write(respBytes)
}

func writeError(w http.ResponseWriter, id any, code int, msg string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		Error: rpcError{
			Code:    code,
			Message: msg,
		},
		ID: id,
	}
	respBytes, _ := json.Marshal(resp)
	w.Write(respBytes)
}

type Authority struct {
	WeightThreshold uint32  `json:"weight_threshold"`
	AccountAuths    [][]any `json:"account_auths"`
	KeyAuths        [][]any `json:"key_auths"`
}

type EnrichedAccountData struct {
	Name                   string        `json:"name"`
	VotingPower            float64       `json:"voting_power"`
	VotingManabar          state.Manabar `json:"voting_manabar"`
	LastVoteTime           string        `json:"last_vote_time"`
	Balance                string        `json:"balance"`
	HbdBalance             string        `json:"hbd_balance"`
	VestingShares          string        `json:"vesting_shares"`
	Created                string        `json:"created"`
	SavingsBalance         string        `json:"savings_balance"`
	SavingsHbdBalance      string        `json:"savings_hbd_balance"`
	PostCount              uint32        `json:"post_count"`
	Reputation             any           `json:"reputation"`
	RewardHiveBalance      string        `json:"reward_hive_balance"`
	RewardHbdBalance       string        `json:"reward_hbd_balance"`
	RewardVestingBalance   string        `json:"reward_vesting_balance"`
	RewardVestingHive      string        `json:"reward_vesting_hive"`
	DelegatedVestingShares string        `json:"delegated_vesting_shares"`
	ReceivedVestingShares  string        `json:"received_vesting_shares"`
	VestingWithdrawRate    string        `json:"vesting_withdraw_rate"`
	NextVestingWithdrawal  string        `json:"next_vesting_withdrawal"`
	Owner                  Authority     `json:"owner"`
	Active                 Authority     `json:"active"`
	Posting                Authority     `json:"posting"`
	MemoKey                string        `json:"memo_key"`
}

func enrichAccount(acc state.AccountData) EnrichedAccountData {
	activeKey := "STM5kQ1uy2CGNSwibSeYyLELWFng3HTyYVSsQd4Bjd4sWfqgKgtgJ"
	postingKey := "STM8Ep2rQp1wPzBPE2tS7tfcvU2JpbnkeyhfsYB1Jcnz7S2w8H9Q3"
	ownerKey := "STM5kQ1uy2CGNSwibSeYyLELWFng3HTyYVSsQd4Bjd4sWfqgKgtgJ"
	memoKey := "STM5kQ1uy2CGNSwibSeYyLELWFng3HTyYVSsQd4Bjd4sWfqgKgtgJ"

	savingsBalance := acc.SavingsBalance
	if savingsBalance == "" {
		savingsBalance = "0.000 HIVE"
	}
	savingsHbdBalance := acc.SavingsHbdBalance
	if savingsHbdBalance == "" {
		savingsHbdBalance = "0.000 HBD"
	}

	return EnrichedAccountData{
		Name:                   acc.Name,
		VotingPower:            acc.VotingPower,
		VotingManabar:          acc.VotingManabar,
		LastVoteTime:           acc.LastVoteTime,
		Balance:                acc.Balance,
		HbdBalance:             acc.HbdBalance,
		VestingShares:          acc.VestingShares,
		Created:                acc.Created,
		SavingsBalance:         savingsBalance,
		SavingsHbdBalance:      savingsHbdBalance,
		PostCount:              0,
		Reputation:             "1000000000000",
		RewardHiveBalance:      "0.000 HIVE",
		RewardHbdBalance:       "0.000 HBD",
		RewardVestingBalance:   "0.000000 VESTS",
		RewardVestingHive:      "0.000 HIVE",
		DelegatedVestingShares: "0.000000 VESTS",
		ReceivedVestingShares:  "0.000000 VESTS",
		VestingWithdrawRate:    "0.000000 VESTS",
		NextVestingWithdrawal:  "1970-01-01T00:00:00",
		Owner: Authority{
			WeightThreshold: 1,
			AccountAuths:    [][]any{},
			KeyAuths:        [][]any{{ownerKey, 1}},
		},
		Active: Authority{
			WeightThreshold: 1,
			AccountAuths:    [][]any{},
			KeyAuths:        [][]any{{activeKey, 1}},
		},
		Posting: Authority{
			WeightThreshold: 1,
			AccountAuths:    [][]any{},
			KeyAuths:        [][]any{{postingKey, 1}},
		},
		MemoKey: memoKey,
	}
}

func (h *RPCHandler) route(method string, params json.RawMessage) (any, *rpcError) {
	switch method {
	case "condenser_api.get_dynamic_global_properties", "database_api.get_dynamic_global_properties":
		return h.handleGetDynamicGlobalProperties()

	case "condenser_api.get_accounts", "database_api.get_accounts":
		return h.handleGetAccounts(params)

	case "condenser_api.get_account_count":
		return h.handleGetAccountCount()

	case "condenser_api.lookup_accounts":
		return h.handleLookupAccounts(params)

	case "condenser_api.lookup_account_names":
		return h.handleLookupAccountNames(params)

	case "database_api.list_accounts":
		return h.handleListAccounts(params)

	case "condenser_api.get_transaction", "account_history_api.get_transaction":
		return h.handleGetTransaction(params)

	case "account_history_api.get_account_history", "condenser_api.get_account_history":
		return h.handleGetAccountHistory(method, params)

	case "account_history_api.get_ops_in_block":
		return h.handleGetOpsInBlock(params)

	case "account_history_api.enum_virtual_ops":
		return h.handleEnumVirtualOps(params)

	case "database_api.find_accounts":
		return h.handleFindAccounts(params)

	case "condenser_api.get_key_references", "account_by_key_api.get_key_references":
		return h.handleGetKeyReferences(params)

	case "condenser_api.get_content", "tags_api.get_discussion":
		return h.handleGetContent(params)

	case "condenser_api.get_content_replies", "tags_api.get_content_replies":
		return h.handleGetContentReplies(params)

	case "condenser_api.get_active_votes":
		return h.handleGetActiveVotes(params)

	case "database_api.find_votes", "database_api.list_votes":
		return h.handleDatabaseVotes(params)

	case "condenser_api.get_blog", "follow_api.get_blog":
		return h.handleGetBlog(params)

	case "condenser_api.get_blog_entries", "follow_api.get_blog_entries":
		return h.handleGetBlogEntries(params)

	case "condenser_api.get_discussions_by_author_before_date", "condenser_api.get_discussions_by_blog",
		"condenser_api.get_discussions_by_comments", "condenser_api.get_discussions_by_created",
		"condenser_api.get_discussions_by_feed", "condenser_api.get_discussions_by_hot",
		"condenser_api.get_discussions_by_trending", "condenser_api.get_comment_discussions_by_payout",
		"condenser_api.get_post_discussions_by_payout", "tags_api.get_discussions_by_author_before_date",
		"tags_api.get_discussions_by_blog", "tags_api.get_discussions_by_comments",
		"tags_api.get_discussions_by_created", "tags_api.get_discussions_by_hot",
		"tags_api.get_discussions_by_trending", "tags_api.get_comment_discussions_by_payout",
		"tags_api.get_post_discussions_by_payout":
		return h.handleGetDiscussions(params)

	case "condenser_api.get_account_reputations", "follow_api.get_account_reputations", "reputation_api.get_account_reputations":
		return h.handleGetAccountReputations(params)

	case "condenser_api.get_followers", "follow_api.get_followers":
		return h.handleGetFollowers(params)

	case "condenser_api.get_following", "follow_api.get_following":
		return h.handleGetFollowing(params)

	case "condenser_api.get_reblogged_by", "follow_api.get_reblogged_by":
		return []string{}, nil

	case "database_api.find_comments":
		return h.handleFindComments(params)

	case "condenser_api.broadcast_transaction", "condenser_api.broadcast_transaction_synchronous",
		"network_broadcast_api.broadcast_transaction", "network_broadcast_api.broadcast_transaction_synchronous":
		return h.handleBroadcastTransaction(params)

	case "condenser_api.find_rc_accounts", "rc_api.find_rc_accounts":
		return h.handleFindRCAccounts(params)

	case "condenser_api.list_rc_accounts", "rc_api.list_rc_accounts":
		return h.handleListRCAccounts(method, params)

	case "condenser_api.list_rc_direct_delegations", "rc_api.list_rc_direct_delegations":
		return h.handleListRCDirectDelegations(method)

	case "rc_api.get_rc_operation_stats", "rc_api.get_rc_stats", "rc_api.get_resource_params", "rc_api.get_resource_pool":
		return h.handleOpenAPIExample(method)

	case "database_api.get_config", "condenser_api.get_config":
		return h.handleGetConfig()

	case "condenser_api.get_chain_properties":
		return h.handleGetChainProperties()

	case "condenser_api.get_version", "database_api.get_version":
		return h.handleGetVersion()

	case "condenser_api.get_hardfork_version":
		return "1.28.6", nil

	case "condenser_api.get_active_witnesses", "database_api.get_active_witnesses",
		"condenser_api.get_current_median_history_price", "database_api.get_current_price_feed",
		"condenser_api.get_feed_history", "database_api.get_feed_history",
		"condenser_api.get_next_scheduled_hardfork", "database_api.get_hardfork_properties",
		"condenser_api.get_reward_fund", "database_api.get_reward_funds",
		"condenser_api.get_witness_schedule", "database_api.get_witness_schedule",
		"debug_node_api.debug_get_future_witness_schedule", "debug_node_api.debug_get_hardfork_property_object",
		"debug_node_api.debug_get_witness_schedule":
		return h.handleOpenAPIExample(method)

	case "condenser_api.get_witness_count":
		return 21, nil

	case "condenser_api.lookup_witness_accounts":
		return h.handleLookupWitnessAccounts(params)

	case "condenser_api.get_witness_by_account":
		return h.handleGetWitnessByAccount(params)

	case "condenser_api.get_witnesses", "condenser_api.get_witnesses_by_vote",
		"database_api.find_witnesses", "database_api.list_witnesses":
		return h.handleOpenAPIExample(method)

	case "condenser_api.find_proposals", "database_api.find_proposals",
		"condenser_api.list_proposals", "database_api.list_proposals":
		return h.handleProposalList(method)

	case "condenser_api.list_proposal_votes", "database_api.list_proposal_votes":
		return h.handleProposalVoteList(method)

	case "database_api.list_witness_votes":
		return map[string]any{"votes": []any{}}, nil

	case "condenser_api.find_recurrent_transfers", "condenser_api.get_collateralized_conversion_requests",
		"condenser_api.get_conversion_requests", "condenser_api.get_expiring_vesting_delegations",
		"condenser_api.get_open_orders", "condenser_api.get_owner_history", "condenser_api.get_savings_withdraw_from",
		"condenser_api.get_savings_withdraw_to", "condenser_api.get_vesting_delegations",
		"condenser_api.get_withdraw_routes", "condenser_api.get_replies_by_last_update",
		"condenser_api.get_trending_tags":
		return []any{}, nil

	case "condenser_api.get_escrow", "condenser_api.get_recovery_request":
		return nil, nil

	case "database_api.find_account_recovery_requests", "database_api.find_change_recovery_account_requests",
		"database_api.find_collateralized_conversion_requests", "database_api.find_decline_voting_rights_requests",
		"database_api.find_escrows", "database_api.find_hbd_conversion_requests", "database_api.find_limit_orders",
		"database_api.find_owner_histories", "database_api.find_recurrent_transfers", "database_api.find_savings_withdrawals",
		"database_api.find_vesting_delegation_expirations", "database_api.find_vesting_delegations",
		"database_api.find_withdraw_vesting_routes", "database_api.list_account_recovery_requests",
		"database_api.list_change_recovery_account_requests", "database_api.list_collateralized_conversion_requests",
		"database_api.list_decline_voting_rights_requests", "database_api.list_escrows", "database_api.list_hbd_conversion_requests",
		"database_api.list_limit_orders", "database_api.list_owner_histories", "database_api.list_savings_withdrawals",
		"database_api.list_vesting_delegation_expirations", "database_api.list_vesting_delegations",
		"database_api.list_withdraw_vesting_routes":
		return h.handleEmptyDatabaseState(method)

	case "database_api.get_comment_pending_payouts":
		return h.handleCommentPendingPayouts(params)

	case "condenser_api.get_transaction_hex", "database_api.get_transaction_hex":
		return h.handleGetTransactionHex(method, params)

	case "condenser_api.get_potential_signatures", "database_api.get_potential_signatures":
		return h.handlePotentialSignatures(method, params)

	case "condenser_api.get_required_signatures", "database_api.get_required_signatures":
		return h.handleRequiredSignatures(method, params)

	case "condenser_api.verify_authority", "database_api.verify_authority", "database_api.verify_account_authority", "database_api.verify_signatures":
		return h.handleVerifyAuthority(method, params)

	case "debug_node_api.debug_get_head_block":
		return h.handleDebugHeadBlock()

	case "debug_node_api.debug_generate_blocks":
		return h.handleDebugGenerateBlocks(params)

	case "debug_node_api.debug_generate_blocks_until":
		return h.handleDebugGenerateBlocksUntil(params)

	case "debug_node_api.debug_get_json_schema":
		return h.handleOpenAPIExample(method)

	case "debug_node_api.debug_has_hardfork":
		return true, nil

	case "debug_node_api.debug_set_hardfork", "debug_node_api.debug_set_vest_price":
		return map[string]any{"ok": true}, nil

	case "debug_node_api.debug_throw_exception":
		return nil, &rpcError{Code: -32000, Message: "debug exception requested"}

	case "hive.db_head_state":
		return h.handleDBHeadState()

	case "search_api.find_text":
		return h.handleSearchText(params)

	case "bridge.get_post", "bridge.normalize_post":
		return h.handleBridgePost(params)

	case "bridge.get_post_header":
		return h.handleBridgePostHeader(params)

	case "bridge.get_discussion":
		return h.handleBridgeDiscussion(params)

	case "bridge.get_account_posts":
		return h.handleBridgeAccountPosts(params)

	case "bridge.get_ranked_posts":
		return h.handleBridgeRankedPosts(params)

	case "bridge.get_profile":
		return h.handleBridgeProfile(params)

	case "bridge.get_profiles":
		return h.handleBridgeProfiles(params)

	case "bridge.get_community":
		return h.handleBridgeCommunity(params)

	case "bridge.get_community_context":
		return map[string]any{"role": "guest", "subscribed": false, "title": ""}, nil

	case "bridge.list_communities":
		return h.handleBridgeListCommunities(params)

	case "bridge.list_pop_communities":
		return h.handleBridgeListPopCommunities(params)

	case "bridge.get_trending_topics":
		return h.handleBridgeTrendingTopics(params)

	case "bridge.get_payout_stats":
		return h.handleBridgePayoutStats()

	case "bridge.list_subscribers", "bridge.list_community_roles", "bridge.list_all_subscriptions", "bridge.get_follow_list":
		return []any{}, nil

	case "bridge.does_user_follow_any_lists":
		return false, nil

	case "bridge.get_relationship_between_accounts":
		return map[string]any{"blacklists": false, "follows": false, "follows_blacklists": false, "follows_muted": false, "ignores": false}, nil

	case "bridge.account_notifications", "bridge.post_notifications":
		return []any{}, nil

	case "bridge.unread_notifications":
		return map[string]any{"lastread": time.Now().UTC().Format("2006-01-02 15:04:05"), "unread": 0}, nil

	case "bridge.list_muted_reasons_enum":
		return map[string]any{
			"MUTED_COMMUNITY_MODERATION": 0,
			"MUTED_COMMUNITY_TYPE":       1,
			"MUTED_PARENT":               2,
			"MUTED_REPUTATION":           3,
			"MUTED_ROLE_COMMUNITY":       4,
		}, nil

	case "condenser_api.get_market_history", "market_history_api.get_market_history",
		"condenser_api.get_market_history_buckets", "market_history_api.get_market_history_buckets",
		"condenser_api.get_order_book", "database_api.get_order_book", "market_history_api.get_order_book",
		"condenser_api.get_recent_trades", "market_history_api.get_recent_trades",
		"condenser_api.get_ticker", "market_history_api.get_ticker",
		"condenser_api.get_trade_history", "market_history_api.get_trade_history",
		"condenser_api.get_volume", "market_history_api.get_volume":
		return h.handleOpenAPIExample(method)

	case "condenser_api.get_block", "database_api.get_block", "block_api.get_block":
		return h.handleGetBlock(params, method)

	case "condenser_api.get_block_header", "database_api.get_block_header", "block_api.get_block_header":
		return h.handleGetBlockHeader(params, method)

	case "block_api.get_block_range":
		return h.handleGetBlockRange(params)

	case "condenser_api.get_ops_in_block":
		return h.handleCondenserGetOpsInBlock(params)

	case "condenser_api.get_follow_count", "follow_api.get_follow_count":
		return h.handleGetFollowCount(params)

	case "condenser_api.is_known_transaction", "database_api.is_known_transaction":
		return h.handleIsKnownTransaction(method, params)

	case "transaction_status_api.find_transaction":
		return h.handleFindTransaction(params)

	case "jsonrpc.get_methods":
		return h.handleGetMethods()

	case "jsonrpc.get_signature":
		return h.handleGetSignature(params)

	case "hive.get_info":
		return h.handleHiveInfo()

	default:
		if result, ok := h.handleKnownOpenAPIMethod(method, params); ok {
			return result, nil
		}
		return nil, &rpcError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", method)}
	}
}

func (h *RPCHandler) handleGetDynamicGlobalProperties() (any, *rpcError) {
	props, err := h.state.GetDynamicProperties()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return props, nil
}

func (h *RPCHandler) handleGetAccounts(params json.RawMessage) (any, *rpcError) {
	var outerParams [][]string
	if err := json.Unmarshal(params, &outerParams); err != nil || len(outerParams) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	names := outerParams[0]
	var results []EnrichedAccountData

	for _, name := range names {
		var baseAcc state.AccountData
		acc, err := h.state.GetAccount(name)
		if err == nil && acc != nil {
			baseAcc = *acc
		} else {
			baseAcc = state.AccountData{
				Name:        name,
				VotingPower: 10000,
				VotingManabar: state.Manabar{
					CurrentMana:    10000,
					LastUpdateTime: time.Now().Unix(),
				},
				LastVoteTime:  "1970-01-01T00:00:00",
				Balance:       "100.000 HIVE",
				HbdBalance:    "10.000 HBD",
				VestingShares: "5000000.000000 VESTS",
				Created:       "2018-01-01T00:00:00",
			}
		}
		results = append(results, enrichAccount(baseAcc))
	}

	return results, nil
}

func (h *RPCHandler) handleFindAccounts(params json.RawMessage) (any, *rpcError) {
	var args struct {
		Accounts []string `json:"accounts"`
	}
	if err := json.Unmarshal(params, &args); err != nil || len(args.Accounts) == 0 {
		var arrArgs []struct {
			Accounts []string `json:"accounts"`
		}
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			args.Accounts = arrArgs[0].Accounts
		}
	}

	if len(args.Accounts) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	var results []EnrichedAccountData

	for _, name := range args.Accounts {
		var baseAcc state.AccountData
		acc, err := h.state.GetAccount(name)
		if err == nil && acc != nil {
			baseAcc = *acc
		} else {
			baseAcc = state.AccountData{
				Name:        name,
				VotingPower: 10000,
				VotingManabar: state.Manabar{
					CurrentMana:    10000,
					LastUpdateTime: time.Now().Unix(),
				},
				LastVoteTime:  "1970-01-01T00:00:00",
				Balance:       "100.000 HIVE",
				HbdBalance:    "10.000 HBD",
				VestingShares: "5000000.000000 VESTS",
				Created:       "2018-01-01T00:00:00",
			}
		}
		results = append(results, enrichAccount(baseAcc))
	}

	return map[string]any{
		"accounts": results,
	}, nil
}

func (h *RPCHandler) handleGetAccountCount() (any, *rpcError) {
	accounts, err := h.state.ListAccounts()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return len(accounts), nil
}

func (h *RPCHandler) handleLookupAccounts(params json.RawMessage) (any, *rpcError) {
	var args []any
	if err := json.Unmarshal(params, &args); err != nil || len(args) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	lowerBound, _ := args[0].(string)
	limit := 1000
	if len(args) > 1 {
		if rawLimit, ok := args[1].(float64); ok && rawLimit > 0 {
			limit = int(rawLimit)
		}
	}
	if limit > 1000 {
		limit = 1000
	}

	accounts, err := h.state.ListAccounts()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	results := make([]string, 0, limit)
	for _, acc := range accounts {
		if acc.Name >= lowerBound {
			results = append(results, acc.Name)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (h *RPCHandler) handleLookupAccountNames(params json.RawMessage) (any, *rpcError) {
	var args [][]string
	if err := json.Unmarshal(params, &args); err != nil || len(args) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	results := make([]any, 0, len(args[0]))
	for _, name := range args[0] {
		acc, err := h.state.GetAccount(name)
		if err != nil || acc == nil {
			results = append(results, nil)
			continue
		}
		results = append(results, enrichAccount(*acc))
	}
	return results, nil
}

func (h *RPCHandler) handleListAccounts(params json.RawMessage) (any, *rpcError) {
	var args struct {
		Start any    `json:"start"`
		Limit uint32 `json:"limit"`
		Order string `json:"order"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		var arrArgs []struct {
			Start any    `json:"start"`
			Limit uint32 `json:"limit"`
			Order string `json:"order"`
		}
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			args = arrArgs[0]
		}
	}
	if args.Limit == 0 {
		args.Limit = 1000
	}
	if args.Limit > 1000 {
		args.Limit = 1000
	}

	startName := ""
	switch start := args.Start.(type) {
	case string:
		startName = start
	case map[string]any:
		if name, ok := start["name"].(string); ok {
			startName = name
		}
	}

	accounts, err := h.state.ListAccounts()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	results := make([]EnrichedAccountData, 0, args.Limit)
	for _, acc := range accounts {
		if acc.Name >= startName {
			results = append(results, enrichAccount(acc))
			if len(results) >= int(args.Limit) {
				break
			}
		}
	}
	return map[string]any{"accounts": results}, nil
}

func (h *RPCHandler) handleGetKeyReferences(params json.RawMessage) (any, *rpcError) {
	var outerParams [][]string
	if err := json.Unmarshal(params, &outerParams); err != nil || len(outerParams) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	keys := outerParams[0]
	refs, err := h.state.GetKeyReferences(keys)
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	var results [][]string
	for _, ref := range refs {
		results = append(results, []string{ref})
	}
	for len(results) < len(keys) {
		results = append(results, []string{})
	}

	return results, nil
}

func (h *RPCHandler) handleGetContent(params json.RawMessage) (any, *rpcError) {
	author, permlink := parseContentRef(params)
	if author == "" || permlink == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	post, err := h.state.GetContent(author, permlink)
	if err == nil && post != nil {
		return post, nil
	}

	fallback := state.PostData{
		Author:       author,
		Permlink:     permlink,
		Category:     "blog",
		Title:        "Mock Post Title",
		Body:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit. This is a mock post generated by Hoverfly. 🛸",
		JSONMetadata: "{}",
		Created:      time.Now().UTC().Format("2006-01-02T15:04:05"),
		ActiveVotes:  []string{},
	}
	return fallback, nil
}

func parseContentRef(params json.RawMessage) (string, string) {
	var args []string
	if err := json.Unmarshal(params, &args); err == nil && len(args) >= 2 {
		return args[0], args[1]
	}

	var objectArgs struct {
		Author   string `json:"author"`
		Account  string `json:"account"`
		Permlink string `json:"permlink"`
	}
	if err := json.Unmarshal(params, &objectArgs); err == nil {
		author := objectArgs.Author
		if author == "" {
			author = objectArgs.Account
		}
		return author, objectArgs.Permlink
	}

	var wrappedObjectArgs []struct {
		Author   string `json:"author"`
		Account  string `json:"account"`
		Permlink string `json:"permlink"`
	}
	if err := json.Unmarshal(params, &wrappedObjectArgs); err == nil && len(wrappedObjectArgs) > 0 {
		author := wrappedObjectArgs[0].Author
		if author == "" {
			author = wrappedObjectArgs[0].Account
		}
		return author, wrappedObjectArgs[0].Permlink
	}

	return "", ""
}

func (h *RPCHandler) handleGetContentReplies(params json.RawMessage) (any, *rpcError) {
	author, permlink := parseContentRef(params)
	if author == "" || permlink == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	var replies []state.PostData
	for _, post := range posts {
		if post.ParentAuthor == author && post.ParentPermlink == permlink {
			replies = append(replies, post)
		}
	}
	return replies, nil
}

func (h *RPCHandler) handleGetActiveVotes(params json.RawMessage) (any, *rpcError) {
	author, permlink := parseContentRef(params)
	if author == "" || permlink == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	post, err := h.state.GetContent(author, permlink)
	if err != nil || post == nil {
		return []any{}, nil
	}

	votes := make([]any, 0, len(post.ActiveVotes))
	for _, voter := range post.ActiveVotes {
		votes = append(votes, map[string]any{
			"voter":      voter,
			"author":     author,
			"permlink":   permlink,
			"percent":    10000,
			"rshares":    0,
			"time":       post.Created,
			"reputation": "1000000000000",
		})
	}
	return votes, nil
}

func (h *RPCHandler) handleDatabaseVotes(params json.RawMessage) (any, *rpcError) {
	votes, rpcErr := h.handleGetActiveVotes(params)
	if rpcErr != nil {
		return map[string]any{"votes": []any{}}, nil
	}
	return map[string]any{"votes": votes}, nil
}

func (h *RPCHandler) handleGetBlog(params json.RawMessage) (any, *rpcError) {
	account, limit := parseAccountLimit(params)
	if account == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	posts, err := h.postsByAuthor(account, limit)
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return posts, nil
}

func (h *RPCHandler) handleGetBlogEntries(params json.RawMessage) (any, *rpcError) {
	account, limit := parseAccountLimit(params)
	if account == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	posts, err := h.postsByAuthor(account, limit)
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	entries := make([]any, 0, len(posts))
	for i, post := range posts {
		entries = append(entries, map[string]any{
			"entry_id":  i,
			"author":    post.Author,
			"permlink":  post.Permlink,
			"blog":      account,
			"reblog_on": post.Created,
		})
	}
	return entries, nil
}

func parseAccountLimit(params json.RawMessage) (string, int) {
	limit := 20
	var args []any
	if err := json.Unmarshal(params, &args); err == nil && len(args) > 0 {
		account, _ := args[0].(string)
		if len(args) > 2 {
			if rawLimit, ok := args[2].(float64); ok && rawLimit > 0 {
				limit = int(rawLimit)
			}
		} else if len(args) > 1 {
			if rawLimit, ok := args[1].(float64); ok && rawLimit > 0 {
				limit = int(rawLimit)
			}
		}
		return account, clampLimit(limit, 500)
	}

	var objectArgs struct {
		Account string `json:"account"`
		Limit   int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &objectArgs); err == nil {
		if objectArgs.Limit > 0 {
			limit = objectArgs.Limit
		}
		return objectArgs.Account, clampLimit(limit, 500)
	}
	return "", limit
}

func clampLimit(limit, max int) int {
	if limit <= 0 {
		return max
	}
	if limit > max {
		return max
	}
	return limit
}

func (h *RPCHandler) postsByAuthor(account string, limit int) ([]state.PostData, error) {
	posts, err := h.state.ListContent()
	if err != nil {
		return nil, err
	}

	results := make([]state.PostData, 0, limit)
	for _, post := range posts {
		if post.Author == account {
			results = append(results, post)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (h *RPCHandler) handleGetDiscussions(params json.RawMessage) (any, *rpcError) {
	limit := 20
	tag := ""
	var objectArgs struct {
		Limit int    `json:"limit"`
		Tag   string `json:"tag"`
	}
	if err := json.Unmarshal(params, &objectArgs); err == nil {
		if objectArgs.Limit > 0 {
			limit = objectArgs.Limit
		}
		tag = objectArgs.Tag
	}
	var wrappedObjectArgs []struct {
		Limit int    `json:"limit"`
		Tag   string `json:"tag"`
	}
	if tag == "" {
		if err := json.Unmarshal(params, &wrappedObjectArgs); err == nil && len(wrappedObjectArgs) > 0 {
			if wrappedObjectArgs[0].Limit > 0 {
				limit = wrappedObjectArgs[0].Limit
			}
			tag = wrappedObjectArgs[0].Tag
		}
	}
	limit = clampLimit(limit, 100)

	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	results := make([]state.PostData, 0, limit)
	for _, post := range posts {
		if tag != "" && post.Category != tag {
			continue
		}
		results = append(results, post)
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (h *RPCHandler) handleFindComments(params json.RawMessage) (any, *rpcError) {
	var args struct {
		Comments []struct {
			Author   string `json:"author"`
			Permlink string `json:"permlink"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(params, &args); err != nil || len(args.Comments) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	comments := make([]any, 0, len(args.Comments))
	for _, ref := range args.Comments {
		post, err := h.state.GetContent(ref.Author, ref.Permlink)
		if err != nil || post == nil {
			continue
		}
		comments = append(comments, post)
	}
	return map[string]any{"comments": comments}, nil
}

func (h *RPCHandler) handleGetAccountReputations(params json.RawMessage) (any, *rpcError) {
	account, limit := parseAccountLimit(params)
	if limit <= 0 {
		limit = 100
	}

	accounts, err := h.state.ListAccounts()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	results := make([]any, 0, limit)
	for _, acc := range accounts {
		if account != "" && acc.Name < account {
			continue
		}
		results = append(results, map[string]any{
			"account":    acc.Name,
			"reputation": "1000000000000",
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (h *RPCHandler) handleGetFollowers(params json.RawMessage) (any, *rpcError) {
	account, _ := parseAccountLimit(params)
	if account == "" {
		account = "thecrazygm"
	}
	return []any{
		map[string]any{"follower": "alice", "following": account, "what": []string{"blog"}},
		map[string]any{"follower": "bob", "following": account, "what": []string{"blog"}},
	}, nil
}

func (h *RPCHandler) handleGetFollowing(params json.RawMessage) (any, *rpcError) {
	account, _ := parseAccountLimit(params)
	if account == "" {
		account = "thecrazygm"
	}
	return []any{
		map[string]any{"follower": account, "following": "alice", "what": []string{"blog"}},
		map[string]any{"follower": account, "following": "bob", "what": []string{"blog"}},
	}, nil
}

func (h *RPCHandler) handleCommentPendingPayouts(params json.RawMessage) (any, *rpcError) {
	author, permlink := parseContentRef(params)
	if author == "" || permlink == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	post, err := h.state.GetContent(author, permlink)
	if err != nil || post == nil {
		return map[string]any{"cashout_infos": []any{}}, nil
	}

	return map[string]any{
		"cashout_infos": []any{
			map[string]any{
				"author":                 post.Author,
				"permlink":               post.Permlink,
				"cashout_time":           "1969-12-31T23:59:59",
				"total_vote_weight":      0,
				"max_accepted_payout":    "0.000 HBD",
				"percent_hbd":            10000,
				"allow_curation_rewards": true,
			},
		},
	}, nil
}

func (h *RPCHandler) handleFindRCAccounts(params json.RawMessage) (any, *rpcError) {
	var args struct {
		Accounts []string `json:"accounts"`
	}
	if err := json.Unmarshal(params, &args); err != nil || len(args.Accounts) == 0 {
		var arrArgs []any
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			if names, ok := arrArgs[0].([]any); ok {
				for _, n := range names {
					if s, ok := n.(string); ok {
						args.Accounts = append(args.Accounts, s)
					}
				}
			}
		}
	}

	if len(args.Accounts) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	type rcAccount struct {
		Account   string        `json:"account"`
		RcManabar state.Manabar `json:"rc_manabar"`
		MaxRC     string        `json:"max_rc"`
	}

	var results []rcAccount
	for _, name := range args.Accounts {
		results = append(results, rcAccount{
			Account: name,
			RcManabar: state.Manabar{
				CurrentMana:    16450459302631,
				LastUpdateTime: time.Now().Unix(),
			},
			MaxRC: "16450459302631",
		})
	}

	return map[string]any{
		"rc_accounts": results,
	}, nil
}

func (h *RPCHandler) handleListRCAccounts(method string, params json.RawMessage) (any, *rpcError) {
	var args struct {
		Start string `json:"start"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		var arrArgs []struct {
			Start string `json:"start"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			args = arrArgs[0]
		}
	}
	limit := clampLimit(args.Limit, 1000)

	accounts, err := h.state.ListAccounts()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	type rcAccount struct {
		Account                 string        `json:"account"`
		RcManabar               state.Manabar `json:"rc_manabar"`
		MaxRC                   string        `json:"max_rc"`
		DelegatedRC             int64         `json:"delegated_rc"`
		ReceivedDelegatedRC     int64         `json:"received_delegated_rc"`
		MaxRCCreationAdjustment any           `json:"max_rc_creation_adjustment"`
	}

	results := make([]rcAccount, 0, limit)
	for _, acc := range accounts {
		if args.Start != "" && acc.Name < args.Start {
			continue
		}
		results = append(results, rcAccount{
			Account: acc.Name,
			RcManabar: state.Manabar{
				CurrentMana:    16450459302631,
				LastUpdateTime: time.Now().Unix(),
			},
			MaxRC:               "16450459302631",
			DelegatedRC:         0,
			ReceivedDelegatedRC: 0,
			MaxRCCreationAdjustment: map[string]any{
				"amount":    "0",
				"precision": 6,
				"nai":       "@@000000037",
			},
		})
		if len(results) >= limit {
			break
		}
	}

	if method == "condenser_api.list_rc_accounts" {
		return results, nil
	}
	return map[string]any{"rc_accounts": results}, nil
}

func (h *RPCHandler) handleListRCDirectDelegations(method string) (any, *rpcError) {
	if method == "condenser_api.list_rc_direct_delegations" {
		return []any{}, nil
	}
	return map[string]any{"rc_direct_delegations": []any{}}, nil
}

func (h *RPCHandler) handleGetConfig() (any, *rpcError) {
	return map[string]any{
		"HIVE_BLOCKCHAIN_VERSION":  "1.28.6",
		"HIVE_BLOCKCHAIN_HARDFORK": 28,
	}, nil
}

func (h *RPCHandler) handleGetChainProperties() (any, *rpcError) {
	return map[string]any{
		"account_creation_fee":   "3.000 HIVE",
		"maximum_block_size":     65536,
		"hbd_interest_rate":      1500,
		"account_subsidy_budget": 797,
		"account_subsidy_decay":  347321,
	}, nil
}

func (h *RPCHandler) handleGetVersion() (any, *rpcError) {
	return map[string]any{
		"blockchain_version": "1.28.6",
		"hive_revision":      "hoverfly",
		"fc_revision":        "hoverfly",
		"haf_revision":       "hoverfly",
		"chain_id":           "beeab0de00000000000000000000000000000000000000000000000000000000",
		"node_type":          "testnet",
	}, nil
}

func (h *RPCHandler) handleLookupWitnessAccounts(params json.RawMessage) (any, *rpcError) {
	lowerBound := ""
	limit := 1000
	var args []any
	if err := json.Unmarshal(params, &args); err == nil && len(args) > 0 {
		lowerBound, _ = args[0].(string)
		if len(args) > 1 {
			if rawLimit, ok := args[1].(float64); ok && rawLimit > 0 {
				limit = int(rawLimit)
			}
		}
	}
	limit = clampLimit(limit, 1000)

	witnesses := activeWitnessNames()
	results := make([]string, 0, limit)
	for _, witness := range witnesses {
		if witness >= lowerBound {
			results = append(results, witness)
			if len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (h *RPCHandler) handleGetWitnessByAccount(params json.RawMessage) (any, *rpcError) {
	account := ""
	var args []string
	if err := json.Unmarshal(params, &args); err == nil && len(args) > 0 {
		account = args[0]
	}
	if account == "" {
		var objectArgs struct {
			Account string `json:"account"`
		}
		if err := json.Unmarshal(params, &objectArgs); err == nil {
			account = objectArgs.Account
		}
	}
	if account == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	if slices.Contains(activeWitnessNames(), account) {
		return mockWitness(account), nil
	}
	return nil, nil
}

func activeWitnessNames() []string {
	return []string{
		"abit", "arcange", "ausbitbank", "blocktrades", "deathwing", "emrebeyler",
		"good-karma", "gtg", "guiltyparties", "ocd-witness", "pharesim", "quochuy",
		"roelandp", "smooth.witness", "steempeak", "stoodkev", "themarkymark",
		"therealwolf", "threespeak", "yabapmatt", "anyx",
	}
}

func mockWitness(owner string) map[string]any {
	return map[string]any{
		"id":                       0,
		"owner":                    owner,
		"created":                  "2016-03-24T16:00:00",
		"url":                      "https://hoverfly.local/witness",
		"votes":                    "0",
		"virtual_last_update":      "0",
		"virtual_position":         "0",
		"virtual_scheduled_time":   "0",
		"total_missed":             0,
		"last_aslot":               0,
		"last_confirmed_block_num": 0,
		"signing_key":              "STM6ipXFLZyBeJRLFkXNRzAeQDz5T9zawSzYUdMShPsBHqB9W4SaC",
		"props": map[string]any{
			"account_creation_fee":   "3.000 HIVE",
			"maximum_block_size":     65536,
			"hbd_interest_rate":      1500,
			"account_subsidy_budget": 797,
			"account_subsidy_decay":  347321,
		},
		"hbd_exchange_rate": map[string]any{
			"base":  "0.200 HBD",
			"quote": "1.000 HIVE",
		},
	}
}

func (h *RPCHandler) handleProposalList(method string) (any, *rpcError) {
	if strings.HasPrefix(method, "database_api.") {
		return map[string]any{"proposals": []any{}}, nil
	}
	return []any{}, nil
}

func (h *RPCHandler) handleProposalVoteList(method string) (any, *rpcError) {
	if strings.HasPrefix(method, "database_api.") {
		return map[string]any{"proposal_votes": []any{}}, nil
	}
	return []any{}, nil
}

func (h *RPCHandler) handleEmptyDatabaseState(method string) (any, *rpcError) {
	field := "items"
	switch {
	case strings.Contains(method, "recovery"):
		field = "requests"
	case strings.Contains(method, "conversion"):
		field = "requests"
	case strings.Contains(method, "decline_voting"):
		field = "requests"
	case strings.Contains(method, "escrow"):
		field = "escrows"
	case strings.Contains(method, "limit_order"):
		field = "orders"
	case strings.Contains(method, "owner_histor"):
		field = "owner_auths"
	case strings.Contains(method, "recurrent_transfer"):
		field = "recurrent_transfers"
	case strings.Contains(method, "savings_withdraw"):
		field = "withdrawals"
	case strings.Contains(method, "vesting_delegation"):
		field = "delegations"
	case strings.Contains(method, "withdraw_vesting"):
		field = "routes"
	}
	return map[string]any{field: []any{}}, nil
}

func (h *RPCHandler) handleGetTransactionHex(method string, params json.RawMessage) (any, *rpcError) {
	var tx crypto.Transaction
	if method == "condenser_api.get_transaction_hex" {
		var args []crypto.Transaction
		if err := json.Unmarshal(params, &args); err != nil || len(args) == 0 {
			return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
		}
		tx = args[0]
	} else {
		var args struct {
			Trx crypto.Transaction `json:"trx"`
		}
		if err := json.Unmarshal(params, &args); err != nil {
			var wrapped []struct {
				Trx crypto.Transaction `json:"trx"`
			}
			if err := json.Unmarshal(params, &wrapped); err != nil || len(wrapped) == 0 {
				return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
			}
			args = wrapped[0]
		}
		tx = args.Trx
	}

	bytes, err := tx.Serialize()
	if err != nil {
		return nil, &rpcError{Code: -32602, Message: err.Error()}
	}

	hexValue := hex.EncodeToString(bytes)
	if method == "database_api.get_transaction_hex" {
		return map[string]any{"hex": hexValue}, nil
	}
	return hexValue, nil
}

func (h *RPCHandler) handlePotentialSignatures(method string, params json.RawMessage) (any, *rpcError) {
	keys := extractAvailableKeys(params)
	if method == "database_api.get_potential_signatures" {
		return map[string]any{"keys": keys}, nil
	}
	return keys, nil
}

func (h *RPCHandler) handleRequiredSignatures(method string, params json.RawMessage) (any, *rpcError) {
	keys := extractAvailableKeys(params)
	if method == "database_api.get_required_signatures" {
		return map[string]any{"keys": keys}, nil
	}
	return keys, nil
}

func extractAvailableKeys(params json.RawMessage) []string {
	var args struct {
		AvailableKeys []string `json:"available_keys"`
		Keys          []string `json:"keys"`
	}
	if err := json.Unmarshal(params, &args); err == nil {
		if len(args.AvailableKeys) > 0 {
			return args.AvailableKeys
		}
		return args.Keys
	}

	var arrArgs []any
	if err := json.Unmarshal(params, &arrArgs); err == nil {
		for _, arg := range arrArgs {
			if keyList, ok := arg.([]any); ok {
				keys := make([]string, 0, len(keyList))
				for _, rawKey := range keyList {
					if key, ok := rawKey.(string); ok {
						keys = append(keys, key)
					}
				}
				if len(keys) > 0 {
					return keys
				}
			}
		}
	}
	return []string{}
}

func (h *RPCHandler) handleVerifyAuthority(method string, params json.RawMessage) (any, *rpcError) {
	valid := true
	if method == "database_api.verify_account_authority" {
		var args struct {
			Account string   `json:"account"`
			Keys    []string `json:"keys"`
		}
		if err := json.Unmarshal(params, &args); err == nil && args.Account != "" {
			valid = len(args.Keys) > 0
		}
	}
	if strings.HasPrefix(method, "database_api.") {
		return map[string]any{"valid": valid}, nil
	}
	return valid, nil
}

func (h *RPCHandler) handleDebugHeadBlock() (any, *rpcError) {
	props, err := h.state.GetDynamicProperties()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return h.handleGetBlock(mustMarshal(map[string]uint32{"block_num": props.HeadBlockNumber}), "block_api.get_block")
}

func (h *RPCHandler) handleDBHeadState() (any, *rpcError) {
	props, err := h.state.GetDynamicProperties()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return map[string]any{
		"head_block": map[string]any{
			"block_num": props.HeadBlockNumber,
			"block_id":  props.HeadBlockID,
			"time":      props.Time,
		},
		"last_irreversible_block_num": props.LastIrreversibleBlockNum,
	}, nil
}

func (h *RPCHandler) handleDebugGenerateBlocks(params json.RawMessage) (any, *rpcError) {
	count := uint32(1)
	var args []any
	if err := json.Unmarshal(params, &args); err == nil {
		for _, arg := range args {
			if rawCount, ok := arg.(float64); ok && rawCount > 0 {
				count = uint32(rawCount)
				break
			}
		}
	}
	if count > 10000 {
		count = 10000
	}
	return h.advanceBlocks(count)
}

func (h *RPCHandler) handleDebugGenerateBlocksUntil(params json.RawMessage) (any, *rpcError) {
	target := ""
	var args []any
	if err := json.Unmarshal(params, &args); err == nil && len(args) > 0 {
		target, _ = args[0].(string)
	}
	props, err := h.state.GetDynamicProperties()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	if target == "" {
		return props.HeadBlockNumber, nil
	}
	targetTime, err := time.Parse("2006-01-02T15:04:05", target)
	if err != nil {
		return nil, &rpcError{Code: -32602, Message: err.Error()}
	}
	currentTime, err := time.Parse("2006-01-02T15:04:05", props.Time)
	if err != nil {
		currentTime = time.Now().UTC()
	}
	if !targetTime.After(currentTime) {
		return props.HeadBlockNumber, nil
	}
	blocks := uint32(targetTime.Sub(currentTime).Seconds() / 3)
	if blocks == 0 {
		blocks = 1
	}
	if blocks > 10000 {
		blocks = 10000
	}
	return h.advanceBlocks(blocks)
}

func (h *RPCHandler) advanceBlocks(count uint32) (any, *rpcError) {
	props, err := h.state.GetDynamicProperties()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	props.HeadBlockNumber += count
	if props.HeadBlockNumber > 10 {
		props.LastIrreversibleBlockNum = props.HeadBlockNumber - 10
	}
	props.Time = time.Now().UTC().Format("2006-01-02T15:04:05")
	props.HeadBlockID = fmt.Sprintf("05f5e100f72d57fd5a542459a94f3a8153c68c%02d", props.HeadBlockNumber%100)
	if err := h.state.SaveDynamicProperties(props); err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	return props.HeadBlockNumber, nil
}

func (h *RPCHandler) handleSearchText(params json.RawMessage) (any, *rpcError) {
	query := ""
	var args struct {
		Q     string `json:"q"`
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &args); err == nil {
		query = args.Q
		if query == "" {
			query = args.Query
		}
	}
	if query == "" {
		var arrArgs []any
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			query, _ = arrArgs[0].(string)
		}
	}
	query = strings.ToLower(query)

	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	results := make([]any, 0)
	for _, post := range posts {
		haystack := strings.ToLower(post.Title + " " + post.Body + " " + post.Author + " " + post.Permlink)
		if query == "" || strings.Contains(haystack, query) {
			results = append(results, post)
		}
	}
	return map[string]any{"results": results}, nil
}

func (h *RPCHandler) handleBridgePost(params json.RawMessage) (any, *rpcError) {
	author, permlink := parseContentRef(params)
	if author == "" || permlink == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	post, err := h.state.GetContent(author, permlink)
	if err != nil || post == nil {
		return bridgePost(state.PostData{
			Author:       author,
			Permlink:     permlink,
			Category:     "blog",
			Title:        "Mock Post Title",
			Body:         "Local Hoverfly bridge post for script testing.",
			JSONMetadata: "{}",
			Created:      time.Now().UTC().Format("2006-01-02T15:04:05"),
			ActiveVotes:  []string{},
		}), nil
	}
	return bridgePost(*post), nil
}

func (h *RPCHandler) handleBridgePostHeader(params json.RawMessage) (any, *rpcError) {
	post, rpcErr := h.handleBridgePost(params)
	if rpcErr != nil {
		return nil, rpcErr
	}
	postMap, _ := post.(map[string]any)
	return map[string]any{
		"author":   postMap["author"],
		"permlink": postMap["permlink"],
		"category": postMap["category"],
		"depth":    postMap["depth"],
	}, nil
}

func (h *RPCHandler) handleBridgeDiscussion(params json.RawMessage) (any, *rpcError) {
	author, permlink := parseContentRef(params)
	if author == "" || permlink == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	discussion := map[string]any{}
	root, err := h.state.GetContent(author, permlink)
	if err != nil || root == nil {
		root = &state.PostData{
			Author:       author,
			Permlink:     permlink,
			Category:     "blog",
			Title:        "Mock Post Title",
			Body:         "Local Hoverfly bridge discussion root.",
			JSONMetadata: "{}",
			Created:      time.Now().UTC().Format("2006-01-02T15:04:05"),
			ActiveVotes:  []string{},
		}
	}
	discussion[root.Author+"/"+root.Permlink] = bridgePost(*root)

	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	for _, post := range posts {
		if post.ParentAuthor == author && post.ParentPermlink == permlink {
			discussion[post.Author+"/"+post.Permlink] = bridgePost(post)
		}
	}
	return discussion, nil
}

func (h *RPCHandler) handleBridgeAccountPosts(params json.RawMessage) (any, *rpcError) {
	var args struct {
		Account string `json:"account"`
		Sort    string `json:"sort"`
		Limit   int    `json:"limit"`
	}
	if err := json.Unmarshal(params, &args); err != nil || args.Account == "" {
		var wrapped []struct {
			Account string `json:"account"`
			Sort    string `json:"sort"`
			Limit   int    `json:"limit"`
		}
		if err := json.Unmarshal(params, &wrapped); err == nil && len(wrapped) > 0 {
			args = wrapped[0]
		}
	}
	if args.Account == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}
	limit := clampLimit(args.Limit, 100)

	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	results := make([]any, 0, limit)
	for _, post := range posts {
		if post.Author != args.Account {
			continue
		}
		if args.Sort == "replies" && post.ParentAuthor == "" {
			continue
		}
		results = append(results, bridgePost(post))
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (h *RPCHandler) handleBridgeRankedPosts(params json.RawMessage) (any, *rpcError) {
	limit, tag := bridgeLimitAndTag(params)
	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	results := make([]any, 0, limit)
	for _, post := range posts {
		if tag != "" && post.Category != tag && post.ParentPermlink != tag {
			continue
		}
		results = append(results, bridgePost(post))
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func bridgePost(post state.PostData) map[string]any {
	metadata := map[string]any{}
	if post.JSONMetadata != "" {
		_ = json.Unmarshal([]byte(post.JSONMetadata), &metadata)
	}
	depth := 0
	if post.ParentAuthor != "" {
		depth = 1
	}
	category := post.Category
	if category == "" {
		category = post.ParentPermlink
	}
	if category == "" {
		category = "blog"
	}
	url := fmt.Sprintf("/%s/@%s/%s", category, post.Author, post.Permlink)
	if post.ParentAuthor != "" {
		url = fmt.Sprintf("/%s/@%s/%s#@%s/%s", category, post.ParentAuthor, post.ParentPermlink, post.Author, post.Permlink)
	}

	activeVotes := make([]any, 0, len(post.ActiveVotes))
	for _, voter := range post.ActiveVotes {
		activeVotes = append(activeVotes, map[string]any{"voter": voter, "rshares": 0})
	}

	return map[string]any{
		"active_votes":         activeVotes,
		"author":               post.Author,
		"author_payout_value":  "0.000 HBD",
		"author_reputation":    25,
		"beneficiaries":        []any{},
		"blacklists":           []any{},
		"body":                 post.Body,
		"category":             category,
		"children":             0,
		"created":              post.Created,
		"curator_payout_value": "0.000 HBD",
		"depth":                depth,
		"is_paidout":           false,
		"json_metadata":        metadata,
		"max_accepted_payout":  "1000000.000 HBD",
		"parent_author":        post.ParentAuthor,
		"parent_permlink":      post.ParentPermlink,
		"payout":               0,
		"pending_payout_value": "0.000 HBD",
		"percent_hbd":          10000,
		"permlink":             post.Permlink,
		"post_id":              stablePostID(post.Author, post.Permlink),
		"reblogs":              0,
		"replies":              []any{},
		"stats":                map[string]any{"flag_weight": 0, "gray": false, "hide": false, "total_votes": len(activeVotes)},
		"title":                post.Title,
		"updated":              post.Created,
		"url":                  url,
	}
}

func stablePostID(author, permlink string) uint32 {
	sum := sha256.Sum256([]byte(author + "/" + permlink))
	return uint32(sum[0])<<24 | uint32(sum[1])<<16 | uint32(sum[2])<<8 | uint32(sum[3])
}

func (h *RPCHandler) handleBridgeProfile(params json.RawMessage) (any, *rpcError) {
	account := bridgeProfileAccount(params)
	if account == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}
	return h.bridgeProfile(account), nil
}

func (h *RPCHandler) handleBridgeProfiles(params json.RawMessage) (any, *rpcError) {
	accounts := bridgeProfileAccounts(params)
	if len(accounts) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}
	results := make([]any, 0, len(accounts))
	for _, account := range accounts {
		results = append(results, h.bridgeProfile(account))
	}
	return results, nil
}

func (h *RPCHandler) bridgeProfile(account string) map[string]any {
	created := time.Now().UTC().Format("2006-01-02T15:04:05")
	if acc, err := h.state.GetAccount(account); err == nil && acc != nil {
		created = acc.Created
	}
	posts, _ := h.postsByAuthor(account, 500)
	return map[string]any{
		"active":     created,
		"blacklists": []any{},
		"context":    map[string]any{"followed": false, "muted": false},
		"created":    created,
		"id":         stablePostID(account, "profile"),
		"metadata": map[string]any{
			"profile": map[string]any{
				"about": "", "blacklist_description": "", "cover_image": "", "location": "",
				"muted_list_description": "", "name": account, "profile_image": "", "website": "",
			},
		},
		"name":       account,
		"post_count": len(posts),
		"reputation": 25,
		"stats":      map[string]any{"followers": 0, "following": 0, "rank": 0},
	}
}

func bridgeProfileAccount(params json.RawMessage) string {
	var args struct {
		Account string `json:"account"`
	}
	if err := json.Unmarshal(params, &args); err == nil && args.Account != "" {
		return args.Account
	}
	var wrapped []struct {
		Account string `json:"account"`
	}
	if err := json.Unmarshal(params, &wrapped); err == nil && len(wrapped) > 0 {
		return wrapped[0].Account
	}
	var arr []string
	if err := json.Unmarshal(params, &arr); err == nil && len(arr) > 0 {
		return arr[0]
	}
	return ""
}

func bridgeProfileAccounts(params json.RawMessage) []string {
	var args struct {
		Accounts []string `json:"accounts"`
	}
	if err := json.Unmarshal(params, &args); err == nil && len(args.Accounts) > 0 {
		return args.Accounts
	}
	var wrapped []struct {
		Accounts []string `json:"accounts"`
	}
	if err := json.Unmarshal(params, &wrapped); err == nil && len(wrapped) > 0 {
		return wrapped[0].Accounts
	}
	var arr [][]string
	if err := json.Unmarshal(params, &arr); err == nil && len(arr) > 0 {
		return arr[0]
	}
	return nil
}

func (h *RPCHandler) handleBridgeCommunity(params json.RawMessage) (any, *rpcError) {
	return bridgeCommunity(bridgeCommunityName(params)), nil
}

func (h *RPCHandler) handleBridgeListCommunities(params json.RawMessage) (any, *rpcError) {
	limit, _ := bridgeLimitAndTag(params)
	if limit <= 0 {
		limit = 1
	}
	communities := []any{bridgeCommunity(bridgeCommunityName(params))}
	if limit < len(communities) {
		communities = communities[:limit]
	}
	return communities, nil
}

func (h *RPCHandler) handleBridgeListPopCommunities(params json.RawMessage) (any, *rpcError) {
	name := bridgeCommunityName(params)
	community := bridgeCommunity(name)
	return []any{[]any{community["name"], community["title"]}}, nil
}

func (h *RPCHandler) handleBridgeTrendingTopics(params json.RawMessage) (any, *rpcError) {
	name := bridgeCommunityName(params)
	community := bridgeCommunity(name)
	return []any{[]any{community["name"], community["title"]}}, nil
}

func (h *RPCHandler) handleBridgePayoutStats() (any, *rpcError) {
	posts, err := h.state.ListContent()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}
	authors := map[string]struct{}{}
	for _, post := range posts {
		authors[post.Author] = struct{}{}
	}
	return map[string]any{
		"blogs": 0,
		"items": []any{
			[]any{"@local", "@local", 0, len(posts), len(authors)},
		},
		"total": 0,
	}, nil
}

func bridgeCommunity(name string) map[string]any {
	if name == "" {
		name = "hive-123456"
	}
	return map[string]any{
		"about":       "Local Hoverfly community for script testing.",
		"admins":      []string{},
		"avatar_url":  "",
		"context":     map[string]any{"role": "guest", "subscribed": false, "title": ""},
		"created_at":  "2019-10-27 08:28:51",
		"description": "Local Hoverfly community for script testing.",
		"flag_text":   "",
		"id":          stablePostID(name, "community"),
		"is_nsfw":     false,
		"lang":        "en",
		"name":        name,
		"num_authors": 0,
		"num_pending": 0,
		"settings":    map[string]any{},
		"subscribers": 0,
		"sum_pending": 0,
		"team":        []any{[]any{name, "owner", ""}},
		"title":       name,
		"type_id":     1,
	}
}

func bridgeCommunityName(params json.RawMessage) string {
	var args struct {
		Name      string `json:"name"`
		Community string `json:"community"`
	}
	if err := json.Unmarshal(params, &args); err == nil {
		if args.Name != "" {
			return args.Name
		}
		return args.Community
	}
	var wrapped []struct {
		Name      string `json:"name"`
		Community string `json:"community"`
	}
	if err := json.Unmarshal(params, &wrapped); err == nil && len(wrapped) > 0 {
		if wrapped[0].Name != "" {
			return wrapped[0].Name
		}
		return wrapped[0].Community
	}
	return ""
}

func bridgeLimitAndTag(params json.RawMessage) (int, string) {
	limit := 20
	tag := ""
	var args struct {
		Limit int    `json:"limit"`
		Tag   string `json:"tag"`
	}
	if err := json.Unmarshal(params, &args); err == nil {
		if args.Limit > 0 {
			limit = args.Limit
		}
		tag = args.Tag
	}
	var wrapped []struct {
		Limit int    `json:"limit"`
		Tag   string `json:"tag"`
	}
	if err := json.Unmarshal(params, &wrapped); err == nil && len(wrapped) > 0 {
		if wrapped[0].Limit > 0 {
			limit = wrapped[0].Limit
		}
		if wrapped[0].Tag != "" {
			tag = wrapped[0].Tag
		}
	}
	return clampLimit(limit, 100), tag
}

func (h *RPCHandler) handleGetBlock(params json.RawMessage, method string) (any, *rpcError) {
	var blockNum uint32
	var arrArgs []uint32
	if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
		blockNum = arrArgs[0]
	} else {
		var objArgs struct {
			BlockNum uint32 `json:"block_num"`
		}
		if err := json.Unmarshal(params, &objArgs); err == nil && objArgs.BlockNum > 0 {
			blockNum = objArgs.BlockNum
		} else {
			var arrObjArgs []struct {
				BlockNum uint32 `json:"block_num"`
			}
			if err := json.Unmarshal(params, &arrObjArgs); err == nil && len(arrObjArgs) > 0 {
				blockNum = arrObjArgs[0].BlockNum
			} else {
				return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
			}
		}
	}

	blockObj := map[string]any{
		"block_id":                fmt.Sprintf("05f5e100f72d57fd5a542459a94f3a8153c68c%02d", blockNum%100),
		"previous":                "05f5e0fff72d57fd5a542459a94f3a8153c68c4a",
		"timestamp":               time.Now().UTC().Format("2006-01-02T15:04:05"),
		"witness":                 "blocktrades",
		"transaction_merkle_root": "0000000000000000000000000000000000000000",
		"extensions":              []any{},
		"witness_signature":       "207f...mock...sig",
		"transactions":            []any{},
		"transaction_ids":         []any{},
		"signing_key":             "STM6ipXFLZyBeJRLFkXNRzAeQDz5T9zawSzYUdMShPsBHqB9W4SaC",
	}

	if method == "block_api.get_block" {
		return map[string]any{
			"block": blockObj,
		}, nil
	}
	return blockObj, nil
}

func (h *RPCHandler) handleGetBlockHeader(params json.RawMessage, method string) (any, *rpcError) {
	var blockNum uint32
	var arrArgs []uint32
	if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
		blockNum = arrArgs[0]
	} else {
		var objArgs struct {
			BlockNum uint32 `json:"block_num"`
		}
		if err := json.Unmarshal(params, &objArgs); err == nil && objArgs.BlockNum > 0 {
			blockNum = objArgs.BlockNum
		} else {
			var arrObjArgs []struct {
				BlockNum uint32 `json:"block_num"`
			}
			if err := json.Unmarshal(params, &arrObjArgs); err == nil && len(arrObjArgs) > 0 {
				blockNum = arrObjArgs[0].BlockNum
			} else {
				return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
			}
		}
	}

	headerObj := map[string]any{
		"previous":                fmt.Sprintf("05f5e100f72d57fd5a542459a94f3a8153c68c%02d", (blockNum-1)%100),
		"timestamp":               time.Now().UTC().Format("2006-01-02T15:04:05"),
		"witness":                 "blocktrades",
		"transaction_merkle_root": "0000000000000000000000000000000000000000",
		"extensions":              []any{},
	}

	if method == "block_api.get_block_header" {
		return map[string]any{
			"header": headerObj,
		}, nil
	}
	return headerObj, nil
}

func (h *RPCHandler) handleGetBlockRange(params json.RawMessage) (any, *rpcError) {
	var args struct {
		StartingBlockNum uint32 `json:"starting_block_num"`
		Count            uint32 `json:"count"`
	}
	if err := json.Unmarshal(params, &args); err != nil || args.Count == 0 {
		var arrArgs []struct {
			StartingBlockNum uint32 `json:"starting_block_num"`
			Count            uint32 `json:"count"`
		}
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			args = arrArgs[0]
		}
	}
	if args.Count == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}
	if args.Count > 1000 {
		args.Count = 1000
	}

	blocks := make([]any, 0, args.Count)
	for i := uint32(0); i < args.Count; i++ {
		block, rpcErr := h.handleGetBlock(mustMarshal(map[string]uint32{"block_num": args.StartingBlockNum + i}), "block_api.get_block")
		if rpcErr != nil {
			return nil, rpcErr
		}
		if blockMap, ok := block.(map[string]any); ok {
			blocks = append(blocks, blockMap["block"])
		}
	}

	return map[string]any{"blocks": blocks}, nil
}

func mustMarshal(v any) json.RawMessage {
	bytes, _ := json.Marshal(v)
	return bytes
}

func (h *RPCHandler) handleBroadcastTransaction(params json.RawMessage) (any, *rpcError) {
	var outerParams []crypto.Transaction
	if err := json.Unmarshal(params, &outerParams); err != nil || len(outerParams) == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	tx := outerParams[0]

	chainID := "0000000000000000000000000000000000000000000000000000000000000000"
	recoveredKeys, err := crypto.VerifySignatures(&tx, chainID)
	if err != nil {
		testnetChainID := "beeab30de373dca1e2f036c30d4970470d0d57d055748a30de53070470d0d57d"
		recoveredKeys, err = crypto.VerifySignatures(&tx, testnetChainID)
		if err != nil {
			log.Warnf("Transaction signature verification FAILED: %v", err)
			return nil, &rpcError{Code: -32000, Message: fmt.Sprintf("signature verification failed: %v", err)}
		}
	}

	log.Infof("Transaction verified successfully. Recovered signing key(s): %v", recoveredKeys)

	for _, rawOp := range tx.Operations {
		var tuple []json.RawMessage
		if err := json.Unmarshal(rawOp, &tuple); err == nil && len(tuple) == 2 {
			var opName string
			json.Unmarshal(tuple[0], &opName)

			switch opName {
			case "transfer":
				var op struct {
					From   string `json:"from"`
					To     string `json:"to"`
					Amount string `json:"amount"`
					Memo   string `json:"memo"`
				}
				if err := json.Unmarshal(tuple[1], &op); err == nil {
					h.mutateTransfer(op.From, op.To, op.Amount)
				}

			case "transfer_to_savings":
				var op struct {
					From   string `json:"from"`
					To     string `json:"to"`
					Amount string `json:"amount"`
					Memo   string `json:"memo"`
				}
				if err := json.Unmarshal(tuple[1], &op); err == nil {
					h.mutateTransferToSavings(op.From, op.To, op.Amount)
				}

			case "comment":
				var op struct {
					Author         string `json:"author"`
					Permlink       string `json:"permlink"`
					ParentAuthor   string `json:"parent_author"`
					ParentPermlink string `json:"parent_permlink"`
					Category       string `json:"category"`
					Title          string `json:"title"`
					Body           string `json:"body"`
					JSONMetadata   string `json:"json_metadata"`
				}
				if err := json.Unmarshal(tuple[1], &op); err == nil {
					h.mutateComment(op.Author, op.Permlink, op.ParentAuthor, op.ParentPermlink, op.Category, op.Title, op.Body, op.JSONMetadata)
				}
			}
		}
	}

	txBytes, _ := tx.Serialize()
	hash := sha256.Sum256(txBytes)
	txID := hex.EncodeToString(hash[:20])

	props, _ := h.state.GetDynamicProperties()
	blockNum := uint32(100000001)
	if props != nil {
		blockNum = props.HeadBlockNumber + 1
	}

	// Save transaction to state for later polling (get_transaction)
	var ops []any
	for _, rawOp := range tx.Operations {
		var op []any
		if err := json.Unmarshal(rawOp, &op); err == nil {
			ops = append(ops, op)
		}
	}

	h.state.SaveTransaction(&state.TransactionData{
		TransactionID:  txID,
		BlockNum:       blockNum,
		TransactionNum: 1,
		RefBlockNum:    tx.RefBlockNum,
		RefBlockPrefix: tx.RefBlockPrefix,
		Expiration:     tx.Expiration,
		Operations:     ops,
		Extensions:     tx.Extensions,
		Signatures:     tx.Signatures,
	})

	return map[string]any{
		"id":        txID,
		"block_num": blockNum,
		"trx_num":   1,
		"expired":   false,
	}, nil
}

func (h *RPCHandler) handleGetTransaction(params json.RawMessage) (any, *rpcError) {
	txID := parseTransactionID(params)
	if txID == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	tx, err := h.state.GetTransaction(txID)
	if err != nil {
		return nil, &rpcError{Code: -32000, Message: fmt.Sprintf("Transaction not found: %s", txID)}
	}

	return tx, nil
}

func (h *RPCHandler) handleGetAccountHistory(method string, params json.RawMessage) (any, *rpcError) {
	account, limit := parseAccountLimit(params)
	if account == "" {
		var args []any
		if err := json.Unmarshal(params, &args); err == nil && len(args) > 0 {
			account, _ = args[0].(string)
			if len(args) > 2 {
				if rawLimit, ok := args[2].(float64); ok && rawLimit > 0 {
					limit = int(rawLimit)
				}
			}
		}
	}
	if account == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}
	limit = clampLimit(limit, 1000)

	transactions, err := h.state.ListTransactions()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	history := make([]any, 0, limit)
	index := 0
	for _, txData := range transactions {
		for _, op := range txData.Operations {
			if !operationTouchesAccount(op, account) {
				continue
			}
			entry := map[string]any{
				"trx_id":       txData.TransactionID,
				"block":        txData.BlockNum,
				"trx_in_block": txData.TransactionNum,
				"op_in_trx":    0,
				"timestamp":    txData.Expiration,
				"op":           op,
				"virtual_op":   false,
			}
			if method == "condenser_api.get_account_history" {
				history = append(history, []any{index, entry})
			} else {
				history = append(history, entry)
			}
			index++
			if len(history) >= limit {
				break
			}
		}
		if len(history) >= limit {
			break
		}
	}

	if method == "condenser_api.get_account_history" {
		return history, nil
	}
	return map[string]any{"history": history}, nil
}

func operationTouchesAccount(op any, account string) bool {
	bytes, err := json.Marshal(op)
	if err != nil {
		return false
	}
	return strings.Contains(string(bytes), `"`+account+`"`)
}

func (h *RPCHandler) handleGetOpsInBlock(params json.RawMessage) (any, *rpcError) {
	blockNum, onlyVirtual := parseBlockNumAndVirtual(params)
	if blockNum == 0 {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	transactions, err := h.state.ListTransactions()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	ops := make([]any, 0)
	if !onlyVirtual {
		for _, txData := range transactions {
			if txData.BlockNum != blockNum {
				continue
			}
			for idx, op := range txData.Operations {
				ops = append(ops, map[string]any{
					"trx_id":       txData.TransactionID,
					"block":        txData.BlockNum,
					"trx_in_block": txData.TransactionNum,
					"op_in_trx":    idx,
					"timestamp":    txData.Expiration,
					"op":           op,
					"virtual_op":   false,
				})
			}
		}
	}
	return map[string]any{"ops": ops}, nil
}

func (h *RPCHandler) handleCondenserGetOpsInBlock(params json.RawMessage) (any, *rpcError) {
	result, rpcErr := h.handleGetOpsInBlock(params)
	if rpcErr != nil {
		return nil, rpcErr
	}
	resultMap, _ := result.(map[string]any)
	ops, _ := resultMap["ops"].([]any)
	return ops, nil
}

func parseBlockNumAndVirtual(params json.RawMessage) (uint32, bool) {
	var args []any
	if err := json.Unmarshal(params, &args); err == nil && len(args) > 0 {
		blockNum, _ := args[0].(float64)
		onlyVirtual := false
		if len(args) > 1 {
			onlyVirtual, _ = args[1].(bool)
		}
		return uint32(blockNum), onlyVirtual
	}

	var objectArgs struct {
		BlockNum    uint32 `json:"block_num"`
		OnlyVirtual bool   `json:"only_virtual"`
	}
	if err := json.Unmarshal(params, &objectArgs); err == nil {
		return objectArgs.BlockNum, objectArgs.OnlyVirtual
	}
	return 0, false
}

func (h *RPCHandler) handleEnumVirtualOps(params json.RawMessage) (any, *rpcError) {
	return map[string]any{
		"ops":                    []any{},
		"ops_by_block":           []any{},
		"next_block_range_begin": 0,
		"next_operation_begin":   "0",
	}, nil
}

func parseTransactionID(params json.RawMessage) string {
	var stringArgs []string
	if err := json.Unmarshal(params, &stringArgs); err == nil && len(stringArgs) > 0 {
		return stringArgs[0]
	}

	var objectArgs struct {
		ID            string `json:"id"`
		TransactionID string `json:"transaction_id"`
		TrxID         string `json:"trx_id"`
	}
	if err := json.Unmarshal(params, &objectArgs); err == nil {
		switch {
		case objectArgs.ID != "":
			return objectArgs.ID
		case objectArgs.TransactionID != "":
			return objectArgs.TransactionID
		case objectArgs.TrxID != "":
			return objectArgs.TrxID
		}
	}

	var wrappedObjectArgs []struct {
		ID            string `json:"id"`
		TransactionID string `json:"transaction_id"`
		TrxID         string `json:"trx_id"`
	}
	if err := json.Unmarshal(params, &wrappedObjectArgs); err == nil && len(wrappedObjectArgs) > 0 {
		switch {
		case wrappedObjectArgs[0].ID != "":
			return wrappedObjectArgs[0].ID
		case wrappedObjectArgs[0].TransactionID != "":
			return wrappedObjectArgs[0].TransactionID
		case wrappedObjectArgs[0].TrxID != "":
			return wrappedObjectArgs[0].TrxID
		}
	}

	return ""
}

func (h *RPCHandler) mutateTransfer(from, to, amountStr string) {
	parts := strings.Fields(amountStr)
	if len(parts) != 2 {
		return
	}
	var val float64
	fmt.Sscanf(parts[0], "%f", &val)
	symbol := parts[1]

	updateBal := func(name string, add float64) {
		acc, err := h.state.GetAccount(name)
		if err != nil {
			acc = &state.AccountData{
				Name:        name,
				VotingPower: 10000,
				VotingManabar: state.Manabar{
					CurrentMana:    10000,
					LastUpdateTime: time.Now().Unix(),
				},
				LastVoteTime:  "1970-01-01T00:00:00",
				Balance:       "0.000 HIVE",
				HbdBalance:    "0.000 HBD",
				VestingShares: "0.000000 VESTS",
				Created:       time.Now().UTC().Format("2006-01-02T15:04:05"),
			}
		}

		if symbol == "HIVE" {
			var current float64
			fmt.Sscanf(acc.Balance, "%f", &current)
			acc.Balance = fmt.Sprintf("%.3f HIVE", current+add)
		} else if symbol == "HBD" {
			var current float64
			fmt.Sscanf(acc.HbdBalance, "%f", &current)
			acc.HbdBalance = fmt.Sprintf("%.3f HBD", current+add)
		}

		h.state.SaveAccount(acc)
	}

	updateBal(from, -val)
	updateBal(to, val)
	log.Infof("State Mutated (Transfer): %s -> %s (%s)", from, to, amountStr)
}

func (h *RPCHandler) mutateTransferToSavings(from, to, amountStr string) {
	parts := strings.Fields(amountStr)
	if len(parts) != 2 {
		return
	}
	var val float64
	fmt.Sscanf(parts[0], "%f", &val)
	symbol := parts[1]

	// Deduct from sender's liquid balance
	accFrom, err := h.state.GetAccount(from)
	if err == nil && accFrom != nil {
		if symbol == "HIVE" {
			var current float64
			fmt.Sscanf(accFrom.Balance, "%f", &current)
			accFrom.Balance = fmt.Sprintf("%.3f HIVE", current-val)
		} else if symbol == "HBD" {
			var current float64
			fmt.Sscanf(accFrom.HbdBalance, "%f", &current)
			accFrom.HbdBalance = fmt.Sprintf("%.3f HBD", current-val)
		}
		h.state.SaveAccount(accFrom)
	}

	// Add to receiver's savings balance
	accTo, err := h.state.GetAccount(to)
	if err != nil {
		accTo = &state.AccountData{
			Name:        to,
			VotingPower: 10000,
			VotingManabar: state.Manabar{
				CurrentMana:    10000,
				LastUpdateTime: time.Now().Unix(),
			},
			LastVoteTime:  "1970-01-01T00:00:00",
			Balance:       "0.000 HIVE",
			HbdBalance:    "0.000 HBD",
			VestingShares: "0.000000 VESTS",
			Created:       time.Now().UTC().Format("2006-01-02T15:04:05"),
		}
	}

	if accTo.SavingsBalance == "" {
		accTo.SavingsBalance = "0.000 HIVE"
	}
	if accTo.SavingsHbdBalance == "" {
		accTo.SavingsHbdBalance = "0.000 HBD"
	}

	if symbol == "HIVE" {
		var current float64
		fmt.Sscanf(accTo.SavingsBalance, "%f", &current)
		accTo.SavingsBalance = fmt.Sprintf("%.3f HIVE", current+val)
	} else if symbol == "HBD" {
		var current float64
		fmt.Sscanf(accTo.SavingsHbdBalance, "%f", &current)
		accTo.SavingsHbdBalance = fmt.Sprintf("%.3f HBD", current+val)
	}
	h.state.SaveAccount(accTo)

	log.Infof("State Mutated (Transfer to Savings): %s -> %s (%s)", from, to, amountStr)
}

func (h *RPCHandler) mutateComment(author, permlink, parentAuthor, parentPermlink, category, title, body, jsonMeta string) {
	post := &state.PostData{
		Author:         author,
		Permlink:       permlink,
		ParentAuthor:   parentAuthor,
		ParentPermlink: parentPermlink,
		Category:       category,
		Title:          title,
		Body:           body,
		JSONMetadata:   jsonMeta,
		Created:        time.Now().UTC().Format("2006-01-02T15:04:05"),
		ActiveVotes:    []string{},
	}
	h.state.SaveContent(post)
	log.Infof("State Mutated (Comment): @%s/%s", author, permlink)
}

func (h *RPCHandler) handleGetFollowCount(params json.RawMessage) (any, *rpcError) {
	var names []string
	if err := json.Unmarshal(params, &names); err == nil && len(names) > 0 {
		return map[string]any{
			"account":         names[0],
			"follower_count":  10,
			"following_count": 5,
		}, nil
	}

	var args struct {
		Account string `json:"account"`
	}
	if err := json.Unmarshal(params, &args); err == nil && args.Account != "" {
		return map[string]any{
			"account":         args.Account,
			"follower_count":  10,
			"following_count": 5,
		}, nil
	}

	return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
}

func (h *RPCHandler) handleIsKnownTransaction(method string, params json.RawMessage) (any, *rpcError) {
	txID := parseTransactionID(params)
	if txID == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	_, err := h.state.GetTransaction(txID)
	isKnown := err == nil
	if method == "database_api.is_known_transaction" {
		return map[string]any{"is_known": isKnown}, nil
	}
	return isKnown, nil
}

func (h *RPCHandler) handleFindTransaction(params json.RawMessage) (any, *rpcError) {
	txID := parseTransactionID(params)
	if txID == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}

	tx, err := h.state.GetTransaction(txID)
	if err != nil {
		return map[string]any{"status": "unknown"}, nil
	}

	props, err := h.state.GetDynamicProperties()
	if err == nil && props != nil && tx.BlockNum <= props.LastIrreversibleBlockNum {
		return map[string]any{"status": "within_irreversible_block"}, nil
	}
	return map[string]any{"status": "within_reversible_block"}, nil
}

func (h *RPCHandler) handleGetMethods() (any, *rpcError) {
	methods := make([]string, 0, len(knownHiveAPIMethods))
	for method := range knownHiveAPIMethods {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return methods, nil
}

func (h *RPCHandler) handleGetSignature(params json.RawMessage) (any, *rpcError) {
	var args struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(params, &args); err != nil || args.Method == "" {
		var arrArgs []struct {
			Method string `json:"method"`
		}
		if err := json.Unmarshal(params, &arrArgs); err == nil && len(arrArgs) > 0 {
			args.Method = arrArgs[0].Method
		}
	}
	if args.Method == "" {
		return nil, &rpcError{Code: -32602, Message: "Invalid parameters"}
	}
	if _, ok := knownHiveAPIMethods[args.Method]; !ok {
		return nil, &rpcError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", args.Method)}
	}

	var responseShape any
	if raw, ok := openAPIMockResponses[args.Method]; ok {
		_ = json.Unmarshal([]byte(raw), &responseShape)
	}

	return map[string]any{
		"args": map[string]any{
			"method": args.Method,
			"params": "see https://api.syncad.com/hived-api/",
		},
		"ret": responseShape,
	}, nil
}

func (h *RPCHandler) handleHiveInfo() (any, *rpcError) {
	props, err := h.state.GetDynamicProperties()
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: err.Error()}
	}

	return map[string]any{
		"database_head_block":          props.HeadBlockNumber,
		"database_schema_version":      2,
		"database_patch_date":          props.Time,
		"database_patched_to_revision": "hoverfly",
		"hivemind_version":             "hoverfly-0.1.0",
		"hivemind_git_rev":             "hoverfly",
		"hivemind_git_date":            props.Time,
	}, nil
}

func (h *RPCHandler) handleKnownOpenAPIMethod(method string, params json.RawMessage) (any, bool) {
	if _, ok := knownHiveAPIMethods[method]; !ok {
		return nil, false
	}

	log.Debugf("Using generic OpenAPI mock response for %s", method)

	result, rpcErr := h.handleOpenAPIExample(method)
	if rpcErr == nil {
		return result, true
	}

	return map[string]any{}, true
}

func (h *RPCHandler) handleOpenAPIExample(method string) (any, *rpcError) {
	if raw, ok := openAPIMockResponses[method]; ok {
		var result any
		if err := json.Unmarshal([]byte(raw), &result); err == nil {
			return result, nil
		}
	}

	return nil, &rpcError{Code: -32603, Message: fmt.Sprintf("invalid OpenAPI mock response for %s", method)}
}
