---
name: "task-feature/playback-progress-versioned-persistence"
description: "feature/playback-progress-versioned-persistenceで進行中のtask。"
when_to_use: "feature/playback-progress-versioned-persistenceで作業を再開する時、PRを作る前に未完了を確かめる時。"
non-scope: "branch非依存の未完了とOpen questions（→wiki）、完了済みのtask、sessionの経緯（→daily）。"
---

1. playback/workerの同形の型・写しを削る範囲をshimと決める
    - 調査済み（editなし）。候補は`ProgressMergeIncoming`の再構築3箇所の削除、progress→contract写像の1関数化、`ProgressWriteUseCaseOutput`を`Pick<EpisodeProgress, ...>`へ寄せること
    - `EpisodeManuscript`と`EpisodeItem`の統合は`application.md`§7-2との関係で慎重に扱う。shimの許可待ち
2. 時刻の検査をdomainでどう持つか決める（案C=domainがzodに依存、案D=domainはzod非依存で`Date`を受ける）
    - Opus 5.5の相談結果: 推奨は案D。文法・暦の手書きは、contractsの`z.iso.datetime({offset:true})`に任せれば消える。domainは`toUtcInstant(date: Date)`（範囲と固定幅）と`isUtcInstant`（往復比較）だけにする。最内側をzodに結合させる費用（zod 3→4でbrand・iso・errorのAPIが変わった前例）を払わずに済む
    - 私の以前の推奨は案C。zod skill §1（schemaを唯一の正とする）に沿うが、差は小さい。決め手は「domainがzodに依存してよいか」
    - どちらの案でも: 小数を3桁へ切ってから`new Date(v)`に渡す（4桁以上はECMA-262が規定せず、Nodeで切り捨ては実測、workerdは未確認）。D1境界は`+09:00`を受ける変換と別に、正規形の確認（往復比較）が要る
    - 案Dを選ぶと、`domain/utc-instant.ts`を縮め（文法・暦・offset算術が消える）、`contracts/http.ts`の時刻schemaを`z.iso.datetime({offset:true})`に戻して`toUtcInstant`を呼ぶ。`domain-vocabulary-shared-by-web-and-worker.md`のReason 5・Rejected 6を更新する
3. `episodeId`を`EpisodeId`にするか決める（cursorの形式も同じ問い）
    - 案a=別名のみ、案b=`z.string().min(1).brand()`をdomainに置く（`episodeId`を含むfileは非test36・test39、test内literal166箇所）、案c=worker内だけ。Opus 5.5の見立て: brandにする根拠（`entities.md`§10-1の『繰り返し破られた』証拠）が無い。範囲外のままか、案aまで
    - `ProgressCursor`は別名のままで、`ProgressVersion`のbrandと不揃い
    - 論点: 「非空」の定義が4箇所に分かれる（contractsの`min(1)`、`manuscript.schema.json`の`minLength`、webのhash adapterの`"" → null`、`episodePath`の`@require`）。正本の候補は、言語横断の`manuscript.schema.json`と、その写しとしてのdomain。`z.toJSONSchema(z.string().min(1))`は`{type:string,minLength:1}`で、`manuscript.schema.json`の`episodeId`と同形（実測）
    - 論点: `episode-layout.md`は`episodeId`を「不透明な対応キー。生成規則はGeneratorに閉じ、Readerはstem一致だけ見る」とする。文字種は未定義なので、domainが言える不変は非空まで
    - 論点: R2 adapterは`${episodeId}.json`のkey組立を無検査で行う。`/`を含むidの扱いは未確認（Honoのparam decodeも未確認）
4. `PROGRESS_WRITE_MAX_ATTEMPTS`と`progressWriteRetryableHttpErrorCodes`をcontractsから出す
    - codeの使用者はゼロ（コメントのみ）。意図した使用者のVMは、depcruiseの`web-viewmodels-no-components-contracts`でcontractsをimportできない
    - 置き場の案: 試行回数はweb側（VM）、retry可否の導出は`web/api`。workerのinfra commentは定数名を指さず方針を書く
5. `EpisodeManuscript`型と`manuscript.schema.json`の食い違い検出を決める
    - 現状はajv後の`as`でつなぐだけ。Goはembedしたschemaで検証し、組立structは手書き。TSは手書き型とHTTPの`episodeItemSchema`が写し（HTTP側は層境界のDecisionで意図的に別）
    - shimの判断: fixtureをschemaで検証するcontract test（案i）は大変なのでやめる
    - 残る案: ii=型の生成（要tool導入と版固定。未導入）、iii=`z.fromJSONSchema`。実測: 検証は動く（空id・余分なfield・負のdurationを拒否）が、出力型は静的に推論されない
    - iiとzodの関係: 生成するのは`EpisodeManuscript`型だけで、zodはHTTP契約とdomainの葉schemaに残る。JSON Schemaの`minLength`は型にならないので、`EpisodeId`のbrandは生成型から出ない（manuscript境界で`episodeIdSchema.parse`する）
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
10. 相談用sub-agentが作った一時file`apps/playback/.zprobe.mts`を削除する
    - read-only契約に反してrepo内に作られた（597B、untracked、gitignore対象外）。`rm`はhookに拒否された。shimが実行する
    - 削除前に`/commit --repo`をしない（`--repo`は未追跡fileも拾う）
