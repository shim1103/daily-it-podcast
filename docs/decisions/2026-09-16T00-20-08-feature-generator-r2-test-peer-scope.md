---
name: generator R2 Adapter NI の正 peer は httptest。local S3 は peer 到達 Verification のみ
date: 2026-09-16T00:20:08
branch: feature/generator-r2-test-peer-scope
---

## 1. Decision

1. 本 file は先行 Decision `2026-09-15T12-02-48` の **generator Adapter NI gate peer 条**を置き換える（supersede）。本番口（generator=S3 互換 HTTPS、playback=Workers R2 binding）は `2026-09-14T11-04-30` を維持する。契約値は写さない。
2. **置き換え範囲（`12-02-48`）**: §1-3（Writer/Lookup NI の gate 正 peer = wrangler experimental local S3）。**維持**: `httptest` を NI から捨てない、local S3 / wrangler を `test-unit` に混ぜない、SU に wrangler 必須起動を載せない、BI に wrangler local S3 を必須化しない、generator に binding を持たせない。playback local binding infra の正手段は `getPlatformProxy`（`12-02-48` §1-6・§1-8）を維持する。
3. **generator R2 Adapter**（`EpisodeWriter` / `CompletedEpisodeLookup`）の Narrow Integration における振る舞い・error 行列の正 peer は **`httptest`（controllable TLS double）** とする。
4. **experimental local S3** の Verification は **peer 到達（HTTP）** に限る。本番 Adapter を local S3 に刺さない。起動失敗は skip せず見える化する。
5. local S3 peer の起動・注入契約の置き場は **`apps/generator/test/r2locals3/`** とする。本番 package（`internal/infrastructure/r2`）に local peer 用の endpoint override や build-tag seam を置かない。
6. local S3 peer Verification の CI 置き場は **`test-unit` に混ぜず、別 integration script**（Integration gate から呼ぶ）とする。

## 2. Reason

1. Adapter の retry / 4xx fail-fast / 成功系 path は、応答を制御できる peer でなければ手堅く固定できない。`httptest` はその制御を持つ。local S3 実 peer は注入が難しく、成功系偏重になりやすい。
2. Adapter 振る舞いと peer 起動を同じ NI に載せると、赤の原因が Adapter か wrangler か切り分けにくい（Fault Isolation）。
3. local S3 用に本番 Adapter へ endpoint override を足すと、Composition が使わない test seam が本番型に残る。test 専用は `test/` 配下へ閉じる既存慣習（他 NI は httptest / Dial で刺す）と揃える。
4. peer 到達だけを別 Verification に残すと、experimental 起動不能を黙って緑にせず、列 6 前の peer 生存確認は保てる。

## 3. Rejected

1. **Writer/Lookup を local S3 に刺し、それを Adapter NI の gate 正にする案**（`12-02-48` §1-3）— error 行列の制御が弱く、本番 Adapter に test seam が要り、失敗時の切り分けが壊れる。
2. **本番 `internal/infrastructure/r2` に `//go:build r2locals3` の seam を撒く案** — test 専用が本番 package tree を汚染する。
3. **local S3 peer Verification 自体を無くし httptest のみにする案** — experimental peer の起動失敗が見えなくなる。到達確認の価値は残す。
4. **local S3 / `getPlatformProxy` を `test-unit` 必須に混ぜる案** — unit の Fast を壊す（先行維持）。
