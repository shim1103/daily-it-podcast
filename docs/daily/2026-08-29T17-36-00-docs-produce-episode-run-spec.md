---
name: ProduceEpisode Run spec の A/B 固定と C/D scope-split
date: 2026-08-29T17:36:00
session_id: none
branch: docs/produce-episode-run-spec
prev: なし
---

## 1. Summary

ProduceEpisode.Run の Ask→scope-split で A 契約（application/build stub・entities 定数・WriterOutput wire）と B Decision 13 本を branch 上に固定し、C Issue 4 本と lane D 概要を整備した。brief Prompt は entities/constants 1 本化と field ごとの limits placeholder 注入へ整理。pr-completion で 4 commit・rebase develop・PR 作成まで進めた。

## 2. Changes

1. static / unit gate pass（generator coverage 91.1%）。pre-commit は node 22 で playback biome が必要だった。
2. `origin/develop` へ rebase 成功（conflict なし）。
3. PR #85 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `4b8c8af`
- `587bd68`
- `dd88ede`
- `7c70fd2`
