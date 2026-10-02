---
name: 実instanceへ未適用の間、進捗表のmigrationは初期SQLを直接編集し、増分fileを作らない
date: 2026-10-02T15:44:40
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 進捗表が実 instance（本番 D1・TEST D1）へ適用される前の間は、初期 SQL（`apps/playback/worker/migrations/0001_episode_progress.sql`）を直接編集する。増分の migration file（`0002_…` など）は作らない。履歴の版管理は `git log` が持つ。
2. 実 instance へ適用された後は、この方針は当てはまらない。適用済みの file は変えず、変更は後続の migration とする。その後の履歴の持ち方は、Drizzle の生成物が担う（`2026-10-01T23-33-20-feature-playback-progress-d1-local-peer.md`）。
3. Drizzle へ移る時、現行の SQL を引き継ぐ baseline 扱いは不要とする。`drizzle-kit generate` が出す初期 migration を正本にして、作り直す。
4. migration file を恒久の置き場（`apps/playback/worker/migrations/`）に残す判断（`2026-09-30T13-30-00-feature-playback-progress-d1-smoke.md`）とは矛盾しない。0001 は残る。変えるのは、未適用の間に増分 file を足さないことだけである。

## 2. Reason

1. 増分 file が要るのは、適用済みの台帳（`d1_migrations`）と file 名の整合を保つ相手がいる時である。D1 は適用済みの file を名前で記録するので、適用後に中身を変えると食い違う。実 instance にまだ進捗表が無い今は、その相手が存在しない。
2. 版管理は `git log` で足りる。増分 file は、未適用の間は履歴の複製にしかならない。
3. 増分 file を持つと、複数 file を昇順に流す適用 helper（`;` での分割など）が要る。Drizzle の移行では、`drizzle-kit generate` の生成物が正本になり、増分 file と helper は捨てる仮物になる。捨てる物を増やさない方が単純である（`design-philosophy.md` §2-3）。
4. Drizzle の初期 migration を正本にするので、現行 SQL を baseline として引き継ぐ必要は無い。引き継ぐ対象の適用済み instance が無く、baseline を作る手間と、生成物との二重管理だけが残る。
5. 適用後に当てはまらなくなるのは、台帳が file 名を記録するからである。その時点から、適用済みの file は変えない通常の運用になる。

## 3. Rejected

1. **増分 `0002` を足す案** — 採らない。`seq` の列と索引の追加は 0001 に畳んだ。台帳と整合を取る相手が無く、増分 file とそれを流す適用 helper が、Drizzle 移行で捨てる仮物を増やす。
2. **適用済みを想定し、台帳を保つ運用を今から敷く案** — 適用済みの instance が無い。存在しない相手のための規律で、未適用の間の変更を重くするだけである。
3. **Drizzle 移行で、現行 SQL を baseline として引き継ぐ案** — 引き継ぐ適用済み instance が無い。生成物を正本にして作り直す方が、二重管理を避けられる。
