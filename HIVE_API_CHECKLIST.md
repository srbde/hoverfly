# Hive JSON-RPC API Mock Checklist

Source: `https://api.syncad.com/hived-api/` (`info.version` 1.28.6), fetched 2026-05-25.

Legend:

- `[x]` Hoverfly has a first-class route for this method. Status is `useful` when it is good enough for developer app testing, or `partial` when it only has a placeholder first-class answer.
- `[ ]` Hoverfly answers this method through the OpenAPI 200-response example fallback, but it still needs review or a richer first-class mock before it counts as useful.
- `stateful` means the mock likely needs to read or mutate Hoverfly state, or at least return state-shaped data that changes with seeded accounts, blocks, posts, transactions, market/order data, follow data, or governance data.
- `static` means a deterministic template response is probably enough for first coverage.

Current first-class coverage: 215/215 OpenAPI methods routed by Hoverfly.
Current answer coverage: 215/215 OpenAPI methods return a JSON-RPC result through either a first-class route or the OpenAPI example fallback.
Current useful coverage: 215/215 methods are useful first-class mocks; 0/215 are first-class but partial; 0/215 still rely on the unrouted OpenAPI fallback.

## `account_by_key_api` (1/1 done)

- [x] `account_by_key_api.get_key_references` - useful - Returns all accounts that have the key associated with their owner or active authorities.

## `account_history_api` (4/4 done)

- [x] `account_history_api.enum_virtual_ops` - useful - Allows specifying a range of blocks to retrieve virtual operations for.
- [x] `account_history_api.get_account_history` - useful - Returns a history of all operations for a given account.
- [x] `account_history_api.get_ops_in_block` - useful - Returns all operations contained in a block.
- [x] `account_history_api.get_transaction` - useful - Returns the details of a transaction based on a transaction ID, including signatures, operations, and the block number it was included in.

## `block_api` (3/3 done)

- [x] `block_api.get_block` - useful - Retrieve a full, signed block of the referenced block, or null if no matching block was found.
- [x] `block_api.get_block_header` - useful - Retrieve a block header of the referenced block, or null if no matching block was found.
- [x] `block_api.get_block_range` - useful - Retrieve a range of full, signed blocks. The list may be shorter than requested if count blocks would take you past the current head block.

## `bridge` (24/24 done)

- [x] `bridge.account_notifications` - useful - Account notifications. (Supported by Hivemind)
- [x] `bridge.does_user_follow_any_lists` - useful - Checks if a given observer follows any blacklists or mute lists. (Supported by Hivemind)
- [x] `bridge.get_account_posts` - useful - Lists posts related to a given account. (Supported by Hivemind)
- [x] `bridge.get_community` - useful - Get community details. (Supported by Hivemind)
- [x] `bridge.get_community_context` - useful - Gets the role, subscription status, and title for a given account in a given community. (Supported by Hivemind)
- [x] `bridge.get_discussion` - useful - Gives a flattened discussion tree starting at given post. (Supported by Hivemind)
- [x] `bridge.get_follow_list` - useful - Returns blacklisted/muted accounts or list of blacklists/mute lists followed by a given observer. (Supported by Hivemind)
- [x] `bridge.get_payout_stats` - useful - Lists communities ordered by payout with stats (total payout, number of posts and authors). (Supported by Hivemind)
- [x] `bridge.get_post` - useful - Gives single selected post. (Supported by Hivemind)
- [x] `bridge.get_post_header` - useful - Gives very basic information on given post. (Supported by Hivemind)
- [x] `bridge.get_profile` - useful - Gets profile. (Supported by Hivemind)
- [x] `bridge.get_profiles` - useful - Gets a list of profiles. (Supported by Hivemind)
- [x] `bridge.get_ranked_posts` - useful - Get ranked posts. (Supported by Hivemind)
- [x] `bridge.get_relationship_between_accounts` - useful - Tells what relations connect given accounts from the perspective of first account. (Supported by Hivemind)
- [x] `bridge.get_trending_topics` - useful - Gets a list of trending communities. (Supported by Hivemind)
- [x] `bridge.list_all_subscriptions` - useful - List all subscriptions, titles, and roles to a community for an account. (Supported by Hivemind)
- [x] `bridge.list_communities` - useful - Gets community. (Supported by Hivemind)
- [x] `bridge.list_community_roles` - useful - List community roles and labels for each account in the community. (Supported by Hivemind)
- [x] `bridge.list_muted_reasons_enum` - useful - Gets a muted reasons enum map. (Supported by Hivemind)
- [x] `bridge.list_pop_communities` - useful - Gets a list of popular communities. (Supported by Hivemind)
- [x] `bridge.list_subscribers` - useful - Gets a list of subscribers for a given community. (Supported by Hivemind)
- [x] `bridge.normalize_post` - useful - Transforms legacy post objects into a new standardized format. (Supported by Hivemind)
- [x] `bridge.post_notifications` - useful - Gets a post notifications. (Supported by Hivemind)
- [x] `bridge.unread_notifications` - useful - Gets a count of unread notifications for an account. (Supported by Hivemind)

