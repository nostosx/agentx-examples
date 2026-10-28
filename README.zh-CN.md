# agentx-examples

[English](README.md) | 简体中文

组合 agentx 与 agentx-contrib 公开 API 的可运行示例，独立模块为
`github.com/nostosx/agentx-examples`，要求 Go 1.27.3。
库旁的 godoc `Example` 测试继续保留。

v1.2.0 快照使用 agentx v1.2.1、provider v1.2.0、filestore v1.0.0、
servex v1.3.0。v1.1.0 标签保留历史依赖；快照版本标识示例源码及依赖，
不代表稳定的库 API。

## CLI 生成的工程

每个仓库是一个 CLI 单体布局工程，启用 agentx 能力。根 `cmd/` 仅保留命令入口，
六个命令行演示的实现位于 `application/`；模型配置、技能加载和进程取消位于
`infrastructure/`。两个服务入口解析配置后调用 `assistant.Run(ctx, cfg)`，
服务的 `app.go` 集中管理依赖、协议清理与路由注册；HTTP 适配器位于自己的 `internal/port/`。

session、AI SDK、AG-UI、A2A 使用 agentx 原生 HTTP 协议，
不包含 Proto、gRPC 监听或契约工具。独立入口
`go run ./services/assistant/cmd/assistant` 与 `cmd/server` 复用相同装配。
实际生成步骤和应用扩展见 [CLI-GENERATION.md](CLI-GENERATION.md)。

## 本地运行

使用 Go 1.27.3，独立克隆本仓库即可。依赖从 Go 模块代理下载，
无需同级仓库或本地替换：

```sh
git clone https://github.com/nostosx/agentx-examples.git
cd agentx-examples
GOWORK=off go mod download
```

从本目录运行：

```sh
export AGENTX_PROVIDER=anthropic
# 在环境中设置 ANTHROPIC_API_KEY。
GOWORK=off go run ./cmd/basic "什么是 agent loop？"
```

`AGENTX_PROVIDER` 支持 `anthropic`、`openai`、`gemini`；
未指定时，按此顺序选择已配置的 API key。
分别使用 `ANTHROPIC_API_KEY`、`OPENAI_API_KEY`、`GEMINI_API_KEY`；
Gemini 也支持 `GOOGLE_API_KEY`。`AGENTX_MODEL` 可覆盖默认模型。
对应的 `ANTHROPIC_BASE_URL`、`OPENAI_BASE_URL` 或 `GEMINI_BASE_URL`
可选择可信的兼容端点，凭据会发送到该地址；不设置时使用供应商的公开 API。
OpenAI 默认使用 Responses API，因此这些命令不能直接使用只支持 Chat Completions 的端点。

## 示例

| 命令 | 内容 |
|---|---|
| `go run ./cmd/basic` | 单轮问答 |
| `go run ./cmd/tools` | 类型化工具、hook 与演示审批策略 |
| `go run ./cmd/streaming` | 流式输出和多轮会话，输入 `/exit` 结束 |
| `go run ./cmd/skills` | 内嵌 skill；通过 `-skills-dir .skills` 显式加载可信的本地 skill |
| `go run ./cmd/multi-agent` | 协调者向多个专门 agent 委派任务 |
| `go run ./cmd/workflow` | 并行起草、审阅与条件编辑 |
| `go run ./cmd/server` | HTTP 协议、审批及 `.agentx/` 文件持久化 |

这些命令会调用真实模型 API。工具和部署操作属于演示，改造时应配置应用自己的授权策略。
本地 skill 可注册可执行 hook，因此只应通过 `-skills-dir` 加载可信目录。
六个命令行示例响应 SIGINT/SIGTERM，并设置五分钟期限（流式聊天按每轮计算）；
聊天的输入、输出失败会结束交互。

服务仅监听 `127.0.0.1:8080`，
使用 `-addr` 选择其他字面量 loopback IP 和端口。
启动前将 `API_TOKEN` 设置为至少 32 个不含空白的可打印 ASCII 字符，所有 API 请求均须携带
`Authorization: Bearer <token>`。会话、审批、AI SDK、AG-UI 和 A2A 路由共享此密钥；
该示例面向单个本地信任域，没有用户或租户隔离。
`http://localhost:3000` 的浏览器客户端可发送预检请求，实际请求仍须携带密钥；
不要将密钥写入前端构建产物或 URL。
收到 SIGINT/SIGTERM 后，服务取消活动模型调用，为 HTTP 关闭保留十秒，随后关闭剩余连接。
HTTP 健康检查公开，Agent 协议使用该 token。HTTP 写超时为零以支持流式响应，
应用退出时主动取消未完成的模型请求。

首次运行服务时，保留前面的模型配置，在一个终端启动：

```sh
export API_TOKEN="$(openssl rand -hex 32)"
GOWORK=off go run ./cmd/server
```

在另一个终端将 `API_TOKEN` 设置为相同的值，再执行：

```sh
curl -N http://127.0.0.1:8080/sessions/demo/runs \
  -H "Authorization: Bearer $API_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"input":"Deploy the latest build to staging"}'
curl http://127.0.0.1:8080/sessions/demo \
  -H "Authorization: Bearer $API_TOKEN"
```

运行请求会自动创建 `demo` 会话。模型调用 `deploy` 时，事件流暂停并等待审批。
查看会话的 `pending` 字段，将返回的调用 ID 放入
`{"approvals":{"<call_id>":{"approved":true}}}`，携带相同认证头 POST 到
`/sessions/demo/resume` 恢复执行。
部署工具只返回演示消息，不会实际部署构建产物。
状态保存在仓库根目录下的 `.agentx/` 中，使用 Ctrl+C 停止服务。

## 检查与依赖版本

```sh
task check
task lint
task vulncheck
```

统一公开八个入口：`check`、`test`、`build`、`lint`、`vulncheck`、`fmt`、`fix`、`tidy`；
直接运行 `task` 等同于 `task check`。`task --list` 查看入口，`task test RACE=true`
启用竞态检测，`task test RACE=true COVER=true` 同时生成 `coverage.out` 并打印覆盖率。
格式、现代化、依赖和 vet 检查作为内部步骤保留。

CI 编译全部示例并运行回归测试，不调用真实模型 API。
agentx v1.2.1、provider v1.2.0、filestore v1.0.0、servex v1.3.0
和 kitx v1.0.0 均为已发布版本，
无需同级源码检出或本地替换。库不反向依赖示例项目。
依赖更新需要通过独立构建，随后发布新的源码快照；本项目没有自动发布工作流。
工具安装与检查说明见 [CONTRIBUTING.md](CONTRIBUTING.md)，
可重现快照及依赖更新见 [RELEASING.md](RELEASING.md)，私密漏洞报告见
[SECURITY.md](SECURITY.md)，变更记录见 [CHANGELOG.md](CHANGELOG.md)。
采用 [MIT 许可证](LICENSE)。

本工程仅使用 HTTP，不需要 Buf、Proto 生成或契约 CI job。
普通构建使用随源码保存的生成代码；没有 smoke 任务和 Python 工具依赖。
