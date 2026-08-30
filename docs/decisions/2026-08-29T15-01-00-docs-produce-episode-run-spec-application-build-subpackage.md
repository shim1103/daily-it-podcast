---
name: ProduceEpisode Builder 詳細は application/build サブ package に置く
date: 2026-08-29T15:01:00
branch: docs/produce-episode-run-spec
---

## 1. Decision

1. **Gate UseCase**（`FetchSourceItems` / `WriteEpisode`）と **Port IF**（`application/port`）は **`application` 直下**に置く。
2. **Builder 詳細**（TextWriter 出力 parse stub、WAV 尺・結合 stub）は **`application/build` サブ package** に置く。`ProduceEpisode` 本体は `application` 直下の orchestrator のまま。
3. `application/build` から export するのは **`ProduceEpisode.Run` が呼ぶ stub / helper のみ**とする。Gate や Port を `build` に寄せない。

## 2. Reason

1. `application/` 直下に Gate・Builder・helper・port がフラットに混在すると、変更理由の単位が読み取りにくい。Builder 詳細を package 境界で束ねると、C 実装の置き場が固定される（SRP）。
2. Go では file 分割だけでは import 境界を作れない。`build` サブ package にすると Composition 以外から Builder 詳細が import されにくい。
3. `application/internal/build` までは今は不要。`application` 配下の 1 段サブ package で足りる（YAGNI）。

## 3. Rejected

1. 現状どおり helper を `application` 直下の unexported func に残す案 — Gate と Builder 詳細が同一 package フラットのまま。
2. Gate も `build` へ移す案 — Fetch / Write の独立ゲート単位が崩れる（既 Decision 違反）。
3. `application/internal/build` 二重 internal 案 — path が冗長。import 制限の追加メリットが今無い。
