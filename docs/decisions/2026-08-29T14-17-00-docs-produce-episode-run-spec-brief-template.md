---
name: TextWriter への brief は平文 1 本で podcast 設定・SourceItems・出力形式を含める
date: 2026-08-29T14:17:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. `ProduceEpisode` が組み立てる **brief** は **平文 1 本**とし、`port.TextWriter.Write` へ渡す。JSON envelope ではない。
2. brief の構成は **（1）最小の podcast 固定設定（2）SourceItems ブロック（3）出力形式ブロック** とする。見出し定数の正本は `entities/constants/brief_sections.go`。
3. SourceItems は JSON ではなく **階層が読める平文**。`SourceID` と `OccurredAt` 以外は opaque `Context` をそのまま載せる。Application は `Context` を parse しない。
4. 出力形式ブロックは **Cursor が Intro / Topics / ClosingSummary / Title だけ**生成することを指示する。**OpeningGreeting / ClosingFarewell は brief に含めない**。
5. Topic タイトルは簡潔な題名、episode `Title` は最重要 topic から気を引く見出し、という **意味上の方針**を brief に書く。数値 range は定数を brief に埋め込む（正本は constants）。

## 2. Reason

1. TextWriter Port は `brief string` のみ。vendor argv / stdin 形式は Cursor Adapter が知る。brief 内容方針は Builder が知るべきで、Port 署名に載せない（DIP）。
2. Cursor と完成 schema は一致しない（`2026-08-18T16-30-00`）。brief で出力形式を固定し、Builder parse へ渡すのが最短経路。
3. SourceItems を structured JSON にすると、情報源ごとに field が増え Port 再設計に向かう。opaque Context は既存 ItemSource Decision と一致する。

## 3. Rejected

1. brief 組み立てを Cursor Adapter 内に置く案 — Infra が episode 生成方針を知る。
2. brief を Composition Root が書く案 — 結線点に手順が混ざる（`2026-08-25T22-37-30` Rejected）。
3. podcast 固定設定を Decision 本文に長文で再掲する案 — 文言は D（未決）でもよい。構造と見出し定数だけ A/B で固定する。
