---
name: 進捗WriteはclientAtで先勝ち／後勝ちmergeし、create・update・completeとplaying／stopped同期を分ける
date: 2026-09-22T19:13:35
branch: docs/playback-audio-history
---

## 1. Decision

1. 進捗行の merge 判定時刻は **client が送る `clientAt`** とする（server 受理時刻ではない）。契約の字段は A（`ProgressWriteRequestSchema` 等）を正とし、ここへ写さない。
2. 字段ごとの勝ち方は次とする。
   1. `firstPlayedAt` / `firstCompletedAt` — **常に先勝ち**（より早い `clientAt` 側を残す。既に値があっても、より早い正当な時刻なら置き換える）
   2. `positionSec` / `lastPlayedAt` — **後勝ち**（より遅い `clientAt` 側を採用）
3. HTTP 操作の意味は次とする。path / method / body の正本は A。
   1. **create（POST progress）** — 初回 play（行が無いときの開始）
   2. **update（PATCH progress）** — 途中更新および stop 時の位置同期。末尾で強制 stop（`positionSec = durationSec`）も update
   3. **complete（POST progress/complete）** — UI が **完走ゾーン**（`positionSec >= durationSec - PROGRESS_COMPLETE_ZONE_SEC`）に入ったとき。末尾 stop そのものではない
4. 端末間・session 内の同期方針は次とする。
   1. **playing 中** — push（create/update/complete）のみ。応答の `positionSec` は見ない／載せない（A の Write 応答形）。負けてもローカル seek をサーバ位置へ引き戻さない
   2. **stopped 中** — stop 時に push したうえで、定期 **pull** で他端末反映。stopped 中の seek だけは端末固有で共有しない
5. 重複 create の冪等 200・行なし update の 404・HTTP code を増やさないことは先行 Decision（`2026-09-22T18-58-38-docs-playback-audio-history.md`）に従う。

## 2. Reason

1. 複数端末では到着順と操作順がずれうる。server 時計だけだと「後から届いた早い初回」を正しく先勝ちできない。操作が起きた側の `clientAt` を merge 鍵にすると、先勝ち／後勝ちの意味が端末横断で一貫する。
2. `first*` は「いつ初めて触った／完走したか」の事実であり、遅い再送で上書きすると履歴が壊れる。より早い正当値だけが勝つ先勝ちが、一回系の意味に合う。`positionSec` / `lastPlayedAt` は「いま共有すべき最新カーソル」なので後勝ちが合う。
3. complete を「末尾 stop」に結びつけると、強制 stop の update と完走記録が同じ操作になり、完走前のゾーン突入を逃す。完走ゾーン定数（A）で complete を切り、末尾 stop は通常 update に残すと責務が分かれる。
4. playing 中に pull や応答位置で seek を引き戻すと、再生中のカーソルが他端末や負けた push に引っ張られ体験が壊れる。stopped だけ pull すれば、聞いていないあいだの他端末結果を取り込めば足りる。seek 単体の共有は要件に無く、push 契機を増やして噪声になる。

## 3. Rejected

1. **server 受理時刻だけで merge する案** — 遅延した早い初回を後勝ちで潰し、先勝ちの意味が立たない。
2. **first* を「一度入ったら絶対不変」にする案** — より早い正当な `clientAt` の遅延到着を捨てる。先勝ち（min）と矛盾する。
3. **complete を末尾 stop（`positionSec = durationSec`）専用にする案** — ゾーン突入と stop が潰れ、完走記録が stop タイミングに依存する。
4. **playing 中も pull して位置を合わせる案** — 再生中カーソルが他端末に引き戻され、主価値（聴取）を損なう。
5. **stopped seek もサーバへ共有する案** — 端末固有の覗き見位置まで共有対象になり、push が増え要件外。
