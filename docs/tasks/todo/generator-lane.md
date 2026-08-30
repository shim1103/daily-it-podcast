## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。decisions は各 task file / 必要時に辿る。

- [x] go.mod（module path）と `ItemSource` / `SourceItem` / 監視定数の境界 stub
- [x] 情報取得 Adapter（GetXAPIはproduction結線済み。TwitterAPI.io旧実装は削除済み）
- [x] 監視 user 一括取得 UseCase（`application.FetchSourceItems`）
- [x] `SpeechSynthesizer` / Gemini Adapter / Drive / OAuth / WriteEpisode / Cursor CLI
- [x] process-env command launcher / HTTP transport
- [x] AgentSecrets / `local_real`除去
- [x] cmd 入口（薄い Driving Adapter）
- [x] TwitterAPI.io旧artifact除去
- [x] Cursor CLI GitHub Actions capability probe
- [x] runtime config loader実装（`internal/config` KISS化まで。error最終形は下記 error-taxonomy-unify）
- [ ] error 3層表現統一（Domain/Infra/Config → Infra pattern）— `docs/tasks/todo/generator-error-taxonomy-unify.md`
- [x] Integration gate収集境界（secretなし Narrow）
- [x] Composition HTTP Adapter移行（M1）
- [x] HTTP SU/NI latest（getxapi / oauth / gemini / gdrive）
- [x] build ComposeBrief
- [x] build manuscript draft parse（`ManuscriptDraftFromWriterOutput` 実装済み。尺モデルは Decision `2026-08-30T03-06-53`）
- [x] build WAV concat / duration（`WavDurationSec` / `ConcatWAV` 実装済み）
- [ ] composition ProduceEpisode 結線 — `docs/tasks/todo/generator-composition-produce-episode-wiring.md`
- [ ] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）— **D**
- [ ] GHA workflow で定期または手動実行 — **D**
- [x] Cursor Narrow（`cursorcli_narrow_integration_test.go`。processenv SU から実 process 観測を Narrow へ分離。`SandboxValue` は GHA 前提で `disabled`）
- [ ] Broad Integration / System・E2E — **D**（Decision `2026-08-26T17-47-00`）

### 依存（実装順）

```text
A/B runtime config・secret方針（済）
C-01 AgentSecrets / local_real除去（済）
C-02 TwitterAPI.io除去（済）

C-03 Cursor CLI GHA capability probe（済）
C-04 runtime config loader（済）
  ├→ error-taxonomy-unify
  └→ M1 Composition HTTP Adapter移行（済）
        └→ SU/NI latest: getxapi / oauth / gemini / gdrive（済）

produce-episode（A/B 済。docs/decisions/2026-08-29T14-10 〜 17-00）
  ├→ build compose-brief / draft parse / wav concat（並行可）
  ├→ composition ProduceEpisode 結線（並行可）
  └→ D: ProduceEpisode.Run → GHA production workflow

C-03実測
  └→ child env再設計（済）
        └→ Cursor Narrow（済）
```

### D（未決・未実測・文案）

| topic | 概要 |
|---|---|
| `ProduceEpisode.Run` | C build + composition 後に orchestration 実装 |
| Prompt / limits 文案・数値 | 尺モデル（秒正本・合計対象・定数整合）は Decision `2026-08-30T03-06-53` で確定・実装済み。残るのは実運用データを見ての `CharsPerSecond` と各 field 秒数の微調整 |
| 挨拶文案 | `ClosingFarewell` 最終 copy、OpeningGreeting の date 読み上げ整形 |
| composite 高度化 | dedup / sort（2 情報源後） |
| 第 2 情報源 Adapter | 別 Issue 化待ち |
| GHA production workflow | Run green 後 |
| Broad / System・E2E | Cursor Narrow 後 |

### Integration test 方針

```text
gate = secret なし Narrow / System 非 CI（DESIGN・既存Decision）
着手順 = 依存（実装順）を正。Cursor Narrow後回しは Decision 2026-08-28T12-49-01
```
