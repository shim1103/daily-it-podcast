---
name: playback Biome + tsc 静的検査導入
date: 2026-08-18T19:10:00
session_id: none
branch: pr-c-playback-biome-tsc
prev: none
---

## 1. Summary

`apps/playback` へ Biome（formatter + linter）と `tsc --noEmit` を導入し、`scripts/check-static.sh` から検知できる状態にした。`./scripts/check-static.sh` / `./scripts/test-unit.sh` の両方 exit 0 を確認した。

## 2. Changes

- `apps/playback/tsconfig.json` を新設（strict、noEmit、vanilla TS 向け。React/Next 専用オプションなし）
- `apps/playback/biome.json` を新設（formatter + linter、react 系 rule は無効のまま）
- `scripts/playback/check-static.sh` を新設し、`scripts/generator/check-static.sh` と対称の contract tag・構造で揃えた
- `scripts/check-static.sh` に playback 呼び出しを追加
- `apps/playback/package.json` に `typecheck` / `lint` / `format:check` script と `@biomejs/biome` / `typescript` を追加
- Biome format により既存 18 ファイル（`contracts/`・`worker/src/`）を機械的整形（ロジック変更なし）
- `code-review`（medium）と `/simplify`（4 観点並列 review）を実施し、`biome.json` の冗長な default 値明記を除去
- working tree 上で `.env.example` が削除された状態を検知したが、これは今回の変更と無関係な既存汚染（直前 `develop` 相当の内容が誤って working tree へ反映された事象に由来）と判明し、shim が `git checkout HEAD -- .env.example` で復元した
- 完了issue（`docs/tasks/todo/pr-c-playback-biome-tsc.md`）を削除した

### Commits

- 069fe5c
- 6a55541
