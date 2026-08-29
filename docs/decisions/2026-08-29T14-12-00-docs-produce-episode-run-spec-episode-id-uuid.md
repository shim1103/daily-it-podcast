---
name: episodeId は ProduceEpisode が発行する opaque な UUID とする
date: 2026-08-29T14:12:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. Drive stem および原稿 JSON の `episodeId` は **`ProduceEpisode` が `WriteEpisode.Run` 直前に 1 回発行**する opaque な **UUID** とする。
2. 表示用 `date`（JST 暦日、`YYYY-MM-DD`）とは **独立**する。`2026-08-15T16-23-07`（id と date 分離）を維持する。
3. UUID 生成 library の選定は C の実装詳細とする。B 固定は「opaque UUID であること」のみ。

## 2. Reason

1. timestamp や暦日を id に埋め込むと、UI や Reader が Generator の時刻組み立て規則を id から推測できる。表示 date は JSON field だけが正であるべき（Orthogonality）。
2. UUID は `contracts/drive-layout.md` の「不透明な対応キー」にそのまま合致する。
3. id 発行を Gate に寄せると、検証 UseCase が identifier 生成方針まで持つ。Builder が完成 manuscript を組む流れと一致しない。

## 3. Rejected

1. `date` や実行時刻から stem を組み立てる案 — id と表示 date の意味が混ざる（`2026-08-15T16-23-07` Rejected と同型）。
2. `WriteEpisode` または `EpisodeWriter` Adapter が id を発行する案 — 書込 Gate が生成方針を持つ。
3. Cursor 出力に `episodeId` を含めさせる案 — 生成単位と永続単位が Port 上で混ざる（`2026-08-25T23-02-35` と同型）。
