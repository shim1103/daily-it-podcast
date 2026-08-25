---
name: 初期の多文書ウォーターフォールと全系モック先行を常設手法にしない
date: 2026-08-25T07:01:00
branch: cursor/docs-early-design-decisions-3502
---

## 1. Decision

1. README / PROPOSAL / SPEC / DESIGN / QUESTIONS の 5 点セットと、権威序列 `README > PROPOSAL > SPEC > DESIGN` を常設しない。文書の正は `2026-08-15T16-23-08`（README・DESIGN・contracts）。
2. 「原稿生成 / podcast 出力 / 再生 UI / 横断オーケストレーション」の 3 系統 + DI オーケストレーションを常設しない。ランタイムの正は `2026-08-15T16-23-06` の 2 系統。
3. 「全システムを固定レスポンスのモックで先に E2E 接続する」ことを、実装前の必須ゲートにしない。実境界（Drive・HTTP・secret・cron）に届かないモック結合を、受け入れの主証拠にしない。

## 2. Reason

1. code 前に文書階層・系統数・MVP 定義まで固定した結果、後から文書体系・系統数・UI 基盤・原稿経路がすべてひっくり返った（archive rewrite）。
2. 要求の交差を解く作業と、未検証の構造を文書で固定する作業は別である。後者を常設すると手戻りが前倒しされるだけで減らない。

## 3. Rejected

1. PROPOSAL / SPEC / QUESTIONS を再び常設する案
2. 工程ごとの多系統 + 中央オーケストレーションを再び標準形にする案
3. 実 API・実 Drive の前に全系モック E2E を必須にする案
