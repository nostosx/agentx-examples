# CLI generation and application extensions

This repository is one Go module created from the nostosx-cli v1.4.0 with the
monolith layout, agentx support and only HTTP. Normal builds need no sibling source.
Install CLI v1.4.0 and set NOSTOSX to the absolute nostosx executable path.

## Create the foundation

Use a new destination; scaffolding refuses overwrites:

```sh
"$NOSTOSX" new agentx-examples --module github.com/nostosx/agentx-examples --with agentx --layout monolith --transport http --output ./new-example
"$NOSTOSX" add context assistant --dir ./new-example
"$NOSTOSX" add service assistant --dir ./new-example
```

The generated foundation provides HTTP configuration and server constructors, a context, service app.go,
private HTTP adapter, standalone service entry, Taskfile and CI. It contains no
Proto, Buf configuration, gRPC listener or contract tasks. Source belongs to the
application and is never fetched again at startup.

## Application extensions

- Six CLI demonstrations retain their command names. Implementations live in
  application/basic, tools, streaming, skills, multiagent and workflow.
- Infrastructure owns AGENTX_PROVIDER/AGENTX_MODEL configuration, cancellation,
  all three model providers and trusted local/embedded skills. Actual demos
  replace the unused generic generated Runner and its tests.
- application/assistant assembles the approval-gated release tool. The service
  owns filestore sessions/checkpoints and the native agentx HTTP protocol server.
- Root app.go owns the monolith Run entry and selects assistant.Run for this
  single-service host. The standalone entry calls assistant.Run directly; both
  reuse services/assistant/app.go and the generated HTTP constructors.
- HTTP adapters implement authentication, CORS and owner cancellation.
- Write timeouts are disabled for streaming. Owner cancellation reaches model
  requests before listener shutdown waits. API_TOKEN authenticates Agent routes;
  -addr replaces ADDR and only literal loopback listeners are permitted.

## Verify

```sh
task tidy
task check lint vulncheck
```

No Buf, generation step, smoke job, Python, extra module or go.work is required.
CI checks the actual application on Linux/macOS/Windows. Tests use temporary local
listeners and model fixtures, without paid APIs or shared stores.

Current CLI templates keep process entries limited to configuration, signal context
and Run. Monolith assembly lives in root app.go; standalone assembly lives in
services/<name>/app.go. The CLI no longer generates AGENTS.md, CONTRIBUTING.md,
README.md, README.zh-CN.md, CHANGELOG.md, SECURITY.md or LICENSE. This example
repository maintains its own documentation and license.

The current scaffold defaults to contrib Viper and contrib Zap. config.Load
merges defaults, a process TOML file, environment and explicit flags; app.go constructs,
injects and flushes the Logger. The example uses the same loading/logging files,
with no hand-written configuration engine or global logging backend.

The process configuration layout matches CLI v1.4.0: config/config.toml for
the monolith and services/assistant/config/config.toml for the standalone entry.
Each config.Load call explicitly supplies its default path between name and args.
Paths are relative to the working directory; files are not merged across processes.
Existing YAML-based applications need an explicit migration.
