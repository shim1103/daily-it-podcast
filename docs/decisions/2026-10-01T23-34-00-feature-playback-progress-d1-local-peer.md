---
name: 時刻の妥当性は層で分け、DBにCHECKを入れない。静的な範囲は契約が、現在時刻に依る判定はApplicationが持つ
date: 2026-10-01T23:34:00
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. DB の schema（migration）には、範囲などの domain 契約（`CHECK`）を入れない。
2. 静的な範囲は契約（zod）が持つ。時刻は instant として、1970-01-01T00:00:00Z より後、2100-01-01T00:00:00Z より前とする。値の正本は契約の schema である。
3. 現在時刻に依存する判定は Application が持つ。未来側の skew 上限と、遠い未来の `clientAt` による永久後勝ちの防止がこれに当たる。skew 上限の数値は別 Issue（#197）で決まるので、この Decision では決めない。

## 2. Reason

1. `CHECK` は後から変えにくい。SQLite では表の再作成が要る。既存データに範囲外の行があると、挿入し直す migration が失敗しうる。変えにくい層へ範囲を置くと、後で見直す時の代償が大きい。
2. 契約の schema は純粋であるべきで、現在時刻との比較は呼ぶたびに結果が変わる非決定な判定になる（`design-philosophy.md` §4-4 参照透過性）。時計が要る判定は、時計を持てる Application が持つ。
3. 静的な範囲を契約に置くのは、固定幅 ISO の前提（拡張年表記を避ける）を入口で守るためである。保存形は `2026-10-01T18-54-16-feature-playback-progress-d1-local-peer.md` の UTC 固定幅に従う。
4. 後勝ちの `lastPlayedAt` は遅い `clientAt` が勝つ。遠い未来の `clientAt` を受けると、その行は以後の正当な書込に永久に勝つ。契約の静的範囲では、2100 年より手前の未来を防げない。現在時刻に対する上限は、Application が持つ。

## 3. Rejected

1. **DB の `CHECK` に範囲を持つ案** — 変えにくく、既存データで migration が失敗しうる。範囲の変更が表の再作成を伴う。
2. **契約の schema に現在時刻との比較を持つ案** — schema が非決定になり、同じ入力が呼ぶ時刻で通ったり落ちたりする。時計を持たない層に時計が入る。
3. **上限を設けず、遠い未来の `clientAt` を受け入れる案** — 後勝ちの列が、遠い未来の1書込で永久に固定される。
