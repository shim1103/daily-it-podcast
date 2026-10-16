---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. `agent-standards`の未commitの変更をcommitする（shim。`agent-standards`の`AGENTS.md`は「`/commit`を勝手に実行しない」）
    - Issue／worktree: `doc-terms/taxonomy`§3（titleは`{kind}({branch名}): {要約}`）、`doc-terms/issue.md`§0-9、`doc-terms/epic.md`§0-5、`3:workflow/worktree-setup`（branch名はIssueのtitleに決まっていればそれを使う）、`3:workflow/issue/SKILL.md`（default flowの書き方記述を外した）
    - 知識: 新規`coding-terms/boundary-trust.md`（知らない・保証・信頼、型で信頼を運ぶ、形式と不変の分担）と、`coding-terms/SKILL.md`の表、`backend-terms/{application,infrastructure,controller,route-handler}.md`、`frontend-terms/{api-client,view-model}.md`の差分
2. #218の本文を、実装に合わせて更新する（PRの前）
    - 本文は旧案の`upsertProgressIfVersion`（期待値nullで挿入と更新を兼ねる1面）のまま。実装は`insertProgressIfAbsent`と`replaceProgressIfVersion`の2面、cursorは`number`（Decision `2026-10-01T18-54-16`・`2026-10-01T23-32-40`）。AC 1・2・4とScope 2・3の文を直す
    - #219にも`upsertProgressIfVersion`の記述が1箇所ある。2026-10-08のdailyは「本文は更新済み」と書くが、実際の本文は更新されていなかった（訂正が要る）
3. push後のCI（`test-unit`・`test-integration`、head `fea9d5c`）の結果を確認する
4. PRの範囲を決める（shim）
    - #218のIn Scopeは、D1とIn-memoryの版つき面だけ。branchはepicとの差分が109 file・+3343/-912で、`domain/`の新設（web・worker・contracts）、音声の内容検査、名前の整理、Issueとskillの更新を含む
    - 案: 1つのPRでDeviationsに書く / 範囲外の変更を別PRへ分ける
