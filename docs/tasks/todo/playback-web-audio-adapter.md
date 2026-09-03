# playback web audio を lib/ Adapter 層へ分離

## 1. Summary

`<audio>` の命令的操作（`play` / `pause` / `currentTime` reset / lifecycle event 購読）を `lib/audio-element.ts`（Driven Adapter）へ寄せ、`useEpisodePlayback` を「Adapter を購読して `playbackPhase` を同期する薄い hook」へ縮める。hash 同期の 3 層構造（`lib/location-hash` → `lib/hash-selection-adapter` → `use-hash-selection-sync`）と対称にする。

## 2. Context

現状 `use-episode-playback.ts` は `audio.addEventListener` 4 種、`listenedRef`（`{ audio, handlers: Map }`）、`attachListeners` / `detachListeners` を ViewModel の中に生で持つ。hash 側は `addEventListener("hashchange", ...)` が `lib/location-hash.ts` に隠蔽され `use~` から `window` が見えないのに対し、playback は Browser API が ViewModel に露出していて読みにくい（listener 保持・解除の 2 ref、`Map`、attach/detach の対）。

`external-dependencies.md` §3-4 は「要素の markup は Component、命令的操作（`element.play()` / `addEventListener` での lifecycle event 購読）は本層の wrapper 関数へ」と定める。`use-episode-playback.ts` の audio 操作はこの命令的操作に当たる。

## 3. Canonical Sources

- 命令的 API の配置 — `architecture/frontend/external-dependencies.md` §2・§3-4
- ViewModel は同期ロジックのみ — `architecture/frontend/view-model.md` §2-1・§7
- 対称にする参照実装 — `apps/playback/web/src/lib/location-hash.ts`、`apps/playback/web/src/lib/hash-selection-adapter.ts`、`apps/playback/web/src/view-models/use-hash-selection-sync.ts`
- hash 同期の React 方針（本 Issue が延長する判断） — `docs/decisions/2026-08-27T19-20-30-feature-playback-web-page-jsx-mount.md`
- selection と playback の直交 — `docs/decisions/2026-09-02T15-00-00-feature-playback-list-episodes-audio-ref-playback-web-selection-playback-orthogonality.md`
- 現状の hook 契約 — `apps/playback/web/src/view-models/use-episode-playback.ts`、`apps/playback/web/src/view-models/playback-state.ts`

## 4. Scope

### In Scope

- `apps/playback/web/src/lib/audio-element.ts` 新設（Driven Adapter）
  - lifecycle event（`playing` / `pause` / `ended` / `error`）を `PlaybackPhase` へ写して購読する関数（例 `subscribeAudioPhase(el, onPhaseChange): () => void`）
  - 別 episode 切替時の reset（`pause()` + `currentTime = 0` + `load()`）をまとめる関数（例 `resetAudio(el): void`）
  - 再生開始（`el.play()` の Promise 化と rejection の受け口）をまとめる関数
- `apps/playback/web/src/lib/audio-element.sociable_unit.test.ts` 新設（実オブジェクト property を Spy/Stub 差し替え、`external-dependencies.md` §6）
- `use-episode-playback.ts` を Adapter 購読へ書き換え。`addEventListener` / `Map` / `attachListeners` / `detachListeners` / 2 個目の ref を hook から除去
- 対応 `use-episode-playback.sociable_unit.test.ts` の更新（Fake audio Adapter を注入する形へ。挙動観点は現状維持）
- A artifact の更新: `useEpisodePlayback` が Adapter を DI で受け取る形へ signature 変更が必要なら、**着手前に** `use-episode-playback.ts` の stub と型を先に固定する（§5 参照）

### Out of Scope

- `deriveBlockingError` の再設計（early return / invariant throw / `null` 2 義解消）は別 topic。本 Issue で `playback-state.ts` の derive ロジックに触れない
- hash 同期側（`use-hash-selection-sync.ts` / `lib/hash-selection-adapter.ts`）の変更
- `<audio>` の markup 位置・page wire（`playback-web-ui-rewrite`）
- seek UI

## 5. Contract

