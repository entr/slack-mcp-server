# entr/slack-mcp-server

Personal fork of [korotovsky/slack-mcp-server](https://github.com/korotovsky/slack-mcp-server).
Upstream has not merged anything since v1.3.0 (2026-05-14), so this fork
collects open upstream PRs and fixes from other forks that matter for local
stdio use with browser session tokens (xoxc/xoxd) in a large workspace.

There is no npm package. Build from source:

```sh
go build -o build/slack-mcp-server ./cmd/slack-mcp-server
```

## Base: upstream PRs (via [Seylox](https://github.com/Seylox/slack-mcp-server))

| PR | Change |
|---|---|
| #233 | Resolve user mentions in message text to display names |
| #321 | `oldest` parameter for `conversations_history` / `conversations_replies` |
| #335 | Thread reply count and latest reply ts on messages |
| #338 | Retry and checkpoint the users cache walk |
| #344 | gofmt-clean tree |
| #347 | `conversations_threads` tool (Slack "Threads" view) |
| #361 | Keep block text when a message also has a text fallback |
| #367 | Scope session cookies to HTTPS Slack hosts; parse insecure TLS flag as a boolean |

## Wave 1: cherry-picks

| Source | Change |
|---|---|
| cyberpapiii `33e850b` | Slack Connect enrichment is best-effort in the users cache |
| cyberpapiii `dc6db10` | CAS-safe `PatchUser` snapshot update |
| cyberpapiii `0a4dc06` | Edge 429 retry replays the request body; guard nil responses |
| provectus `613169b` | Don't persist a partial channel list on pagination failure |
| cyberpapiii `5c936ec` | Fix `UsersList` data race; share one Tier 3 limiter |
| cyberpapiii `5c59667`, `727b517` | Tool-contract fixes: search limit, IM/MPIM filter by id, `saved_update date_due:0`, duration limit with cursor |
| cupskeee `b010d07` (#303) | `conversations_update_message` tool |
| cupskeee #339 | `conversations_add_message` Markdown dialect docs and `content_type` enum |
| cupskeee `7d688df` (#323) | `conversations_info` tool (last read, unread count) |
| cupskeee `aa31053` | Docs/wording reconcile for the above |

## Wave 2: manual ports

| Source | Change |
|---|---|
| PiotrMackowski `a0fded6` | `edgeFailed` as `atomic.Bool` |
| cyberpapiii `fc3d29a` | Deadline on background users/channels cache refreshes |
| cyberpapiii `17b7f47`, `573a9a1` | Process-wide shared rate limiters; own bucket for `search.messages` |
| cyberpapiii `a5919d9`, `277faf5` | Stop rendering block text the message text already holds |
| cyberpapiii `9dc0eb3` | Unreads: skip the redundant history backfill |
| cyberpapiii `57818d5`, `9d9a573` | Serve stdio while caches warm; one `auth.test` at startup |
| cyberpapiii `784d121` (concept) | Revalidate users/channels caches past their TTL at runtime |

## Fork-only

- Tool calls made while the caches warm wait up to 30s (`AwaitReady`)
  instead of failing immediately.
- Block-text dedupe ignores emoji form, so a message whose text has `✅` and
  whose section block has `:white_check_mark:` is not rendered twice.

## Candidates not yet adopted

- Search `has:` modifiers and `sort=timestamp` (cyberpapiii `2b832e1`)
- Markdown to native rich_text when posting (provectus `a650672`, `96f716d`)

## Syncing with upstream

```sh
git fetch upstream
git rebase upstream/master   # or merge; resolve, then:
go build ./... && go vet ./... && go test -race ./... -skip Integration
```

Integration tests need real Slack/OpenAI tokens and are skipped locally.
