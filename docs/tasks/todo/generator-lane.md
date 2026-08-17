## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

- [x] go.mod（module path）と `PostSource` / `Post` / 監視定数の境界 stub
- [x] 情報取得 Adapter（TwitterAPI.io / `PostSource`。Composition 結線済み。Issue 未作成）
- [x] 監視 user 一括取得 UseCase（Issue 未作成。`application.FetchWatchedPosts`）
- [ ] GetXAPI Adapter（Issue3 draft: `x-getxapi-adapter.md`。Issue 未作成）
- [ ] cmd 入口
- [ ] Cursor CLI / Gemini / Drive の Infrastructure
- [ ] GHA workflow で定期または手動実行
