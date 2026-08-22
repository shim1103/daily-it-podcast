## Cursor CLI TextWriter: 秘密境界の是正と test 構造の整理

Issue draft。`create-issue` で正式化する前の一時置き場。

type: `refactor` / scope: `generator`

## 1. Summary

このIssueでは、実装済みの Cursor CLI `TextWriter` Adapter に対し、秘密供給を `agentsecrets env --` 経由へ移し、あわせて review で判明した test 構造と命名の誤りを取り除く。完了後、Cursor CLI へ渡る環境変数は必要最小限になり、`composition/` に前例のない結線 test と、定数 SSoT を写経するだけの test が無くなる。振る舞いの追加は無い。

## 2. Context

1. 先行 Issue `generator-cursor-text-writer.md`（完了・削除済み）で Adapter を実装し、全 gate pass・`cursorcli` coverage 100% に到達している。このIssueはその後続で、先行 Issue を復活させるものではない
2. `agentsecrets` CLI には `env` subcommand が実在する。`agentsecrets env -- <command> [args...]` は keychain から秘密を解決して子 process の環境変数へ注入し、disk へは何も書かない。先行 decision はこの存在を知らずに書かれた
3. 既存の `internal/infrastructure/agentsecrets/proxy.go` は HTTP proxy 専用で、公開 API は `Do` のみ。秘密値を取り出す API を意図的に持たない（`proxy.go:15` の why comment）。したがって Go 側から値を取得する経路は存在しない
4. 現在の `runCursorCLI` は `exec.CommandContext` をそのまま使うため、親 process の環境を暗黙に全継承する。generator が持つ他 vendor の秘密が Cursor CLI へ到達している
5. `composition/` は `scripts/generator/test-unit.sh` が coverage 計測から除外する層である（awk による除外と `# why: Composition Root は結線だけ。` の comment）。既存 4 factory のうち `gemini.go` / `getxapi.go` / `twitterapiio.go` の 3 つには結線 test が無い。`gdrive_sociable_unit_test.go` は結線 test ではなく UseCase の validation 順序検証である
6. `newTextWriterForTest` は production の `NewTextWriter()` が呼ぶ唯一の生成経路であり、test 専用ではない。中身は `return &TextWriter{runFn: runFn}` の 1 行のみで struct literal と等価

## 3. Canonical Sources

1. `docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md` — CLI 境界の秘密供給を `agentsecrets env --` に定める。HTTP 境界との機構の分岐と、子 process へ渡る env の範囲
2. `docs/decisions/2026-08-18T16-30-00-feature-cursor-text-writer.md` — Port 境界、argv 固定、envelope 解釈、stderr の扱い。§4 のみ上記 decision が上書き済み
3. `docs/decisions/2026-08-16T00-06-30-docs-agentsecrets-secret-export.md` — local 秘密は名前参照で一本化し、実行主体で経路を分けない
4. `apps/generator/internal/infrastructure/manuscript/cursorcli/constants.go` — Cursor CLI argv 決定値の SSoT
5. `apps/generator/internal/infrastructure/manuscript/cursorcli/text_writer.go` — Adapter 実装と compile-time assertion
6. `apps/generator/internal/application/port/text_writer.go` — Port 境界
7. `testing-strategy` skill `quality.md` §7 — test 有効性の判定基準
8. `coding-style` skill `naming.md` §1 — 名前と実態の一致
9. `philosophy` skill — §4-1 / §4-3 / §5-1

## 4. Scope

### In Scope

1. `apps/generator/internal/composition/cursorcli_sociable_unit_test.go` を削除する。Port 適合の担保は `text_writer.go` の `var _ port.TextWriter = (*TextWriter)(nil)` が compile-time に果たす
2. `newTextWriterForTest` を廃止し、呼び出し箇所を `&TextWriter{runFn: ...}` の struct literal へ置き換える。`NewTextWriter()` と test の両方を追随させる。あわせて `TestNewTextWriter_usesRealExec_whenConstructedForProduction` の存在理由を再評価し、`runFn` が非 nil であることの確認以上を検証していないなら削除する
3. `apps/generator/internal/infrastructure/manuscript/cursorcli/.claude/` を削除する。git 未追跡の作業残骸で、他 adapter には存在しない
4. 秘密供給を `agentsecrets env --` 経由へ変更する。`constants.go` へ wrapper 用の定数を追加し、`buildArgs` / `runCursorCLI` を改修する。子 process へ渡る env を最小化する
5. `TestWrite_omitsForceAndAutoAndFastModel_whenExecSucceeds` を削除する。先行 Issue の AC-4 相当は `TestWrite_buildsArgvInFixedOrderWithoutOmission_whenExecSucceeds` の argv 構造 test と `constants.go` 自身の SSoT 性で担保する
6. 上記に伴う test の追加・修正

