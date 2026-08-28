---
name: Cursor CLI の GitHub Actions capability probe を実測完了
date: 2026-08-28T19:06:23
session_id: none
branch: feature/generator-cursor-cli-github-actions-probe
prev: なし
---

## 1. Summary

`docs/tasks/todo/generator-cursor-cli-github-actions-probe.md` の probe を実測完了した。GitHub Actions runner 上で Cursor CLI の公式 install・現行 argv の可否・child environment 要件を観測し、`--sandbox enabled` が runner の AppArmor 非対応で起動不能（`--sandbox disabled` なら 4 case 全て `run_exit=0`）、認証は `CURSOR_API_KEY` 環境変数のみで成立、`HOME`/`PATH`/`TMPDIR` は成功に非依存、を確定した。manager（non-edit / audit）が executor へ実装を委譲、reviewer が code-review + simplify を査読、複数回の再実装を audit した。probe 5 commit と除去 1 commit を push。probe 専用 code・workflow・task file を除去し、lane の未完了マーカーを完了へ更新した。実測結果と検証は PR body に残す。

## 2. Changes

1. probe 構成: `scripts/generator/probe-cursor-cli.sh`（install → 現行 argv 実行 → stdout/stderr の byte・行数のみ数値化 → 失敗段階の機械分類 → metadata 出力）、`probe-cursor-cli-classify-test.sh`（`classify_failure` / `build_cursor_args` / `build_env_prefix` の最小 TDD runner、bats 不在のため自作）、`.github/workflows/probe-cursor-cli.yml`（`workflow_dispatch` のみ、matrix 4 case `[full, no-home, no-tmpdir, minimal-path]` + `fail-fast: false`）。
2. `workflow_dispatch` は定義ファイルが default branch に無いとトリガ API が 404。workflow 定義のみを PR [#78](https://github.com/shim1103/daily-it-podcast/pull/78) で `master` へ先行 merge（`a0670e4`、yml 1 ファイルのみ cherry-pick）。実行は `--ref feature/...` 指定で feature branch の中身が走る。
3. dispatch 1回目（run 33157428638）: 4 case job success だが観測結果を `GITHUB_STEP_SUMMARY` へしか出さず `gh` から回収不能 → `append_summary` を `tee -a "${GITHUB_STEP_SUMMARY:-/dev/null}"` へ変更し stdout 併用（`e958637`）。
4. dispatch 2回目（run 33157854280）: 4 case とも `install_exit=0 / run_exit=1 / stderr 259 byte / stdout 0`。本文非出力のため原因不明。`PROBE_REVEAL_STDERR=1` opt-in で stderr 先頭 300 byte を1回だけ開示する分岐を追加（`bdc1028`、既定は非開示で Contract 3 維持）。
5. dispatch 3回目（run 33160392008）: 開示された stderr = `Error: Sandbox mode is enabled but not available on this system. Sandbox failed to start, possibly due to AppArmor configuration.`。原因は `--sandbox enabled` の環境非対応で確定。暫定分類 `entitlement`（stderr byte 数ヒューリスティック）は誤りだった。
6. `PROBE_SANDBOX_MODE`（`enabled`（既定・constants.go 再現）/ `disabled` / `off`）で `--sandbox` 指定を切替える分岐を `build_cursor_args` に追加（`170de4f`）。`PROBE_REVEAL_STDERR` とは独立した一時措置で、各々分岐ごと削除可能。
7. dispatch 4回目（run 33161307905、`PROBE_SANDBOX_MODE=disabled`）: 4 case 全て `install_exit=0 / run_exit=0 / stdout 306〜314 byte 1 行 / stderr 0 / 分類 success`。GHA で Cursor CLI が成功する条件を確定。
8. local 補助調査: scratchpad の使い捨て script（TDD 対象外）で `cursor-agent` を3段階（最小 argv / +model / constants.go 完全版）実行。完全版は local で成功（`{"type":"result","subtype":"success","result":"2"}`）。local 成功時は `~/.cursor/cli-config.json` に `authInfo` が存在。CI runner にはこれが無く `CURSOR_API_KEY` 経路のみになる差を切り分けた。`~/.cursor` の一時退避は harness sandbox が書き込みを弾き未実行（ディレクトリは無傷を確認）。
9. Phase 4 除去（`7d4a5f0`）: probe script 2本・workflow・task file を `git rm`、`docs/tasks/todo/generator-lane.md` の該当行を `- [ ]` → `- [x]` + 実測要点の注記へ。`git grep` で `probe-cursor-cli` / `PROBE_SANDBOX_MODE` / `PROBE_REVEAL_STDERR` の参照ゼロ、通常 workflow（`test-unit.yml` / `test-integration.yml`）無変更、`constants.go` / `text_writer.go` 無変更を確認。
10. 検証: `classify` test は各再実装後に全 pass（最終 26 件）。`bash -n` 両 script OK（shellcheck 未導入のため未実行）。dispatch 4回の run URL と case 別 result を回収。全 run の job log を `grep` し `CURSOR_API_KEY` は `***` マスクのみ、prompt 本文・install 出力・stderr 本文（開示分を除く）が非出力であることを確認。
11. 残課題: `master` の `probe-cursor-cli.yml`（PR #78 分）の revert PR が別途必要。PR Deviations に follow-up として記録。`constants.go` の `SandboxValue = "enabled"` を GHA 前提で見直すかは Decision 化 scope（本 probe の Out of Scope）。

### Commits

- `038f784`
- `680565e`
- `e958637`
- `bdc1028`
- `170de4f`
- `7d4a5f0`
