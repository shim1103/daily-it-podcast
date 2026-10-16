---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. playback/workerの同形の型・写しを削る範囲をshimと決める
    - 調査済み（editなし）。候補は`ProgressMergeIncoming`の再構築3箇所の削除、progress→contract写像の1関数化、`ProgressWriteUseCaseOutput`を`Pick<EpisodeProgress, ...>`へ寄せること
    - `EpisodeManuscript`と`EpisodeItem`の統合は`application.md`§7-2との関係で慎重に扱う。shimの許可待ち
2. 時刻以外の共有語彙（cursorの形式、`episodeId`）を`domain/`へ寄せるか、brandにするかを決める
    - 時刻の`UtcInstant`は`domain/`へ移した（`domain-vocabulary-shared-by-web-and-worker.md`）。cursorと`episodeId`は範囲外にした
    - `ProgressCursor`は別名のままで、`ProgressVersion`のbrandと不揃い
3. `PROGRESS_WRITE_MAX_ATTEMPTS`と`progressWriteRetryableHttpErrorCodes`をcontractsから出す
    - codeの使用者はゼロ（コメントのみ）。意図した使用者のVMは、depcruiseの`web-viewmodels-no-components-contracts`でcontractsをimportできない
    - 置き場の案: 試行回数はweb側（VM）、retry可否の導出は`web/api`。workerのinfra commentは定数名を指さず方針を書く
4. `EpisodeManuscript`型と`manuscript.schema.json`の食い違い検出を決める
    - 現状はajv後の`as`でつなぐだけ。案: fixtureをschemaで検証するcontract test（安価）、型の生成（要tool導入と版固定）
5. skillsを更新する（`agent-standards`側。このsessionでは更新しない指示）
    - `coding-terms/backend-terms/entities.md`§7（他の内部層をimportしない）と`ring-model.md`§4-7（root直下は層のdirだけ）が、web・worker共有の最内側dir`domain/`を想定していない。本branchは`domain-vocabulary-shared-by-web-and-worker.md`を正とした
    - `defensive-design.md`§2-1（呼ばれ方を仮定せず壊れない）と§2-3（callerの責務と重なる防御は入れない）の文面の衝突を直す
    - `/std`で再反映した後も、上の3箇所の文面は変わっていない（確認済み）
6. commitの扱いをshimと決める
    - 本branchに4 commit（`86e7e39`・`9778ce8`・`741414f`・`18dc12c`）を作った。根拠は`AGENTS.md`§4の常設指示で、このsessionでshimが明示したものではない
    - `18dc12c`（`.codex/config.toml`）はshimの未commit編集を`--repo`で入れた。残すか、`git reset --soft`で戻すかの判断待ち
    - `/std`の再反映で`.codex/config.toml`にコメント1行が増え、未commitのまま残っている
7. epic branch名の確認
    - `origin/feature/playback-progress`と推定。`gh`がTLS errorで取れず未確認
