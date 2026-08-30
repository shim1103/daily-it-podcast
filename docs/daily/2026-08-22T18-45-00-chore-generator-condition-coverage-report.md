---
name: generator condition coverage report を追加
date: 2026-08-22T18:45:00
session_id: none
branch: chore/generator-condition-coverage-report
prev: なし
---

## 1. Summary

generator Unit の Boolean condition 実行状況を local で可視化する report を追加した。statement coverage の既存 gate は変更せず、condition coverage は参考指標として保持した。

## 2. Changes

1. `gobco v1.3.4` report は 15 package で `240/300`、80.0% を出力した。
2. report、statement coverage gate、script contract test、syntax check を実行した。
3. review 指摘により、temporary path を report path へ戻す置換を literal match に修正した。

### Commits

1. `71e504b`
