---
name: dispatch 専用の新規 yml は実装完了後に、master 起点の一時 checkout へ cherry-pick して master へ載せる。契約確定前に先行させない
date: 2026-09-30T12:32:03
branch: feature/playback-progress-d1-smoke
---

## 1. Decision

1. `workflow_dispatch` 専用の新規 yml は、実装が current branch で完了し、引数・Secret・env・step が確定してから master へ載せる。
2. 載せ方は、master を起点にした一時 checkout（一時 worktree）へ、yml だけを current branch から cherry-pick し、そこから master へ merge する。current branch 全体は master へ向けない。
3. 実行は従来どおり `gh workflow run <yml> --ref <branch>` とする（`DEPLOY.md` §5）。

## 2. Reason

1. `--ref` は branch 側の yml を実行するため、master 側の yml は file 名を default branch へ通すためだけに要る。一方、引数・Secret・env・step は実装中に動くことが多い。契約の確定前に載せると、master 側の中身は実装完了時に陳腐化するか、変更のたびに master へ追加の変更が要る。確定後に1回載せれば、その往復が要らない。
2. cherry-pick の対象を yml 1 file に絞れば、疎通専用の実装（PASS 後に削除する probe・script）が master へ混ざらない。R2 疎通の yml 先行 PR（`4b84935`）が守っていた「配線・deploy 対象 code を master へ混ぜない」条件は、この形でも保たれる。
3. 一時 checkout を使うと、実装中の current branch の作業状態を動かさずに master 側の作業ができる。

## 3. Rejected

1. **yml を契約の確定前に master へ先行させる案**（R2 疎通の前例 `4b84935`） — 確定前の引数・Secret・env が master に残り、実装完了後に修正の変更が要る。
2. **current branch 全体を master へ PR する案** — 削除する前提の probe と script が master に混ざり、削除も master へ届ける必要が生じる。