- A/B は本 Issue 着手前に整合させる（scope-split §2、C は A/B 確定後）。具体的には:
  1. `useEpisodePlayback` の新 signature（Adapter を optional DI で受けるか、`lib/audio-element.ts` の関数を module 参照で使い test は module mock で差し替えるか）を決め、`use-episode-playback.ts` の stub と `EpisodePlaybackViewModel` 型を先に固定する
  2. `lib/audio-element.ts` の export 関数 signature を stub で固定する
  3. この方向づけは新 Decision を立てず、`2026-08-27T19-20-30`（hash 同期を lib + 専用 hook + `useSyncExternalStore` に寄せる判断）の適用範囲を audio へ広げたものとして扱う。Issue の本節を Canonical Source とし、hash Decision を参照する
- `useEpisodePlayback` の公開 ViewModel（`playedEpisodeId` / `playbackPhase` / `audioElementRef` / `play` / `stop`）の外形は変えない。変えるのは内部実装と、あれば DI 引数のみ
- B Decision の直交を維持する（`play` / `stop` は `selectedEpisodeId` を触らない、別 episode `play` で前 audio を reset）
- `use~` から `addEventListener` / `removeEventListener` / `window` / raw DOM event を消す。listener の登録・解除は `lib/audio-element.ts` に閉じる

## 6. Constraints

- `throw` しない。失敗は state / 戻り値（`view-model.md` §4、`external-dependencies.md` §4）
- `lib/audio-element.ts` は ViewModel・Component・純粋関数層を import しない（`external-dependencies.md` §5）
- `<audio>` の markup そのものは本層に持たない（§3-4）
- test のためだけの export を公開 API に足さない（§6-3）。DI か module mock かは §5 で決める

## 7. Acceptance Criteria

- [ ] AC-1: `cd apps/playback && npm run test:unit` が pass する（coverage gate 含む）
- [ ] AC-2: `npm run lint:layers` が pass する（`lib/audio-element.ts` の依存方向違反なし）
- [ ] AC-3: `rg 'addEventListener|removeEventListener' apps/playback/web/src/view-models/use-episode-playback.ts` が 0 件
- [ ] AC-4: 別 episode へ `play` で前 audio が `pause` + `currentTime=0` reset される（`lib/audio-element.ts` の test で検証）
- [ ] AC-5: audio の `error` event で `playbackPhase` が `error` になり、`play()` rejection でも `error` になる（test で検証）
- [ ] AC-6: `useEpisodePlayback` の SU test の挙動観点（初期 idle、play で loading、event → phase、unmount で listener 解除、`play`/`stop` の参照不変）が現状と同数以上で pass する

## 8. Verification

```bash
cd apps/playback && npm run test:unit && npm run lint:layers && npm run typecheck
rg 'addEventListener|removeEventListener|window\.' apps/playback/web/src/view-models/use-episode-playback.ts
```

## 9. Dependencies

- `playback-web-view-models`（ViewModel 本実装）完了が前提。本 Issue はそのリファクタ
- `playback-web-ui-rewrite` とは独立（page wire に触れない）。順序はどちらが先でもよいが、両方 in-flight なら `use-episode-playback.ts` の衝突に注意

## 10. Risks

- SU test を Fake Adapter 注入へ移すとき、現状の 14 case が担保する観点（特に unmount cleanup、`play` rejection、別 episode 切替の reset）を取りこぼす risk — case を 1:1 で移植し、`lib/audio-element.ts` 側と `use-episode-playback.ts` 側のどちらが各観点を所有するか test-minimization で明示する
- `useSyncExternalStore` を audio phase の購読に使う場合、`getSnapshot` が新規オブジェクトを返すと無限 render になる（hash 側と同じ罠）。phase は string enum なので Object.is 安定に保てる

## 11. Notes

- hash 同期との対称性がゴール。完了後は `lib/{location-hash, hash-selection-adapter, audio-element}` が Browser API を閉じ込め、`view-models/use-{hash-selection-sync, episode-playback}` が同期ロジックだけを持つ形になる
- `external-dependencies.md` §3-4 は本 Issue の設計意図が伝わるよう追記済み（要素 markup は Component、命令的操作は本層）
