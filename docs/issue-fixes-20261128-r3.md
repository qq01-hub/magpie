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