### Out of Scope

1. `gemini` の `newSpeechSynthesizerForTest` の命名是正。同じ命名の誤りを持つが、nil 正規化 logic があるため関数化の理由自体はあり、既存 adapter への波及変更になる。→ Notes 1
2. `TextWriter` Port の signature 変更
3. Cursor envelope の解釈 logic、error 型、stderr の扱い（先行 decision §1・§3・§5 のまま）
4. `TextWriter` を呼ぶ UseCase、TTS、Drive 書込
5. `agentsecrets/proxy.go` への変更。HTTP 境界の機構は現状維持
6. 実 `agent` / 実 `agentsecrets` を起動する Integration test

## 5. Contract

`port.TextWriter` の公開契約は変更しない。`Write(ctx, brief)` の入力・成功・失敗の対応は先行 decision のまま維持する。

変わるのは Adapter 内部の実行手段だけである。

| 項目 | 変更前 | 変更後 |
|---|---|---|
| exec する program | `agent` | `agentsecrets` |
| Cursor CLI の位置 | exec の program 名 | `env --` の後ろの argv |
| Cursor 秘密の供給元 | 親 process の env を暗黙全継承 | `agentsecrets env` が子 process へ注入 |
| 子 process の env | 親の環境全体 | 実行に必要な最小限 |
| Go process が持つ秘密値 | なし（ただし environ 経由で到達可能） | なし |

`BinaryName` の扱いは次の方針で決める。`BinaryName = "agent"` は Cursor CLI 自身の識別子として意味が正しいため名前も値も変えない。`runCursorCLI` が exec する program 名は wrapper の binary になる。既存の argv 構造 test は、この新しい固定順序を検証する形へ更新する。

wrapper 側の定数の**置き場所は `cursorcli/constants.go` ではない**。`constants.go` は冒頭 comment で「Cursor CLI（非SDK）の argv 構成に必要な確定値だけを定義する」と宣言しており、wrapper の binary 名・subcommand・separator は AgentSecrets の知識であって Cursor CLI の知識ではない。これらは `infrastructure/agentsecrets` が `EnvBinary` / `EnvSubcommand` / `ArgSeparator` として所有し、argv の組み立ても `agentsecrets.EnvWrapper.Command` が行う。`cursorcli` は `agentsecrets` を**直接 import** して使う。

新しい interface を挟まない（DIP を適用しない）。Infrastructure 同士の import であり、既存 5 adapter が `agentsecrets.Client` を具体型のまま import しているのと同型である（`architecture/backend/infrastructure.md` §6 は Infrastructure の import 許可先に外部SDKを挙げる）。

## 6. Constraints

1. 既存の `docs/decisions/` 配下の file を書き換えない。過去の判断は書き換えず、新しい decision で上書きする（`logging` §3）
2. Go process が秘密値を保持しない。値を取得する API を `agentsecrets` package へ足さない
3. 子 process の env に、その呼び出しが必要としない秘密を載せない（`philosophy` §4-3）
4. `philosophy` §5-1 の一貫性を理由に、既存 code に存在する誤った pattern を新しい code へ複製しない。`ForTest` 命名と結線 test は「既に他所にあるから」を根拠に残さない
5. 定数や対応表の値そのものが SSoT である層へ、同じ値を書き写して存在確認する test を足さない（`quality.md` §7）
6. coverage 計測対象外である `composition/` に、coverage を理由とした test を置かない
7. argv は決定的に固定する。実行のたびに順序や要素が変わらない

## 7. Acceptance Criteria

