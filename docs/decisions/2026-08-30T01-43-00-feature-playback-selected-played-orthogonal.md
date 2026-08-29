---
name: playback 一覧の selected / played を直交させ item を隠さない
date: 2026-08-30T01-43-00
branch: feature/playback-audio-player-ui-design-api-refactoring
---

## 1. Decision

1. `EpisodeFocus` 型・React state・`isFocused` prop は作らない。紫線は既存 CSS（`:hover` / `:focus-visible`）のみ
2. **selected** と **played** は独立（直交）。一方が他方を暗示しない
3. どの状態でも他の episode item は **隠さない**（`isFocused && !isSelected → null` を削除）
4. 空状態は `null` / `undefined` で表さない。名前付き discriminated union を使う

```ts
type EpisodeSelection =
  | { kind: "none" }
  | { kind: "open"; episodeId: string; detail: EpisodeDetailState };

type EpisodePlayback =
  | { kind: "idle" }
  | { kind: "active"; episodeId: string; audioRef: string };
```

5. **selected 操作**
   - 行 select → detail を開く（他 item も見える）
   - 同じ episode を再 select → detail を閉じる（`kind: "none"`）
   - 別 episode を select → 前の detail は閉じ、新しい detail だけ開く
6. **played 操作**
   - play pill → その episode を再生（`kind: "active"`）。detail 開閉は変えない
   - 別 episode を play → 再生対象だけ乗り換え。他 item は見える
7. `EpisodeList` から `EpisodePlayer`（描画）は除去する。`useEpisodePlayback` は **List が持つ**（page は持たない）。List は map と item 向け状態の配布だけ
8. 共有 hidden `<audio>` は Item が mount する。再生中 item、または global idle 時の先頭 item のみ
9. `useEpisodePlayback` が `audioElementRef`・`EpisodePlayback`・`play` を持つ。page は `playback` / `audioElementRef` を props で持たない・渡さない
10. `useEpisodeListViewModel` の selection は `EpisodeSelection`（`kind: "none" | "open"`）。`selection: null` は廃止
11. Item への props は boolean + `null` を混ぜない。item 単位の discriminated union にする

```ts
type EpisodeListItemSelection =
  | { kind: "none" }
  | { kind: "open"; detail: EpisodeDetailState };

type EpisodeListItemPlayback =
  | { kind: "silent" }                         // audio を mount しない
  | { kind: "host"; audioRef: string };        // mount する（idle 先頭は ""、active は url）
```

`isSelected` / `isPlayed` / `detail: null` / item への global `playback` 丸渡しは禁止
12. item + detail の描画単位は `EpisodeListItem`（内包 `EpisodeDetail`）。`selection.kind === "open"` の時だけ detail を描く。`EpisodeSelectedGroup` は open 行の layout に使う
13. 先行 Decision `2026-08-29T23-30-00-...` の Decision 4 と、page が playback props を抱える実装は **本 Decision が上書きする**

## 2. Reason

1. shim 要求: focused / selected / played を直交させ、選択中も他 item を残す。現行 `isFocused = selection !== null` と hide はこれを破る
2. focused の「離れる→色が戻る」は CSS で足りる。React に上げると blur 追従が余分になる（KISS）
3. List に EpisodePlayer **描画**を置くのは SRP 違反。ただし playback **hook** は List が持つ（page に上げると page が `playback`/`audioElementRef` を props 型で抱える）
4. `null` / boolean の混ぜ書きより item 単位 union の方が「この行の selected/played」が型で読める
5. page がもともと持っていなかった playback 配線を page props に戻すのは、逃げ場所を page に移しただけである

## 3. Rejected

1. `EpisodeFocus` 型・state — shim 明示で不要。CSS に任せる
2. selected 中に他 item を隠す（旧 focus-mode）— 本要求と矛盾
3. hidden audio を EpisodeList / page に **描画として**残す — List/page の SRP を破る
4. Item が apiClient で detail を自前 fetch — `feature-component` 契約違反
5. played を selected に追従させる — 直交要求と矛盾
6. page が `playback` / `audioElementRef` を List へ渡す — page がもともと持たない関心を props 型に露出する
7. Item に `isSelected` + `detail: null` + global `playback` を同居させる — Decision §4 の「null 禁止・明示型」と矛盾する
