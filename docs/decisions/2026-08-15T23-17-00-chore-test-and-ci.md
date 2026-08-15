---
name: commit は Unit、push/GHA は Integration。runner は Vitest と go test
date: 2026-08-15T23:17:00
branch: chore/test-and-ci
---

## 1. Decision

1. commit gate は Unit のみ、push / GHA gate は Integration のみ（YAGNI。E2E・coverage は入れない）
2. Playback の runner は Vitest（`apps/playback`）、Generator は `go test`
3. 実行手順の正は README、配置と gate 規則の正は DESIGN（docs-split を維持）

## 2. Reason

1. Test Pyramid と YAGNI に合わせ、速い Unit を commit、合成の Integration を push に置く
2. Vite 系 Playback には Vitest、Go は標準 `go test` が Least Power
3. 使い方と規則を同一文に書くと README/DESIGN の DRY が崩れる（既存 docs-split）

## 3. Rejected

1. push CI に Unit を「念のため」併記する案（gate 条件の二重化）
2. `~/.cursor/rules` や AGENTS.md を User Rules の代替にする案（製品が読まない / cloud 保存）
3. Go build tag による Integration 分離を最初から入れる案（path 分離で足りる）
