---
name: useHashSync の hash 購読を useSyncExternalStore へ移行
date: 2026-08-28T15:52:47
session_id: none
branch: feature/playback-web-use-sync-external-store
prev: なし
---

## 1. Summary

issue-manager で `apps/playback/web/src/view-models/use-hash-sync.ts` の外部ストア購読部を `useSyncExternalStore` ベースへ移行した。手書きの hashchange listener 登録・`lastSyncedHashRef`・`onHashSelectRef`（stale closure 対策）を除去し、`syncedHashRef` 1個へ集約した。decision `2026-08-27T19-20-30` §1-3 の follow-up 消化。executor が TDD で実装、reviewer が code-review + simplify を査読（blocker なし、should-fix 2件・nit 2件）、executor が指摘反映、manager が全 AC を再実行で audit した。1 commit を push、issue file を削除した。

## 2. Changes

1. `useHashSync` の読み取り方向を `const currentHash = useSyncExternalStore(onLocationHashChange, getLocationHash)` へ移行。`subscribe` / `getSnapshot` に `lib/location-hash.ts` の既存 Driven Adapter をラッパー無しで直接渡す。`getLocationHash` が `#` 除去済み文字列を返すため `Object.is` 安定（Constraint 1）。
2. `lastSyncedHashRef` + `onHashSelectRef`（2個）を除去。`syncedHashRef`（初期値 = mount 時 `currentHash`）1個で、読み取り差分（外部由来の変化か echo か）と書き込み差分（selectedId 追従要否）の双方を判定。`onHashSelectRef` は `useSyncExternalStore` の再 render で最新 `onHashSelect` が closure に載るため不要化。
3. 読み取り effect（dep `[currentHash, onHashSelect]`）を書き込み effect の前に宣言。`currentHash !== syncedHashRef.current` の時だけ `onHashSelect` を呼ぶ。書き込み effect（dep `[selectedId]`）は比較対象を `syncedHashRef.current` に置換しただけで残置（decision §1-4、Out of Scope 3）。
4. reviewer 指摘の反映: `@ensure` / why コメントを「hashchange 由来のみ」→「直近同期値との差分検出。初回既存 hash と自己書き込み echo は流さない」へ（should-fix #1）。2 effect の宣言順依存の why コメントを、実際に効く interleaving（`selectedId` と `onHashSelect` が同一 commit で変化し `currentHash` 未追従の時）へ差し替え（should-fix #2）。reviewer 提案 A（2 effect を1本へ統合）は happy-dom で既存ケースが赤化したため不採用、fallback（宣言順維持 + コメント修正）を採った。
5. test 追随: `useSyncExternalStore` は hashchange → 再 render → effect の経路を通るため、hash 変更 + `hashchange` 発火を `act()` で括る `dispatchHashChange` helper を追加、既存4箇所を置換（assertion 無改変）。「外部 hash 不変で再 render しても余分な再 render が起きない」回帰ガードを1ケース追加（`getSnapshot` の Object.is 安定を render 回数で固定）。use-hash-sync test 8→9 ケース。
6. shim との対話で `useSyncExternalStore` の内部機構（Fiber ノード上の `inst.value`、render 中と commit 直前の2回 `getSnapshot` 呼び出し＝tearing 検出、subscribe callback が再 render トリガ、再 render 対象は該当 Fiber サブツリー）、`act()` が React の非同期更新を flush する役割、旧実装（同期 listener 直呼び）との差を議論した。
7. manager audit: `npm run typecheck` エラー0 / `npm run lint:layers` 0 violations（53 modules）/ `npm run lint` 86 files clean / `npm run test:unit` を5連続実行し全て 37 files・212 tests pass、flaky ゼロ。`use-hash-sync.ts` カバレッジ 100/100/100/100。`use-hash-sync.sociable_unit.test.ts` 9 tests + `episode-list-page.sociable_unit.test.ts` 8 tests（後者は完全無修正）green を独立確認。
8. `npm run dev` での実ブラウザ hash 同期（戻る/進む、`#ep-xxxx` 直リンク）は未実施。happy-dom は hashchange 非同期発火・history push を再現しないため PR review で manual 確認する。
9. pr-completion flow で PR [#77](https://github.com/shim1103/daily-it-podcast/pull/77)（base: develop、関連 GitHub Issue なし・追跡は local task file）を `gh pr` 経由で作成した。`develop` 先行の PR #76 と `docs/lessons/index.md` が conflict（両者とも 223 の直後に追記）。両側保持で解消し、本 session の lesson を 224〜228 → 241〜245 へ振り直した。merge 後に typecheck / test:unit（37 files・212 tests）/ lint:layers を再実行し green。CI（static-and-unit / integration）全 pass、mergeStateStatus CLEAN を確認。

### Commits

- `315babd`
- `ba46c19`
- `3a34fcd`
