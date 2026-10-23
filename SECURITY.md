# Security

Security fixes target the latest v1 source snapshot and current development
code. Update to the latest snapshot before reporting a vulnerability.
The HTTP server uses a shared local token and has no tenant isolation. Local
skills may execute hooks. Model calls can send prompts and tool results to the
selected provider. See the README before running or adapting an example.

For a vulnerability in example wiring, use this repository's GitHub
**Security → Report a vulnerability** channel if available. If unavailable,
open an issue asking for a private contact without exploit details. Do not
include credentials, private sessions or access-bearing URLs. No response-time
guarantee is stated. Include the source revision and a synthetic reproducer.

For a defect in an upstream library, follow that library's SECURITY.md instead:
[agentx](https://github.com/nostosx/agentx/blob/main/SECURITY.md) or
[agentx-contrib](https://github.com/nostosx/agentx-contrib/blob/main/SECURITY.md).
Ordinary bugs can use public issues.
