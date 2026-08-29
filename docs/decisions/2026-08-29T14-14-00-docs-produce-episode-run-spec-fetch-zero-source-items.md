---
name: Fetch 後に SourceItem が 0 件なら Domain Error で終了し WriteEpisode を呼ばない
date: 2026-08-29T14:14:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. `FetchSourceItems.Run` が **空 slice（0 件）**を返した場合、`ProduceEpisode.Run` は **Domain Error**（`no_source_items`）で終了する。
2. この場合 **TextWriter / TTS / WriteEpisode を呼ばない**（書込なし）。正常終了（exit 0）にしない。
3. `ItemSource.List` が 0 件を返すこと自体は Port 契約上 valid である。0 件を episode 未生成として reject するのは **Builder 方針**である。

## 2. Reason

1. 情報源が無い状態で Cursor / TTS を走らせると、コストだけ消費し成果物の意味が無い。fail-fast が Least Astonishment に合う。
2. Infra 障害（List error）と「窓内に該当なし」を Domain で区別すると、運用側が再実行判断しやすい。
3. 空 episode を書く案は、Playback 一覧に意味の無い件が載る。個人 podcast の日次生成として過剰（YAGNI）。

## 3. Rejected

1. 0 件でも exit 0 で終了する案 — 成功と区別できず、GHA が silently skip したように見える。
2. 0 件を Infra Error にする案 — Adapter は正常に空を返している。層分類が崩れる。
3. ダミー SourceItem を合成して続行する案 — 事実と異なる episode を生成する。
