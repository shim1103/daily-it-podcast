---
name: develop 追従の conflict 解消と epic→integration PR 準備
date: 2026-10-02T19:08:05
session_id: なし
branch: refactor/generator-go-performance
prev: 2026-09-23T19-23-27-refactor-generator-go-performance.md
---

## 1. Summary

`origin/develop` を epic へ取り込み、yesterday half-open 窓・retry 結線と Application 側 I/O fan-out を両立させた。並行して `3:workflow/pr --epic` に shim 手動 merge 後の Issue close・branch 削除を正本へ足した。

## 2. Changes

1. apply-standards を current-repo へ再反映した
2. `agent-standards` 側で `/pr --epic` を更新した（shim が merge、message `merge` で close/delete）。`release` の branch 削除と `lifecycle/release` §7 の close 主体も揃えた
3. `origin/develop` merge の conflict を解消した（Composition の Application 合成を維持し、displayLoc / retry / since-until を取り込み、HN/Lobsters の errgroup を載せ直した。`docs/lessons/index.md` は develop の削除に従った）
4. pre-commit の generator / playback unit が緑であることを確認した

### Commits

- `959f1b6`