1. [ ] `apps/generator/internal/composition/cursorcli_sociable_unit_test.go` が存在しない
2. [ ] `newTextWriterForTest` が codebase から消え、`grep -rn "newTextWriterForTest" apps/generator` が 0 件を返す
3. [ ] `apps/generator/internal/infrastructure/manuscript/cursorcli/.claude/` が存在しない
4. [ ] `TestWrite_omitsForceAndAutoAndFastModel_whenExecSucceeds` が存在せず、`constants.go` の `Mode` / `ModelID` の値を書き写す test が `cursorcli` package に無い
5. [ ] exec stub が受け取る program 名が wrapper の binary であり、argv が `env` subcommand と `--` separator を経て `BinaryName` と Cursor flags へ続く固定順序になっている
6. [ ] argv のどの要素にも秘密値が含まれない
7. [ ] 子 process へ渡る env が、親 process の環境全体ではなく明示的に構築された最小集合である。exec stub がこれを観測できる
8. [ ] `--force` / `--yolo` / `auto` / fast model が argv に現れない（AC-5 の固定順序 test が構造として担保する）
9. [ ] 成功時に非空の text 断片を返し、非0 exit・envelope 不正・`result` 欠落で `*cursorcli.Error` を返す既存の振る舞いが維持されている
10. [ ] `text_writer.go` の `var _ port.TextWriter = (*TextWriter)(nil)` が残っている
11. [ ] `cursorcli` package の coverage が 90% gate を満たす

## 8. Verification

```bash
cd apps/generator
go test ./internal/application/... ./internal/infrastructure/... ./internal/composition/...
```

```bash
scripts/generator/test-unit.sh
scripts/generator/check-static.sh
```

すべて pass し、`test-unit.sh` の coverage gate 90% を維持する。実 `agent` / 実 `agentsecrets` の起動は不要。

## 9. Dependencies

先行: Cursor CLI `TextWriter` Adapter の実装（完了）、`docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md`

後続: `TextWriter` を N 回呼ぶ UseCase、cmd 入口、GHA workflow

## 10. Risks

1. wrapper 経由にすると exec する program が `agentsecrets` へ変わるため、`agentsecrets` が PATH に無い環境で Adapter が起動時に失敗する risk。exit 由来の error として既存の失敗経路へ畳まれるので Port 契約は変わらないが、失敗原因が Cursor 側か wrapper 側か区別しにくくなる
2. env を明示構築へ変えると、Cursor CLI が暗黙に依存していた環境変数（`HOME`、`PATH` 等）が落ちて動かなくなる risk。何を最小集合に含めるかは実装時に確定させ、根拠を code の why comment に残す
3. `composition/` の結線 test を削ると、factory の存在自体を検査するものが無くなる。compile-time assertion と build が担保するため許容するが、既存 3 factory と揃うだけで新たな穴ではない

## 11. Notes

1. `gemini` の `newSpeechSynthesizerForTest` も名前と実態が一致しない同種の問題を持つ。ただし nil 正規化 logic を持つため関数化そのものには理由がある。`ForTest` 系の命名は `gemini` と `cursorcli` の 2 箇所のみで、他 4 adapter（`gdrive` / `oauth` / `getxapi` / `twitterapiio`）には無い。別 Issue 候補として切り出す
2. model 名 `composer-2.5` は vendor の世代交代で必ず変わる。`constants.go` の定数更新を止めにかかる test を置かないこと自体が、この Issue の削除判断の根拠である
3. HTTP 境界（既存 4 adapter）の秘密供給は AgentSecrets proxy のまま変えない。境界の種類で機構が分かれることは decision で明示済み
4. 次アクション: `create-issue` で Issue 化する

## 12. shim の手動作業: Cursor 専用 AgentSecrets project の作成

code は既に「Cursor 専用 project の dir で wrapper を起動する」構造で書かれている。その dir がまだ存在しないため、**この作業を行うまで Cursor CLI 呼び出しは wrapper の起動失敗として実行時 error になる**。

### 12.1 なぜ project 分離が要るか

1. `agentsecrets env` に、解決する秘密を絞る flag は存在しない。`--help` の Flags は `-h` のみで、説明文も「Resolves **all secrets** from the active project」である
2. `agentsecrets secrets policy` は allowed domains / HTTP methods だけを制御し、CLI の env 注入は絞れない
3. `agentsecrets exec` は秘密値を stdout へ返すため採用しない。Go process が値を見た時点で zero-knowledge が崩れる（decision `2026-08-22T11-55-22` §3-1）
4. AgentSecrets の active project は git root の自動検出ではなく、**実行時 current directory 直下の設定 file** で決まる（`agentsecrets` skill `project-binding.md` §1）

