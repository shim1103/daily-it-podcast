---
name: playback web の ViewModel state は判別可能 union で矛盾状態を排除し、page が見る loading/error は PageStatus 1 型へ集約する
date: 2026-09-02T23:00:00
branch: feature/playback-web-view-models
---

## 1. Decision

1. playback web の ViewModel が持つ state のうち、複数 field の組で矛盾しうるものは **判別可能 union** で表現し、無効な組を型で表現不能にする。
   1. 選択: `{ selected: false } | { selected: true; episodeId: string }`。「選択なしなのに id を読む」を型で消す
   2. 再生: `{ played: false } | { played: true; episodeId: string; phase: "loading" | "playing" | "paused" | "ended" | "error" }`。`phase` は `played: true` の枝にのみ置く。「再生対象が無いのに phase がある」「`error` なのに episodeId が無い」を型で消す
2. page（統合点の Feature component）が全画面の振る舞いを決めるために見る state は、**`PageStatus` 1 型**に集約する。
   1. `PageStatus = { kind: "loading" } | { kind: "error"; reason: PageErrorReason } | { kind: "ready" }`
   2. `PageErrorReason` は「catalog 取得失敗 / 選択 id が一覧に無い / audio 取得失敗」の 3 値。page の**構造**は `reason` で分岐しない（error 文言の出し分けにのみ使う）
   3. `derivePageStatus` は生 state（`catalogStatus` / `episodes` / 選択 union / 再生 union）から `PageStatus` を導出する。返り値に `null` を使わない
3. `playback-state.ts` の 1 行 derive 関数（`deriveIsPlaying` / `deriveIsSelected` / `deriveIsPlayed`）は削除し、呼び出し側で union の判別式（`playback.played && playback.phase === "playing"` 等）に置き換える。`deriveSelectedEpisode` / `derivePlayedEpisode`（id → episode 実体の lookup）は残す。
4. 契約値（union の形・`PageStatus` の形・関数 signature）の正本は A artifact（`apps/playback/web/src/view-models/playback-state.ts` と各 hook の型）。本 Decision はその形を写さず、方針だけを固定する。

置き換え範囲: 先行 Decision（`2026-09-02T15-00-00-feature-playback-list-episodes-audio-ref-playback-web-selection-playback-orthogonality.md`）§1-6 の「catalog 失敗・一覧に無い episodeId・audio load 失敗は blocking error として同一 Error surface で示す」の**表現手段**を、本 Decision の `PageStatus` へ置き換える。同 §1-8 の「型・derive 規則は A artifact を正とし本 file へ再掲しない」に、本 Decision 1-3 の「矛盾を判別可能 union で排除する」方針を追加する。維持範囲: 同 Decision の selection と playback の**直交**（§1-2〜§1-5）、Row / Entry / AudioControls の domain 配置（§1-7）、1 page 前提、hash 同期、物語・視覚は参照のみで維持する。「1 surface で示す」という結論自体は変えない。

## 2. Reason

1. `deriveBlockingError` は `BlockingError | null` を返し、`null` が「catalog loading 中（判定材料が無い）」と「異常なし」の 2 意味を兼ねていた。`error-handling/defensive-design.md` §7 が禁じる null 多義であり、catalog loading を「blocking なし」に読み替える握り潰しでもあった。`PageStatus` に `{ kind: "loading" }` を明示値として持たせれば、loading は「blocking でない何か」ではなく page の 1 モードになり、`null` が消える。
2. 型名 `BlockingError` と関数名 `deriveBlockingError` は「blocking でない error」との対比を含意するが、先行 Decision §1-6 が「1 page app なので surface は 1 つ」と決めた以上、non-blocking error は設計上存在しない。存在しない対比を名前が示すのは Least Astonishment（`design-philosophy.md` §5-1）の次点。`PageStatus` は「page がどう振る舞うか」だけを表し、blocking という語彙自体を不要にする。page が `catalogStatus` と `blockingError` の 2 つを突き合わせて初めて「loading か error か ready か」を判断していたのを、1 型の判別で済むようにする（`view-model.md` §3、統合点に状態判断を積まない）。
3. `playbackPhase: string | null` と `playedEpisodeId: string | null` を独立 field で持つと、`playbackPhase === "error" && playedEpisodeId === null` のような組が型上は表現でき、`deriveBlockingError` は `&& playedEpisodeId !== null` という後付けガードでそれを弾いていた（弾いた先は握り潰し）。`phase` を `played: true` の枝に閉じ込めれば、`error` phase は必ず `episodeId` を伴い、後付けガードも握り潰しも要らなくなる（make illegal states unrepresentable）。同じ理由で選択も union にする。
4. flat な primitive state（`string | null`・union 文字列）は「読み方」を関数化する必要があった（`deriveIsPlaying` 等）。判別可能 union にすると読み方が型に埋まり、1 行関数は式に戻せる。関数を残すと「union の判別」と「関数呼び出し」の 2 経路ができ DRY を崩す。lookup（`deriveSelectedEpisode` 等）は cache 引き当てという実処理があるので残す。
5. 契約の形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。本 Decision は「矛盾を union で排除する」「page の error/loading を 1 型に集約する」という**再発する方針**だけを持ち、union の具体的な枝や signature は `playback-state.ts` を正とする。

## 3. Rejected

1. `deriveBlockingError` の `BlockingError | null` を維持し、`{ kind: "error" | null }` の判定だけ整える — null 多義（loading の null / 異常なしの null）が残る。`{ kind: "loading" }` を明示値にしない限り、catalog loading を「異常なし」に読み替える握り潰しは消えない。
2. `PageErrorReason` を廃し `PageStatus = loading | error | ready` の 3 値だけにする — error 文言を「catalog が落ちた」「その episode は無い」「音声が落ちた」で出し分けられなくなる。page の**構造**は 1 surface のままだが、文言の材料として `reason` は要る。3 種の failure を別 surface にする（先行 Decision が Rejected 済み）のとは別で、`reason` は同一 surface 内の表示分岐にとどまる。
3. 選択・再生を判別可能 union にせず `string | null` のまま、`deriveBlockingError` だけ直す — `phase === "error"` と `playedEpisodeId` の整合を毎回コードで守り続けることになる。矛盾を型で消せる場所を primitive のまま残すのは、`error-handling/defensive-design.md` §7 の「判定不能を表現し、既定値へ黙って読み替えない」に反する方向。
4. `PlaybackPhase` に `idle` / `loading` を含む 6 値の flat union を維持し、`played` boolean を別に持つ — `played: false` かつ `phase: "playing"` が型で表現できてしまう。`played` と `phase` の整合をコードで守る負債が残る。
5. `deriveIsPlaying` / `deriveIsSelected` / `deriveIsPlayed` を「呼び出し側の式が散らばると DRY 違反」として関数のまま残す — union の判別式は 1 項の比較で、抽象化するほどの重複ではない（`design-philosophy.md` §2-3 KISS）。関数を残すと union を直接見る経路と関数経路が併存する方が DRY を崩す。
6. 本 Decision を先行 orthogonality Decision の本文へ追記する — 「selection と playback の直交」と「state を判別可能 union で表現する / page error を 1 型に集約する」は独立した判断軸。`decisions.md` §2-7 の「overarching な決定と派生を別 file にする」に従い、直交 Decision を参照する別 Decision にする。片方だけの supersede を効かせるため。
