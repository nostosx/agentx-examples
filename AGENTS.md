# agentx-examples development

Parent workspace instructions apply. Always load `golang-how-to` for Go work,
then the relevant Go skills. Read README.md, CONTRIBUTING.md and
.github/workflows/ci.yml before editing.

- Go 1.27.3; one CLI-generated module. cmd/ holds thin entries; application/ owns
  example behavior, infrastructure/ owns configuration/lifecycle, and services/
  holds app.go, private adapters and standalone service entries.
- Root app.go selects the monolith composition. Commands parse configuration,
  create a signal context and call Run; keep assembly in root or service app.go.
- Compose public APIs with explicit parameters. Do not add a DI container or
  move application wiring back into the libraries.
- Core, contrib and kitx dependencies use published versions without
  local replacements. Keep standalone builds reproducible with GOWORK=off;
  do not commit go.work files or sibling checkout requirements.
- Code comments are English. Preserve package-level Example tests in libraries.
- `task check` checks formatting, modernizer/tidy diffs, vet, builds and race
  tests with randomized order. `task lint` uses golangci-lint v2.15.0 with
  Go 1.27.3; `task vulncheck` uses govulncheck. `task test` is ordinary
  testing; `task test RACE=true` runs the race suite and `COVER=true` adds coverage. `task fmt`, `task fix` and
  `task tidy` modify files. Only check/test/build/lint/vulncheck/fmt/fix/tidy are public; bare task runs check.
  See RELEASING.md for dependency/snapshot checks.
- CI does not call paid model APIs. Run real providers only when explicitly
  requested; never store credentials or local session data.

- Follow CLI-GENERATION.md for the actual new/context/service generation recipe.
  Keep domain/application free of generated contracts and transport imports.
  This application only enables HTTP. Do not add empty Proto/gRPC services, Buf
  configuration or contract gates. Do not add smoke or Python tooling.

- Default process configuration uses contrib config/viper, including config/config.toml,
  APP_ overrides and explicit flags. app.go constructs contrib logger/zap console/file
  logging and injects the core Logger; keep globals and concrete backends out of domain/application.

- TOML process files are independent: config/config.toml for cmd/server and
  services/assistant/config/config.toml for the standalone service. Entry points pass
  their default path to config.Load; preserve file/env/flag precedence.

The application layout separates business contexts from deployable services. Service-private `internal/application/command.go` declares `Command`, which assembles typed CQRS chains and metadata. `app.go` constructs dependencies through `CommandDeps` and injects the command handler into the tool adapter.
