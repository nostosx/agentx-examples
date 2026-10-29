# agentx-examples

This snapshot uses published servex v1.4.0 and the independent contrib
HTTP/configuration/logging modules, without sibling source replacements.

English | [简体中文](README.zh-CN.md)

Runnable applications composing the public APIs of agentx and agentx-contrib.
This is a separate Go 1.27.3 module, `github.com/nostosx/agentx-examples`.
Library godoc Example tests remain with their packages.

Snapshot v1.4.0 uses agentx v1.2.1, provider v1.2.0, filestore v1.0.0 and
servex v1.4.0. The v1.1.0 tag retains its historical dependencies.
Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## CLI-generated application

This repository is one monolith-layout CLI project with optional agent support.

```text
app.go                       monolith Run; reuses the single service assembly
config/config.toml           monolith process settings
services/assistant/config/config.toml  standalone service settings
cmd/                         thin command entries; cmd/server calls application.Run
domain/assistant/            domain declarations
application/                 agent demonstrations and release-assistant assembly
infrastructure/              model/skill configuration, command cancellation, server construction
services/assistant/          app.go, private HTTP adapters and cmd/assistant
```

The six command-line entries retain their commands and delegate to application
packages. Provider selection and trusted skills live in infrastructure.
The monolith entry parses configuration and calls root `app.go`'s `Run(ctx, cfg)`,
which reuses `assistant.Run`. The standalone entry calls `assistant.Run` directly.
The service app.go owns file-backed sessions, protocol cleanup and explicit
handler registration with the shared transport constructors. Session, AI SDK, AG-UI and A2A
use agentx's native HTTP contracts. No Proto, gRPC listener or contract tooling is included.

Run `go run ./services/assistant/cmd/assistant` instead of `./cmd/server` for the
standalone service; both use the same flags, token and module. See
[CLI-GENERATION.md](CLI-GENERATION.md) for the actual generation recipe and extensions.

Server configuration is loaded by contrib Viper through infrastructure/config.Load:
defaults < TOML file < environment < explicit flags. The default file is optional;
--config selects a required file. APP_ADDRESS, APP_LOG_LEVEL and APP_TOKEN override
file settings; APP_TOKEN takes precedence over API_TOKEN. --log-level selects the Zap level (default info).
Monoliths load `config/config.toml`; the standalone service loads
`services/assistant/config/config.toml`. Run from the repository root because
paths are relative to the working directory. These files are independent; neither
entry merges the other process's configuration. Use --config to select a deployment path.
Both entries keep the same default ports; configure distinct addresses when running them together.

Composition roots create contrib Zap console text and optional JSON file logging, inject the core Logger into
the application and transport boundaries, and flush after shutdown.

Console records render modules and fields as [Module] message [key:value],
with [操作:…], [耗时:…], [状态:…] and [错误:…] summaries. Newlines and tabs
in messages/fields are escaped so one event occupies one physical line.
An explicit component selects the module; otherwise it uses the caller package.
Console records use local timestamps in YYYY-MM-DD HH:mm:ss format, uppercase levels,
short caller locations and readable durations. File records use one JSON object
per line, the same second-precision local timestamps and numeric durations in seconds. Business fields
remain separate in both outputs. Set log_file in the process TOML file, APP_LOG_FILE or
--log-file to enable JSON file output; an explicit empty --log-file disables it.
The default is console only. Enabled files rotate at 100 MiB, retain at most five
backups for up to seven days, and compress backups with lumberjack v2.2.1.
Give each service process its own file path; close the logger after its users stop.

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
