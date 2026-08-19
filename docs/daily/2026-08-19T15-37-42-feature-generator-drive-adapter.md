---
name: generator Drive 保存の KISS / Issue 化準備
date: 2026-08-19T15-37-42
session_id: none
branch: feature-generator-drive-adapter
prev: none
---

## 1. Summary

`feature/generator-drive-adapter` で Drive 保存アダプタ側の境界を先に固定し、DESIGN/README と generator todo を **transcript なしで実装できる KISS 粒度**に整理した。
あわせて Drive writer sociable unit の repo root 読み取り依存を除去してテストを安定化し、`./scripts/test-unit.sh` と `./scripts/generator/check-static.sh` が pass する状態にした。

## 2. Changes

- `apps/generator` 側で Port / secret inject / Drive 保存の枠を commit 1 で固定
- `DESIGN.md` / `README.md` / `contracts/drive-layout.md` と generator の todo を commit 2 で KISS 化し、Issue 化の分割単位を明確化
- `gdrive/writer_sociable_unit_test.go` の repo root 取得を `runtime.Caller` 基準に変更し、パス組み立ての揺れを除去

### Commits

- e8ddcfa
- 601277a
