---
name: playback web の page は分岐と配置だけを持ち、一覧取得の起動は useEpisodeCatalog、deep-link 復元は useHashSelectionSync へ寄せる。EpisodeRowViewModel は episode 実体を持ち page の別配列 index 引きをなくす
date: 2026-09-04T01:10:00
branch: feature/playback-web-ui-rewrite
---

## 1. Decision

1. `episode-list-page.tsx` から React の `useEffect` / `useRef` を排し、page の責務を「`useEpisodeListPage()` を呼ぶ」「`pageStatus.kind` で全画面を 3 分岐する」「`rows` を map して Feature Component を配置する」だけにする。副作用（一覧取得の起動・hash 同期・deep-link 復元）を page に置かない。

2. 一覧取得の起動を `useEpisodeCatalog` の内部へ寄せる。mount 時に自動で `load()` を 1 回呼ぶ。`useEpisodeListPage` は下位 hook を compose するだけで、起動用の `useEffect` を持たない。

3. deep-link 復元（初回ロード時、既存の `location.hash` に episodeId があればその episode を選択する）を `useHashSelectionSync` の内部へ寄せる。この hook は既に「hash → selection」方向を持ち、echo 抑止のため catalog 完了時点の既存 hash を流していなかった。その既存 hash を、catalog 完了後の初回に 1 回だけ `onHashEpisodeIdChange` へ流す挙動へ変える。page 側に初期化ゲート（`deepLinkRestoredRef` 等）や `getLocationHash` 参照を持たせない。

4. `EpisodeRowViewModel` に `episode: EpisodeData` を持たせ、`deriveEpisodeRows` が各 row に episode 実体を載せる。page は `rows.map((row) => <EpisodeRow episode={row.episode} …/>)` で描画し、`episodes` 配列を別途受け取って index で対応付ける処理をなくす。

5. `EpisodeListPageViewModel` から page が使わない投影・アクションを落とす。残すのは page が実際に参照する `selectedEpisode` / `playback` / `rows` / `pageStatus` / `toggleSelection` / `play` / `seek` / `stop` / `audioElementRef`。生 union（`selection` / `catalogStatus`）と `deselect` / `select` / `load` / `episodes` は返さない（hook 内部の合成材料であり、page も他 component も使わない）。

6. 契約値（`EpisodeRowViewModel` の形・`EpisodeListPageViewModel` の残存フィールド・各 hook の signature）の正本は A artifact（`apps/playback/web/src/view-models/`）。本 Decision は方針だけを固定し形を写さない。

置き換え範囲: 先行 Decision（`2026-09-03T13-40-00-feature-playback-web-view-models.md`）§1-6 の「ViewModel が持つ生 state は catalog / selection / playback の 3 つに固定する。derive 関数は component が必要とする形の粒度で切り、compose hook には分岐・三項・callback 再構築を残さない」を、本 Decision §4・§5 で具体化する。`deriveEpisodeRows` の投影粒度を「`episodeId` + boolean のみの最小」から「episode 実体を含む、Row がそのまま描ける形」へ広げ、`useEpisodeListPage` の戻り値から page 非使用フィールドを落とす。維持範囲: 同 §1-3（`SelectionState` の選択枝は `EpisodeData` 実体を持つ、`useEpisodeSelection` が実在検証する、playback は catalog 非依存）、§1-5（`derivePageStatus` の入力は `CatalogStatus` 1 つ）、先行 Decision `2026-09-04T00-45-00` の `PlaybackState.active` に `audioRef`・`derivePlayedEpisode` 廃止、`2026-08-27T19-20-30` の hash 同期を routing library なしで `lib` + 専用 hook + `useSyncExternalStore` に寄せる方針、`2026-09-02T15-00-00` の selection と playback の直交は参照のみで維持する。

前回の実装 flow で「deep-link 復元と load 起動を `useEpisodeListPage` に `useEffect` として足す」案を一度実装し却下した。却下理由は「compose hook に冗長な副作用を積んだ」ことであり、副作用を hook 層へ寄せること自体は否定されていない。本 Decision は寄せ先を「最下層の責務を持つ hook」（catalog / hash-sync）にすることで、compose hook を素のままに保つ。

## 2. Reason

1. `page-route.md` §3 は page に「ビジネスロジック・状態管理を置かない」と定める。現 page は `useEffect`×2 + `useRef`×2 を持ち、うち deep-link 復元は「初回だけ実行するゲート」「mount 時 hash の lazy 保持」という状態管理そのもの。`2026-08-27T19-20-30` §2 も「hash 同期を page の `useEffect` に直書きするのは §3 違反、custom hook へ畳め」と既に決めている。page から副作用を除けば、page の変更理由は「表示の分岐と配置」だけになり、hook の変更理由（state machine・同期機構）と分離する（SRP）。

2. 一覧取得の起動は `useEpisodeCatalog` の責務。この hook は「一覧の fetch と cache」を担い、`load()` を公開している。「いつ最初に load するか」は fetch のライフサイクルの一部で、呼び出し側（page や compose hook）が毎回 `useEffect(() => void load())` を書くのは DRY 違反。`useEpisodeCatalog` が mount 時に自動 load すれば、その 1 行がなくなり、`load()` は「明示 reload」専用として残る（retry 導線が要る時のため）。compose hook に effect を足さないので、前回却下された形にならない。

