---
name: 進捗のmergeはApplicationの純関数が持ちD1 adapterは条件付き書込に徹する。競合はUseCaseが再read→再mergeで解き、時刻は契約境界でUTC固定幅ISOへ正規化する
date: 2026-10-01T18:54:16
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 先勝ち・後勝ちの merge は、Application の**純関数**が持つ。置き場と、merge・冪等・行なし 404 を Application に置く理由は `2026-10-01T18-54-14-feature-playback-progress-application.md` が正で、ここへ写さない。規則の意味は `2026-09-22T19-13-35-feature-playback-progress.md` が正。D1 adapter は merge 済みの行を、条件付きの 1 文でそのまま保存する薄い永続とする。
2. 2端末が同じ行を同時に書く**競合**は、server 側で自動解決する。
   1. UseCase が `read → merge（純関数）→ 条件付き書込（期待値＝読んだ時の版）` を行う。競合（条件が満たされず、書込が 0 行）なら、再 read → 再 merge を最大 3 試行まで行う（回数の正本は定数）。尽きたら UseCase は Domain Error `ProgressWriteConflictError` を throw する。Controller の `mapInternalErrorToExternal` の既定経路（"other"）が `UnavailableError`（503）へ写し、browser の既存の retry に委ねる。HTTP の公開 error code は増やさない（`2026-09-22T18-58-38-feature-playback-progress.md`）。
   2. merge の結果が既存の行と同じなら、書込を省略する（`seq` を無駄に進めない）。
   3. 新規行の期待値は null とし、`INSERT … ON CONFLICT DO NOTHING` の changes=0 を競合として扱う。
3. adapter は、条件付きの 1 文を実行して `written` か `conflict` を返すだけで、再試行しない。D1 の一時失敗も、従来どおり adapter では再試行せず、browser の retry に任せる（`2026-09-22T19-29-15-feature-playback-progress.md`）。`seq` の採番は adapter の SQL が同じ文の中で行い、UseCase は版を不透明な値として受け渡すだけとする（`seq` の意味は `2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md`）。
4. 時刻列は、**契約境界**でUTC固定幅のISO表記へ正規化する。Application の文字列比較は、正規化済みの値を前提とし、adapter は変換しない。pull の更新印は時刻でなく `seq` で、時刻の比較には使わない（`2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md`）。応答の時刻もUTCの `Z` 表記になる。契約が許す入力の形（offset 付き）は狭めない。範囲の検証は `2026-10-01T23-34-00-feature-playback-progress-d1-local-peer.md` が正。

## 2. Reason

1. 勝ち側を決める規則の意味は `2026-09-22T19-13-35-feature-playback-progress.md` の1箇所が正である。merge を D1 内の1文 upsert に置くと、同じ規則が SQL と TS（In-Memory の Fake）の2箇所に実装され、片方だけの修正で食い違う（`design-philosophy.md` §2-2 DRY）。純関数に一本化すれば、実装は1つで、永続は結果を書くだけになる。
2. SQL の merge は、実 SQLite を刺す Narrow Integration でしか検証できない。Application の純関数なら Port double だけで勝敗を検証でき、#197 の「Port double だけで閉じる」と整合する。旧答えは、Port の `@ensure`（勝ち側の `first*` を返す）の形が A の時点で永続側に勝敗を寄せていたが、#196・#197 は merge 本文を永続側でなく Application 側に置く契約だった。この食い違いを、永続を薄く戻して解く。
3. read-then-write の全列置換では lost update が起きる。2端末が同じ行を読んでから書くと、後から書いた側が先の書込を上書きし、`first*`（先勝ち）は恒久的に失われうる。書込が成功した端末は再送しないので、失われた値は復元されない。条件付き書込（期待値＝読んだ時の版）にすれば、版が変わっていた書込は 0 行で弾かれ、再 read → 再 merge で取り込み直せる。merge を SQL へ戻さずに、原子性が取れる。
4. 再試行は「再 read → 再 merge → 再書込」で、merge の純関数を持つのは Application だけである（`2026-10-01T18-54-14-feature-playback-progress-application.md` の境界）。だから競合の再試行は UseCase が持ち、adapter は条件付き 1 文の結果（`written` か `conflict`）を返すだけにする。adapter が再 merge すると、merge が永続層に入る。
5. 競合は一時失敗でなく、論理的な衝突である。server が即座に解決でき、ネットワークの往復も要らない。これは `2026-09-22T19-29-15-feature-playback-progress.md` が避ける「adapter と browser の再試行が掛け算になる」話とは、失敗のクラスが別である。あちらは D1 の一時失敗に対する再試行で、その予算は browser が持ち、adapter は再試行しない（不変）。競合の再試行は UseCase が持ち、上限を使い切った時だけ、同じ 503 で browser の予算に戻す。
6. 条件付き書込が原子的であることは、docs と実測で確かめた。D1 は各データベースで単一スレッドでクエリを 1 つずつ処理し、Sessions API を使わない限り全クエリは primary で実行される。ローカルの実 SQLite（miniflare D1）では、期待値が一致すれば changes=1、不一致なら 0（書込 0）で、同じ期待値の条件付き UPDATE を 10 本同時に走らせると勝者はちょうど 1 本だった。本番 D1 での同実測は未確認である。
7. merge の結果が既存の行と同じなら書込を省くのは、`seq` が書込のたびに進む（`2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md`）ためである。実質の変更が無い書込で `seq` を進めると、他端末の pull に変更の無い行が現れ、書込量も無駄に増える。
8. 新規行の期待値を null にするのは、行が無い状態も「読んだ時の版」の一つとして同じ型で扱えるからである。2端末が同時に新規を作ると、片方の `ON CONFLICT DO NOTHING` が changes=0 になる。これを競合として扱えば、新規と更新で競合の扱いが分かれない。
9. 時刻を文字列で比べる箇所は Application の merge である（pull の差分印は時刻でなく `seq`）。文字列の大小が時系列と一致する固定幅が前提になる。比較は adapter より前（Application）で走るので、正規化を adapter に置くと、merge は未正規化の値を比べることになる。契約境界で1回だけ正規化すれば、write の入力・保存・応答・比較が同じ形で揃う。列ごと・層ごとに形が異なると、形を意識する箇所が増える。応答が常に `Z` 表記である現状の挙動も、merge の置き場や正規化の置き場の変更を理由に変えない。
10. 競合で試行が尽きた時の失敗を Domain Error にするのは、Application が throw してよいのは Domain Error のみ（`architecture/backend/application` §6）だからである。External の型（`UnavailableError`）は Controller 層の写像が持つ。尽きた時の失敗は、既存の写像の既定経路（"other"）で 503 になるので、写像に分岐を足さず、HTTP の公開 error code も増えない。
11. 入口で `+09:00` 等を拒むと、契約が許す入力を狭める。offset 付きの入力を受け、同じ instant の UTC 表記へ寄せれば、契約が許す入力を変えずに形だけを一致させられる。

