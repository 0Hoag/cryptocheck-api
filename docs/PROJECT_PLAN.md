# CryptoCheck — API implementation plan

> Status convention: `[x]` = implemented and verified; `[ ]` = not complete.
> Every completed task must include automated checks and a short evidence note in
> its section before being marked `[x]`.

## Operating rules

- Make one focused change set per commit; push it after its checks pass.
- Do not mark an item complete from compilation alone when it changes a user flow.
- For every protected endpoint, verify both an authenticated and unauthenticated case.
- Keep secrets only in `.env`; never commit them.

## 0. Local foundation

- [x] MongoDB + RabbitMQ local Docker dependencies can be started from `make run-api`.
- [x] RabbitMQ is optional for local API startup and no longer crashes the server.
- [x] Scanner explorer keys were moved from the legacy local environment to the new local `.env` (not committed).
- [ ] Replace the Gemini API key reported by Google as leaked, then verify AI analysis returns JSON instead of regex fallback.

Pass checks:

```bash
make run-api
curl -fsS http://localhost:8080/swagger/index.html >/dev/null
```

## 1. Scanner reliability

- [x] Resolve a token symbol through DexScreener and keep the resolved supported network.
- [x] Prefer exact token symbol/address/name matches over unrelated high-liquidity pairs.
- [x] Prevent native coins such as BNB/ETH/BTC from being treated as contracts in the client.
- [x] ENA integration check: resolves Ethereum contract and returns a score.
- [ ] Add scanner integration tests with mocked DexScreener and explorer responses.
- [ ] Add a persisted scan history with owner, input, network, score, engine version and timestamps.
- [ ] Define and enforce Free/Premium quotas and analysis depth server-side.

### 1.1 Multi-market and early-launch coverage

- [x] Phase 1: return explicit `contract`, `native_asset`, and `market_asset` modes; never present a contract-security score when no source scan exists. (scanner package tests + ENA smoke test passed)
- [x] Phase 1: expand DexScreener discovery beyond the supported EVM explorers and retain market liquidity/volume for unsupported chains. (ENA smoke test passed)
- [x] Phase 2: expose the five strongest de-duplicated matches so the client can ask users to choose the chain when a symbol is ambiguous. (BONK candidates smoke test passed)
- [x] Add a first chain-specific Solana SPL mint analyzer (mint/freeze authority only); label it as limited on-chain analysis, not a full program audit. (BONK candidate smoke test passed)
- [ ] Add chain-specific analyzers for other high-volume non-EVM chains before claiming source-code coverage.
- [x] Add market metadata (DEX pair URL, liquidity, 24h volume, pair creation time, source/provider) and a market-data confidence level to every market-only result. (ENA smoke test passed)
- [x] Add a prelaunch/watchlist record for projects without a deployed contract: public list/detail plus authenticated create and owner-only update/delete; record project URL, official socials, claimed chain, launch date, verification evidence and risk flags. (prelaunch and HTTP server packages passed)
- [ ] Integrate a launch-calendar/presale provider only after its API terms, rate limits and data attribution are reviewed; cache results and record the provider timestamp.
- [ ] Add background discovery/monitoring for new listings and contract deployments, with deduplication and alert thresholds.

Acceptance rule: do not claim “all coins” coverage. Results must clearly state whether they are contract analysis, native-asset profile, market-data profile, or prelaunch due diligence, plus the data freshness and source coverage.

Pass checks:

```bash
go test ./internal/adapters/dexscreener ./internal/scanner/usecase ./internal/httpserver
curl -fsS 'http://localhost:8080/api/v1/news-feed/scanner?token=ENA&lang=vi'
```

## 2. Social API correctness

- [x] Post CRUD endpoints are available with authenticated write access.
- [x] Reaction creation no longer recurses indefinitely.
- [x] Reaction `type` is persisted.
- [x] Reaction and comment list endpoints filter by `post_id` correctly.
- [x] Audit flow passed: create post → like → comment → list one reaction/comment → delete all test data.
- [x] Post feed accepts a validated `author_id` filter for profile/community views.
- [x] Post update and delete enforce ownership and return 403 for another user's post.
- [x] Comment update/delete enforce ownership, return the persisted update, and return 403 for another user (integration tested).
- [ ] Add table-driven handler/usecase tests for post, reaction and comment ownership.
- [x] Enforce one reaction per user/post/type with a Mongo unique index and duplicate-safe API response (integration tested).
- [x] Return aggregate reaction/comment counts with feed posts to remove client N+1 calls (integration tested).
- [x] Return sanitized author summaries with feed posts (integration tested; no phone exposure).
- [x] Enforce public and private (`justme`) post visibility for guest, owner and other members (integration tested).
- [ ] Add followers-only and group-only post visibility with authorization tests.
- [ ] Add pagination, sorting and rate limits appropriate for a public community.

Pass checks:

```bash
go test ./internal/post/... ./internal/comment/...
# Manual smoke test: create post -> like -> comment -> list by post_id -> delete test data
```

## 3. Community and groups

- [ ] Add `groups` domain: group, membership, role (owner/admin/mod/member), join policy and visibility.
- [ ] Add group post feed and moderation controls.
- [x] Add authenticated follow/unfollow and author-filtered profile feed endpoints (integration tested).
- [x] Add a public aggregate follower/following counts endpoint without exposing individual follow relationships. (HTTP handler test passed)
- [ ] Define privacy-aware access rules for individual follower/following lists.
- [ ] Add notifications for reactions, comments, follows and group events.
- [ ] Add reporting/moderation audit trail for posts and comments.

Pass checks:

```bash
go test ./internal/group/... ./internal/notification/...
```

## 4. Premium entitlement

- [ ] Model plans, subscriptions, entitlement periods and payment-provider references.
- [ ] Add middleware/service to evaluate entitlements; never trust a client-side Premium flag.
- [ ] Gate private group creation and enhanced scanner quotas/depth with entitlement checks.
- [ ] Choose payment provider and implement webhook signature verification before enabling billing.
- [ ] Add cancellation, expiry and downgrade behavior tests.

## 5. Production hardening

- [ ] Set release-mode logging, trusted proxy configuration and restrictive CORS origins.
- [ ] Add request validation/error contract tests and structured observability.
- [ ] Add CI for `go test ./...`, formatting and vulnerability checks.
- [ ] Document environment variables and deployment runbook without secrets.
