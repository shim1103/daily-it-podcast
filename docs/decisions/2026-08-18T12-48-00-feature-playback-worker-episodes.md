---
name: Domain 不在は class で分類し message は診断専用
date: 2026-08-18T12:48:00
branch: feature/playback-worker-episodes
---

## 1. Decision

1. JSON 欠落・schema 不適合・wav 欠落・stem 不一致は、いずれも `EpisodeNotFoundError` 1 class に畳む
2. `Error.message` は server 境界が log する診断文であり、HTTP / UI の表示文ではない
3. client 向け文（例: エピソードが見つからない）と External `{ code }` 写像は HTTP 層が所有する

## 2. Reason

1. 不完全ペア専用の外部分類を出すと Drive 内部事情が漏れる（既存 HTTP 契約と同じ畳み）
2. throw は制御フロー脱出であり表示手段ではない。表示変換は境界だけが行う
3. Unit が診断文の文字列を写して固定すると、copy 変更のたびに test が壊れ、class 写像という公開契約を測れない

## 3. Rejected

1. 原因ごとに Domain Error class を増やす案（HTTP が全部 404 に写すなら消費されない）
2. Domain `message` に UI 文を入れて client へ転写する案（内外契約の混同）
3. Unit で `message` 文字列一致と「文が互いに異なること」を仕様として固定する案