## `condenser_api` (79/79 done)

- [x] `condenser_api.broadcast_transaction` - useful - Used to broadcast a transaction.
- [x] `condenser_api.broadcast_transaction_synchronous` - useful - Used to broadcast a transaction and waits for it to be processed synchronously.
- [x] `condenser_api.find_proposals` - useful - Finds proposals by `proposal.id` (not `proposal.proposal_id`).
- [x] `condenser_api.find_rc_accounts` - useful - Returns the available resource credits of accounts.
- [x] `condenser_api.find_recurrent_transfers` - useful - Finds transfers of any liquid asset every fixed amount of time from one account to another.
- [x] `condenser_api.get_account_count` - useful - Returns the number of accounts.
- [x] `condenser_api.get_account_history` - useful - Returns a history of all operations for a given account.
- [x] `condenser_api.get_account_reputations` - useful - Returns a list of account reputations. (Supported by Hivemind)
- [x] `condenser_api.get_accounts` - useful - Returns accounts, queried by name.
- [x] `condenser_api.get_active_votes` - useful - Returns all votes for the given post. (Supported by Hivemind)
- [x] `condenser_api.get_active_witnesses` - useful - Returns the list of active witnesses.
- [x] `condenser_api.get_block` - useful - Returns a block.
- [x] `condenser_api.get_block_header` - useful - Returns a block header.
- [x] `condenser_api.get_blog` - useful - Returns the list of blog entries for an account. (Supported by Hivemind)
- [x] `condenser_api.get_blog_entries` - useful - Returns a list of blog entries for an account. (Supported by Hivemind)
- [x] `condenser_api.get_chain_properties` - useful - Returns the chain properties.
- [x] `condenser_api.get_collateralized_conversion_requests` - useful - Returns objects corresponding with collateralized_convert operations.
- [x] `condenser_api.get_comment_discussions_by_payout` - useful - Returns a list of discussions based on payout. (Supported by Hivemind)
- [x] `condenser_api.get_config` - useful - Returns information about compile-time constants.
- [x] `condenser_api.get_content` - useful - Returns the content (post or comment). (Supported by Hivemind)
- [x] `condenser_api.get_content_replies` - useful - Returns a list of replies. (Supported by Hivemind)
- [x] `condenser_api.get_conversion_requests` - useful - Returns a list of conversion request.
- [x] `condenser_api.get_current_median_history_price` - useful - Returns median history price.
- [x] `condenser_api.get_discussions_by_author_before_date` - useful - Returns a list of discussions based on author before date. (Supported by Hivemind)
- [x] `condenser_api.get_discussions_by_blog` - useful - Returns a list of discussions based on blog. (Supported by Hivemind)
- [x] `condenser_api.get_discussions_by_comments` - useful - Returns a list of discussions based on comments. (Supported by Hivemind)
- [x] `condenser_api.get_discussions_by_created` - useful - Returns a list of discussions based on created timestamp. (Supported by Hivemind)
- [x] `condenser_api.get_discussions_by_feed` - useful - Returns a list of discussions based on feed. (Supported by Hivemind)
- [x] `condenser_api.get_discussions_by_hot` - useful - Returns a list of discussions based on recent popularity. (Supported by Hivemind)
- [x] `condenser_api.get_discussions_by_trending` - useful - Returns a list of discussions based on popularity. (Supported by Hivemind)
- [x] `condenser_api.get_dynamic_global_properties` - useful - Returns the current dynamic global properties.
- [x] `condenser_api.get_escrow` - useful - Returns the escrow for a certain account by id.
- [x] `condenser_api.get_expiring_vesting_delegations` - useful - Returns the expiring vesting delegations for an account.
- [x] `condenser_api.get_feed_history` - useful - Returns the history of price feed values.
- [x] `condenser_api.get_follow_count` - useful - Returns the count of followers/following for an account. (Supported by Hivemind)
- [x] `condenser_api.get_followers` - useful - Returns the list of followers for an account. (Supported by Hivemind)
- [x] `condenser_api.get_following` - useful - Returns the list of accounts that are following an account. (Supported by Hivemind)
- [x] `condenser_api.get_hardfork_version` - useful - Returns the current hardfork version.
- [x] `condenser_api.get_key_references` - useful - Returns all accounts that have the key associated with their owner or active authorities.
- [x] `condenser_api.get_market_history` - useful - Returns the market history for the internal HBD:HIVE market.
- [x] `condenser_api.get_market_history_buckets` - useful - Returns the bucket seconds being tracked by the plugin.
- [x] `condenser_api.get_next_scheduled_hardfork` - useful - Returns the next scheduled hardfork.
- [x] `condenser_api.get_open_orders` - useful - Returns the open orders for an account.
- [x] `condenser_api.get_ops_in_block` - useful - Returns all operations contained in a block.
- [x] `condenser_api.get_order_book` - useful - Returns the internal market order book.
- [x] `condenser_api.get_owner_history` - useful - Returns the owner history of an account.
- [x] `condenser_api.get_post_discussions_by_payout` - useful - Returns a list of posts based on payout. (Supported by Hivemind)
- [x] `condenser_api.get_potential_signatures` - useful - This method will return the set of all public keys that could possibly sign for a given transaction.
- [x] `condenser_api.get_reblogged_by` - useful - Returns a list of authors that have reblogged a post. (Supported by Hivemind)
- [x] `condenser_api.get_recent_trades` - useful - Returns the most recent trades for the internal HBD:HIVE market.
- [x] `condenser_api.get_recovery_request` - useful - Returns the recovery request for an account.
- [x] `condenser_api.get_replies_by_last_update` - useful - Returns a list of replies to a comment. (Supported by Hivemind)
- [x] `condenser_api.get_required_signatures` - useful - This API will take a partially signed transaction and a set of public keys that the owner has the ability to sign for and return the minimal subset of public keys that should add signatures to the transaction.
- [x] `condenser_api.get_reward_fund` - useful - Returns information about the current reward funds.
- [x] `condenser_api.get_savings_withdraw_from` - useful - Returns savings withdraw from an account.
- [x] `condenser_api.get_savings_withdraw_to` - useful - Returns savings withdraw to an account.
- [x] `condenser_api.get_ticker` - useful - Returns the market ticker for the internal HBD:HIVE market.
- [x] `condenser_api.get_trade_history` - useful - Returns the trade history for the internal HBD:HIVE market.
- [x] `condenser_api.get_transaction` - useful - Returns the details of a transaction based on a transaction id.
- [x] `condenser_api.get_transaction_hex` - useful - Returns a hexdump of the serialized binary form of a transaction.
- [x] `condenser_api.get_trending_tags` - useful - Returns a list of trending tags. (Supported by Hivemind)
- [x] `condenser_api.get_version` - useful - Returns the versions of blockchain, hive, and FC.
- [x] `condenser_api.get_vesting_delegations` - useful - Returns the vesting delegations by an account.
- [x] `condenser_api.get_volume` - useful - Returns the market volume for the past 24 hours.
- [x] `condenser_api.get_withdraw_routes` - useful - Returns the withdraw routes for an account.
- [x] `condenser_api.get_witness_by_account` - useful - Returns the witness of an account.
- [x] `condenser_api.get_witness_count` - useful - Returns the witness count.
- [x] `condenser_api.get_witness_schedule` - useful - Returns the current witness schedule.
- [x] `condenser_api.get_witnesses` - useful - Returns current witnesses.
- [x] `condenser_api.get_witnesses_by_vote` - useful - Returns current witnesses by vote.
- [x] `condenser_api.is_known_transaction` - useful - Only return true if the transaction has not expired or been invalidated. If this method is called with a VERY old transaction we will return false, use account_history_api.get_transaction.
- [x] `condenser_api.list_proposal_votes` - useful - Returns all proposal votes, starting with the specified voter or `proposal.id`.
- [x] `condenser_api.list_proposals` - useful - Returns all proposals, starting with the specified creator or start date.
- [x] `condenser_api.list_rc_accounts` - useful - Find accounts and their RC delegations.
- [x] `condenser_api.list_rc_direct_delegations` - useful - Get list of “from” “to” which account how much RC was delegated.
- [x] `condenser_api.lookup_account_names` - useful - Looks up account names.
- [x] `condenser_api.lookup_accounts` - useful - Looks up accounts starting with name.
- [x] `condenser_api.lookup_witness_accounts` - useful - Looks up witness accounts starting with name.
- [x] `condenser_api.verify_authority` - useful - Returns true if the transaction has all of the required signatures.

