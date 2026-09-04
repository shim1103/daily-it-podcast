---
name: playback web の再生 state は kind tag union にし、再生位置を state で保持して play/seek を「指定秒から鳴らす」「指定秒へ移動する」の 2 操作に分ける
date: 2026-09-03T16:20:00
branch: feature/playback-web-view-models
---

## 1. Decision

1. `PlaybackState` の判別子を boolean（`played`）から文字列 tag（`kind`）へ変える。
   1. `{ kind: "idle" } | { kind: "active"; episodeId; phase; positionSec; durationSec }`
   2. `idle`（再生対象なし）は付随データを持たない。`active` は再生対象があり、`phase` に関わらず現在位置（`positionSec`）と長さ（`durationSec`、metadata 未取得なら null）を必ず持つ
2. 再生位置を ViewModel state で保持する。`active` 枝の `positionSec` は再生中は `<audio>` の `timeupdate` で更新し、`seek` / `play` でも更新する。`durationSec` は `loadedmetadata` で埋める。停止中でも位置は state に残る。
3. `play` と `seek` を別操作にする。共有するのは「再生対象の切替」「位置の移動」「phase 購読の貼り直し」で、`<audio>.play()` を呼ぶか否かが違う。
   1. `play(episodeId, positionSec?)` — 指定秒から**鳴らす**。`positionSec` 省略時、同じ episode が `active` なら現在の `positionSec` から、そうでなければ 0 から。`<audio>.play()` を呼び phase は `loading`（以後 event で遷移）
   2. `seek(episodeId, positionSec)` — 指定秒へ**移動する**だけ。`<audio>.play()` は呼ばない。phase は「同じ episode が直前 `playing` だったら `playing` 継続、それ以外（違う episode / 同じ episode で停止中）は `paused`」
   3. seek の 3 分岐: 同じ episode で再生中 → 移動して再生継続。同じ episode で停止中 → 移動して停止のまま。違う episode → 切替して移動して停止
4. `use-episode-playback.ts` の「現在の再生対象を setState 非同期を跨いで読む」ための ref を、`string | null` の切り出し変数から `PlaybackState` と同型の ref へ変える。state と ref の更新を 1 つのラッパ関数に通し、片方だけ書くことを構造で防ぐ。
5. `lib/audio-element.ts` の命令的操作を上記に合わせる。
   1. `seekAudioElement(el, positionSec, opts: { play: boolean })` — `currentTime` を動かし、`opts.play` のときだけ `play()` する。`play()` の Promise を返す（rejection は呼び出し側が握る）
   2. `startAudioPlayback` は `seekAudioElement(el, 0, { play: true })` 相当へ吸収し、専用関数を残さない
   3. `<audio>` の状態購読を lifecycle event（phase）に加え `timeupdate`（位置）と `loadedmetadata`（長さ）へ広げる。ViewModel は 3 種のうち必要なものを受け取る
6. 「違う episode への `play` / `seek` で `<audio src>` を張り替える順序」は本 Decision の範囲外。現状は component が `playedEpisode.audioRef` から `<audio src>` を JSX で制御しており、ViewModel が `src` を持たないため、違う episode への seek/play を ViewModel 単体で完結できない。src を ViewModel が返す形へ寄せる判断は表示側配線の Issue で行う。本 Issue では違う episode 経路の state 遷移だけ実装し、audio 反映は「要素に正しい src が載っている前提」で書く。
7. 契約値（union の枝・関数 signature・`positionSec` の更新頻度）の正本は A artifact（`apps/playback/web/src/view-models/` と `lib/audio-element.ts`）。本 Decision は方針だけを固定し形を写さない。

置き換え範囲: 先行 Decision（`2026-09-03T13-40-00-feature-playback-web-view-models.md`）§1-4 の `PlaybackState` の phase 判別可能 union 化を、本 Decision §1（`kind` tag 化 + `positionSec` / `durationSec` 追加）で拡張する。`phase` を判別可能 union にする §1-4 の趣旨は維持し、`played: boolean` を `kind: "idle" | "active"` に、`active` 枝に位置情報を足す。先行 Decision（`2026-09-02T15-00-00-...-orthogonality.md`）§1-3（別 episode への `Play` で前 audio を stop し `currentTime` を reset）は維持し、`seek` にも同じ reset を適用する（違う episode への seek でも前 audio を reset）。selection と playback の直交（同 Decision §1-2）は維持。

## 2. Reason

