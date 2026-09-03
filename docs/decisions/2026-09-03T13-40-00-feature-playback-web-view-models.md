---
name: playback web の error は blocking / non-blocking / 一過性の 3 層へ分け、invalid な選択は state に入れず、selection は選択中 episode の実体を持つ
date: 2026-09-03T13:40:00
branch: feature/playback-web-view-models
---

## 1. Decision

1. error の表示責務を 3 層に分ける。1 page 前提（route なし）を維持したまま、「使えない度」で表示の重み・位置・寿命を変える。
   1. **blocking**（本体を描けない）: catalog 取得失敗と catalog loading。全画面を占める。寿命は `CatalogStatus` の派生に固定し、独立した error flag state を持たない。回復は明示 retry（`load()` 再実行）。閉じる操作は提供しない
   2. **non-blocking**（本体は描ける、一部が失敗）: audio 取得失敗。再生 controls 近傍に inline 表示する。寿命は `PlaybackState` の error 枝が立っている限り。回復は「その episode の再生 retry」。閉じる操作は提供せず、「直す操作」だけ提供する
   3. **一過性**（起きた瞬間に一度伝える）: 無効な hash を弾いたことの告知など。captcha-message-app 式の `Notification | null`（別 state、時間または操作で消滅）。導入は必要になってからでよい
2. 無効な選択（catalog success かつ hash 由来の episodeId が一覧に無い）は **error にしない**。hash → selection 反映の時点で reject し、selection は「選択なし」のまま、hash も消す。「一覧に存在する episode を選択中」という不変条件を型で保てない状態を作らない。
3. `SelectionState` の「選択中」枝は、episode の **id ではなく実体（`EpisodeData`）** を持つ。lookup は選択を確定する 1 箇所でだけ行う。`useEpisodeSelection` は catalog の episodes を受け取り、選択確定時に実在を検証する。playback は catalog に依存させない（一覧から消えた episode を再生し続けられる。`derivePlayedEpisode` の「無ければ null」は残す）。
4. `PlaybackState` の再生中 phase は、string の並びではなく **判別可能 union** にする。`error` phase は失敗理由（`reason`）を枝に持つ。他 phase は追加データを持たない。
5. `PageStatus` は blocking の判定だけを担う型にする。`derivePageStatus` の入力は `CatalogStatus` 1 つ（`episodes` / `selection` / `playback` を受け取らない）。non-blocking の error は `PlaybackState` の枝が持ち、`PageStatus` からは分離する。
6. ViewModel が持つ生 state は catalog / selection / playback の 3 つに固定する。hash は Driven Adapter が持つ外部状態で ViewModel state ではない。derive 関数は「component が必要とする形」の粒度で切り、compose hook には分岐・三項・callback 再構築を残さない。hash 同期を catalog 完了まで保留する判断は `useHashSelectionSync` の内部へ置く（page の関心事にしない）。
7. 契約値（union の枝・`PageStatus` の形・関数 signature・derive の粒度）の正本は A artifact（`apps/playback/web/src/view-models/` の型と関数）。本 Decision は方針だけを固定し、形を写さない。

置き換え範囲: 先行 Decision（`2026-09-02T23-00-00-feature-playback-web-view-models.md`）の §1-2（`PageErrorReason` に `catalog-load-failed` / `invalid-selection` / `audio-load-failed` の 3 値を持たせ、`derivePageStatus` が生 state から `PageStatus` を導出する）を、本 Decision §1・§2・§5 で置き換える。`invalid-selection` と `audio-load-failed` は `PageStatus` から外れる。`derivePageStatus` の入力は `CatalogStatus` 1 つになる。維持範囲: 先行 Decision の §1-1（判別可能 union で矛盾状態を排除する）、§1-3（1 行 derive 関数を式へ、lookup 関数は残す）、§1-4（契約値の正は A artifact）は維持する。さらに先の Decision（`2026-09-02T15-00-00-...-orthogonality.md`）の selection と playback の直交、Row / Entry / AudioControls の domain 配置は参照のみで維持する。「使えない状態を 1 page で示す」という結論は維持し、その示し方を「使えない度で 3 層に分ける」へ具体化する。

## 2. Reason

