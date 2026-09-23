---
name: Playback WorkerはController結線をrequest-scoped middlewareに1回寄せ、routeはcontextから取る
date: 2026-09-23T07:44:31
branch: docs/playback-audio-history
---

## 1. Decision

1. Playback Worker の Hono app は、**1 request につき Composition Root（`createPlaybackControllers`）を 1 回だけ**呼ぶ。呼び口は専用 middleware（request-scoped wiring）に置き、各 route handler は `c.get("controllers")` から Controller を取る。
2. Composition Root の責務（具象選択・結線）は変えず、**HTTP pipeline 上の呼び口だけ**を一本化する。
3. requestId / access log / secureHeaders など **cross-cutting** の middleware と、本 Decision の **wiring middleware** は役割を混ぜない（観測・防御 vs 依存組み立て）。

## 2. Reason

1. endpoint（list / audio / progress Write・pull）が増えると、route ごとに同じ `createPlaybackControllers(env, { mode: "r2" }, overrides)` が並び、結線の変更理由が N 箇所に散る。pipeline 先頭で 1 回載せれば、route は HTTP 入出力だけに戻る。
2. Composition Root を消して route 内で Adapter を new する案は、層の唯一点（CR）を壊す。middleware は CR を **呼ぶ位置**を寄せるだけで、DIP / Composition Root と両立する。
3. 学習目的としても、middleware を「Hono API」ではなく **request pipeline の段**（cross-cutting と request-scoped DI の違い）として固定できる。

## 3. Rejected

1. **各 route が個別に Composition Root を呼ぶ案の維持** — endpoint 増加で重複と散在が増える。
2. **Composition Root を廃し route / middleware 内で Adapter を直接 new する案** — 結線特権が複数点に割れ、Dependency Rule が崩れる。
3. **cross-cutting middleware に結線を同居させる案** — 観測・header と DI 組み立てが同じ変更理由になり、段の意味が曖昧になる。
