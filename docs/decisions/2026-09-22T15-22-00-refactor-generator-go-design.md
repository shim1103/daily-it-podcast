---
name: JS/TS tooling の package.json は apps/playback に閉じ、repo root に置かない
date: 2026-09-22T15-22-00
branch: refactor/generator-go-design
---

## 1. Decision

1. Node / Biome / Vitest 等の npm 依存と lockfile の正本は `apps/playback/package.json`（およびその `package-lock.json`）とする。
2. repo root に `package.json` / `package-lock.json` を置かない。Biome を root の `devDependencies` だけのために二重管理しない。

## 2. Reason

1. この monorepo の JS/TS runtime・静的検査の実行入口は playback 側である（`scripts/playback/*` が `cd apps/playback` する）。root に Biome だけを置くと「monorepo 全体 lint」に見え、Go 系統と混線する。
2. playback は既に同版の `@biomejs/biome` を持っていた。root の複製は版ドリフトと「どちらが正本か」のコストだけを増やす。
3. generator は `go.mod` が依存の正本であり、root npm に載せない。

## 3. Rejected

1. **root に Biome 専用 `package.json` を残す案** — 実行は playback のみなのに入口が二つになる。
2. **npm workspaces で root を hub にする案** — 現状の script / CI は playback 単体 `npm ci` で足りており、workspace 導入は過剰。
