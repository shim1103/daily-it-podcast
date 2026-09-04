# 安定 E2E fixture（人手 upload）

本番 Worker の `DRIVE_FOLDER_ID` **直下**へ、次の 2 file をそのまま置く（sub folder なし）。

| file | 役割 |
|------|------|
| `8ff4177b-26fe-4036-ab7b-d2a4e9e7639d.json` | 原稿（`date` = 2026-09-04）。`body.opening` / `body.ending` は `{ text, startSec }` object。`text` は TTS が読み上げる朗読全文（opening = 定型挨拶 + intro、ending = closingSummary + 定型締め）。 |
| `8ff4177b-26fe-4036-ab7b-d2a4e9e7639d.wav` | 上記 `durationSec`（518.56 秒）と一致する合成音声。 |

この pair は `generator-system`（`TestProduceEpisodeSystem`）が本番相当の経路で produce した完成成果物であり、`ManuscriptDraftFromWriterOutput`（`manuscript_draft_limits`）と完成稿の `WriteEpisode`（`manuscript.schema.json` + stem 一致 + 非空音声）を通過している。日次 produce が増えてもこの pair は残す（本番一覧に見える）。

`apps/playback/test/e2e/authenticated_playback.e2e.spec.ts` の `FIXTURE_*` 定数は本 json の値を写したもの。json を差し替えたら spec の定数も揃える。

方針: `docs/decisions/2026-08-31T00-12-01-feature-playback-e2e-deploy.md`
検証上書き: `docs/decisions/2026-08-31T00-22-00-feature-playback-e2e-deploy.md`（ProduceEpisode / TextWriter Domain Rule + `WriteEpisode`）
契約統一: merge(develop) `opening/ending` を `{ text, startSec }` へ統一（本 branch）
配置契約: `contracts/drive-layout.md`
