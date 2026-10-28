# Changelog

Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## Unreleased

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
