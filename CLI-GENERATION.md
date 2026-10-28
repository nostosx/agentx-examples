# CLI generation and application extensions

This repository is one Go module created from the nostosx-cli v1.2.0 with the
monolith layout, agentx support and only HTTP. Normal builds need no sibling source.
Install CLI v1.2.0 and set NOSTOSX to the absolute nostosx executable path.

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
- HTTP adapters implement authentication, CORS and owner cancellation. Root and
  standalone entries reuse the same app.go and generated HTTP configuration and server constructors.
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