1. 先行 Decision `2026-09-02T23-00-00` は「1 surface で示す」を `PageStatus` の error 枝 1 つへ集約した。しかし 4 つの失敗は「本体を描けるか」が違う。catalog 取得失敗は描くものが 0 なので全画面が正しい（`defensive-design.md` §8「外部依存・恒久的失敗 → 専用画面」）。audio 取得失敗は一覧も原稿も見られる状態で、user の再生操作への応答なので操作の近傍に inline で返すのが Least Astonishment（同 §8「入力起因・回復可能 → 入力近傍に inline」）。同じ型の同じ枝に押し込めると、表示位置を後から文言で再現することになり、位置情報が型から失われる。
2. 無効な選択を error にしていたのは 3 重に誤り。(a) hash は URL・bookmark・共有リンク由来で user の操作起因ではなく、見せても user に取れる action がない。(b) 一覧は正常に描ける。(c) `SelectionState` が「実在する episode を選択中」を保証しないため、`derivePageStatus` が毎回実在 lookup を行い、`deriveSelectedEpisode` の lookup と重複する（先行レビューで DRY 違反の芽と指摘済み）。無効な選択を state に入れなければ、この 3 つが同時に消える。`make illegal states unrepresentable` の直接適用。
3. `SelectionState` が id だけを持つと「その id が一覧にある」保証が型にないため、選択中 episode を使う箇所すべてが lookup + null 分岐を負う。選択確定時に実体を持たせれば lookup は 1 箇所に集約し、`deriveSelectedEpisode` が不要になる。`useEpisodeSelection` が catalog に依存するのは正しい依存で、「選択」は「何から選ぶか」抜きに定義できない。playback を同じにしないのは寿命が違うため（再生対象は一覧の更新から独立して生き続けてよい）。非対称だが、その非対称は state の寿命差を正しく写している。
4. `PlaybackState` の phase が string の並びだと、`error` だけが追加情報（失敗理由）を持ちたいのに全 phase が同じ形になり、理由を別 state に逃がすか捨てるかになる。判別可能 union にすれば `error` 枝にだけ `reason` を持たせられ、他 phase は増えない。先行 Decision §1-1 の「矛盾を型で排除する」を phase 内部にも適用するだけで、新しい方針ではない。
5. `derivePageStatus` が 5 入力を対等に受け取りながら、実装は catalog を主軸に分岐している（先行実装で catalog の 2 分岐は無条件 return、success だけ子関数）。型と実装の構造が食い違い、それが子関数の命名の座りの悪さの原因になっていた。blocking 判定は catalog だけで決まる（selection/playback の失敗は non-blocking へ移す）ので、入力を `CatalogStatus` 1 つに絞れば型と実装が一致し、Parameter Object も不要な単項関数になる。
6. state を正規化せずに derive を細かく割ると、compose hook に「正規化されていない state を組み立て直す式」が溜まる（`selection.selection` の二重ネスト、union から boolean を作り直す 3 つの callback）。derive を「component が必要とする形」まで投影しきれば、compose hook は下位 hook 呼び出しと derive 呼び出しの直線コードになり、component は map するだけになる（`feature-component.md` §4「ViewModel から受け取った状態を props で描画するのみ」）。hash 同期の保留判断が page 側にあるのは責務の置き違いで、「catalog 完了まで同期しない」は hash 同期の関心事。
7. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。

## 3. Rejected

1. 4 つの失敗を `PageStatus` の error 枝に集約したまま維持する（先行 Decision の形）— 表示位置・寿命・回復手段が違うものを 1 つの型で表し、表示側が reason 文字列から位置を復元することになる。catalog 失敗と audio 失敗を同じ全画面にすると、音声が鳴らないだけで一覧も原稿も読めなくなる。
2. 無効な選択を error として表示し、user に deselect させる — user に取れる action がない（hash を手で直せは UI ではない）。判断を user へ押し付けるだけ。自動 deselect + hash クリアで回復できる。
3. 無効な選択の告知を全く出さない（自動 deselect のみ、無言）— 共有リンクが古かった等を user が知る手段がなくなる。ただし出すなら一過性の toast 1 回で足り、`PageStatus` に持たせる必要はない。導入は必要になってから（YAGNI）。
4. audio 失敗の inline に閉じる（×）ボタンを付ける — 「閉じた」という第 4 の状態が要り、`{ phase: "error"; dismissed: boolean }` へ育つ。原因が解消済みなのに dismissed が残る illegal state を作る。retry ボタン（直す操作）だけ提供すれば、再生成功で phase が遷移して表示が自然に消える。
5. catalog 失敗の error を独立した `useState` の flag で持つ — catalog が success になったのに error flag が残る illegal state が表現可能になる。`CatalogStatus` の派生に固定すれば型上不可能。
6. `SelectionState` を id のままにして branded type（`EpisodeId`）で hash 由来の未検証 string と区別する — 選択枝に episode 実体を持たせれば実在が型で保証され、branded type は不要になる。KISS 優先。
7. compose を JSX 層へ逃がす（captcha-message-app 式）/ Feature component が個別に小 hook を使う — catalog / selection / playback は相互依存し、row 投影という共通の合流点がある。JSX へ逃がすと合流ロジックが JSX 内に散り test 不能になる。component ごとに hook を呼ぶと同じ state の複数 instance か Context が要る。1 page SPA に Context は KISS 違反。単一 compose hook を保ち、肥大の原因（正規化不足）を潰すのが筋。
8. union をやめて flat な object（`{ played: boolean; episodeId?: string; phase?: string }`）にする — undefined チェックが同じ回数必要で、しかも忘れても compile が通る。判別可能 union の narrow コストは「触ってはいけない時に触れない」利益そのもの。state の直積が本質的な場面（全組合せが legal）でも、枝が 10 を超える場面でもない。
