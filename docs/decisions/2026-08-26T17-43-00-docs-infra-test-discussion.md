---
name: Generator の local 実物 suite は build tag で gate 収集から除外する
date: 2026-08-26T17:43:00
branch: docs/infra-test-discussion
---

## 1. Decision

1. Generator の local 実物 Integration / System suite を gate 収集から除外する手段は **Go build tag** とする。tag 名と入口の正本は code 契約を参照する。
2. file 名規約は分類ラベルとして使ってよいが、収集除外の正本にはしない。
3. Integration gate script は当該 tag を付けずに実行する。local 専用入口だけが tag 付きで実行する。

## 2. Reason

1. `go test ./package/...` は package 内の `*_test.go` を file 名パターンでは除外しない。file 名だけを「除外契約」にすると、実際には gate が local 実物を実行してしまう。Least Astonishment に反する。
2. script 側で path を1 file ずつ列挙除外すると、suite 追加のたびに gate script が変わり、収集契約が散在する（DRY / Orthogonality）。build tag なら「tag 付き file は既定 build に入らない」が一箇所の言語仕様で完結する。
3. 別 directory に切る案もあるが、既存の Integration 置き場契約（`apps/generator/test/`）を増やさず、同じ package で Sociability 分類名を file 名に出せる方が、現行 DESIGN の命名規則と衝突しない。

## 3. Rejected

1. file 名（例: `*_local_*`）だけで除外する案 — 収集が変わらないため契約として偽。
2. gate script が個別 path を列挙して除外する案 — 追加のたびに script 更新が必要で、除外漏れが gate 汚染になる。
3. local 実物を別 Go module にする案 — 過剰分割。今必要なのは収集境界だけである（YAGNI）。
