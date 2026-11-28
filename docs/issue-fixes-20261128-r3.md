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
