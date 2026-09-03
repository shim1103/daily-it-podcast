---
name: playback web ViewModel の本実装と state 再設計（union 化・error 3層・play/seek 分離）
date: 2026-09-03T17:30:25
session_id: none
branch: feature-playback-web-view-models
prev: なし
---

## 1. Summary

playback web の ViewModel hook（catalog / selection / hash-sync / playback / compose）を A 契約 stub から本実装へ置き換えた。実装を進める中で shim との設計対話から state 表現を段階的に作り直し、3 本の Decision に固定した。判別可能 union で矛盾状態を型排除し、error を blocking（`PageStatus`）と non-blocking（`PlaybackState` の error 枝）へ分離、再生 state を `kind` tag union にして再生位置を state で保持、`play` / `seek` を「指定秒から鳴らす」「指定秒へ移動する」の 2 操作へ分けた。`<audio>` の命令的操作は `lib/audio-element.ts` の Driven Adapter へ閉じ込めた。

## 2. Changes

1. 各段階（stub→本実装、union 化、error 3層、`kind` tag + 位置保持、可読性 refactoring）を独立 Decision + executor + reviewer + audit のサイクルで回した。段階ごとに Decision が前段の一部を supersede し、置き換え範囲は新 Decision 側にのみ記した。
2. `useEpisodeSelection` は catalog の episodes を受け、選択確定時に実在検証して episode 実体を state に持つ。`deriveSelectedEpisode` は削除、`derivePageStatus` の入力は `CatalogStatus` 1 つに縮小、無効な hash 由来の選択は state に入れず reject。
3. `lib/audio-element.ts` に `subscribeAudioState`（phase + timeupdate + loadedmetadata）と `seekAudioElement(el, sec, { play })` を新設、`startAudioPlayback` / `subscribeAudioPhase` を削除。`useEpisodePlayback` は Adapter 通知を state へ同期するだけの hook にした。
4. `deriveEpisodeRows` で row 単位の投影（`isSelected` / `isPlaying`）を作りきり、compose hook から `isX(id)` 系 callback を除去。`useHashSelectionSync` に catalog 完了までの同期保留判断を移設。
5. `use-episode-playback.ts` を可読性重視で refactoring。docstring 33→11 行、`apply*` の重複 3 関数を `updateActive` helper へ、`// why:` コメント 12→4 個、三項ネストを名前付き述語へ。挙動不変（既存 SU test 25 件を無改変で緑）。
6. 検証: `test:unit` 333 passed（55 files）、`lint:layers` / `typecheck` / `lint` / `format:check` 全 pass、view-models と lib の branch coverage 100%。generator 側も pre-commit hook で全 pass。
7. lessons に 23 件追記（union の判別子設計、state 正規化、error 3層、media 操作の分離、ref と state 同型化、命令的 API の層分離、上位モデル相談の前提確認 等）。
8. follow-up として旧 `useEpisodeListViewModel` 撤去・表示側配線・`<audio src>` の ViewModel 管理化を `playback-web-audio-adapter.md` 他へ残した。

### Commits

- `54c38bd`
- `68cd110`
