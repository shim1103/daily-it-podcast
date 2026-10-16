---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. playback/workerの同形の型・写しを削る範囲をshimと決める
    - 調査済み（editなし）。候補は`ProgressMergeIncoming`の再構築3箇所の削除、progress→contract写像の1関数化、`ProgressWriteUseCaseOutput`を`Pick<EpisodeProgress, ...>`へ寄せること
    - `EpisodeManuscript`と`EpisodeItem`の統合は`application.md`§7-2との関係で慎重に扱う。shimの許可待ち
2. 時刻の正規化で、手書きの文法・暦・offset検査を減らす案をshimと決める
    - 実測: zodの`z.iso.datetime({offset:true})`は、2月30日・24時・60秒・`+24:00`・offset無し・日付のみを全て拒否する。`domain/utc-instant.ts`の手書きと重複する
    - 案B（推奨）: 文法はcontracts（zod）だけが持つ。domainは`toUtcInstant(date: Date)`（有効性・範囲・brand）と`isUtcInstant(string)`だけにする。案A=現状維持、案C=domainがzodに依存
    - 案Bを選ぶと、`domain-vocabulary-shared-by-web-and-worker.md`のReason 5とRejected 6（`z.iso.datetime`を入口に残す案）を更新する
3. `episodeId`を`EpisodeId`にするか決める（cursorの形式も同じ問い）
    - 案a=別名のみ（意味の明確化、安い）、案b=brand＋非空の入口（`episodeId`を含むfileは非test36・test39、test内literal166箇所）、案c=worker内だけ
    - `ProgressCursor`は別名のままで、`ProgressVersion`のbrandと不揃い
4. `PROGRESS_WRITE_MAX_ATTEMPTS`と`progressWriteRetryableHttpErrorCodes`をcontractsから出す
    - codeの使用者はゼロ（コメントのみ）。意図した使用者のVMは、depcruiseの`web-viewmodels-no-components-contracts`でcontractsをimportできない
    - 置き場の案: 試行回数はweb側（VM）、retry可否の導出は`web/api`。workerのinfra commentは定数名を指さず方針を書く
5. `EpisodeManuscript`型と`manuscript.schema.json`の食い違い検出を決める
    - 現状はajv後の`as`でつなぐだけ。案: fixtureをschemaで検証するcontract test（安価）、型の生成（要tool導入と版固定）
6. skillsを更新する（`agent-standards`側。このsessionでは更新しない指示）
    - `coding-terms/backend-terms/entities.md`§7（他の内部層をimportしない）と`ring-model.md`§4-7（root直下は層のdirだけ）が、web・worker共有の最内側dir`domain/`を想定していない。本branchは`domain-vocabulary-shared-by-web-and-worker.md`を正とした
    - `defensive-design.md`§2-1（呼ばれ方を仮定せず壊れない）と§2-3（callerの責務と重なる防御は入れない）の文面の衝突を直す
    - `/std`で再反映した後も、上の3箇所の文面は変わっていない（確認済み）
7. commitの扱いをshimと決める
    - 本branchに4 commit（`86e7e39`・`9778ce8`・`741414f`・`18dc12c`）を作った。根拠は`AGENTS.md`§4の常設指示で、このsessionでshimが明示したものではない
    - `18dc12c`（`.codex/config.toml`）はshimの未commit編集を`--repo`で入れた。残すか、`git reset --soft`で戻すかの判断待ち
    - `/std`の再反映で`.codex/config.toml`にコメント1行が増え、未commitのまま残っている
8. epic branch名の確認
    - `origin/feature/playback-progress`と推定。`gh`がTLS errorで取れず未確認
9. `.codex/config.toml`を追跡から外す（shimのqueue）
    - `.gitignore`に`.codex`があるのに、`config.toml`だけtrackedのまま。`git rm --cached .codex/config.toml`で外す案
    - 影響: 他のcheckoutではpullでfileが消える。`/std`が再生成する。`18dc12c`と`/std`で増えた1行は、外せば不要になる
    - read-onlyのturnだったので未実行。shimの許可待ち
