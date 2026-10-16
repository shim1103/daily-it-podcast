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
2. epic branch名の確認
    - `origin/feature/playback-progress`と推定。`gh`がTLS errorで取れず未確認
3. 秒を表示する関数の引数の型を決める
    - `formatDurationMmSs(durationSec: DurationSec)`が、位置（`startSec`: `PositionSec`）の表示にも使われ、型の意味がずれる。別名なので型は通る
    - 案: 引数を秒の共通の語にする / 位置用と長さ用に分ける / 名前を見直す
4. `agent-standards`に入れたskillの編集をcommitする（shim。`agent-standards`の`AGENTS.md`は「`/commit`を勝手に実行しない」）
    - 編集済み・未commit: `coding-terms/ring-model.md`§4-7、`coding-terms/backend-terms/entities.md`§7、`coding-terms/backend-terms/SKILL.md`§3、`error-terms/defensive-design.md`§1-2・3
    - 同repoの`skills/2:platform/swift/layout.md`の未commit差分は別の作業のもの（触っていない）
    - 追加の候補: 「層の入出力の束は、domainの葉と層固有の都合から合成する」の原則を`dto-naming.md`か`application.md`に書く（依頼されたtask 6の範囲外）
5. Write の retry を実装する時、上限の定数をview-modelへ置く
    - 上限は初回を含め3試行。Decision `2026-09-22T19-23-39-feature-playback-progress.md`が設計意図の置き場。再試行してよい語彙は`catalogErrorPresentation`の`retryable`を使う
