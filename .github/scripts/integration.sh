#!/usr/bin/env bash
set -euo pipefail
: "${AGENTX_EXAMPLES_IT_POSTGRES_DSN:?set an owned vector-enabled PostgreSQL DSN}"
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$root"
result="$(mktemp)"
trap 'rm -f "$result"' EXIT
status=0
GOWORK=off "${GO:-go}" test -race -shuffle=on -count=1 -timeout=5m -tags integration -run '^TestIntegration' -json ./... | tee "$result" || status=$?
GOWORK=off "${GO:-go}" run "$root/.github/scripts/integrationcheck/main.go" "$result"
exit "$status"
