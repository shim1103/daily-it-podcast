---
name: 進捗付き初回listはembedのno-store、session中の他端末反映はprogress pullのno-storeのみ
date: 2026-09-19T19:12:01
branch: docs/playback-audio-history
---

## 1. Decision

1. 初回の一覧表示は、先行Decision（`2026-09-19T16-26-00-docs-playback-audio-history.md`）どおり **embed 1 Get** で catalog と進捗を同時に返す。
2. その embed 応答、および session 中の進捗 **pull** 応答は、HTTP cache させない（`Cache-Control: no-store` 相当。具体header文字列の正本は cache 政策の実装／既存 `cache-policy` に置き、ここへ写さない）。
3. SPA が mount したあとの session 中は、catalog を再取得しない（現行どおり list は mount 時のみ）。他端末の進捗反映は **progress 専用 pull** だけとする。
4. 置き換え範囲: 先行Decision（`2026-09-19T16-26-00-docs-playback-audio-history.md`）§1-3「進捗専用の list 向け Get を初手では分けない」のうち、**初回一覧の代替として分けない**趣旨は維持する。**session 中の増分／差分 pull として進捗専用 Get を持つ**ことは本 Decision が許す（初回 embed を廃して常時2 Get にする話ではない）。

## 2. Reason

1. embed 応答に進捗が載る以上、既存 list 向けの短い TTL（browser 約1分・edge 約5分）に乗せると、フル reload の cache hit で D1 を見ず古い進捗が残る。catalog は日次更新でも進捗は高頻度に変わるため、寿命をcatalog側に合わせられない。
2. 音声の長い edge TTL（日／週）と list の短い TTL は、更新頻度が違う資源を分けた結果である。進捗はさらに短い（実質貯めない）側に置く。
3. list が mount 一回しか走らない前提では、session 中に catalog だけを TTL 付きで再取得する第三の Get は今の UI 寿命に対して余分である。他端末反映に必要なのは進捗だけなので pull に閉じる。
4. 初回を catalog + progress の2 Get 完了待ちにすると、遅い方までの待ち・片方失敗時の分岐・合流配線が増える。mount 一回の初回表示では embed 1本の方が要件（同一 timing）に対して短い。

## 3. Rejected

1. **embed 応答を既存 list の短い TTL のままにする案** — reload 直後に古い進捗が返り、端末横断の正本（D1）を読んだことにならない。
2. **初回から catalog Get と progress Get を常時分離する案** — cache はきれいだが、今の mount 一回＋同一 timing 要件に対して待ちと失敗分岐が余分である。
3. **session 中に catalog 再取得 API を用意する案** — 現行 UI が list を mount 以外で呼ばない前提と合わず、今は不要である。
4. **進捗 pull にも list 並みの max-age を付ける案** — 他端末の stop 結果が見えるまでの遅延が固定され、pull の目的と反する。
