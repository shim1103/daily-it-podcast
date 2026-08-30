---
name: ManuscriptDraft の尺は秒を正本にし文字数を CharsPerSecond の畳み込みで導出、合計は朗読 field のみ
date: 2026-08-30T03:06:53
branch: feature/generator-manuscript-draft-limits-seconds-model
---

## 1. Decision

1. 朗読 field（`intro` / `closingSummary` / 各 topic の `preface` / `detail`）と全体尺の許容範囲の正本は **秒**とする。`entities/constants/manuscript_draft_seconds.go` が `Draft*Sec` 系と `CharsPerSecond` を持つ。文字数（`manuscript_draft_limits.go` の `Draft*Len` 系）は `秒 × CharsPerSecond` の const 畳み込みで導出する。値の一覧は両 file を正本とし、本 Decision へ再掲しない。
2. `application/build` は「秒」を知らない。`validateTotalChars` を含む parse は文字数定数（`Draft*Len` / `DraftTotalChars*`）だけを見る。秒→文字の変換は Entities 内（const 畳み込み）で完結する。
3. **全体文字数の合計対象は `intro + closingSummary + Σ_topics(preface + detail)`** とする。`title` と各 `topic.title` は朗読されない見出しなので合計に入れない。`openingGreeting` / `closingFarewell` も別工程で付与するため合計に入れない。
4. `title` / `topic.title` は見出しとして **非空 + 日本語含有 + rune 数 range** のみ検証する。**末尾句点は課さない**。これは先行 Decision（`2026-08-29T14-11-00-docs-produce-episode-run-spec-manuscript-draft-parse-domain-rules.md`）3 項「各朗読 field は末尾が `。`」の対象から `title` / `topic.title` を外す部分 supersede である。同 Decision の他項（機械判定可能なものはすべて parse で検証、正本は constants + models、Infra Error にしない、全体文字数 range は TTS 尺に直結）は維持する。
5. 定数群の整合（各 field を自分の range 内で動かしつつ全体尺も満たす draft が構築可能であること、および昇順・target 一致）は contract test（`manuscript_draft_limits_contract_test.go`）で固定する。検証する不変条件は「最小構成の合計 ≦ 全体上限」「最大構成の合計 ≧ 全体下限」「target 構成の合計 = 全体 target」「各 min/target/max が昇順」。畳み込み式そのものの再掲は自己参照 assert のため書かない。
6. `closingSummary` の文字数定数（`DraftClosing*`）を新設する。これは先行 Decision（`2026-08-29T17-00-00-docs-produce-episode-run-spec-brief-prompt-field-limits-merge.md`）Rejected 案 3「数値は将来 limits 追加時に placeholder を足す」の「将来」の到来であり、同 Decision の 1 本化方針と矛盾しない。
7. brief prompt の全体尺 guideline に分数（`朗読でおよそ N〜M 分`）を併記する。`N` / `M` は `DraftTotalMinSec / 60` / `DraftTotalMaxSec / 60` を `embedManuscriptDraftLimits` が埋める placeholder（`{{TOTAL_MINUTES_MIN}}` / `{{TOTAL_MINUTES_MAX}}`）とし、prompt 文字列にハードコードしない。

## 2. Reason

1. 尺の設計は「何分の episode か」で人が考える。文字数は TTS の発話速度（`CharsPerSecond`）を介した派生値でしかない。正本を文字数にすると、発話速度を変えたとき全 field 定数を手で計算し直すことになり、かつ「元々何秒を意図したか」が失われる。正本を秒にして文字数を畳み込みで導出すれば、`CharsPerSecond` の 1 箇所変更で全 field が追随し、意図（秒）が残る。
2. 秒→文字の変換を `application/build` に置くと Use Cases が「秒」という Entities の詳細を知ることになり Ring の依存方向（内→外のみ）に反する。Go の const は整数演算を畳むので Entities 内で `Draft*Sec * CharsPerSecond` を評価でき、build は文字数だけを見れば済む。層をまたぐ変換ロジックが不要になる。
3. `title` / `topic.title` は画面や一覧に出す見出しであって朗読されない。TTS 入力にも WAV 尺にも寄与しないため、全体文字数（先行 Decision `14-11` で TTS 尺に直結すると定めた値）に数えるのは誤り。見出しを合計に混ぜると、本文を削って見出しを長くしても尺が変わらないのに合計だけ増える、という歪みが出る。
4. 見出しに文末の句点「。」を強制するのは日本語の見出しとして不自然（新聞・雑誌の見出しに句点は付かない）。一方、日本語番組なので見出しも日本語であることと、空でないこと、長さの範囲は担保したい。よって朗読 field 用の検証（`checkNarrationBasics`）から句点チェックを外した `checkHeadingBasics` を分ける。
5. 「各 field の min/max」と「全体尺の min/max」は独立した制約で、valid な draft は両方を同時に満たす。両者が数学的に両立するか（各 field range 内で全体尺 range も満たせる組み合わせが存在するか）は定数を変えた瞬間に崩れうる不変条件であり、`CharsPerSecond` や field 秒数を誰かが調整したときに気づけるよう contract test で固定する。畳み込み式の再掲（`DraftIntroMinLen == DraftIntroMinSec * CharsPerSecond`）は同語反復で検出力を持たないため、検証するのは定数どうしの関係だけにする。
6. 分数を prompt にハードコードすると `DraftTotalMinSec` / `DraftTotalMaxSec` と二重管理になり、尺を変えたとき prompt の文言だけ古くなる。placeholder 化して定数から埋めれば SSOT が秒定数に一本化される（先行 Decision `17-00` の「数値 placeholder は `manuscript_draft_limits.go` が SSOT、ComposeBrief は定数を埋めるだけ」と同じ方針）。

## 3. Rejected

1. 文字数を正本にし秒を持たない案 — `CharsPerSecond` 変更時に全 field 定数の手計算が必要で、かつ「元々何秒を意図したか」が記録に残らない。発話速度は TTS エンジンや読み上げ速度設定で変わる前提なので、派生値を正本にすべきでない。
2. 秒→文字の変換を `application/build` の関数で行う案 — Use Cases が「秒」を知り Ring の依存方向に反する。parse の SSOT が「constants の秒数 + build の変換式」に分裂する。const 畳み込みで Entities 内に閉じれば変換関数自体が不要。
3. `title` / `topic.title` も全体文字数に含める案 — 見出しは朗読されず TTS 尺に寄与しない。合計に入れると本文と見出しのトレードオフが尺に反映されない歪みが出る。先行 Decision `14-11` の「全体文字数 range は TTS 尺に直結」と整合しない。
4. `title` にも末尾句点を課したまま合計からだけ外す案 — 合計対象かどうかと文末表現の要否は別問題。日本語の見出しに句点は不自然で、Cursor に不自然な出力を強制することになる。
5. 畳み込み式そのものを contract test で assert する案（`DraftIntroMinLen == DraftIntroMinSec * CharsPerSecond` 等）— 実装の定義式を test にコピーするだけの自己参照 assert で、bug も一緒にコピーされ検出力を持たない。値の正しさはその値を消費する `validateTotalChars` の unit test が担い、contract test は定数どうしの数学的関係だけを見る。
6. 全体尺の min/max を各 field min/max の単純合計に一致させる案 — 「各 field は範囲内だが全体としてこれくらい」という融通（どれか短め・どれか長めでも全体尺に収まればよい）が効かなくなる。両者は独立の窓とし、両立可能性だけを contract test で保証する。
