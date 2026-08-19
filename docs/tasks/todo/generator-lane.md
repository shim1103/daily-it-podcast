## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

- [x] go.mod（module path）と `PostSource` / `Post` / 監視定数の境界 stub
- [x] 情報取得 Adapter（TwitterAPI.io / `PostSource`。Composition 結線済み。Issue 未作成）
- [x] 監視 user 一括取得 UseCase（Issue 未作成。`application.FetchWatchedPosts`）
- [x] `SpeechSynthesizer` / `SpeechAudio` / Gemini Adapter 定数（空）の境界 stub
- [x] GetXAPI Adapter（`PostSource`。Composition 結線済み。Issue 未作成）
- [ ] cmd 入口
- [ ] Cursor CLI の Infrastructure（`TextWriter` Adapter。Issue は `generator-cursor-text-writer.md`）
- [x] Gemini TTS Adapter（`SpeechSynthesizer`。Composition 結線済み。Issue 未作成）
- [ ] Drive 保存 Adapter — `generator-drive-storage-adapter.md`
- [ ] Google OAuth refresh Adapter — `generator-google-oauth-adapter.md`
- [ ] Application 原稿検証 + WriteEpisode — `generator-episode-validation.md`
- [ ] GHA workflow で定期または手動実行

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `generator-cursor-text-writer.md` | Cursor CLI `TextWriter` |
| `generator-drive-storage-adapter.md` | Drive 保存（REST + `EpisodeWriter`） |
| `generator-google-oauth-adapter.md` | OAuth refresh + TokenSource |
| `generator-episode-validation.md` | Application schema 検証 + WriteEpisode |

### 依存（実装順）

```text
contracts / SpeechSynthesizer（済）
  → drive-storage-adapter
  → google-oauth-adapter
  → episode-validation
  → 原稿→TTS→書込 UseCase
  → cmd / GHA
```

方針: `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`

playback の読取 Adapter とは共有しない。音声は wav。