## `database_api` (54/54 done)

- [x] `database_api.find_account_recovery_requests` - useful - Returns a list of account recovery requests.
- [x] `database_api.find_accounts` - useful - Returns accounts, queried by name.
- [x] `database_api.find_change_recovery_account_requests` - useful - Returns a list of requests to change the recovery account.
- [x] `database_api.find_collateralized_conversion_requests` - useful - Returns objects corresponding with collateralized_convert operations.
- [x] `database_api.find_comments` - useful - Search for comments by author/permlink. (Supported by Hivemind)
- [x] `database_api.find_decline_voting_rights_requests` - useful - Returns a list of requests to decline voting rights.
- [x] `database_api.find_escrows` - useful - Returns a list of escrows.
- [x] `database_api.find_hbd_conversion_requests` - useful - Returns the list of HBD conversion requests for an account.
- [x] `database_api.find_limit_orders` - useful - Returns a list of limit orders.
- [x] `database_api.find_owner_histories` - useful - Returns owner authority history.
- [x] `database_api.find_proposals` - useful - Finds proposals by `proposal.id`.
- [x] `database_api.find_recurrent_transfers` - useful - Finds transfers of any liquid asset every fixed amount of time from one account to another.
- [x] `database_api.find_savings_withdrawals` - useful - Returns the list of savings withdrawls for an account.
- [x] `database_api.find_vesting_delegation_expirations` - useful - Returns the expiring vesting delegations for an account.
- [x] `database_api.find_vesting_delegations` - useful - Returns the list of vesting delegations for an account.
- [x] `database_api.find_votes` - useful - Returns all votes for the given post. (Supported by Hivemind)
- [x] `database_api.find_withdraw_vesting_routes` - useful - Returns the list of vesting withdraw routes for an account.
- [x] `database_api.find_witnesses` - useful - Search for witnesses.
- [x] `database_api.get_active_witnesses` - useful - Returns the list of active witnesses.
- [x] `database_api.get_comment_pending_payouts` - useful - Get comment pending payout data.
- [x] `database_api.get_config` - useful - Returns information about compile-time constants. Some properties may not be present.
- [x] `database_api.get_current_price_feed` - useful - Returns the current price feed.
- [x] `database_api.get_dynamic_global_properties` - useful - Returns the current dynamic global properties.
- [x] `database_api.get_feed_history` - useful - Returns the history of price feed values.
- [x] `database_api.get_hardfork_properties` - useful - Returns the current properties about the blockchain’s hardforks.
- [x] `database_api.get_order_book` - useful - Returns the order book.
- [x] `database_api.get_potential_signatures` - useful - This method will return the set of all public keys that could possibly sign for a given transaction.
- [x] `database_api.get_required_signatures` - useful - Return the minimal subset of public keys that should add signatures to the transaction.
- [x] `database_api.get_reward_funds` - useful - Returns information about the current reward funds.
- [x] `database_api.get_transaction_hex` - useful - Returns a hexdump of the serialized binary form of a transaction.
- [x] `database_api.get_version` - useful - Returns the compile time versions of blockchain, hived, FC.
- [x] `database_api.get_witness_schedule` - useful - Returns the current witness schedule.
- [x] `database_api.is_known_transaction` - useful - Only return true if the transaction has not expired or been invalidated.
- [x] `database_api.list_account_recovery_requests` - useful - Returns a list of account recovery requests.
- [x] `database_api.list_accounts` - useful - List accounts ordered by specified key.
- [x] `database_api.list_change_recovery_account_requests` - useful - Returns a list of recovery account change requests.
- [x] `database_api.list_collateralized_conversion_requests` - useful - Returns objects corresponding with collateralized_convert operations.
- [x] `database_api.list_decline_voting_rights_requests` - useful - Returns a list of decline voting rights requests.
- [x] `database_api.list_escrows` - useful - Returns a list of escrows.
- [x] `database_api.list_hbd_conversion_requests` - useful - Returns the list of HBD conversion requests for an account.
- [x] `database_api.list_limit_orders` - useful - Returns a list of limit orders.
- [x] `database_api.list_owner_histories` - useful - Returns a list of limit orders.
- [x] `database_api.list_proposal_votes` - useful - Returns all proposal votes, starting with the specified voter or `proposal.id`.
- [x] `database_api.list_proposals` - useful - Returns all proposals, starting with the specified creator or start date.
- [x] `database_api.list_savings_withdrawals` - useful - Returns a list of savings withdrawls.
- [x] `database_api.list_vesting_delegation_expirations` - useful - Returns a list of vesting delegation expirations.
- [x] `database_api.list_vesting_delegations` - useful - Returns a list of vesting delegations.
- [x] `database_api.list_votes` - useful - Returns all votes, starting with the specified voter and/or author and permlink. (Supported by Hivemind)
- [x] `database_api.list_withdraw_vesting_routes` - useful - Returns a list of vesting withdraw routes.
- [x] `database_api.list_witness_votes` - useful - Returns a list of witness votes.
- [x] `database_api.list_witnesses` - useful - Returns the list of witnesses.
- [x] `database_api.verify_account_authority` - useful - Returns true if the keys are valid for the specified account.
- [x] `database_api.verify_authority` - useful - Verify transaction signatures.
- [x] `database_api.verify_signatures` - useful - This method validates if transaction was signed by person listed in required_owner, required_active or required_posting parameter.

