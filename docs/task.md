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
6. `ProgressRow`・`ProgressUpdatedEntry`を`domain/`へ昇格するか決める（旧task 2の一部。判断せずに落としていたので戻した）
    - 今は`worker/src/application/ports/progress-repository.ts`にある。`ProgressRow`は`{ episodeId } & EpisodeProgress`、`ProgressUpdatedEntry`は`{ episodeId, progress }`
    - 基準案: web・worker・contractsのうち2つ以上が使うか。`ProgressUpdatedEntry`は契約のpull応答の1件と同形
7. 旧task 4（retry定数）を、計画と違う手段で解決した点をshimが確認する
    - 計画は「試行回数はweb側のVM、retry可否の導出は`web/api`へ移す」。実際は使用者ゼロなので削除し、設計意図をDecisionに残した
    - 戻す案: web側へ定数を置く（使用者ゼロのまま）。このまま案: 実装時にVMへ置く
8. 旧task 3（brandの手書き）と旧task 1（写しの範囲）は、shimの答えが無いまま私が決めた。結論を確認する
    - brand: 現状維持（3つは別の層の別概念。共通helperは、routesとapplicationにdomainのimportを強いるだけ）
    - 写し: `toEpisodeProgressResponse`に共通化。`EpisodeManuscript`と`EpisodeItem`は統合しない
9. `Uint8Array`が残る2箇所（`routes/audio-response.ts`、`r2-episode-repository.ts`の`readBytes`）の方針を確認する
    - 指示欄に本文が無く、現状維持と解釈した。routesはentitiesをimportできず、`readBytes`はjsonと音声の共用
