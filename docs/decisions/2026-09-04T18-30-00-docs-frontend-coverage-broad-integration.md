---
name: playback frontend に Page 合成入口 1 つの Broad Integration を置く。合成入口単位で 1 file、外部は HTTP と <audio> 再生を double、location.hash は実物
date: 2026-09-04T18:30:00
branch: docs/frontend-coverage-broad-integration
---

## 1. Decision

1. playback frontend に Broad Integration を **1 file** 置く。対象は合成入口 `EpisodeListPage`（`useEpisodeListPage` 配下の catalog / selection / hash-sync / playback、feature / primitive Component、`playback-state.ts` の derive を実物で通す）。file 分割軸は **合成入口単位**であり、hook 単位でも Component 単位でもない。playback frontend の合成入口は単一 page なので Broad file も 1 本。
2. Broad で double にするのは **真に外部の依存だけ**とする。
   1. HTTP（`PlaybackApiClient`）— `ApiResult` を返す Stub。実 network I/O 契約は Narrow の所有（先行 Decision `2026-08-30T16-20-02` §1-1/§1-2 を維持）。
   2. `<audio>` 再生（`lib/audio-element.ts` の `play` / `pause` / `currentTime` / lifecycle event 購読）— Fake。happy-dom に再生エンジンが無く、実 browser 挙動は E2E の所有。
   3. `location.hash` / `hashchange` は **実物のまま**使う。happy-dom が実装を持ち、選択 ↔ hash 双方向同期の合成そのものが検証対象のため、ここを double にすると見る対象が消える。
3. Broad が assert してよいのは、合成で初めて生じる関係（接続・引数伝播・状態伝播・error 伝播・直交性）の **代表**に限る。下位 hook の内部分岐（`toPlaybackPhase` の全枝、`derivePageStatus` の 3 分岐、`use-hash-selection-sync` の echo 抑止の細かい枝、`listEpisodes` の全 error code）は再 assert しない。それらは mapping test / 単一 hook SU が所有する（`testing-strategy/minimization.md` §2-1/§2-2）。
4. file 名分類語は **`broad_integration`**。配置は `test/integration/` 直下（worker の Broad と同じ専用領域）。docstring に `scope: Broad Integration` / `real:` / `double:` を明記する。
5. Broad Integration を Unit coverage 計測分母に **含める**（Sociable Unit + secret なし Narrow + Broad の合算）。先行 Decision `2026-08-30T16-20-01` §1-2「Broad 以上は分母に入れない」を frontend Broad についてのみ置き換える。理由は §2-5。
6. 観点（catalog 状態分岐 / selection ↔ manuscript / 再生操作の配線 / hash ↔ selection 同期 等）で file を割らない。観点は `describe`、1 振る舞いは `it` の GWT 1 個で表す（`testing-strategy/naming-and-layout.md` §1/§2）。
7. test の具体（どの DOM を見るか、Stub / Fake の実装、case 名）の正本は A artifact（`apps/playback/test/integration/` の Broad file と `web/src/lib/` の Fake）。本 Decision は方針だけを固定し形を写さない。

置き換え範囲: 先行 Decision（`2026-08-30T16-20-02-docs-playback-integration-e2e-plan.md`）§1-3「Page → hooks → api / lib を全部実物にした frontend Broad Integration は当面やらない。配線の最終結果は認証後 E2E と下位 SU に寄せる」を本 Decision §1・§2・§3 で置き換える。同 Decision §1-1/§1-2（`fetch` / apiClient Stub は Sociable Unit、実 network だけ Narrow）は維持する。同 Decision Rejected §2「層ごと（component–hooks、hooks–lib）に専用 Integration を今切る案」は維持する（本 Decision は層ペア単位の Integration を作らず、合成入口 1 つに集約するため矛盾しない）。先行 Decision `2026-08-30T16-20-01` §1-2 のうち Broad を Unit coverage 分母から外す方針を、frontend Broad についてのみ本 Decision §5 で置き換える。worker Broad を分母に入れない方針、および E2E を分母に入れない方針（同 §1-2）は維持する。selection と playback の直交（`2026-09-02T15-00-00` §1-2）、E2E の収集・定時（`2026-08-30T16-20-00`）は維持する。

## 2. Reason

