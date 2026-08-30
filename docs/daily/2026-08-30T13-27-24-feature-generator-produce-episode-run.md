---
name: ProduceEpisode.Run 本体を Builder として実装し非決定依存を Composition 注入へ寄せる
date: 2026-08-30T13:27:24
session_id: none
branch: feature/generator-produce-episode-run
prev: なし
---

## 1. Summary

`ProduceEpisode.Run` 本体を実装した。Fetch → `build.ComposeBrief` → TextWriter →
`build.ManuscriptDraftFromWriterOutput` → 表示日付確定 → TTS（Greeting / Intro /
topic ごとの Preface・Detail / ClosingSummary / Farewell を各 1 Synthesize）→
`build.WavDurationSec` で無音込み累積 startSec・durationSec → `build.ConcatWAV` →
opaque UUID episodeId → `build.MarshalManuscript` → `WriteEpisode.Run` を Builder
として結線する。純粋計算は `build` の `SpeechTexts` / `Timeline` / `MarshalManuscript`
へ切り出し、Run は Port orchestration だけに保つ。

Fetch 後 0 件は Run で自前判定せず `build.ComposeBrief` 既存の `no_source_items`
throw をそのまま propagate する。`build` helper の内部不整合（segment 数・topic 数
不一致）は panic をやめ新 Op `inconsistent_episode_assembly` の Domain Error で返す。

非決定な依存（opaque UUID 発行・表示タイムゾーンの tzdata 解決）は Application が
直接触れず Composition の production runtime 既定値（`newEpisodeID` /
`sharedDisplayLocation`）として `NewProduceEpisode` へ注入する。`displayDate` は
`*time.Location` を受ける純粋関数へ戻し、`time.LoadLocation` の I/O と err 戻り値を
Run から排除した。`config.Load` も `os.LookupEnv` 直呼びから `sharedLookupEnv()`
経由へ揃えた。

挨拶定型は `OpeningGreetingTemplate` / `ClosingFarewell` とも `%s` 入り template
（Builder が `fmt.Sprintf` で読み上げ日付 `YYYY年M月D日` を注入）へ shim が確定した。

## 2. Changes

- manager(non-edit, audit 専任) + haiku/sonnet executor 体制。executor 実装 →
  manager 監査 → shim 指摘 → executor 再修正のループを複数回。
- shim による割り込み修正: `ClosingFarewell` を空文字→勝手な文案へ変えた executor 変更を
  revert（shim 本人が後で `%s` 入り template として確定）、`newEpisodeID` の層帰属確認
  （Composition 維持が正）、`time.LoadLocation` の Application 混入を Composition へ移動、
  `build` helper の panic 全廃、`os.LookupEnv` の shared 経由化、`jstDate`→`displayDate`
  rename、不要な定数 test file 削除。
- sociable-unit test: `ProduceEpisode.Run` 9 case（Fault Isolation で協調先の判定結果を
  再 assert せず入力伝播と非書込のみ）、`build.episode_assembly` 8 case、`newEpisodeID`
  1 case。fixture（valid wire JSON・固定尺 WAV・manuscript Unmarshal helper）は
  application test 側に別 file で分離。test の Location は tzdata 非依存の
  `time.FixedZone("JST", 9*3600)`。
- pre-commit / pre-push は generator + playback 全系統。generator unit coverage 91.1%。
  push は sandbox proxy が git SSH 認証・`gh` API TLS を弾いたため sandbox 無効化で実行。
- GitHub Issue は無し（local lane が正）。PR #97 base `develop`。
- PR 作成後、session 中に develop が #95（Cursor CLI Narrow）/ #96（Broad・System
  E2E plan）で先行していたことが判明。`episode_greetings.go`（`ClosingFarewell` 文言）・
  `generator-lane.md`（lane 再編）・`lessons/index.md`（append 位置）が衝突。
  `origin/develop` を merge で取り込み、`ClosingFarewell` は develop の文案更新
  commit（`c14d929`）を採用、lane は develop の再編へ Run 完了を反映、lessons は
  両 branch の append を統合。merge 後 gate 全 green を確認して push。

### Commits

- `fb77325`
- `f9d29cc`
- `f7a55aa`
- `70bf057`
- `7163608`
- `a1da922`（merge origin/develop・conflict 解消）
