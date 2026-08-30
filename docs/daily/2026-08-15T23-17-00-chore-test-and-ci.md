---
name: Unit/Integration gate と Vitest・Go test 足場
date: 2026-08-15T23:17:00
session_id: none
branch: chore/test-and-ci
prev: なし
---

## 1. Summary

Vitest（Playback）と Go test 入口、commit=Unit / push=Integration の CI gate（local hook + GHA）を導入した。README は使い方、DESIGN は規則に分業し直した。

## 2. Changes

1. `scripts/test-unit.sh` / `scripts/test-integration.sh` と git hook・`install-hooks.sh` を追加
2. `apps/playback` に Vitest（unit / integration project）を追加
3. GHA `test-integration.yml` を追加（push/PR で Integration）
4. gate script の Narrow Integration self-check を `test/` に追加
5. README（実行手順）と DESIGN（配置・gate 規則）の責務を docs-split に合わせて分離

### Commits

1. `b5409e0` — chore(playback): Vitest で unit/integration project を用意する
2. `6f39ee2` — chore(ci): commit は Unit、push は Integration の gate を入れる
3. `387903d` — docs: test の使い方と規則を README / DESIGN に分ける
4. （本 commit）docs(log): セッションログ
