# feature: 原稿 TextWriter fallback（Gemini secondary + 切り替え UseCase）本実装

## 1. Summary

このIssueでは、A stub の `application/manuscript.TextWriter`（切り替え UseCase）と `infrastructure/manuscript/geminiapi.TextWriter`（Gemini generateContent Adapter）に、Decision どおりの切り替え条件と retry を実装し、SU / Narrow まで完了状態にする。完了後、`ProduceEpisode` から見た原稿取得は、Cursor が create 401/403（`port.ErrSourceExhausted`）を返したときだけ 1 回だけ Gemini へ切り替わり、それ以外の Cursor error はそのまま伝播する。

## 2. Context

- A（`application/manuscript` stub・`application/port/errors.go` の番兵・`infrastructure/manuscript/geminiapi` stub と定数・Composition 結線・`cursorapi` の 401/403 wrap）と B（Decision `2026-09-07T19-06-00`）は固定済み。
- 現状 `manuscript.TextWriter.Write` と `geminiapi.TextWriter.Write` は zero value を返す stub。A の足場 test（`TestNewTextWriter_returnsNonNil` 等）が coverage gate を緑に保っているだけ。
- `cursorapi.TextWriter` の create 401/403 → `port.ErrSourceExhausted` wrap は A で実装済み・SU 緑。本 Issue で `cursorapi` は触らない。
- Broad Integration は `broadTextWriter`（単一 `port.TextWriter` double）を `ProduceEpisode` へ直接渡すため、UseCase を経由せず緑のまま。本 Issue で Broad を切り替え経路へ載せ替える必要はない。
- `GEMINI_API_KEY` は本番 GHA Secret 登録済み（TTS と共用）。新規 Secret は不要。

## 3. Canonical Sources

- 切り替え判断・provider 選定・retry 方針 — `docs/decisions/2026-09-07T19-06-00-feature-generator-text-writer-fallback.md`
- 番兵 error — `apps/generator/internal/application/port/errors.go`（`ErrSourceExhausted`。変更しない）
- 切り替え UseCase 契約値 — `apps/generator/internal/application/manuscript/text_writer.go`
- Gemini Adapter 契約値（endpoint・header・model・retry 上限）— `apps/generator/internal/infrastructure/manuscript/geminiapi/`
- Composition 結線 — `apps/generator/internal/composition/produce_episode.go` / `geminiapi.go` / `runtime.go`（`logManuscriptSourceSwitched`）
- Port — `apps/generator/internal/application/port/text_writer.go`（変更しない）
- HTTP Adapter 対称例 — `apps/generator/internal/infrastructure/manuscript/cursorapi/`（retry 分岐・`backoffSleepFn` seam・secret 非露出）
- test 方針 — `testing-strategy` SKILL
- 地図 — `DESIGN.md` §3

## 4. Scope

### In Scope

- `manuscript.TextWriter.Write` 本実装（primary → `errors.Is(err, port.ErrSourceExhausted)` なら `onFallback` + secondary。切り替えは高々 1 回）
- `geminiapi.TextWriter.Write` 本実装（generateContent 1 回、`candidates[0].content.parts[].text` 抽出、`finishReason` 検査、Decision の retry）
- Sociable Unit（`manuscript`: `port.TextWriter` fake の Spy。`geminiapi`: httptest / RoundTripper double）
- Narrow Integration（`geminiapi`: local httptest。secret なし。`apps/generator/test/geminiapi_narrow_integration_test.go`）
- A の足場 test（`TestNewTextWriter_returnsNonNil` / `TestCtxSleep_*`）を behavior test へ置き換え
- lane index への本 task 登録追随（未完了 checkbox → 完了）

### Out of Scope

- `cursorapi` の変更（401/403 wrap は A 済み）
- `port.ErrSourceExhausted` の定義変更、`port.TextWriter` signature 変更
- 3 provider 目の追加
- 構造化 log 基盤の導入・切り替え発火の能動通知（D）
- `gemini-2.5-flash-lite` の原稿品質・token 消費・尺下限割れの実測（D）
- Gemini 応答の `safetyRatings` / `promptFeedback` の詳細分類
- GitHub Issue 化（別判断）

## 5. Contract

- `port.TextWriter.Write(ctx, brief) (string, error)` の signature は不変。`manuscript.TextWriter` と `geminiapi.TextWriter` の両方がこれを満たす。
- `manuscript.TextWriter`: primary 成功 → その断片、secondary 未呼び出し。primary が `port.ErrSourceExhausted` → `onFallback` 1 回 + secondary 1 回、secondary の戻りを透過。primary の非枯渇 error → そのまま返し secondary 未呼び出し。切り替えは高々 1 回。
- `geminiapi.TextWriter`: 成功時は非空 text 断片。失敗時は `*geminiapi.Error`、断片は空。endpoint は `EndpointURLTemplate` + `ModelID`、key は `APIKeyHeader`。vendor 固有の応答構造を Port へ露出しない。
- Composition は `manuscript.NewTextWriter(newCursorTextWriter(...), newGeminiTextWriter(...), logManuscriptSourceSwitched)` 形を維持する。

