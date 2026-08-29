---
name: playback 一覧UIの責務を ViewModel / Entry / Row に再分離する
date: 2026-08-30T01-56-00
branch: feature/playback-audio-player-ui-design-api-refactoring
---

## 1. Decision

1. **根本原因**：hidden `<audio>` を item 間で回す host にしたため、`RefObject`・`host/silent`・item 内 `useEffect`・`baseUrl` 貫通が発生した
2. **共有 audio は再生中の `EpisodeListEntry` が宣言的 `<audio>`（`EpisodeAudio`）を 1 つ mount する。** Page / List の「一覧全体の上」には置かない。`new Audio()` / External Dependencies で wrap しない（`feature-component.md` §2-3）
3. **Page** が `useEpisodeListViewModel` と `useEpisodePlayback(baseUrl)` を呼び、`EpisodeList` だけを compose する。`EpisodeAudio` は Page に置かない。再生 UI（pill / seek）も Page に置かない
4. **`useEpisodePlayback`** が再生 state・src・seek・timeupdate 購読を持つ。`baseUrl` による URL 組立は VM 内で閉じる。Page は `audioElementRef` / `resolvedSrc` を List へ渡し、**再生中 entry だけ**が `EpisodeAudio` に配線する。seek 時に audio 未 mount なら pending として覚え、mount 後に適用する
5. **`EpisodeList`** は ViewModel hook を呼ばない。`selection` / `playback` / callbacks / audio 配線を受け、entry 向け union を導出して並べるだけ。自身は `EpisodeAudio` を描画しない
6. **`EpisodeListEntry`** が `playback.kind === "playing"` の時だけ直下に `EpisodeAudio` を置く（その行の上＝その entry 内）
7. **`EpisodePlayer`（3 variant）を削除**し、次へ分解する
   - `EpisodeAudio`：`<audio>` のみ
   - `EpisodePlayButton`：表示 + `onPlay` のみ
   - `EpisodeSeekBar`：表示 + `onSeek` のみ（自前 `useState` / listener を持たない）
8. **`EpisodeSelectedGroup` を削除**し、`EpisodeListEntry` の CSS modifier（`--selected`）に畳む
9. 状態は discriminated union。`null` / boolean 混在禁止

```ts
EpisodePlayback =
  | { kind: "idle" }
  | { kind: "playing"; episodeId: string; positionSec: number; durationSec: number };

EpisodeEntrySelection =
  | { kind: "closed" }
  | { kind: "open"; detail: EpisodeDetailState };

EpisodeEntryPlayback =
  | { kind: "stopped" }
  | { kind: "playing"; positionSec: number; durationSec: number };
```

10. selected / played は直交。item を隠さない
11. `EpisodeFocus` は作らない（CSS）
12. `utils/seek-audio-element.ts` の副作用は VM（または Page が渡す `onSeek`）へ移し、utils からは削除する
13. 先行 Decision `2026-08-30T01-43-00-...` のうち、List が `useEpisodePlayback` を持つ・item が audio host する・`host/silent` で回す条項は **本 Decision が上書きする**

## 2. Reason

1. Feature が ViewModel を所有すると `feature-component.md` §3（状態管理を置かない）に反する。現状 `EpisodeList` が `useEpisodePlayback` を呼んでいる
2. `EpisodeListItem` が row・detail・audio host・auto-play を同時に知り、変更理由が交差する
3. `EpisodePlayer` の 3 variant は共通実装がほぼ無く、dispatcher が負債
4. skill は宣言的 `<audio>` を Feature に置き、External Dependencies で wrap しないと定める。再生中 entry が `EpisodeAudio` を持つと「list-item の上」という shim 意図と、宣言的 audio の両方を満たす
5. Page 直下に置くと「全部の list の上に再生中 audio が1つ」になり、行との対応が読めない。entry 内ならどの episode の音かが DOM 位置で分かる
6. `RefObject` は Page→List→再生中 Entry の配線に限り、Row / Detail / PlayButton の props には載せない

## 3. Rejected

1. item が idle 時まで audio を持ち回る（host/silent）— Ref 貫通と副作用の温床
2. List / Page が一覧全体の上に `EpisodeAudio` を置く — 行との対応が消え、shim の「list-item の上」と矛盾
3. List が `useEpisodePlayback` を所有 — Feature が状態管理する
4. `new Audio()` を External Dependencies に閉じる — `feature-component.md` §2-3 違反
5. React Context で playback 配布 — 深さ3に対して過剰
6. `EpisodePlayer` を variant 削減だけで残す — 共通実装が無く中間層が残る
7. `EpisodeSelectedGroup` を別名で残す — class だけの wrapper
8. list VM に再生 state を合流 — 直交を壊す
