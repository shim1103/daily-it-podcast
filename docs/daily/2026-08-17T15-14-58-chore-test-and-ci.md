---
name: generator Unit に coverage と depguard 層 gate を入れる
date: 2026-08-17T15:14:58
session_id: none
branch: chore/test-and-ci
prev: 2026-08-15T23-17-00-chore-test-and-ci.md
---

## 1. Summary

generator の Unit（commit）gate に statement coverage 90%（除外後）と depguard（strict allow、Infrastructure は Port のみ）を入れた。playback と Integration / GHA は触らず、`origin/develop` を merge して lessons 衝突を両方残した。`shim gh create-pr` は `workflow/project.toml` 欠如で失敗し、`gh pr create` は使っていない。GHA `test-integration` は success。

## 2. Changes

1. `golangci-lint` + `depguard` を generator に追加し、`test-unit.sh` / pre-commit から呼ぶ
2. `check-generator-unit-coverage.sh` で除外後 statement 90% を Unit 入口に載せる
3. AgentSecrets proxy の未cover分岐を sociable unit で埋める（除外後 92.1%）
4. DESIGN / README / decision に gate 規則と実行手順を分ける
5. `origin/develop` の lessons 27–29 と本 branch の 27–32 を番号振り直さず併記する

### Commits

1. `de4ff24` — test(generator): AgentSecrets proxy の未cover分岐を Unit で埋める
2. `93ba69a` — chore(ci): generator の Unit gate に depguard と coverage を入れる
3. `d4a0039` — docs: generator の coverage と層依存 gate を DESIGN に書く
4. `3a0d832` — docs(log): セッションログ
5. `8daa286` — docs(log): develop の lessons 追記と衝突を両方残す
