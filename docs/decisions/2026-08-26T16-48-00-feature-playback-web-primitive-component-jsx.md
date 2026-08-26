---
name: Primitive を先に JSX 化する時、未 JSX の Feature 向けには createRoot 橋ではなく静的 markup 橋を置く
date: 2026-08-26T16:48:00
branch: feature/playback-web-primitive-component-jsx
---

## 1. Decision

1. Primitive Component を JSX 関数コンポーネントへ先に移行し、Feature がまだ DOM API で組み立てる期間は、Feature 側に一時的な橋渡し（HTMLElement を返す helper）を置いて Verification を緑に保つ
2. 橋渡しは `createRoot` + 子 DOM の抜き出しではなく、`renderToStaticMarkup` で root を持たない経路にする
3. 動的 camelCase の dataset key は `ref` 副作用ではなく、browser dataset と同値の `data-*` 属性へ写像して declarative に渡す
4. 橋渡し helper は Feature 本格 JSX 化 Issue で削除する（恒久 abstraction にしない）

## 2. Reason

1. AC で旧 `.ts` 工場関数を消すと、Feature の import が壊れ typecheck / unit が落ちる。Feature 本格 JSX 化は別 Issue のまま、機械的追従だけを許可しないと Verification と Out of Scope が両立しない
2. `createRoot` で render した子を `append` 先へ移し `unmount` しないと orphan root が残る。`unmount` すると React 管理下の子が壊れる。静的 markup なら寿命問題が消える
3. 静的 markup では `ref` が走らない。dataset を declarative `data-*` にしないと橋と Primitive の観測が乖離する
4. 橋を Primitive 層へ置くと「HTMLElement を返す」責務が JSX Primitive と同居し SRP が壊れる。Feature 側の一時 helper に留め、次 Issue で消す前提にする

## 3. Rejected

1. Feature を同時に全面 JSX 化する — 別 Issue の契約を侵食する
2. `createRoot` + `flushSync` + 子抜き出しを橋にする — orphan root / unmount との両立が取れない
3. `ref` で `dataset[key]` を書く — static markup 経路と非互換で、表示専用 Primitive に commit 後 mutation を残す
4. 旧 `createLabeledText` を HTMLElement 返却のまま残す dual API — AC（`.ts` 廃止・JSX 関数コンポーネント化）と矛盾する
