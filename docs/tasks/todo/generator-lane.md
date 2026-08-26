## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。decisions は各 task file / 必要時に辿る。

- [x] go.mod（module path）と `ItemSource` / `SourceItem` / 監視定数の境界 stub
- [x] 情報取得 Adapter（TwitterAPI.io / GetXAPI。Composition 結線済み）
- [x] 監視 user 一括取得 UseCase（`application.FetchSourceItems`）
- [x] `SpeechSynthesizer` / Gemini Adapter / Drive / OAuth / WriteEpisode / Cursor CLI
- [x] process-env command launcher / HTTP transport
- [x] AgentSecrets HTTP proxy 正本吸収 / local AgentSecrets Cursor command launcher
- [x] cmd 入口（薄い Driving Adapter）
- [x] Integration 収集境界（A）— gate と `local_real` 分離
- [ ] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）— **D**
- [ ] GHA workflow で定期または手動実行 — **未切り出し**
- [ ] Narrow（C）— `docs/tasks/todo/generator-narrow-*.md` が正（GitHub Issue 化しない）
- [ ] Broad Integration / System・E2E — **D**（Decision `2026-08-26T17-47-00`）

### 依存（実装順）

```text
Drive / OAuth / 原稿検証 / Cursor CLI / secret transport（済）
  ├→ process-env command / HTTP（済）
  │     └→ AgentSecrets command / HTTP（済）
  └→ cmd 入口（済）
        ├→ ProduceEpisode.Run（D）
        └→ GHA
```

### Integration test 方針

```text
A/B 済: gate = secret なし Narrow / local_real 除外 / System 非 CI / CDC 非導入
C: docs/tasks/todo/generator-narrow-*.md（local_real + vendor gate）。実装は各 file の AC
D: Broad / System・E2E / vendor 実 API
```
