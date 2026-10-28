# agentx-examples

English | [简体中文](README.zh-CN.md)

Runnable applications composing the public APIs of agentx and agentx-contrib.
This is a separate Go 1.27.3 module, `github.com/nostosx/agentx-examples`.
Library godoc Example tests remain with their packages.

Snapshot v1.2.0 uses agentx v1.2.1, provider v1.2.0, filestore v1.0.0 and
servex v1.3.0. The v1.1.0 tag retains its historical dependencies.
Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## CLI-generated application

This repository is one monolith-layout CLI project with optional agent support.

```text
cmd/                         thin command entries; cmd/server calls assistant.Run
domain/assistant/            domain declarations
application/                 agent demonstrations and release-assistant assembly
infrastructure/              model/skill configuration, command cancellation, server construction
services/assistant/          app.go, private HTTP adapters and cmd/assistant
```

The six command-line entries retain their commands and delegate to application
packages. Provider selection and trusted skills live in infrastructure. Both server entries parse configuration and call `assistant.Run(ctx, cfg)`.
The service app.go owns file-backed sessions, protocol cleanup and explicit
handler registration with the shared transport constructors. Session, AI SDK, AG-UI and A2A
use agentx's native HTTP contracts. No Proto, gRPC listener or contract tooling is included.

Run `go run ./services/assistant/cmd/assistant` instead of `./cmd/server` for the
standalone service; both use the same flags, token and module. See
[CLI-GENERATION.md](CLI-GENERATION.md) for the actual generation recipe and extensions.

## Local setup

Use Go 1.27.3. Clone this repository on its own; dependencies download from
the Go module proxy, with no sibling checkouts or local replacements:

```sh
git clone https://github.com/nostosx/agentx-examples.git
cd agentx-examples
GOWORK=off go mod download
```

Run from this directory with `GOWORK=off` (the Taskfile.yml already sets it).
`AGENTX_PROVIDER` selects `anthropic`, `openai` or `gemini`. Without it, the
first configured API key is used, in that order. Set the corresponding
`ANTHROPIC_API_KEY`, `OPENAI_API_KEY` or `GEMINI_API_KEY` (`GOOGLE_API_KEY`
also works for Gemini). `AGENTX_MODEL` overrides the provider default.
The corresponding `ANTHROPIC_BASE_URL`, `OPENAI_BASE_URL` or `GEMINI_BASE_URL`
can select a trusted compatible endpoint. Credentials are sent to that endpoint;
leave it unset to use the provider's public API. OpenAI uses the Responses API
by default, so a Chat Completions-only endpoint is not sufficient for these commands.

```sh
export AGENTX_PROVIDER=anthropic
# Set ANTHROPIC_API_KEY in your environment.
GOWORK=off go run ./cmd/basic "What is an agent loop?"
```

## Programs

| Command | Demonstrates |
|---|---|
| `go run ./cmd/basic` | Single-turn Q&A |
| `go run ./cmd/tools` | Typed tools, hooks and a demo approval policy |
| `go run ./cmd/streaming` | Streaming output and multi-turn sessions; `/exit` ends the chat |
| `go run ./cmd/skills` | Embedded skills; `-skills-dir .skills` opts into trusted local skills |
| `go run ./cmd/multi-agent` | A coordinator delegating to specialists |
| `go run ./cmd/workflow` | Parallel drafts, review and conditional editing |
| `go run ./cmd/server` | HTTP protocols and approval, with `.agentx/` file-backed state |

Commands use real model APIs. The tools and deployment actions are demonstrations;
apply your own authorization policy when adapting them. Local skills may register
executable hooks, so only pass a directory you trust to `-skills-dir`.
The six command-line examples stop on SIGINT/SIGTERM and have a five-minute
deadline (per turn for streaming). Input and output failures end the chat.

The host only listens on `127.0.0.1:8080`.
Use `-addr` to select another literal loopback address.
Set `API_TOKEN` to at least 32 printable ASCII characters without whitespace
before starting it, and send `Authorization: Bearer <token>` with every API
request. The session, approval, AI SDK, AG-UI and A2A routes share this token.
This demonstrates one local trust domain; it does not isolate users or tenants.
Browser clients at `http://localhost:3000` can preflight requests and must also
send the token. Keep it out of client bundles and URLs. On SIGINT/SIGTERM, the
server cancels active model calls, allows ten seconds for HTTP shutdown and
closes remaining connections. Health routes are public; agent protocol routes
require the token. The HTTP write deadline is disabled for streaming; active model
requests are tied to the application cancellation context.

For a first server request, keep the provider configuration above and start
the server in one terminal:

```sh
export API_TOKEN="$(openssl rand -hex 32)"
GOWORK=off go run ./cmd/server
```

In another terminal, set `API_TOKEN` to the same value and run:

```sh
curl -N http://127.0.0.1:8080/sessions/demo/runs \
  -H "Authorization: Bearer $API_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"input":"Deploy the latest build to staging"}'
curl http://127.0.0.1:8080/sessions/demo \
  -H "Authorization: Bearer $API_TOKEN"
```

The run creates session `demo` automatically. If the model calls `deploy`,
the event stream pauses for approval. Inspect the session's `pending` field,
then POST `{"approvals":{"<call_id>":{"approved":true}}}` to
`/sessions/demo/resume` with the same Authorization header and returned call ID.
The deployment tool only returns a message; it does not deploy a real build.
State is stored under `.agentx/` relative to the repository root. Use Ctrl+C
to stop the server.

## Checks and dependency versions

```sh
task check
task lint
task vulncheck
```

The eight common tasks are `check`, `test`, `build`, `lint`, `vulncheck`, `fmt`,
`fix` and `tidy`; bare `task` runs `check`. Use `task --list` to view them,
`task test RACE=true` for race detection, and `task test RACE=true COVER=true`
for race coverage in `coverage.out`. Check helpers are internal.

CI builds the seven demonstrations and the standalone assistant entry, and runs
regression tests without real model calls. Dependencies resolve from published
modules without sibling checkouts or local replacements.
The libraries do not depend on this project. Dependency updates must pass a
standalone build before a new source snapshot is published. There is no
automatic publishing workflow.
This HTTP-only application needs no Buf, Proto generation or contract CI job.
There are no smoke tasks or Python tooling.
See [CONTRIBUTING.md](CONTRIBUTING.md) for tool installation
and check details, [RELEASING.md](RELEASING.md) for reproducible snapshots and
dependency updates, [CHANGELOG.md](CHANGELOG.md) for changes and
[SECURITY.md](SECURITY.md) for private reports. Licensed under the [MIT License](LICENSE).
