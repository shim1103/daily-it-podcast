---
name: develop SSOT 化と rewrite 前 mock 文書の削除
date: 2026-09-01T12:46:00
session_id: none
branch: cursor-docs-ssot-develop-cleanup-9567
prev: 2026-08-31T02-19-00-drafts-playback-ops-ssot.md
---

## 1. Summary

rewrite 以前の mock MVP 文書（`docs/SPEC` 系・旧 `docs/DESIGN`・旧 architecture 図）を削除し、README / DEPLOY の develop SSOT 表記と generator lane の済み要約を現状へ揃えた。git history 改変は行わない。

## 2. Changes

1. `docs/SPEC.md`・`docs/PROPOSAL.md`・`docs/MEMO.md`・`docs/DESIGN.md`・`docs/diagrams/architecture.mmd` を削除した。
2. README から archive 行を除去し、branch 表の `develop` を SSOT とした。
3. DEPLOY から `ProduceEpisode.Run` 未完の stale 行を削除し、`workflow_dispatch` の default branch を `develop` と明記した。
4. `generator-lane.md` の済み項目（error 3 層・本番 produce workflow）を要約へ移し、未完了は System のみ残した。

### Commits

- `581d7f0`
- `07379fc`
- `a3fe7d6`