## `debug_node_api` (11/11 done)

- [x] `debug_node_api.debug_generate_blocks` - useful - Generate blocks locally.
- [x] `debug_node_api.debug_generate_blocks_until` - useful - Generate blocks locally until a specified head block time. Can generate them sparsely.
- [x] `debug_node_api.debug_get_future_witness_schedule` - useful - Returns the future witness schedule.
- [x] `debug_node_api.debug_get_hardfork_property_object` - useful - Returns the current hardfork property object.
- [x] `debug_node_api.debug_get_head_block` - useful - Returns the current head block.
- [x] `debug_node_api.debug_get_json_schema` - useful - Returns the JSON schema.
- [x] `debug_node_api.debug_get_witness_schedule` - useful - Returns the witness schedule.
- [x] `debug_node_api.debug_has_hardfork` - useful - Returns true if the specified hardfork has been applied.
- [x] `debug_node_api.debug_set_hardfork` - useful - Sets the hardfork to the specified version.
- [x] `debug_node_api.debug_set_vest_price` - useful - Sets the price feed used to convert VESTS to HIVE.
- [x] `debug_node_api.debug_throw_exception` - useful - Throws an exception.

## `follow_api` (7/7 done)

- [x] `follow_api.get_account_reputations` - useful - Returns a list of account reputations. (Supported by Hivemind)
- [x] `follow_api.get_blog` - useful - Returns the list of blog entries for an account. (Supported by Hivemind)
- [x] `follow_api.get_blog_entries` - useful - Returns a list of blog entries for an account. (Supported by Hivemind)
- [x] `follow_api.get_follow_count` - useful - Returns the count of followers/following for an account. (Supported by Hivemind)
- [x] `follow_api.get_followers` - useful - Returns the list of followers for an account. (Supported by Hivemind)
- [x] `follow_api.get_following` - useful - Returns the list of accounts that are following an account. (Supported by Hivemind)
- [x] `follow_api.get_reblogged_by` - useful - Returns a list of authors that have reblogged a post. (Supported by Hivemind)

