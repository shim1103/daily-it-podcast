---
name: Drive 同名 file の find は同 stem の upsert とする
date: 2026-08-30T23:31:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Generator の Drive 書込 Adapter は、所定 folder 内で **同名 file を find** したとき **media を update** する（無ければ create）。これは **同 stem（`episodeId`）の upsert** であり、原稿 content が同じだと決め打ちしない。
2. find を理由に opaque `episodeId` を再発行しない。find を Domain Error にもしない。
3. content 二重（同日・別 UUID）の扱いは本 Decision の対象外である。正は Decision `2026-08-30T23-30-00`。

## 2. Reason

1. stem は opaque UUID（先行 Decision `2026-08-29T14-12-00`）なので、通常の新規 Run では同名 find は起きにくい。起きうる主因は **同一 `episodeId` の部分書込の続き**（例: json 成功・wav 失敗のあと、同 ID で再 put）である。このとき update が正しい。
2. Adapter が content 同一性を判断すると、配置契約（name / MIME / byte）を超えて原稿意味を知ることになる。
3. Write 途中での ID 再発行は、既に manuscript に埋め込まれた `episodeId` と stem 契約を壊す。

## 3. Rejected

1. find したら Domain Error にする案 — 部分書込のリカバリ経路を自分で潰す。
2. find したら opaque ID を振り直す案 — manuscript の `episodeId` とファイル名がずれる。
3. find を content 同一の証拠とみなす案 — name 一致と content 一致は別である。