したがって、渡る秘密の範囲を絞る唯一の手段は **wrapper を起動する working directory を Cursor 専用 project の設定 dir へ向けること**である。code はこれを `exec.Cmd.Dir` で行う。

**分離が無い状態で動かすと、その cwd が指す project の全 secret が Cursor CLI へ渡る。** 現在 repo root の `.agentsecrets/project.json` は `daily-it-podcast` 単一 project を指し、全 vendor の secret がそこに入っている。

### 12.2 code が期待する構造

| 項目 | 期待値 |
|---|---|
| project dir | `agentsecrets.DefaultProjectDir(projectName)` が解決する。dir の配置規約は `agentsecrets.ProjectsRootName`、project 名は `cursorcli` の `projectName` が所有する |
| dir 直下 | `.agentsecrets/project.json` が存在する |
| project が持つ secret | `CURSOR_API_KEY` のみ |
| path 形式 | 絶対 path。相対 path は `agentsecrets.EnvWrapper.Validate` が実行時 error として弾く |

`$HOME` 配下へ置く理由: repo 内へ置くと秘密境界の構成が repo の clone 状態と結びつき、clone していない環境で絞り込みが黙って外れる。repo root の `.agentsecrets/project.json` は git 追跡されているため、その隣に置く選択も同じ問題を持つ。

### 12.3 project 分離の手順と実測

手順は次の 5 段で、いずれも実施済みである。

1. `mkdir -p "$HOME/.agentsecrets-projects/cursor"`
2. その dir で `agentsecrets use-project cursor` を実行し、`daily-it-podcast` とは別の project を紐付ける
3. その project へ `CURSOR_API_KEY` だけを登録する。他 vendor の secret を入れない
4. 既存 `daily-it-podcast` project から `CURSOR_API_KEY` を削除する。残すと分離の意味が半減する（Cursor 以外の呼び出し先へ Cursor key が届く経路が残る）
5. 両方の dir で `agentsecrets env -- sh -c 'env | grep -oE "^[A-Z_]+_(API_KEY|ID)"'` 相当を実行し、注入される secret を対比する

5 の実測結果。cwd だけが違い、注入される範囲が変わることを確認した。

| cwd | 注入 |
|---|---|
| repo root | `Injecting 5 secrets`（`GETX_API_KEY` / `TWITTER_IO_API_KEY` / `GEMINI_API_KEY` / `DRIVE_FOLDER_ID` を含む） |
| Cursor 専用 project dir | `Injecting 1 secret: CURSOR_API_KEY` |

分離しなければ Cursor CLI へ他 vendor の 4 鍵が渡っていた。`agentsecrets env` に絞り込み flag が無いという前提と、cwd で active project が決まるという前提の両方が、実 CLI で裏付けられている。

`.agentsecrets/project.json` は運用設定であり、agent は変更しない。

### 12.4 宣言と実注入の乖離

`cursorcli.NewTextWriter` は `secretnames.CursorAPIKeyName` を `agentsecrets.EnvWrapper.SecretKeys` へ渡し、`RequiredSecretKeys()` で読み出せる。

**これは宣言であって、実際に注入される秘密の一覧ではない。** `agentsecrets env` は key 名を受け取らないため、この値は実行時に wrapper へ渡らない。実注入は 12.2 の project 構成が決める。宣言が意味を持つのは次の 3 点である。

1. 既存 5 adapter が `Inject` へ key 名を渡すのと同型に、依存する秘密を code へ明示する
2. 12.3-5 の検証で「注入されるべき集合」の期待値になる
3. 将来 wrapper が key 指定機構を持った時の受け口になる

乖離しうる事実は `EnvWrapper.SecretKeys` と `RequiredSecretKeys` の `warn:` comment に書いてある。

### 12.5 PATH 不在時の扱い（対処しない）

`agentsecrets` または `agent` が PATH に無い場合、`exec.LookPath` による事前検査は**行わない**。

既存 5 adapter は全て HTTP 境界で、AgentSecrets proxy 未起動時も接続 error として実行時に畳まれる。事前検査の前例は codebase に無い。`philosophy` §5-1（一貫性が他の全原則に優先する）に従い、環境側の不備は wrapper の起動失敗として同じ実行時 error 経路へ落とす。

同じ理由で、`EnvWrapper.Validate` は project dir の**実在**も検査しない。検査するのは「絞り込みが構造として成立しているか」（空でない・絶対 path である）だけで、環境の状態は実行時に現れる。
