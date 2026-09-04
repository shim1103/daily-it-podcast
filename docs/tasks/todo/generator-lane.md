## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor Cloud Agents REST 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。

### 済み（要約）

1. `ProduceEpisode.Run` / Broad Integration / error 3 層 / 本番 produce workflow
2. 情報源3 Adapter（HackerNews / Lobsters / ITmedia）を composite `ItemSource` へ結線。Broad Integration が3源 double で緑
3. 原稿 TextWriter を Cursor CLI から Cloud Agents REST（`manuscript/cursorapi`）へ移行。`commandlaunch` / `processenv` / CLI install を廃止
4. System e2e 1 回通し（`TestProduceEpisodeSystem`）と rate 計測 2 本（`TestGeminiTTSRate` / `TestCursorAPIDraftRate`）を配置。`generator-draft-rate.yml` は実 API dispatch で 3/3 PASS 確認済み

### 済み（要約・続き）

5. System — `generator-system.yml` suite 本体・`TEST_*` 登録・e2e 1 回通しの実 dispatch 確認（run 33857369881 PASS、Drive 実到達）。残りの rate 計測 follow-up（TTS rate 実 dispatch / draft 尺 A/B）は `docs/tasks/todo/generator-system-e2e-produce-episode.md`。運用方針は `DEPLOY.md` §5

### 未完了

（なし。rate 計測 follow-up は上記 child todo が index）

### D（未決・未実測・文案）

再発する判断の正は `docs/decisions/`。ここは残りの未実測・文案のみ index する。

| topic | 概要 |
|---|---|
| Prompt / limits 文案・数値 | 尺モデルは確定済み。残は実運用後の微調整 |
| 挨拶文案 | Opening/Closing 定数は date placeholder 入り template で確定。実運用での文言微調整のみ残 |
| composite の source またぎ sort | 3 情報源で `OccurredAt` 順の混在が起きる。dedup は `SourceID` が全源で異なるため不要。時系列 sort を Application/Composition のどちらで持つかは別判断（事実: 現状は登録順 concat のみ） |
| 別媒体の報道源追加 | Publickey / InfoQ / はてブ IT 等は各々専用 Adapter を新設（`infrastructure/<媒体>/`。RSS 汎用 Adapter は作らない）。RSS 2.0 parse の重複が三度現れたら共通化を検討（未実測） |
| 議論 comment のスレッド深掘り | HN / Lobsters は 1 階層のみ取得（上限は Adapter stub 定数）。ネストした議論を辿るかは未決 |
| TextWriter の web_fetch 実測 | `links:` の URL を Cloud Agents 経路が実際に fetch できるか未実測。できない場合の補完は TextWriter 経路の内側（Application には置かない） |
| no-repo 原稿品質・token・job timeout | Cloud Agents no-repo が ask 相当の断片になるか、Pro 日次消費、SSE 待ちが GHA job に収まるかは未実測 |
| draft 尺の下限マージン | `generator-draft-rate` 実測（run 33840526373）で default variant の 1 回が下限 +2 文字。variant `a` の A/B か `constants.TextWriterBriefPrompt` の detail 目安引き上げを検討 |

### 方針 index

閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。
