# Changelog

Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

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
