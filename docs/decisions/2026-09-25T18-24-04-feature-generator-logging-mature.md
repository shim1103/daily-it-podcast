---
name: retry/fallback attemptの途中経過をport.RetryReporterでrealtime観測する
date: 2026-09-25T18:24:04
branch: feature/generator-logging-mature
---

## 1. Decision

1. `port`へ`RetryReporter`（`Retry(step string, attempt, max int, reason string)`）を新設する。`ProgressReporter`・`FallbackReporter`と並ぶ第3の観測interfaceとする。
2. retryループを持つInfrastructure Adapter（`manuscript/geminiapi`・`manuscript/cursorapi`・`speech/gemini`・`r2`）のconstructorへ`RetryReporter`をDIし、各attemptが失敗して次のattemptへ進む直前に`Retry`を呼ぶ。成功時・全attempt失敗で最終errorを返す時は呼ばない（成功は`ProgressReporter.Done`、最終失敗は呼び出し元のerror返却で表現済みのため）。
3. `delivery.LogWriter`に`Retry`を実装する。出力は既存の`Event(category, name, fields...)`のラッパーとし、`category=retry`固定、`fields`に`attempt`/`max`/`reason`を持つ。
4. retryを持たない単発呼び出しAdapter（記事source 5種: hackernews/lobsters/publickey/techcrunch/cloudwatch、`audio/ffmpeg`）には`RetryReporter`をDIしない。単発失敗は既存の`ProgressReporter`（Doneが呼ばれないことで失敗を示す規約）で表現済みであり、新規報告点を追加しない。
5. `LogWriter`に「全体終了原因」用の`Error`相当メソッドは追加しない。`main.go`の`writeExternalError`（`delivery.Format`）が既にこの役割を持つ。`Retry`は「この時点のこの試行が失敗した」というrealtime観測であり、「全体が最終的に失敗した」という`Format`の役割とは別の情報を運ぶため、同じ出力経路にまとめない。
6. `composition`側の各adapterコンストラクタ（`newGeminiTextWriterPrimary`等）のsignatureへ`logw`を追加し、`RetryReporter`として渡す。`produce_episode()`（UseCase層）自身は引き続き8段階のStart/Doneだけを持ち、下位Adapterへの観測interface配線責務は持たせない。

## 2. Reason

1. 09-24・09-25の`generator-produce-episode` workflow失敗調査で、Cursor adapterが最大5 attemptのretryループ内で`buildFn`（原稿検証）失敗を`lastBuildErr`へ握りつぶしており、何回・どのfieldで・何回同じ理由で失敗したかがログに一切残らないことが判明した（`previous attempt rejected`という文言だけが最終errorに埋め込まれ、attempt単位の経過は再現不能）。同型のretryループは`speech/gemini`・`r2`にも存在し、同じ問題を抱える。
2. `1:terms/observability/logging-boundary.md`§2は「外部systemを呼び出す点」でのログ出力を許可している。retryループはこの境界点そのものであり、各attemptの結果を都度報告することは既存原則の範囲内。
3. `1:terms/observability/logging-boundary.md`§6（本Decisionと同時に追記）が明確化した通り、「logging禁止」は出力実体（Logger直書き）を持つことの禁止であり、DIされた観測interfaceの呼び出しはInfrastructure層でも許可される。`RetryReporter`はこの区別に沿う設計。
4. 記事source・ffmpegはretryを持たず単発で失敗するため、新規報告点を足しても「1回失敗した」以上の情報が増えない。既存のProgressReporterの失敗表現（Doneが出ない）で十分であり、対象を広げると§3「同一errorを複数層でlogしない」（重複禁止）に抵触するリスクだけが増える。
5. `Error`のような全体失敗用の別メソッドを作らない理由は、`main.go`の`writeExternalError`が既にプロセス終了時の唯一の失敗表示点であるため。同じ役割のメソッドを2箇所に持つと、どちらが最終原因を表すかが呼び出し側の判断に委ねられ、重複出力または欠落のどちらかを招く。

## 3. Rejected

1. `Warn`/`Log`のような汎用メソッド名を追加する案 — 既に`Event(category, name, fields...)`という汎用layerが存在し、その上に汎用メソッドを重ねると「何が起きたか」が型として表現されず、後続の集計・grepが事象名ベースでできなくなる。具体的な事象名（`Retry`）を選ぶ。
2. 全Infrastructure Adapter（記事source・ffmpeg含む）へ一律`RetryReporter`をDIする案 — retryを持たないAdapterに「1回だけ呼ばれるRetry」を実装させても情報が増えず、DIする意味のない引数が増えるだけになる。retryループを実際に持つ4 Adapterだけに限定する。
3. `produce_episode()`（UseCase層）が下位Adapterへ`RetryReporter`を配る案 — UseCase層は8段階の大きな区切りだけを持つ設計を保ち、下位Adapterの配線責務は`composition`（Composition Root）に閉じる。UseCase層に配線責務を持たせると、Composition Rootの「全Portに対する具体的なAdapterを結線する唯一の場所」という責務が分散する。
4. `RetryReporter.Retry`を`ProgressReporter`のメソッドとして統合する案 — `ProgressReporter`は「段階の開始・完了」という1つのUseCase内の大きな区切りを表す。retry attemptは同じ段階内で複数回起きる、意味の異なるイベントであり、同一interfaceに混ぜると呼び出し側が「これは段階か、試行か」を都度判別する必要が生まれる。
