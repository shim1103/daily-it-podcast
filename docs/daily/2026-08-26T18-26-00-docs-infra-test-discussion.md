---
name: generator Integration 収集境界と Narrow task 固定
date: 2026-08-26T18:26:00
session_id: none
branch: docs/infra-test-discussion
prev: なし
---

## 1. Summary

generator の Integration gate を secret なし Narrow に限定し、local_real 収集を build tag と local 入口へ分離した。方針は Decision 6本、vendor/local Narrow の達成契約は `docs/tasks/todo/generator-narrow-*.md` に置き、lane から Issue 化待ち表と decisions 再掲を削って DRY にした。

## 2. Changes

1. gate script / GHA / composer と `local_real` tag・local 入口を code で凍結した。
2. Decision 6本（gate・tag・秘密供給・System非CI・CDC非導入・Broad/SystemはD）を追加した。
3. DESIGN / README に地図を足した。
4. vendor gate 6本と local AgentSecrets 2本の task file を追加し、Unit/Narrow 責務分離を AC に入れた。generator/playback lane を薄くした。
5. develop を取り込み lessons / playback-lane 衝突を解消した。
6. PR #67 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `025b162`
- `13ce49d`
- `b240dc4`
- `6c93305`
- `cacd7d4`
- `1d675c9`
