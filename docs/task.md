---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. `EpisodeManuscript`型と`manuscript.schema.json`の食い違い検出を決める
    - 現状はajv後の`as`でつなぐだけ。Goはembedしたschemaで検証し、組立structは手書き。TSは手書き型とHTTPの`episodeItemSchema`が写し（HTTP側は層境界のDecisionで意図的に別）
    - shimの判断: fixtureをschemaで検証するcontract test（案i）は大変なのでやめる
    - 残る案: ii=型の生成（要tool導入と版固定。未導入）、iii=`z.fromJSONSchema`。実測: 検証は動く（空id・余分なfield・負のdurationを拒否）が、出力型は静的に推論されない
    - iiとzodの関係: 生成するのは`EpisodeManuscript`型だけで、zodはHTTP契約とdomainの葉schemaに残る。JSON Schemaの`minLength`は型にならないので、`EpisodeId`のbrandは生成型から出ない（manuscript境界で`episodeIdSchema.parse`する）
2. `agent-standards`の未commitの変更をcommitする（shim。`agent-standards`の`AGENTS.md`は「`/commit`を勝手に実行しない」）
    - Issue／worktree: `doc-terms/taxonomy`§3（titleは`{kind}({branch名}): {要約}`）、`doc-terms/issue.md`§0-9、`doc-terms/epic.md`§0-5、`3:workflow/worktree-setup`（branch名はIssueのtitleに決まっていればそれを使う）、`3:workflow/issue/SKILL.md`（default flowの書き方記述を外した）
    - 知識: 新規`coding-terms/boundary-trust.md`（知らない・保証・信頼、型で信頼を運ぶ、形式と不変の分担）と、`coding-terms/SKILL.md`の表、`backend-terms/{application,infrastructure,controller,route-handler}.md`、`frontend-terms/{api-client,view-model}.md`の差分
3. 既存の子Issue（#195〜#200・#217〜#223）のtitleに、branch名を足すか決める
    - 新しい規約は、titleに`{kind}({branch名}): {要約}`を求める。既存のtitleは`feat(playback): …`か`playback: …`で、branch名を含まない
    - #193（epic）のtitleも`epic({epic branch名}): …`に揃えるか
