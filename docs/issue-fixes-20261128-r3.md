# Issue fixes, 2026-11-28, round 3

Each independent branch starts at `e87965dcd302f62f9c11ba1c622d1cdab0bdc4d8`
The integration branch is [qq01-hub/magpie:fix/unanswered-issues-20261128-r3](https://github.com/qq01-hub/magpie/tree/fix/unanswered-issues-20261128-r3)
Issue state is checked again immediately before a comment is posted

## Reviewed changes

| Issue | Behavior | Independent branch | Reviewed head |
| --- | --- | --- | --- |
| [#8068](https://github.com/yetone/magpie/issues/8068) | OpenCode's browser integration directory alone no longer counts as an installed agent; a real config file still does | [fix/8068-opencode-browser-detection](https://github.com/qq01-hub/magpie/tree/fix/8068-opencode-browser-detection) | `00e4d12856ece501f9a01787e0ae2e8214b2d099` |
| [#7165](https://github.com/yetone/magpie/issues/7165) | Azure's foreign-resource history rejection uses the existing bounded retry and refusal memory, including masked stream errors | [fix/7165-azure-foreign-seal](https://github.com/qq01-hub/magpie/tree/fix/7165-azure-foreign-seal) | `d58a42524673df8be0767ce68972d9f6c4018621` |
| [#7432](https://github.com/yetone/magpie/issues/7432) | Grok Build sessions stay distinct across affinity, plugin calls and remote gateway hops | [fix/7432-grok-session-headers](https://github.com/qq01-hub/magpie/tree/fix/7432-grok-session-headers) | `d211934fd6a177b906e8fad99e90b1a68adb2b39` |
| [#6456](https://github.com/yetone/magpie/issues/6456) | GLM Coding Plan searches call the renamed `web_search_prime` MCP tool | [fix/6456-glm-search-tool-name](https://github.com/qq01-hub/magpie/tree/fix/6456-glm-search-tool-name) | `2ad2e263141c4af33c152e794c1f9092d3103ae3` |
| [#7385](https://github.com/yetone/magpie/issues/7385) | OAuth callback ports inside a SPA fragment query are recognized, enabling the paste callback entry | [fix/7385-oauth-fragment-callback](https://github.com/qq01-hub/magpie/tree/fix/7385-oauth-fragment-callback) | `47caa25b0d18c2b9844aff99e94ecbf9a10f1957` |
| [#6614](https://github.com/yetone/magpie/issues/6614) | Hermes custom-prefixed Magpie providers match their catalog options, profile lifecycle and drift checks | [fix/6614-hermes-custom-magpie](https://github.com/qq01-hub/magpie/tree/fix/6614-hermes-custom-magpie) | `23306813e28de0c4f386c025697607f78b0a637d` |
| [#7434](https://github.com/yetone/magpie/issues/7434) | AiHubMix's exact unlimited-key sentinel is no longer displayed as an account balance | [fix/7434-aihubmix-unlimited-quota](https://github.com/qq01-hub/magpie/tree/fix/7434-aihubmix-unlimited-quota) | `578672a3d582f133f419539d0864a3bb513c42d9` |
| [#5330](https://github.com/yetone/magpie/issues/5330) | Plugin accounts are recorded under their provider identity so quota cards follow the account in use | [fix/5330-plugin-served-id](https://github.com/qq01-hub/magpie/tree/fix/5330-plugin-served-id) | `1537f03e5195906d6dbfcadfc5fe48f4e8f8a481` |
| [#7072](https://github.com/yetone/magpie/issues/7072) | Structured plugin empty replies use the bounded empty-reply retry budget in HTTP and SSE | [fix/7072-plugin-empty-reply-kind](https://github.com/qq01-hub/magpie/tree/fix/7072-plugin-empty-reply-kind) | `eb3a87c20c8a131aea6c88d7d2ce2d778b9fae9c` |
| [#7091](https://github.com/yetone/magpie/issues/7091) | MCP OAuth discovery tries path-appended AS metadata after RFC/OIDC candidates, parsing each response independently | [fix/7091-mcp-oauth-metadata-path](https://github.com/qq01-hub/magpie/tree/fix/7091-mcp-oauth-metadata-path) | `90a1fc583ec34b5762068e9480015a1a8a24da2d` |
| [#3057](https://github.com/yetone/magpie/issues/3057) | Drawer discovery advertises image output in the standard output_modalities field | [fix/3057-drawer-output-modalities](https://github.com/qq01-hub/magpie/tree/fix/3057-drawer-output-modalities) | `79ec3f010ae415443a7f79fb0d1e99acd564296a` |
| [#5229](https://github.com/yetone/magpie/issues/5229) | dsh reads and writes the last effective default-model entry, leaving metadata and other plugins alone | [fix/5229-dsh-last-effective](https://github.com/qq01-hub/magpie/tree/fix/5229-dsh-last-effective) | `be82702568901091e39013dcd2a7c02deb0d162a` |
| [#4028](https://github.com/yetone/magpie/issues/4028) | Provider imports accept single-label LAN names with ports and IPv6 literals, preserving URL validation and gateway loop prevention | [fix/4028-lan-host-url](https://github.com/qq01-hub/magpie/tree/fix/4028-lan-host-url) | `04d9c78e649e849afa926c6edd6558ce63f512b3` |
| [#2690](https://github.com/yetone/magpie/issues/2690) | Codex titles drop localized labels and read localized JSON fields with deterministic precedence | [fix/2690-codex-title-labels](https://github.com/qq01-hub/magpie/tree/fix/2690-codex-title-labels) | `9baab8957866d611b6c490a7d4e8b213168d468a` |
| [#6148](https://github.com/yetone/magpie/issues/6148) | vLLM grammar-format custom-tool refusals use the existing same-Responses-API translation | [fix/6148-custom-tool-grammar](https://github.com/qq01-hub/magpie/tree/fix/6148-custom-tool-grammar) | `2b8a1ea8e295d57fd7cfe62fae7c57c491d9b1ea` |
| [#4312](https://github.com/yetone/magpie/issues/4312) | An empty implicit DSH Web profile yields to a configured profile, while explicit leads and shared start settings keep precedence | [fix/4312-dsh-configured-profile](https://github.com/qq01-hub/magpie/tree/fix/4312-dsh-configured-profile) | `8df2ebf618aac72e5309a38397ce74dc88672f1f` |
| [#7564](https://github.com/yetone/magpie/issues/7564) | Named host and WSL OMP profiles receive Library resources in their own directories, with version-aware shared skills | [fix/7564-omp-profile-target](https://github.com/qq01-hub/magpie/tree/fix/7564-omp-profile-target) | `5546f31dbc34149c533c2ce0f3dba185849cf310` |
| [#7019](https://github.com/yetone/magpie/issues/7019) | Client-run tool search converts whole-number floats before handing the call to Codex | [fix/7019-tool-search-float-args](https://github.com/qq01-hub/magpie/tree/fix/7019-tool-search-float-args) | `4cf37fe9668a96b93715f1e8856043a16257e9e7` |
| [#2815](https://github.com/yetone/magpie/issues/2815) | Claude Code installed with Homebrew's latest cask gets its own channel's update | [fix/2815-claude-brew-latest](https://github.com/qq01-hub/magpie/tree/fix/2815-claude-brew-latest) | `b2b8ba80cf8a3406668f501c692b84a2f2b7a07e` |
| [#7431](https://github.com/yetone/magpie/issues/7431) | Qwen Code receives Library MCP servers and tool exclusions in its settings file | [fix/7431-qwen-library-target](https://github.com/qq01-hub/magpie/tree/fix/7431-qwen-library-target) | `75c2e408c2fc7d8d32ac3807e7d5f409cb13e19c` |
| [#3735](https://github.com/yetone/magpie/issues/3735) | Copied Claude fork heads no longer inflate session usage, including full parsing and cached resume | [fix/3735-claude-fork-count](https://github.com/qq01-hub/magpie/tree/fix/3735-claude-fork-count) | `85e996ea6bda4ec077942c14dd8764b472883f35` |
| [#6696](https://github.com/yetone/magpie/issues/6696) | Remote delta requires matching output content, preserving client edits and counting retained cache data | [fix/6696-remote-delta-content](https://github.com/qq01-hub/magpie/tree/fix/6696-remote-delta-content) | `2202e65de5d66070fd018a86f2602f48a30f7e97` |
| [#5221](https://github.com/yetone/magpie/issues/5221) | Explicit InternLM content-check codes are classified as refusals without broad message matching | [fix/5221-internlm-safety-refusal](https://github.com/qq01-hub/magpie/tree/fix/5221-internlm-safety-refusal) | `29a39f55ebe78daa3fbdbd3950c7526bc8ef94b8` |
| [#7563](https://github.com/yetone/magpie/issues/7563) | OMP project skills, MCP servers and plugin parts use their native project files | [fix/7563-omp-project-targets](https://github.com/qq01-hub/magpie/tree/fix/7563-omp-project-targets) | `b314126c7447fbdf7235832d2ce7fd77b4a32ac8` |
| [#2835](https://github.com/yetone/magpie/issues/2835) | Complete current-platform editor SDK caches keep Claude and Codex detected, including beside Cindy | [fix/2835-agent-host-sdk](https://github.com/qq01-hub/magpie/tree/fix/2835-agent-host-sdk) | `1418ec294bf055b40dbac5855f7ad8d571078696` |
| [#2609](https://github.com/yetone/magpie/issues/2609) | Windows update lookup follows a running executable moved away from its launch path, including long Unicode paths | [fix/2609-windows-moved-executable](https://github.com/qq01-hub/magpie/tree/fix/2609-windows-moved-executable) | `1b7ec71e3cdc059940c5b35415cf4cebc1821b39` |
| [#6064](https://github.com/yetone/magpie/issues/6064) | Plugin sign-ins asking only nonsecret prompts retain the separate API key field | [fix/6064-plugin-nonsecret-prompts](https://github.com/qq01-hub/magpie/tree/fix/6064-plugin-nonsecret-prompts) | `0a3675b2982b00826a2174c701cabd98ba1ebde0` |
| [#6710](https://github.com/yetone/magpie/issues/6710) | Git skills bind and regroup from verified source records or checkout origins without merging case-sensitive refs and paths | [fix/6710-git-skill-grouping](https://github.com/qq01-hub/magpie/tree/fix/6710-git-skill-grouping) | `a54aed801e7b38c55d06f3f5608c24f1d5502230` |
| [#3604](https://github.com/yetone/magpie/issues/3604) | Explicit LobeHub requests display their client name and color icon without adding a configurable agent | [fix/3604-lobehub-client](https://github.com/qq01-hub/magpie/tree/fix/3604-lobehub-client) | `2604c2d62db54a0aa0a46218b3f01d01e0da641e` |
| [#7134](https://github.com/yetone/magpie/issues/7134) | Kettu requests display their own name and PNG icon while retaining the original script budgets | [fix/7134-kettu-client](https://github.com/qq01-hub/magpie/tree/fix/7134-kettu-client) | `5ac4ae5e76909fdb8e6bb5fdba8d9aa6057daea0` |
| [#7430](https://github.com/yetone/magpie/issues/7430) | Synthetic model lists retain live reasoning and output limits, distinguish Infinite and carry known output into both OpenCode formats | [fix/7430-synthetic-live-models](https://github.com/qq01-hub/magpie/tree/fix/7430-synthetic-live-models) | `c12f672a94aa646b0d68ccd6be6c39d5385c5a60` |
| [#6576](https://github.com/yetone/magpie/issues/6576) | Desktop advisor aliases and tier picks resolve with the original caller, preserving key limits and subscription preference | [fix/6576-desktop-advisor-alias](https://github.com/qq01-hub/magpie/tree/fix/6576-desktop-advisor-alias) | `1d8000afe69a00a64147d77f0c916c6ca499d990` |
| [#6377](https://github.com/yetone/magpie/issues/6377) | Plugin MCP icons survive full Library reads and partial refreshes while explicit icons retain precedence | [fix/6377-library-plugin-icons](https://github.com/qq01-hub/magpie/tree/fix/6377-library-plugin-icons) | `dd038edada1ad91f0565a394b2f4315c527cac09` |
| [#6565](https://github.com/yetone/magpie/issues/6565) | Buddy usage reads only newly completed lines, retaining reply replacement, disk continuation and old-cache rebuilding | [fix/6565-workbuddy-incremental-calls](https://github.com/qq01-hub/magpie/tree/fix/6565-workbuddy-incremental-calls) | `45b6cfb32b783ee84ff2edd38966b7fda1a38eb5` |
| [#7221](https://github.com/yetone/magpie/issues/7221) | Completely shared subgroups keep their metadata and chosen paths, and their models display under the correct headings | [fix/7221-nested-group-trace](https://github.com/qq01-hub/magpie/tree/fix/7221-nested-group-trace) | `cb43fe135b1b3f9562c9ca4e2b29fc5c48947aad` |

## Reproduction and checks

- #8068: browser-only and browser-plus-skills directories both incorrectly detected OpenCode on the old code; both pass after the change, with a real `opencode.json` still detected
- #8068 review: 6 top-level tests and 2 subtests passed, covering Pi, sign-in/library-only directories, shared folders and the agent registration guards
- #7165: old code returned the Azure HTTP 400 or masked SSE error; after the port, Responses and all four Messages client/upstream streaming combinations retry without the rejected sealed item and remember the refusal for the next turn
- #7165: positive and unrelated-error negative cases passed; the same focused selection passed under `-race`
- #7432: old code returned an empty session id or the same opening hash for separate Grok sessions; the port passes separate-session, resume, title, subagent, post-turn-summary and header-precedence checks
- #7432: plugin and remote-gateway propagation passed, including a focused `-race` run; Windows and Linux gateway vet checks passed
- #6456: the old camelCase tool name fails against a local MCP fixture accepting the reported name; the end-to-end search, result parser, endpoint selection and key-only path pass after the fix
- #7385: the reported LobsterAI URL yields no callback port on old code, versus `42961` after the fix; non-loopback, ordinary-query and fragment-port checks pass, as does the simulated plugin's pasted-callback flow
- Combined review after integrating the three production changes: 14 top-level gateway tests, 92 tests and subtests passed, zero failures or skips in 20.402 seconds
- Final five-fix integration at `4728b2a2a82baaa02b562f5e9d09edb4787d9209`: 27 top-level tests, 107 tests and subtests across agent, gateway and provider, zero failures or skips; this adds the four GLM checks and callback/key sign-in paths to the checks above
- Windows `go vet ./...`, Linux `go vet -tags nogui ./...`, Windows GUI-enabled build and Linux `nogui` build passed with the final production changes; all 16 changed Go files are formatted and the complete diff passes `git diff --check`

Machine-readable results: [first five-fix batch](issue-fixes-20261128-r3-batch1-results.json)

The combined selection was:

```text
go test -tags nogui -run '^(TestForeignReasoningWords|TestForeignSealAzureOtherResource|TestForeignCarriedReasoningAzureOtherResource|TestForeignCarriedReasoningAskedAgainWithout|TestForeignSealAcrossGroupMembers|TestForeignSealAzureSummary|TestForeignSealGrokRelayCompactionBlob|TestGrokBuildNamesItsSession|TestPluginToldTheSession|TestRemoteMagpieKeepsTheSession|TestConversationID|TestDshMarksSessionAndCompaction|TestDshMarksOnEveryRoute|TestPluginSessionEdges)$' -count=1 -json -timeout 3m ./internal/gateway
```

The individual commit messages record the reproduction, test commands and limitations
Original patch authors retain their credit: yhzang for #7165, Wyman Wen for #7432, zzy7517 for #6456 and isdoge for #7385; MizuIro-H is credited for #8068's root-cause analysis

## Second batch checks

- #6614: the unchanged provider predicate leaves the selected picker value at `deepseek/pro`, instead of the matching `magpie/deepseek/pro` option; the regression passes with normalization
- #6614: separate temporary configs cover `custom:magpie` and `custom:magpie-anthropic`, main/profile values, entry lookup, model follow, restore, disconnect/reconnect and an unrelated custom provider; four focused top-level tests passed in the independent review
- #6614: a broader selection also exposed the existing Windows `TestHermesProfileModels` fixture sharing the package's temporary `LOCALAPPDATA` instead of its individual test home; it was left out of the focused passing selection and was not changed by this fix
- #7434: the old parser renders the unlimited sentinel as `$100000000.00`; the port retains the original author's change and adds numeric/string cases for finite, zero, negative and exact-sentinel values, including the immediate finite neighbors
- #7434: the balance readers and account management-key path passed, as did the CLI single-card refresh test; no live balance request was made
- Combined at `39181a4fba1602c7f710185e3eed9ac1d20a1cc0`: all seven selected top-level tests and twelve subtests passed across agent and provider, with zero failures or skips
- #5330: old code persists a generic `plugin@user` identity that the provider-specific quota card cannot find; after the original author's patch, the right card receives its timestamp and the same user in two plugins remains separate
- #5330: five provider tests passed under `-race`, plus the tray-follow and menu-bar notification tests; the independent integration review passed all six selected provider/tray tests
- #5330: old generic identities cannot safely be attributed to a particular plugin and are left to expire; the next answer records the correct plugin-specific identity
- #7072: unchanged code asks a plugin returning a structured empty reply 15 times in the fixture; after the fix it asks twice, and an ordered group asks its first member once and last member twice
- #7072: the contributor's substring matcher was shown to mistake unrelated metadata for the error kind and miss JSON whitespace; the supplement parses `error.diagnostics.kind` in JSON and SSE data, including unnamed SSE events, only for busy 5xx statuses
- #7072: the five new tests and their ten subtests passed under `-race`; the independent integration review passed 12 top-level tests and 19 subtests, covering ordinary errors, authentication, existing empty-reply behavior, group failover and patience limits
- #7091: old code cannot discover metadata exposed only at the path-appended AS URL; the contributor's fallback completes discovery, client registration, PKCE authorization and token exchange against a loopback fake
- #7091: a separate regression exposed candidate fields leaking into later responses, producing a false complete metadata record; each candidate now starts empty, and candidate-order and incomplete-response checks pass
- #7091: the MCP OAuth package passed, both new metadata tests passed under `-race`, and cancellation of an in-flight callback was checked for new and existing sign-ins without saving a token
- Credits for this batch: OrionRay / bigbitbox for #6614's analysis, Zigzag / GoodbyeAntwerpBlue for #7434, 李群 / jytx for #5330, 陈新 / CodeNovice2017 for #7072 and sean / xinnyu for #7091
- Final integration at `3a8aa5d88f245416f4c2f2c30e6ff2af7bdef4f3`: 37 top-level tests, 76 tests and subtests across agent, provider, gateway, gui, mcpauth and cli; zero failures and zero skips
- Windows `go vet ./...`, Linux `go vet -tags nogui ./...`, a GUI-enabled Windows build and a Linux `nogui` build passed for the second batch; changed Go files are formatted and `git diff --check` is clean

Machine-readable results: [second five-fix batch](issue-fixes-20261128-r3-batch2-results.json)

## Third batch checks

- #3057: the original `/v1/models` handler response with the drawers header has no `output_modalities`; the port returns exactly `["image"]` and retains the existing discovery fields
- #3057: the HTTP regression and hidden-drawer neighbor passed under `-race`; the independent integration review passed all four selected tests, including model/key and account restrictions for image/video lists
- #3057: the test network guard refused a background models.dev lookup during the restriction tests; assertions used local fixtures and passed without contacting it
- #3057: frankdzh's production patch retains its author; tests and documentation are a separate Q commit
- #5229: three new tests fail against the old code because set/drop, keep-pick and profile fill change or inspect the first duplicate instead of the last effective entry
- #5229: readers and writers now use the same selection, including effort and route migration; entries with `insert`, another plugin's name, or metadata without config are not selected for config writes
- #5229: the independent integration review passed 14 top-level tests and 16 tests/subtests, zero failures or skips, covering duplicate entries, home-layer overrides and restoration of the user's model
- #5229: a broader DSH-only run passed with two platform skips; an accidental whole-agent-package run was stopped and is not counted as verification
- #5229: yetone's root-cause analysis is credited in the production commit
- #4028: the old code rejects the reported LAN host and IPv6 literal and misses IPv6 loopback when checking for the gateway itself; the new tests reproduce each failure
- #4028: review found that the first port check still accepted zero and empty IPv6 ports; separate failing tests now guard those cases, with 1 and 65535 accepted and malformed, out-of-range and unbracketed IPv6 addresses refused
- #4028: a CC Switch config.json fixture is read, offered and imported into the saved provider with its original endpoint and key; imported IPv6 and mapped-IPv6 loopbacks on the gateway's port remain excluded
- #4028: four targeted tests and 15 subtests passed under `-race`; the independent integration review passed 10 top-level tests and 25 tests/subtests with zero failures or skips
- #4028: tests used temporary files only, with no LAN service or real relay access; the earlier broader import selection had one fake-shell platform skip, excluded from the final selection
- #2690: the original code retains Chinese title prefixes and loses localized JSON title fields; tests using the reporter's original input fail before the change
- #2690: review exposed nondeterministic precedence between `title` and `Title`; exact string keys now win, including an empty canonical title, followed by fixed localized order and sorted case-insensitive aliases
- #2690: Chinese, Japanese and Korean JSON/text labels, non-string fields, bare labels and words beginning with a label are covered; the HTTP handler regression verifies a Japanese title returned through the actual Codex title path
- #2690: four focused tests passed under `-race`; the independent integration review passed 10 top-level tests and 12 tests/subtests with zero failures or skips, including title settings, description fallbacks and no-description recording
- #2690: DamienZCollins is credited as the reporter; the original fork was unavailable, so this is an independent implementation
- #6148: nlimpid's original patch is ported with its author preserved; only the optional literal grammar-format wording is added to the refusal matcher
- #6148: the old code returns the reported HTTP 400 for both streamed and complete responses; both pass with the change, including the same-API retry, custom call reconstruction, immediate conversion on the next request, grammar description and unchanged accepting-upstream controls
- #6148: invalid grammar definitions and unrelated errors remain unmatched; three focused top-level tests and four subtests passed
- #6148: the custom-tool selection passed under `-race`, including call IDs, raw custom input preservation and streaming grammar conversion
- Final integration at `494c7478d0eacc5eaa2bd958563f78d86d5857cf`: 41 top-level tests, 64 tests and subtests across agent, provider and gateway; zero failures and zero skips
- Windows `go vet ./...`, Linux `go vet -tags nogui ./...`, a GUI-enabled Windows build and a Linux `nogui` build passed for the third batch; changed Go files are formatted and `git diff --check` is clean

Machine-readable results: [third five-fix batch](issue-fixes-20261128-r3-batch3-results.json)

## Fourth batch checks

- #4312: the old reader returns an empty model and effort and the Web path when Web is an empty template and Desktop has a start model
- #4312: the selected read profile now supplies the model, effort, path, Joined and Check; explicit leads, nonempty or unreadable Web files, shared settings and home patches retain precedence, and writes keep Web-first order
- #4312: 21 top-level tests and 31 tests/subtests passed in the independent integration review, including the prior #5229 duplicate-entry fix, home overrides, restore and sibling-profile diagnostics
- #7564: both host and WSL named-profile target lookups return nil on the old code; the new tests perform real instruction, MCP and skill writes in temporary profile directories
- #7564: six integration tests passed, covering custom PI_CONFIG_DIR, separate base/work/lab profiles, preservation of user instructions, old/current/unknown OMP versions, and a fake WSL probe whose version is independent of the host
- #7564: the extra TestEveryAgentInLibrary guard fails for an unrelated DS Bot target on both the fixed base and this branch; that baseline failure is recorded separately and not counted as a passing check
- #7019: the old relayed Responses path hands Codex limit:8.0 under whole, byte and decimal-boundary stream cuts; normalization at tool_search_call creation returns limit:8 in each case
- #7019: translated Chat controls, fractional numbers, numbers inside query strings, a string-valued limit and raw custom-tool inputs retain their existing behavior; the targeted race run passed
- #7019: non-streaming Responses passthrough remains outside searchTidy and is unchanged by this fix; the local tests do not establish live Codex or MCP service behavior
- #2815: the old CLI reader recognizes the installed version in a fake latest-channel Caskroom but offers no updater; the new end-to-end reader test obtains the latest version and the command brew upgrade --cask claude-code@latest
- #2815: the original TestHowInstalled has its existing Unix-symlink skip on Windows; the added isolated test runs on this host and executes no brew, installer or real client
- #7019 and #2815: their combined integration review passed 15 top-level tests and 25 tests/subtests, including tool-search rendering, IDs, custom input and CLI version parsing
- #7431: the original Qwen target has no MCP path; the port writes HTTP, SSE and command servers to the existing settings file and carries disabled tools in excludeTools
- #7431: tests cover aliases, preserved user settings/comments/servers, same-name adoption, removal, environment references, fake WSL and actual writes after QWEN_HOME changes with the old file unchanged
- Credits: Junfei-Z for #4312, bsbandwl for #7564's source analysis, timonwong for #7019, GinWU for #2815 and Wyman Wen for #7431
- Upstream was inspected again at 1038298ffabb167600a2afe8ddb4a165a2794541; the affected production logic remains present there, and the intervening Claude artifact-rule change is separate from Qwen's MCP format selection
- Final integration at `efa5a1a23b63a67cba53b9a94c03f1eb55271993`: 46 top-level tests and 66 tests/subtests across agent, library and gateway passed, with zero failures and zero skips
- Windows `go vet ./...`, Linux `go vet -tags nogui ./...`, a GUI-enabled Windows build and a Linux `nogui` build passed for the fourth batch; changed Go files are formatted and `git diff --check` is clean

Machine-readable results: [fourth five-fix batch](issue-fixes-20261128-r3-batch4-results.json)

## Fifth batch checks

- #3735: the original parent/fork fixture counts 250 input / 25 output / 30 cache-read tokens on the old code instead of 150 / 15 / 20
- #3735: the author port passes that fixture but still fails full-JSON marker parsing and binary-cache resume; the follow-up handles reordered fields and stores ForkHead in the serialized rare state
- #3735: ordinary subagents, marker words inside normal text, replay/revision, parser invalidation and state schema controls pass; the same parent/fork fixture's AgentCalls returns exactly two calls with the correct totals, through its existing cross-file message-id deduplication
- #3735: 15 selected top-level tests passed under race detection, plus a final race run of the call-detail assertion
- #6696: JSON and SSE requests through two loopback gateway hops reproduce the loss of a client-edited {} argument to a 4,613,947-character cached output; both pass after content matching
- #6696: changed tool names and text force complete history, while unchanged content and recognized metadata omissions retain delta; final_answer omission is accepted but commentary remains distinct
- #6696: long identifiers and many digests fail the old cache size accounting and pass the corrected accounting; scope, expiry and per-route limits pass, and 11 selected top-level tests passed with -race -count=3
- #5221: raw, wrapped and streamed InternLM errors fail classification on the old code; the author's message substring rule also misclassifies unrelated errors, so the final change accepts only the two explicit code/type values
- #5221: real loopback HTTP/SSE handler tests verify zero retries asks one key and records failRefused without resting the account; message-only, quoted and similar codes stay ordinary errors, and focused race tests pass
- #7563: all three new project tests fail on the old code; native skill links, stdio/HTTP/SSE servers and plugin parts are written and removed successfully after the change
- #7563: existing user skills, servers, comments, compatibility files and PI_CONFIG_DIR settings remain intact; the project-native paths follow the checked oh-my-pi source at e045d0a1208f6395f9d5ab85faf9ebc74a7eba4e and do not rely on shared-skill support
- #2835: the old code rejects complete editor SDK caches; the author port falsely accepts another platform and still hides independent SDKs beside Cindy, both reproduced before the follow-up
- #2835: eight selected tests cover complete/partial/wrong-platform caches and Cindy coexistence; Node platform/architecture names are used, and unsupported targets remain absent
- #2835: native Linux/macOS and the newer Claude Linux-musl-specific cache variant were not exercised; the implementation recognizes the exact OS/architecture cache target
- Credits: Silas Su for #3735, Xulin Wu for #6696, Hermes Agent for #5221, bsbandwl for #7563's source references and WenyuRiver for #2835
- The affected production files have no intervening changes between the fixed base and inspected upstream 1038298ffabb167600a2afe8ddb4a165a2794541
- Final integration at `df13223fe1107ea1da68126f17084e2c84ee96ae`: 53 top-level tests and 117 tests/subtests across sessions, gateway, agent and library passed, with zero failures and zero skips
- Windows go vet ./..., Linux go vet -tags nogui ./..., a GUI-enabled Windows build and a Linux nogui build passed; changed Go files are formatted and git diff --check is clean

Machine-readable results: [fifth five-fix batch](issue-fixes-20261128-r3-batch5-results.json)

## Sixth batch checks

- #2609: a copied test helper waits for the parent to move its running executable, then calls the production Executable function; the old implementation returns the obsolete launch path for both short and over-MAX_PATH Unicode destinations
- #2609: the port retains the original author and the supplement uses existing x/sys/windows bindings with a bounded UTF-16 buffer; the independent integration review passed six top-level tests and eight tests/subtests, including invalid handles, the unchanged-path fast path, Scoop junctions and unresolved symlink fallback
- #2609: only temporary test executables were moved; no live Magpie binary, installer or elevated operation was run, and native Linux/macOS execution remains untested
- #6064: both PackyAPI method shapes from egenchen/magpie-packycode at 2a2eeaa fail with gave no key on the old host; the real plugin host now retains the separate key input and saves the expected key and metadata
- #6064: the two integration tests also cover one-time secret entry, secrets excluded from metadata, no-authorize/no-prompts methods and rejection when the key is absent; no real account or browser sign-in was used
- #6710: the old code leaves an identical Git skill unbound and regroups zero skills; local Git repositories now exercise exact copies, older copies, recorded sources, typed locks and real checkout origins
- #6710: review reproduced two additional mistakes in the initial implementation, merging Release/release or Kit/kit and accepting a stale full-URL From record; case-sensitive refs and Git paths remain separate, expired locks cannot claim unrelated skills, and the checkout's actual origin is retained
- #6710: the final independent integration review passed 13 top-level tests and 20 tests/subtests in 175.597 seconds, including existing GitHub and well-known grouping; an earlier worker selection exceeded its three-minute timeout and is not counted as a passing run
- #3604: the old registry has no named LobeHub metadata or embedded icon and does not normalize the mixed-case UA; the author port adds only a request client, with no configuration writer
- #3604: the original SVG fails the existing icon-sharpness guard; the supplement removes two shadow filters while preserving paths and gradients, then verifies the actual embedded asset's HTTP status, MIME and color-icon metadata
- #3604: five selected Go tests passed in independent integration review; headless Chromium and WebKit each passed English and Chinese routing-page checks for the name, icon URL and decoded image, and screenshots were inspected
- #7134: the original registry leaves Kettu under its raw name with a generic icon; the port retains the contributor's captured UA and logo, and adds request-client-only checks plus the actual embedded PNG's HTTP status, MIME and 64-by-64 decoding
- #7134: the first patch exceeds the existing core script budget by nine bytes; the supplement retains both original budgets and removes only separators' spaces in the unchanged PNG icon set
- #7134: six focused Go tests pass; Chromium and WebKit pass in English and Chinese, and removing only the PNG registration reproduces the generic cube in Chromium before restoration passes
- Credits for these changes: 3ZEROS12 for #2609, C0nstFoL for #6064's diagnosis and fixture source, yetone for #6710's diagnosis, Hardy for #3604 and Amigoer for #7134
- Final integration at `1ebc87d614a5253bca64abbbb07e9066008dfccc`: 30 top-level tests and 39 tests/subtests across update, plugin, library, agent and gui passed, with zero failures and zero skips
- The combined routing-client browser test passed all four Chromium/WebKit and English/Chinese cases with both new icons present, and the final WebKit screenshot was inspected
- Windows go vet ./..., Linux go vet -tags nogui ./..., a GUI-enabled Windows build and a Linux nogui build passed; all 82 changed Go files are formatted and git diff --check is clean
- Upstream was checked at 1038298ffabb167600a2afe8ddb4a165a2794541; the only intervening change in these production files adds Orca to the configurable-agent list, separately from the request-client entries

Machine-readable results: [sixth five-fix batch](issue-fixes-20261128-r3-batch6-results.json)

## Seventh batch checks

- #7430: the reporter's two Synthetic records reproduce identical names, missing Infinite reasoning and output, and stale regular-model effort and output metadata on the fixed base
- #7430: only the official Synthetic host interprets its reasoning_parameters, supported_features and max_output_length fields; live effort lists now survive catalog decoration, with missing fields retaining fallback
- #7430: review reproduced fractional and out-of-range output limits overwriting valid legacy data, plus known output being lost in OpenCode when context is unknown; positive integral limits are bounded by the architecture's int range and both OpenCode formats preserve the output
- #7430: the actual Gateway Handler is checked for plain, OpenCode 1 and OpenCode 2 lists using the real registered OpenCode builder; context zero follows the checked OpenCode 1.19.6 default, rather than inventing an Infinite context window
- #7430: 22 top-level tests and 40 tests/subtests passed in independent integration review, with zero failures or skips; the affected catalog, provider and agent checks also passed under race detection
- #6576: MaJingyiFighting's port retains its author and reproduces Desktop aliases leaking upstream, cross-provider advisor loss and missing auto-group stand-ins on the old implementation
- #6576: review adds exact caller-token/path precedence, unrelated-UA controls, subscription-bridge alias preservation and effort/context suffix handling, with real handler regressions for allowed and denied keys
- #6576: another review regression demonstrates that resolving a missing auto group before the original caller is available selects the first relay instead of the caller's own subscription; parsing now retains the identifier until the caller-aware request stage
- #6576: the accepted subscription-priority fixture first verifies the account, caller identity and direct stand-in selection; earlier experiments lacking the subscription's live catalog are not counted as valid red evidence
- #6576: 15 top-level tests and 41 tests/subtests passed in independent integration review, with zero failures or skips; eight focused top-level tests passed with -race -count=3 in 27.040 seconds, including zero advisor calls for disallowed keys and one for allowed keys
- #6377: KiraMio's patch retains its author and makes plugin-generated MCP servers inherit their installed icon only when there is no Library override; cached files use the existing validated icon endpoint
- #6377: review reproduces icon loss after an MCP-only refresh and applies the same enrichment to that response; the actual HTTP endpoint is checked for PNG content and rejection of encoded path traversal and non-image extensions
- #6377: four top-level tests and six tests/subtests passed in independent integration review; the selected GUI checks also passed under race detection, and 32 Chromium/WebKit tests cover five languages, real image loading, explicit overrides, unsafe-file fallback and partial refresh with no remote network access
- #6377: both browser screenshots were saved and the final WebKit screenshot was inspected by the main reviewer
- #6565: WorkBuddy, WorkBuddy AI and CodeBuddy now use the existing incremental line reader; the old code reparses the complete file after an append and counts a complete JSON object before its terminating newline, both reproduced before the fix
- #6565: the last reply ID and model-config stamp survive disk continuation, repeated reply updates mark the old call for replacement, and cloned carry preserves earlier returned calls; truncation, rewriting, ownership changes on append and parser-version upgrades rebuild the shard
- #6565: 18 top-level tests and 34 tests/subtests passed in independent integration review; the final selected new and adjacent tests also passed with race detection at one and four CPUs, including all three version-zero shard rebuilds
- #6565: unnecessary global timezone changes in three existing all-history tests caused a race on the unchanged base; their removal is a separate test-isolation commit, with the original failure and repeated passing checks retained
- #6565: a single measured append to the same real CodeBuddy fixture took 437.82 ms before and 2.01 ms after at 10 MiB, and 6410.53 ms before and 6.87 ms after at 100 MiB; this is a local one-iteration comparison, not a statistical throughput claim or proof of constant cost, since call copying and persistence still depend on history size
- #7221: the reporter's four-group configuration reproduces a missing fully shared Review group and the wrong chosen-member path; nested rule fixtures also reproduce losing Review/Inner ancestors
- #7221: group metadata retains primary and shared positions in source order, and a selected rule carries its full group path while execution candidates remain deduplicated; six real Handler and adjacent tests passed in independent integration review and under repeated race detection
- #7221: browser fixtures use the serialized real Handler traces; visual review then reproduced a third Review model sitting under Task low, fixed by keeping display rows together by their existing paths without changing execution order
- #7221: 44 Chromium/WebKit checks passed across five languages and wide/narrow viewports, including group ownership, unique model rows, fold behavior, multiple routes and the existing group-fold case; main review inspected the wide WebKit shared-group screenshot and narrow Chromium nested-group screenshot
- Credits: zent00 for #7430's original response and diagnosis, MaJingyiFighting for #6576's patch, KiraMio for #6377's patch, yetone for #6565's diagnosis and Riablo for #7221's report and configuration
- All five Go changes together at `b6076ac4b` passed 65 top-level tests and 127 tests/subtests across catalog, provider, agent, gateway, gui and sessions, with zero failures and zero skips
- The final UI follow-up at `2928a1cac` changes no Go source; four combined-head Chromium/WebKit cases passed, followed by successful GUI-enabled Windows and nogui Linux builds
- Windows go vet ./... and Linux go vet -tags nogui ./... passed; all 107 changed Go files are formatted and git diff --check is clean
- The independent UI branch's generated asset archive matched its source tree and served successfully, and its original core-script budget remained unchanged

Machine-readable results: [seventh five-fix batch](issue-fixes-20261128-r3-batch7-results.json)

## Environment and limits

Windows 11 amd64, Go 1.26.8 and Bun 1.4.5
The Go SDK, `GOCACHE`, `GOMODCACHE`, `GOPATH`, `GOTMPDIR`, `TEMP` and `TMP` are on D:
Tests use temporary homes, local fixtures and loopback servers
No real client, installer, elevated action or live gateway restart was used
Real Azure/Grok accounts and Linux/macOS runtime behavior were not tested
No full repository suite was run in this round, to avoid tests that launch desktop applications or restart services
This report describes completed checks only and does not claim a fully passing repository suite

#8010 and #8071 were not reproduced on the fixed base and are not counted as fixes
#8070 remains deferred while the T3 Code title source needs a representative local database fixture
