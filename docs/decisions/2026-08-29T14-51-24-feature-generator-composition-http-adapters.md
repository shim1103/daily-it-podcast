---
name: Cursor CLI 経路の secret 値を Composition→processenv の closure に閉じ、cursorcli には inject env 名（N2）だけを流す
date: 2026-08-29T14:51:24
branch: feature/generator-composition-http-adapters
---

## 1. Decision

1. Cursor CLI の secret 値は `internal/composition` が `config` から取り出し、`processenv` の factory 構築関数へ直接渡す。値は factory の closure に閉じ、`cursorcli`（vendor Adapter）を経由しない。
2. `processenv` は「inject env 名を受け取り、閉じ込めた secret 値と親環境アクセス手段で `commandlaunch.Launcher` を組む」factory を返す。factory の型（`func(envName string) commandlaunch.Launcher`）は `commandlaunch`（抽象）が所有する。
3. `cursorcli` は自身が所有する inject env 名（N2、`CursorAPIKeyEnvName`）を factory へ渡して `Launcher` を得る。`cursorcli` は secret 値・親環境アクセス手段・runtime 実装 package を知らない。
4. 親環境アクセス手段（production では `os.LookupEnv`）は `internal/composition/runtime.go` が `sharedHTTPClient()` と同じ層で production 既定値として供給し、factory 構築時に注入する。`processenv.Launcher` は暗黙 default を持たず、未注入なら `Launch` が起動前に error を返す。

## 2. Reason

- 結論1: 前タスクの案P は N2 を cursorcli へ隠蔽したが、`cursorcli.NewTextWriter(apiKey string, ...)` が secret 値を引数で受け取り struct 経由で保持する形になった。これは `docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md` §1「Go process は秘密値を一度も保持せず」の思想、および先行 Decision（`docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md`）§5「vendor Adapter は…秘密値…を知らない」に反する。secret 値の生存範囲を composition→processenv の closure に閉じ、cursorcli を通さなければ、vendor Adapter は値を一度も見ない。HTTP 経路との対称性: `gemini` Adapter は `*http.Client` と revealed string を受け取るが、これは HTTP transport が値を注入で受ける形で、command 経路も「値は runtime 実装（processenv）が closure で保持し、Adapter は名前だけ」にすると、値を保持する主体が runtime 実装に一本化される。
- 結論2: factory の型を `commandlaunch`（出口の I/O 契約 = 抽象）が持てば、`processenv`（runtime 実装）も `cursorcli`（vendor Adapter）も `commandlaunch` だけを import すればよい。`cursorcli` が `processenv` を import すると先行 Decision §4「vendor Adapter は runtime 具象に依存しない」に反し、`processenv` が `cursorcli` を import すると同 §5後段「runtime 実装は vendor flag を知らない」に反する。両方向の依存を `commandlaunch` 経由の抽象に倒す。この制約は `.golangci.yml` の depguard `infrastructure` ルールが `internal/infrastructure` prefix 全体を allow するため機械検出されない。設計規律と review で守る。
- 結論3: N2（inject env 名）は Decision-1 で `cursorcli` 所有と確定した。`cursorcli` が factory へ N2名を渡す1行が、N1（config が読む key 名）→ N2（child へ inject する env 名）の写像点になる（`ports-adapters` §12「隣接層の同名識別子は写像表で1対1」）。cursorcli が secret 値を持たないことで、Cursor CLI の呼び出し仕様（argv・envelope・inject env 名）だけを知る Adapter になり、先行 Decision §5「vendor Adapter は出口の話し方だけを知る」を満たす。
- 結論4: `os.LookupEnv`（親環境アクセス能力への入口）は HTTP Client（外部 I/O 能力への入口）と同じ種類の production runtime 既定値。`runtime.go` が両方を `shared*()` で供給し、Composition が明示注入することで、`configuration-boundary`（「Composition Root が具体依存を選択・生成・結線する」）の責務を infrastructure が肩代わりしない。暗黙 default（`nil` → `os.LookupEnv`）は Composition が注入をサボっても動く経路を残すため廃止し、未注入を `Launch` の起動前 error で顕在化させる。

## 3. Rejected

1. `cursorcli.NewTextWriter(apiKey string, newLauncher func(SecretEnv) Launcher)` が secret 値を受け取り `SecretEnv{Name: CursorAPIKeyEnvName, Value: apiKey}` を組む案（前タスクの案P）。N2 の隠蔽は達成するが、vendor Adapter が secret 値を引数で受け取り struct 経由で保持するため、`docs/decisions/2026-08-22T11-55-22-*` §1「Go process は秘密値を一度も保持せず」と先行 Decision §5「vendor Adapter は秘密値を知らない」に反する。
2. `composition/cursorcli.go` が `cursorcli.CursorAPIKeyEnvName` を参照して `SecretEnv` を組み、`processenv.NewLauncher` へ渡す案（Decision-1 直後の実装）。Composition が inject env 名を読み取るため Cursor CLI protocol 知識が結線層へ漏れ、HTTP 経路（Composition は header 名を知らない）と非対称。
3. factory 型を `cursorcli` または `processenv` package が所有する案。`cursorcli` 所有なら `processenv` が `cursorcli` を import し先行 Decision §5後段に反する。`processenv` 所有なら `cursorcli` が `processenv`（runtime 具象）を import し先行 Decision §4 に反する。抽象（`commandlaunch`）に置く以外は依存方向が破綻する。
4. `commandlaunch.Command` に `SecretEnvName` または `SecretValue` field を追加し `Launch` 時に渡す案。`docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` §3「command invocation は秘密値を持たない」および `cursorcli` の `@ensure Launch へ渡す Command は Program・Args・Stdin のみ` を破る。
5. `processenv` の `lookupEnv == nil` 暗黙 `os.LookupEnv` fallback を維持する案。Composition が注入をサボっても動くため「Composition が必ず runtime 手段を注入する」規律を骨抜きにする。

Decision-3 は Decision-1 / Decision-2 を supersede しない（両者の延長。N2 所有は Decision-1、結線対称化は Decision-2 が扱い、本 Decision は「secret 値の生存範囲」と「runtime 手段の注入」に答える）。関連 Decision 参照は file 名で書く。
