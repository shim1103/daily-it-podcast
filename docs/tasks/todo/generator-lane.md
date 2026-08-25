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
- [ ] process-env HTTP transport — `generator-processenv-http-transport.md`
- [ ] local AgentSecrets Cursor command launcher — `generator-agentsecrets-cursor-command-launcher.md`
- [ ] local AgentSecrets HTTP transport — `generator-agentsecrets-http-transport.md`
- [ ] 原稿 → TTS → 書込 を束ねる UseCase — **未切り出し**
- [ ] cmd 入口（`cmd/generator` は `.gitkeep` のみ）— **未切り出し**
- [ ] GHA workflow で定期または手動実行 — **未切り出し**

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `generator-processenv-http-transport.md` | production process-env HTTP transport + Adapter の `secrettransport` 切替 |
| `generator-agentsecrets-cursor-command-launcher.md` | local AgentSecrets × command（Cursor 専用 project） |
| `generator-agentsecrets-http-transport.md` | local AgentSecrets × HTTP（`secrettransport.Client`） |

### 依存（実装順）

```text
Drive 保存 / OAuth / 原稿検証 / Cursor CLI / secret transport contract（済）
  ├→ process-env command launcher（済）
  │     └→ local AgentSecrets Cursor command launcher
  ├→ process-env HTTP transport
  │     └→ local AgentSecrets HTTP transport
  └→ 原稿→TTS→書込 UseCase
      └→ cmd / GHA
```

Issue 分割の正: `docs/decisions/2026-08-25T14-20-18-feature-generator-processenv-command-launcher.md`  
2軸の正: `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md`  
方針: `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`

playback の読取 Adapter とは共有しない。音声は wav。
