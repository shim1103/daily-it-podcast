---
name: 進捗のWorker→D1はadapter内再試行せず補償rollbackせず、Workers Cacheにも載せない
date: 2026-09-22T19:29:15
branch: feature/playback-progress
---

## 1. Decision

1. **Worker→D1（Infrastructure）の retry**: adapter は D1 の一時失敗を **内側で回り直さない**。試行回数の正本は A（`PROGRESS_D1_ADAPTER_MAX_ATTEMPTS = 1`）。失敗は Infrastructure Error として上げ、既存写像どおり External `UnavailableError`（503）へ畳む。有限の再送予算は **browser→Worker の HTTP retry**（先行 Decision `2026-09-22T19-23-39` と contracts の `PROGRESS_WRITE_MAX_ATTEMPTS`）が持つ。
2. **Worker→D1 の rollback**: 補償操作は作らない。1文（1行 upsert / 読取）の成否に任せ、失敗時は行が変わらない。成功済みの別操作を打ち消さない（先行 Decision `2026-09-22T19-23-39` と同じ軸を Worker 内にも明示する）。
3. **Workers Cache（edge）**: 進捗を含む応答は **Workers Cache に載せない**。対象は progress embed 済み list、progress pull、progress Write 成功応答、および既存どおり error 応答。header の正本は A（`noStoreCacheHeaders` / `progressPullCacheHeaders` / `progressWriteCacheHeaders`）。音声の Workers Cache 方針（Decision `2026-09-16T19-45-00`）は進捗には適用しない。
4. 置き換え範囲: Decision `2026-09-16T19-45-00` §1-3「episode一覧は短い browser / edge cache」のうち、**progress を embed した list** については本 Decision と先行 `2026-09-19T19-12-01` が優先する（短い TTL ではなく no-store）。音声・progress 無しの仮想定は本 Decision の対象外。

## 2. Reason

1. browser が既に最大3試行する。adapter でも再試行すると積が最大で掛け算になり、D1 障害時に負荷と遅延だけが増える。予算の単一所有者は HTTP 境界の caller（browser）に置く方が、防御責務の所在がはっきりする。
2. D1 進捗は1行更新であり、Worker 内で「途中まで書いた複数資源」を補償で戻す対象が無い。単文原子性で足りる。
3. Workers Cache は hit 時に Worker と D1 をスキップする。進捗は高頻度に変わり、edge に短い TTL でも載せると他端末反映と no-store 方針が壊れる。音声のように原則不変な資源向けの edge cache を進捗へ流用しない。
4. `Cache-Control: no-store` だけでは実装・運用者が edge 専用 header と混同しやすい。edge 向けも明示的に no-store とし、Workers Cache を効かせない意図を契約値に残す。

## 3. Rejected

1. **D1 一時失敗を adapter 内で N 回 retry する案** — HTTP retry と二重化し、障害時の扇状負荷になる。
2. **Worker 内で失敗 Write の補償 DELETE を用意する案** — 単行更新に対して過剰。
3. **進捗 list / pull を短い edge TTL で Workers Cache する案** — D1 を見ない stale 進捗が残り、端末横断の正本と矛盾する。
4. **音声と同じ Cloudflare-CDN-Cache-Control 長期 TTL を list に残す案** — progress embed 後は不適（先行 no-store Decision と衝突）。