## `hive` (2/2 done)

- [x] `hive.db_head_state` - useful - Gets information about current headblock in the database.
- [x] `hive.get_info` - useful - Gets information about current status of the hivemind node.

## `jsonrpc` (2/2 done)

- [x] `jsonrpc.get_methods` - useful - Returns a list of methods supported by the JSON RPC API.
- [x] `jsonrpc.get_signature` - useful - Returns the signature information for a JSON RPC method including the arguments and expected response JSON.

## `market_history_api` (7/7 done)

- [x] `market_history_api.get_market_history` - useful - Returns the market history for the internal HBD:HIVE market.
- [x] `market_history_api.get_market_history_buckets` - useful - Returns the bucket seconds being tracked by the plugin.
- [x] `market_history_api.get_order_book` - useful - Returns the internal market order book.
- [x] `market_history_api.get_recent_trades` - useful - Returns the most recent trades for the internal HBD:HIVE market.
- [x] `market_history_api.get_ticker` - useful - Returns the market ticker for the internal HBD:HIVE market.
- [x] `market_history_api.get_trade_history` - useful - Returns the trade history for the internal HBD:HIVE market.
- [x] `market_history_api.get_volume` - useful - Returns the market volume for the past 24 hours.

## `network_broadcast_api` (1/1 done)

