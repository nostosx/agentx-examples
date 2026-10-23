# agentx-examples

[English](README.md) | 简体中文

组合 agentx 与 agentx-contrib 公开 API 的可运行示例，独立模块为
`github.com/nostosx/agentx-examples`，要求 Go 1.27.3。
库旁的 godoc `Example` 测试继续保留。

首个源码快照计划为 v1.0.0，尚未发布。快照版本标识示例源码及依赖，
不代表稳定的库 API。

## 本地运行

agentx 及其 contrib 模块仍处于开发阶段，请将以下目录放在同一级。
kitx 已使用发布的 v1.0.0，由 Go 模块代理下载，无需本地检出：

```text
nostosx/
├── agentx/
├── agentx-contrib/
└── agentx-examples/
```

`go.mod` 通过相对 `replace` 解析尚未发布的 core 和 contrib 模块。从本目录运行：

```sh
export AGENTX_PROVIDER=anthropic
# 在环境中设置 ANTHROPIC_API_KEY。
GOWORK=off go run ./cmd/basic "什么是 agent loop？"
```

`AGENTX_PROVIDER` 支持 `anthropic`、`openai`、`gemini`；
未指定时，按此顺序选择已配置的 API key。
分别使用 `ANTHROPIC_API_KEY`、`OPENAI_API_KEY`、`GEMINI_API_KEY`；
Gemini 也支持 `GOOGLE_API_KEY`。`AGENTX_MODEL` 可覆盖默认模型。

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

服务默认监听 `127.0.0.1:8080`，`ADDR` 仅允许选择其他回环地址或端口。
启动前将 `AGENTX_SERVER_TOKEN` 设置为不含空白的随机密钥，所有 API 请求均须携带
`Authorization: Bearer <token>`。会话、审批、AI SDK、AG-UI 和 A2A 路由共享此密钥；
该示例面向单个本地信任域，没有用户或租户隔离。
`http://localhost:3000` 的浏览器客户端可发送预检请求，实际请求仍须携带密钥；
不要将密钥写入前端构建产物或 URL。
收到 SIGINT/SIGTERM 后，服务取消活动模型调用，为 HTTP 关闭保留十秒，随后关闭剩余连接。
请求示例见 [cmd/server/main.go](cmd/server/main.go)。

## 检查与依赖版本

```sh
make check
make lint
make vulncheck
```

CI 将 core、contrib 放在同级目录，编译全部示例并运行回归测试，不调用真实模型 API。
core 和 contrib 使用占位版本及本地替换，kitx 使用发布的 v1.0.0；库不依赖示例项目。

实际发布后，分别为 core、provider、filestore 选择版本并移除对应替换，
再执行 `GOWORK=off go mod tidy` 和检查；三个模块路径详见
[英文说明](README.md#checks-and-dependency-versions)。
本项目没有发布工作流。工具安装与检查说明见 [CONTRIBUTING.md](CONTRIBUTING.md)，
可重现快照及依赖更新见 [RELEASING.md](RELEASING.md)，私密漏洞报告见
[SECURITY.md](SECURITY.md)，变更记录见 [CHANGELOG.md](CHANGELOG.md)。
采用 [MIT 许可证](LICENSE)。
