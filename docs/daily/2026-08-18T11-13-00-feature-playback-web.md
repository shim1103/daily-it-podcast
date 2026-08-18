---
name: playback web技術選定をTS+Pico.cssへ固定
date: 2026-08-18T11:13:00
session_id: none
branch: feature/playback-web
prev: 2026-08-17T19-10-00-feature-playback-web.md
---

## 1. Summary

philosophy（KISS / YAGNI / Least Power）に基づき、playback web を **Vite + TypeScript + Pico.css（classless）** に固定した。React / Next.js / shadcn/ui / Tailwind は採用しない。worker は TS、generator は Go stdlib のまま。判断の詳細は `docs/decisions/2026-08-18T11-12-00-feature-playback-web.md` に記録し、`README.md` / `DESIGN.md` / playback todo を DRY に整合させた。

## 2. Changes

1. decision `2026-08-18T11-12-00-feature-playback-web.md` を追加（TS + Pico.css、React/Next/shadcn/worker Go Wasm を Rejected）
2. `README.md` の技術選定表と全体図から React / Tailwind を除去
3. `DESIGN.md` に web スタックと decision 参照を追記
4. `playback-web-api-client.md` / `playback-lane.md` の React 前提を TS-only に修正

### Commits

- `c8cc991` — docs(decisions): playback webはTS+Pico.css、React/Next/shadcnを使わない
- `41739b1` — docs(playback): README/DESIGNの技術選定をTS+Pico.cssへ更新
- `e13e337` — docs(tasks): playback todoのReact前提をTS-onlyへ修正
