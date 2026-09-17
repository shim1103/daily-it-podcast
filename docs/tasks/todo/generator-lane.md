## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → 原稿（Gemini/Cursor fallback）→ Gemini TTS（fallback）→ R2 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。

### 済み（要約）

1. `ProduceEpisode.Run` 本実装・5 情報源 Adapter を composite `ItemSource` へ結線
2. 原稿 TextWriter を Cursor Cloud Agents REST（`manuscript/cursorapi`）へ移行
3. System e2e 1 回通し・rate 計測 dispatch（`generator-system.yml` / `generator-tts-rate.yml` / `generator-draft-rate.yml`）を配置
4. ConcatWAV 後 ffmpeg で mp3 化する Adapter を実装・結線
5. `httpget` helper・source Narrow・接続 cache を整備
6. Storage を Google Drive から Cloudflare R2 へ完全移行（`EpisodeWriter` / `CompletedEpisodeLookup` 本実装、Drive codebase 削除）
7. TextWriter/TTS の model 切り替え fallback 本実装 — 原稿は `GEMINI_API_KEY`(free)→`CURSOR_API_KEY`→`SPARE_GEMINI_API_KEY`(paid, final) の 3 段、TTS は `GEMINI_API_KEY`(free)→`SPARE_GEMINI_API_KEY`(paid, final) の 2 段。system-test は `SYSTEM_TEST_TOPIC_COUNT` で topic 数を任意指定できる。判断は Decision `2026-09-16T00-39-21` / `2026-09-16T11-41-59` / `2026-09-17T10-36-05`

### 未完了

（storage 移行は完了。次の未完了項目は無し）

storage 実施順の正は Decision `2026-09-14T12-49-26` / `playback-lane.md` の実施順 index。R2 Adapter NI の正 peer は `2026-09-16T00-20-08`（本番口は `11-04-30`）。

### D（未決・未実測・文案）

再発する判断の正は `docs/decisions/`。ここは残りの未実測・文案のみ index する。R2 実施の形は `2026-09-14T11-04-30`（方針は `14-22-55`）。error は `11-19-21`。実施順は `12-49-26`。R2 Adapter NI peer は `2026-09-16T00-20-08`。cache 方針は `14-23-30`。

| topic | 概要 |
|---|---|
| composite の source またぎ sort | 情報源で `OccurredAt` 順の混在が起きる。dedup は `SourceID` が全源で異なるため不要。時系列 sort を Application/Composition のどちらで持つかは別判断（事実: 現状は登録順 concat のみ） |
| Links 件数→UseCase 先 fetch | `SourceBody.Links` 件数を見て TextWriter fetch 上限前に Application が先 fetch するかは未決。Adapter HTML scrape はしない（`14-41-02`） |
| 議論 comment のスレッド深掘り | HN / Lobsters は 1 階層のみ取得（上限は Adapter 定数）。ネストした議論を辿るかは未決 |
| TextWriter の web_fetch / url_context 実測 | `Detail.Links` / `Discourse.Links` を Cloud Agents / Gemini 経路が実際に fetch できるか・件数上限は未実測 |
| no-repo 原稿品質・token・job timeout | Cloud Agents no-repo が ask 相当の断片になるか、Pro 日次消費、SSE 待ちが GHA job に収まるかは未実測 |
| Gemini free-tier RPD の実運用値 | 公称 RPD≈1,000 だが実測で下振れ報告あり。1 日 1 回 produce + draft retry 最大 5 でも問題ないはずだが未確認 |

### 方針 index

閾値・入口の正は `DESIGN.md` / `DEPLOY.md`。音声形式は mp3（`contracts/episode-layout.md` / `2026-09-13T13-40-29`）。R2 実施の形は `2026-09-14T11-04-30`。実施順は `2026-09-14T12-49-26`（index は `playback-lane.md`）。
