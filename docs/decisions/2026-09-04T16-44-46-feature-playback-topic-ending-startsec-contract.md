---
name: 原稿 body の opening / closing は seek 用の開始秒を contract に持つ。frontend が 0 や総尺で補わない
date: 2026-09-04T16:44:46
branch: feature/playback-topic-ending-startsec-contract
---

## 1. Decision

1. `manuscript.schema.json` と `apps/playback/contracts/http.ts` の `body.opening` / `body.closing` を、topics[] と同型の「本文 + `startSec`」object にする。`opening` は `{ text, startSec }`、`closing` は `{ summary, startSec }`。3 種の bookend（opening / topics / closing）がすべて「本文 + 音声上の開始位置」を持つ形に揃える。
2. `closing.startSec` は generator が算定する。`build.Timeline` が末尾 segment（closingSummary+farewell 束）の開始累積秒 `closingStartSec` を返し、`MarshalManuscript` が `closing.startSec` へ書く。`opening.startSec` は先頭 segment なので定義上つねに 0 で、`MarshalManuscript` が 0 を直書きする（Timeline を経由しない）。
3. playback web の `EpisodeManuscript` は「導入」「まとめ」の seek bar とラベルを `body.opening.startSec` / `body.closing.startSec` から描く。frontend が `0` や `durationSec` を補って seek 先を決めることをやめる。
4. `opening` / `closing` 本文のキー名（`text` / `summary`）自体の推敲は本 Decision の対象外。ここで固定するのは「seek 用開始秒を contract 側に持つ」という所在の方針。

置き換え範囲: 先行 Decision `2026-08-25T05-10-48-feature-playback-ui-structure.md` §1-2 の「`episode-topic` の contract に `startSec` を含めない」と §1-5 の「seek 機能は scope に含めない」を、opening / closing についても本 Decision で置き換える（topics については先行 Decision `2026-08-29T14-10-00-docs-produce-episode-run-spec-tts-segment-split.md` §1-2 が既に `body.topics[].startSec` を確定済み）。先行 Decision `2026-09-02T13-55-00-feature-generator-system-e2e-produce-episode.md` §1-3 の「greeting+intro / summary+farewell は開始秒を持たない固定 segment」を、`closing`（summary+farewell 束）について置き換える。closing 束は `closingStartSec` を持ち、Timeline がそれを返す。

維持範囲: 先行 Decision `2026-08-29T14-10-00` §1-1 の TTS segment 分割（Opening = OpeningGreeting と Intro を別々、Closing = ClosingSummary と ClosingFarewell を別々）は維持する。本 Decision は manuscript JSON の形だけを変え、TTS へ渡す束ね方（`SpeechTexts`）は変えない。`2026-09-02T13-55-00` §1-2 の「`topicStartSecs[i]` は i 番目 topic 束の開始累積秒」、同 §1-1 の topic+2 束、`2026-08-29T14-15-00-docs-produce-episode-run-spec-wav-concat-segment-silence.md` §1-3 の「無音 insert 分を含めた累積尺を正とする」は維持する。`2026-09-04T01-40-00-feature-playback-web-ui-rewrite.md` の「原稿を opening / topics / closing に組む `EpisodeManuscript` → `EpisodeTopic` の構造」も維持する。

契約値（object の形・generator の算定経路）の正本は A artifact（`contracts/manuscript.schema.json`、`apps/playback/contracts/http.ts`、`apps/generator/internal/application/build/episode_assembly.go`）。本 Decision は所在の方針だけを固定し、形を写さない。

## 2. Reason

