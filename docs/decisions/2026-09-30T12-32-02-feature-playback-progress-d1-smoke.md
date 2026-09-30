---
name: 進捗D1の一過性疎通は Worker が使う binding 面で write+read を往復し、表が無ければ dispatch 内で migration file を適用する
date: 2026-09-30T12:32:02
branch: feature/playback-progress-d1-smoke
---

## 1. Decision

1. 疎通は、Worker が使う D1 binding（`EPISODE_PROGRESS`）を `getPlatformProxy` の remote binding で取り、adapter が呼ぶ最小面（`prepare` / `bind` / `first` / `run`）で write → read → delete の往復を1回通す。`wrangler d1 execute` の CLI 実行だけで済ませない。
2. TEST D1 に進捗表が無いときは、同じ dispatch 内で repo の migration file を適用してから往復する。表の DDL を dispatch 側（yml・script・probe）へ複製しない。
3. test 固定・PASS 後削除・merge 意味を見ない点は、R2 疎通の Decision（`2026-09-16T10-45-14-feature-r2-smoke-migrate-cutover` / `2026-09-16T13-46-41-feature-r2-smoke-remove-and-migrate`）と同じ答えとし、本 file へ写さない。

## 2. Reason

1. CLI は SQL を直接流すだけで、adapter が実際に呼ぶ面（先行 Decision `2026-09-23T08-06-58-feature-playback-progress` §1-2 が A で固定した `prepare` / `bind` / `first` / `run`）を通らない。CLI が緑でも、binding 経由の到達や面の不一致は見逃す。同じ面で往復すれば、面と結線を1回で確かめられる。playback smoke の R2 も remote binding 経由で疎通しており、同じ型に揃う。
2. 疎通の目的は生 I/O 経路の確認である（R2 Decision と同じ）。表が無いことは経路の故障ではなく、schema が未適用という環境状態である。表なしで落として手で適用し再 dispatch する運用は、一過性 artifact の目的を二度手間にする。
3. DDL の正本は migration file 1本である。dispatch 側が DDL を持つと第二の SSOT になり、列変更時に片方だけ更新される。file をそのまま実行する形なら、本番へ適用するときも同じ file を使える。

## 3. Rejected

1. **CLI（`wrangler d1 execute --remote`）だけで疎通を見る案** — adapter が呼ぶ面を通らず、結線ミスが緑で通る。
2. **表を Dashboard 等で手作業適用してから dispatch する案** — 適用内容が repo の migration file と乖離しうり、適用の記録も repo に残らない。
3. **probe が `CREATE TABLE IF NOT EXISTS` を自前で持つ案** — DDL が migration file と probe の2箇所になり、列契約が二重管理になる。
