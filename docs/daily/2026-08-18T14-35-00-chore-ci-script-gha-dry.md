---
name: Unit/GHA 入口を片系 script に分け、GHA でも Unit を回す
date: 2026-08-18T14:35:00
session_id: none
branch: chore/ci-script-gha-dry
prev: なし
---

## 1. Summary

test 手順を `scripts/generator/` と `scripts/playback/` の片系入口へ分け、root 集約と hook / GHA はそれを呼ぶだけにした。GHA に Unit workflow を足し、Integration 専用だった remote gate を上書きする decision を残した。以降の script / docs 更新は commit → push → PR 作成まで実施した。

## 2. Changes

1. 片系 script（generator: static / unit / integration、playback: unit / integration）を追加し、root 入口は委譲だけにした
2. `check-generator-unit-coverage.sh` を `scripts/generator/test-unit.sh` へ移して削除した
3. pre-commit が `check-static.sh` と `test-unit.sh` を順に呼ぶようにした
4. `.github/workflows/test-unit.yml` を追加し、既存 Integration workflow は script 呼び出しのままにした
5. DESIGN / README の gate 規則と実行手順を更新した
6. composer 契約の sociable unit（`.shell`）を scripts 隣に置いた

### Commits

- `ee77a98` chore(ci): unify static/unit/integration entrypoints
- `d9c30ef` chore(generator): adjust unit coverage gate and generator tests
- `917b429` docs(tasks): add generator/playback static gate tasks

PR: [#25](https://github.com/shim1103/daily-it-podcast/pull/25)
