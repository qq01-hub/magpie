# Verification of ten unanswered issue fixes

The individual branches are based on `7281f97bed6d10f7160ed084fc4c98d316370cc5`
The combined branch is based on `5456743b731e1bc0d2c73b295d25cf70fb4e3c7e`
The combined production-code head tested is `537aac95a1b95a4dbe8f3e3dd6626034fca27d7a`
The original authors' commits were retained when importing their patches

## Changes and regression evidence

| Issue | Branch in qq01-hub/magpie | Behavior and verification |
| --- | --- | --- |
| [#7901](https://github.com/yetone/magpie/issues/7901) | [fix/7901-tolerant-model-modalities](https://github.com/qq01-hub/magpie/tree/fix/7901-tolerant-model-modalities) | Non-object modalities no longer reject a complete model list; the original test failed with `not a model list`, and twelve additional cases cover empty arrays, invalid shapes, top-level precedence, embeddings and repeated fields |
| [#7958](https://github.com/yetone/magpie/issues/7958) | [fix/7958-codex-store-cli-version](https://github.com/qq01-hub/magpie/tree/fix/7958-codex-store-cli-version) | The Windows Store app contributes its bundled CLI version to the async-compaction gate; the old reader had an empty version, while the fixed sync writes the percentage and still blocks old or unknown readers |
| [#7952](https://github.com/yetone/magpie/issues/7952) | [fix/7952-skill-sync-installed-files](https://github.com/qq01-hub/magpie/tree/fix/7952-skill-sync-installed-files) | Skill and local-plugin bundles exclude installed dependencies and Python bytecode before charging the size budget; lockfiles remain, stale Missing entries are filtered, and unchanged local installs survive Put |
| [#7954](https://github.com/yetone/magpie/issues/7954) | [fix/7954-gemini-json-arrays](https://github.com/qq01-hub/magpie/tree/fix/7954-gemini-json-arrays) | A Gemini request for any JSON value does not receive a second instruction limiting it to an object; Messages and Chat format-refusal paths passed after failing before the patch |
| [#7955](https://github.com/yetone/magpie/issues/7955) | [fix/7955-gemini-error-status](https://github.com/qq01-hub/magpie/tree/fix/7955-gemini-error-status) | Gemini replies use Google's status names and preserve known mid-stream status codes; direct encoding and real translated-stream tests failed with 502 in place of 429 before the patch |
| [#7957](https://github.com/yetone/magpie/issues/7957) | [fix/7957-gemini-schema-properties](https://github.com/qq01-hub/magpie/tree/fix/7957-gemini-schema-properties) | Schema property names stay names, nullable is translated, and literal default/const/enum/examples values are not rewritten; four literal-value cases also failed on the reporter's initial patch |
| [#7961](https://github.com/yetone/magpie/issues/7961) | [fix/7961-codex-other-input-items](https://github.com/qq01-hub/magpie/tree/fix/7961-codex-other-input-items) | Non-OpenAI requests omit unknown Codex history placeholders, and Copilot input-item enum refusals allow plain-text compaction; tests cover first-attempt forwarding, actual summary retry, protected history and unrelated refusals |
| [#7842](https://github.com/yetone/magpie/issues/7842) | [fix/7842-stream-repeated-tool-header](https://github.com/qq01-hub/magpie/tree/fix/7842-stream-repeated-tool-header) | A repeated open tool header at a new index continues streaming its arguments; removing the held-call guard makes the added interleaving regression fail |
| [#7843](https://github.com/yetone/magpie/issues/7843) | [fix/7843-custom-tool-numbers](https://github.com/qq01-hub/magpie/tree/fix/7843-custom-tool-numbers) | Custom tool input retains values such as 2.0 and 1.0.2; parser-to-encoder and parser-to-renderer tests failed on the old code for both direct and namespaced tools, while ordinary function integer conversion remains |
| [#7945](https://github.com/yetone/magpie/issues/7945) | [fix/7945-serve-new-account-models](https://github.com/qq01-hub/magpie/tree/fix/7945-serve-new-account-models) | The serving gateway continues fetching newly signed-in accounts after its initial lists are complete; the real server test failed with zero requests before the patch and passed after its next 60-second polling round |

## Environment and completed checks

- Windows amd64, Go 1.26.8, Bun 1.4.5
- Go SDK, module cache, build cache and test temporary directories are under `D:\AI\magpie-validation`
- Tests isolate HOME, USERPROFILE, APPDATA, LOCALAPPDATA and agent-specific configuration
- Upstream responses in regression tests come from local httptest servers
- The installed Windows Store Codex CLI `0.171.0-alpha.13.2` loaded the generated configuration through `features list` in a temporary Codex home
- `gofmt -l` on all changed Go files and `git diff --check` were clean
- `go vet ./...` on Windows passed
- `GOOS=linux go vet -tags nogui ./...` passed
- Windows GUI-enabled `go build` and Linux `go build -tags nogui` passed
- Combined gateway regression run passed 355 tests and subtests; its one Bun-dependent skipped test was then run with Bun present and passed
- The final combined regression selection across catalog, agent, library,
  gateway and provider passed 29 top-level tests and 74 total tests and
  subtests, with no failures or skips
- The new #7961 regression group passed under the race detector
- #7945 provider polling, retry and in-flight cancellation tests passed under `-race -count=20`
- #7945's real ListenAndServe regression passed under `-race` in 63.160 seconds

## Full Windows suite: incomplete and not passing

The full command was `go test -tags nogui -count=1 -timeout 45m -json ./...`
It ran from 2026-11-27 17:51:22 to 18:36:27 Asia/Shanghai and was stopped
after 45 minutes 5 seconds of overall elapsed time
Go's `-timeout` applies per package, so the command did not enforce this
overall time limit itself
Only the recorded Go process and its remaining test children were stopped

The saved JSON stream records 8,459 passing, 521 skipped and 42 failing test
events, including parent tests and subtests
Of 134 packages, 47 emitted a final package result: 36 passed, 6 failed and
5 had no tests; 4 packages were still running and 83 had not started
These partial counts are not a full-suite pass or complete coverage result

The same named failures for 40 of the 42 failure events were reproduced
on the unmodified combined baseline
Examples include Windows allocation budgets, Unix desktop-entry fixtures,
the unavailable `sh` command, Git absent from an isolated test PATH, and
the Windows listener-file error in the gateway handover test
The remaining two failures, `TestHermesProfilesFollowMagpieModel` and
`TestHermesProfilesAreNotListedAsAgents`, passed in focused runs on both
the baseline and the combined branch
Their failure in the full sequence remains unconfirmed; the focused
passes do not establish that the complete sequence is free of regressions

At interruption, the unfinished top-level tests were
`TestPiEnabledModelsLoseDeselected`, `TestGroupRuleCmd`,
`TestHandedOffGatewayTakesNoNewConnection` and `TestDesktopSubagentModelNotice`
The CLI had progressed beyond `TestEveryAgentFieldNameWritesNothing`
The handoff test also failed in focused runs on both the baseline and the
combined branch with the same Windows listener-file error

The [machine-readable results](issue-fixes-20261127-results.json) list all
42 failures, their baseline evidence, the 29 final focused tests and the
individual branch heads
Raw logs remain in `D:\AI\magpie-validation\logs`, including
`batch-full-windows.jsonl`, `batch-new-regressions-final.jsonl`,
`baseline-tail-confirm.log` and `batch-tail-focused.log`

## Limits

No real provider account was charged and no production gateway or agent configuration was changed
The speech, Gemini, Copilot and custom-tool protocol checks use representative local upstream replies
The skill tests exercise bundle collection, sharing and installation, not a live WebDAV service
No GUI asset changed, so browser UI suites were not run
Linux was compiled and vetted from Windows, not executed; macOS was not available
