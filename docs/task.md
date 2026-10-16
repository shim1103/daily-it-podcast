---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. playback/workerの同形の型・写しを削る範囲をshimと決める
    - 調査済み（editなし）。候補は`ProgressMergeIncoming`の再構築3箇所の削除、progress→contract写像の1関数化、`ProgressWriteUseCaseOutput`を`Pick<EpisodeProgress, ...>`へ寄せること
    - `EpisodeManuscript`と`EpisodeItem`の統合は`application.md`§7-2との関係で慎重に扱う。shimの許可待ち
2. domainの検査を手書きからzodへ寄せるか決める（時刻・`EpisodeId`共通）
    - 実測（zod 4.4.3）: `z.iso.datetime({offset:true}).refine(範囲).transform(toISOString).brand()`で、手書きの文法・暦・offset検査が全て消える。型は`z.output`で`string & $brand`になり、素の`string`は型errorで弾かれる。`satisfies z.ZodType<Domain, unknown>`も通る。失敗issueに入力値は載らない（custom issueも）
    - 規範: `selection-terms/criteria.md`第1基準が再発明の回避。zod skillは「schemaを唯一の正とし型は導出」。zodは既にweb・worker・contractsの依存
    - 案A=現状維持（手書き）、案B=手書き縮小（`toUtcInstant(Date)`）、案C（推奨）=domainがzodだけを許可して依存（`domain-no-npm`をzodのallowlistに変える）
    - 案Cを選ぶと、`domain/utc-instant.ts`の手書き（約70行）と`isUtcInstant`が、schemaと`safeParse`に置き換わる。`domain-vocabulary-shared-by-web-and-worker.md`の「zodもimportしない」「外部依存ゼロ」とReason 5・Rejected 6を更新する
3. `episodeId`を`EpisodeId`にするか決める（cursorの形式も同じ問い）
    - 案a=別名のみ、案b=`z.string().min(1).brand()`をdomainに置く（`episodeId`を含むfileは非test36・test39、test内literal166箇所）、案c=worker内だけ。案Cを採るなら、非空はzodが保証し`isEpisodeId`は不要
    - `ProgressCursor`は別名のままで、`ProgressVersion`のbrandと不揃い
    - 論点: 「非空」の定義が4箇所に分かれる（contractsの`min(1)`、`manuscript.schema.json`の`minLength`、webのhash adapterの`"" → null`、`episodePath`の`@require`）。正本の候補は、言語横断の`manuscript.schema.json`と、その写しとしてのdomain。`z.toJSONSchema(z.string().min(1))`は`{type:string,minLength:1}`で、`manuscript.schema.json`の`episodeId`と同形（実測）
    - 論点: `episode-layout.md`は`episodeId`を「不透明な対応キー。生成規則はGeneratorに閉じ、Readerはstem一致だけ見る」とする。文字種は未定義なので、domainが言える不変は非空まで
    - 論点: R2 adapterは`${episodeId}.json`のkey組立を無検査で行う。`/`を含むidの扱いは未確認（Honoのparam decodeも未確認）
4. `PROGRESS_WRITE_MAX_ATTEMPTS`と`progressWriteRetryableHttpErrorCodes`をcontractsから出す
    - codeの使用者はゼロ（コメントのみ）。意図した使用者のVMは、depcruiseの`web-viewmodels-no-components-contracts`でcontractsをimportできない
    - 置き場の案: 試行回数はweb側（VM）、retry可否の導出は`web/api`。workerのinfra commentは定数名を指さず方針を書く
5. `EpisodeManuscript`型と`manuscript.schema.json`の食い違い検出を決める
    - 現状はajv後の`as`でつなぐだけ。Goはembedしたschemaで検証し、組立structは手書き。TSは手書き型とHTTPの`episodeItemSchema`が写し（HTTP側は層境界のDecisionで意図的に別）
    - 案: fixtureをschemaで検証するcontract test（安価）、型の生成（要tool導入と版固定。未導入）、`z.fromJSONSchema`（zod 4.4.3にある。型は推論されない見込みで未検証）
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