1. 先行 Decision `2026-08-30T16-20-02` §1-3 が frontend Broad を見送った主因は「Page 合成の最終結果を E2E が既に見るなら minimization に反する重複」だった。しかし現行 E2E（`test/e2e/authenticated_playback.e2e.spec.ts`）は `PLAYWRIGHT_BASE_URL` / `PLAYWRIGHT_STORAGE_STATE` 未設定で `test.skip` される remote 専用で、PR gate にも pre-push にも載らない（`2026-08-30T16-20-00` §1-2）。つまり「catalog 完了 → hash 同期開始 → deep-link 復元」「`listEpisodes` 失敗 → 全画面 Error UI」「selection ⊥ playback」といった **4 hook を跨ぐ配線の回帰を PR 前に落とす層が今どこにもない**。E2E は実 Access・実 Drive 到達の最終確認であって、CI 常時 gate ではないため、重複の前提が崩れている。合成入口 1 つの Broad をローカル / PR gate で回せば、この穴が埋まる。
2. file を合成入口（Page）単位にするのは、Broad の検証対象が「複数の独立した振る舞い単位の**関係**」（`testing-strategy/levels.md` §4）だからである。hook 2 個の組み合わせはまだ単一 unit とその内部協調で Sociable Unit の範囲（同 §3-2）。hook 単位で file を割るとペア数だけ file が増え、しかもそれぞれが SU にしかならない。Component 単位で割ると `EpisodeItem` の Broad は `onPlay` Stub を受けるだけで hook 間配線を通らず、既存の Component SU と重複する。playback frontend の合成入口は単一 page なので、Broad file も 1 本が過不足ない。
3. Broad と Narrow で外部境界の double 方針は逆である（`testing-strategy/levels.md` §3-2）。Narrow は境界 provider を実物にして実 I/O 契約を見る。Broad は境界を double にして、応答が合成経路を流れて UI になる伝播だけを見る。HTTP 実 I/O 契約（status → `ApiResult` 変換、schema validation、connection error 吸収）は既存 `playback_api_client.narrow_integration.test.ts` の所有であり、Broad で `apiClient` を Stub しても二重にならない（同 §7）。
4. `location.hash` だけ実物にするのは、`use-hash-selection-sync` の検証したい関係が「selection 変化 → hash 書き込み → `hashchange` 発火 → caller 通知 → echo 抑止で無限ループしない」であり、実 `location.hash` と実 event を回さないと現れないからである。`hash-selection-adapter.fake.ts` を使うと `hashchange` の発火タイミングが再現できず、echo 抑止の合成が検証できない。`<audio>` 再生と HTTP は happy-dom / test 環境が実装を持たない、または実 I/O が別 Scope の所有なので double にする。「テスト環境が実装を持つブラウザ API は実物、持たない副作用と別 Scope 所有の境界は double」という切り分け。
5. frontend Broad を Unit coverage 分母に入れるのは、`2026-08-30T16-20-01` §1-2 が worker Broad / E2E を分母から外した理由（「結線・最終結果が目的で、行 coverage の逃げ道にしてはならない」）が frontend Broad には当てはまらないからである。frontend Broad が実物で通す `useEpisodeListPage` とその配下 hook・derive は、Drive HTTP のような真の外部を持たず、Broad の経路がそのまま製品ロジックの実行になる。SU で個別に埋めた分岐と Broad の合成経路で二重に数えても「上位 Scope で下位の穴を埋める」構図にならない（worker Broad は Drive Stub を挟むため経路の一部が製品外）。逆に分母から外すと、Broad へ移した test が計測していた行が SU 側に無い場合に検出力がすり抜ける。
6. 観点で file を割らないのは、`testing-strategy/naming-and-layout.md` §1 が要求する「file 名から Scope × Sociability が一意に判別」が Scope 軸の話であり、runner の収集も Scope でしか行われない（同 §1 末尾「分類名で収集する runner 設定を同時に用意」）からである。観点別の収集設定は存在しない。観点を file 名に入れると `episode_list_page.error_propagation.broad_integration.test.ts` のように冗長化し、同じ合成入口の Broad が複数 file へ散って `minimization.md` §2-5「同じ例外 case を複数 test file で重複検証しない」の監視が効かなくなる。観点は `describe`、1 振る舞いは `it` の GWT 1 個。
7. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。

## 3. Rejected

1. frontend Broad を引き続き作らず、hook 間配線の回帰を認証後 E2E に委ねる（先行 §1-3 の維持）— E2E が CI 常時 gate でなく remote skip 前提である以上、PR 前に配線回帰を落とす層が無いままになる。「E2E が見るから重複」の前提が現行構成では成立しない。
2. hook ペア（catalog × selection、selection × hash-sync 等）ごとに Integration file を切る — hook 2 個の合成は Sociable Unit の範囲で、Broad にならない（`levels.md` §3-2）。ペア数だけ file が増え、いずれも SU にしかならず、`2026-08-30T16-20-02` Rejected §2 が退けた「層ペア専用 Integration」に該当する。
3. Component 単位（`EpisodeItem` の Integration、`AudioControls` の Integration）で file を割る — Component を実物にしただけでは、その Component と子・Stub された handler の合成であり Sociable Unit（`levels.md` §3-2）。hook 間配線を通らず既存 Component SU と重複する。
4. Broad でも `<audio>` を happy-dom の生要素のまま使い、`src` 属性だけ見る（現行 `episode-list-page.sociable_unit.test.ts` の方式を据え置く）— phase 通知（`onPhaseChange("playing")` → `deriveEpisodeRows` の `isPlaying` → DOM 強調）と audio load 失敗の非 blocking error 伝播（`phase:"error"` が `pageStatus` を汚さない）という、合成で初めて見える状態伝播が検証できない。能動的に phase event を発火できる Fake が要る。
5. Broad で `location.hash` も Fake（`hash-selection-adapter.fake.ts`）にする — `hashchange` の実発火タイミングが再現できず、echo 抑止（hash 書き戻しで `listEpisodes` が再呼びされない）の合成が検証できない。実 `location.hash` は happy-dom が持つので double にする理由が無い。
6. frontend Broad を Unit coverage 分母から外す（`2026-08-30T16-20-01` §1-2 をそのまま適用）— frontend Broad は真の外部を Stub せず経路がそのまま製品ロジックの実行になるため、worker Broad（Drive Stub を挟む）と同じ「行 coverage の逃げ道」懸念が当たらない。分母から外すと SU に無い行を Broad が計測していた場合に検出力が漏れる。
7. 観点別（error 伝播 / 状態伝播 / 直交性 …）に Broad file を分ける — runner は Scope でしか収集せず、観点別収集設定が無い。file が散って重複検証の監視が効かなくなる。観点は `describe` で表す。
