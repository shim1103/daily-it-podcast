---
name: 進捗の完走ゾーンはentities、D1 binding・表列・adapter再試行はinfrastructure/d1に置く
date: 2026-09-23T08:06:58
branch: docs/playback-audio-history
---

## 1. Decision

1. **完走ゾーン幅**（業務規則の秒数）だけを Entities の Domain 定数として置く。正本 path は A（entities constants）。
2. **D1 binding 名・表名・列名・adapter 内再試行回数**、および **D1 binding の最小 method 面**（`prepare` / `bind` / `first` / `all` / `run`）は **Infrastructure（`infrastructure/d1/`）** に置く。PlaybackEnv はそこを参照するだけにする。
3. `ProgressRepository`（Port）は D1 固有型を露出しない（先行 Port 契約を維持）。

## 2. Reason

1. backend Entities は「仕様で固定された業務値」を持ち、**runtime config や永続のどう書くか**を持たない。完走ゾーンは「いつ complete するか」の Domain 規則なので Entities。binding 名・SQL 列・D1 API 面は「どう永続するか」なので Infrastructure。
2. すべてを entities に同居させると、Domain が D1 / Workers を知っているように見え、層の変更理由が混ざる。R2 が binding 最小面を adapter 隣に切ったのと同じ軸で、D1 も opaque `object` ではなく **adapter が実際に呼ぶ面**を A で固定する（実装前の不確実性を減らす）。
3. Port に D1 を出さないことで、Application は vendor を知らず、差し替えと test double が Port 単位で閉じる。

## 3. Rejected

1. **進捗関連定数をすべて `entities/constants` に置く案** — runtime / 永続契約が Domain に漏れ、Entities の外部依存ゼロと衝突する。
2. **D1 binding を `object` opaque のまま Composition に置く案** — adapter が使う method が A で見えず、C が面をその場発明する。完成系に寄せる方針と逆。
3. **Port に D1 prepared statement 面を載せる案** — vendor が Application 境界を貫通し、R2 と同様の「Port は storage 非依存」と矛盾する。
