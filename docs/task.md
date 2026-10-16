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
2. `Uint8Array`が残る2箇所（`routes/audio-response.ts`、`r2-episode-repository.ts`の`readBytes`）の方針を決める
    - shimの問い（外部の byte をどこで保証し`EpisodeAudio`にするか、routes はどちらを要求するか、r2 adapter が infra 内で保証するか）に、非editで答えた。答えを見て決める
    - 私の見立て: routes は`Uint8Array`（platform の型。entities を import できない）、保証は I/O 成功まで infra、内容の妥当性は application（`verifyManuscript`と同じ型）
3. `agent-standards`に入れたskillの編集をcommitする（shim。`agent-standards`の`AGENTS.md`は「`/commit`を勝手に実行しない」）
    - 未commit: `3:workflow/shape`→`3:workflow/split`のrenameと参照の置換、`coding-terms/split.md`§1-6（使われない定数・error型もAとして置く。使用者が無くcoverageに影響するなら`TODO:`でコメントアウト）、`doc-terms/epic.md`（Sub-featuresの章を無くし、ACが子の一覧）、`3:workflow/issue/SKILL.md`（`--epic`から書き方を外し、merged済みの子Issueを閉じる手順を足した）、`doc-terms/commit.md` Granularity§10と`3:workflow/commit/SKILL.md`（indexへ合成した内容をstageしない）
4. `3:workflow/issue/SKILL.md`の default flow §2 と anti-pattern §3 にも、`doc-terms`が持つ書き方の記述が残る（依頼は`--epic`のみ）。外すか決める
