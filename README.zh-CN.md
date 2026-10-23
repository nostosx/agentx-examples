# agentx-examples

[English](README.md) | 简体中文

组合 agentx 与 agentx-contrib 公开 API 的可运行示例，独立模块为
`github.com/nostosx/agentx-examples`，要求 Go 1.27.3。
库旁的 godoc `Example` 测试继续保留。

v1.0.0 源码快照使用已发布的 v1.0.0 库依赖。快照版本标识示例源码及依赖，
不代表稳定的库 API。

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

CI 编译全部示例并运行回归测试，不调用真实模型 API。
agentx、provider、filestore 和 kitx 均使用已发布的 v1.0.0，
无需同级源码检出或本地替换。库不反向依赖示例项目。
依赖更新需要通过独立构建，随后发布新的源码快照；本项目没有自动发布工作流。
工具安装与检查说明见 [CONTRIBUTING.md](CONTRIBUTING.md)，
可重现快照及依赖更新见 [RELEASING.md](RELEASING.md)，私密漏洞报告见
[SECURITY.md](SECURITY.md)，变更记录见 [CHANGELOG.md](CHANGELOG.md)。
采用 [MIT 许可证](LICENSE)。