## 3. Rejected

1. **D1 内の1文 upsert（`INSERT ... ON CONFLICT DO UPDATE ... RETURNING`）で merge し、単行原子性に任せる案**（旧答え） — merge が SQL と TS（In-Memory Fake）に二重化する。実 SQLite の Narrow Integration でしか検証できず、#197 の「Port double だけで閉じる」と整合しない。Port の `@ensure` が永続側に勝敗を寄せる形で、#196・#197 の契約（merge 本文は Application 側）とずれる。競合を SQL 1 文の merge で解く案としても同じ理由で採らない（`2026-10-01T18-54-14-feature-playback-progress-application.md` で却下済み）。原子性は、merge を SQL に戻さず、条件付き書込で別に取る。
2. **merge を entities の Domain 純関数に置く案**（旧答え） — merge・冪等・行なし 404 を持つのは Application という境界（`2026-09-26T03-48-06-feature-playback-progress.md`、`2026-10-01T18-54-14-feature-playback-progress-application.md`）と食い違う。規則の置き場が Application と entities に分かれ、同じ規則を探す場所が2つになる。
3. **SQLの時刻関数（`julianday` / `strftime`）で offset 混在を比較する案** — SQL に merge を置く前提の案で、merge を SQL の外に置いたため要らない。D1のSQLiteの版と関数の許可は未確認で、確認していないものへ依存を増やさない。
4. **adapter が Date で比較して merge する案** — merge が永続層に入り、規則の置き場が Application と永続に分かれる。永続を実 SQLite で刺さないと勝敗を検証できない点も、1 文 upsert の案と同じ問題を残す。
5. **offsetのまま保存する案** — merge の正しさには要らなくなったが、列の形が入力の書き方に依存して揺れ、Application の文字列比較の前提が崩れる。応答の時刻の形も入力で変わり、外から見える挙動の変更になる。
6. **正規化を adapter 内で行う案**（旧答え） — merge が Application の文字列比較になった今、比較が adapter の正規化より先に走り、merge が未正規化の値を比べる。契約境界の正規化に加えて adapter でも変換すると、同じ変換が2箇所になる（`design-philosophy.md` §2-2 DRY）。
7. **lost update を許容する案**（旧答え） — `first*` が先に勝った側を失うと、復元されない。窓が小さいという前提は、条件付き書込で原子性を取れると分かった今、失う根拠にならない。旧答えは、原子性を取り戻すには merge を SQL へ戻すしかないと見ていたが、期待値つきの条件付き書込で、merge を Application に置いたまま取れる。
8. **D1 の `batch` で、読んで計算して書くを包む案** — `batch` は SQL の transaction（全体で commit か rollback）だが、送信前に全文が確定している必要がある。読んだ結果をアプリで計算して書く流れは包めない。interactive な transaction は、確認した範囲の docs に記載が無い。
9. **競合を browser の retry だけに委ねる案** — server が即座に解決できる論理的な衝突を、client に見せる。往復と遅延が増える。
10. **adapter 内で再 read → 再 merge する案** — adapter は merge を持たない。持たせると、merge の置き場が Application と永続に分かれる。
11. **競合で試行が尽きた時に、UseCase が External の `UnavailableError`（503）を直接 throw する案**（旧答え） — Application が throw してよいのは Domain Error のみで、External の型を直接 throw すると層の境界を越える。External への写像は Controller の `mapInternalErrorToExternal` が持つ。