## 6. Constraints

- Decision `2026-09-07T19-06-00` の切り替え条件（401/403 限定）・retry 方針・model 定数固定を破らない。契約値は A を参照し Issue へ写さない。
- `manuscript.TextWriter` は `cursorapi` / `geminiapi` を import しない（`errors.Is` と `port` のみ）。
- `geminiapi` は `speech/gemini` の重装 retry（`SynthesizeBudget` / call gap / 分単位 backoff）を流用しない。`generateContent` は idempotent なので 5xx も 1 回再試行してよい。
- `sharedHTTPClient()`（30s）を Gemini Adapter に渡さない。長文生成は `sharedHTTPClientWithoutTimeout()`。
- secret（`x-goog-api-key` 値・key を含む URL）を error message へ写さない。key は URL query に載せず header に載せる。
- 429 backoff は `Retry-After` delta-seconds を尊重し `MaxRetryAfter` で clamp。`MaxAttempts` は A 定数。

## 7. Acceptance Criteria

- [ ] AC-1: `manuscript` — primary 成功で断片透過、secondary 未呼び出し、`onFallback` 未呼び出し（SU）
- [ ] AC-2: `manuscript` — primary が `port.ErrSourceExhausted` wrap error → `onFallback` 1 回・secondary 1 回・secondary の戻りを透過、secondary へ渡す brief は primary と同一（SU）
- [ ] AC-3: `manuscript` — primary が非枯渇 error → その error を返し secondary 未呼び出し・`onFallback` 未呼び出し（SU）
- [ ] AC-4: `manuscript` — secondary も `port.ErrSourceExhausted` を返しても 3 つ目の呼び出しは無い（SU）
- [ ] AC-5: `geminiapi` — 成功応答（`finishReason: "STOP"` + 非空 text）で `Write` が非空断片を返す（SU）
- [ ] AC-6: `geminiapi` — `client.Do` error / 5xx は +1 即再試行（SU）
- [ ] AC-7: `geminiapi` — 429 は `MaxAttempts` まで backoff（`backoffSleepFn` seam で待ちを観測。上限到達で `*geminiapi.Error`）（SU）
- [ ] AC-8: `geminiapi` — 401 / 403 / その他 4xx は非 retry で `*geminiapi.Error`、断片空（SU）
- [ ] AC-9: `geminiapi` — `finishReason` が STOP 以外、または text が空/空白のみ → 非 retry で `*geminiapi.Error`（SU）
- [ ] AC-10: `geminiapi` — error message に API key 実値が出ない。key は `APIKeyHeader` に載り URL query に出ない（SU）
- [ ] AC-11: `geminiapi` — Narrow（httptest）が AC-5 相当の成功経路を通す
- [ ] AC-12: A の足場 test（`TestNewTextWriter_returnsNonNil` 等）が behavior test へ置き換わり、`go test ./...`（apps/generator）が pass
- [ ] AC-13: `manuscript` / `geminiapi` が Unit coverage gate（statement 90%）の分母で緑

## 8. Verification

```bash
cd apps/generator && go test ./internal/application/manuscript/... ./internal/infrastructure/manuscript/geminiapi/...
cd apps/generator && go test ./test/ -run Gemini -count=1
cd apps/generator && go test ./...
bash scripts/generator/test-unit.sh
rg -n 'cursorapi|geminiapi' apps/generator/internal/application/manuscript   # import してはいけない
```

## 9. Dependencies

- A / B 完了（本 branch で固定済み）
- D の実測（Gemini 原稿品質・token・尺）とは独立

## 10. Risks

- `gemini-2.5-flash-lite` が JSON 原稿を安定して返さない risk — SU / Narrow は double で契約を固定し、実 API 品質は D の手動計測に残す
- 切り替え発火が成功時に痕跡を残さない risk — `logManuscriptSourceSwitched` の stderr 1 行で最小限の観測を確保。恒久策は D
- `generateContent` 応答形の版差（`candidates` 構造）risk — Adapter 定数と parse を 1 箇所に閉じ、失敗は `*geminiapi.Error` で落として定数修正で対応

## 11. Notes

- follow-up（本 Issue 外・D）: Gemini 原稿の尺下限割れ率の計測（`generator-draft-rate.yml` 同型の dispatch）、切り替え発火の能動通知、Gemini free-tier RPD の実運用値、Cursor 復帰の運用気づき手段
- `manuscript.TextWriter` を将来 3 実装以上へ拡張する場合は decorator chain ではなく「優先順位付き slice + 番兵」を検討（本 Issue では 2 実装固定）
