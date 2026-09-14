---
name: 情報源 Adapter の共有は httpget helper、NI は源ごと実境界、接続 cache は別 suite（CI 外）
date: 2026-09-14T13:06:11
branch: feature/generator-textwriter-prompt-source-criteria-adapters
---

## 1. Decision

1. **GET + 最小 retry（および HTML 正規化）の共有置き場は `infrastructure/httpget/` とする。** 媒体 Adapter（`hackernews` / `lobsters` / `publickey` / `techcrunch` / `cloudwatch`）は維持し、写像・URL・`SourceID`・方言 parse は各 dir に残す。
2. **`infrastructure/rss/` は作らない。** 各 source を `rss/` 配下へ移さない。RSS 汎用 facade 禁止（先行 `2026-09-02T14-41-00` / `14-41-01` / `2026-09-13T15-08-55`）を維持する。
3. **Narrow Integration は源ごと 1 file（5 源）とし、実境界を対象にする。** `httptest` + DialTLS redirect による合成 upstream NI は廃し、その役割は実境界 NI が吸収する。`httpget` に NI は置かない（helper の SU のみ）。
4. **接続確認 + `.cache/` 保存は NI とは別 suite とする。** `apps/generator/test/` に 1 file・5 case（源ごと正常系のみ）。secret / `workflow_dispatch` は不要。production Adapter は cache を知らない。本 suite は Integration gate（CI）から除外する。

契約値（定数・path・tag 名）の正本は code / script。本文へ写さない。

## 2. Reason

5 源すべてに同型の GET+retry / HTML 正規化が重複している。共有物は HTTP 取得であり RSS parse ではない（HackerNews / Lobsters は JSON）。dir 名を `rss` にすると JSON Adapter が RSS に依存して見え、Least Astonishment に反する。媒体ごとの Adapter を `rss/` に寄せると `SourceID`・dir・中身の一致が壊れ、Rejected 済みの「多媒体を 1 module が捌く facade」に近づく。

写像・since・retry・件数上限は Sociable Unit が所有する。NI は外部境界の実 I/O 契約を見る。`httptest` 合成 upstream は「実境界」ではなく、実 NI と所有が二重になる。接続 cache は到達確認と fixture 保存が目的で、写像や failure 表を再 assert しない。network 依存のため CI gate に載せると Repeatable な secret なし Integration と失敗理由が混ざる。

## 3. Rejected

1. **`infrastructure/rss/` に helper または各 source を置く案** — 共有物が RSS 非依存なのに名が嘘になる。源をまとめると facade 化に寄る。
2. **`httptest` NI を残し実境界 NI と併存する案** — 同じ「外向き HTTP」を二重所有する。実 NI が吸収する。
3. **`httpget` の Narrow Integration を置く案** — helper は Adapter から呼ばれる部品。境界契約の所有者は各 `ItemSource` NI。
4. **接続 cache を NI file に混ぜる案 / `infra/feedcache` を production 化する案** — Scope が混線し、Adapter が test 用 cache を知ることになる。
5. **接続 cache を Integration gate（CI）に載せる案** — 実 network 依存で gate の Repeatable 性を壊す。
6. **RSS 汎用 Adapter（多媒体 1 `List`）案** — 先行 Decision で Rejected。本 Decision も踏襲する。
