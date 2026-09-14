---
name: 情報源 Narrow は controllable peer。本番直撃は接続 cache のみ
date: 2026-09-14T15:05:00
branch: feature/generator-item-source-httpget-ni-cache
---

## 1. Decision

1. **源ごと Narrow Integration は、境界 provider（標準 `*http.Client` / TLS）を実物にし、本番 peer を controllable upstream（`httptest` + DialTLS redirect 等）へ double する。** 成功時の request 観測（GET・認証 header 無し等）と、失敗時の `*adaptererror.Error` への包みを切り分ける。
2. **本番 peer へのインターネット直撃を Narrow と呼ばない。** 到達確認と `.cache/` 保存は接続 cache suite（gate 外）の所有とする。
3. **先行 Decision `2026-09-14T13-06-11` §1.3・§2 の「httptest NI を廃し本番実境界 NI が吸収する」および Rejected §3-2（httptest 併存禁止の根拠としての「実境界=本番」読み）を、本 Decision が置き換える。** `httpget` helper・接続 cache 別 suite・RSS facade 禁止（同 Decision §1.1・§1.2・§1.4）は維持する。
4. **`httpget` に Narrow は置かない**（helper は SU のみ）。境界契約の所有者は各源 `ItemSource` NI。

## 2. Reason

1. `testing-strategy/levels.md` の Narrow「実物」は境界 provider であり、本番 peer ではない。controllable peer が無いと HTTP status 系と Adapter の infra 包みを Fault Isolation できない。
2. 先行 `2026-09-02T16-57-00` も generator Narrow を fake upstream（httptest + DialTLS）と明記している。`2026-09-14T13-06-11` の「実境界=本番 GET」読みはこれおよび levels と衝突した。
3. 接続 cache は到達と fixture 保存が目的で、NI の Repeatable な境界契約検証とは所有が違う。両方を本番直撃 NI に混ぜると gate の失敗理由が混線する。

## 3. Rejected

1. **本番 GET を Narrow に残し cache だけ gate 外にする案** — Narrow の定義と Repeatable 性を壊す。
2. **httptest NI と本番直撃 NI を併存する案** — 同じ外向き HTTP 契約を二重所有する。本番直撃は cache（または System）へ寄せ、NI は controllable peer に一本化する。
3. **接続 cache を NI file に混ぜる案** — Scope と目的が混線する（先行 Rejected を維持）。
