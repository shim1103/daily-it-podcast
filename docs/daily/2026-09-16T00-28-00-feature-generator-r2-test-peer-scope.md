---
name: R2 test peer を httptest NI + local S3 到達に固定し PR 化する
date: 2026-09-16T00:28:00
session_id: pr-completion-generator-r2-test-peer-scope
branch: feature/generator-r2-test-peer-scope
prev: なし
---

## 1. Summary

C1（`generator-r2-test-peer-scope`）で local S3 peer 起動・playback `getPlatformProxy` infra・Integration 配線を入れたあと、Adapter を local S3 に刺す seam を本番 `r2` から外し、local S3 Verification を peer 到達のみへ閉じた。方針は新 Decision `2026-09-16T00-20-08`（create、旧 `12-02-48` の Adapter NI gate 条を supersede）。lane と参照 comment を latest policy に揃えた。

## 2. Changes

1. experimental local S3 peer（`test/r2locals3`）と playback `getPlatformProxy` infra を Integration gate から辿れる状態にした
2. 本番 Adapter の endpoint override / build-tag seam を削除し、local S3 NI を HTTP 到達観測のみにした
3. Decision `2026-09-16T00-20-08` を create し、generator / playback lane の正参照を更新した

### Commits

- `6a435e2`
- `015ced8`
- `6ff1529`
- `2d4d5c5`
- `87c9296`
- `03b86b5`
