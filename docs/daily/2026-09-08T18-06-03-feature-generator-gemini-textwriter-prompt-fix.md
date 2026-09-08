---
name: generator-draft-rate を api 入力化し gemini の brief prompt を 9/10 まで調整
date: 2026-09-08T18:06:03
session_id: none
branch: feature/generator-gemini-textwriter-prompt-fix
prev: なし
---

## 1. Summary

`generator-draft-rate` に入力 `api`（`cursor` | `gemini`、既定 `cursor`）を足し、1 dispatch = 1 API で draft PASS 率を測れるようにした。rate 計測 test を `DRAFT_RATE_API` で分岐（key env と TextWriter 実物を差し替え、環境要因除外は `*cursorapi.Error` / `*geminiapi.Error` の `Op=="do"` で API 非依存に統一）、cursor 専用だった test file / 関数 / `-run` filter を汎用名へ rename。`workflow_dispatch` の入力定義は default branch のものが読まれるため yml + script を master へ先行配置（PR #138、merge は shim）。gemini の default prompt を `api=gemini` dispatch で実測しながら調整し、9/10（閾値 0.9）到達（run 34207700009）。数値 range（`manuscript_draft_limits.go`）は不変で、`constants.TextWriterBriefPrompt` の指導文と `build.marshalWriterOutputExample` の形式例だけを直した。session 途中で shim から 3 点の指摘（不要 test・prompt の裸の数字・lane の冗長）を受けて是正。

## 2. Changes

- **api 入力化**: `generator-draft-rate.yml` に `api`（`type: choice`）、`TEST_GEMINI_API_KEY` 注入、`DRAFT_RATE_API` / `API` env 伝播。`draft_rate_system_test.go`（`cursorapi_` から rename、関数 `TestDraftRate...`）に `resolveDraftAPITarget` / `isEnvUnreachable`。script 2 本の doc header と `-run` filter を追従。判断は Decision `2026-09-08T07-40-00`。
- **master 先行配置**: PR #138（yml + script 3 file のみ）を `drafts/draft-rate-api-input` から作成。CI 4/4 pass、shim が merge。`gh pr merge` は hook で block されるため shim 手動。
- **prompt 調整の実測**: `api=gemini` dispatch を 11 回反復。0/5 → 4/5（run 34204186623、threshold 0.8）→ 一旦後退（裸の数字を消した refactor で 3/10）→ 9/10（run 34207700009、threshold 0.9）。各 run の落ちた field と対応修正は Decision `2026-09-08T07-40-00` §4 の表が SSOT。到達 prompt の要点: 短い field の実質下限・上限を `{{X_MIN/MAX}}` ではなく `{{X_TARGET}}` placeholder として扱わせる（TARGET は正本連動・境界から離れる）、intro / closingSummary は 3 文固定、`# Self-check` と `# 提出前の必須修正手順` を完了条件付きの 1 loop へ統合、形式例の intro / closingSummary / title を目安長へ。
- **shim 指摘の是正**:
  - 不要 test: 本 session で足した `TestTextWriterBriefPrompt_*` 4 本（prompt の指導文文字列を assert に写経）を削除し、pre-branch の 3 本（`build.ComposeBrief` が消費する placeholder の構造契約を見る）へ戻した。
  - prompt の裸の数字: 「上限マイナス 20 文字」「`{{TOTAL_MIN}}` + 400」等のマジックナンバーを全廃。`{{X}} + 400` の算術は build が評価せず生文字列で LLM へ渡っていた。狙い値は `{{X_TARGET}}` placeholder（`manuscript_draft_limits.go` 由来）へ。
  - lane KISS/DRY: `generator-lane.md` の済み項目を 1 行へ、D 表の解決済み行を削除。実測の run ID・反復詳細は Decision へ寄せた。
- **shim の session 外 commit**: `03feada`（`# Length strategy` に「ソース素材が薄い場合は推測で文脈を補足せず topic に採用しない」を追加）。以降の prompt 修正はこの推測禁止と整合する形（「伸ばす」= ソース事実の詳述に限定）にした。
- push: `origin/feature/generator-gemini-textwriter-prompt-fix`（SSH auth のため sandbox 無効で実行）。develop / master への feature PR は create-pr flow で作成予定。

### Commits

- `d29fb25`
- `6bca8a5`
- `1402176`
- `ca49f13`
- `5474a0c`
- `91cb732`
- `12ad7cb`
- `5414af7`
- `03feada`（shim）
- `4848dd9`
- `2925aa9`
- `0a2de92`
- `0508861`
- `1af18c3`
- `8382b9b`
- `c2374fb`（PR #138 で master へ）
