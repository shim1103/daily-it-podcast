---
name: PlaybackState の active 枝に audioRef（不変値）だけ持たせ derivePlayedEpisode を廃し、再生 UI の分岐を union tag のナローイングへ寄せる
date: 2026-09-04T00:45:00
branch: feature/playback-web-ui-rewrite
---

## 1. Decision

1. `PlaybackState` の `active` 枝に `audioRef: string`（episode の音源 path、契約 `ListEpisodesResponse` 由来）を追加する。`play` / `seek` が再生対象を確定する 1 箇所で catalog から引き当てて state に固定する。
2. `derivePlayedEpisode(episodes, playback)`（`playback.episodeId` を毎 render で catalog から lookup し `EpisodeData | null` を返す）を廃止する。`AudioControls` が要る値は `audioRef` だけで、それは `active` 枝が直接持つ。
3. page（統合点）の再生 UI 分岐を `playedEpisode !== null` から `playback.kind === "active"` の判別可能 union ナローイングへ変える。`AudioControls` へ渡す src は `buildRequestUrl(baseUrl, playback.audioRef)` で page が組む（URL 構築は表示側に残す）。
4. selection 側は変えない。`SelectionState` の「選択中」枝が `EpisodeData` 実体を持ち、`selectedEpisode` は `selection.selected ? selection.episode : null` の union ナローイングで得る形（先行 Decision `2026-09-03T13-40-00` §1-3 が確定済み）を維持する。page の Entry 分岐は `row.isSelected && selectedEpisode !== null` の二重条件をやめ、`selectedEpisode?.episodeId === row.episodeId` の単独判定にする。
5. 契約値（`active` 枝の形・`play` / `seek` の引き当てタイミング）の正本は A artifact（`apps/playback/web/src/view-models/playback-state.ts` と `use-episode-playback.ts`）。本 Decision は方針だけを固定し形を写さない。

置き換え範囲: 先行 Decision（`2026-09-03T13-40-00-feature-playback-web-view-models.md`）§1-3 の「playback は id だけ持ち、実体 lookup は `derivePlayedEpisode` で毎回行う（`derivePlayedEpisode` の『無ければ null』は残す）」を、本 Decision §1・§2 で部分的に置き換える。`active` 枝が持つのは `EpisodeData` 全体ではなく `audioRef` 1 値。`derivePlayedEpisode` は廃止する。維持範囲: 同 §1-3 の「playback を catalog に依存させない（一覧から消えた episode を再生し続けられる）」「selection の『選択中』枝は `EpisodeData` 実体を持つ」「`useEpisodeSelection` が選択確定時に実在検証する」は維持する。同 §1-4（phase の判別可能 union）、§1-5（`derivePageStatus` の入力は `CatalogStatus` 1 つ）、§1-6（derive は component が必要とする形の粒度、compose hook に分岐を残さない）は維持する。さらに先の Decision（`2026-09-03T16-20-00`）の play / seek の 2 操作分離、`2026-09-02T15-00-00` の selection と playback の直交、Row / Entry / AudioControls の domain 配置は参照のみで維持する。

## 2. Reason

1. 先行 Decision `2026-09-03T13-40-00` §1-3 は selection には実体を持たせ playback には持たせない非対称を選んだ。その Reason 3 は「playback と selection は寿命が違う。再生対象は一覧の更新から独立して生き続けてよいので、`EpisodeData` 全体を固定すると title 等が stale になる」だった。この理由は `EpisodeData` 全体には当てはまるが、`audioRef` には当てはまらない。`audioRef` は episode の音源 path で、episode が存在する限り不変であり、一覧が再取得されても変わらない。stale になりうる field（title・date・body）を state に入れず、不変の `audioRef` 1 値だけ入れれば、§1-3 の「catalog 非依存で寿命が独立」を保ったまま lookup を消せる。

2. `derivePlayedEpisode` は `playback.episodeId` を毎 render で `episodes.find` する。返り値 `EpisodeData | null` の `null` は「再生なし」と「再生中だが一覧に無い」を兼ね、page はそれを `playedEpisode !== null` で判定していた。これは `PlaybackState` が既に `kind: "idle" | "active"` の判別可能 union（`2026-09-02T23-00-00` §1-1 で確定）を持つのに、その tag を使わず lookup 結果の null 有無で分岐している。`active` 枝が `audioRef` を持てば、page は `playback.kind === "active"` で narrow でき、lookup も null 分岐も消える。`make illegal states unrepresentable` と「union の判別式は 1 項の比較、関数に抽象化しない」（`2026-09-02T23-00-00` §1-3）の両方に沿う。

