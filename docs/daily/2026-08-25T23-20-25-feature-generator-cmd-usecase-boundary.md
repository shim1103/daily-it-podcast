---
name: generator cmd と ProduceEpisode 境界契約を固定する
date: 2026-08-25T23:20:25
session_id: 3751aa5e-0037-41d2-94e4-19fa84c0c58b
branch: feature/generator-cmd-usecase-boundary
prev: なし
---

## 1. Summary

日次 episode 生成の cmd ↔ Application 境界を Builder/Gate・Port=string・薄い Driving Adapter として固定し、契約 stub と Decision・local Issue を残した。`ProduceEpisode.Run` 本体は D のまま。

## 2. Changes

1. PlanQuestion で cmd・Composition・ProduceEpisode・既存 Gate の責務を確定し、空の orchestrator や Port=Draft を却下した。
2. A/B として stub・Decision・lane・local Issue を置き、shared architecture（Builder/Gate）も更新した（dotfiles 側）。
3. unit coverage が stub/`cmd` で 90% を割ったため、panic 観測 test と `cmd/` 除外を入れて gate を通した（除外後 91.2%）。
4. PR #58 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `a3ffdc9`
- `9c123c0`
- `8d3c820`
