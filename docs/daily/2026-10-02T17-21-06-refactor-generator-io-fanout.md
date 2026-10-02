---
name: generatorのI/O fan-outとComposition責務分けを#201として閉じる
date: 2026-10-02T17:21:06
session_id: none
branch: refactor/generator-io-fanout
prev: 2026-10-01T16-29-05-refactor-generator-wav-duration.md
---

## 1. Summary

#201 として `apps/generator` の I/O-bound 5 箇所を `errgroup` fan-out した。続けて ItemSource composite を Composition に残す責務分けを Decision・code・architecture skill に揃え、`pairStemCandidates` はメモリ分類のままと固定した。`~/settings` の architecture skill 更新は non-scope のため未 commit。

## 2. Changes

1. Issue AC の 5 fan-out（composite / HN story / HN comment / Lobsters detail / HasPair GET）と `x/sync` direct・depguard 許可を実装した。local hook（unit・integration）は緑。
2. composite を Application へ移す誤読をやめ、Composition / Application fallback / Adapter 内部の 3 軸で分ける形へ Decision `07-17-48`・`item_source`・`produce_episode` invariant を直した。
3. `pairStemCandidates` は GET でないため逐次のまま（Decision `07-47-56`）。
4. architecture skill（composition-root / application / infrastructure）は `~/settings/agent-standards` と worktree `.cursor` へ反映済み。settings 側は commit しない。
5. PR は #216。base は親 epic `refactor/generator-go-performance`（`lifecycle/release` の sub-feature→epic）。

### Commits

1. `9b6d5b0`
2. `5653350`
3. `43b9bf9`
4. `ca5343a`
5. `04c4a25`
6. `d9e595a`
7. `718262f`
8. `751d6a8`
9. `c81f749`
10. `e978dad`
11. `71a4f15`
12. `cc2d256`
13. `0d1adb6`
14. `1245642`
15. `f4d042e`
