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
- [ ] Add table-driven handler/usecase tests for post, reaction and comment ownership.
- [x] Enforce one reaction per user/post/type with a Mongo unique index and duplicate-safe API response (integration tested).
- [ ] Return author summary and aggregate reaction/comment counts with feed posts to remove client N+1 calls.
- [ ] Add post visibility rules (public, followers, group-only, private) and authorization tests.
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
- [ ] Add follower/following counts and privacy-aware follow access rules.
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
