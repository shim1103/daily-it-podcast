---
name: generator X adapter Unit Test の時刻依存を除去
date: 2026-08-22T18:40:00
session_id: none
branch: test/generator-time-determinism
prev: なし
---

## 1. Summary

generator の X adapter にある vendor error Unit Test から current wall-clock input を除去し、固定 UTC 時刻で error 契約を検証できるようにした。

## 2. Changes

1. GetXAPI と TwitterAPI.io の vendor error case は、nil result と error の既存 assertion を維持した。
2. `rg 'time\.Now' apps/generator --glob '*_test.go'` が match なしであることを確認した。
3. `./scripts/generator/test-unit.sh` が `-shuffle=on` と `-count=1` で pass し、除外後 coverage は `91.1%` だった。

### Commits

1. `90e53af`
