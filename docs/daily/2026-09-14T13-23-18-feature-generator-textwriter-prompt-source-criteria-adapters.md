---
name: 新3源 List 実装・MaxStories 揃え・httpget/NI/cache 方針固定
date: 2026-09-14T13:23:18
session_id: none
branch: feature/generator-textwriter-prompt-source-criteria-adapters
prev: なし
---

## 1. Summary

Publickey / TechCrunch / CloudWatch の `List` を本番取得・写像へ置き換え、5 源の `MaxStoriesScanned` を揃え、engagement discard 専用 SU を外した。続く httpget helper・実境界 NI・接続 cache の方針を Decision / Issue / lane / DESIGN へ固定し、develop を merge して PR 準備した。

## 2. Changes

1. issue-manager で新 3 源 Adapter 実装。粗すぎた 3 源一括 commit を媒体ごとに分割して force-with-lease
2. shim 指示で MaxStories（RSS 系 15、HN/Lobsters 20）と discard 専用 test 削除
3. Decision `2026-09-14T13-06-11` と Issue `generator-item-source-httpget-ni-cache`、lane / DESIGN 更新
4. `origin/develop`（PR #149 mp3 encode）を merge。lane 未完了の conflict を解消

### Commits

- `d4238a6`
- `0565ef9`
- `7ac8a83`
- `c983c90`
- `c27b133`
- `069124c`
- `303a7b2`
- `9e7f028`
