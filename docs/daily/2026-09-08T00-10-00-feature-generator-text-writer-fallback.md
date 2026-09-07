---
name: 原稿 TextWriter fallback を本実装し実 API smoke と system e2e で疎通実証、枯渇 trigger に 400 usage_limit_exceeded を追加
date: 2026-09-08T00:10:00
session_id: none
branch: feature/generator-text-writer-fallback
prev: 2026-09-07T20-46-32-feature-generator-text-writer-fallback.md
---

## 1. Summary

前 session で A/B/C/D として固定した原稿 fallback の C（実装 Issue `generator-text-writer-fallback.md`）を `/issue-manager` flow で完遂した。`manuscript.TextWriter`（`errors.Is(port.ErrSourceExhausted)` で primary→secondary を高々 1 回切替）と `geminiapi.TextWriter`（generateContent 1 回 + Do error/5xx を 1 回・429 を `MaxAttempts` backoff）を stub から本体へ、SU / Narrow / A 足場 test の behavior 化まで。続けて「fallback なしで `geminiapi` を実 API へ 1 往復させる最小 dispatch workflow」を master へ PR（#135、マージ済み）。その smoke run が `gemini-2.5-flash-lite` の 404 を検出したため `geminiapi.ModelID` を現行 GA 版へ差し替え。さらに `generator-system.yml`（e2e 1 回通し）が Cursor create の **HTTP 400 + `usage_limit_exceeded`**（Background Agent 利用枠喪失）で赤くなり、これを 401/403 と同じ枯渇シグナルとして番兵 wrap する後続 Decision（`2026-09-07T23-30-00`）を起こして実装。再 dispatch した system e2e が fallback 経路（Cursor 400 → Gemini 原稿 → Drive 到達）で緑になり実証完了。

## 2. Changes

- **C 実装（`/issue-manager`）**: `manuscript.TextWriter.Write` 本体（空 brief は `domainerrors.DomainErr(OpEmptyBrief)` で返す。`entities/errors` に `OpEmptyBrief` を追加）。`geminiapi.TextWriter.Write` 本体は `cursorapi` の `streamResult`/`fetchOnce`/`retryAfter`/`backoffDelay` と同型構造。SU は `cursorapi` の `fakeRoundTripper`/`sleepSpy` 様式を踏襲。Narrow は `test/geminiapi_narrow_integration_test.go`（TLS redirect、secret なし）。`constants_test.go` で `EndpointURLTemplate` が query を持たないことを固定。AC-1〜13 充足、`test-unit.sh` coverage 91.7% ≥ 90%、`check-static.sh` 0 issues、`go test ./...` 全緑。code-reviewer 査読は must-fix なし、should-fix 4 + nit を再実装で反映。
- **model 名 SSOT**: 当初 reviewer 指摘で Decision / lane に model 値を書いていたのを撤回。値は `geminiapi.ModelID` の 1 行のみ、docs は「`geminiapi.ModelID` を正とする」の参照だけ（DRY）。
- **smoke workflow（PR #135 → master マージ済み）**: `generator-geminiapi-smoke.yml`（`workflow_dispatch` のみ、`TEST_GEMINI_API_KEY`）+ `test/system/geminiapi_smoke_system_test.go`（`//go:build system && ratemeasure`）+ `test-geminiapi-smoke.sh` / `geminiapi-smoke-summary.sh`。yml は master 必須（`workflow_dispatch` は default branch の yml を要求）、test は二重 tag で `test-unit.sh`/`test-integration.sh`/`generator-system.yml`（`-tags=system` のみ）のどの gate からも除外。`--ref <fallback 実装 branch>` で回す。PR 差分は 4 file のみ（fallback 5 commit は含めない）。
- **ModelID 修正**: `generator-geminiapi-smoke` run 34129174375 が `gemini-2.5-flash-lite` の base alias で HTTP 404（NOT_FOUND）。web 調査で 2026-09 時点の現行 GA Flash-Lite を確認し `geminiapi.ModelID` を差し替え。再 dispatch run 34131089167 PASS（断片 14 文字 / 0.6s）。
- **400 usage_limit_exceeded の番兵化**: `generator-system.yml` run 34132242410 が `cursorapi: create_status: create status 400` で FAIL。切り分けのため `cursorapi` に `bodySnippet`（`speech/gemini` と同型）を足し create 失敗 error へ応答 body を載せた（commit `2a1723a`）。run 34132953055 で body が `{"error":{"code":"usage_limit_exceeded",...}}` と判明。Decision `2026-09-07T23-30-00`（先行 `2026-09-07T19-06-00` §4 を supersede せず観測点を 1 つ足す）を起こし、`cursorapi.createAgent` の番兵 wrap 条件へ `res.StatusCode == 400 && bytes.Contains(raw, []byte("usage_limit_exceeded"))` を追加。error JSON は struct parse せず code 文字列の存在だけを見る。
- **system e2e 実証**: 再 dispatch run 34133797530 PASS（426.5s、episodeId `182c84ff-563f-41b5-a62e-46ab0942ffa1`）。stderr に `manuscript text writer switched: from=cursor to=gemini reason=source_exhausted` が出て、Cursor 400 → Gemini 原稿 → TTS → OAuth+Drive 書込 → Drive 実到達を確認。
- **develop 取り込み**: `git merge origin/develop`（PR #134 の playback e2e fix 等）。`docs/lessons/index.md` が conflict（253 行 vs 5 行）。履歴を遡ると 2026-09-04 の merge `3f48fd9` で index.md が 247/248 行 → 0 行に切り詰められ、別 session が `fe465f2` で 258 行に union 復元、その後 `53c725a docs: by shim` が再び 0 行化、develop 側 `c5b5522` は 0 行ベースから playback-e2e lesson 3 本だけ append、という経緯。shim 判断で develop 側（5 行）を採用し feature 側 253 行は破棄。
- **lane 更新（KISS/DRY）**: `generator-lane.md` §済み 6 を AC 列挙から要約へ絞り、smoke workflow を §済み 7 として追加。D 表に「Cursor 枯渇 error code の網羅」を新設、model 値・番兵条件は constant / Decision 参照のみ。
- push: `origin/feature/generator-text-writer-fallback`（sandbox proxy が SSH auth を弾くため sandbox 無効で実行）。develop への PR 未作成。

### Commits

- `7feebb2`
- `aae76d5`
- `56add8c`
- `9da0921`
- `e5f0bae`
- `67618f7`（PR #135 で master へ）
- `85c01c2`
- `2a1723a`
- `e0e2c09`
- `f1925f0`（merge origin/develop）
- `2153bb5`
