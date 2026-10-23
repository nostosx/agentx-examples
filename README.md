# agentx-examples

English | [简体中文](README.zh-CN.md)

Runnable applications composing the public APIs of agentx and agentx-contrib.
This is a separate Go 1.27.3 module, `github.com/nostosx/agentx-examples`.
Library godoc Example tests remain with their packages.

The v1.0.0 source snapshot uses published v1.0.0 library dependencies.
Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## Local setup

Use Go 1.27.3. Clone this repository on its own; dependencies download from
the Go module proxy, with no sibling checkouts or local replacements:

```sh
git clone https://github.com/nostosx/agentx-examples.git
cd agentx-examples
GOWORK=off go mod download
```

Run from this directory with `GOWORK=off` (the Makefile already sets it).
`AGENTX_PROVIDER` selects `anthropic`, `openai` or `gemini`. Without it, the
first configured API key is used, in that order. Set the corresponding
`ANTHROPIC_API_KEY`, `OPENAI_API_KEY` or `GEMINI_API_KEY` (`GOOGLE_API_KEY`
also works for Gemini). `AGENTX_MODEL` overrides the provider default.

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

The server listens on `127.0.0.1:8080`. `ADDR` may select another loopback
address or port. Set `AGENTX_SERVER_TOKEN` to a random secret without whitespace
before starting it, and send `Authorization: Bearer <token>` with every API
request. The session, approval, AI SDK, AG-UI and A2A routes share this token.
This demonstrates one local trust domain; it does not isolate users or tenants.
Browser clients at `http://localhost:3000` can preflight requests and must also
send the token. Keep it out of client bundles and URLs. On SIGINT/SIGTERM, the
server cancels active model calls, allows ten seconds for HTTP shutdown and
closes remaining connections. Request examples are in [cmd/server/main.go](cmd/server/main.go).

## Checks and dependency versions

```sh
make check
make lint
make vulncheck
```

CI builds all seven programs and runs regression tests without calling real
model providers. It resolves agentx, provider and filestore v1.0.0, plus kitx
v1.0.0, from published modules without sibling checkouts or local replacements.
The libraries do not depend on this project. Dependency updates must pass a
standalone build before a new source snapshot is published. There is no
automatic publishing workflow.
See [CONTRIBUTING.md](CONTRIBUTING.md) for tool installation
and check details, [RELEASING.md](RELEASING.md) for reproducible snapshots and
dependency updates, [CHANGELOG.md](CHANGELOG.md) for changes and
[SECURITY.md](SECURITY.md) for private reports. Licensed under the [MIT License](LICENSE).
