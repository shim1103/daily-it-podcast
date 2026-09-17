---
name: TextWriter の invalid-draft retry を Adapter 内部へ移し、model 切り替え fallback を source 配列化する
date: 2026-09-16T11:41:26
branch: feature/generator-textwriter-adapter-fallback
---

## 1. Decision

1. `port.TextWriter.Write` の signature に、生 response を `models.ManuscriptDraft` へ解釈する関数（`buildFn`）を DI で追加する。Adapter は `application/build.ManuscriptDraftFromWriterOutput` を直接 import しない。
2. **invalid-draft retry（`TextWriterMaxAttempts` 回、brief へ rejection 文言を足す既存ロジック）を `application/produce_episode.go` から Adapter 実装（`infrastructure/manuscript/geminiapi` 等）内部へ移す。** `manuscript.TextWriter`（application 層の合成 layer）は model 切り替え fallback だけを持ち、invalid-draft retry を持たない。
3. `manuscript.TextWriter` の `primary, secondary port.TextWriter` という固定 2 引数を `sources []port.TextWriter` へ変える。fallback ループは 1 つで、何段の source でも同じコードで処理する。
4. invalid-draft retry を使い切ってもなお invalid なら、その状態を **`ErrSourceExhausted` とは別の番兵 error（`ErrDraftRejected`）** で表す。両者を暗黙に同一視しない。
5. `manuscript.TextWriter` は `errors.Is` で両番兵を区別し、次 source へ渡す brief を変える：`ErrSourceExhausted` なら素の brief、`ErrDraftRejected` なら前 source の最後の rejection 理由を brief へ織り込む（既存の `"\n\n# Previous attempt rejected\n" + err.Error() + ...` と同じ組み立て）。
6. `TextWriterMaxAttempts`（429 retry 回数とは別の、invalid-draft retry 回数）は現状値のまま据え置く。RPD に余裕があっても増やさない。

## 2. Reason

1. 現行実装は invalid-draft retry を `produce_episode.go`（Application 層）に置き、`manuscript.TextWriter.Write` を 1 回だけ呼んでいた。この構成だと、primary が invalid-draft retry を `TextWriterMaxAttempts` 回使い切って失敗した時点で全体が失敗し、**secondary への切り替えが一度も試されない**欠陥があった。invalid-draft retry を Adapter 内部へ移し、その使い切りを fallback 判定対象にすることで、この欠陥を解消する。
2. buildFn を DI で渡すのは、Adapter（Infrastructure 層）が `application/build` の Domain Rule 関数を直接 import すると Ring Model の依存方向（内→外のみ）に反するため。Adapter は「response を drafts へ解釈する関数」を受け取るだけで、その関数の中身（何が valid か）を知らない。
3. `ErrSourceExhausted`（枯渇）と `ErrDraftRejected`（invalid-draft使い切り）を同一の番兵で扱うと、`manuscript.TextWriter` が「素の brief で次へ渡すべきか、rejection 理由を織り込むべきか」を区別できない。2 つの異なる状態を 1 つの番兵に握り潰すのは、意味の異なる失敗を同一視する設計であり、区別が必要になった時点（今回）で型を分けるべきという判断（暗黙の握り潰しより明示的な別番兵を優先）。
4. `MaxAttempts`（429 用 retry 上限）を増やさない理由：既存の「同種 error（同じ Op）が 2 回連続したら打ち切り」ロジック（Decision `2026-09-02T13-56-00`、TTS 側で先に導入済み）が、RPD の余裕とは無関係に一過性と決定的な失敗を区別する。RPD に余裕があっても、同じ理由で reject され続けるなら retry を増やす意味が薄い。

## 3. Rejected

1. **invalid-draft retry を Application 層に残したまま、fallback だけ配列化する案** — primary の invalid-draft retry 使い切りが fallback 対象にならない欠陥が残る。
2. **`ErrSourceExhausted` と `ErrDraftRejected` を同一の番兵にまとめる案** — 枯渇と invalid-draft使い切りでは次 source へ渡す brief の扱いが異なり（素の brief か rejection 理由を織り込むか）、1 つの番兵では区別できない。
3. **`TextWriterMaxAttempts` を RPD の余裕に応じて増やす案** — 同種 error 2 連続打ち切りロジックが既にあるため、retry 回数を増やしても実効的には使い切らないことが多く、増やす実益が薄い。
