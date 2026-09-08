## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor Cloud Agents REST 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。

### 済み（要約）

1. `ProduceEpisode.Run` / Broad Integration / error 3 層 / 本番 produce workflow
2. 情報源3 Adapter（HackerNews / Lobsters / ITmedia）を composite `ItemSource` へ結線。Broad Integration が3源 double で緑
3. 原稿 TextWriter を Cursor CLI から Cloud Agents REST（`manuscript/cursorapi`）へ移行。`commandlaunch` / `processenv` / CLI install を廃止
4. System e2e 1 回通し（`TestProduceEpisodeSystem`）と dispatch 専用 test（`TestGeminiTTSRate` / `TestDraftRate` / `TestGeminiAPISmoke`）を配置

### 済み（要約・続き）

5. System — `generator-system.yml` suite 本体・`TEST_*` 登録・e2e 1 回通しの実 dispatch 確認（run 33857369881 PASS、Drive 実到達、episodeId `8ff4177b-26fe-4036-ab7b-d2a4e9e7639d`）。運用方針は `DEPLOY.md` §5
6. 原稿 TextWriter fallback 本実装 — `manuscript.TextWriter`（`errors.Is(port.ErrSourceExhausted)` で高々 1 回切替）/ `geminiapi.TextWriter`（generateContent 1 回 + retry）を SU / Narrow 込みで実装。切替 trigger は `cursorapi` create の 401/403 と 400 + `usage_limit_exceeded`。`generator-system.yml` run 34133797530 で fallback 経路の e2e を実証（Cursor 400 → Gemini 原稿 → Drive 到達）。判断は Decision `2026-09-07T19-06-00` / `2026-09-07T23-30-00`
7. `geminiapi` 実 API 疎通 smoke — `generator-geminiapi-smoke.yml`（dispatch 専用、`TEST_GEMINI_API_KEY`）。yml は master（PR #135）、test は `system && ratemeasure` tag で gate 外。`--ref <fallback 実装 branch>` で回す
8. `generator-draft-rate` を `api`（cursor | gemini）入力化。gemini の default prompt を 9/10 まで調整（数値 range 不変、指導文のみ）。判断・実測は Decision `2026-09-08T07-40-00`

### 未完了

（現在なし。rate 計測 follow-up は下記 D 表が index）

### D（未決・未実測・文案）

再発する判断の正は `docs/decisions/`。ここは残りの未実測・文案のみ index する。

| topic | 概要 |
|---|---|
| Prompt / limits 文案・数値 | 尺モデルは確定済み（正は `entities/constants/manuscript_draft_seconds.go` / `manuscript_draft_limits.go`）。topic 数 6/8/10・全体尺 14/16/18 分へ一度伸ばしたが、本番 produce run 34209712652 が gemini fallback 経路で HTTP 429（出力 token 増で free-tier rate limit に到達）で失敗したため旧尺（topic 3/5/7・全体 8/10/12 分）へ戻した。尺を再度伸ばすなら先に Gemini quota（TPM/RPD）か有料 tier を手当てする |
| 挨拶文案 | Opening/Closing 定数は date placeholder 入り template で確定。実運用での文言微調整のみ残 |
| composite の source またぎ sort | 3 情報源で `OccurredAt` 順の混在が起きる。dedup は `SourceID` が全源で異なるため不要。時系列 sort を Application/Composition のどちらで持つかは別判断（事実: 現状は登録順 concat のみ） |
| 別媒体の報道源追加 | Publickey / InfoQ / はてブ IT 等は各々専用 Adapter を新設（`infrastructure/<媒体>/`。RSS 汎用 Adapter は作らない）。RSS 2.0 parse の重複が三度現れたら共通化を検討（未実測） |
| 議論 comment のスレッド深掘り | HN / Lobsters は 1 階層のみ取得（上限は Adapter stub 定数）。ネストした議論を辿るかは未決 |
| TextWriter の web_fetch 実測 | `links:` の URL を Cloud Agents 経路が実際に fetch できるか未実測。できない場合の補完は TextWriter 経路の内側（Application には置かない） |
| no-repo 原稿品質・token・job timeout | Cloud Agents no-repo が ask 相当の断片になるか、Pro 日次消費、SSE 待ちが GHA job に収まるかは未実測 |
| TTS rate 実 dispatch | `TestGeminiTTSRate` が実 API でまだ走っていない。1 度 dispatch して尺帯ごとの PASS 率・所要を台帳化する |
| `interactionResponse.Status` | 現状未使用。`status != "completed"` の扱いは未決 |
| gemini prompt 修正の cursor 影響 | 済み 8 の prompt 修正が cursor 側 PASS 率を落としていないか未確認（`generator-draft-rate -f api=cursor` 再 dispatch）|
| Cursor 枯渇 error code の網羅 | 番兵 wrap は 401/403 と 400 + `usage_limit_exceeded` のみ（Decision `2026-09-07T23-30-00`）。`billing_*` 等の別 code が出たら都度 Decision を継ぐ |
| Gemini fallback 発火の観測 | 現状 `logManuscriptSourceSwitched` の stderr 1 行のみ。切り替え成功時は原稿が出るので痕跡が薄い。GHA run summary への出力 / Drive metadata への provenance / 構造化 log 基盤の導入は未決 |
| Gemini free-tier RPD の実運用値 | 公称 RPD≈1,000 だが実測で下振れ報告あり。1 日 1 回 produce + draft retry 最大 5 でも問題ないはずだが未確認 |
| Cursor 復帰の運用気づき | 毎回 primary（Cursor）を先に試すので枠復活後は自動で戻るが、「毎日 Gemini に落ちている」状態を運用が能動的に気づく手段は未整備 |


### 方針 index

閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。
