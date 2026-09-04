---
name: playback web の <audio> は ViewModel が src を命令的に張る。stop は位置保持の pause、topic sec bar は state 不問でそこから再生開始とする
date: 2026-09-04T13:17:57
branch: refactor/playback-web-audio-adapter-design-2
---

## 1. Decision

1. `<audio>` の音源指定を Component の React props（controlled `<audio src>`）から外し、`useEpisodePlayback` が `play` / `seek` の直前に ref 経由で命令的に張る。Adapter に「`el.src` を差し替え `load()` する」操作を持たせ、hook は「今 `<audio>` に張った URL」を自前の ref で覚えて差分判定する（`el.src` の getter は常に絶対 URL を返すため、baseUrl が空で相対の音源 path を渡す構成だと `el.src !== 引数` が毎回真になり、同じ episode の seek のたび `load()` で再生が途切れる）。
2. 別 episode への切替で `<audio>` を頭出しする専用操作（pause + `currentTime = 0` + `load`）を廃し、src 差し替え操作の `load()` に頭出しを兼ねさせる。
3. `stop()`（UI 表示は「停止」）は頭出ししない。その位置で `pause()` するだけにし、state は `active` の `phase: "paused"`（`positionSec` 保持）を維持して `idle` には戻さない。同じ episode を再び `play` すると保持した `positionSec` の続きから再生する。
4. `seek(episodeId, positionSec)`（topic の sec bar から呼ぶ）は「移動だけ」をやめ、押した時点の state（`idle` / 別 episode 再生中 / 同じ episode の停止中）に関わらず、その episode をその位置から再生開始する（`shouldPlay` は常に true）。別 episode を指していれば §1-1 の src 差し替えを伴う。
5. 音源 URL の組み立て（`buildRequestUrl(baseUrl, audioRef)`）を page から `useEpisodeListPage` へ移す。page は `baseUrl` を hook へ渡すだけで、`play` / `seek` へ絶対 URL を解決するのは hook の責務。`AudioControls` は `src` prop を持たず `audioRef`（ref）だけ受け、再生中かに関わらず常に mount する。
6. 行の「再生 ↔ 停止」トグルの判定は `phase` が `"loading"` または `"playing"` のとき（`isActivePlayback`）。`"paused"` / `"ended"` / `"error"` は「再生」表示に戻し、押すと §1-3 の続き再生になる。「今まさに音が出ているか」（`"playing"` のみ）は別 flag として視覚強調にだけ使う。
7. 契約値（Adapter 関数の signature、`stop` / `seek` が遷移させる `phase`、`useEpisodeListPage` の引数）の正本は A artifact（`apps/playback/web/src/view-models/` と `lib/audio-element.ts`、`pages/`）。本 Decision は方針だけを固定し形を写さない。

置き換え範囲: 先行 Decision（`2026-09-03T16-20-00-feature-playback-web-view-models.md`）を次のとおり部分的に置き換える。§1-3（`seek` は `<audio>.play()` を呼ばない「移動だけ」、seek の 3 分岐で「同じ episode 停止中 → 移動して停止のまま」「違う episode → 切替して移動して停止」）を本 Decision §1-4 で置き換え、seek は常にそこから再生開始とする。§1-5-2（`startAudioPlayback` を `seekAudioElement(el, 0, {play:true})` へ吸収）は維持。§1-5 が前提にしていた「別 episode 切替で前 audio を stop し `currentTime` を reset する」専用操作（先行 `2026-09-02T15-00-00` §1-3 由来）を本 Decision §1-2 で置き換え、src 差し替えの `load()` に頭出しを兼ねさせる。§6（「違う episode への `<audio src>` 張り替え順序」は本 Issue 範囲外、src を ViewModel が持つ形は表示側配線 Issue で行う）を本 Decision §1-1・§1-5 で解決し、src は ViewModel（hook）が命令的に持つ形へ確定する。維持範囲: §1-1（`PlaybackState` の `kind` tag union）、§1-2（再生位置を state で保持、停止中も位置は残る）、§1-4（state と ref を 1 ラッパに通す）、§1-5-1（`seekAudioElement` の `{play: boolean}` オプション）、§1-5-3（`timeupdate` / `loadedmetadata` の購読）は維持する。先行 Decision `2026-09-04T00-45-00` の「`active` 枝に `audioRef`（不変値）を持たせ catalog 非依存」は維持し、その §1-3 が page に置いた `buildRequestUrl` 呼び出しを本 Decision §1-5 で hook へ移す。selection と playback の直交（`2026-09-02T15-00-00` §1-2）は維持。

## 2. Reason

1. `<audio src>` を `playback` state から controlled で与えると、`play` / `seek` が state を commit した直後に命令的な `seekAudioElement` / `play()` を呼ぶ経路で、React の再 render・DOM 反映が挟まる前に古い（または空の）src の要素に対して `currentTime` セットと `play()` が走る。`play()` は src の無い要素で reject し、seek の `currentTime` は反映後に上書きされる。src 指定も Adapter 経由の命令的操作へ寄せ、seek/play の直前に必ず正しい音源が張られる順序を hook が保証すれば、controlled と命令再生のタイミング競合が構造的に消える。先行 Decision `2026-09-03T16-20-00` §6 がこれを「別 Issue」として保留していたのを、本 refactor の対象として解く。

