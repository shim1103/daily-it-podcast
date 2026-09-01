---
name: README 地図から rewrite archive 参照を残さない
date: 2026-09-01T12:58:01
branch: cursor-docs-ssot-develop-cleanup-9567
---

## 1. Decision

README の地図・branch 表は **現行 2 系統（Generator / Playback）と現行 branch 役割だけ**を示す。`archive/2026-08-15-pre-rewrite` 等の rewrite 履歴ナレーションは README に書かない。integration 正本 branch は `develop` とし、表記は **SSOT** とする（`master` は release）。

## 2. Reason

1. README の役割は地図・使い方・受け入れ（`2026-08-15T16-23-08`）。rewrite 以前の monorepo 説明は現行読者の入口としてノイズになり、どの DESIGN が正か迷わせる（根 `DESIGN.md` と旧 `docs/DESIGN.md` 重複の温床だった）。
2. archive は git tag として残る。地図文書へ履歴を再掲すると、tag と README の二重 SSOT になる。必要なら git / daily を見れば足りる。
3. `develop` を「base」より「SSOT」と呼ぶのは、文書・workflow・`workflow_dispatch` の正本 branch が develop であることを一目で示すため。release 用 `master` と役割を対で読める。

## 3. Rejected

1. README に archive 行を残して「旧実装がある」と注記する案 — 新規読者は現行だけ追えばよい。注記は完了ナレーションに近く、地図が肥大する。
2. rewrite 経緯を README §Branch に長文で書く案 — 変更理由は daily / decision にあり、README は latest のみ。
3. default branch 表記だけ直し archive 行は残す案 — 旧 `docs/DESIGN` への誘導が残り、今回の削除意図と矛盾する。
