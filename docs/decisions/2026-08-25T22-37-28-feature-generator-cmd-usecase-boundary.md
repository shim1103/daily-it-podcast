---
name: ProduceEpisode は生成方針の所有者であり空洞の順送り UseCase にしない
date: 2026-08-25T22:37:28
branch: feature/generator-cmd-usecase-boundary
---

## 1. Decision

1. `ProduceEpisode` は全日次 episode 生成の**方針**（brief、定型と Draft の結合、TTS へ渡す順、尺の書き込み、完成原稿の構築）を持つ。
2. 原稿用・音声用の UseCase を対称のために新設し、`ProduceEpisode` を順序呼び出しだけにしない。独立ゲートとして残すのは既存の `FetchSourceItems` と `WriteEpisode` とする。
3. 原則の正は architecture `backend/application.md` の UseCase 責務・UseCase 同士の依存。

## 2. Reason

1. Clean Arch の UseCase はアプリ固有ビジネス手順である。「今日の episode を作る」方針そのものが UseCase の中身であり、順序だけ残して方針を下へ逃がすと、方針の所在が読取れなくなる。
2. 取得窓と永続前検査はすでに単独の変わり方と test がある。台本組み立てと TTS 順は同じ「生成方針」の一面であり、今は別 UseCase に割ると空洞 orchestrator と方針の分散が同時に起きる（過剰分割）。
3. 後から原稿方針だけが独立に頻繁に変わる実測が出たら、そのとき切り出す（YAGNI）。対称性は切り出し理由にしない。

## 3. Rejected

1. `ProduceEpisode` を Fetch → 原稿UC → 音声UC → Write の順送り専用にする案 — 方針が散り、上位が空洞になる。
2. `FetchSourceItems` / `WriteEpisode` まで `ProduceEpisode` に内包してゲート UseCase を消す案 — 既存の窓計算・検証迂回防止の単位を壊す。
