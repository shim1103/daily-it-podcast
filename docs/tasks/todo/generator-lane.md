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
- [ ] Cursor CLI GitHub Actions capability probe — `docs/tasks/todo/generator-cursor-cli-github-actions-probe.md`
- [ ] runtime config loader実装 — `docs/tasks/todo/generator-runtime-config-loader.md`
- [x] Integration gate収集境界（secretなし Narrow）
- [ ] Composition HTTP Adapter移行（M1）— `docs/tasks/todo/generator-composition-http-adapters.md`
- [ ] HTTP SU/NI latest（getxapi / oauth / gemini / gdrive）— `docs/tasks/todo/generator-su-ni-*.md`
- [ ] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）— **D**
- [ ] GHA workflow で定期または手動実行 — **未切り出し**
- [ ] Cursor Narrow — `docs/tasks/todo/generator-narrow-gate-vendor-cursorcli.md`（child env再設計後）
- [ ] Broad Integration / System・E2E — **D**（Decision `2026-08-26T17-47-00`）

### 依存（実装順）

```text
A/B runtime config・secret方針（済）
C-01 AgentSecrets / local_real除去（済）
C-02 TwitterAPI.io除去（済）

C-03 Cursor CLI GHA capability probe（並行可）
C-04 runtime config loader（並行可）
  └→ M1 Composition HTTP Adapter移行
        └→ SU/NI latest: getxapi / oauth / gemini / gdrive

C-03実測
  └→ child env再設計（未切り出し）
        └→ Cursor Narrow（既存gate file、後回し）

既存Generator graph
  └→ ProduceEpisode.Run（D）
        └→ GHA production workflow
```

### Integration test 方針

```text
gate = secret なし Narrow / System 非 CI（DESIGN・既存Decision）
着手順 = 依存（実装順）を正。Cursor Narrow後回しは Decision 2026-08-28T12-49-01
```
