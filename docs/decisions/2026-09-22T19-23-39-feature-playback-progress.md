---
name: 進捗Writeは単行原子性に任せ補償rollbackせず、retryは一時不能のみ有限回
date: 2026-09-22T19:23:39
branch: feature/playback-progress
---

## 1. Decision

1. 進捗 Write の **rollback（補償操作）は作らない**。D1 への1行 upsert / complete が失敗したら、その試行の変更は入らない（成功前の行が残る）。成功済みの別 Write や再生状態を打ち消す DELETE / 逆更新はしない。
2. client 側も、失敗した Write のためにローカル再生位置や既に成功反映した共有進捗UIを **巻き戻さない**。共有進捗UIは成功後だけ更新する（先行 Decision `2026-09-19T19-12-30-feature-playback-progress.md`）。
3. **retry** の回数上限と、HTTP 契約 code として再試行してよい語彙の正本は A（`PROGRESS_WRITE_MAX_ATTEMPTS` / `progressWriteRetryableHttpErrorCodes`）とする。本 Decision は「どの失敗を回すか」の方針だけを固定する。
   1. 契約 code **`unavailable`（一時不能）** は有限回 retry する
   2. **`validation_error` / `episode_not_found` / `configuration_error`** は retry しない
   3. web だけの輸送失敗（例: `network_error`）は、契約 code ではないが **unavailable と同様に有限回 retry** してよい（上限は同じ A 定数）
4. 有限回を使い切ったら捨て、session またぎキューは持たない（先行 Decision `2026-09-19T19-12-30`）。

## 2. Reason

1. 進捗永続は1 episode=1行の単一更新であり、複数資源にまたがる「途中まで書いた」状態を補償で戻す対象が無い。失敗時ゼロ副作用の範囲は「その1文が成功するまで行が変わらない」で足り、事後rollback経路を足すと rollback 失敗という新しい失敗面が増える。
2. 再生は主価値で進捗同期は補助である。失敗時に再生カーソルや成功済みUIを戻すと、聴取体験が同期の成否に引きずられる。
3. 400 / 404 系は同じ body を再送しても成功にならない。503（unavailable）と輸送断だけが時間で変わりうるので、retry 対象をそこに限ると無駄な再送とログ噪声が減る。
4. 回数を Decision に書くと A の定数と二重になる。有限である理由は先行 Decision、数と code 集合は A。

## 3. Rejected

1. **失敗した Write の補償 DELETE / 逆 PATCH を用意する案** — 単行更新に対して過剰で、補償自体の失敗経路が増える。
2. **失敗時にローカル再生位置をサーバ／直前成功値へ巻き戻す案** — 主操作を同期に結合する。
3. **validation / not_found も同じ回数 retry する案** — 結果が変わらず遅延とログだけが増える。
4. **retry 回数を Decision 本文の正本にする案** — 契約値の二重化。A を正とする。
