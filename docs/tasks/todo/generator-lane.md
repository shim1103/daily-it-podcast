## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。依存順は各 task file の Dependencies を正とする。


<<<<<<< HEAD
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
- [x] composition ProduceEpisode 結線（composite `ItemSource` 経由で Fetch。factory は入れず結線直書き）
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
  ├→ composition ProduceEpisode 結線（済）
  └→ D: ProduceEpisode.Run → GHA production workflow

C-03実測
  └→ child env再設計（済）
        └→ Cursor Narrow（済）
```
=======
- [ ] composition ProduceEpisode 結線 — `docs/tasks/todo/generator-composition-produce-episode-wiring.md`
- [ ] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）— **D**
- [ ] Broad Integration — `docs/tasks/todo/generator-broad-integration-produce-episode.md`
- [ ] Cursor Narrow — `docs/tasks/todo/generator-narrow-gate-vendor-cursorcli.md`（child env再設計後）
- [ ] GHA 本番 produce — workflow 済（`generator-produce-episode.yml`）。Run 実装後に緑化。Secret/Variable 登録は人手
- [ ] System — workflow 済（`generator-system.yml`）。suite 実装と TEST_* 登録は後続（`generator-system-e2e-produce-episode`）
>>>>>>> 5771e48 (docs(tasks): Broad Integration 達成契約を切り D を lane に残す)

### D（未決・未実測・文案）

| topic | 概要 |
|---|---|
| `ProduceEpisode.Run` | composition 後に orchestration 実装 |
| Prompt / limits 文案・数値 | 尺モデルは Decision `2026-08-30T03-06-53`。残は実運用後の微調整 |
| 挨拶文案 | `ClosingFarewell` 最終 copy、OpeningGreeting の date 読み上げ整形 |
| composite 高度化 | dedup / sort（2 情報源後） |
| 第 2 情報源 Adapter | 別 Issue 化待ち |
| GHA production workflow | YAML・inventory 名は済。Run 未完のため定時は赤になりうる。repo へ本番 Secret/Variable を登録する人手作業が残る |
| `generator-system-e2e-produce-episode` | workflow 済。suite 本体・assert・TEST_* 値の登録が未。schedule / required は Decision `2026-08-30T12-49-01`（週次・required は対象外） |

### Integration test 方針

```text
gate = secret なし Narrow + Broad（Decision 2026-08-30T11-56-00）
System = gate 外・週次 + dispatch（Decision 2026-08-30T12-49-01）
本番 produce = 毎日 07:00 JST + dispatch（同 Decision）
Cursor Narrow 後回し = Decision 2026-08-28T12-49-01
```
