---
name: 進捗のmergeはentitiesのDomain純関数が持ちD1 adapterは保存に徹する。競合のlost updateは許容し、時刻はUTC固定幅ISOへ正規化して保存する
date: 2026-10-01T18:54:16
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 先勝ち・後勝ちの merge は、entities の **Domain 純関数**が持つ。規則の意味は `2026-09-22T19-13-35-feature-playback-progress.md` が正で、ここへ写さない。D1 adapter は merge 済みの行を保存する薄い永続に戻す。Port は「merge 済み行の保存」へ変わる。変更の実施は別 Issue が持ち、この Decision は選択だけを持つ。
2. 2端末が同じ行を read-then-write して起きる **lost update** は許容する。`first*` は先に勝った側が失われうる（一度失うと復元されない）。`positionSec` は次の push で回復する。
3. 時刻列は、保存前にUTC固定幅のISO表記へ正規化して保存する。応答の時刻もUTCの `Z` 表記になる。正規化は契約schemaを通った後の adapter 内で行い、契約が許す入力の形は変えない。

## 2. Reason

1. 勝ち側を決める規則の意味は `2026-09-22T19-13-35-feature-playback-progress.md` の1箇所が正である。merge を D1 内の1文 upsert に置くと、同じ規則が SQL と TS（In-Memory の Fake）の2箇所に実装され、片方だけの修正で食い違う（`design-philosophy.md` §2-2 DRY）。Domain 純関数に一本化すれば、実装は1つで、永続は結果を書くだけになる。
2. SQL の merge は、実 SQLite を刺す Narrow Integration でしか検証できない。Domain 純関数なら Port double だけで勝敗を検証でき、#197 の「Port double だけで閉じる」と整合する。旧答えは、Port の `@ensure`（勝ち側の `first*` を返す）の形が A の時点で永続側に勝敗を寄せていたが、#196・#197 は merge 本文を永続側でなく Application 側に置く契約だった。この食い違いを、永続を薄く戻して解く。
3. lost update を許容する根拠は2つある。単一ユーザー（Access 配下）で、2端末が同じ行を読んでから書くまでの窓が小さい。勝ち規則は `clientAt` の大小で決まる先勝ち（早い側）・後勝ち（遅い側）なので、書込の到着順が入れ替わっても結論は同じになる。原子性を取り戻すには merge を SQL へ戻すしかなく、Reason 1・2 を壊す。失う範囲は、`first*` の復元不能と `positionSec` の一時的な巻き戻りに限る。競合のための補償や再読込を持たない方針（`2026-09-22T19-23-39-feature-playback-progress.md`）とも同じ側に立つ。
4. merge を Domain に置くと、時刻の勝敗は SQL の文字列順に依存しなくなり、offset 混在は merge の正しさの条件でなくなった。それでも保存形は UTC 固定幅に揃える。SQL で TEXT 列を比べる箇所（pull の差分印）は、文字列の大小が時系列と一致する固定幅が前提になる。契約の静的範囲も、拡張年表記を避ける固定幅 ISO を前提に置く。列ごとに形が異なると、保存・応答・SQL 比較の三箇所で形を意識することになる。応答が常に `Z` 表記である現状の挙動も、merge の置き場の変更を理由に変えない。
5. 入口で `+09:00` 等を拒むと、契約が許す入力を狭める。契約を変えず、保存の形だけで一致させる。

## 3. Rejected

1. **D1 内の1文 upsert（`INSERT ... ON CONFLICT DO UPDATE ... RETURNING`）で merge し、単行原子性に任せる案**（旧答え） — merge が SQL と TS（In-Memory Fake）に二重化する。実 SQLite の Narrow Integration でしか検証できず、#197 の「Port double だけで閉じる」と整合しない。Port の `@ensure` が永続側に勝敗を寄せる形で、#196・#197 の契約（merge 本文は Application 側）とずれる。原子性で守れていた lost update は、単一ユーザーで窓が小さいため許容して手放す。
2. **SQLの時刻関数（`julianday` / `strftime`）で offset 混在を比較する案** — SQL に merge を置く前提の案で、merge を Domain に置いたため要らない。D1のSQLiteの版と関数の許可は未確認で、確認していないものへ依存を増やさない。
3. **read-then-write で、adapter が Date で比較して merge する案** — lost update を許容する点は採用案と同じだが、merge が永続層に入り、規則の置き場が Domain と永続に分かれる。永続を実 SQLite で刺さないと勝敗を検証できない点も、旧答えと同じ問題を残す。
4. **offsetのまま保存する案** — merge の正しさには要らなくなったが、列の形が入力の書き方に依存して揺れ、SQL で比べる差分印と形が分かれる。応答の時刻の形も入力で変わり、外から見える挙動の変更になる。
