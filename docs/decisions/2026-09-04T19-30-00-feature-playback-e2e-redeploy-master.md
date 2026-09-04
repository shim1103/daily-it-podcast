---
name: manuscript body.opening / body.ending は { text, startSec } object で、text は定型込みの朗読全文とする
date: 2026-09-04T19:30:00
branch: feature/playback-e2e-redeploy-master
---

## 1. Decision

1. `contracts/manuscript.schema.json` と `apps/playback/contracts/http.ts` の `body.opening` / `body.ending` を、`topics[]` と同型の `{ text, startSec }` object にする。3 bookend（opening / topics / ending）がすべて「本文 + 音声上の開始秒」を持つ形に揃う。
2. `opening.text` は **定型挨拶 + intro**、`ending.text` は **closingSummary + 定型締め** をこの順で連結した朗読全文とする（`SpeechTexts` の先頭束 / 末尾束と同一文字列）。連結 delimiter は `SpeechTexts` の束境界と同じ改行 3 個（`"\n\n\n"`）。
3. `opening.startSec` は先頭 segment なので `MarshalManuscript` が 0 を直書きする（`build.Timeline` を経由しない）。`ending.startSec` は末尾束（closingSummary+farewell）の無音込み累積開始秒で、`build.Timeline` が `endingStartSec` として返し `MarshalManuscript` が書く。
4. bookend 本文のキー名は `text` とする（opening / ending で対称。`topics[].preface` / `detail` のような複数本文を持たないため）。

置き換え範囲:
- 先行 Decision `2026-09-04T16-00-00-feature-playback-e2e-redeploy-master.md` §1（`body` field 名 `opening` / `topics` / `ending`）は維持。同 §2-3（opening / ending は string）を「`{ text, startSec }` object。string は `.text`」へ、§5（連結 delimiter は改行 1 個）を「改行 3 個（`SpeechTexts` の束境界に一致。`2026-09-04T17-05-00` §1）」へ置き換える。「朗読全文を契約へ載せる」「定型を application 内に隠さない」方針は維持。
- 先行 Decision `2026-09-04T16-44-46-feature-playback-topic-ending-startsec-contract.md` の「seek 用開始秒を bookend の contract 側に持つ」方針・generator 算定経路（§1-2 / §1-3）は維持。同 §1-1 の key `closing` → `ending`、`summary` → `text` へ改名。同 §1-4（本文キー名は対象外）を本 Decision §1-4 で確定。同 Decision の「`closing.summary` に締めの挨拶を含めない」を破棄し、`ending.text` は定型締めを含む朗読全文とする（`16-00-00` の朗読 SSoT 方針が優先）。

維持範囲: `2026-08-29T14-10-00` §1-1 の TTS segment 分割、`2026-09-02T13-55-00` の topic+2 束・`topicStartSecs` の意味、`2026-08-29T14-15-00` §1-3 の無音込み累積尺、`2026-09-04T17-05-00` の束境界改行 3 個 / detail 改行最大 1 は変えない。契約値・generator 算定経路の正本は artifact（`contracts/manuscript.schema.json`・`apps/playback/contracts/http.ts`・`apps/generator/internal/application/build/episode_assembly.go`）。

## 2. Reason

1. `origin/develop`（PR #127, `16-44-46`）と本 branch（`16-00-00`）が、同じ `body.opening` / `body.ending` に対し**別方向の変更**を merge 前に持っていた。PR #127 は「seek 用 startSec を持つ object」、本 branch は「定型込み朗読全文の string」。どちらも「3 bookend を対称化する」「Drive JSON を朗読 SSoT にする」という上位方針の帰結であり、両立する。両立形は「object の形（`16-44-46`）で、本文は朗読全文（`16-00-00`）」の 1 つに定まる。片方を丸ごと採ると、seek 先の SSoT が frontend ハードコードへ戻る（`16-44-46` を捨てた場合）か、定型が契約から落ちて消費者が全文を再構成できない（`16-00-00` を捨てた場合）。
2. delimiter を改行 3 個へ揃えるのは、`opening.text` / `ending.text` が `SpeechTexts` の束と**同一文字列**である以上、束境界の規約（`17-05-00`）がそのまま契約側の連結規約になるため。`16-00-00` §5 が「改行 1 個」と書いたのは `17-05-00` より前で、束境界がまだ改行 1 個だった時点の記述。SSoT が 1 つなら delimiter も 1 つ。
3. `closing` → `ending` の改名は `16-00-00` §3（「`closing` は『まとめだけ』と誤読されやすい」）と同じ理由。締め定型を含む朗読全文なので `ending`。本文キーは `text` 単数で、opening / ending が構造的に同型になり、`manuscriptBookendJSON` の 1 型で両方を marshal できる（DRY）。

## 3. Rejected

1. `16-44-46` をそのまま採り、`ending`（旧 `closing`）は `{ summary, startSec }` で summary は closingSummary のみ、farewell は音声専用 — Drive JSON が朗読 SSoT でなくなる（`16-00-00` の破棄）。Playback が「まとめ」表示に farewell を出せず、音声と表示が乖離する。
2. `16-00-00` をそのまま採り、`body.opening` / `body.ending` は string のまま、seek 先は frontend が「導入 = 0」「まとめ = durationSec」で補う — seek 先の知識が topics（contract）と bookend（frontend 定数）に二重化（`16-44-46` §2-1 の DRY 違反）。かつ「まとめ = 総尺」は音声の実開始位置ではなく、seek すると常に最後尾へ飛ぶ。
3. bookend 本文キーを `opening.text` / `ending.summary` のように非対称にする — 同型 struct 1 つで marshal できず、schema / zod / Go struct が bookend ごとに別定義になる。`topics[]` が `preface` / `detail` と複数本文を持つのと違い、bookend は本文 1 つなので対称 `text` で足りる。
4. 新 Decision を書かず `16-00-00` / `16-44-46` の本文を両方書き換える — `decisions.md` §2-5（近い問いを分ける時、片方を正・他方は参照）に反する。reconciliation という 1 判断は 1 file にして、両先行 Decision はこの file を参照する形にする。
