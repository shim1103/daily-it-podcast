---
name: R2 の NI gate peer は generator=experimental local S3、playback=getPlatformProxy。C1 が infra、列 5 が振る舞い
date: 2026-09-15T12:02:48
branch: feature/generator-r2-write-adapter
---

## 1. Decision

1. 本 file は先行 Decision `2026-09-14T19-29-08` の **test peer / gate 条**を置き換える（supersede）。本番口（generator=S3 互換 HTTPS、playback=Workers R2 binding）は先行 `2026-09-14T11-04-30` および `19-29-08` §1-1 を維持する。契約値は写さない。
2. **置き換え範囲（`19-29-08`）**: §1-3（NI 既定=`httptest`・local S3 は任意・gate 必須にしない）、および Rejected「NI 既定を wrangler に置換して `httptest` を捨てる案」の結論。**維持**: SU に wrangler を載せない、BI に wrangler local S3 を必須化しない、generator に binding を持たせない、Lookup/Writer の同 credential・列 6 同着。
3. **generator Narrow Integration** の gate 正 peer は **wrangler experimental local S3** とする。必須対象は **Writer NI と Lookup NI**（Lookup 本実装完了後）。Verification は Issue 内で赤から緑にする（先出し専用 smoke Issue は作らない）。
4. **`httptest`（controllable TLS double）** は NI から捨てない。local S3 が gate 正、`httptest` は退路・高速代表として併存してよい。
5. local S3 必須化の CI 置き場は **`test-unit` に混ぜず、別 integration job / script** とする。
6. **playback** の local binding test 基盤は **`getPlatformProxy`（または同等の埋め込み platform proxy）** を正とする。`wrangler dev` 常時 process を gate 必須にしない。
7. Issue 分担: **`generator-r2-test-peer-scope`（C1）= peer / gate infra**。**列 5 `playback-r2-read-adapter`= Adapter 振る舞い本実装**。C1 は列 5 の list/get 契約本体を持たない。列 5 は C1 が固定した local binding 手段の上で振る舞う。
8. playback の Scope: **NI（必要なら Broad）は実 local binding**。**SU は Fake**（実 proxy 起動を SU に載せない）。

## 2. Reason

1. 「使いたい」が任意のままだと、本番口（S3 / binding）と test peer（常に `httptest` / Fake だけ）が永久にずれ、列 6 手前で初めて差が出る。gate 必須にすると差を Adapter Issue の完了条件へ前倒しできる。
2. `19-29-08` が任意に留めた主因は experimental の未実測だった。未実測を D に逃がし続けると導入が起きない。Issue 内で赤から通す（X1）なら、方針 SSOT を先に固定しつつ、通らなければ同じ Issue で失敗が見える。
3. `httptest` を捨てると、local S3 起動不能時に NI 全体が死ぬ。併存なら gate 正は本番寄り、退路は残る。
4. `test-unit` に wrangler を混ぜると FIRST（Fast / Repeatable）と coverage 入口が肥大する。別 integration job なら unit の速さを守れる。
5. playback の本番口は binding であり S3 HTTP ではない。local も binding を刺さないと列 5 の NI が嘘になる。`getPlatformProxy` は別 process の `wrangler dev` より CI 埋め込み向きで、vitest-pool-workers 全面移行より既存構成への差分が小さい。
6. peer 起動と Adapter 振る舞いを 1 Issue に混ぜると、失敗時に infra か振る舞いが切り分けられない。C1=infra・列 5=behavior は Fault Isolation のため。
7. SU に実 proxy を載せると Adapter 分岐表の赤が runtime 起動失敗と区別しにくい。NI に実 binding を置く。

## 3. Rejected

1. **`19-29-08` のまま local S3 を任意・非 gate に固定し続ける案** — 導入意図と矛盾し、本番口との peer 差が残る。
2. **NI から `httptest` を捨てて local S3 のみにする案** — experimental / CI 起動の単一障害点になる。
3. **local S3 / local binding を `test-unit` 必須に混ぜる案** — unit の Fast を壊す。
4. **playback local binding の gate 手段を `wrangler dev` 常時 process にする案** — 起動・port・flake が gate 向きでない。
5. **列 5 に peer/gate infra まで全部載せる案** — Adapter 振る舞いと起動基盤の失敗が混ざる。
6. **C1 に列 5 の EpisodeRepository 振る舞い本実装まで載せる案** — 実施順列 5 の達成契約と二重になる。
7. **BI に wrangler local S3 を必須化する案** — 先行 `19-29-08` 維持。BI は合成関係の所有であり NI I/O の再 assert になる。
8. **先に専用 smoke だけして Decision を後回しにする案** — 方針 SSOT が無く Issue AC がぶれる。実測は C1 Verification に内包する。
