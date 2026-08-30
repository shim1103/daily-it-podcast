---
name: dev-onlyでuse case overridesが必要なHono appはfactory関数化する
date: 2026-08-26T18:00:53
branch: feature/playback-worker-hono-entry-cutover
---

## 1. Decision

1. production用に単一export（`export const app = new Hono(...)`）していたHono instanceは、非productionの呼び出し元がuse case実装を差し替える必要が生じた時点で`createApp(useCaseOverrides?)`のようなfactory関数へ変える。production側の`app`はfactoryの引数無し呼び出し結果とし、route定義・error写像（`notFound`/`onError`）はfactory内の単一箇所にのみ書く。
2. overridesの型は、Hono app側で新規定義せず、Composition Root（`createPlaybackControllers`）が既に持つ`PlaybackUseCaseOverrides`をそのまま再利用する。

## 2. Reason

1. `web/vite.config.ts`のdev-only middlewareは、production routerと同じHTTPハンドリング（route一致・Response組み立て・error変換）を必要としながら、backendはfake use caseへ差し替える必要がある。この2要件を両立する経路が、単一Hono instance exportには無かった。
2. factory化以外の選択（後述Rejected）はいずれもroute定義かerror写像のどちらかを複製するため、DRY（`philosophy` §2-2）に反する。productionとdev-onlyが同じHTTPハンドリング経路を通ることは、Issue Contract（`worker-entry.ts`のexport interface不変）にもPrinciple of Least Astonishment（`philosophy` §4-5、production/dev-onlyで挙動が分岐しない）にも資する。
3. overridesの型を独自定義せず`root.ts`の既存契約を再利用したのは、`PlaybackEnv`（Cloudflare Workers native binding）に`overrides`を混在させる代替案が`runtime-config.ts`の既存契約（`@require env は Worker binding 由来`）に反するため。型の正本を二重に持たないことがLaw of Demeter（`philosophy` §3-2）的にも妥当。

## 3. Rejected

1. dev-only middlewareが`app.ts`のroute定義を独自に複製する案。production route変更のたびにdev-only側の追従漏れが起きる（旧`fetch.ts`時代に実際発生していたDRY違反の再発）。
2. Hono instanceを複数export（production用・dev用）し、それぞれ個別にroute定義する案。route定義・error写像が2箇所に分岐し、Issue Contract上「production HTTP入口の変更」がdev側にも波及することの検知が難しくなる。
3. `PlaybackEnv`型自体に`overrides`フィールドを追加する案。Cloudflare Workers native bindingという既存契約の意味を汚染し、production runtimeが実際には使わないfieldを型に持ち込む。
