# feature: Cursor Cloud Agents REST TextWriter 本実装

## 1. Summary

このIssueでは、A stub の `cursorapi.TextWriter` に Cloud Agents REST（create → SSE → `result.text`）と Decision どおりの retry を実装し、SU / Narrow / CI 残骸掃除まで完了状態にする。完了後、`Write(brief)` が httptest double で非空断片を返し、production 結線が stub の `not_implemented` を返さない。

## 2. Context

- A（`manuscript/cursorapi` stub・`commandlaunch` / `cursorcli` / probe / CLI install 不在・Composition 結線）と B（Decision `2026-09-03T17-03-33`）は固定済み。
- 現状 `Write` は brief validate 後に `infraErr("not_implemented", ...)` を返す。
- Broad は既に `port.TextWriter` double で緑。本 Issue で Broad を本番 Adapter に載せ替える必要はない。
- C の旧案（abolition / CI を別 Issue）は本 1 file に束ねる。削除の多くは A 済み。残るのは script / coverage 文言・lane 追随と本実装。

## 3. Canonical Sources

- 移行判断 — `docs/decisions/2026-09-03T17-03-33-feature-generator-cursor-cli-to-http-api.md`
- 契約値（dir・constructor・定数・retry 上限）— `apps/generator/internal/infrastructure/manuscript/cursorapi/`
- Composition 結線 — `apps/generator/internal/composition/cursorapi.go` / `runtime.go`（`sharedHTTPClientWithoutTimeout`）
- Port — `apps/generator/internal/application/port/text_writer.go`（変更しない）
- HTTP Adapter 対称例 — `apps/generator/internal/infrastructure/speech/gemini/`
- test 方針 — `testing-strategy` SKILL
- 地図 — `DESIGN.md` §3

## 4. Scope

### In Scope

- `cursorapi.TextWriter.Write` 本実装（no-repo create、SSE 終端、`result.text`、Decision の retry）
- Sociable Unit（httptest / RoundTripper double）
- Narrow Integration（local httptest。secret なし）
- A 後に残る CLI 言及の掃除（例: `scripts/generator/test-unit.sh` コメント、必要なら gate needle）
- lane index への本 task 登録追随（未完了 checkbox）

### Out of Scope

- Domain Draft invalid / DomainError 変更
- `GET /v1/models` runtime 確認
- agent 再利用・repo 付き cloud・SDK sidecar
- System suite 本体（別 lane 項目）
- no-repo 原稿品質・token 消費・GHA job 所要の実測（D）
- GitHub Issue 化（別判断）

## 5. Contract

- `port.TextWriter.Write(ctx, brief) (string, error)` の signature は不変。
- 成功時は非空 text 断片。失敗時は `*cursorapi.Error`、断片は空。
- vendor 固有の agentId / runId / SSE event を Port へ露出しない。
- Composition は `NewTextWriter(sharedHTTPClientWithoutTimeout(), Reveal())` 形を維持する。

## 6. Constraints

- Decision の retry / timeout / no-repo / 毎回 create を破らない。契約値は A を参照し Issue へ写さない。
- `sharedHTTPClient()`（30s）を Cursor Adapter に渡さない。長時間・streaming は `sharedHTTPClientWithoutTimeout()`。
- POST create の 5xx / 曖昧 timeout を再試行して二重 agent を作らない。
- secret・応答本文を error message へ写さない（既存 infra Error 慣習）。

## 7. Acceptance Criteria

- [ ] AC-1: 成功 SSE（終端 `result` に非空 text）で `Write` が非空断片を返す（SU）
- [ ] AC-2: 空 brief は `validate_brief` 系 Infra Error、断片空（SU）
- [ ] AC-3: 401 / 403 / 400 は非 retry で Infra Error（SU）
- [ ] AC-4: 429 は有限回 backoff 後に成功または上限到達で Infra Error（SU。`MaxAttempts` は A 定数）
- [ ] AC-5: Do error / GET 5xx は +1 即再試行の契約どおり（SU）
- [ ] AC-6: POST create の 5xx は再試行しない（SU）
- [ ] AC-7: Narrow（httptest）が AC-1 相当の成功経路を通す
- [ ] AC-8: `go test ./...`（apps/generator）が pass
- [ ] AC-9: production/test に `commandlaunch` / `cursorcli` / `probe-cursor-cli` / `not_implemented` stub 経路が残らない（本実装後の `Write`）

## 8. Verification

```bash
cd apps/generator && go test ./internal/infrastructure/manuscript/cursorapi/...
cd apps/generator && go test ./test/ -count=1
cd apps/generator && go test ./...
rg -n 'commandlaunch|cursorcli|probe-cursor-cli|not_implemented' apps/generator --glob '!.git/**'
```

## 9. Dependencies

- A / B 完了（本 branch で固定済み）
- System suite・D 実測とは独立

## 10. Risks

- Cloud Agents が ask 相当の断片を返さない risk — Narrow/SU は double で契約を固定し、実 API 品質は D の手動 probe に残す
- SSE 実装ミスで共有 30s Client に戻す risk — Composition 結線と AC で `WithoutTimeout` を維持する

## 11. Notes

- follow-up（本 Issue 外）: no-repo 品質実測、Pro token、GHA job timeout、archive/delete、web_fetch 題材更新
