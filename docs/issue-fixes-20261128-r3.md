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
