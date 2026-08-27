## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

未完了の達成契約は `docs/tasks/todo/generator-*.md` が正。本 lane は進捗 index のみ。decisions は各 task file / 必要時に辿る。

- [x] go.mod（module path）と `ItemSource` / `SourceItem` / 監視定数の境界 stub
- [x] 情報取得 Adapter（GetXAPIはproduction結線済み。TwitterAPI.io旧実装はC-02で削除）
- [x] 監視 user 一括取得 UseCase（`application.FetchSourceItems`）
- [x] `SpeechSynthesizer` / Gemini Adapter / Drive / OAuth / WriteEpisode / Cursor CLI
- [x] process-env command launcher / HTTP transport
- [ ] AgentSecrets / `local_real`除去 — `docs/tasks/todo/generator-remove-agentsecrets-local-real.md`
- [x] cmd 入口（薄い Driving Adapter）
- [ ] TwitterAPI.io旧artifact除去 — `docs/tasks/todo/generator-remove-twitterapiio.md`
- [ ] Cursor CLI GitHub Actions capability probe — `docs/tasks/todo/generator-cursor-cli-github-actions-probe.md`
- [ ] runtime config loader実装 — `docs/tasks/todo/generator-runtime-config-loader.md`
- [x] Integration gate収集境界（secretなし Narrow）
- [ ] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）— **D**
- [ ] GHA workflow で定期または手動実行 — **未切り出し**
- [ ] Narrow（C）— `docs/tasks/todo/generator-narrow-*.md` が正（GitHub Issue 化しない）
- [ ] Broad Integration / System・E2E — **D**（Decision `2026-08-26T17-47-00`）

### 依存（実装順）

```text
A/B runtime config・secret方針（済）
  ├→ AgentSecrets / local_real除去（C-01）
  │     └→ TwitterAPI.io除去（C-02）
  ├→ Cursor CLI GHA capability probe（C-03、並行可）
  └→ runtime config loader（C-04、並行可）

既存Generator graph
  └→ ProduceEpisode.Run（D）
        └→ GHA production workflow
```

### Integration test 方針

```text
A/B 済: gate = secret なし Narrow / System 非 CI / CDC 非導入
C-01: local_real入口とAgentSecrets Narrowを削除
C-02: TwitterAPI.io Narrow taskを削除
C: 残るdocs/tasks/todo/generator-narrow-*.mdはsecretなしvendor gate。実装は各fileのAC
D: Broad / System・E2E / vendor実API
```
