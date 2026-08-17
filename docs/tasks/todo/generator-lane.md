## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

- [x] go.mod（module path）と `PostSource` / `Post` / 監視定数の境界 stub
- [x] 情報取得 Adapter（TwitterAPI.io / `PostSource`。Composition 結線済み。Issue 未作成）
- [x] 監視 user 一括取得 UseCase（Issue 未作成。`application.FetchWatchedPosts`）
- [x] `SpeechSynthesizer` / `SpeechAudio` / Gemini Adapter 定数（空）の境界 stub
- [ ] GetXAPI Adapter（Issue3・下に詳細。Issue / draft md 未作成）
- [ ] cmd 入口
- [ ] Cursor CLI の Infrastructure
- [ ] Gemini TTS Adapter（下の draft。Issue / gh 未作成）
- [ ] Drive の Infrastructure
- [ ] GHA workflow で定期または手動実行

### 未完了: Issue3 GetXAPI Adapter（本運用差し替え）

Issue1（TwitterAPI.io）で Port 契約を実測したあと、同じ `PostSource` を GetXAPI で満たす。create-issue template の正式 draft はまだ作らない。終わってない task として境界だけ残す。

- 目的: 本運用 vendor を GetXAPI にし、Composition から切り替え可能にする
- 依存: Issue1 完了後が安全（契約・変換の実測後）。path は独立
- 所有 path: `apps/generator/internal/infrastructure/x/getxapi/`
- 触ってよい: 上記 dir、Composition の **Adapter 選択・結線切り替えのみ**（twitterapiio 実装を消さない／無関係リファクタしない）
- 触禁止: `application/port/`、`entities/`、`infrastructure/x/twitterapiio/` の無関係変更、後回し decision 対象
- 受け入れの正: 既存 `PostSource` 契約（変えない・満たす）。vendor 型非漏出
- 秘密: 名前の正は README。値は secrets。code に書かない
- やること: HTTP client、cursor / since 打ち切り、raw → `models.Post`、Infrastructure Error、契約用 test、Composition 切り替え
- やらない: UseCase、trends、Reply/Repost/引用、profile cache、upsert/DB、media ローカル保存本体、GHA cron 全体
- 参照: `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md`、`docs/decisions/2026-08-15T17-43-09-feature-x-api-adoption.md`、Port / Post / constants、Issue1 draft
- 次アクション: Issue1 検証後に create-issue template で正式 draft → gh 作成

### 未完了: Gemini TTS Adapter（`SpeechSynthesizer`）

境界 stub 済み。HTTP / mp3 / retry / 定数の中身 / Composition は未実装。GitHub Issue は未作成。

- 詳細の正: `docs/tasks/todo/gemini-tts-adapter.md`
- 所有 path: `apps/generator/internal/infrastructure/speech/gemini/`
- 触ってよい: 上記 dir、Composition のこの Adapter 結線、空定数への公式値
- 触禁止: Port / `SpeechAudio` の契約変更、UseCase、Drive、cmd、GHA
- 受け入れの正: 既存 `SpeechSynthesizer` 契約（変えない・満たす）。vendor 型非漏出。戻りは mp3 bytes
- 次アクション: `shim gh create-issue` は別指示。実装着手は draft 本文で足りる
