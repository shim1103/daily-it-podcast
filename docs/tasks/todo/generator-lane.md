## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

- [x] go.mod（module path）と `ItemSource` / `SourceItem` / 監視定数の境界 stub
- [x] 情報取得 Adapter（TwitterAPI.io / `ItemSource`。Composition 結線済み。Issue 未作成）
- [x] 監視 user 一括取得 UseCase（Issue 未作成。`application.FetchSourceItems`）
- [x] `SpeechSynthesizer` / `SpeechAudio` / Gemini Adapter 定数（空）の境界 stub
- [x] GetXAPI Adapter（`ItemSource`。Composition 結線済み。Issue 未作成）
- [x] Gemini TTS Adapter（`SpeechSynthesizer`。Composition 結線済み。Issue 未作成）
- [x] Drive 保存 Adapter（`gdrive.RawEpisodeWriter`。PR #34 で完了）
- [x] Google OAuth refresh Adapter（`oauth.TokenSource`。PR #37 で完了）
- [x] Application 原稿検証 + WriteEpisode（`application.WriteEpisode`。PR #39 で完了）
- [x] Cursor CLI の Infrastructure（`TextWriter` Adapter）
- [x] process-env command launcher — production Cursor path 完了（todo file 削除済み）
- [x] process-env HTTP transport — production process-env HTTP 完了（todo file 削除済み）
- [x] AgentSecrets HTTP proxy 正本吸収 — `secrettransport/agentsecrets` 単一実装（todo 削除済み）
- [ ] local AgentSecrets Cursor command launcher — `generator-agentsecrets-cursor-command-launcher.md`（素材は `commandlaunch/agentsecrets`）
- [ ] 原稿 → TTS → 書込 UseCase（`ProduceEpisode.Run` 本体）— **D / 未切り出し**（入口契約は cmd 入口の正）
- [x] cmd 入口 — 薄い Driving Adapter 完了（todo 削除済み）
- [ ] GHA workflow で定期または手動実行 — **未切り出し**

### Issue 化待ち（詳細は file）

| file | 内容 |
|---|---|
| `generator-agentsecrets-cursor-command-launcher.md` | local AgentSecrets × command（`commandlaunch.Launcher`。EnvWrapper 素材は `commandlaunch/agentsecrets`） |

### 依存（実装順）

```text
Drive 保存 / OAuth / 原稿検証 / Cursor CLI / secret transport contract（済）
  ├→ process-env command launcher（済）
  │     └→ local AgentSecrets Cursor command launcher
  ├→ process-env HTTP transport（済）
  │     └→ AgentSecrets HTTP proxy 正本吸収（済。`secrettransport/agentsecrets`）
  └→ cmd 入口（済。Decision: `docs/decisions/2026-08-25T23-12-31-feature-generator-cmd-usecase-boundary.md`）
        ├→ ProduceEpisode.Run 本体（D）
        └→ GHA
```

cmd 入口の正: `docs/decisions/2026-08-25T23-12-31-feature-generator-cmd-usecase-boundary.md`  
cmd 成否観測の正: `docs/decisions/2026-08-26T14-42-16-feature-generator-cmd-entrypoint.md`  
Issue 分割の正: `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md`  
2軸の正: `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md`  
HTTP proxy 正本: `docs/decisions/2026-08-25T19-36-11-feature-generator-agentsecrets-http-transport.md`  
command 素材の置き場: `docs/decisions/2026-08-26T12-14-00-feature-generator-agentsecrets-http-proxy-absorb.md`  
方針: `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`
