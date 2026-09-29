---
name: hono/middleware/secure-headers を全route共通で導入する
date: 2026-09-23T03:44:59
branch: refactor/playback-worker-hono-design
---

## 1. Decision

`hono/middleware/secure-headers`（`secureHeaders()`）を、`app.ts` の `.use()` chainへ
`requestLoggingMiddleware` と並べて全route共通で導入する。個別 route ごとの適用分けは行わない。

## 2. Reason

`node_modules/hono/dist/middleware/secure-headers/secure-headers.js` を読むと、この middleware
は `X-Content-Type-Options: nosniff`・`X-Frame-Options: SAMEORIGIN`・
`Referrer-Policy: no-referrer` 等、複数のセキュリティ関連 header をデフォルトで一括付与する。
現状このrepoにはセキュリティ関連 header の付与が一切無い。

このAPIは認証機構を持たない公開API（前回セッションで確認済み）だが、公開APIであることは
セキュリティ header が不要である理由にはならない。特に `X-Content-Type-Options: nosniff` は
音声mp3・一覧JSONいずれのresponseでも、ブラウザによる MIME sniffing（宣言された
Content-Type を無視してブラウザが独自に中身を判定してしまう挙動）を防ぐ効果があり、
どの route にも一律で効かせるべき性質の header である。route ごとに必要な header が
異なるわけではないため、全route共通の `.use()` として導入する。

## 3. Rejected

1. **route ごとに個別設定する案** — このAPIの2 route（一覧・音声）はどちらも同じ配信目的
   （読み取り専用の公開data配信）であり、必要なセキュリティ header の種類に差が無い。
   `etag`（route間でコスト対効果が非対称）とは性質が異なり、分ける理由がない。
2. **導入しない案** — 「認証が無い公開APIだから不要」という判断は、MIME sniffing対策等の
   基礎的な防御まで省く理由にはならない。導入コストは応答ヘッダーの追加のみで、CPU・帯域への
   影響も無視できる規模のため、見送る理由がない。
