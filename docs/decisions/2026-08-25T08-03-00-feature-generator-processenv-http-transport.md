---
name: processenv runtime 実装は Infrastructure Error 型を持ち、proxy 系との plain error 混同を再発させない
date: 2026-08-25T08:03:00
branch: feature/generator-processenv-http-transport
---

## 1. Decision

1. `commandlaunch` / `secrettransport` の runtime 実装（`processenv` パッケージ群）は、vendor Adapter（`getxapi`・`gdrive` 等）と同じ Infrastructure 層に属する。Infrastructure Error クラス（`type Error struct { Op string; Err error }` + `Error()` / `Unwrap()` + `infraErr` helper）を持ち、`fmt.Errorf` の plain error だけで失敗を表さない。両義務は同型・同スコープで扱い、片方だけを免除しない。
2. `secrettransport/processenv.Client` は AgentSecrets proxy 経由の `agentsecrets.Client` と異なり、秘密値を **この process 内で HTTP request へ直接注入**して送信する。中間 proxy が無いため、送信先 upstream が受け取る request は vendor への実 request そのものであり、`X-AS-Target-URL` / `X-AS-Method` のような meta header 経由の contract を持たない。processenv 系の Narrow / Sociable Unit Test は `r.Method` / `r.URL.String()` のような標準 HTTP フィールドで実 request を検証し、proxy 系 test に残る `X-AS-*` header 読み取りパターンをそのまま流用しない。
3. Narrow Integration Test は 1 test 1 Given-When-Then を守る。同一 `Inject`（Bearer / Header / Form / JSON）内の複数 field 種別を検証する場合でも、種別ごとに test 関数を分ける。test 関数名から検証対象の種別が判別できない束ね方をしない。
4. TargetURL が host 付き絶対 https URL であることの検証は `secrettransport` package が `ValidateAbsoluteHTTPSURL` として一元的に持ち、`secrettransport.Client` の runtime 実装（`processenv.Client`・`agentsecrets.Client`）はこれを呼ぶだけで独自に再実装しない。この検証は runtime 具象に依存しない契約レベルの判断であり、`secrettransport` package（`contract.go` と同じ場所）が所有する。
5. `commandlaunch` / `secrettransport` の `processenv` 実装は、環境変数 lookup をコンストラクタで DI する（`lookupEnv func(key string) (string, bool)` を `NewLauncher` / `NewClient` の引数として受け取り、nil のとき `os.LookupEnv` へ fallback する）。`os.LookupEnv` を method 内部や private helper へ直書きしない。2 つの processenv 実装が同じ形の DI を持つことを、次に processenv runtime を追加する agent が確認できる基準にする。

## 2. Reason

1. `infrastructure.md` §4 は Infrastructure Error クラスの責務を層の規約として定義しており、契約実装（`commandlaunch.Launcher` / `secrettransport.Client` の runtime 具象）も vendor Adapter と同じ Infrastructure 層に属する（先行 `2026-08-25T13-53-55` §1-4「契約 package は Infrastructure に置く」）。層で規約が変わるという例外を認めると、次に runtime 実装を追加する agent が同じ欠落を繰り返す。
2. `agentsecrets.Client`（proxy 方式）と `processenv.Client`（直接注入方式）は、同じ `secrettransport.Client` interface を満たすが、実 I/O 契約（meta header 経由か標準フィールドか）が異なる。この違いを明文化せずに test だけ流用すると、次に runtime 実装を追加する agent が誤って旧 proxy 方式の header 読み取りパターンを新方式へコピーする（`philosophy` §2-1 Orthogonality、実装は違えど regression の温床になる）。
3. `naming-and-layout.md` §2-4「複数 case の Given・When・Then を top-level へ直列に並べない」への違反は、review 4 角度（reuse / simplification / efficiency / altitude）でも「borderline」として見送られ、確定判断に至らなかった。shim の audit で初めて修正対象として固定した。同種の借越判断を simplify レビューだけに委ねず、1 test 1 GWT を再発防止の判断として明記する。
4. `secrettransport/processenv/client.go` と `agentsecrets/proxy.go` は同じ TargetURL 検証（scheme が https か、host が非空か）を個別に実装していた。`design-philosophy.md` §2-2 DRY「同一 logic を複数箇所に実装しない」に反する状態であり、2 つ目の runtime 実装（`processenv`）が生まれた時点が「重複を発見したら、より上位の layer へ一元化する」を適用するタイミングである。共通 helper への切り出しは `philosophy` §2-3 KISS が言う「現時点で必要な最小の構造」の範囲に収まり、先送りする理由がない。
5. `commandlaunch/processenv.Launcher` は `os.LookupEnv` を `Launch` method 内部へ直書きしていた。同じ `processenv` という runtime 名を持つ `secrettransport/processenv.Client` がコンストラクタ DI 済みである以上、片方だけ直書きのままだと `philosophy` §4-5 Principle of Least Astonishment（一貫性は他の全原則に優先する、`design-philosophy.md` §5-1）に反する。2 つの processenv 実装は同じ設計判断（DI 済みかどうか）に同じ答えを返すべきであり、後から見つかった非対称を放置しない。

## 3. Rejected

1. `commandlaunch/processenv/launcher.go` の既存 plain error を「先行 Issue で完了済みだから今回は踏襲でよい」とする案。先行実装の欠落を新規実装が模倣すると、欠落が既成事実として拡大する。両方まとめて是正する。
2. Narrow Integration Test の 3 シナリオ（Bearer+Header／Form／JSON）を 1 関数のまま「upstream server 起動コストの amortize」を理由に維持する案。1 test 1 GWT は上位 `testing-strategy` の命名規則であり、実行コストの節約を理由に規約から逸脱しない。
3. TargetURL 検証の重複を「`agentsecrets` は Out of Scope だから触らない」を理由に放置する案。検証ロジックの共通化は `agentsecrets.Client` の振る舞い・契約を変えず、呼び出し方法を共通 helper へ委譲するだけであり、Out of Scope 制約（削除しない・振る舞いを変えない）と両立する。
4. `commandlaunch/processenv.Launcher` の lookup 直書きを「`secrettransport/processenv.Client` が正であり `Launcher` 側は追従不要」として据え置く案。一貫性は優先順位で他の全原則に優先するため、正しい側の設計を確認した時点で揃えないままにしない。
