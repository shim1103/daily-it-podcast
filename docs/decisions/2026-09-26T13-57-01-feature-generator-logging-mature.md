---
name: RetryReporterのnil検査はAdapter個別panicではなくComposition Root 1点へ集約する
date: 2026-09-26T13:57:01
branch: feature/generator-logging-mature
---

## 1. Decision

1. 先行Decision（`2026-09-26T12-53-38-feature-generator-logging-mature.md`）が採用した「全Adapter constructorで`retry == nil`をpanicする」方針を撤回し、次の方針で置き換える。
2. 各Adapter（`geminiapi`・`cursorapi`・`speech/gemini`・`r2`・記事source5種）のconstructorはnil検査を持たない。`retry`を単に構造体へ格納し、`retry.Retry(...)`を直接呼ぶ。contractコメントは`@require retry != nil（Composition Root の結線責務）`とし、`FallbackReporter`・`ProgressReporter`と同じ表現に揃える。
3. fail-fast検査は`internal/composition/produce_episode.go`の`newProduceEpisodeWithTopicCount`（Composition Rootの唯一の内部entrypoint）の先頭1点に集約する。`logw *delivery.LogWriter`という具体型に対して`if logw == nil { panic(...) }`を1回だけ行う。
4. `port.NoopRetryReporter{}`は削除する。既存test（61箇所）は、Retryが実際に呼ばれないパスは`nil`を渡し、呼ばれるパス（共有helperが複数testから使われ、呼ばれるかどうかをtestごとに見極めるコストが高い場合を含む）は各パッケージの`_test.go`内へ個別定義したSpy（`retryReporterSpy`）を渡す。

## 2. Reason

1. Opus 5.5への設計相談で、先行Decisionが見落としていた事実が判明した：`delivery.LogWriter.Event`は既に`if l == nil || l.w == nil { return }`という無言の握り潰しガードを持っており、`configuration-boundary.md`§7が禁止する「暗黙default」の実体はAdapter層ではなくこちらにあった。Adapter個別のpanicは、この経路の未注入を検出できていなかった。
2. Goの`typed nil`問題（`*T`型のnilをinterfaceへ代入すると`== nil`比較がfalseになる）により、Adapter側で`port.RetryReporter`型（interface）に対して`retry == nil`を検査しても、Compositionが`(*delivery.LogWriter)(nil)`のような具体型nilを渡した場合を検出できない。fail-fastは値が具体型のまま存在する最も内側の地点（Composition Root入口）でのみ有効に機能する。
3. `if w.retry != nil { w.retry.Retry(...) }`（先行Decision以前の実装）は、`configuration-boundary.md`§7が禁止する「未注入時に本物の別実装へ黒魔術的に切り替わる」形ではなく、「未注入時にその1件の観測イベントだけを送らない」という別category（optional observerの省略）である。診断（§7違反の疑い）は誤りだったが、「本番でretry logが要らない構成は存在しない、必須にすべき」という結論自体は正しかった。過剰だったのは処置（Adapterごとのpanic + Dummy新設）の方。
4. 兄弟interfaceである`FallbackReporter`・`ProgressReporter`は元々どちらもpanicを持たず、doc契約（`@require`）だけで非nilを表現している。`RetryReporter`だけpanicを持つ非対称は、新設したinterfaceの特別扱いに過ぎず、既存の設計と整合しない。
5. `port.NoopRetryReporter`は本番で使われる正当な理由（`io.Discard`のような実運用ユースケース）が無いtest専用のDummyであり、interface定義を持つ`port`パッケージに置くべきではない。このrepoの既存慣習（Fake/Stub/Spyは全部`_test.go`内に個別定義）に従う。

## 3. Rejected

1. 先行Decisionの方針（Adapter個別panic + port.NoopRetryReporter）をそのまま維持する案 — typed nil問題により実際の結線漏れを検出できず、かつ`port`パッケージへtest専用実装を混入させる副作用がある。
2. `retry`をnil許容にした上で、Adapter constructorにもfail-fast検査を重ねて置く案（Composition Rootと二重に検査） — 値の生まれる場所が1箇所（`main.go`→`newProduceEpisodeWithTopicCount`）である以上、検査も1点で足りる。複数箇所に複製すると、将来同種のinterfaceが増えるたびに全Adapterへ機械的な追加が要り、かつ「なぜここにも検査があるのか」を次の読み手が問う非対称が生まれる。
3. test専用のcommon helper package（例: `internal/testsupport`）を新設し`NoopRetryReporter`相当を1箇所に共有する案 — このrepoにはFake/Stub/Spy/Dummyを共有packageへ集約する既存慣習が無く、新規package 1つを追加する設計コストが、1行の空メソッド定義を各テストファイルへ複製するコストを上回る。