1. `played` は過去分詞で「再生し終わった」と読める。実際は「再生対象 episode が選ばれている」で、`phase: "loading"`（まだ鳴っていない）も `phase: "ended"`（鳴り終わった）も `played: true` に入る。どちらの語義で読んでも矛盾に見え、型を読んだだけでは解けない（Least Astonishment の次点）。`kind: "idle" | "active"` なら両側に名前が付き、`CatalogStatus`（`status`）・`PageStatus`（`kind`）の文字列 tag と repo 内で揃う。boolean 判別子は「false 側が何なのか」を名前で表現できず、将来 3 状態目が要ると 2 boolean になって矛盾が復活する。`selected` は「選択なし」と素直に読めるので `SelectionState` は巻き込まない。
2. 「停止中に再生位置だけ動かす」UI（seek bar のドラッグ、章ボタンからの頭出しで再生はしない）を満たすには、再生状態と再生位置を分けて持つ必要がある。位置を state に持たないと、(a) 停止中に seek した後 `play` を押すと 0 秒に戻る、(b) `<audio>` 要素が未 mount の間に来た seek を保持できない、(c) seek bar が「今どこ」を描けない。`positionSec` を `active` 枝に phase 不問で必須にすることで、これらが型で保証される。`idle` 枝は再生対象がないので位置の概念がなく、持たせない（make illegal states unrepresentable）。
3. `play` と `seek` を 1 関数に統合すると「seek すると必ず再生が始まる」か「play で位置を指定できない」のどちらかになる。両方満たすには「`play()` を呼ぶか否か」だけが違う 2 操作として分けるのが素直。共通処理（切替・移動・購読貼り直し）は private helper に括る（`function-design.md` §4）。
4. 現状の `playedEpisodeIdRef: string | null` は `PlaybackState` から episodeId だけを切り出した別変数で、`play` / `stop` のたびに「ref と state の両方を更新」を手で書く必要があり、片方を忘れる bug が型検査を通る。`useMemo` では解決しない（`useMemo` は render 中の計算で、イベントハンドラ実行中に「今書いた値を今読む」ができるのは ref だけ）。ref を `PlaybackState` と同型にし、更新をラッパ関数 1 つに通せば、切り出しによる不整合が構造的に消える。
5. `seekAudioElement` が無条件で `play()` する現状では seek と play を分けられない。`{ play: boolean }` オプションで分岐させ、`startAudioPlayback` を吸収すれば、「位置移動 + 条件付き再生」という 1 つの命令的操作に統一できる。`timeupdate` / `loadedmetadata` の購読は `<audio>` の状態を state へ写すのに必須で、既存の `subscribeAudioPhase` と同じ「購読して解除関数を返す」形で足せる。`timeupdate` は約 4Hz で発火するが、seek bar を動かすのに必要な re-render なので `positionSec` の差分更新（functional setState）で許容する。高頻度が問題化してから ref + `useSyncExternalStore` への分離を検討する（YAGNI）。
6. 違う episode への seek/play は `<audio src>` の張り替えを伴い、現状の「component が `playedEpisode.audioRef` から src を JSX で制御する」設計だと、`playback.episodeId` 変化 → component 再 render → `<audio src>` 変化 → `currentTime` セット、の順序に依存する。`currentTime` セットが src 反映より先だと無効になる。これを ViewModel 単体で解くには src を ViewModel が返す形（`audioSrc: string | null` を戻り値に持ち、component は `<audio src={vm.audioSrc}>` するだけ）への変更が要り、これは表示側配線 Issue のスコープ。本 Issue では state 遷移だけ正しくし、audio 反映は「要素に正しい src が載っている前提」で書く。
7. 形を Decision 本文へ写すと A artifact と二重 SSOT になる（`decisions.md` §4-4）。

## 3. Rejected

1. `PlaybackState` の判別子を `isActive: boolean` にする — 矛盾排除（`isActive: false` 枝に `episodeId` なし）は効くが、false 側に名前が付かず、3 状態目が要ると 2 boolean に増えて矛盾が復活する。`kind` 文字列 tag なら 1 語増やすだけで、repo の他の状態型（`CatalogStatus` / `PageStatus`）とも揃う。
2. `positionSec` を `idle` 枝にも持たせる（`kind` 不問で必須）— 再生対象がない状態に位置の概念はない。持たせると「何も鳴らしていないのに位置が 42 秒」という無意味な state が表現できる。`active` 枝に閉じる。
3. `play` と `seek` を 1 関数（`playFrom(episodeId, startSec)`）に統合する — seek が必ず再生を開始してしまう。「停止中に位置だけ動かす」が表現できない。`play()` を呼ぶか否かを引数フラグで切ると boolean flag 引数（`function-design.md` §1 が避ける形）になる。2 操作に分けて共通処理を helper に括るのが筋。
4. `playedEpisodeIdRef` を `useMemo` に置き換える — `useMemo` は render 中に値を計算してキャッシュする仕組みで、イベントハンドラ実行中に「setState 予約した値を同じ関数内で読む」用途には使えない。ref だけが同期的な read-after-write を提供する。
5. `positionSec` の `timeupdate` 反映を本 Issue で見送り、seek 3 分岐だけ先に実装する — seek 3 分岐は `positionSec` を state に持てば `timeupdate` なしでも成立するが、その `positionSec` が停止中に seek した値のまま固定され、再生を再開しても更新されない half-baked な state になる。位置を state に載せると決めた以上、更新経路（`timeupdate`）も同じ Issue で入れる。
6. 違う episode への src 張り替えを本 Issue で ViewModel 管理（`audioSrc` を返す）へ変える — 現状の component が `<audio src>` を制御する設計を変えるのは表示側配線に波及し、本 Issue（ViewModel の state 設計）のスコープを超える。src 反映の順序問題は「表示側 Issue で解く」として TODO に残し、本 Issue では state 遷移だけ完成させる。
