---
name: Cursor Cloud Agents REST TextWriter 本実装と pr-completion
date: 2026-09-04T00:00:51
session_id: none
branch: feature-generator-cursor-http-text-writer
prev: 2026-09-03T17-46-27-feature-generator-cursor-cli-to-http-api.md
---

## 1. Summary

A stub と B Decision（`2026-09-03T17-03-33`）で固定済みの transport 移行を受け、`cursorapi.TextWriter.Write` を Cloud Agents REST（no-repo create → SSE 終端 `result.text` → Decision §5 の retry）で本実装した。SU 16 本 / Narrow 2 本で AC-1〜AC-9 を固定し、`not_implemented` stub 経路を除去。issue-manager flow（manager plan → executor → reviewer → executor 再実装 → manager audit → issue file 削除）で進めた。その後、production file に残っていた `ForTest` 命名を cursorapi / gemini 両 Adapter から除去し、`produce` job に未設定だった `timeout-minutes` を明示した。

## 2. Changes

1. issue-manager flow: manager は read-only で plan / audit のみ。executor が TDD で本実装 + test、reviewer が must-fix 1（`Retry-After` 無限界・ctx 非対応）+ should-fix 8 を指摘、executor が再実装（`streamRetryKind` で transient / rate-limited を分離、`ctxSleep` で ctx 先行 cancel 観測、`MaxRetryAfter` クランプ、`run.id` 一本化、SSE parse 簡約）。
2. manager audit で AC-1〜AC-9 の充足を検証コマンドで確認（`go test ./...` / `go vet` / `-race` / `check-static.sh` 0 issues / coverage 92.4% ≥ 90% / needle grep 0 件）。issue file `generator-cursor-http-text-writer.md` を削除、lane index 未完了 1 を完了へ。
3. shim の指摘で `newTextWriterForTest` → `newTextWriter`（cursorapi）、`newSpeechSynthesizerForTest` → `newSpeechSynthesizer`（gemini）へリネーム。production code が test 用途を名前に出さない。挙動不変、SU が担保。
4. `.github/workflows/generator-produce-episode.yml` の `produce` job は `timeout-minutes` 未設定で GitHub 既定 360 分まで走りうる状態だった。Adapter が Client Timeout を持たず run 全体上限を job timeout に委ねる設計の受け皿として `timeout-minutes: 25` を追加。厳密値は no-repo run の実測（lane D）後に絞る。
5. `.claude/settings.json` の `defaultMode: bypassPermissions` 差分は本 task 無関係で、shim が別途 `5bbb49c` で commit 済み。
6. DESIGN.md §3 / lane は既に現行前提（Cloud Agents REST / `manuscript/cursorapi` / Decision 参照）を指しており、stub → 本実装で更新すべき差分なし。

### Commits

- `66db1ea`
- `1696875`
- `c09cfda`
- `03f1835`
- `73be459`
