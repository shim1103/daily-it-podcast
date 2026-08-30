# 安定 E2E fixture（人手 upload）

本番 Worker の `DRIVE_FOLDER_ID` **直下**へ、次の2 file をそのまま置く（sub folder なし）。

| file | 役割 |
|------|------|
| `eb14426e-2f4c-4157-9175-301ae4e7808d.json` | 原稿（`date` = 2026-08-30 = 配置時の前日 JST 定時暦日） |
| `eb14426e-2f4c-4157-9175-301ae4e7808d.wav` | ProduceEpisode と同順の固定尺 segment を結合した無音 WAV（`durationSec` と一致） |

方針: `docs/decisions/2026-08-31T00-12-01-feature-playback-e2e-deploy.md`  
検証上書き: `docs/decisions/2026-08-31T00-22-00-feature-playback-e2e-deploy.md`（ProduceEpisode / TextWriter Domain Rule + `WriteEpisode`）  
配置契約: `contracts/drive-layout.md`

本 fixture は `ManuscriptDraftFromWriterOutput`（`manuscript_draft_limits`）と完成稿の `WriteEpisode`（schema + stem 一致 + 非空音声）を通過する完成成果物である。日次 produce が増えてもこの pair は残す（本番一覧に見える）。
