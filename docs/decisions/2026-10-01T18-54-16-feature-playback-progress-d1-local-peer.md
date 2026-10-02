---
name: 進捗のmergeはApplicationの純関数が持ちD1 adapterは保存に徹する。競合のlost updateは許容し、時刻は契約境界でUTC固定幅ISOへ正規化する
date: 2026-10-01T18:54:16
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 先勝ち・後勝ちの merge は、Application の**純関数**が持つ。置き場と、merge・冪等・行なし 404 を Application に置く理由は `2026-10-01T18-54-14-feature-playback-progress-application.md` が正で、ここへ写さない。規則の意味は `2026-09-22T19-13-35-feature-playback-progress.md` が正。D1 adapter は merge 済みの行をそのまま保存する薄い永続とする。
2. 2端末が同じ行を read-then-write して起きる **lost update** は許容する。`first*` は先に勝った側が失われうる（一度失うと復元されない）。`positionSec` は次の push で回復する。
3. 時刻列は、**契約境界**でUTC固定幅のISO表記へ正規化する。Application の文字列比較と adapter の SQL 比較は、正規化済みの値を前提とし、adapter は変換しない。応答の時刻もUTCの `Z` 表記になる。契約が許す入力の形（offset 付き）は狭めない。範囲の検証は `2026-10-01T23-34-00-feature-playback-progress-d1-local-peer.md` が正。

## 2. Reason

1. 勝ち側を決める規則の意味は `2026-09-22T19-13-35-feature-playback-progress.md` の1箇所が正である。merge を D1 内の1文 upsert に置くと、同じ規則が SQL と TS（In-Memory の Fake）の2箇所に実装され、片方だけの修正で食い違う（`design-philosophy.md` §2-2 DRY）。純関数に一本化すれば、実装は1つで、永続は結果を書くだけになる。
2. SQL の merge は、実 SQLite を刺す Narrow Integration でしか検証できない。Application の純関数なら Port double だけで勝敗を検証でき、#197 の「Port double だけで閉じる」と整合する。旧答えは、Port の `@ensure`（勝ち側の `first*` を返す）の形が A の時点で永続側に勝敗を寄せていたが、#196・#197 は merge 本文を永続側でなく Application 側に置く契約だった。この食い違いを、永続を薄く戻して解く。
3. lost update を許容する根拠は2つある。単一ユーザー（Access 配下）で、2端末が同じ行を読んでから書くまでの窓が小さい。勝ち規則は `clientAt` の大小で決まる先勝ち（早い側）・後勝ち（遅い側）なので、書込の到着順が入れ替わっても結論は同じになる。原子性を取り戻すには merge を SQL へ戻すしかなく、Reason 1・2 を壊す。失う範囲は、`first*` の復元不能と `positionSec` の一時的な巻き戻りに限る。Application の read→merge→upsert が read-then-write である点とも矛盾しない。競合のための補償や再読込を持たない方針（`2026-09-22T19-23-39-feature-playback-progress.md`）とも同じ側に立つ。
4. 時刻を文字列で比べる箇所は2つある。Application の merge と、SQL の pull 差分印である。どちらも、文字列の大小が時系列と一致する固定幅が前提になる。比較は adapter より前（Application）で走るので、正規化を adapter に置くと、merge は未正規化の値を比べることになる。契約境界で1回だけ正規化すれば、write の入力・保存・応答・両方の比較が同じ形で揃う。列ごと・層ごとに形が異なると、形を意識する箇所が増える。応答が常に `Z` 表記である現状の挙動も、merge の置き場や正規化の置き場の変更を理由に変えない。
5. 入口で `+09:00` 等を拒むと、契約が許す入力を狭める。offset 付きの入力を受け、同じ instant の UTC 表記へ寄せれば、契約が許す入力を変えずに形だけを一致させられる。

## 3. Rejected

1. **D1 内の1文 upsert（`INSERT ... ON CONFLICT DO UPDATE ... RETURNING`）で merge し、単行原子性に任せる案**（旧答え） — merge が SQL と TS（In-Memory Fake）に二重化する。実 SQLite の Narrow Integration でしか検証できず、#197 の「Port double だけで閉じる」と整合しない。Port の `@ensure` が永続側に勝敗を寄せる形で、#196・#197 の契約（merge 本文は Application 側）とずれる。原子性で守れていた lost update は、単一ユーザーで窓が小さいため許容して手放す。
2. **merge を entities の Domain 純関数に置く案**（旧答え） — merge・冪等・行なし 404 を持つのは Application という境界（`2026-09-26T03-48-06-feature-playback-progress.md`、`2026-10-01T18-54-14-feature-playback-progress-application.md`）と食い違う。規則の置き場が Application と entities に分かれ、同じ規則を探す場所が2つになる。
3. **SQLの時刻関数（`julianday` / `strftime`）で offset 混在を比較する案** — SQL に merge を置く前提の案で、merge を SQL の外に置いたため要らない。D1のSQLiteの版と関数の許可は未確認で、確認していないものへ依存を増やさない。
4. **read-then-write で、adapter が Date で比較して merge する案** — lost update を許容する点は採用案と同じだが、merge が永続層に入り、規則の置き場が Application と永続に分かれる。永続を実 SQLite で刺さないと勝敗を検証できない点も、旧答えと同じ問題を残す。
5. **offsetのまま保存する案** — merge の正しさには要らなくなったが、列の形が入力の書き方に依存して揺れ、SQL で比べる差分印と形が分かれる。応答の時刻の形も入力で変わり、外から見える挙動の変更になる。
6. **正規化を adapter 内で行う案**（旧答え） — merge が Application の文字列比較になった今、比較が adapter の正規化より先に走り、merge が未正規化の値を比べる。契約境界の正規化に加えて adapter でも変換すると、同じ変換が2箇所になる（`design-philosophy.md` §2-2 DRY）。
