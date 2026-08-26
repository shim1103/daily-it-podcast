---
name: Hono の AppType を hc で使う時、route 登録は mutation ではなく method chain にする
date: 2026-08-26T19:28:00
branch: feature/playback-web-view-model-react-hooks
---

## 1. Decision

1. `export type AppType = typeof app` を `hc<AppType>()` の入力にする前提では、`app.get(...)` の mutation 連打ではなく、`new Hono().get(...).get(...)` の **method chain** で route を載せる
2. handler の振る舞い自体はこの判断の対象外。型が route を載せるための登録形だけを固定する

## 2. Reason

1. Hono の型は chain の戻り値に route schema が積まれる。mutation の `app.get` だと `typeof app` が空に近い schema のまま残り、`hc<AppType>()` の呼び出し面が `unknown` になる（本 repo で実機 `tsc` により確認済み）
2. RPC 採用の merit（path/method の型同期）は `AppType` に route が載って初めて成立する。route 定義 Issue 完了後でも、登録形が mutation のままだと web 側 AC（RPC 差し替え）が型 safe に進まない
3. chain 化は handler 中身の変更を要求しない。振る舞い SSOT を動かさず、型だけを RPC 利用可能にする最小変更になる

## 3. Rejected

1. mutation のまま `AppType` を手で組み立てる / 別の型宣言を足す — framework が持つ chain 型を捨て、型同期の二重管理に戻る
2. `unknown` のまま RPC を呼び、呼び出し側で assertion する — 採用理由の型同期を捨て、実行時までずれを遅延させる
3. route 定義を web 側の手書き path に逃がして RPC を避ける — Hono RPC 採用決定と矛盾する
