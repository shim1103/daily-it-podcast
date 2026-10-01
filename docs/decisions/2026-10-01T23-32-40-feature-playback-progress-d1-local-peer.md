---
name: pullの差分はclient時刻でなく、serverが書込の到着時に付ける`updated_at`で取る
date: 2026-10-01T23:32:40
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. pull の差分印は、client 時刻（`clientAt` 由来の `last_played_at`）ではなく、server が書込の到着時に付ける `updated_at` とする。cursor は server 発行の印を使う。
2. 契約の形（cursor を返すか、`updatedAt` を返すか）は未決で、この Decision では決めない。
3. `clientAt` を使う判断（`2026-09-22T19-13-35-feature-playback-progress.md`）は merge の鍵であり、pull の更新印とは別の問いとして扱う。

## 2. Reason

1. client 時刻を基準にすると、offline や retry で遅れて届く書込（`clientAt` が、他端末が持つ cursor より古い）が、次の pull の条件に当たらない。他端末は、その書込を取りこぼす。到着時刻なら、cursor を取った後に届いた書込は、それより後の印を持つ（server 時計の揺れは Reason 4）。
2. 後勝ちの `lastPlayedAt` が動かず、`first_completed_at` だけが動く変更は、client 時刻の列に差分として現れない。他端末は完走の事実を取りこぼす。到着時刻なら、どの列が動いても行が更新された印が進む。
3. `clientAt` は「操作が起きた順」を表し、merge に必要な時刻である。pull の更新印が答えるのは「この cursor 以降に、server が何を受けたか」で、別の問いである。記録上、pull の更新印を client 時刻にした理由は無く、更新印の列が表に無かった現状の帰結だった。
4. 差分印を SQL で比べる時の保存形は、`2026-10-01T18-54-16-feature-playback-progress-d1-local-peer.md` の UTC 固定幅に従う。同一時刻に届いた書込の扱い、server 時計の揺れの扱いは未確認で、契約の形と一緒に決める。

## 3. Rejected

1. **全件 pull（`since` を廃す）案** — 応答の payload と D1 の読取が、行数 × ポーリング回数に比例して増える。shim が却下した。ポーリング間隔・D1 の読取上限の数値は未確認で、ここへは載せない。
2. **client 時刻（`last_played_at`）を差分印にする案**（現行） — 遅れて届く書込と、`first_completed_at` だけが動く変更を、他端末が取りこぼす。
