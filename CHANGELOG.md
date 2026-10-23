# Changelog

No snapshot has been released yet. Snapshot versions identify example source
and dependencies; they do not establish a stable library API.

## Unreleased

## v1.0.0

Prepared first source snapshot; not yet published.

### Added

- Seven runnable applications for basic calls, tools, streaming, skills,
  multi-agent coordination, workflows and an authenticated local HTTP server.
- Local regression tests and Makefile checks for Go 1.27.3, with independent
  cross-platform CI, lint and vulnerability jobs.

### Dependencies

- Use the published kitx v1.0.0 through agentx and filestore.
- Keep local replacements for the unreleased agentx, provider and filestore
  modules. Independently buildable snapshots require published versions; see
  [RELEASING.md](RELEASING.md).