1. 先行 Decision `2026-08-25T05-10-48` が seek と `startSec` を scope 外に置いたのは、当時の goal が「1 page 統合と component 分解」で seek が要件になかったため（同 §2-5）。その後 `2026-08-29T14-10-00` が topics に `startSec` を入れ、seek は実装済みの機能になった。opening / closing だけ `startSec` を持たないと、`EpisodeManuscript` が「導入は 0」「まとめは総尺」という **contract に無い値を frontend が知っているつもりで補う**ことになる。同じ「seek 先の開始秒」という知識が、topics は contract、opening / closing は frontend のハードコードに二重化する（`design-philosophy.md` §2-2 DRY 違反）。3 bookend すべてを contract 側に寄せると、seek 先の SSOT が manuscript JSON に一意化する。
2. `closing.startSec` を generator が算定するのは、開始秒が「無音を含む累積尺」（`2026-08-29T14-15-00` §1-3）であり、その情報を持つのは WAV 尺を測る generator だけだから。frontend は総尺しか知らず、まとめが音声のどこから始まるかを算定できない。topics の `startSec` を generator が持つのと同じ理由（`2026-09-02T13-55-00` §3）。
3. `opening.startSec` を 0 直書きにして Timeline を経由しないのは、opening が必ず先頭 segment（先行する無音なし）で開始位置が定義上 0 に固定されるため。Timeline に「常に 0 を返す返り値」を足すと、呼び出し側が使わない定数を戻り値で運ぶことになり indirection のコストに見合わない（`design-philosophy.md` §2-3 KISS）。`MarshalManuscript` にコメントで「先頭 segment なので 0」と書けば意図は追える。
4. `2026-09-02T13-55-00` §1-3 が「summary+farewell は開始秒を持たない固定 segment」と書いたのは、束ね導入時に timeline の外形（topic 開始秒と総尺）が変わらないことを示すためで、closing 束の開始秒を将来も contract に載せないという禁止ではない。本 Decision は timeline の外形を変えず（topic 開始秒・総尺の算定は不変）、closing 束の開始秒という**既に Timeline のループが通過している値**を返り値に足すだけ。
5. opening / closing 本文のキー名を本 Decision の対象外にしたのは、それが「seek 用開始秒の所在」とは独立した軸だから。キー名（`text` / `summary` / `greeting` 等）は片方だけ後から変えられる選択で、`decisions.md` §2 の「軸が独立した選択を 1 file に束ねない」に従い分離する。

## 3. Rejected

1. opening / closing を `startSec` なしの文字列のまま残し、`EpisodeManuscript` が「導入は 0、まとめは `durationSec`」を補い続ける — 「まとめは総尺」は音声の実開始位置ではなく、seek すると常にエピソード最後尾へ飛ぶ。まとめ本文が始まる時刻へ seek できない。かつ seek 先の知識が topics（contract）と opening / closing（frontend 定数）に二重化し DRY に反する。
2. `opening.startSec` も `closing.startSec` も Timeline の返り値から取る（対称性のため Timeline に `openingStartSec` を足す） — opening の開始秒は定義上 0 で自由度がない。返り値に定数を足すと呼び出し側が使わない値を運ぶ。対称性は「contract の object 形が 3 bookend で揃う」ことで既に得られており、算定経路まで対称にする必要はない。
3. `closing` に `startSec` を持たせず、まとめの seek 先を「最後の topic の `startSec` + その topic の尺」と frontend で計算する — topic の尺は contract に無い（`startSec` の差分からは無音込みの尺しか出ず、topic 本文の実尺と一致しない）。まとめが音声のどこから始まるかは generator の WAV 尺情報がないと出ない。
4. `opening` / `closing` 本文のキー名をこの Decision で `greeting` / `summary` に確定する — キー名の軸は seek 用開始秒の所在とは独立で、別 issue で推敲する。ここで束ねると片方だけの supersede が効かなくなる（`decisions.md` §2）。

## 4. 後続 supersede

`feature/playback-e2e-redeploy-master` との merge で、key `closing` → `ending`、本文キー `summary` → `text` へ改名し、§1-4 で保留した本文キー名を確定した。`ending.text` は closingSummary + 定型締めの朗読全文（本 Decision の「summary に締めの挨拶を含めない」は破棄）。seek 用開始秒を bookend の contract 側に持つ方針・generator 算定経路（§1-2 / §1-3）は維持。詳細は `2026-09-04T19-30-00-feature-playback-e2e-redeploy-master.md`。
