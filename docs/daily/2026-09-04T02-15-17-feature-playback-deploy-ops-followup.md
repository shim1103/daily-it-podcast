---
name: Playback 運用後続の docs・Decision・契約固定
date: 2026-09-04T02:15:17
session_id: none
branch: feature/playback-deploy-ops-followup
prev: なし
---

## 1. Summary

Playback 運用後続を docs / Decision / `wrangler.jsonc` の契約まで固定し、再 deploy・CD・DAST・Dependabot を完了条件から外した。lane の複合 checkbox を閉じた。

## 2. Changes

1. rollback 一次手段・observability 常時 ON・運用後続完了境界（再 deploy 非 scope）を Decision 3 file に分けた。
2. `observability.enabled` を契約へ載せ、DEPLOY に Logs / rollback 節を追加した。本番 deploy は実行していない。
3. lane から複合未完了行を外し、方針 index へ Decision 参照を足した。
4. PR #122 を作成した。`develop` merge で lessons conflict を解消し、削除済み達成契約を指す lane 行を落とした。

### Commits

- `cbeec99`
- `3ba83d1`
- `6008b64`
- `01effe3`
- `e87ac9c`
