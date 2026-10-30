# Changelog

Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## v1.5.0 - 2026-10-31

- Name the service-private application type Command in command.go and keep it separate from protocol adapters. Construct individual use cases in app.go, compose CQRS chains in handler aggregates, and inject handler aggregates into protocol boundaries.

- Route the approval-gated deployment tool through an injected servex/cqrs command assembled in service app.go; validate environments and preserve cancellation and approval boundaries.

## v1.4.0 - 2026-10-29

- **Breaking:** Move monolith configuration to config/config.toml and standalone service
  configuration to services/assistant/config/config.toml. Use independent TOML
  files with an explicit default path in config.Load; preserve defaults/file/env/flag
  precedence and --config overrides. Convert precedence fixtures to TOML and
  verify process isolation and both checked-in configuration files.

## v1.3.0 - 2026-10-29

- Adopt published servex v1.4.0 and independent contrib HTTP, configuration
  and logging adapters; resolve dependencies without source replacements.

- Use readable console text and separate JSON file encoding with timestamps,
  uppercase levels and caller locations. Preserve structured fields, support
  log_file/APP_LOG_FILE/--log-file and rotate optional files with lumberjack.

- Use contrib Viper for defaults/file/env/flag configuration and contrib Zap for
  readable console logs and optional JSON files. Inject the core Logger and flush after shutdown; retain
  API_TOKEN while supporting APP_TOKEN and configurable log levels.

- Add the monolith Run entry in root app.go. cmd/server delegates to it, while
  the standalone entry calls assistant.Run. Both reuse the service assembly.

- Add the unreleased servex-contrib/transport/client module required by the
  matching HTTP/gRPC backends. Refresh its actual version and checksum during
  the coordinated release.

- Migrate the process HTTP host to servex-contrib/transport/httpx.
  This source revision requires matching unreleased core and contrib backends;
  public-proxy consumer verification remains a coordinated-release gate.

## v1.2.0

Release date: 2026-10-28.

- Consolidate public Task commands to check/test/build/lint/vulncheck/fmt/fix/tidy;
  race and coverage use `task test RACE=true COVER=true`, with checks kept internal.

- Use only the native HTTP transport for Agent protocols. Remove the empty
  AssistantService, gRPC listener/dependencies, Proto artifacts, Buf configuration
  and contract CI/tasks. Preserve authentication, cancellation and streaming.

- Adopt one actual CLI-generated application per repository: domain/application,
  shared infrastructure, service app.go files and private adapters. See CLI-GENERATION.md for generation and extensions.
- Move six agent demonstrations out of main packages and reuse generated
  transport constructors for the assistant protocol service. Select agentx v1.2.1,
  provider v1.2.0, filestore v1.0.0 and servex v1.3.0.
- **Breaking (example startup):** API_TOKEN replaces AGENTX_SERVER_TOKEN; use
  -addr instead of ADDR. The HTTP default is 8080 and the listener requires
  a literal loopback IP. There is no gRPC listener.

- Use config.Parse in commands and service.Run(ctx, cfg) in app.go for shared
  assembly. Register only mounts handlers; server.New only constructs transports.

- Development tooling: replace Make entry points with Task v3.53.1 in local
  checks, CI and contribution/release instructions. Use `task check`, `task lint`
  and `task vulncheck`; check coverage, build tags and failure gates are preserved.

## v1.1.0

Release date: 2026-10-27.

- Require published agentx v1.1.0 and agentx-contrib provider v1.1.0;
  filestore stays at v1.0.0. The programs need no source changes.
- Document compatible endpoint configuration and a first authenticated server
  request, including approval, resume and local state expectations.

## v1.0.0

Release date: 2026-10-23.

### Added

- Seven runnable applications for basic calls, tools, streaming, skills,
  multi-agent coordination, workflows and an authenticated local HTTP server.
- Local regression tests and Makefile checks for Go 1.27.3, with independent
  cross-platform CI, lint and vulnerability jobs.

### Dependencies

- Use the published kitx v1.0.0 through agentx and filestore.
- Require published agentx v1.0.0 and the following contrib modules at v1.0.0:
  `github.com/nostosx/agentx-contrib/provider`,
  `github.com/nostosx/agentx-contrib/filestore`.
- Resolve dependencies through the Go module proxy without sibling checkouts
  or local replacements.
