---
name: ViewModel を hook 化した後、未 JSX の page 向けには createRoot 一時橋を page 側に置く
date: 2026-08-26T19:29:00
branch: feature/playback-web-view-model-react-hooks
---

## 1. Decision

1. `createEpisodeListViewModel`（subscribe / getState 工場）を廃止し、`useEpisodeListViewModel` だけを ViewModel の正とする（dual API にしない）
2. Page がまだ DOM 組み立ての間は、page 側に一時橋（`createRoot` + hook Host）を置き、旧工場と同型の `getState` / `subscribe` / `load` / `select` 面を提供して Verification を緑に保つ
3. hook 本体に `flushSync` を置かない。同期観測が必要なのは一時橋の都合であり、橋側に閉じる
4. 一時橋は `playback-web-page-jsx-mount` で page が hook を直接使う時に削除する（恒久 abstraction にしない）

## 2. Reason

1. ViewModel の platform 実装は React hooks（`frontend/view-model.md`）。工場を残す dual API は AC の「書き換える」と矛盾し、次の Feature/Page JSX でもどちらが正か揺れる（Primitive JSX Decision の Rejected と同型）
2. Page / Feature 本格 JSX は別 Issue。旧工場を消すと page の typecheck / unit が落ちる。呼び出し側への機械的追従（寿命付き橋）だけを許すと Out of Scope と Verification が両立する
3. Primitive の静的 markup 橋は表示専用で足りた。ViewModel は state・async・subscribe が必要で、`renderToStaticMarkup` では再現できない。stateful な橋は `createRoot` が必要になる
4. `flushSync` を hook に入れると、page JSX 化後も同期強制が ViewModel に残り続ける。race 判定は `useRef` の同期更新で足り、橋だけが await 後の観測を汲めばよい

## 3. Rejected

1. 旧工場と hook の dual API を残す — 書き換え AC と矛盾し、正本が二つになる
2. Page / Feature を同時に全面 JSX 化する — 別 Issue の契約を侵食する
3. Primitive と同型の静的 markup 橋で ViewModel を載せる — state / subscription を持てない
4. hook 本体内で常時 `flushSync` する — 一時橋の都合を恒久の ViewModel 契約へ漏らす