- [x] `network_broadcast_api.broadcast_transaction` - useful - Used to broadcast a transaction.

## `rc_api` (7/7 done)

- [x] `rc_api.find_rc_accounts` - useful - Returns the available resource credits of accounts.
- [x] `rc_api.get_rc_operation_stats` - useful - Returns rc operation statistics.
- [x] `rc_api.get_rc_stats` - useful - Returns rc statistics.
- [x] `rc_api.get_resource_params` - useful - Exports all relevant resource size constants, in particular the measurement-based execution time parameters.
- [x] `rc_api.get_resource_pool` - useful - Returns a list of all tracked pools.
- [x] `rc_api.list_rc_accounts` - useful - Returns a list of rc accounts.
- [x] `rc_api.list_rc_direct_delegations` - useful - Returns a list of rc delegations.

## `reputation_api` (1/1 done)

- [x] `reputation_api.get_account_reputations` - useful - Returns the reputation of accounts.

## `search_api` (1/1 done)

- [x] `search_api.find_text` - useful - Finds posts related to entry pattern.

## `tags_api` (10/10 done)

- [x] `tags_api.get_comment_discussions_by_payout` - useful - Returns a list of discussions based on payout. (Supported by Hivemind)
- [x] `tags_api.get_content_replies` - useful - Returns a list of replies. (Supported by Hivemind)
- [x] `tags_api.get_discussion` - useful - Returns the content (post or comment). (Supported by Hivemind)
- [x] `tags_api.get_discussions_by_author_before_date` - useful - Returns a list of discussions based on author before date. (Supported by Hivemind)
- [x] `tags_api.get_discussions_by_blog` - useful - Returns a list of discussions based on blog. (Supported by Hivemind)
- [x] `tags_api.get_discussions_by_comments` - useful - Returns a list of discussions based on comments. (Supported by Hivemind)
- [x] `tags_api.get_discussions_by_created` - useful - Returns a list of discussions based on created timestamp. (Supported by Hivemind)
- [x] `tags_api.get_discussions_by_hot` - useful - Returns a list of discussions based on recent popularity. (Supported by Hivemind)
- [x] `tags_api.get_discussions_by_trending` - useful - Returns a list of discussions based on popularity. (Supported by Hivemind)
- [x] `tags_api.get_post_discussions_by_payout` - useful - Returns a list of posts based on payout. (Supported by Hivemind)

## `transaction_status_api` (1/1 done)

- [x] `transaction_status_api.find_transaction` - useful - Returns the status of a given transaction id.
