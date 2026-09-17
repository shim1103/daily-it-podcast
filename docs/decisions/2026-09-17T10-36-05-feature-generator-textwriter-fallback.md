---
name: system-test の topic 数を環境変数で任意指定可能にし、source 取得件数と LLM 生成 topic 数の両方に反映する
date: 2026-09-17T10:36:05
branch: feature/generator-textwriter-fallback
---

## 1. Decision

1. 本番の topic 数（`entities/constants.DraftTopicCountTarget`、および `manuscript_draft_seconds.go` が畳み込む全体尺・margin）は const のまま変更しない。
2. system-test 実行時の topic 数は、`test/system` package が `os.Getenv`（`SYSTEM_TEST_TOPIC_COUNT`）で直接読み、int へ parse した上で、`composition.NewProduceEpisodeFromEnvWithTopicCount(topicCount, logw)` へ渡す。本番経路 `NewProduceEpisodeFromEnv` はこの汎用関数へ `constants.DraftTopicCountTarget` を渡して委譲する（system-test 専用の別 entry point を新設するのではなく、topicCount を引数化した単一経路に本番・test 双方が乗る）。raw environment を読む境界は `test/system`（Frameworks & Drivers 相当）に閉じ、Entities・Application・Use Case 層は raw env を読まない（`architecture/configuration-boundary.md` §2 の原則を維持）。
3. 環境変数が指定する topic 数は、**source 取得件数の上限**（`sourceMaxItems = topicCount * constants.SourceItemsPerTopic`。本番 topicCount 時は 0 を渡し各 Adapter の既存 `MaxStoriesScanned` へフォールバックさせ、既定動作を変えない）と、**LLM に生成させる topic 数**（`ManuscriptDraftFromWriterOutput` の topic 数検証、prompt への埋め込み）の両方に反映する。
4. topic 数に依存する秒数・文字数・margin の計算式（`manuscript_draft_seconds.go` の `DraftTotalTgtSec` 等と同型の畳み込み）を、`DraftTopicCountTarget` という const 専用ではなく、任意の topic 数を引数に取る関数（`TotalTgtSecFor` 等）へ一般化する。本番はこの関数へ `DraftTopicCountTarget` を渡し、system-test は環境変数由来の値を渡す。
5. `application/build` の topic 数検証（`draft_from_writer.go`）と prompt 埋め込み（`brief_limits_embed.go` 等）は、期待 topic 数を引数として受け取る形に変える（`ManuscriptDraftFromWriterOutput` を `buildFn` として DI する既存パターンと同型）。

## 2. Reason

1. 本番の topic 数を可変にしないのは、`DraftTopicCountTarget` を正本とする定数畳み込み（先行 Decision `2026-09-15T23-42-51`）が、topic 数を固定値にすることで「各 field の margin 合計が全体 margin を上回る」という圧力契約を成立させていたため。topic 数を本番でも可変にすると、この圧力契約の前提（`Σ_field margin > TotalMarginSec` が topic 数に依存しない形にした設計）が崩れる。system-test だけの一時的な値として扱うことで、本番の契約を壊さずに済む。
2. `test/system` が `os.Getenv` を直接読む設計は、一見 `architecture/configuration-boundary.md` の「Entities・Application は raw environment API を読まない」という原則に反するように見えるが、`test/system` package はテストコードであり Composition Root より外側（テスト実行環境という Frameworks & Drivers 相当の境界）に位置する。raw env を読んで typed 値へ変換する責務がこの境界に閉じている限り、原則には反しない。
3. topic 数を「source 取得件数の上限」だけでなく「LLM 生成 topic 数」にも反映する理由：system-test の目的は「1 回の実行あたりの Gemini 呼び出し回数（TTS の `topics+2` 束、TextWriter の invalid-draft retry 等）を絞って quota 消費を抑える」ことであり、source 取得件数だけを絞っても LLM が生成する topic 数（およびそれに伴う TTS 呼び出し回数）が減らなければ節約にならない。
4. 畳み込み計算式を関数化する理由：`DraftTopicCountTarget` を const のまま、topic 数に依存する秒数・margin を「任意の topic 数を受け取る関数」として一般化すれば、本番は定数を渡すだけで済み、system-test 専用の別定数セットを重複して持つ必要がない（DRY）。
5. 本番・test を「別 entry point」ではなく「単一の topicCount 引数化経路への委譲」にした理由：`newProduceEpisode` と `newProduceEpisodeWithTopicCount` を完全に別実装で持つと、結線ロジック（source・TextWriter・TTS の組み立て）が二重管理になる。本番用の薄い委譲関数 1 つに留めることで、結線の正本を単一に保てる。

## 3. Rejected

1. **本番の `DraftTopicCountTarget` 自体を実行時に上書き可能な変数にする案** — 本番の圧力契約（先行 Decision `2026-09-15T23-42-51`）が topic 数固定を前提にしており、可変にすると契約が崩れる。
2. **system-test 専用の topic 数を固定 1 パターンのみ（例: 常に 1）にする案** — 任意の値を受け取れる汎用パラメータが要求であり、固定 1 パターンでは要求を満たさない。
3. **system-test 専用の秒数・margin 定数セットを、本番の定数ブロックとは別に複数個（topic 数のパターンごとに）用意する案** — 任意の topic 数に対応するには無限にパターンを増やす必要があり非現実的。計算式を関数化する方が任意の値に対応でき、かつ DRY。
4. **`application/build` の topic 数検証・prompt 埋め込みを、Composition Root から正式な DI 経路で伝播させる案** — 本番経路の挙動を変更しないという方針（§1-1）と矛盾する。system-test からの引数注入に留める方が、本番への影響範囲を最小化できる。
5. **本番用 `newProduceEpisode` と system-test 用の結線を完全に別関数として重複実装する案** — 結線ロジック（5 情報源・TextWriter fallback・TTS 合成の組み立て）の正本が二重になり、片方だけ更新し忘れる regression リスクが生まれる。
