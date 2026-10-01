---
name: Access 付き remote e2e の失敗を切り分け、storageState を手で組み立てる運用で PASS させた
date: 2026-10-01T09:30:38
session_id: 5e60167a-d898-4f60-ab70-6695f6b66416
branch: feature/playback-e2e-session-recovery
prev: 2026-09-29T18-18-32-develop.md
---

## 1. Summary

`playback-e2e` が Access の login 画面で落ち続けていた。期限確認の step と失敗時の画面の保存を足して切り分け、原因が JSON の `domain` に付けた `https://` だったと確かめて、PASS させた。途中で、JWT から storageState を組み立てる登録 script と CLI を作ったが不安定だったため、人が JSON を組み立てて登録する運用へ替えた。

## 2. Changes

1. 実 run の経過（すべて `--ref feature/playback-e2e-session-recovery` の dispatch）。

| run | 結果 | 観測 |
|---|---|---|
| 36692642623 | failure | 一覧が出ない。切り分けの手段がまだ無かった |
| 36696634336 | failure | 期限確認が「失効」で止まった。`expires` が雛形の `0`（1970 年）のまま登録されていた |
| 36700825917 | failure | 期限確認は有効（残り 30 日）。E2E は Access の login 画面で落ちた。JSON の `domain` に `https://` が付いていた |
| 36702518853 | success | `domain` を host のみに直して再登録した。期限確認は有効、E2E は 1 passed |

2. 期限確認は `domain` を見ない。そのため、`domain` の誤りは期限確認を通り、E2E の login 画面でしか見えなかった。誤入力を最初の step で止める案（`domain` の検査、または `normalize` での host への変換）は提示した段階で、未実装である（shim の指示待ち）。
3. 登録 script と CLI の組み立て（JWT の `exp` から `expires` を作る経路）は、shim が不安定と判断して削除した。残したのは、期限確認（`check`）と、`expires` の日時を Unix 秒へ直して file へ書く `normalize` だけである。
4. 雛形 `storage-state.example.json` の `expires` を数値の `0` から文字列の placeholder へ直したのは、`0` が置き換え忘れに見えなかったため。DevTools の日時（例 `2026-10-30T16:44:30.650Z`）は、そのまま貼れる。
5. 検証: `check-static`・`test-unit`・`test-integration` が終了 status 0。unit は 426 test 通過、branch coverage は 100%。`actionlint` と `shellcheck` は手元に無く、workflow の yml は実 run で確かめた。
6. biome の `--write` が、無関係な file の import 順を変えたことがあった。commit の前に元へ戻した。
7. 同じ session で、agent-standards へ skill `3:workflow/transient-smoke` を作った。最初の版は shim が commit した（`9157e9c`）。
8. shim の指摘で、疎通の知識を `lifecycle/smoke.md` へ移し、skill を手順だけに書き直した。ymlだけを base branch 起点の tmp-branch に置き、実装は current-branch に置く。yml の commit は、base への merge 後に `rebase --onto` で載せ替えて取り込む。base は取り込まない。
9. PR #210（`closes #195`、base は epic）は merge されたが、#195 は OPEN のままだった。GitHub の `closes` は、既定 branch への merge でしか Issue を閉じない。shim の指示で、sub-feature の達成契約 Issue は shim が手動で close する規則を `lifecycle/release.md` §7 へ書いた（#195 は未 close）。
10. agent-standards の commit は 4 件（`39464a5` `70f0e3d` `198e74a` `847c394`）。Decision は 2 件で、疎通の配置と同期、Issue の close。既存の `2026-09-22T18-48-51-main.md` が `base branch` を role の呼称として退けているため、`smoke.md` では限定した意味でだけ使った。
11. label=`lesson` を作り、Issue #211 を起こした。`docs/lessons/index.md` の 30 行と、今回の 9 行を移し、file を消した。
12. wiki #192 の「e2e が Access session の失効で赤」の行を、`domain` 誤入力の検知が未実装であることの行へ置き換えた。
13. `develop` 向けの PR を、この後に作る。

### Commits

- daily-it-podcast: `dc4eb2d` `04034c1` `938ebae` `36c82d1` `f599e3c` `58bb2b0` `208194e` `da8b8a6` `b777aa7` `8ddf0b8` `5ae30c8` `9a3801e` `8a4d182` `5fd10b1` `c06718d` `8730dfb` `f9511c8` `4426065` `d02de59` `7ec5581` `971e0ba` `d0cba4b` `684f866`
- agent-standards: `39464a5` `70f0e3d` `198e74a` `847c394`
