# 2026-11-28 无回复 issue 修复与验证

这一批从 `0d310af00f2934b4331609251340e7336a1f9d0c` 开始，每项保留独立分支，已有完整作者提交的保留原 Author，自己的补测和补强另行提交

账号为 [qq01-hub](https://github.com/qq01-hub)，整合分支为 [fix/unanswered-issues-20261128](https://github.com/qq01-hub/magpie/tree/fix/unanswered-issues-20261128)

以下是已经完成并审核的第一组 8 项，本轮其余问题仍在独立工作树处理

| Issue | 修复与边界 | 独立分支 | 审核 HEAD |
| --- | --- | --- | --- |
| [#7949](https://github.com/yetone/magpie/issues/7949) | Gemini 非流式回复遵守 includeThoughts，保留 usage 和纯思考回复的候选结果，流式行为保持原有 loop guard 约束 | [fix/7949-gemini-include-thoughts](https://github.com/qq01-hub/magpie/tree/fix/7949-gemini-include-thoughts) | `60c760256ffef6df50b88b54fac0f48a5a922c66` |
| [#7956](https://github.com/yetone/magpie/issues/7956) | Gemini inline audio 与 Chat input_audio 发给明确支持音频的插件 Messages 模型，普通供应商、未知能力、音频 URL 和 assistant 音频保留文本降级 | [fix/7956-audio-author](https://github.com/qq01-hub/magpie/tree/fix/7956-audio-author) | `f84126c75375e005cc9357855771e3ce28641198` |
| [#7777](https://github.com/yetone/magpie/issues/7777) | DeepSeek Hooks 经 YAML 重排仍可识别，用户修改路径、额外设置、同级插件或包名后不会被覆盖或移除 | [fix/7777-dsh-hook-ownership](https://github.com/qq01-hub/magpie/tree/fix/7777-dsh-hook-ownership) | `08c50d19e45ca9746f4211bb7f017c3d77b51462` |
| [#7965](https://github.com/yetone/magpie/issues/7965) | OpenCode Go 及其路由组恢复 Codex Ultra，preset-free 官方地址同样处理，混有通用 ChatGPT relay 的组保留密封任务限制 | [fix/7965-opencode-seals-ultra](https://github.com/qq01-hub/magpie/tree/fix/7965-opencode-seals-ultra) | `148a4735bb760c1dd44bc4b501f23a54147dfea4` |
| [#7633](https://github.com/yetone/magpie/issues/7633) | Claude 钥匙串不可读或返回损坏内容时保持未知状态，不从旁边的空凭据文件误判为需要重新登录，缺失项和明确清空仍按各自语义处理 | [fix/7633-claude-unread-keychain](https://github.com/qq01-hub/magpie/tree/fix/7633-claude-unread-keychain) | `6017b1eda78fa6b9e0bb4633541e0f5e6b2eed7d` |
| [#7584](https://github.com/yetone/magpie/issues/7584) | 显示托管 Claude Code 和 agy CLI 的版本，普通 Claude 安装优先，托管副本与 agy 不增加不支持的更新或安装操作 | [fix/7584-managed-agent-versions](https://github.com/qq01-hub/magpie/tree/fix/7584-managed-agent-versions) | `d3a1f3efbb6f2563ead6bd26e2d3897d64d5be94` |
| [#7960](https://github.com/yetone/magpie/issues/7960) | 同账号 Gemini 4 argon/carbon/hydrogen/neon 共享 opaque，旧缓存按共享键、当前 tier 的两种拼写、其他 tier 的顺序读取，其他账号及模型保持隔离 | [fix/7960-copilot-gemini-opaque-family](https://github.com/qq01-hub/magpie/tree/fix/7960-copilot-gemini-opaque-family) | `8684801341a4aba8e0f0c3962db3ac715e19f347` |
| [#7939](https://github.com/yetone/magpie/issues/7939) | WebSocket 转内部 POST 时复用大正文，保留小正文的独立缓冲和实际字节数限制，伪造较小 Content-Length 仍返回 413 | [fix/7939-responses-websocket-body](https://github.com/qq01-hub/magpie/tree/fix/7939-responses-websocket-body) | `8e158cc05fcb56f4be73c0b401bb69e7c6888392` |

## 实际失败复现与验证

以下命令均在隔离的测试 HOME 执行，Go SDK 为 1.26.8，Bun 为 1.4.5，Go 缓存、模块缓存、GOPATH、GOTMPDIR、TEMP 和 TMP 全在 D 盘

| Issue | 修复前的实际失败 | 修复后的验证 |
| --- | --- | --- |
| #7949 | 未设置 includeThoughts 时仍出现 thought parts，纯思考回复未过滤 | 三个新回归以及 loop guard、tool search、监听地址和 self-proof 邻接测试通过 |
| #7956 | Gemini 和 Chat 音频均变成 `[attachment audio/wav]` | 两种输入入口、能力开关、URL 和 role 负例及视频邻接测试通过 |
| #7777 | 原主干不能识别重排后的 Hooks 行，作者补丁仍会改写五类用户编辑边界 | 7 个顶层测试和 8 个边界子测试通过 |
| #7965 | 实际生成的 Codex catalog 中 Go 单模型、组及无 preset 的官方地址都缺 Ultra | provider、codexcat、gateway 共 51 个顶层测试通过 |
| #7633 | 36/51/128 错误误标重新登录，作者补丁对成功返回非法 JSON 的情况仍误标 | 真实生产 reader 经 Windows fake security 命令覆盖 unreadable、missing、malformed、known-empty 和成功优先级，定向通过 |
| #7584 | 托管 Claude 与 agy 的版本为空 | 两个新增跨平台行为测试及版本解析通过，三个既有 Unix shell fixture 测试在 Windows 跳过 |
| #7960 | 切换 tier 丢失 opaque，旧缓存冲突时取到 sibling 而非当前 tier，前缀别名也有同类问题 | 10 个顶层相关测试和 5 个子测试通过，包含真实本地 HTTP 请求和协议翻译路径 |
| #7939 | 33554444 字节的 socket 请求仍整份复制 | 三个 body ownership / limit 测试通过 race，另 42 个相关顶层 socket/request/unrouted 测试通过 |

说明：#7777 提交正文中边界红测的 “on the base” 指作者补丁上的边界版本，原始主干复现与作者补丁补强是两个独立失败记录

第一组整合代码 HEAD 为 `de23e87fc866a840bcca3fec9cdc9771530e1efa`，最终统一回归执行：

```text
go test -tags nogui -run '^Test(GeminiThoughtsOnlyWhenIncluded|GeminiThoughtsOnlyReplyKeepsItsCandidate|GeminiStreamStillCarriesThoughts|AudioReachesAPluginModelThatTakesIt|ChatAudioReachesAPluginMessagesModelThatTakesIt|HooksDsh.*|CodexSeals|OpenCodeGoCodexCatalogOffersUltra|OpenCodeGoIsNotASealedTaskReader|ClaudeUnreadKeychainIsNotARefusedSignIn|ClaudeMalformedKeychainIsNotARefusedSignIn|ClaudeClearedKeychainIsKnownEmptyBeforeTheFile|ClaudeMissingKeychainItem.*|ClaudeReadableKeychainTakesPrecedenceOverTheFile|ManagedClaudeVersionFallback|AgyVersionHasNoInstaller|CopilotOpaque.*|SocketPost.*)$' -count=1 -json -timeout 5m ./internal/agent ./internal/library ./internal/provider ./internal/gateway
```

结果为 33 个顶层测试、49 个测试及子测试全部通过，0 失败、0 跳过

以下检查在上述整合代码上实际完成并通过：

- Windows `go vet ./...`
- Linux 目标 `go vet -tags nogui ./...`
- Windows `go build`，包含桌面构建
- Linux 目标 `go build -tags nogui`
- 改动 Go 文件的 gofmt 检查与 `git diff --check`
- #7949 与 #7956 合并后的五项回归通过 race

## 验证范围与限制

尚未完成本轮完整 `go test -tags nogui ./...`，以上是定向回归、静态检查和构建结果，不能解释为整仓测试全部通过

额外完整 provider 包测试遇到 `TestRestartSwitchedCodexAppReal` 失败：helper 无输出后测试访问空切片，该失败在本轮未改动的固定生产基线上也已实际复现，相关 Claude 定向测试通过

本机未提供可运行的 Linux/macOS 环境，跨平台 vet/build 不代表对应系统的运行测试，未调用真实付费模型、真实 macOS 钥匙串或修改现有 agent 配置，供应商相关行为使用隔离的本地 HTTP 服务和 fixture 验证

原作者提供的提交消息中可能包含其本人执行的 Linux、macOS 或 live 验证，本报告只把本次会话实际执行的检查计入结果
