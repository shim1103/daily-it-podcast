---
name: Playback remote E2E は本番 Drive の安定 fixture 1 本以上の正常系に寄せる。専用 test folder は作らない
date: 2026-08-31T00:12:01
branch: feature/playback-e2e-deploy
---

## 1. Decision

1. Playback browser E2E（認証後）は **正常系中心**とし、**episode が 1 件以上ある path** に寄せる。空一覧を E2E の主契約にしない（0 件は下位 Scope が既に所有）。
2. その 1 件以上は、本番 Worker が読む **`DRIVE_FOLDER_ID` folder 直下**へ置く **安定 fixture**（`{episodeId}.json` + `{episodeId}.wav`）で賄う。人手で置く。Generator System の `TEST_*` folder とは分ける（本番読取 path を変えない）。
3. fixture の原稿は **`WriteEpisode` の検証**（`manuscript.schema.json` + stem 一致 + 非空音声）を通るものとする。値・file の正本は repo 内 fixture artifact とし、本 Decision に UUID / 日付 / 本文を写さない。
4. 日次 produce で episode が増えても、安定 fixture は残す。本番一覧に fixture も見えることを受け入れる（個人 Access 前提）。
5. Playback E2E 専用の第3 Drive folder / 別 Worker env は作らない。Drive を E2E 側で double しない先行方針（`2026-08-30T16-20-03`）を維持する。

## 2. Reason

1. E2E の価値は Access 入場後の一覧・原稿・再生の最終結果である。空一覧の振る舞いだけを上層で重ねると、Pyramid 上で下位所有と重複し、原稿・再生の観測ができない。
2. remote E2E は本番 Worker 経路が正である。別 folder を刺すと「本番 path の回帰」ではなくなり、運用面（第2 env）だけが増える。
3. System 用 `TEST_*` は Generator が本番を汚さないための分離であり、Playback 読取側の二重化ではない。
4. fixture を人手配置にするのは、generator 入口を E2E 準備に混ぜず、書込検証境界（`WriteEpisode`）だけを満たせば足りるからである。

## 3. Rejected

1. E2E 主契約を空 Drive（0 件一覧）にする案 — 原稿・再生 UI を観測できず、正常系の最終結果検証にならない。
2. Playback E2E 専用 Drive folder / 第2 Worker を立てる案 — 本番経路から離れ、secret / Variable 面が増える。
3. production path に in-memory / test bypass を足して fixture 不要にする案 — 本番経路を短絡し、先行 Constraints に反する。
4. Generator System を本番 `DRIVE_FOLDER_ID` に向けて fixture を兼ねる案 — System 失敗・テストデータが本番一覧を不安定にする。
