# 安定 E2E fixture（人手 upload）

本番 Worker の `DRIVE_FOLDER_ID` **直下**へ、次の2 file をそのまま置く（sub folder なし）。

| file | 役割 |
|------|------|
| `4d22034f-66a5-42bc-81e1-775e5e82af2f.json` | 原稿（`date` = 2026-09-04。generator system e2e 1 回通しの成果物を契約形へ追随） |
| `4d22034f-66a5-42bc-81e1-775e5e82af2f.wav` | 同上 produce の結合 WAV（`durationSec` と一致） |

方針: `docs/decisions/2026-08-31T00-12-01-feature-playback-e2e-deploy.md`  
検証上書き: `docs/decisions/2026-08-31T00-22-00-feature-playback-e2e-deploy.md`（ProduceEpisode / TextWriter Domain Rule + `WriteEpisode`）  
朗読 field: `docs/decisions/` の manuscript body opening/ending 契約  
配置契約: `contracts/drive-layout.md`

本 fixture は `ManuscriptDraftFromWriterOutput`（`manuscript_draft_limits`）と完成稿の `WriteEpisode`（schema + stem 一致 + 非空音声）を通過する完成成果物である。日次 produce が増えてもこの pair は残す（本番一覧に見える）。

`body.opening` = 定型挨拶 + intro、`body.ending` = closingSummary + 定型締め。本 pair の opening は旧 produce（挨拶のみ）のまま残している。intro 付き全文は次の produce で揃う。
