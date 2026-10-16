---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

0. 時刻のbrand型`UtcInstant`を、web・worker共有の新dir`apps/playback/domain/`へ置く（shim確定。実装中）
    - 計画は`plan.md`（scratchpad）。Opus 5.5のreview済み。Phase 1（domain・設定・contracts・application型）は完了。Phase 2はsub-agent 2本（worker application/controllers、infrastructure/integration）が実行中
    - 残り: Phase 3（diff確認・全gate）、Phase 4（`/decision`で新規1件+更新3件、`/commit --repo`）。skillsとの食い違い（`entities.md`§7、`ring-model`§4-7）は、skill更新の別taskにする
    - 範囲外: `MAX_ATTEMPTS`移設、manuscript fixture test、cursor・EpisodeIdのbrand、Controllerのfield copy削減、skillsの更新
    - 注意: `.codex/config.toml`はshimの未commit編集。stageしない
1. playback/workerの同形の型・写しを削る範囲をshimと決める
    - 調査済み（editなし）。候補は`ProgressMergeIncoming`の再構築3箇所の削除、progress→contract写像の1関数化、`ProgressWriteUseCaseOutput`を`Pick<EpisodeProgress, ...>`へ寄せること
    - `EpisodeManuscript`と`EpisodeItem`の統合は`application.md`§7-2との関係で慎重に扱う。shimの許可待ち
2. `contracts/http.ts`の「domain語彙」と「HTTP形」を分け、語彙をEntitiesへ寄せるか決める
    - shim案: Entitiesにhttp非依存の語彙を置き、contractsがimportしてHTTP契約を宣言し、UseCaseはEntitiesからinterfaceを作る
    - Decision `2026-08-25T18-42-00`のRejected 7（契約全体をEntitiesへ移す案）とは別の問い。語彙だけを分ける案は未判断で、Decisionの追記が要る
    - 先に決めること: contracts→entitiesのimportを許すか（depcruiseに禁止rule無し。skillにも記述無し）、時刻のUTC前提をbrand型にするか
    - 実際の分離候補: `episodeProgressSchema`、`progressInstantSchema`、cursor形式。`PROGRESS_WRITE_MAX_ATTEMPTS`はclientの再試行方針でwire形ではない
3. `progress-repository.ts`の`ProgressCursor`のdocを直す
    - 「変換と形式の検証はController」と書いてあるが、Decision `2026-10-01T23-32-40`は「routeのzod。Controllerは数値→文字列の戻しだけ」と言う
4. epic branch名の確認
    - `origin/feature/playback-progress`と推定。`gh`がTLS errorで取れず未確認
5. `PROGRESS_WRITE_MAX_ATTEMPTS`と`progressWriteRetryableHttpErrorCodes`をcontractsから出す
    - codeの使用者はゼロ（コメントのみ）。意図した使用者のVMは、depcruiseの`web-viewmodels-no-components-contracts`でcontractsをimportできない
    - 置き場の案: 試行回数はweb側（VM）、retry可否の導出は`web/api`。workerのinfra commentは定数名を指さず方針を書く
6. `EpisodeManuscript`型と`manuscript.schema.json`の食い違い検出を決める
    - 現状はajv後の`as`でつなぐだけ。案: fixtureをschemaで検証するcontract test（安価）、型の生成（要tool導入と版固定）
7. 共有語彙の置き場を決める（`worker/src/entities`か、web・worker共有の新dirか）
    - webは応答schemaをruntimeで使う（`playback-api-client.ts`）ため、語彙の関数がweb bundleに入る
    - `entities.md`§7「他の内部層をimportしない」との整合が、共有dir案の論点
