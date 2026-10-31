# Publishing agentx-examples

Tags identify reproducible source snapshots of runnable examples. They do not
establish a stable library API or include prebuilt binaries. Go 1.27.3 is the
toolchain requirement, independent of the snapshot version. There is no
automatic release workflow.

## Published dependencies

The v1.6.0 snapshot uses these released modules, with no local replacements:

| Module | Version |
|---|---|
| `github.com/nostosx/agentx` | v1.4.0 |
| `github.com/nostosx/agentx-contrib/provider` | v1.3.0 |
| `github.com/nostosx/agentx-contrib/filestore` | v1.1.0 |
| `github.com/nostosx/servex` | v1.5.0 |
| `github.com/nostosx/kitx` | v1.0.0 |
| `github.com/nostosx/servex-contrib/config/viper` | v1.0.1 |
| `github.com/nostosx/servex-contrib/logger/zap` | v1.0.0 |
| `github.com/nostosx/servex-contrib/transport/httpx` | v1.1.0 |

The PostgreSQL example additionally pins these published capabilities:

| Module | Version |
|---|---|
| agentx-contrib/retrieval | v1.2.0 |
| servex-contrib/database/postgres | v1.0.0 |
| servex-contrib/vector/pgvector | v1.0.0 |
| servex-contrib/search/postgres | v1.0.0 |

For subsequent snapshots, publish and verify upstream core and contrib versions
first. Select each version explicitly, tidy go.mod/go.sum and inspect the diff.
Go does not inherit dependencies’ replace directives. Never use sibling
checkouts or a workspace file as release acceptance.

## Validate a snapshot

1. Start from a clean checkout on `main`. Review both READMEs, the selected
   versions and the dated entry in [CHANGELOG.md](CHANGELOG.md).
2. Run `task check`, `task lint`, `task vulncheck` with Go 1.27.3.
   Confirm every hosted CI job passes at the exact commit to be tagged.
3. Copy or clone the project into a temporary directory with no sibling
   repositories. With `GOWORK=off`, download dependencies, run `go mod verify`
   and build all commands using `go build -mod=readonly ./...`.
4. Complete the normal and integration tests. Record the source revision,
   exact dependency versions, toolchain, check results and known limitations
   in the release notes. Do not describe unrun live integrations as validated.

Run automated checks without model credentials; they use local fixtures and
do not call paid APIs. Live provider tests require an explicit provider, model
and credentials. Build the server and run its binary from a temporary working
directory with a loopback address and temporary token, so `.agentx/` state
stays isolated. Do not retain tokens or sessions as release evidence.

## Publish

After the checks pass, create a GitHub source release for the reviewed commit,
using the matching changelog section as its notes. The initial snapshot is
`v1.0.0`; later snapshots use new versions. Verify the remote tag resolves to
the checked commit and the source archive can be downloaded. Never move an
already published tag. Dependency updates require a new snapshot.

## PostgreSQL acceptance

The v1.6.0 snapshot requires the PostgreSQL integration job at the exact release
commit, in addition to every existing gate. See [POSTGRESQL.md](POSTGRESQL.md).
Record actual tool/transaction behavior separately from standalone builds.
