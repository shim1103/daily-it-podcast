## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

- [x] go.mod（module path）と `ItemSource` / `SourceItem` / 監視定数の境界 stub
- [x] 情報取得 Adapter（TwitterAPI.io / 旧 `PostSource`。Composition 結線済み。`ItemSource.List` 載せ替えは `generator-x-item-source.md`）
- [x] 情報取得 UseCase（`application.FetchSourceItems`。`List` 1回。監視 user は知らない）
- [x] `SpeechSynthesizer` / `SpeechAudio` / Gemini Adapter 定数（空）の境界 stub
- [x] GetXAPI Adapter（旧 `PostSource`。Composition 結線済み。`ItemSource.List` 載せ替えは `generator-x-item-source.md`）
- [ ] cmd 入口
- [ ] Cursor CLI の Infrastructure（`TextWriter` Adapter。Issue は `generator-cursor-text-writer.md`）
- [x] Gemini TTS Adapter（`SpeechSynthesizer`。Composition 結線済み。Issue 未作成）
- [ ] Drive 書込 Adapter（本番・WAV）— 詳細は `generator-drive-adapter.md`
- [ ] X Adapter を `ItemSource.List` へ載せ替える — 詳細は `generator-x-item-source.md`
- [ ] GHA workflow で定期または手動実行

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `generator-x-item-source.md` | X 両 Adapter を `ItemSource.List` へ。`WatchUserIDs` を infra へ。GitHub Issue 化しない |
| `generator-drive-adapter.md` | Drive 書込 Port + Google Drive API 本番 Adapter（json + wav） |
| `generator-cursor-text-writer.md` | Cursor CLI `TextWriter` Port + Adapter（non-interactive、JSON envelope） |

### 依存（実装順）

```text
ItemSource / FetchSourceItems（済）
  → generator-x-item-source（X Adapter を List へ）
  → contracts / SpeechSynthesizer（済）
  → drive-adapter（書込 Port + 本番 Adapter。Cursor と並行可）
  → 原稿→TTS→書込 UseCase（未切り出し）
  → cmd / GHA（未切り出し）
```

playback の読取 Adapter とは共有しない。音声は wav。