3. deep-link 復元は `useHashSelectionSync` が既に持つべき機能。この hook は「選択 ↔ hash の双方向同期」を担い、`onHashEpisodeIdChange`（hash → selection 方向）を既に持つ。echo 抑止で「catalog 完了時点の既存 hash」を流していなかったのは、無限書き戻しを防ぐためで、deep-link（共有リンクで開いた時の初回展開）とは目的が違う。catalog 完了後の初回 1 回だけ既存 hash を流せば、`useEpisodeListPage` の `onHashEpisodeIdChange` が既に `selection.select` を呼ぶので、追加の hook も ref も要らない。既存 hook の 1 箇所（`syncedEpisodeIdRef` の初期化と初回 read effect）の挙動修正で済む。page 側の `initialHashRef` / `deepLinkRestoredRef` / `getLocationHash` import がすべて消える。

4. `EpisodeRowViewModel` を「`episodeId` + boolean のみの最小投影」にしたのは先行実装の判断で、page が `rows` と `episodes` を別々に受け取り `rows.map((row, i) => episodes[i])` で index 対応させる形を生んだ。`deriveEpisodeRows(episodes, …)` が同じ `episodes` を 1:1 投影する前提に page が暗黙依存し、`rows` の生成条件が将来変わると `episode` が静かにズレる。row に `episode` 実体を載せれば、page は `row.episode` を直接使い、`episodes` 配列を受け取る必要がなくなる。`feature-component.md` の「Row の表示整形は `EpisodeData` から行う」は満たしたまま（Row が受け取るのが `episodes[i]` か `row.episode` かの違いで、整形責務は Row のまま）。「最小投影」は目的ではなく、当時 Row に何が要るか未確定だった名残。Row の props が固まった今は「Row がそのまま描ける形」が正しい粒度（先行 §1-6「component が必要とする形の粒度」）。

5. `useEpisodeListPage` が 15 フィールドを返すうち page が使うのは 9 個。`selection`（生 union）は `selectedEpisode`（射影）があれば page は要らず、`catalogStatus` は `pageStatus` があれば要らず、`deselect` は `toggleSelection` で兼ね、`select` は hash 同期の内部で使うだけ、`load` は §1-2 で auto-load 化すると page から消え、`episodes` は §1-4 で `row.episode` 化すると消える。使われないフィールドを返すのは、hook の内部合成材料を外へ露出してテストがそれを直接覗く構造を許し、カプセル化を弱める（`view-model.md` §2「複数 component で共有する状態」に当たらない）。返り値を page の実使用に絞れば、hook を変える時に「このフィールドを誰が見ているか」が返り値の型だけで分かる。

6. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。本 Decision は「page を分岐と配置のみにする」「起動は catalog、deep-link は hash-sync へ」「row に episode 実体を、戻り値を実使用に絞る」という再発する方針だけを固定する。

## 3. Rejected

1. deep-link 復元と load 起動を `useEpisodeListPage`（compose hook）に `useEffect` として足す — 前回この形を実装して却下された。compose hook は「下位 hook を束ねて derive を返す」だけであるべきで、そこに初期化ゲートや lazy ref を積むと、compose hook の変更理由が「合成の仕方」から「副作用のタイミング」へ広がる。副作用は、それを本来の責務とする最下層の hook（catalog の起動、hash-sync の初回反映）へ置く。

2. deep-link 復元専用の新しい hook（`useDeepLinkRestore(pageStatus, select)` 等）を切る — `useHashSelectionSync` が既に「hash → selection」方向を持つのに、その初回分だけ別 hook に切り出すと、同じ関心事（hash から選択を復元する）が 2 hook に分かれる。echo 抑止と deep-link は「初回の既存 hash をどう扱うか」という 1 つの判断の表裏で、同じ hook 内の分岐で表現するのが DRY。新 hook はフィールド 1 個・呼び出し 1 箇所の薄いラッパになり、前回却下された「冗長な hook 追加」と同型。

3. `EpisodeRowViewModel` を最小投影のまま維持し、page の `rows.map((row, i) => episodes[i])` に「`deriveEpisodeRows` が 1:1 投影する」旨の why コメントを付けて許容する — コメントは暗黙依存を明示するだけで、依存自体は消えない。`rows` にフィルタや並べ替えが入った瞬間に index がズレる。row に `episode` を載せれば構造的に対応が保証される。

4. `EpisodeListPageViewModel` の全フィールドを維持し、page が使わないものはテスト用の観測点として許容する — テストは hook の公開契約（page が使うもの）を検証すべきで、内部合成材料（`selection` 生 union 等）を `result.current.selection` で覗くのは実装詳細への結合。`selectedEpisode` / `pageStatus` / `toggleSelection` で書き換えられる。返り値を絞れば、テストも自然と公開契約だけを見る。

5. `useEpisodeCatalog` の auto-load をやめ、`main.ts`（composition root）が `EpisodeListPage` mount 前に `load()` を呼ぶ — `useEpisodeCatalog` は hook で、mount 前に外から呼べない。`main.ts` に load 呼び出しを置くと、hook のライフサイクルと分離した副作用が composition root に生まれ、テスト（`renderHook`）で auto-load を再現できない。hook 内の `useEffect` が素直。

6. page の deep-link 復元を残し、`useHashSelectionSync` は触らない — page に `useEffect` + `useRef`×2 が残り、§1・`page-route.md` §3・`2026-08-27T19-20-30` §2 の「page に同期の機微を置かない」に反したまま。echo 抑止のロジックが hash-sync hook に、その例外（初回 deep-link）が page に、と 1 つの判断が 2 層に割れる。
