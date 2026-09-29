---
name: requestId の発行を自作 request-context.ts から hono/middleware/request-id へ委譲する
date: 2026-09-23T03:44:57
branch: refactor/playback-worker-hono-design
---

## 1. Decision

`worker/src/routes/request-context.ts` の `requestLoggingMiddleware` が自前で担っている requestId
の**発行**責務を、Hono 公式の `hono/middleware/request-id`（`requestId()`）へ委譲する。

1. **requestId の発行・伝播は `hono/middleware/request-id` に任せる。** `crypto.randomUUID()` を
   直接呼ぶ自作の `createRequestId()` は撤去し、`.use(requestId())` を
   `.use(requestLoggingMiddleware)` より前段に積む。公式 middleware が `c.set("requestId", ...)` と
   `X-Request-Id` header 付与まで行う。
2. **access log（開始・完了）と branded type（`RequestId`）は `request-context.ts` /
   `worker/src/routes/logger.ts` に残す。** 公式 middleware はログ出力も型安全性も持たないため、
   この2つは自作を維持する。`requestLoggingMiddleware` は `c.get("requestId")` を読むだけの
   middleware へ縮小する。

`worker/src/routes/logger.ts` の `RequestId` branded type・`logInfo`/`logError` の
requestId 必須化は変更しない（`[[skills-1:terms-architecture-logging-policy]]` の
正常系ログ原則を実装する層として維持する）。

## 2. Reason

`node_modules/hono/dist/middleware/request-id/request-id.js` を実際に読んだところ、
既存の自作実装と機能がほぼ重複していた。

```js
var requestId = ({
  limitLength = 255,
  headerName = "X-Request-Id",
  generator = () => crypto.randomUUID()
} = {}) => {
  return async function requestId2(c, next) {
    let reqId = headerName ? c.req.header(headerName) : void 0;
    if (!reqId || reqId.length > limitLength || /[^\w\-=]/.test(reqId)) {
      reqId = generator(c);
    }
    c.set("requestId", reqId);
    if (headerName) {
      c.header(headerName, reqId);
    }
    await next();
  };
};
```

さらに自作版には無い機能を持つ——**client が送ってきた `X-Request-Id` が妥当な形式
（`\w\-=` のみ、255文字以内）なら、新規発行せずそのまま再利用する**。分散 trace で
複数 service を跨ぐ 1 つの request に同じ correlation ID を持たせる、という実務上
標準的なパターンに公式側は最初から対応しており、自作版はこれを持たない。

車輪の再発明を避け、Hono が標準で解決している問題（発行・client 値の尊重・形式検証）を
自前実装で持ち続ける理由はない。一方、Hono 標準の `requestId` middleware は
access log（`request_start`/`request_end` の構造化 log 出力）も、requestId の型安全性
（`RequestId` branded type によるただの `string` との区別）も持たない。この2つは
`[[skills-1:terms-architecture-logging-policy]]` §5（外部入力を受け取る点の正常系log）が
定める、このrepo固有の観測要件であり、汎用 middleware の責務ではない。よって
「発行は委譲、logと型安全性は自作維持」という分割が妥当。

## 3. Rejected

1. **`request-context.ts` を丸ごと維持し、公式 middleware を使わない案** — 車輪の再発明を
   放置する。client 由来の `X-Request-Id` を尊重する機能（分散 trace 対応）を自前で
   実装し直すコストに見合う理由がない。
2. **`hono/middleware/request-id` をそのまま使い、access log と branded type を諦める案** —
   `logging-policy.md` が定める正常系アクセスログの原則を満たせなくなる。`RequestId` が
   ただの `string` に戻ると、branded type によって防いでいた「任意の文字列を requestId として
   誤って渡す」型エラー検出が失われる（`worker/src/routes/logger.ts` の `InfoPayload`/
   `ErrorPayload` が requestId を必須にしている前提と矛盾する）。