3. `AudioControls` の `@require` は `audioSrc` = 再生対象 episode の URL。その材料は `audioRef`（契約 path）と `baseUrl`。`baseUrl` は page が prop で持つ既定値（`2026-08-27T19-20-30` 系で `""` 固定）。`audioRef` を `active` 枝に載せれば、page は `buildRequestUrl(baseUrl, playback.audioRef)` を組むだけになり、`playedEpisode` という射影 state を hook 戻り値から消せる。URL 文字列そのものを hook が返す案は先行 Decision `2026-09-03T16-20-00` Rejected 6 が「`<audio src>` の張り替え順序問題が表示側に波及する、別 Issue」として却下済みで、本 Decision はそれを侵さない。hook が持つのは契約由来の path、URL 構築は page、という分界を保つ。

4. selection 側を触らないのは、先行 Decision §1-3 が既に「選択枝は実体を持つ、lookup は 1 箇所、`deriveSelectedEpisode` 廃止」で到達点を定義しているため。`selectedEpisode` は `SelectionState` union の素直なナローイング結果（`T | null`）で、これ以上寄せる先がない。page の `row.isSelected && selectedEpisode !== null` は、`selectedEpisode` が既に選択 episode を一意に指すので `isSelected` との AND が冗長。`selectedEpisode?.episodeId === row.episodeId` にすれば 1 式で、かつ「どの row の下に Entry を出すか」を `selectedEpisode` 単独で決められる。

5. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。本 Decision は「`audioRef` を `active` 枝へ、`derivePlayedEpisode` を廃止、再生分岐を union tag へ」という再発する方針だけを固定する。

## 3. Rejected

1. `PlaybackState` の `active` 枝に `EpisodeData` 実体全体を持たせる（user の当初提案）— 先行 Decision `2026-09-03T13-40-00` §1-3 Reason 3 の「再生対象は一覧の更新から独立して生き続けてよく、`EpisodeData` 全体を固定すると title 等が stale になる」に正面から反する。`play` 後に catalog が再取得され title が変わっても、再生中の episode は古い title を持ち続ける。必要なのは `audioRef` だけで、それは不変なので stale 問題が無い。全体を巻き込む必要がない。

2. `derivePlayedEpisode` を残したまま page の `playedEpisode !== null` だけ `playback.kind === "active"` に変える — `kind === "active"` かつ `derivePlayedEpisode` が `null`（再生中だが一覧に無い）のとき、page は `active` と判定するのに `AudioControls` へ渡す `EpisodeData` が無い、という不整合が残る。`audioRef` を枝に載せて lookup 自体を消さないと、tag と lookup 結果の二重管理が続く。

3. `AudioControls` に `audioRef` と `baseUrl` を別々に渡し、URL 構築を `AudioControls` の中でやる — `buildRequestUrl` の呼び出しが Feature component に入る。`build-request-url.ts` は utils（純粋関数層）で component から import してよいが、「再生対象の URL をどう組むか」は page の統合配線の一部で、component は組み上がった src を受け取るだけにする方が `feature-component.md` §4「ViewModel から受け取った状態を props で描画するのみ」に沿う。現状の `audioSrc: string` prop を維持する。

4. `selectedEpisode` も `derivePlayedEpisode` と同様に廃止し、page が `selection.selected` で分岐して `episodes` から実体を引く — `SelectionState` の「選択中」枝は既に `episode` 実体を持つ（§1-3）。`episodes` から引き直す必要がない。`selection.selected ? selection.episode : null` は union の直接ナローイングで、`derivePlayedEpisode` のような catalog lookup を伴わない。廃止対象ではない。

5. `PlaybackState` を変えず、page で `playedEpisode !== null` を許容する — `null` の多義（再生なし / 再生中だが一覧に無い）が page の分岐に残り続ける。`PlaybackState` が判別可能 union を持つのにその tag を使わず lookup 結果の null で分岐するのは、`2026-09-02T23-00-00` §1-1・§1-3 が排したパターンそのもの。
