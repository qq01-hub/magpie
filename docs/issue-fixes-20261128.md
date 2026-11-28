# 2026-11-28 无回复 issue 修复与验证

这一批从 `0d310af00f2934b4331609251340e7336a1f9d0c` 开始，每项保留独立分支，已有完整作者提交的保留原 Author，自己的补测和补强另行提交

账号为 [qq01-hub](https://github.com/qq01-hub)，整合分支为 [fix/unanswered-issues-20261128](https://github.com/qq01-hub/magpie/tree/fix/unanswered-issues-20261128)

以下是已经完成并审核的 17 项，本轮其余问题仍在独立工作树处理

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
| [#8019](https://github.com/yetone/magpie/issues/8019) | 站点脚本通过规范化真实路径判断命令入口，路径含空格、`#` 或经符号链接调用时仍执行，作为模块导入时无命令副作用 | [fix/8019-site-entry-paths](https://github.com/qq01-hub/magpie/tree/fix/8019-site-entry-paths) | `2120148bedb0d7cfdffaee9e424465a260ea094b` |
| [#7992](https://github.com/yetone/magpie/issues/7992) | 技能路由描述在原有 1024 UTF-16 单位预算内选取能容纳的最长共同前缀，保留短描述和完整技能名 | [fix/7992-skill-description-budget](https://github.com/qq01-hub/magpie/tree/fix/7992-skill-description-budget) | `feed84b98918cea133aa03684c63cd6e29a3bf51` |
| [#7870](https://github.com/yetone/magpie/issues/7870) | 上游明确的图片大小超限 400/413 可转到组内下一成员，原成员不冷却，普通请求错误、图片格式、数量及速率限制不误归类 | [fix/7870-image-size-group-failover](https://github.com/qq01-hub/magpie/tree/fix/7870-image-size-group-failover) | `99930ace8900c72e56a8c699296b28e036692d97` |
| [#7969](https://github.com/yetone/magpie/issues/7969) | 从插件自身聚合列表发现朗读及转写模型，支持第二账号与本地模型别名，遵守禁用和供应商顺序，不把普通音频聊天模型当成语音接口 | [fix/7969-plugin-speech-discovery](https://github.com/qq01-hub/magpie/tree/fix/7969-plugin-speech-discovery) | `cc7d131bd5c0b547f1987e0ce1cb0dcd5a138717` |
| [#7816](https://github.com/yetone/magpie/issues/7816) | 移除插件账号后将其专有模型移出已选列表，保留手动模型，旧编辑窗口保存不会加回，账号恢复后重新列出 | [fix/7816-plugin-delisted-picks](https://github.com/qq01-hub/magpie/tree/fix/7816-plugin-delisted-picks) | `07c609ae5da16b0892e48aa5b3485bad9e78a4b8` |
| [#7938](https://github.com/yetone/magpie/issues/7938) | AGY 本地摘要请求超限时精简历史图片和大工具输出，UTF-8 字符保持完整，普通请求及原生压缩保持原样，仍超限返回 413 | [fix/7938-oversized-agy-compaction](https://github.com/qq01-hub/magpie/tree/fix/7938-oversized-agy-compaction) | `59c12f9cca2ea0ee2432587e6dca1005079a6093` |
| [#7988](https://github.com/yetone/magpie/issues/7988) | 技能 frontmatter 解析失败时丢弃不完整字段，显示文件夹名及本地化警告，区分来源显示名和安全安装名 | [fix/7988-skill-frontmatter-parse](https://github.com/qq01-hub/magpie/tree/fix/7988-skill-frontmatter-parse) | `5af7e9a49d11ed3bb6511a5aa07f77f4fb42a259` |
| [#7799](https://github.com/yetone/magpie/issues/7799) | 交互模式粘贴前清理 bidi 控制标记，保留原文、ZWJ emoji 和零宽空格，命令前缀保护在清理后仍生效 | [fix/7799-interactive-bidi-paste](https://github.com/qq01-hub/magpie/tree/fix/7799-interactive-bidi-paste) | `38728eb8e52c9a9b6c01054560449f7ba6aaf756` |
| [#7888](https://github.com/yetone/magpie/issues/7888) | 同源余额刷新错开请求，429 后按 Retry-After 或指数退避等待并保留上次余额，探测限流不记录为无余额接口，等待可取消 | [fix/7888-balance-refresh-429](https://github.com/qq01-hub/magpie/tree/fix/7888-balance-refresh-429) | `eff2ed3148b16970a28ffd8de6f664c4cd68df1e` |

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
| #8019 | 三个索引命令在特殊路径和链接入口下缺少生成文件，链接 credits 命令意外返回成功 | Node 24.22.3 定向 27 项全部通过，完整 site 套件为 176 通过、33 失败，33 项均属于 Windows 上的 Unix installer fixture，原基线同名 33 项也失败 |
| #7992 | 14 个技能仍被截到 40 字符档位且预算闲置，35 短描述加 8 长描述退化成只列名称 | 七项 RouterDesc 测试在独立分支及整合代码 `8a294c0cf` 上全部通过，覆盖 Unicode、80 rune 上限和名称溢出；较广的 Library 回归曾达到四分钟套件超时，其中 `TestSkillRouterUnnamedInView` 随后在整合代码上单独运行通过，耗时 214.05 秒，该单项结果不代表整个 Library 套件通过 |
| #7870 | 本地真实 HTTP 请求中首成员的图片大小限制 400 被直接返回，下一成员请求次数为零 | 六项顶层和 26 个子测试全部通过，覆盖小图前后仍用首成员、大小图回退、不冷却、分类负例及原有通用 413 回退 |
| #7969 | 最终 WorkBuddy 模型 fixture 在原生产代码上四项失败：audio-only TTS 被聊天列表中的候补取代，供应商次序后的 TTS 选择错误，第二账号专有模型和本地别名缺失 | 七项测试及相同七项 race 全部通过，0 跳过，覆盖禁用、Select none、用户选择、普通音频聊天负例，真实 Bun 插件宿主；转写在此基线 fixture 已通过，不虚报为另一条红测 |
| #7816 | 撤回后台 relisted 回调后移除账号仍留下两项专有模型，撤回 Fetch models 的插件条件后 dropped 为空 | 两项实际 Bun 插件回归及 race 通过，另两项后台重列与恢复测试通过 race，整合代码同四项通过，未复测原作者的真实 CodeZ 账号 |
| #7938 | 真实超过 512 MiB 的本地 HTTP 请求在旧 handler 返回 413、上游零字节；作者补丁的工具输出截断会切碎 UTF-8 字符 | 大请求压缩及继续对话、UTF-8 边界、普通或仍超限输入、原生压缩四项均通过，整合代码同四项通过，运行未传 `-short`，未使用真实 AGY 账号 |
| #7988 | 旧生产代码的 Library 和 Agent 自有技能视图丢失 YAML 解析错误 | 两项 Go 回归通过，Chromium/WebKit 五语言共 12 项通过，覆盖来源安装、合法邻居、窄屏和 HTML 样式错误值，整合分支同组通过 |
| #7799 | 旧代码的交互 shim 仍将全部 bidi 标记发给模拟 Claude，四种前缀保护失败 | 新 shim 测试、Pasteable 和原有交互发送三项通过，整合分支同组通过；额外完整 claudetui 包运行中止，未作为通过，未运行真实 Claude Code 2.1.332 |
| #7888 | 旧余额路径同源请求间隔为零，429 后第二次读取没有等待；作者补丁对已取消 context 仍返回成功 | 六项余额与两项取消测试及两个子案例通过 `-race -count=20`，整合代码同八项通过，覆盖秒数和 HTTP-date Retry-After、余额不足负例、迟到成功、探测和 last-good；额外 provider 整包运行中止，未作为通过 |

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

#7988 额外界面邻接检查没有全部通过：原样 `skill-manual.test.cjs` 的空 POST 断言被 `/api/update/using` 心跳触发，固定基线同样失败；临时仅过滤测试心跳后 22 项通过，随后恢复原测试文件，没有把过滤算入提交结果

`found-skill-pick-remove.test.cjs` 在修复树为 13 通过、9 失败（含父测试），失败涉及 reach/press 辅助流程；固定基线复现了两个引擎下日语、德语 420px 的四个子测试失败，另外三个 WebKit 子测试失败在该次基线运行没有复现，不能把它们一概认定为基线问题；新增 frontmatter 测试与实际改动路径仍为两引擎五语言全通过

本机未提供可运行的 Linux/macOS 环境，跨平台 vet/build 不代表对应系统的运行测试，未调用真实付费模型、真实 macOS 钥匙串或修改现有 agent 配置，供应商相关行为使用隔离的本地 HTTP 服务和 fixture 验证

原作者提供的提交消息中可能包含其本人执行的 Linux、macOS 或 live 验证，本报告只把本次会话实际执行的检查计入结果
