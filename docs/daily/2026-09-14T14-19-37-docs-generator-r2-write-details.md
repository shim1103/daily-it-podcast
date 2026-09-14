---
name: R2 実施契約と mp3/R2 Issue 順を固定し prod binding を宣言した
date: 2026-09-14T14:19:37
session_id: 5379144b-ba3e-40ea-b2c6-ea2727ebf1ae
branch: docs/generator-r2-write-details
prev: なし
---

## 1. Summary

Drive 継承のまま R2 実施の形・error/retry・Issue 実施順を Decision と Issue file に固定し、配置契約を `episode-layout` へ移して R2 stub を置いた。GHA へ R2 Variable を登録し、Worker の prod `EPISODES` binding を `wrangler.jsonc` に宣言した（deploy はしない）。develop 先端へ同期し、encode-write 完了分と lane を整合させた。

## 2. Changes

- R2 topology / binding / credential / 書込意味論 / storage class を Decision 化。error・retry・観測非露出を別 Decision で具体化。
- storage 実施順を Issue 列 1–7 に固定。release 専用 Issue は作らない。batch は cutover Issue へ吸収。
- Account ID / bucket 名を GHA Variable として登録。Access Key Secret は shim 側・agent 確認時点では未登録。
- wrangler の冗長 why comment は削り、binding 契約値だけ残した。
- 検証: commit hook 経由の generator static/unit と playback format/lint/tsc/layers 緑。
- PR: https://github.com/shim1103/daily-it-podcast/pull/151

### Commits

- `ae60f9f`
- `c3bf12a`
- `b2e8497`
- `e1d4c08`
- `e75c3ba`