2. `el.src` の getter は仕様上つねに絶対 URL へ解決される。`baseUrl` を空文字にして相対の音源 path（契約由来の `/episodes/...`）を渡す構成では、「今 `<audio>` が指している音源」を `el.src === 引数の path` で判定できない（絶対 vs 相対で毎回不一致）。不一致のたび `load()` すると、同じ episode 内の seek（topic sec bar 連打）でも再生が毎回中断する。hook が「最後に張った URL 文字列」をそのまま覚えて比較すれば、渡された値基準で正しく差分判定できる。

3. 「停止」を頭出し（`currentTime = 0`）+ `idle` 化にすると、停止後に同じ episode を再生したとき必ず頭から始まり、`2026-09-03T16-20-00` §1-2 が state に位置を保持させた意味（「停止中でも位置は残る」）が UI に出ない。`pause()` だけにして `active/paused` を維持し `positionSec` を残せば、既存の `play` の「同じ episode が `active` なら現在位置から」ロジックがそのまま続き再生になる。頭出しは別 episode へ切り替えるとき §1-2 の `load()` が担うので、「停止」に頭出しを持たせる必要がない。

4. topic の sec bar は「その章から聴く」ための操作で、押した時点が停止中か・別 episode を聴いている最中かに関わらず「そこから再生」が期待される挙動。`2026-09-03T16-20-00` §1-3 の「seek は移動だけ、停止中は停止のまま、別 episode は停止のまま」は seek bar のドラッグを想定した分岐だが、実際に seek を呼ぶ UI は topic bar だけで、そこでは「移動して止まったまま」が一度も要らない。分岐を残すと「別 episode の topic を押しても鳴らない」という bug に見える挙動（本 refactor 前に実在）を招く。seek を「そこから再生開始」に一本化し、`play()` を呼ぶか否かの分岐を seek 側から無くす。

5. `buildRequestUrl` の呼び出しが page にあると、page が「音源 path → 絶対 URL」という配線知識を持ち、`useEpisodeListPage` が `play(episodeId)` の外部 signature で受けた episodeId を hook 内部で catalog から `audioRef` へ引き当てる責務と分断される。URL 解決も hook に寄せれば、page は `baseUrl` を hook へ渡すだけの純粋な組み立てに戻り（先行 `2026-09-04T01-10-00` の「page は分岐と配置だけ」に沿う）、`<audio src>` を controlled で持たなくなった `AudioControls` も `audioRef`（ref）1 つ受けるだけになる。

6. 行の再生ボタンが `phase: "playing"` のときだけ「停止」表示だと、`play` 直後の `phase: "loading"`（まだ `playing` event が来ていない、実 browser では常に一定時間続く）の間ボタンが「再生」のままで、ユーザーは再生が始まったか分からず、loading 中の音声を止められない。「再生進行中（`loading` または `playing`）」を停止トグルの基準にすれば loading 中も停止でき、`"paused"` は「再生」に戻って続き再生の入口になる。「今音が出ている」視覚強調は `"playing"` 限定の別 flag で担い、2 つの関心を混ぜない。

7. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。

## 3. Rejected

1. `<audio src>` を controlled のまま維持し、`play` / `seek` を `useEffect` で `playback.audioRef` 変化後に発火させる — hook に「src 変化を待って play する」effect が増え、「play を押したら遅れて鳴る」タイミングを制御する追加のフラグ（この play 要求は effect で消化済みか）が要る。命令的経路に一本化するより複雑で、`2026-09-03T16-20-00` §1-4 が構造で消した「state と副作用の二重管理」を別の形で戻す。

2. `moveTo` の src 差分判定を `el.src` と、渡された path を絶対 URL 化した値の比較で行う — `baseUrl` が空のとき「現在ページ URL + path」を組む必要があり、`buildRequestUrl` とは別の URL 正規化ロジックが hook に入る。hook が「最後に張った文字列」を覚えれば正規化なしで足り、比較対象が「実際に `setAudioSource` へ渡した値」と一致するので判定がぶれない。

3. `stop()` は `idle` に戻し、直近の episode 位置だけ別の ref に退避して、次に同じ episode を `play` したらそこから — `PlaybackState` の外に「停止した episode とその位置」という影の state を持つことになり、`kind` union が「再生対象がある/ない」を一意に表す性質（`2026-09-03T16-20-00` §1-1）が崩れる。`active/paused` を維持すれば union の中で完結する。

4. seek の 3 分岐（`2026-09-03T16-20-00` §1-3）を残し、別 episode の topic bar だけ「切替して再生」にする — 「同じ episode 停止中の topic bar は移動して止まったまま」という分岐が残るが、その挙動を要求する UI が無い。使われない分岐を型と実装に残すのは、`illegal states unrepresentable` の逆で「使わない正当な状態」を保守し続けることになる。seek を一本化して分岐ごと消す。

5. 行の再生ボタンを `phase` 不問（`active` かつその episodeId なら「停止」）にする — `"paused"`（停止中）でボタンが「停止」表示のまま「もう一度押すと続き再生」になり、「停止」を押して再生が始まるのは Least Astonishment に反する。停止中は「再生」表示に戻し、押下＝続き再生、が素直。
